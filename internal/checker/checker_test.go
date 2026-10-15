package checker

import (
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/parser"
)

// program wraps body lines in `int main()`. The first body line is line 3.
// extra is appended after main as additional top-level declarations.
func program(body, extra string) string {
	return "package p;\nint main() {\n" + body + "\n    return 0;\n}\n" + extra
}

func check(t *testing.T, src string) []*diagnostics.Diagnostic {
	t.Helper()
	f, diags := parser.Parse([]byte(src))
	if len(diags) != 0 {
		t.Fatalf("parse failed: %v", diags)
	}
	_, diags = Check(f)
	return diags
}

func TestAccepted(t *testing.T) {
	cases := map[string]string{
		"canonical entry":     "package main;\nprotected int main() {\n    return 0;\n}",
		"void main":           "package p;\nvoid main() { trace(\"hi\"); }",
		"sint is int":         program("    sint a = 1;\n    int b = a;\n    sint c = b;", ""),
		"literal to uint":     program("    uint u = 4294967295;", ""),
		"int bounds":          program("    int lo = -2147483648;\n    int hi = 2147483647;", ""),
		"constant folding":    program("    uint u = 2 * 3 - 1;\n    int x = -(5 - 10);", ""),
		"uint arithmetic":     program("    uint a = 3;\n    uint b = a * 2 + 1;\n    bool c = a < b;", ""),
		"strings and bools":   program("    string s = \"a\";\n    bool b = s == \"a\" && !false;\n    bool c = b != true;", ""),
		"compound assign":     program("    int x = 1;\n    x += 2;\n    x -= 1;\n    x *= 3;\n    x /= 2;\n    x %= 2;\n    x++;\n    x--;", ""),
		"for scope":           program("    for (int i = 0; i < 3; i++) { trace(i); }\n    for (int i = 0; i < 3; i++) { }", ""),
		"sibling scopes":      program("    if (true) { int x = 1; } else { int x = 2; }", ""),
		"while break":         program("    while (true) { break; }", ""),
		"call before decl":    program("    trace(add(1, 2));", "int add(int a, int b) { return a + b; }"),
		"if-else returns":     program("    trace(sign(3));", "int sign(int n) {\n    if (n < 0) { return -1; } else if (n == 0) { return 0; } else { return 1; }\n}"),
		"discarded result":    program("    one();", "int one() { return 1; }"),
		"uint param":          program("    trace(twice(21));", "uint twice(uint n) { return n * 2; }"),
		"trace mixed":         program("    trace(\"x\", 1, true, -5);", ""),
		"recursion":           program("    trace(f(3));", "int f(int n) { if (n == 0) { return 0; } return f(n - 1); }"),
		"visibility recorded": "package p;\npublic void main() { }\nprivate void helper() { }",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if diags := check(t, src); len(diags) != 0 {
				t.Errorf("unexpected diagnostics: %v", diags)
			}
		})
	}
}

