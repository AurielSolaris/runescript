package lexer

import (
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

func TestTokensAndPositions(t *testing.T) {
	src := "package main;\nint x = -42; // comment\n/* block\n */ s = \"a\\tb\" <= != ++ += && || !"
	toks, diags := Lex([]byte(src))
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	want := []struct {
		kind      Kind
		text      string
		line, col int
	}{
		{Keyword, "package", 1, 1}, {Ident, "main", 1, 9}, {Semicolon, ";", 1, 13},
		{Keyword, "int", 2, 1}, {Ident, "x", 2, 5}, {Assign, "=", 2, 7}, {Op, "-", 2, 9},
		{Int, "42", 2, 10}, {Semicolon, ";", 2, 12},
		{Ident, "s", 4, 5}, {Assign, "=", 4, 7}, {String, "a\tb", 4, 9},
		{Op, "<=", 4, 16}, {Op, "!=", 4, 19}, {Op, "++", 4, 22}, {Op, "+=", 4, 25},
		{Op, "&&", 4, 28}, {Op, "||", 4, 31}, {Op, "!", 4, 34},
		{EOF, "", 4, 35},
	}
	if len(toks) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(toks), len(want), toks)
	}
	for i, w := range want {
		g := toks[i]
		if g.Kind != w.kind || g.Text != w.text || g.Span.Start.Line != w.line || g.Span.Start.Col != w.col {
			t.Errorf("token %d: got %v %q at %s, want %v %q at %d:%d",
				i, g.Kind, g.Text, g.Span.Start, w.kind, w.text, w.line, w.col)
		}
	}
}

func TestKeywordClasses(t *testing.T) {
	cases := map[string]Kind{
		"if": Keyword, "while": Keyword, "bool": Keyword, "true": Keyword,
		"use": Keyword, "class": Reserved, "float": Keyword, "float32": Keyword, "try": Keyword, "new": Keyword,
		"char": Reserved, "int64": Reserved, "matrix": Reserved, "Error": Ident, "as": Ident,
		"trace": Ident, "function": Ident, "Main": Ident, "_x1": Ident,
	}
	for word, kind := range cases {
		toks, _ := Lex([]byte(word))
		if toks[0].Kind != kind {
			t.Errorf("%q: got %v, want %v", word, toks[0].Kind, kind)
		}
	}
}

func TestColumnsCountCodePoints(t *testing.T) {
	toks, diags := Lex([]byte("\"héllo\" x"))
	if len(diags) != 0 {
		t.Fatal(diags)
	}
	if toks[0].Text != "héllo" || toks[1].Span.Start.Col != 9 {
		t.Errorf("got %q and x at col %d", toks[0].Text, toks[1].Span.Start.Col)
	}
}

func TestLexErrors(t *testing.T) {
	cases := []struct {
		src  string
		code diagnostics.Code
		pos  string
	}{
		{`"abc`, diagnostics.LexUnterminatedString, "1:1"},
		{"x = \"ab\ncd\"", diagnostics.LexUnterminatedString, "1:5"},
		{`"a\qb"`, diagnostics.LexInvalidEscape, "1:3"},
		{"/* never closed", diagnostics.LexUnterminatedBlock, "1:1"},
		{"007", diagnostics.LexInvalidNumber, "1:1"},
		{"12abc", diagnostics.LexInvalidNumber, "1:1"},
		{"99999999999999999999999", diagnostics.LexInvalidNumber, "1:1"},
		{"  1.5x", diagnostics.LexInvalidNumber, "1:3"},
		{"1.2.3", diagnostics.LexInvalidNumber, "1:1"},
		{"00.5", diagnostics.LexInvalidNumber, "1:1"},
		{"1e999", diagnostics.LexInvalidNumber, "1:1"},
		{"1e-999", diagnostics.LexInvalidNumber, "1:1"},
		{"'a'", diagnostics.LexUnsupportedLiteral, "1:1"},
		{"x # y", diagnostics.LexUnexpectedChar, "1:3"},
		{"a & b", diagnostics.LexUnexpectedChar, "1:3"},
	}
	for _, c := range cases {
		_, diags := Lex([]byte(c.src))
		if len(diags) == 0 {
			t.Errorf("%q: expected %s, got no diagnostics", c.src, c.code)
			continue
		}
		if d := diags[0]; d.Code != c.code || d.Span.Start.String() != c.pos {
			t.Errorf("%q: got %s at %s, want %s at %s", c.src, d.Code, d.Span.Start, c.code, c.pos)
		}
	}
}

func TestFloatLiterals(t *testing.T) {
	toks, diags := Lex([]byte("64.3 1e10 2.5E-3 0.5 0e0 1e+2 7"))
	if len(diags) != 0 {
		t.Fatal(diags)
	}
	want := []Kind{Float, Float, Float, Float, Float, Float, Int, EOF}
	for i, tok := range toks {
		if tok.Kind != want[i] {
			t.Errorf("token %d %q: kind %v, want %v", i, tok.Text, tok.Kind, want[i])
		}
	}
	if toks[2].Text != "2.5E-3" {
		t.Errorf("text %q", toks[2].Text)
	}
}

func TestArrowAndMaximalMunch(t *testing.T) {
	toks, diags := Lex([]byte("(int a) => a==>b"))
	if len(diags) != 0 {
		t.Fatal(diags)
	}
	var got []string
	for _, tok := range toks[:len(toks)-1] {
		got = append(got, tok.Text)
	}
	want := "( int a ) => a == > b"
	if s := strings.Join(got, " "); s != want {
		t.Errorf("got %q, want %q", s, want)
	}
}

func TestZeroIsAValidLiteral(t *testing.T) {
	toks, diags := Lex([]byte("0"))
	if len(diags) != 0 || toks[0].Kind != Int || toks[0].Text != "0" {
		t.Errorf("got %v %v", toks, diags)
	}
}

// A file that is not UTF-8 is rejected at its first invalid byte, with the
// column counted in code points.
func TestInvalidUTF8(t *testing.T) {
	toks, diags := Lex([]byte("package p;\n// é then \xe9 (Windows-1252)\n"))
	if len(diags) != 1 || diags[0].Code != diagnostics.LexInvalidUTF8 || diags[0].Span.Start.String() != "2:11" {
		t.Fatalf("diagnostics: %v", diags)
	}
	if len(toks) != 1 || toks[0].Kind != EOF {
		t.Errorf("tokens: %v", toks)
	}
}
