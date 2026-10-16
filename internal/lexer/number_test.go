package lexer

import (
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

func TestIntegerLiteralForms(t *testing.T) {
	cases := map[string]uint64{
		"0":                     0,
		"42":                    42,
		"1_000_000":             1000000,
		"0xFF":                  255,
		"0xff_ff":               65535,
		"0b1010":                10,
		"0b1111_0000":           240,
		"0o17":                  15,
		"0o7_7":                 63,
		"0xFFFF_FFFF_FFFF_FFFF": 18446744073709551615,
		"18446744073709551615":  18446744073709551615,
	}
	for src, want := range cases {
		toks, diags := Lex([]byte(src))
		if len(diags) != 0 {
			t.Errorf("%s: %v", src, diags)
			continue
		}
		if toks[0].Kind != Int || toks[0].Text != src {
			t.Errorf("%s: got %v %q", src, toks[0].Kind, toks[0].Text)
			continue
		}
		if v, err := ParseInt(toks[0].Text); err != nil || v != want {
			t.Errorf("%s: ParseInt = %d, %v; want %d", src, v, err, want)
		}
	}
}

func TestFloatLiteralSeparators(t *testing.T) {
	cases := map[string]float64{"1_000.5": 1000.5, "3.141_592": 3.141592, "1e1_0": 1e10, "6.022_140e2_3": 6.02214e23}
	for src, want := range cases {
		toks, diags := Lex([]byte(src))
		if len(diags) != 0 || toks[0].Kind != Float {
			t.Errorf("%s: %v %v", src, toks[0].Kind, diags)
			continue
		}
		if v, err := ParseFloat(toks[0].Text); err != nil || v != want {
			t.Errorf("%s: ParseFloat = %v, %v; want %v", src, v, err, want)
		}
	}
}

func TestBadNumberLiterals(t *testing.T) {
	cases := []string{
		"0x", "0xG1", "0b102", "0o8", "0X1F", "0B1", "0O7", // prefixes
		"1__000", "1_", "_1", "1_.5", "1._5", "0x_FF", "0xFF_", "1e_5", // underscores ("_1" is a name, see below)
		"0_7", "00x1", // leading zero
		"0x1p3", "12abc", "0xFFu", // suffixes and hex floats
		"0x1_0000_0000_0000_0000", "18446744073709551616", // too large
	}
	for _, src := range cases {
		toks, diags := Lex([]byte(src))
		if src == "_1" {
			if toks[0].Kind != Ident {
				t.Errorf("_1: got %v, want an identifier", toks[0].Kind)
			}
			continue
		}
		if len(diags) == 0 {
			t.Errorf("%s: expected ASX-LEX-005, got tokens %v", src, toks)
			continue
		}
		if d := diags[0]; d.Code != diagnostics.LexInvalidNumber || d.Span.Start.String() != "1:1" {
			t.Errorf("%s: got %s at %s", src, d.Code, d.Span.Start)
		}
	}
}
