// Package srcenc turns the bytes of a source file into UTF-8 text.
//
// RuneScript source is UTF-8 (owner decision, v0.3: "try to convert, else
// reject"). Files in another Unicode encoding that can be recognised with
// certainty are converted: UTF-8 with a byte-order mark, UTF-16 and UTF-32
// with a byte-order mark, and UTF-16 without one (which Windows tools
// sometimes write; it shows as NUL bytes next to ASCII). Anything else is
// returned unchanged, and the lexer rejects it if it is not valid UTF-8.
// Legacy 8-bit encodings such as Windows-1252 are not guessed: any byte
// sequence is valid in them, so a guess could silently change a program.
package srcenc

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
	"unicode/utf8"
)

// Decode returns b as UTF-8 text. from names the encoding it was converted
// from ("UTF-16LE", "UTF-8 with a byte-order mark", ...), or is "" if b was
// returned unchanged.
func Decode(b []byte) (text []byte, from string) {
	switch {
	case bytes.HasPrefix(b, []byte{0xEF, 0xBB, 0xBF}):
		return b[3:], "UTF-8 with a byte-order mark"
	case bytes.HasPrefix(b, []byte{0xFF, 0xFE, 0, 0}):
		if t, ok := utf32(b[4:], binary.LittleEndian); ok {
			return t, "UTF-32LE"
		}
	case bytes.HasPrefix(b, []byte{0, 0, 0xFE, 0xFF}):
		if t, ok := utf32(b[4:], binary.BigEndian); ok {
			return t, "UTF-32BE"
		}
	case bytes.HasPrefix(b, []byte{0xFF, 0xFE}):
		if t, ok := utf16Text(b[2:], binary.LittleEndian); ok {
			return t, "UTF-16LE"
		}
	case bytes.HasPrefix(b, []byte{0xFE, 0xFF}):
		if t, ok := utf16Text(b[2:], binary.BigEndian); ok {
			return t, "UTF-16BE"
		}
	case bytes.IndexByte(b, 0) >= 0:
		// No byte-order mark, but NUL bytes, which RuneScript source never
		// contains: UTF-16 if the NULs fall on one side of every ASCII
		// character.
		if order, ok := guessUTF16(b); ok {
			if t, ok := utf16Text(b, order); ok {
				if order == binary.ByteOrder(binary.LittleEndian) {
					return t, "UTF-16LE"
				}
				return t, "UTF-16BE"
			}
		}
	}
	return b, ""
}

// utf16Text decodes UTF-16 in the given byte order. It fails on an odd
// length, an unpaired surrogate, or a NUL character.
func utf16Text(b []byte, order binary.ByteOrder) ([]byte, bool) {
	if len(b)%2 != 0 {
		return nil, false
	}
	units := make([]uint16, len(b)/2)
	for i := range units {
		units[i] = order.Uint16(b[2*i:])
	}
	var out []byte
	for i := 0; i < len(units); i++ {
		u := units[i]
		switch {
		case u == 0:
			return nil, false
		case utf16.IsSurrogate(rune(u)):
			if u >= 0xDC00 || i+1 == len(units) {
				return nil, false
			}
			r := utf16.DecodeRune(rune(u), rune(units[i+1]))
			if r == utf8.RuneError {
				return nil, false
			}
			out = utf8.AppendRune(out, r)
			i++
		default:
			out = utf8.AppendRune(out, rune(u))
		}
	}
	return out, true
}

// utf32 decodes UTF-32 in the given byte order.
func utf32(b []byte, order binary.ByteOrder) ([]byte, bool) {
	if len(b)%4 != 0 {
		return nil, false
	}
	var out []byte
	for i := 0; i < len(b); i += 4 {
		r := rune(order.Uint32(b[i:]))
		if r == 0 || !utf8.ValidRune(r) {
			return nil, false
		}
		out = utf8.AppendRune(out, r)
	}
	return out, true
}

// guessUTF16 reports whether b looks like UTF-16 without a byte-order mark,
// and in which order: in its first 512 bytes, at least half the code units
// have a NUL high byte (ASCII characters) and only a few have a NUL low
// byte (e.g. the surrogate 0xDE00). Decoding then checks the whole text.
func guessUTF16(b []byte) (binary.ByteOrder, bool) {
	n := len(b)
	if n > 512 {
		n = 512
	}
	n -= n % 2
	if n < 4 {
		return nil, false
	}
	var evenNUL, oddNUL int
	for i := 0; i < n; i += 2 {
		if b[i] == 0 {
			evenNUL++
		}
		if b[i+1] == 0 {
			oddNUL++
		}
	}
	pairs := n / 2
	switch {
	case 2*oddNUL >= pairs && 8*evenNUL <= oddNUL:
		return binary.LittleEndian, true // "p\x00a\x00..."
	case 2*evenNUL >= pairs && 8*oddNUL <= evenNUL:
		return binary.BigEndian, true // "\x00p\x00a..."
	}
	return nil, false
}

// FirstInvalid returns the offset of the first byte of b that is not part
// of valid UTF-8, or -1 if b is valid.
func FirstInvalid(b []byte) int {
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && size <= 1 {
			return i
		}
		i += size
	}
	return -1
}
