package checker

import (
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/loader"
)

// program wraps body lines in `int main()`. The first body line is line 3.
// extra is appended after main as additional top-level declarations.
func program(body, extra string) string {
	return "package p;\npublic int main() {\n" + body + "\n    return 0;\n}\n" + extra
}

func check(t *testing.T, src string) []*diagnostics.Diagnostic {
	t.Helper()
	prog, diags := loader.LoadSource(diagnostics.NewSource("t.rsc", []byte(src)), t.TempDir())
	for _, f := range prog.Files {
		if f.AST == nil {
			t.Fatalf("parse failed: %v", diags)
		}
	}
	_, more := Check(prog)
	return append(diags, more...)
}

func TestAccepted(t *testing.T) {
	cases := map[string]string{
		"canonical entry":     "package main;\nprotected int main() {\n    return 0;\n}",
		"void main":           "package p;\npublic void main() { trace(\"hi\"); }",
		"float constants":     program("    float f = 3;\n    float32 h = 0.5;\n    float64 d = -1e300;\n    float32 big = 16777216;\n    trace(f * 2, h / 4, d < 0);", ""),
		"constant casts":      program("    int a = (int) -64.7;\n    uint b = (uint) 3.9;\n    float32 c = (float32) 0.1;\n    float d = (float) 9007199254740993;", ""),
		"runtime casts":       program("    int n = 3;\n    string s = \"4\";\n    float f = (float) n + (float) s;\n    uint u = (uint) n;\n    bool b = (bool) s;\n    trace((string) f, (int) f, (float32) f, u, b);", ""),
		"error subtypes":      program("    Error e = new RangeError(\"r\");\n    Error(string) make = (string m) => new TypeError(m);\n    trace(e.message, e.name, e.errorID, make(\"t\"));", ""),
		"throw ends a path":   "package p;\npublic int f(int n) {\n    if (n > 0) {\n        return n;\n    }\n    throw new ArgumentError(\"n\");\n}\npublic void main() { trace(f(1)); }",
		"try ends a path":     "package p;\npublic int f() {\n    try {\n        return 1;\n    } catch (Error e) {\n        return 2;\n    }\n}\npublic void main() { trace(f()); }",
		"loop in finally":     program("    try { trace(1); } finally {\n        for (int i = 0; i < 2; i++) { if (i == 1) { break; } }\n    }", ""),
		"sint is int":         program("    sint a = 1;\n    int b = a;\n    sint c = b;", ""),
		"literal to uint":     program("    uint u = 4294967295;", ""),
		"int bounds":          program("    int lo = -2147483648;\n    int hi = 2147483647;", ""),
		"constant folding":    program("    uint u = 2 * 3 - 1;\n    int x = -(5 - 10);", ""),
		"uint arithmetic":     program("    uint a = 3;\n    uint b = a * 2 + 1;\n    bool c = a < b;", ""),
		"chars":               program("    char c = 'x';\n    bool b = c < 'y' && c != '\\n';\n    int n = (int) c;\n    char d = (char) (n + 1);\n    string s = (string) c + $\"{d}\";\n    char e = (char) s;\n    char(char) f = (char x) => x;\n    trace(c, f(e), b);", ""),
		"concatenation":       program("    string s = \"a\" + \"b\";\n    s += s + \"c\";\n    int n = 2;\n    string t = $\"{s} {n} {1.5} {n > 1} {(string) n + \"!\"} {$\"{n}\"}\";", ""),
		"strings and bools":   program("    string s = \"a\";\n    bool b = s == \"a\" && !false;\n    bool c = b != true;", ""),
		"compound assign":     program("    int x = 1;\n    x += 2;\n    x -= 1;\n    x *= 3;\n    x /= 2;\n    x %= 2;\n    x++;\n    x--;", ""),
		"for scope":           program("    for (int i = 0; i < 3; i++) { trace(i); }\n    for (int i = 0; i < 3; i++) { }", ""),
		"sibling scopes":      program("    if (true) { int x = 1; } else { int x = 2; }", ""),
		"while break":         program("    while (true) { break; }", ""),
		"call before decl":    program("    trace(add(1, 2));", "public int add(int a, int b) { return a + b; }"),
		"if-else returns":     program("    trace(sign(3));", "public int sign(int n) {\n    if (n < 0) { return -1; } else if (n == 0) { return 0; } else { return 1; }\n}"),
		"discarded result":    program("    one();", "public int one() { return 1; }"),
		"uint param":          program("    trace(twice(21));", "public uint twice(uint n) { return n * 2; }"),
		"trace mixed":         program("    trace(\"x\", 1, true, -5);", ""),
		"recursion":           program("    trace(f(3));", "public int f(int n) { if (n == 0) { return 0; } return f(n - 1); }"),
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
		{"signed arg to uint param", program("    int s = 1;\n    f(s);", "public void f(uint u) { }"), diagnostics.TypeSignedToUnsigned, "4:7"},
		{"signed return from uint fn", program("", "public uint f(int s) { return s; }"), diagnostics.TypeSignedToUnsigned, "6:31"},
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
		{"int to char", program("    char c = 65;", ""), diagnostics.TypeMismatch, "3:14"},
		{"char to int", program("    int n = 'a';", ""), diagnostics.TypeMismatch, "3:13"},
		{"string to char", program("    char c = \"a\";", ""), diagnostics.TypeMismatch, "3:14"},
		{"char arithmetic", program("    char c = 'a';\n    trace(c + 1);", ""), diagnostics.TypeBadOperand, "4:13"},
		{"char plus char", program("    trace('a' + 'b');", ""), diagnostics.TypeBadOperand, "3:15"},
		{"char and string", program("    trace('a' == \"a\");", ""), diagnostics.TypeBadOperand, "3:15"},
		{"char to float", program("    trace((float) 'a');", ""), diagnostics.TypeBadConversion, "3:11"},
		{"char to bool", program("    trace((bool) 'a');", ""), diagnostics.TypeBadConversion, "3:11"},
		{"char increment", program("    char c = 'a';\n    c++;", ""), diagnostics.TypeBadOperand, "4:5"},
		{"string plus int", program("    trace(\"a\" + 1);", ""), diagnostics.TypeBadOperand, "3:15"},
		{"int plus string", program("    int n = 1;\n    trace(n + \"a\");", ""), diagnostics.TypeBadOperand, "4:13"},
		{"string minus", program("    trace(\"a\" - \"b\");", ""), diagnostics.TypeBadOperand, "3:15"},
		{"interpolated function", program("    trace($\"{f}\");", "public void f() { }"), diagnostics.TypeBadConversion, "3:14"},
		{"interpolated void", program("    trace($\"{f()}\");", "public void f() { }"), diagnostics.TypeVoidValue, "3:14"},
		{"interpolated error", program("    trace($\"{new Error()}\");", ""), diagnostics.TypeBadConversion, "3:14"},
		{"int condition", program("    if (1) { }", ""), diagnostics.TypeCondition, "3:9"},
		{"int while condition", program("    int x = 1;\n    while (x) { }", ""), diagnostics.TypeCondition, "4:12"},
		{"and on ints", program("    bool b = 1 && 2;", ""), diagnostics.TypeBadOperand, "3:16"},
		{"not on int", program("    bool b = !1;", ""), diagnostics.TypeBadOperand, "3:14"},
		{"string compound int", program("    string s = \"a\";\n    s += 1;", ""), diagnostics.TypeBadOperand, "4:5"},
		{"string compound minus", program("    string s = \"a\";\n    s -= \"b\";", ""), diagnostics.TypeBadOperand, "4:5"},
		{"bool increment", program("    bool b = true;\n    b++;", ""), diagnostics.TypeBadOperand, "4:5"},
		{"void variable", program("    void v = f();", "public void f() { }"), diagnostics.TypeVoidValue, "3:5"},
		{"void value", program("    int x = f();", "public void f() { }"), diagnostics.TypeVoidValue, "3:13"},
		{"void param", program("", "public void f(void x) { }"), diagnostics.TypeVoidValue, "6:15"},
		{"unknown type", program("    Foo x = 1;", ""), diagnostics.TypeUnknown, "3:5"},
		{"arg count", program("    f(1);", "public void f() { }"), diagnostics.TypeArgCount, "3:5"},
		{"arg type", program("    f(\"x\");", "public void f(int n) { }"), diagnostics.TypeMismatch, "3:7"},
		{"return value from void", program("", "public void f() { return 1; }"), diagnostics.TypeBadReturn, "6:26"},
		{"bare return from int", program("", "public int f() { return; }"), diagnostics.TypeBadReturn, "6:18"},
		{"missing return", program("", "public int f() {\n    trace(1);\n}"), diagnostics.TypeMissingReturn, "8:1"},
		{"if without else", program("", "public int f(bool b) {\n    if (b) { return 1; }\n}"), diagnostics.TypeMissingReturn, "8:1"},
		{"else-if chain without final else", program("", "public int f(int n) {\n    if (n > 0) { return 1; } else if (n < 0) { return -1; }\n}"), diagnostics.TypeMissingReturn, "8:1"},
		{"loop is not terminating", program("", "public int f() {\n    while (true) { return 1; }\n}"), diagnostics.TypeMissingReturn, "8:1"},

		// Names and scopes.
		{"undefined variable", program("    trace(y);", ""), diagnostics.NameUndefined, "3:11"},
		{"undefined function", program("    g();", ""), diagnostics.NameUndefined, "3:5"},
		{"use before decl", program("    int x = x;", ""), diagnostics.NameUndefined, "3:13"},
		{"out of scope", program("    if (true) { int x = 1; }\n    trace(x);", ""), diagnostics.NameUndefined, "4:11"},
		{"for var out of scope", program("    for (int i = 0; i < 1; i++) { }\n    trace(i);", ""), diagnostics.NameUndefined, "4:11"},
		{"redeclare", program("    int x = 1;\n    int x = 2;", ""), diagnostics.NameRedeclared, "4:9"},
		{"shadow", program("    int x = 1;\n    if (true) { int x = 2; }", ""), diagnostics.NameRedeclared, "4:21"},
		{"shadow param", program("", "public void f(int n) { int n = 1; }"), diagnostics.NameRedeclared, "6:28"},
		{"duplicate param", program("", "public void f(int a, int a) { }"), diagnostics.NameRedeclared, "6:26"},
		{"var named like function", program("    int f = 1;", "public void f() { }"), diagnostics.NameRedeclared, "3:9"},
		{"duplicate function", program("", "public void f() { }\npublic void f() { }"), diagnostics.NameRedeclared, "7:13"},
		{"redeclare trace", program("", "public void trace() { }"), diagnostics.NameBuiltinReused, "6:13"},
		{"call a variable", program("    int x = 1;\n    x();", ""), diagnostics.NameNotCallable, "4:5"},
		{"trace a function value", program("    trace(f);", "public void f() { }"), diagnostics.TypeBadOperand, "3:11"},
		{"trace as value", program("    void(string) t = trace;", ""), diagnostics.NameNotAValue, "3:22"},
		{"assign to param", program("", "public void f(int n) { n = 1; }"), diagnostics.NameAssignToParam, "6:24"},
		{"increment param", program("", "public void f(int n) { n++; }"), diagnostics.NameAssignToParam, "6:24"},
		{"break outside loop", program("    break;", ""), diagnostics.FlowBranchOutsideLoop, "3:5"},
		{"continue in if", program("    if (true) { continue; }", ""), diagnostics.FlowBranchOutsideLoop, "3:17"},

		// Entry point.
		{"no main", "package p;\npublic void f() { }", diagnostics.EntryMissing, "1:1"},
		{"main with params", "package p;\npublic int main(int x) { return 0; }", diagnostics.EntryBadSignature, "2:12"},
		{"main returns string", "package p;\npublic string main() { return \"\"; }", diagnostics.EntryBadSignature, "2:15"},
		{"main returns uint", "package p;\npublic uint main() { return 0; }", diagnostics.EntryBadSignature, "2:13"},
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

// withString wraps body in a program that uses std:string. The first body
// line is line 4.
func withString(body string) string {
	return "package p;\nuse \"std:string\";\npublic int main() {\n" + body + "\n    return 0;\n}\n"
}

func TestStdStringAccepted(t *testing.T) {
	cases := map[string]string{
		"call in initializer": withString(`    string s = string.toUpper("rune");`),
		"call as statement":   withString(`    string.trimSpace(" a ");`),
		"nested calls":        withString(`    string s = string.toUpper(string.trimSpace(" a "));`),
		"bool result":         withString(`    if (string.hasPrefix("rune", "ru") && !string.contains("x", "y")) { }`),
		"int result":          withString(`    int i = string.index("rune", "n") + string.count("aaa", "a");`),
		"int args":            withString(`    string s = string.replace(string.repeat("ab", 3), "a", "", -1);`),
		"trace":               withString(`    trace(string.toLower("ABC"), string.lastIndex("abcabc", "c"));`),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if diags := check(t, src); len(diags) != 0 {
				t.Errorf("unexpected diagnostics: %v", diags)
			}
		})
	}
}

