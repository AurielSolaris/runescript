package srcenc

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

const program = "package p;\n// € and 😀\npublic void main() { trace(\"é\"); }\n"

func encode16(s string, order binary.AppendByteOrder, bom bool) []byte {
	var out []byte
	if bom {
		out = order.AppendUint16(out, 0xFEFF)
	}
	for _, u := range utf16.Encode([]rune(s)) {
		out = order.AppendUint16(out, u)
	}
	return out
}

func encode32(s string, order binary.AppendByteOrder) []byte {
	out := order.AppendUint32(nil, 0xFEFF)
	for _, r := range s {
		out = order.AppendUint32(out, uint32(r))
	}
	return out
}

func TestDecodeConverts(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		from string
	}{
		{"UTF-8", []byte(program), ""},
		{"UTF-8 BOM", append([]byte{0xEF, 0xBB, 0xBF}, program...), "UTF-8 with a byte-order mark"},
		{"UTF-16LE BOM", encode16(program, binary.LittleEndian, true), "UTF-16LE"},
		{"UTF-16BE BOM", encode16(program, binary.BigEndian, true), "UTF-16BE"},
		{"UTF-16LE no BOM", encode16(program, binary.LittleEndian, false), "UTF-16LE"},
		{"UTF-16BE no BOM", encode16(program, binary.BigEndian, false), "UTF-16BE"},
		{"UTF-32LE BOM", encode32(program, binary.LittleEndian), "UTF-32LE"},
		{"UTF-32BE BOM", encode32(program, binary.BigEndian), "UTF-32BE"},
	}
	for _, c := range cases {
		text, from := Decode(c.in)
		if string(text) != program || from != c.from {
			t.Errorf("%s: got %q from %q, want the program from %q", c.name, text, from, c.from)
		}
	}
}

// Bytes that are not recognisably Unicode are returned unchanged, for the
// lexer to reject.
func TestDecodeLeavesOthersAlone(t *testing.T) {
	for name, in := range map[string][]byte{
		"Windows-1252":         []byte("package p; // caf\xe9\n"),
		"odd-length UTF-16":    append(encode16("package p;", binary.LittleEndian, true), 'x'),
		"lone surrogate":       append(encode16("package p;", binary.LittleEndian, true), 0x00, 0xD8),
		"NUL in UTF-8 text":    []byte("package p;\x00 int\n"),
		"invalid UTF-32 rune":  {0xFF, 0xFE, 0, 0, 0, 0, 0x11, 0},
		"short NUL-ish header": {'p', 0},
	} {
		if text, from := Decode(in); from != "" || string(text) != string(in) {
			t.Errorf("%s: converted from %q to %q", name, from, text)
		}
	}
}

func TestFirstInvalid(t *testing.T) {
	for in, want := range map[string]int{"": -1, "abc": -1, "a€b": -1, "ab\xffc": 2, "a\xe2\x82": 1, "\xc0\x80": 0} {
		if got := FirstInvalid([]byte(in)); got != want {
			t.Errorf("FirstInvalid(%q) = %d, want %d", in, got, want)
		}
	}
}
