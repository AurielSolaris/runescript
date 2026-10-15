// Package lexer converts RuneScript source text into tokens.
package lexer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

// Kind classifies a token.
type Kind int

const (
	EOF Kind = iota
	Ident
	Int    // integer literal; Token.Text holds the digits
	String // string literal; Token.Text holds the decoded value
	Keyword
	Reserved // reserved for a future feature; not usable as an identifier
	LParen
	RParen
	LBrace
	RBrace
	Comma
	Semicolon
	Dot
	Assign
	Op // operator; Token.Text holds its spelling
)

var kindNames = map[Kind]string{
	EOF: "end of file", Ident: "identifier", Int: "integer literal",
	String: "string literal", Keyword: "keyword", Reserved: "reserved word",
	LParen: "'('", RParen: "')'", LBrace: "'{'", RBrace: "'}'", Comma: "','",
	Semicolon: "';'", Dot: "'.'", Assign: "'='", Op: "operator",
}

func (k Kind) String() string { return kindNames[k] }

// keywords implemented by the current milestone.
var keywords = map[string]bool{
	"package": true, "return": true,
	"if": true, "else": true, "while": true, "for": true, "break": true, "continue": true,
	"true": true, "false": true, "bool": true,
	"public": true, "protected": true, "private": true,
	"int": true, "sint": true, "uint": true, "string": true, "void": true,
}

// reserved words named by the specification for features that are not yet
// implemented. Reserving them now keeps today's programs valid later.
var reserved = map[string]bool{
	"class": true, "struct": true, "use": true,
	"float": true, "char": true, "array": true, "matrix": true,
	"int8": true, "int16": true, "int32": true, "int64": true,
	"uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true,
}

// Token is a lexical token with its source span.
type Token struct {
	Kind Kind
	Text string
	Span diagnostics.Span
}

// Describe returns a human-readable token description for diagnostics.
func (t Token) Describe() string {
	switch t.Kind {
	case Ident, Keyword, Reserved, Op:
		return fmt.Sprintf("'%s'", t.Text)
	case Int:
		return fmt.Sprintf("integer literal %s", t.Text)
	}
	return t.Kind.String()
}

type lexer struct {
	src   []byte
	off   int
	line  int
	col   int
	toks  []Token
	diags []*diagnostics.Diagnostic
}

// Lex tokenizes src. The token slice always ends with an EOF token.
// Lexical errors are reported as diagnostics; lexing continues after them.
func Lex(src []byte) ([]Token, []*diagnostics.Diagnostic) {
	l := &lexer{src: src, line: 1, col: 1}
	l.run()
	return l.toks, l.diags
}

func (l *lexer) pos() diagnostics.Pos {
	return diagnostics.Pos{Offset: l.off, Line: l.line, Col: l.col}
}

func (l *lexer) peek(n int) byte {
	if l.off+n < len(l.src) {
		return l.src[l.off+n]
	}
	return 0
}

// advance consumes one code point.
func (l *lexer) advance() {
	if l.off >= len(l.src) {
		return
	}
	r, size := utf8.DecodeRune(l.src[l.off:])
	l.off += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
}

func (l *lexer) errorf(code diagnostics.Code, start diagnostics.Pos, format string, args ...any) *diagnostics.Diagnostic {
	d := diagnostics.New(code, diagnostics.Span{Start: start, End: l.pos()}, format, args...)
	l.diags = append(l.diags, d)
	return d
}

func (l *lexer) emit(kind Kind, text string, start diagnostics.Pos) {
	l.toks = append(l.toks, Token{Kind: kind, Text: text, Span: diagnostics.Span{Start: start, End: l.pos()}})
}

var punct = map[byte]Kind{
	'(': LParen, ')': RParen, '{': LBrace, '}': RBrace,
	',': Comma, ';': Semicolon, '.': Dot,
}

// operators, longest first so that maximal munch picks "<=" over "<".
var operators = []string{
	"++", "--", "+=", "-=", "*=", "/=", "%=", "==", "!=", "<=", ">=", "&&", "||",
	"+", "-", "*", "/", "%", "<", ">", "!", "=",
}

func (l *lexer) run() {
	for {
		l.skipSpaceAndComments()
		start := l.pos()
		if l.off >= len(l.src) {
			l.emit(EOF, "", start)
			return
		}
		c := l.src[l.off]
		switch {
		case isIdentStart(c):
			l.ident(start)
		case isDigit(c):
			l.number(start)
		case c == '"':
			l.string(start)
		case c == '\'':
			l.advance()
			l.errorf(diagnostics.LexUnsupportedLiteral, start, "character literals are not supported yet").
				WithHelp("the 'char' type and its literal syntax are not yet specified")
		default:
			if k, ok := punct[c]; ok {
				l.advance()
				l.emit(k, string(c), start)
				continue
			}
			if op := l.matchOperator(); op != "" {
				for range op {
					l.advance()
				}
				kind := Op
				if op == "=" {
					kind = Assign
				}
				l.emit(kind, op, start)
				continue
			}
			r, _ := utf8.DecodeRune(l.src[l.off:])
			l.advance()
			l.errorf(diagnostics.LexUnexpectedChar, start, "unexpected character %q", r)
		}
	}
}