func TestStdStringRejected(t *testing.T) {
	cases := []struct {
		name string
		src  string
		code diagnostics.Code
		pos  string
	}{
		{"missing file", "package p;\nuse \"lib/x.rsc\";\npublic void main() { }", diagnostics.UseNotFound, "2:5"},
		{"go use", "package p;\nuse \"go:strings\";\npublic void main() { }", diagnostics.UseUnsupported, "2:5"},
		{"unknown std package", "package p;\nuse \"std:math\";\npublic void main() { }", diagnostics.UseUnknownPackage, "2:5"},
		{"duplicate use", "package p;\nuse \"std:string\";\nuse \"std:string\";\npublic void main() { }", diagnostics.UseDuplicate, "3:1"},
		{"missing use", "package p;\npublic void main() { trace(string.toUpper(\"a\")); }", diagnostics.NamePackageNotUsed, "2:28"},
		{"unknown qualifier", "package p;\npublic void main() { foo.bar(); }", diagnostics.NamePackageNotUsed, "2:22"},
		{"Go casing", withString(`    trace(string.ToUpper("a"));`), diagnostics.NameNotInPackage, "4:18"},
		{"unknown function", withString(`    trace(string.split("a,b", ","));`), diagnostics.NameNotInPackage, "4:18"},
		{"arg count", withString(`    trace(string.repeat("a"));`), diagnostics.TypeArgCount, "4:11"},
		{"arg type", withString(`    trace(string.toUpper(1));`), diagnostics.TypeMismatch, "4:26"},
		{"uint arg to int param", withString("    uint n = 2;\n    trace(string.repeat(\"a\", n));"), diagnostics.TypeUnsignedToSigned, "5:30"},
		{"int result to uint", withString(`    uint i = string.index("a", "a");`), diagnostics.TypeSignedToUnsigned, "4:14"},
		{"string result to int", withString(`    int n = string.toUpper("a");`), diagnostics.TypeMismatch, "4:13"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			diags := check(t, c.src)
			if len(diags) == 0 {
				t.Fatalf("expected %s at %s, got no diagnostics", c.code, c.pos)
			}
			if d := diags[0]; d.Code != c.code || d.Span.Start.String() != c.pos {
				t.Errorf("got %s at %s (%s), want %s at %s", d.Code, d.Span.Start, d.Message, c.code, c.pos)
			}
		})
	}
}

