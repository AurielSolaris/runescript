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
	// An interpolated string $"a{x}b{y}c" is StrHead "a", the tokens of x,
	// StrMid "b", the tokens of y, and StrTail "c". One without holes is a
	// String.
	StrHead
	StrMid
	StrTail
	Char // character literal; Token.Text holds the one decoded character
	Keyword
	Reserved // reserved for a future feature; not usable as an identifier
	LParen
	RParen
	LBrace
	RBrace
	LBracket
	RBracket
	Comma
	Semicolon
	Dot
	Assign
	Op // operator; Token.Text holds its spelling
)

var kindNames = map[Kind]string{
	EOF: "end of file", Ident: "identifier", Int: "integer literal", Float: "floating-point literal",
	String: "string literal", StrHead: "interpolated string", StrMid: "'}'", StrTail: "'}'", Char: "character literal", Keyword: "keyword", Reserved: "reserved word",
	LParen: "'('", RParen: "')'", LBrace: "'{'", RBrace: "'}'", LBracket: "'['", RBracket: "']'", Comma: "','",
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
	"float": true, "float32": true, "float64": true, "char": true,
	"int8": true, "int16": true, "int32": true, "int64": true,
	"uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"try": true, "catch": true, "finally": true, "throw": true, "new": true,
	"extern": true,
}

// reserved words named by the specification for features that are not yet
// implemented. Reserving them now keeps today's programs valid later.
var reserved = map[string]bool{
	"class": true, "struct": true,
	"array": true, "matrix": true,
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
	'(': LParen, ')': RParen, '{': LBrace, '}': RBrace, '[': LBracket, ']': RBracket,
	',': Comma, ';': Semicolon, '.': Dot,
}

// operators, longest first so that maximal munch picks "<=" over "<".
var operators = []string{
	"<<=", ">>=",
	"=>", "++", "--", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "==", "!=", "<=", ">=", "&&", "||", "<<", ">>",
	"+", "-", "*", "/", "%", "&", "|", "^", "~", "<", ">", "!", "=",
}

func (l *lexer) run() {
	for {
		l.skipSpaceAndComments()
		if l.off >= len(l.src) {
			l.emit(EOF, "", l.pos())
			return
		}
		l.token()
		if l.atTemplate() {
			l.skipSpaceAndComments()
			start := l.pos()
			l.advance()
			l.emit(LBrace, "{", start)
			l.markup()
		}
	}
}

