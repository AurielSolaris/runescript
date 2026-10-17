package stdlib

import (
	"strings"
	"unicode/utf8"
)

// decodeUTF8 turns bytes read from a file or standard input into a valid
// string. Invalid UTF-8 becomes U+FFFD following the WHATWG Encoding
// Standard (one U+FFFD per maximal invalid subsequence), which is what
// the JavaScript target's decoder does, so both targets read the same
// text. (Go's own conversions replace each invalid byte separately.)
func decodeUTF8(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	}
	var sb strings.Builder
	for i := 0; i < len(b); {
		c := b[i]
		if c < 0x80 {
			sb.WriteByte(c)
			i++
			continue
		}
		need, lower, upper := 0, byte(0x80), byte(0xBF)
		var r rune
		switch {
		case c >= 0xC2 && c <= 0xDF:
			need, r = 1, rune(c&0x1F)
		case c >= 0xE0 && c <= 0xEF:
			if c == 0xE0 {
				lower = 0xA0 // no overlong forms
			}
			if c == 0xED {
				upper = 0x9F // no surrogates
			}
			need, r = 2, rune(c&0x0F)
		case c >= 0xF0 && c <= 0xF4:
			if c == 0xF0 {
				lower = 0x90
			}
			if c == 0xF4 {
				upper = 0x8F // nothing above U+10FFFF
			}
			need, r = 3, rune(c&0x07)
		default:
			sb.WriteRune(utf8.RuneError)
			i++
			continue
		}
		j := i + 1
		for ; need > 0; need-- {
			if j >= len(b) || b[j] < lower || b[j] > upper {
				break
			}
			r = r<<6 | rune(b[j]&0x3F)
			lower, upper = 0x80, 0xBF
			j++
		}
		if need > 0 {
			// A broken sequence: one U+FFFD for the bytes so far; the byte
			// that broke it starts the next one.
			r = utf8.RuneError
		}
		sb.WriteRune(r)
		i = j
	}
	return sb.String()
}
