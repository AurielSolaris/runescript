package stdlib

import (
	"reflect"
	"testing"
)

// fault runs f and returns the Fault it panics with ("" type if none).
func fault(f func()) (got Fault) {
	defer func() {
		if r := recover(); r != nil {
			got = r.(Fault)
		}
	}()
	f()
	return
}

func TestCSVParse(t *testing.T) {
	bom := string(rune(0xFEFF))
	cases := []struct {
		text string
		sep  string
		want [][]string
	}{
		{"", ",", [][]string{}},
		{"\n\n\r\n", ",", [][]string{}},
		{"a,b,c", ",", [][]string{{"a", "b", "c"}}},
		{"a,b\nc,d\n", ",", [][]string{{"a", "b"}, {"c", "d"}}},
		{"a,b\r\nc,d\r\n", ",", [][]string{{"a", "b"}, {"c", "d"}}},
		{"a\n\nb", ",", [][]string{{"a"}, {"b"}}},
		{"a,,", ",", [][]string{{"a", "", ""}}},
		{",", ",", [][]string{{"", ""}}},
		{`""`, ",", [][]string{{""}}},
		{`"a,b","c""d"`, ",", [][]string{{"a,b", `c"d`}}},
		{"\"line1\nline2\",x\ny", ",", [][]string{{"line1\nline2", "x"}, {"y"}}},
		{"\"crlf\r\nkept\"", ",", [][]string{{"crlf\r\nkept"}}},
		{"a\rb,c", ",", [][]string{{"a\rb", "c"}}},
		{" a , b ", ",", [][]string{{" a ", " b "}}},
		{bom + "a,b", ",", [][]string{{"a", "b"}}},
		{"a;b,c", ";", [][]string{{"a", "b,c"}}},
		{"a→b→c", "→", [][]string{{"a", "b", "c"}}},
		{"é,ü\n", ",", [][]string{{"é", "ü"}}},
		{"a\tb", "\t", [][]string{{"a", "b"}}},
	}
	for _, c := range cases {
		if got := csvParse(c.text, c.sep); !reflect.DeepEqual(got, c.want) {
			t.Errorf("parse(%q, %q) = %q, want %q", c.text, c.sep, got, c.want)
		}
	}
}

func TestCSVParseErrors(t *testing.T) {
	cases := []struct{ text, want string }{
		{`a"b`, `line 1, column 2: '"' in an unquoted field; quote the field and write the '"' twice`},
		{"ok\nab,c\"d", `line 2, column 5: '"' in an unquoted field; quote the field and write the '"' twice`},
		{`"a"b`, `line 1, column 4: expected a separator or a line break after the closing '"'`},
		{"x\n\"multi\nline\"z", `line 3, column 6: expected a separator or a line break after the closing '"'`},
		{"x,\"open", `line 1, column 3: quoted field has no closing '"'`},
		{"a\nb,\"c\nd", `line 2, column 3: quoted field has no closing '"'`},
		{"é\"", `line 1, column 3: '"' in an unquoted field; quote the field and write the '"' twice`},
	}
	for _, c := range cases {
		f := fault(func() { csvParse(c.text, ",") })
		if f.Type != "Error" || f.Message != c.want {
			t.Errorf("parse(%q): %s %q, want Error %q", c.text, f.Type, f.Message, c.want)
		}
	}
	for _, r := range []rune{'"', '\r', '\n'} {
		if f := fault(func() { csvSep(r) }); f.Type != "ArgumentError" {
			t.Errorf("separator %q: %v, want an ArgumentError", r, f)
		}
	}
}

func TestCSVFormat(t *testing.T) {
	cases := []struct {
		rows [][]string
		sep  string
		want string
	}{
		{[][]string{}, ",", ""},
		{[][]string{{"a", "b"}, {"c"}}, ",", "a,b\nc\n"},
		{[][]string{{"a,b", `q"q`, "line\nbreak", "cr\rx", " sp "}}, ",", "\"a,b\",\"q\"\"q\",\"line\nbreak\",\"cr\rx\", sp \n"},
		{[][]string{{""}}, ",", "\"\"\n"},
		{[][]string{{"", ""}}, ",", ",\n"},
		{[][]string{{"a;b", "c,d"}}, ";", "\"a;b\";c,d\n"},
	}
	for _, c := range cases {
		got := csvFormat(c.rows, c.sep)
		if got != c.want {
			t.Errorf("format(%q, %q) = %q, want %q", c.rows, c.sep, got, c.want)
		}
		if back := csvParse(got, c.sep); !reflect.DeepEqual(back, c.rows) {
			t.Errorf("parse(format(%q)) = %q", c.rows, back)
		}
	}
	f := fault(func() { csvFormat([][]string{{"a"}, {}}, ",") })
	if f.Type != "ArgumentError" || f.Message != "rows[1] has no fields; a CSV record has at least one" {
		t.Errorf("empty record: %v", f)
	}
}