// token lexes one token; the input is not at a space or a comment.
func (l *lexer) token() {
	start := l.pos()
	c := l.src[l.off]
	switch {
	case isIdentStart(c):
		l.ident(start)
	case isDigit(c):
		l.number(start)
	case c == '"':
		l.string(start)
	case c == '$' && l.peek(1) == '"':
		l.interpString(start)
	case c == '\'':
		l.char(start)
	default:
		if k, ok := punct[c]; ok {
			l.advance()
			l.emit(k, string(c), start)
			return
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
			return
		}
		r, _ := utf8.DecodeRune(l.src[l.off:])
		l.advance()
		l.errorf(diagnostics.LexUnexpectedChar, start, "unexpected character %q", r)
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

// number lexes a number literal (OQ-NUM-06): a decimal integer (`42`),
// a hex, binary, or octal integer (`0xFF`, `0b1010`, `0o17`), or a decimal
// floating-point literal (`64.3`, `1e10`, `2.5e-3`). An underscore may
// separate two digits: `1_000_000`, `0xFF_FF`, `1_000.5`. Leading zeros
// (`007`, but not `0.5`), upper-case prefixes, and suffixes are rejected.
func (l *lexer) number(start diagnostics.Pos) {
	if l.src[l.off] == '0' && strings.IndexByte("xXbBoO", l.peek(1)) >= 0 {
		l.prefixed(start)
		return
	}
	l.digitRun(isDigit)
	isFloat := false
	if l.off < len(l.src) && l.src[l.off] == '.' && isDigit(l.peek(1)) {
		isFloat = true
		l.advance()
		l.digitRun(isDigit)
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
			l.digitRun(isDigit)
		}
	}
	text := string(l.src[start.Offset:l.off])
	if l.trailing(start) || !l.underscores(start, text, isDigit) {
		return
	}
	digits := strings.ReplaceAll(text, "_", "")
	if len(digits) > 1 && digits[0] == '0' && isDigit(digits[1]) {
		l.errorf(diagnostics.LexInvalidNumber, start, "number literal %q has a leading zero", text).
			WithHelp("leading zeros are rejected so that 017 can't be mistaken for octal; write 17, or 0o17 for octal")
		return
	}
	if isFloat {
		v, err := strconv.ParseFloat(digits, 64)
		switch {
		case err != nil:
			l.errorf(diagnostics.LexInvalidNumber, start, "floating-point literal %s is too large", text).
				WithHelp("the largest float is about 1.7976931348623157e308")
			return
		case v == 0 && strings.Trim(strings.SplitN(strings.ToLower(digits), "e", 2)[0], "0.") != "":
			l.errorf(diagnostics.LexInvalidNumber, start, "floating-point literal %s is too small and would round to 0", text)
			return
		}
		l.emit(Float, text, start)
		return
	}
	if _, err := ParseInt(text); err != nil {
		l.errorf(diagnostics.LexInvalidNumber, start, "integer literal %s is too large", text).
			WithHelp("the largest integer literal is 18446744073709551615 (0xFFFF_FFFF_FFFF_FFFF)")
		return
	}
	l.emit(Int, text, start)
}

// prefixed lexes a hex, binary, or octal integer literal.
func (l *lexer) prefixed(start diagnostics.Pos) {
	p := l.peek(1)
	l.advance()
	l.advance()
	var digit func(byte) bool
	switch p {
	case 'x', 'X':
		digit = isHex
	case 'b', 'B':
		digit = func(c byte) bool { return c == '0' || c == '1' }
	default:
		digit = func(c byte) bool { return c >= '0' && c <= '7' }
	}
	l.digitRun(func(c byte) bool { return digit(c) || isDigit(c) || isHex(c) })
	text := string(l.src[start.Offset:l.off])
	if l.trailing(start) {
		return
	}
	switch {
	case p == 'X' || p == 'B' || p == 'O':
		l.errorf(diagnostics.LexInvalidNumber, start, "number literal %q has an upper-case prefix", text).
			WithHelp("write the prefix in lower case: 0x, 0b, or 0o")
		return
	case len(text) == 2:
		l.errorf(diagnostics.LexInvalidNumber, start, "number literal %q has no digits after its prefix", text)
		return
	}
	for _, c := range []byte(text[2:]) {
		if c != '_' && !digit(c) {
			l.errorf(diagnostics.LexInvalidNumber, start, "%q is not a digit of %s literal %q", c, map[byte]string{'x': "a hex", 'b': "a binary", 'o': "an octal"}[p], text)
			return
		}
	}
	if !l.underscores(start, text[2:], digit) {
		return
	}
	if _, err := ParseInt(text); err != nil {
		l.errorf(diagnostics.LexInvalidNumber, start, "integer literal %s is too large", text).
			WithHelp("the largest integer literal is 18446744073709551615 (0xFFFF_FFFF_FFFF_FFFF)")
		return
	}
	l.emit(Int, text, start)
}

// digitRun consumes digits accepted by digit, and underscores.
func (l *lexer) digitRun(digit func(byte) bool) {
	for l.off < len(l.src) && (digit(l.src[l.off]) || l.src[l.off] == '_') {
		l.advance()
	}
}

// trailing reports, and consumes, letters or digits stuck to the end of a
// number literal, such as a suffix (`12abc`, `1.5f`).
func (l *lexer) trailing(start diagnostics.Pos) bool {
	if l.off >= len(l.src) || !(isIdentStart(l.src[l.off]) || l.src[l.off] == '.' && (isDigit(l.peek(1)) || l.peek(1) == '_')) {
		return false
	}
	for l.off < len(l.src) && (isIdentStart(l.src[l.off]) || isDigit(l.src[l.off]) || l.src[l.off] == '.') {
		l.advance()
	}
	l.errorf(diagnostics.LexInvalidNumber, start, "invalid number literal %q", string(l.src[start.Offset:l.off])).
		WithHelp("number literals have no suffixes; hex, binary, and octal are written 0xFF, 0b1010, 0o17")
	return true
}

// underscores checks that every '_' in text stands between two digits.
func (l *lexer) underscores(start diagnostics.Pos, text string, digit func(byte) bool) bool {
	for i := 0; i < len(text); i++ {
		if text[i] == '_' && (i == 0 || i == len(text)-1 || !digit(text[i-1]) || !digit(text[i+1])) {
			l.errorf(diagnostics.LexInvalidNumber, start, "misplaced '_' in number literal %q", string(l.src[start.Offset:l.off])).
				WithHelp("an underscore can only separate two digits, as in 1_000_000")
			return false
		}
	}
	return true
}

// ParseInt returns the value of an integer literal as the lexer accepted
// it: decimal, or 0x/0b/0o, with underscores.
func ParseInt(text string) (uint64, error) {
	text = strings.ReplaceAll(text, "_", "")
	base := 10
	if len(text) > 1 && text[0] == '0' {
		switch text[1] {
		case 'x':
			base = 16
		case 'b':
			base = 2
		case 'o':
			base = 8
		}
		if base != 10 {
			text = text[2:]
		}
	}
	return strconv.ParseUint(text, base, 64)
}

// ParseFloat returns the value of a floating-point literal as the lexer
// accepted it.
func ParseFloat(text string) (float64, error) {
	return strconv.ParseFloat(strings.ReplaceAll(text, "_", ""), 64)
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
			l.escape(&b)
			continue
		}
		_, size := utf8.DecodeRune(l.src[l.off:])
		b.Write(l.src[l.off : l.off+size])
		l.advance()
	}
}

