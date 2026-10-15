// Package lexer converts RuneScript source text into tokens.
package lexer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/srcenc"
)

// Kind classifies a token.
type Kind int

const (
	EOF Kind = iota
	Ident
	Int    // integer literal; Token.Text holds the digits
	Float  // floating-point literal; Token.Text holds it as written
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
	EOF: "end of file", Ident: "identifier", Int: "integer literal", Float: "floating-point literal",
	String: "string literal", Keyword: "keyword", Reserved: "reserved word",
	LParen: "'('", RParen: "')'", LBrace: "'{'", RBrace: "'}'", Comma: "','",
	Semicolon: "';'", Dot: "'.'", Assign: "'='", Op: "operator",
}

func (k Kind) String() string { return kindNames[k] }

// keywords implemented by the current milestone.
var keywords = map[string]bool{
	"package": true, "use": true, "return": true,
	"if": true, "else": true, "while": true, "for": true, "break": true, "continue": true,
	"true": true, "false": true, "bool": true,
	"public": true, "protected": true, "private": true,
	"int": true, "sint": true, "uint": true, "string": true, "void": true,
	"float": true, "float32": true, "float64": true,
	"try": true, "catch": true, "finally": true, "throw": true, "new": true,
}

// reserved words named by the specification for features that are not yet
// implemented. Reserving them now keeps today's programs valid later.
var reserved = map[string]bool{
	"class": true, "struct": true,
	"char": true, "array": true, "matrix": true,
	"int8": true, "int16": true, "int32": true, "int64": true,
	"uint8": true, "uint16": true, "uint32": true, "uint64": true,
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
	file  *diagnostics.Source
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
	return LexSource(diagnostics.NewSource("", src))
}

// LexSource tokenizes a source file; every position records the file.
// Source text must be UTF-8 (other Unicode encodings are converted when
// files are read, see internal/srcenc); a file that is not is rejected as a
// whole, at its first invalid byte.
func LexSource(file *diagnostics.Source) ([]Token, []*diagnostics.Diagnostic) {
	l := &lexer{file: file, src: file.Text, line: 1, col: 1}
	if bad := srcenc.FirstInvalid(file.Text); bad >= 0 {
		for l.off < bad {
			l.advance()
		}
		l.errorf(diagnostics.LexInvalidUTF8, l.pos(), "the file is not valid UTF-8: byte 0x%02X cannot appear here", file.Text[bad]).
			WithHelp("save the file as UTF-8; UTF-16 and UTF-32 files are also read")
		l.emit(EOF, "", l.pos())
		return l.toks, l.diags
	}
	l.run()
	return l.toks, l.diags
}

func (l *lexer) pos() diagnostics.Pos {
	return diagnostics.Pos{File: l.file, Offset: l.off, Line: l.line, Col: l.col}
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
	"=>", "++", "--", "+=", "-=", "*=", "/=", "%=", "==", "!=", "<=", ">=", "&&", "||",
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

// number lexes a decimal integer literal (`42`) or a floating-point
// literal (`64.3`, `1e10`, `2.5e-3`): digits, then optionally '.' and
// digits, then optionally an exponent. Both sides of the '.' need digits.
// Leading zeros (`007`, but not `0.5`) and suffixes are rejected because
// their semantics are undecided.
func (l *lexer) number(start diagnostics.Pos) {
	l.digits()
	isFloat := false
	if l.off < len(l.src) && l.src[l.off] == '.' && isDigit(l.peek(1)) {
		isFloat = true
		l.advance()
		l.digits()
	}
	if l.off < len(l.src) && (l.src[l.off] == 'e' || l.src[l.off] == 'E') {
		k := 1
		if c := l.peek(1); c == '+' || c == '-' {
			k = 2
		}
		if isDigit(l.peek(k)) {
			isFloat = true
			for ; k > 0; k-- {
				l.advance()
			}
			l.digits()
		}
	}
	text := string(l.src[start.Offset:l.off])
	if l.off < len(l.src) && (isIdentStart(l.src[l.off]) || l.src[l.off] == '.' && isDigit(l.peek(1))) {
		for l.off < len(l.src) && (isIdentStart(l.src[l.off]) || isDigit(l.src[l.off]) || l.src[l.off] == '.') {
			l.advance()
		}
		l.errorf(diagnostics.LexInvalidNumber, start, "invalid number literal %q", string(l.src[start.Offset:l.off])).
			WithHelp("number literal prefixes and suffixes are not yet specified")
		return
	}
	if len(text) > 1 && text[0] == '0' && isDigit(text[1]) {
		l.errorf(diagnostics.LexInvalidNumber, start, "number literal %q has a leading zero", text).
			WithHelp("leading zeros are rejected until octal/decimal semantics are decided; write the number without them")
		return
	}
	if isFloat {
		v, err := strconv.ParseFloat(text, 64)
		switch {
		case err != nil:
			l.errorf(diagnostics.LexInvalidNumber, start, "floating-point literal %s is too large", text).
				WithHelp("the largest float is about 1.7976931348623157e308")
			return
		case v == 0 && strings.Trim(strings.SplitN(strings.ToLower(text), "e", 2)[0], "0.") != "":
			l.errorf(diagnostics.LexInvalidNumber, start, "floating-point literal %s is too small and would round to 0", text)
			return
		}
		l.emit(Float, text, start)
		return
	}
	if _, err := strconv.ParseUint(text, 10, 64); err != nil {
		l.errorf(diagnostics.LexInvalidNumber, start, "integer literal %s is too large", text)
		return
	}
	l.emit(Int, text, start)
}

func (l *lexer) digits() {
	for l.off < len(l.src) && isDigit(l.src[l.off]) {
		l.advance()
	}
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

// IsName reports whether s is an identifier that is not a keyword or a
// reserved word, i.e. a valid package name segment or function name.
func IsName(s string) bool {
	if s == "" || !isIdentStart(s[0]) || keywords[s] || reserved[s] {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isIdentStart(s[i]) && !isDigit(s[i]) {
			return false
		}
	}
	return true
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