func (l *lexer) skipSpaceAndComments() {
	for l.off < len(l.src) {
		c := l.src[l.off]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			l.advance()
		case c == '/' && l.peek(1) == '/':
			for l.off < len(l.src) && l.src[l.off] != '\n' {
				l.advance()
			}
		case c == '/' && l.peek(1) == '*':
			start := l.pos()
			l.advance()
			l.advance()
			for {
				if l.off >= len(l.src) {
					l.errorf(diagnostics.LexUnterminatedBlock, start, "unterminated block comment")
					return
				}
				if l.src[l.off] == '*' && l.peek(1) == '/' {
					l.advance()
					l.advance()
					break
				}
				l.advance()
			}
		default:
			return
		}
	}
}

func (l *lexer) ident(start diagnostics.Pos) {
	for l.off < len(l.src) && (isIdentStart(l.src[l.off]) || isDigit(l.src[l.off])) {
		l.advance()
	}
	text := string(l.src[start.Offset:l.off])
	switch {
	case keywords[text]:
		l.emit(Keyword, text, start)
	case reserved[text]:
		l.emit(Reserved, text, start)
	default:
		l.emit(Ident, text, start)
	}
}

// number lexes a decimal integer literal. Leading zeros, floating-point
// literals, and suffixes are rejected because their semantics are undecided.
func (l *lexer) number(start diagnostics.Pos) {
	for l.off < len(l.src) && isDigit(l.src[l.off]) {
		l.advance()
	}
	text := string(l.src[start.Offset:l.off])
	if l.off < len(l.src) && l.src[l.off] == '.' && isDigit(l.peek(1)) {
		for l.off < len(l.src) && (isDigit(l.src[l.off]) || l.src[l.off] == '.') {
			l.advance()
		}
		l.errorf(diagnostics.LexUnsupportedLiteral, start, "floating-point literals are not supported yet").
			WithHelp("'float' literal syntax and widening rules are not yet specified")
		return
	}
	if l.off < len(l.src) && isIdentStart(l.src[l.off]) {
		for l.off < len(l.src) && (isIdentStart(l.src[l.off]) || isDigit(l.src[l.off])) {
			l.advance()
		}
		l.errorf(diagnostics.LexInvalidNumber, start, "invalid integer literal %q", string(l.src[start.Offset:l.off])).
			WithHelp("integer literal prefixes and suffixes are not yet specified")
		return
	}
	if len(text) > 1 && text[0] == '0' {
		l.errorf(diagnostics.LexInvalidNumber, start, "integer literal %q has a leading zero", text).
			WithHelp("leading zeros are rejected until octal/decimal semantics are decided; write the number without them")
		return
	}
	if _, err := strconv.ParseUint(text, 10, 64); err != nil {
		l.errorf(diagnostics.LexInvalidNumber, start, "integer literal %s is too large", text)
		return
	}
	l.emit(Int, text, start)
}

func (l *lexer) string(start diagnostics.Pos) {
	l.advance() // opening quote
	var b strings.Builder
	for {
		if l.off >= len(l.src) || l.src[l.off] == '\n' {
			l.errorf(diagnostics.LexUnterminatedString, start, "unterminated string literal")
			return
		}
		c := l.src[l.off]
		if c == '"' {
			l.advance()
			l.emit(String, b.String(), start)
			return
		}
		if c == '\\' {
			escStart := l.pos()
			l.advance()
			var e byte
			if l.off < len(l.src) {
				e = l.src[l.off]
			}
			switch e {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				if e != '\n' && e != 0 {
					l.advance()
				}
				l.errorf(diagnostics.LexInvalidEscape, escStart, "invalid escape sequence in string literal").
					WithHelp(`supported escapes are \n \t \r \\ \"`)
				continue
			}
			l.advance()
			continue
		}
		_, size := utf8.DecodeRune(l.src[l.off:])
		b.Write(l.src[l.off : l.off+size])
		l.advance()
	}
}

func (l *lexer) matchOperator() string {
	rest := l.src[l.off:]
	for _, op := range operators {
		if strings.HasPrefix(string(rest[:min(len(rest), 2)]), op) {
			return op
		}
	}
	return ""
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