// char lexes a character literal: one character or escape between single
// quotes (OQ-CHAR-01). \' is a quote; the other escapes are those of
// strings.
func (l *lexer) char(start diagnostics.Pos) {
	l.advance() // opening quote
	var b strings.Builder
	for {
		if l.off >= len(l.src) || l.src[l.off] == '\n' {
			l.errorf(diagnostics.LexUnterminatedString, start, "unterminated character literal")
			return
		}
		c := l.src[l.off]
		switch {
		case c == '\'':
			l.advance()
			if n := utf8.RuneCountInString(b.String()); n != 1 {
				l.errorf(diagnostics.LexBadChar, start, "a character literal holds exactly one character, found %d", n).
					WithHelp(`use a string, "…", for text`)
				return
			}
			l.emit(Char, b.String(), start)
			return
		case c == '\\' && l.peek(1) == '\'':
			b.WriteByte('\'')
			l.advance()
			l.advance()
		case c == '\\':
			l.escape(&b)
		default:
			_, size := utf8.DecodeRune(l.src[l.off:])
			b.Write(l.src[l.off : l.off+size])
			l.advance()
		}
	}
}

// escape decodes the escape sequence at the input into b.
func (l *lexer) escape(b *strings.Builder) {
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
	case '\'':
		b.WriteByte('\'')
	case '0':
		b.WriteByte(0)
	case 'u':
		l.advance()
		l.unicodeEscape(b, escStart)
		return
	default:
		if e != '\n' && e != 0 {
			l.advance()
		}
		l.errorf(diagnostics.LexInvalidEscape, escStart, "invalid escape sequence in string literal").
			WithHelp(`supported escapes are \n \t \r \0 \\ \" \' and \u{1F600}`)
		return
	}
	l.advance()
}

