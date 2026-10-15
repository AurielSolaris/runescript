package diagnostics

import "testing"

func TestFormat(t *testing.T) {
	src := []byte("package p;\r\npublic int main() {\r\n\tuint u = signedValue;\r\n}")
	// "signedValue" starts at byte 36, line 3, column 11 (the tab counts as one column).
	d := New(TypeSignedToUnsigned, Span{
		Start: Pos{Offset: 36, Line: 3, Col: 11},
		End:   Pos{Offset: 47, Line: 3, Col: 22},
	}, "Cannot assign signed integer '%s' to unsigned integer '%s'.", "sint", "uint").
		WithHelp("Use an explicit checked conversion, or change the destination type.")

	want := "f.rs:3:11: error ASX-TYPE-004: Cannot assign signed integer 'sint' to unsigned integer 'uint'.\n" +
		"     uint u = signedValue;\n" +
		"              ^^^^^^^^^^^\n" +
		"  help: Use an explicit checked conversion, or change the destination type.\n"
	if got := Format(d, "f.rs", src); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatAtEndOfFile(t *testing.T) {
	src := []byte("x")
	d := New(LexUnterminatedString, Span{Start: Pos{Offset: 1, Line: 1, Col: 2}, End: Pos{Offset: 1, Line: 1, Col: 2}}, "msg")
	want := "f:1:2: error ASX-LEX-002: msg\n    x\n     ^\n"
	if got := Format(d, "f", src); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCodeString(t *testing.T) {
	if got := TypeSignedToUnsigned.String(); got != "ASX-TYPE-004" {
		t.Errorf("got %s", got)
	}
}
