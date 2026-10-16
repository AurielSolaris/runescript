package lexer

import (
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

func TestUnicodeEscapes(t *testing.T) {
	cases := map[string]string{
		`"caf\u{E9}"`:         "café",
		`"\u{1F600}!"`:        "😀!",
		`"\u{41}\u{000041}"`:  "AA",
		`"\u{10FFFF}"`:        "\U0010FFFF",
		`"a\0b"`:              "a\x00b",
		`"it\'s"`:             "it's",
		`'\u{E9}'`:            "é",
		`'\0'`:                "\x00",
		`$"\u{E9}{x}\u{E8}"`:  "é",
		`"\u{e9}\u{fe0f}"`:    "é️",
		`"tab\tthen\u{20}sp"`: "tab\tthen sp",
	}
	for src, want := range cases {
		toks, diags := Lex([]byte(src))
		if len(diags) != 0 {
			t.Errorf("%s: %v", src, diags)
			continue
		}
		if toks[0].Text != want {
			t.Errorf("%s: got %q, want %q", src, toks[0].Text, want)
		}
	}
}

func TestBadUnicodeEscapes(t *testing.T) {
	cases := []struct{ src, pos string }{
		{"\"\\" + "u0041\"", "1:2"}, // braces required (backslash-u0041 is not an escape)
		{`"\u{}"`, "1:2"},           // no digits
		{`"\u{1234567}"`, "1:2"},    // 7 digits
		{`"\u{D800}"`, "1:2"},       // surrogate
		{`"\u{110000}"`, "1:2"},     // beyond U+10FFFF
		{`"\u{12G4}"`, "1:2"},       // not hex
		{`x = '\u{DFFF}'`, "1:6"},   // in a char too
		{`"\x41"`, "1:2"},           // byte escapes are not part of the language
	}
	for _, c := range cases {
		_, diags := Lex([]byte(c.src))
		if len(diags) == 0 {
			t.Errorf("%s: expected an error", c.src)
			continue
		}
		if d := diags[0]; d.Code != diagnostics.LexInvalidEscape || d.Span.Start.String() != c.pos {
			t.Errorf("%s: got %s at %s, want %s at %s", c.src, d.Code, d.Span.Start, diagnostics.LexInvalidEscape, c.pos)
		}
	}
}