// unicodeEscape decodes the rest of \u{X…}: 1 to 6 hex digits naming a
// Unicode scalar value (OQ-STR-02). The input is just after the 'u'.
func (l *lexer) unicodeEscape(b *strings.Builder, escStart diagnostics.Pos) {
	bad := func(msg string) {
		l.errorf(diagnostics.LexInvalidEscape, escStart, "%s", msg).
			WithHelp(`write a Unicode escape as \u{...} with 1 to 6 hex digits, e.g. \u{E9} or \u{1F600}`)
	}
	if l.off >= len(l.src) || l.src[l.off] != '{' {
		bad(`\u must be followed by {`)
		return
	}
	l.advance()
	var v int64
	n := 0
	for l.off < len(l.src) && isHex(l.src[l.off]) {
		v = v*16 + hexVal(l.src[l.off])
		n++
		l.advance()
		if n > 6 {
			for l.off < len(l.src) && isHex(l.src[l.off]) {
				l.advance()
			}
			bad(`a \u{...} escape has at most 6 hex digits`)
			return
		}
	}
	if n == 0 || l.off >= len(l.src) || l.src[l.off] != '}' {
		bad(`a \u{...} escape needs 1 to 6 hex digits and a closing }`)
		return
	}
	l.advance()
	switch {
	case v >= 0xD800 && v <= 0xDFFF:
		bad(fmt.Sprintf(`\u{%X} is a surrogate, which is not a character`, v))
	case v > 0x10FFFF:
		bad(fmt.Sprintf(`\u{%X} is beyond U+10FFFF, the last Unicode code point`, v))
	default:
		b.WriteRune(rune(v))
	}
}

func isHex(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func hexVal(c byte) int64 {
	switch {
	case c >= 'a':
		return int64(c-'a') + 10
	case c >= 'A':
		return int64(c-'A') + 10
	}
	return int64(c - '0')
}

// interpString lexes $"…{x}…" (OQ-STR-01): the text pieces become StrHead,
// StrMid, and StrTail tokens, with each hole's tokens between them. In the
// text, {{ and }} are literal braces, and escapes work as in strings.
func (l *lexer) interpString(start diagnostics.Pos) {
	l.advance() // $
	l.advance() // "
	kind, piece := StrHead, start
	for {
		text, closed, ok := l.interpText(start)
		if !ok {
			return
		}
		if closed {
			if kind == StrHead {
				kind = String // no holes: an ordinary string
			} else {
				kind = StrTail
			}
			l.emit(kind, text, piece)
			return
		}
		l.emit(kind, text, piece)
		if !l.hole(start) {
			return
		}
		piece = l.pos()
		l.advance() // }
		kind = StrMid
	}
}

// interpText reads text up to and including the '{' that opens a hole or
// the closing quote; closed reports which. ok is false after an error that
// ends the string.
func (l *lexer) interpText(start diagnostics.Pos) (text string, closed, ok bool) {
	var b strings.Builder
	for {
		if l.off >= len(l.src) || l.src[l.off] == '\n' {
			l.errorf(diagnostics.LexUnterminatedString, start, "unterminated string literal")
			return "", false, false
		}
		switch c := l.src[l.off]; {
		case c == '"':
			l.advance()
			return b.String(), true, true
		case c == '{' && l.peek(1) == '{', c == '}' && l.peek(1) == '}':
			b.WriteByte(c)
			l.advance()
			l.advance()
		case c == '{':
			l.advance()
			return b.String(), false, true
		case c == '}':
			brace := l.pos()
			l.advance()
			l.errorf(diagnostics.LexInvalidEscape, brace, "single '}' in an interpolated string").
				WithHelp("write }} for a literal '}'")
		case c == '\\':
			l.escape(&b)
		default:
			_, size := utf8.DecodeRune(l.src[l.off:])
			b.Write(l.src[l.off : l.off+size])
			l.advance()
		}
	}
}

// hole lexes the tokens of an interpolation hole, up to its closing '}',
// which is not consumed. Braces inside the hole nest.
func (l *lexer) hole(start diagnostics.Pos) bool {
	depth := 0
	for {
		l.skipSpaceAndComments()
		if l.off >= len(l.src) {
			l.errorf(diagnostics.LexUnterminatedString, start, "unterminated interpolated string: a '{' has no closing '}'")
			return false
		}
		switch l.src[l.off] {
		case '}':
			if depth == 0 {
				return true
			}
			depth--
		case '{':
			depth++
		}
		l.token()
	}
}

func (l *lexer) matchOperator() string {
	rest := l.src[l.off:]
	for _, op := range operators {
		if strings.HasPrefix(string(rest[:min(len(rest), 3)]), op) {
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