func TestGoCasingHint(t *testing.T) {
	diags := check(t, withString(`    trace(string.ToUpper("a"));`))
	if len(diags) != 1 || !strings.Contains(diags[0].Help, "'toUpper'") {
		t.Errorf("expected a camelCase hint, got %v", diags)
	}
}

func TestFunctionsAccepted(t *testing.T) {
	cases := map[string]string{
		"callback lambda": program("    trace(apply((int a, int b) => a * b, 6, 7));",
			"public int apply(int(int, int) op, int a, int b) { return op(a, b); }"),
		"named function as value": program("    int(int, int) f = add;\n    trace(f(1, 2));",
			"public int add(int a, int b) { return a + b; }"),
		"nested function":  program("    public int sq(int n) { return n * n; }\n    trace(sq(3));", ""),
		"nested recursion": program("    private int fact(int n) { if (n <= 1) { return 1; } return n * fact(n - 1); }\n    trace(fact(5));", ""),
		"nested mutual recursion": program("    private bool isEven(int n);\n"+
			"    private bool isOdd(int n) { if (n == 0) { return false; } return isEven(n - 1); }\n"+
			"    private bool isEven(int n) { if (n == 0) { return true; } return isOdd(n - 1); }\n"+
			"    trace(isEven(10));", ""),
		"top-level prototype":        program("    trace(f(1));", "public int f(int);\npublic int f(int x) { return x; }"),
		"prototype names may differ": program("", "public int f(int a);\npublic int f(int b) { return b; }"),
		"returning functions": program("    trace(pick(true)(4));",
			"public int(int) pick(bool b) {\n    private int dbl(int n) { return n * 2; }\n    if (b) { return dbl; }\n    return (int n) => n;\n}"),
		"void lambda with call body":    program("    void() hi = () => trace(\"hi\");\n    hi();", ""),
		"inferred expression lambda":    program("    trace(((int x) => x + 1)(2));", ""),
		"inferred block lambda":         program("    trace(((int x) => { if (x > 0) { return x; } return -x; })(-3));", ""),
		"reassign function variable":    program("    int(int) f = (int x) => x;\n    f = (int x) => x * 2;\n    trace(f(2));", ""),
		"contextual unsigned result":    program("    uint(uint) inc = (uint x) => x + 1;\n    trace(inc(1));", ""),
		"sibling nested functions":      program("    private int a() { return 1; }\n    private int b() { return a() + 1; }\n    trace(b());", ""),
		"same name across boundary":     program("    int n = 1;\n    private int f() { int n = 2; return n; }\n    trace(f(), n);", ""),
		"higher-order chain":            program("    trace(twice((int x) => x + 3)(1));", "public int(int) twice(int(int) f) { return f; }"),
		"function-typed parameter call": program("    run(() => trace(\"ran\"));", "public void run(void() cb) { cb(); }"),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if diags := check(t, src); len(diags) != 0 {
				t.Errorf("unexpected diagnostics: %v", diags)
			}
		})
	}
}

