package lexer

import (
	"html"
	"strings"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

// Templates (rune:dom, OQ-DOM-01). `template {` switches the lexer to
// markup: tags, attributes, and text, until the matching `}`. Inside
// markup, `{{ expr }}`, the condition of `$if (cond)` and
// `$elseif (cond)`, and the header of `$foreach (T x in a)` are lexed as
// ordinary RuneScript tokens.
//
// In template text, '{' and '}' are reserved: `{{` starts an
// interpolation and a lone `}` ends the template, so a literal brace is
// written {{ "{" }}. HTML character references (&amp;, &lt;, &#169; ...)
// are decoded in text and in quoted attribute values.

// Markup token kinds. Text holds the tag or attribute name, the decoded
// text, the decoded attribute value, or the directive ("$if").
const (
	MText        Kind = iota + 100 // text between tags
	MTagOpen                       // `<name`
	MTagEnd                        // `>` ending an opening tag
	MSelfClose                     // `/>`
	MTagClose                      // `</name>`
	MAttr                          // an attribute name inside a tag
	MAttrValue                     // a quoted attribute value
	MInterpOpen                    // `{{`
	MInterpClose                   // `}}`
	MDirective                     // `$if`, `$elseif`, `$else`, `$endif`, `$foreach`, `$endforeach`
)

func init() {
	for k, name := range map[Kind]string{
		MText: "template text", MTagOpen: "'<'", MTagEnd: "'>'", MSelfClose: "'/>'", MTagClose: "closing tag",
		MAttr: "attribute", MAttrValue: "attribute value", MInterpOpen: "'{{'", MInterpClose: "'}}'", MDirective: "directive",
	} {
		kindNames[k] = name
	}
}

var directives = []string{"$endforeach", "$foreach", "$elseif", "$endif", "$else", "$if"} // longest first

// atTemplate reports whether the last token is the identifier template
// and the next character, after spaces and comments, is '{'.
func (l *lexer) atTemplate() bool {
	if len(l.toks) == 0 {
		return false
	}
	last := l.toks[len(l.toks)-1]
	if last.Kind != Ident || last.Text != "template" {
		return false
	}
	save, line, col := l.off, l.line, l.col
	l.skipSpaceAndComments()
	ok := l.off < len(l.src) && l.src[l.off] == '{'
	l.off, l.line, l.col = save, line, col
	return ok
}

func (l *lexer) has(prefix string) bool {
	return strings.HasPrefix(string(l.src[l.off:min(len(l.src), l.off+len(prefix))]), prefix)
}

func (l *lexer) skip(n int) {
	for i := 0; i < n; i++ {
		l.advance()
	}
}

// directive returns the directive at the input, or "".
func (l *lexer) directive() string {
	for _, d := range directives {
		if l.has(d) {
			next := l.off + len(d)
			if next >= len(l.src) || !(isIdentStart(l.src[next]) || isDigit(l.src[next])) {
				return d
			}
		}
	}
	return ""
}

// markup lexes a template's content, up to and including its closing '}'.
func (l *lexer) markup() {
	for {
		start := l.pos()
		switch {
		case l.off >= len(l.src):
			l.errorf(diagnostics.LexUnterminatedTemplate, start, "the template has no closing '}'")
			return
		case l.src[l.off] == '}':
			l.advance()
			l.emit(RBrace, "}", start)
			return
		case l.has("<!--"):
			l.skip(4)
			for !l.has("-->") {
				if l.off >= len(l.src) {
					l.errorf(diagnostics.LexUnterminatedBlock, start, "unterminated comment <!-- ... -->")
					return
				}
				l.advance()
			}
			l.skip(3)
		case l.has("</"):
			l.closeTag(start)
		case l.src[l.off] == '<' && l.off+1 < len(l.src) && isIdentStart(l.src[l.off+1]):
			l.tag(start)
		case l.has("{{"):
			l.interpolation(start)
		case l.src[l.off] == '{':
			l.advance()
			l.errorf(diagnostics.LexUnexpectedChar, start, "'{' in a template starts an interpolation, {{ expression }}").
				WithHelp(`write a literal brace as {{ "{" }}`)
		case l.directive() != "":
			d := l.directive()
			l.skip(len(d))
			l.emit(MDirective, d, start)
			if d == "$if" || d == "$elseif" || d == "$foreach" {
				l.condition(d)
			}
		default:
			l.text(start)
		}
	}
}

// text lexes template text up to the next tag, brace, or directive.
func (l *lexer) text(start diagnostics.Pos) {
	for l.off < len(l.src) {
		c := l.src[l.off]
		if c == '{' || c == '}' || c == '<' && (l.has("</") || l.has("<!--") || l.off+1 < len(l.src) && isIdentStart(l.src[l.off+1])) ||
			c == '$' && l.directive() != "" {
			break
		}
		l.advance()
	}
	l.emit(MText, html.UnescapeString(string(l.src[start.Offset:l.off])), start)
}

func isMarkupName(c byte) bool {
	return isIdentStart(c) || isDigit(c) || c == '-' || c == ':' || c == '.'
}

func (l *lexer) name() string {
	from := l.off
	for l.off < len(l.src) && isMarkupName(l.src[l.off]) {
		l.advance()
	}
	return string(l.src[from:l.off])
}

func (l *lexer) skipMarkupSpace() {
	for l.off < len(l.src) && strings.IndexByte(" \t\r\n", l.src[l.off]) >= 0 {
		l.advance()
	}
}

// tag lexes an opening tag: `<name`, its attributes, and `>` or `/>`.
func (l *lexer) tag(start diagnostics.Pos) {
	l.advance()
	l.emit(MTagOpen, l.name(), start)
	for {
		l.skipMarkupSpace()
		at := l.pos()
		switch {
		case l.off >= len(l.src):
			l.errorf(diagnostics.LexUnterminatedTemplate, start, "the tag is not closed with '>'")
			return
		case l.src[l.off] == '>':
			l.advance()
			l.emit(MTagEnd, ">", at)
			return
		case l.has("/>"):
			l.skip(2)
			l.emit(MSelfClose, "/>", at)
			return
		case isIdentStart(l.src[l.off]):
			l.emit(MAttr, l.name(), at)
			l.skipMarkupSpace()
			if l.off >= len(l.src) || l.src[l.off] != '=' {
				continue // a boolean attribute
			}
			eq := l.pos()
			l.advance()
			l.emit(Assign, "=", eq)
			l.skipMarkupSpace()
			l.attrValue()
		default:
			l.advance()
			l.errorf(diagnostics.LexUnexpectedChar, at, "unexpected %q in a tag", l.src[at.Offset]).
				WithHelp(`attributes are written name="value", name={{ expression }}, or name alone for a boolean attribute`)
		}
	}
}

// attrValue lexes a quoted attribute value or {{ expression }}.
func (l *lexer) attrValue() {
	start := l.pos()
	switch {
	case l.has("{{"):
		l.interpolation(start)
	case l.off < len(l.src) && (l.src[l.off] == '"' || l.src[l.off] == '\''):
		q := l.src[l.off]
		l.advance()
		from := l.off
		for l.off < len(l.src) && l.src[l.off] != q {
			l.advance()
		}
		if l.off >= len(l.src) {
			l.errorf(diagnostics.LexUnterminatedString, start, "unterminated attribute value")
			return
		}
		value := html.UnescapeString(string(l.src[from:l.off]))
		l.advance()
		l.emit(MAttrValue, value, start)
	default:
		l.errorf(diagnostics.LexUnexpectedChar, start, "an attribute value is quoted, or an interpolation").
			WithHelp(`write name="value" or name={{ expression }}`)
	}
}

// closeTag lexes `</name>`.
func (l *lexer) closeTag(start diagnostics.Pos) {
	l.skip(2)
	name := l.name()
	l.skipMarkupSpace()
	if l.off >= len(l.src) || l.src[l.off] != '>' {
		l.errorf(diagnostics.LexUnterminatedTemplate, start, "the closing tag </%s is not closed with '>'", name)
		return
	}
	l.advance()
	l.emit(MTagClose, name, start)
}

// interpolation lexes `{{ expression }}`: RuneScript tokens up to the
// `}}` that is not inside braces of the expression, e.g. a lambda body.
func (l *lexer) interpolation(start diagnostics.Pos) {
	l.skip(2)
	l.emit(MInterpOpen, "{{", start)
	depth := 0
	for {
		l.skipSpaceAndComments()
		at := l.pos()
		switch {
		case l.off >= len(l.src):
			l.errorf(diagnostics.LexUnterminatedTemplate, start, "the interpolation has no closing '}}'")
			return
		case depth == 0 && l.has("}}"):
			l.skip(2)
			l.emit(MInterpClose, "}}", at)
			return
		}
		l.token()
		switch l.toks[len(l.toks)-1].Kind {
		case LBrace:
			depth++
		case RBrace:
			depth--
		}
	}
}

// condition lexes the parenthesized condition after $if or $elseif.
func (l *lexer) condition(d string) {
	l.skipSpaceAndComments()
	if l.off >= len(l.src) || l.src[l.off] != '(' {
		help := d + " (count() > 0)"
		if d == "$foreach" {
			help = "$foreach (string name in names) … $endforeach"
		}
		l.errorf(diagnostics.LexUnexpectedChar, l.pos(), "%s needs %s in parentheses", d, map[bool]string{true: "a header", false: "a condition"}[d == "$foreach"]).
			WithHelp(help)
		return
	}
	depth := 0
	for {
		l.skipSpaceAndComments()
		if l.off >= len(l.src) {
			l.errorf(diagnostics.LexUnterminatedTemplate, l.pos(), "the condition of %s has no closing ')'", d)
			return
		}
		l.token()
		switch l.toks[len(l.toks)-1].Kind {
		case LParen:
			depth++
		case RParen:
			depth--
			if depth == 0 {
				return
			}
		}
	}
}