func TestRejected(t *testing.T) {
	cases := []struct {
		name string
		src  string
		code diagnostics.Code
		pos  string
	}{
		// Signedness (CLAUDE.md §4.5).
		{"signed var to uint", program("    sint signedValue = -10;\n    uint unsignedValue = signedValue;", ""), diagnostics.TypeSignedToUnsigned, "4:26"},
		{"positive signed var to uint", program("    int s = 10;\n    uint u = s;", ""), diagnostics.TypeSignedToUnsigned, "4:14"},
		{"signed assign to uint", program("    int s = 1;\n    uint u = 0;\n    u = s;", ""), diagnostics.TypeSignedToUnsigned, "5:9"},
		{"signed arg to uint param", program("    int s = 1;\n    f(s);", "void f(uint u) { }"), diagnostics.TypeSignedToUnsigned, "4:7"},
		{"signed return from uint fn", program("", "uint f(int s) { return s; }"), diagnostics.TypeSignedToUnsigned, "6:24"},
		{"uint to int", program("    uint u = 1;\n    int s = u;", ""), diagnostics.TypeUnsignedToSigned, "4:13"},
		{"negative literal to uint", program("    uint u = -1;", ""), diagnostics.TypeNegativeToUnsigned, "3:14"},
		{"negative folded to uint", program("    uint u = 1 - 2;", ""), diagnostics.TypeNegativeToUnsigned, "3:14"},
		{"mixed arithmetic", program("    int s = 1;\n    uint u = 1;\n    trace(s + u);", ""), diagnostics.TypeMixedSignedness, "5:13"},
		{"mixed comparison", program("    int s = 1;\n    uint u = 1;\n    bool b = s < u;", ""), diagnostics.TypeMixedSignedness, "5:16"},
		{"negate uint", program("    uint u = 1;\n    trace(-u);", ""), diagnostics.TypeBadOperand, "4:11"},
		{"negative constant in uint expr", program("    uint u = 1;\n    trace(u + -1);", ""), diagnostics.TypeNegativeToUnsigned, "4:15"},

		// Numeric ranges (widening not implemented).
		{"int too big", program("    int x = 2147483648;", ""), diagnostics.TypeConstOutOfRange, "3:13"},
		{"int too small", program("    int x = -2147483649;", ""), diagnostics.TypeConstOutOfRange, "3:13"},
		{"uint too big", program("    uint x = 4294967296;", ""), diagnostics.TypeConstOutOfRange, "3:14"},
		{"folded too big", program("    int x = 65536 * 65536;", ""), diagnostics.TypeConstOutOfRange, "3:13"},
		{"trace literal too big", program("    trace(3000000000);", ""), diagnostics.TypeConstOutOfRange, "3:11"},
		{"constant div by zero", program("    int x = 1 / 0;", ""), diagnostics.TypeConstDivByZero, "3:15"},

		// Other type errors.
		{"string to int", program("    int x = \"a\";", ""), diagnostics.TypeMismatch, "3:13"},
		{"int to string", program("    string s = 1;", ""), diagnostics.TypeMismatch, "3:16"},
		{"bool to int", program("    int x = true;", ""), diagnostics.TypeMismatch, "3:13"},
		{"string concat", program("    trace(\"a\" + \"b\");", ""), diagnostics.TypeBadOperand, "3:15"},
		{"int condition", program("    if (1) { }", ""), diagnostics.TypeCondition, "3:9"},
		{"int while condition", program("    int x = 1;\n    while (x) { }", ""), diagnostics.TypeCondition, "4:12"},
		{"and on ints", program("    bool b = 1 && 2;", ""), diagnostics.TypeBadOperand, "3:16"},
		{"not on int", program("    bool b = !1;", ""), diagnostics.TypeBadOperand, "3:14"},
		{"string compound", program("    string s = \"a\";\n    s += \"b\";", ""), diagnostics.TypeBadOperand, "4:5"},
		{"bool increment", program("    bool b = true;\n    b++;", ""), diagnostics.TypeBadOperand, "4:5"},
		{"void variable", program("    void v = f();", "void f() { }"), diagnostics.TypeVoidValue, "3:5"},
		{"void value", program("    int x = f();", "void f() { }"), diagnostics.TypeVoidValue, "3:13"},
		{"void param", program("", "void f(void x) { }"), diagnostics.TypeVoidValue, "6:8"},
		{"unknown type", program("    Foo x = 1;", ""), diagnostics.TypeUnknown, "3:5"},
		{"arg count", program("    f(1);", "void f() { }"), diagnostics.TypeArgCount, "3:5"},
		{"arg type", program("    f(\"x\");", "void f(int n) { }"), diagnostics.TypeMismatch, "3:7"},
		{"return value from void", program("", "void f() { return 1; }"), diagnostics.TypeBadReturn, "6:19"},
		{"bare return from int", program("", "int f() { return; }"), diagnostics.TypeBadReturn, "6:11"},
		{"missing return", program("", "int f() {\n    trace(1);\n}"), diagnostics.TypeMissingReturn, "8:1"},
		{"if without else", program("", "int f(bool b) {\n    if (b) { return 1; }\n}"), diagnostics.TypeMissingReturn, "8:1"},
		{"loop is not terminating", program("", "int f() {\n    while (true) { return 1; }\n}"), diagnostics.TypeMissingReturn, "8:1"},

		// Names and scopes.
		{"undefined variable", program("    trace(y);", ""), diagnostics.NameUndefined, "3:11"},
		{"undefined function", program("    g();", ""), diagnostics.NameUndefined, "3:5"},
		{"use before decl", program("    int x = x;", ""), diagnostics.NameUndefined, "3:13"},
		{"out of scope", program("    if (true) { int x = 1; }\n    trace(x);", ""), diagnostics.NameUndefined, "4:11"},
		{"for var out of scope", program("    for (int i = 0; i < 1; i++) { }\n    trace(i);", ""), diagnostics.NameUndefined, "4:11"},
		{"redeclare", program("    int x = 1;\n    int x = 2;", ""), diagnostics.NameRedeclared, "4:9"},
		{"shadow", program("    int x = 1;\n    if (true) { int x = 2; }", ""), diagnostics.NameRedeclared, "4:21"},
		{"shadow param", program("", "void f(int n) { int n = 1; }"), diagnostics.NameRedeclared, "6:21"},
		{"duplicate param", program("", "void f(int a, int a) { }"), diagnostics.NameRedeclared, "6:19"},
		{"var named like function", program("    int f = 1;", "void f() { }"), diagnostics.NameRedeclared, "3:9"},
		{"duplicate function", program("", "void f() { }\nvoid f() { }"), diagnostics.NameRedeclared, "7:6"},
		{"redeclare trace", program("", "void trace() { }"), diagnostics.NameBuiltinReused, "6:6"},
		{"call a variable", program("    int x = 1;\n    x();", ""), diagnostics.NameNotCallable, "4:5"},
		{"function as value", program("    trace(f);", "void f() { }"), diagnostics.NameNotAValue, "3:11"},
		{"assign to param", program("", "void f(int n) { n = 1; }"), diagnostics.NameAssignToParam, "6:17"},
		{"increment param", program("", "void f(int n) { n++; }"), diagnostics.NameAssignToParam, "6:17"},
		{"break outside loop", program("    break;", ""), diagnostics.FlowBranchOutsideLoop, "3:5"},
		{"continue in if", program("    if (true) { continue; }", ""), diagnostics.FlowBranchOutsideLoop, "3:17"},

		// Entry point.
		{"no main", "package p;\nvoid f() { }", diagnostics.EntryMissing, "1:1"},
		{"main with params", "package p;\nint main(int x) { return 0; }", diagnostics.EntryBadSignature, "2:5"},
		{"main returns string", "package p;\nstring main() { return \"\"; }", diagnostics.EntryBadSignature, "2:8"},
		{"main returns uint", "package p;\nuint main() { return 0; }", diagnostics.EntryBadSignature, "2:6"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			diags := check(t, c.src)
			if len(diags) == 0 {
				t.Fatalf("expected %s at %s, got no diagnostics", c.code, c.pos)
			}
			d := diags[0]
			if d.Code != c.code || d.Span.Start.String() != c.pos {
				t.Errorf("got %s at %s (%s), want %s at %s", d.Code, d.Span.Start, d.Message, c.code, c.pos)
			}
		})
	}
}

// The signed-to-unsigned diagnostic must follow the style in CLAUDE.md §4.5.
func TestSignedToUnsignedDiagnosticText(t *testing.T) {
	diags := check(t, program("    sint signedValue = -10;\n    uint unsignedValue = signedValue;", ""))
	if len(diags) != 1 {
		t.Fatalf("got %v", diags)
	}
	d := diags[0]
	if d.Code.String() != "ASX-TYPE-004" {
		t.Errorf("code = %s", d.Code)
	}
	if want := "Cannot assign signed integer 'sint' to unsigned integer 'uint'."; d.Message != want {
		t.Errorf("message = %q, want %q", d.Message, want)
	}
	if !strings.Contains(d.Help, "explicit checked conversion") || !strings.Contains(d.Help, "change the destination type") {
		t.Errorf("help = %q", d.Help)
	}
}

func TestReportsMultipleErrors(t *testing.T) {
	diags := check(t, program("    uint a = -1;\n    string s = 2;\n    trace(nope);", ""))
	if len(diags) != 3 {
		t.Errorf("got %d diagnostics, want 3: %v", len(diags), diags)
	}
}