func TestFunctionsRejected(t *testing.T) {
	cases := []struct {
		name string
		src  string
		code diagnostics.Code
		pos  string
	}{
		{"outer variable in nested function", program("    int n = 1;\n    private int f() { return n; }", ""), diagnostics.NameAcrossBoundary, "4:30"},
		{"outer variable in lambda", program("    int n = 1;\n    int(int) f = (int x) => x + n;", ""), diagnostics.NameAcrossBoundary, "4:33"},
		{"assign outer variable", program("    int n = 1;\n    private void f() { n = 2; }", ""), diagnostics.NameAcrossBoundary, "4:24"},
		{"outer parameter in nested", program("", "public int f(int p) {\n    private int g() { return p; }\n    return g();\n}"), diagnostics.NameAcrossBoundary, "7:30"},
		{"prototype never defined", program("    private int g(int);", ""), diagnostics.NameNeverDefined, "3:17"},
		{"prototype defined in an inner block", program("    private int g(int);\n    if (true) { private int g(int x) { return x; } }", ""), diagnostics.NameRedeclared, "4:29"},
		{"prototype mismatch", program("", "public int f(int);\npublic int f(string s) { return 0; }"), diagnostics.TypeProtoMismatch, "7:12"},
		{"duplicate prototype", program("", "public int f(int);\npublic int f(int);\npublic int f(int x) { return x; }"), diagnostics.NameRedeclared, "7:12"},
		{"prototype after definition", program("", "public int f(int x) { return x; }\npublic int f(int);"), diagnostics.NameRedeclared, "7:12"},
		{"call a non-function value", program("    trace((1)(2));", ""), diagnostics.TypeNotCallable, "3:11"},
		{"function value arg count", program("    int(int) f = (int x) => x;\n    trace(f(1, 2));", ""), diagnostics.TypeArgCount, "4:11"},
		{"lambda parameter mismatch", program("    int(int) f = (string s) => 1;", ""), diagnostics.TypeMismatch, "3:18"},
		{"lambda result mismatch", program("    int(int) f = (int x) => \"a\";", ""), diagnostics.TypeMismatch, "3:29"},
		{"lambda signed to unsigned result", program("    int s = 1;\n    uint(int) f = (int x) => x;", ""), diagnostics.TypeSignedToUnsigned, "4:30"},
		{"named function type mismatch", program("    int(int) f = add;", "public int add(int a, int b) { return a + b; }"), diagnostics.TypeMismatch, "3:18"},
		{"break across boundary", program("    while (true) {\n        private void f() { break; }\n        break;\n    }", ""), diagnostics.FlowBranchOutsideLoop, "4:28"},
		{"lambda missing return", program("    int(int) f = (int x) => { if (x > 0) { return 1; } };", ""), diagnostics.TypeMissingReturn, "3:56"},
		{"nested function clashes with variable", program("    int f = 1;\n    private int f() { return 1; }", ""), diagnostics.NameRedeclared, "4:17"},
		{"void in function type", program("    void(void) f = () => trace(1);", ""), diagnostics.TypeVoidValue, "3:10"},
		{"compare functions", program("    bool b = add == add;", "public int add(int a, int b) { return a + b; }"), diagnostics.TypeBadOperand, "3:18"},
		{"assign to function", program("    add = add;", "public int add(int a, int b) { return a + b; }"), diagnostics.NameAssignToParam, "3:5"},
		{"call int variable", program("    int x = 1;\n    trace(x(2));", ""), diagnostics.NameNotCallable, "4:11"},
		{"lambda returns void call", program("    trace(((int x) => { return trace(x); })(1));", ""), diagnostics.TypeVoidValue, "3:32"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			diags := check(t, c.src)
			if len(diags) == 0 {
				t.Fatalf("expected %s at %s, got no diagnostics", c.code, c.pos)
			}
			if d := diags[0]; d.Code != c.code || d.Span.Start.String() != c.pos {
				t.Errorf("got %s at %s (%s), want %s at %s", d.Code, d.Span.Start, d.Message, c.code, c.pos)
			}
		})
	}
}
