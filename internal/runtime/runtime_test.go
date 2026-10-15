package runtime_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

func run(t *testing.T, src string) (string, int, error) {
	t.Helper()
	var out strings.Builder
	code, diags, err := driver.Run([]byte(src), &out)
	if diags != nil {
		t.Fatalf("compile errors: %v", diags)
	}
	return out.String(), code, err
}

func TestPrograms(t *testing.T) {
	cases := []struct {
		name, src, out string
		code           int
	}{
		{"exit code", "package p;\npublic int main() { return 7; }", "", 7},
		{"void main", "package p;\npublic void main() { trace(\"hi\"); }", "hi\n", 0},
		{"trace joins with spaces", "package p;\npublic void main() { trace(\"a\", 1, -2, true, false); }", "a 1 -2 true false\n", 0},
		{"trace no args", "package p;\npublic void main() { trace(); }", "\n", 0},
		{"string escapes", "package p;\npublic void main() { trace(\"q\\\"\\\\\"); }", "q\"\\\n", 0},
		{"arithmetic", "package p;\npublic void main() { int x = 7; trace(x + 2, x - 9, x * 3, x / 2, x % 4, -x); }", "9 -2 21 3 3 -7\n", 0},
		{"truncating division", "package p;\npublic void main() { int x = -7; trace(x / 2, x % 2); }", "-3 -1\n", 0},
		{"precedence", "package p;\npublic void main() { trace(2 + 3 * 4, (2 + 3) * 4); }", "14 20\n", 0},
		{"constant folding", "package p;\npublic void main() { trace(10 - 3, 7 / 2, -7 % 2, -7 / 2, -(4 - 9)); }", "7 3 -1 -3 5\n", 0},
		{"large intermediate constants fold exactly", "package p;\npublic void main() { trace(4000000000 / 2 - 1999999999); }", "1\n", 0},
		{"comparisons", "package p;\npublic void main() { int a = 1; int b = 2; trace(a < b, a <= a, a > b, b >= a, a == b, a != b); }", "true true false true false true\n", 0},
		{"string equality", "package p;\npublic void main() { string s = \"x\"; trace(s == \"x\", s != \"x\"); }", "true false\n", 0},
		{"assignment", "package p;\npublic void main() { int x = 1; x = 5; x += 2; x -= 1; x *= 3; x /= 2; x %= 5; trace(x); }", "4\n", 0},
		{"inc dec", "package p;\npublic void main() { int x = 0; x++; x++; x--; trace(x); }", "1\n", 0},
		{"if else chain", "package p;\npublic void f(int n) { if (n < 0) { trace(\"neg\"); } else if (n == 0) { trace(\"zero\"); } else { trace(\"pos\"); } }\npublic void main() { f(-1); f(0); f(1); }", "neg\nzero\npos\n", 0},
		{"while", "package p;\npublic void main() { int i = 0; while (i < 3) { trace(i); i++; } }", "0\n1\n2\n", 0},
		{"for", "package p;\npublic void main() { for (int i = 3; i > 0; i--) { trace(i); } }", "3\n2\n1\n", 0},
		{"break continue", "package p;\npublic void main() { for (int i = 0; i < 10; i++) { if (i == 1) { continue; } if (i == 3) { break; } trace(i); } }", "0\n2\n", 0},
		{"for without clauses", "package p;\npublic void main() { int i = 0; for (;;) { i++; if (i == 4) { break; } } trace(i); }", "4\n", 0},
		{"while continue", "package p;\npublic void main() { int i = 0; while (i < 4) { i++; if (i % 2 == 0) { continue; } trace(i); } }", "1\n3\n", 0},
		{"nested loops break inner", "package p;\npublic void main() { for (int i = 0; i < 2; i++) { for (int j = 0; j < 5; j++) { if (j == 2) { break; } trace(i, j); } } }", "0 0\n0 1\n1 0\n1 1\n", 0},
		{"return from loop", "package p;\npublic int find() { for (int i = 0; i < 100; i++) { if (i * i > 50) { return i; } } return -1; }\npublic void main() { trace(find()); }", "8\n", 0},
		{"short circuit", "package p;\npublic bool boom() { trace(\"boom\"); return true; }\npublic void main() { trace(false && boom(), true || boom()); }", "false true\n", 0},
		{"recursion", "package p;\npublic int fact(int n) { if (n <= 1) { return 1; } return n * fact(n - 1); }\npublic int main() { trace(fact(10)); return 0; }", "3628800\n", 0},
		{"arguments are copied", "package p;\npublic int inc(int n) { int m = n; m++; return m; }\npublic void main() { int x = 1; trace(inc(x), x); }", "2 1\n", 0},
		{"scopes reset per iteration", "package p;\npublic void main() { for (int i = 0; i < 2; i++) { int y = i * 10; trace(y); } }", "0\n10\n", 0},
		{"uint max", "package p;\npublic void main() { uint u = 4294967295; trace(u); }", "4294967295\n", 0},
		{"int bounds", "package p;\npublic void main() { int lo = -2147483648; int hi = 2147483647; trace(lo, hi); }", "-2147483648 2147483647\n", 0},
		{"sint same as int", "package p;\npublic int main() { sint s = -3; int i = s; return i + 3; }", "", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, code, err := run(t, c.src)
			if err != nil {
				t.Fatalf("runtime error: %v", err)
			}
			if out != c.out || code != c.code {
				t.Errorf("got output %q exit %d, want %q exit %d", out, code, c.out, c.code)
			}
		})
	}
}

func TestRuntimeErrors(t *testing.T) {
	cases := []struct {
		name, src string
		code      diagnostics.Code
		pos       string
	}{
		{"int overflow", "package p;\npublic void main() {\n    int x = 2147483647;\n    x = x + 1;\n}", diagnostics.RunOverflow, "4:11"},
		{"int overflow increment", "package p;\npublic void main() {\n    int x = 2147483647;\n    x++;\n}", diagnostics.RunOverflow, "4:5"},
		{"int underflow", "package p;\npublic void main() {\n    int x = -2147483648;\n    x -= 1;\n}", diagnostics.RunOverflow, "4:5"},
		{"negate min int", "package p;\npublic void main() {\n    int x = -2147483648;\n    trace(-x);\n}", diagnostics.RunOverflow, "4:11"},
		{"mul overflow", "package p;\npublic void main() {\n    int x = 65536;\n    trace(x * x);\n}", diagnostics.RunOverflow, "4:13"},
		{"uint below zero", "package p;\npublic void main() {\n    uint u = 0;\n    u--;\n}", diagnostics.RunOverflow, "4:5"},
		{"uint overflow", "package p;\npublic void main() {\n    uint u = 4294967295;\n    trace(u + 1);\n}", diagnostics.RunOverflow, "4:13"},
		{"divide by zero", "package p;\npublic void main() {\n    int z = 0;\n    trace(1 / z);\n}", diagnostics.RunDivideByZero, "4:13"},
		{"modulo by zero", "package p;\npublic void main() {\n    int z = 0;\n    trace(1 % z);\n}", diagnostics.RunDivideByZero, "4:13"},
		{"infinite recursion", "package p;\npublic void f() {\n    f();\n}\npublic void main() { f(); }", diagnostics.RunCallDepth, "3:5"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := run(t, c.src)
			var re *runtime.Error
			if !errors.As(err, &re) {
				t.Fatalf("expected runtime error %s, got %v", c.code, err)
			}
			if re.Diag.Code != c.code || re.Diag.Span.Start.String() != c.pos {
				t.Errorf("got %s at %s (%s), want %s at %s", re.Diag.Code, re.Diag.Span.Start, re.Diag.Message, c.code, c.pos)
			}
		})
	}
}

func TestOutputBeforeRuntimeErrorIsKept(t *testing.T) {
	out, _, err := run(t, "package p;\npublic void main() { trace(\"before\"); int z = 0; trace(1 / z); }")
	if err == nil || out != "before\n" {
		t.Errorf("got %q, %v", out, err)
	}
}

func TestStdString(t *testing.T) {
	src := `package p;
use "std:string";
public void main() {
    trace(string.clone("rune"));
    trace(string.compare("a", "b"), string.compare("b", "b"), string.compare("c", "b"));
    trace(string.contains("seafood", "foo"), string.contains("seafood", "bar"));
    trace(string.containsAny("failure", "ui"), string.containsAny("foo", ""));
    trace(string.count("cheese", "e"), string.count("five", ""));
    trace(string.equalFold("Go", "GO"));
    trace(string.hasPrefix("Gopher", "Go"), string.hasSuffix("Amigo", "go"));
    trace(string.index("chicken", "ken"), string.index("chicken", "dmr"));
    trace(string.indexAny("golang", "y"), string.indexAny("golang", "ly"));
    trace(string.lastIndex("go gopher", "go"), string.lastIndexAny("go gopher", "ordent"));
    trace(string.repeat("na", 2));
    trace(string.replace("oink oink oink", "k", "ky", 2));
    trace(string.replace("oink oink oink", "oink", "moo", -1));
    trace(string.replaceAll("oink oink oink", "oink", "moo"));
    trace(string.toLower("Gopher"), string.toUpper("Gopher"), string.toTitle("loud"));
    trace(string.toValidUTF8("a\tb", "?"));
    trace(string.trim("¡¡¡Hello, Gophers!!!", "!¡"));
    trace(string.trimLeft("xxhixx", "x"), string.trimRight("xxhixx", "x"));
    trace(string.trimPrefix("prefix-body", "prefix-"), string.trimSuffix("body.rs", ".rs"));
    trace(string.trimSpace("  hi  "), "|");
}
`
	out, code, err := run(t, src)
	if err != nil || code != 0 {
		t.Fatalf("err=%v code=%d", err, code)
	}
	want := `rune
-1 0 1
true false
true false
3 5
true
true true
4 -1
-1 2
3 8
nana
oinky oinky oink
moo moo moo
moo moo moo
gopher GOPHER LOUD
a	b
Hello, Gophers
hixx xxhi
body body
hi |
`
	if out != want {
		t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", out, want)
	}
}

// Positions are UTF-8 byte offsets, as in Go.
func TestStdStringByteOffsets(t *testing.T) {
	out, _, err := run(t, `package p;
use "std:string";
public void main() {
    trace(string.index("héllo", "l"), string.indexAny("héllo", "é"), string.count("héllo", ""));
}`)
	if err != nil || out != "3 1 6\n" {
		t.Errorf("got %q, %v", out, err)
	}
}

func TestStdStringRuntimeErrors(t *testing.T) {
	_, _, err := run(t, "package p;\nuse \"std:string\";\npublic void main() {\n    int n = -1;\n    trace(string.repeat(\"x\", n));\n}")
	var re *runtime.Error
	if !errors.As(err, &re) || re.Diag.Code != diagnostics.RunNativeFailed || re.Diag.Span.Start.String() != "5:11" {
		t.Fatalf("expected ASX-RUN-004 at 5:11, got %v", err)
	}
	if !strings.Contains(re.Diag.Message, "string.repeat failed") || !strings.Contains(re.Diag.Message, "negative Repeat count") {
		t.Errorf("message = %q", re.Diag.Message)
	}
}

func TestFunctionFeatures(t *testing.T) {
	cases := []struct{ name, src, out string }{
		{"callback", `package p;
public int apply(int(int, int) op, int a, int b) { return op(a, b); }
public int mul(int a, int b) { return a * b; }
public void main() {
    trace(apply(mul, 6, 7), apply((int a, int b) => a - b, 10, 4));
}`, "42 6\n"},
		{"nested and recursive", `package p;
public void main() {
    private int fib(int n) {
        if (n < 2) { return n; }
        return fib(n - 1) + fib(n - 2);
    }
    trace(fib(15));
}`, "610\n"},
		{"mutual recursion with prototype", `package p;
public void main() {
    private bool isEven(int n);
    private bool isOdd(int n) { if (n == 0) { return false; } return isEven(n - 1); }
    private bool isEven(int n) { if (n == 0) { return true; } return isOdd(n - 1); }
    trace(isEven(10), isOdd(7), isEven(3));
}`, "true true false\n"},
		{"top-level prototype", `package p;
public int twice(int x);
public void main() { trace(twice(21)); }
public int twice(int x) { return x * 2; }`, "42\n"},
		{"returning functions", `package p;
public int(int) pick(string op) {
    private int dbl(int n) { return n * 2; }
    if (op == "double") { return dbl; }
    return (int n) => n * n;
}
public void main() {
    trace(pick("double")(5), pick("square")(5));
}`, "10 25\n"},
		{"block lambda and void callbacks", `package p;
public void repeat(int times, void(int) body) {
    for (int i = 0; i < times; i++) { body(i); }
}
public void main() {
    repeat(3, (int i) => { trace("tick", i); });
    repeat(2, (int i) => trace("tock", i));
}`, "tick 0\ntick 1\ntick 2\ntock 0\ntock 1\n"},
		{"function variables", `package p;
public void main() {
    int(int) f = (int x) => x + 1;
    trace(f(1));
    f = (int x) => x * 10;
    trace(f(1));
    int(int)(int) adder = (int n) => (int x) => x + 2;
    trace(adder(0)(5));
}`, "2\n10\n7\n"},
		{"lambda early return", `package p;
public void main() {
    int(int) abs = (int x) => { if (x < 0) { return -x; } return x; };
    trace(abs(-4), abs(4));
}`, "4 4\n"},
		{"nested function has its own frame", `package p;
public void main() {
    int n = 1;
    private int f() { int n = 50; n++; return n; }
    trace(f(), f(), n);
}`, "51 51 1\n"},
		{"higher-order with std", `package p;
use "std:string";
public string each(string s, string(string) f) { return f(s); }
public void main() {
    trace(each("rune", (string s) => string.toUpper(s)));
}`, "RUNE\n"},
		{"callee evaluated before arguments", `package p;
public int(int) get() { trace("callee"); return (int x) => x; }
public int arg() { trace("arg"); return 1; }
public void main() { trace(get()(arg())); }`, "callee\narg\n1\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, _, err := run(t, c.src)
			if err != nil {
				t.Fatalf("runtime error: %v", err)
			}
			if out != c.out {
				t.Errorf("got %q, want %q", out, c.out)
			}
		})
	}
}

func TestLambdaRecursionDepthLimit(t *testing.T) {
	_, _, err := run(t, `package p;
public void loop() { (() => loop())(); }
public void main() { loop(); }`)
	var re *runtime.Error
	if !errors.As(err, &re) || re.Diag.Code != diagnostics.RunCallDepth {
		t.Fatalf("expected ASX-RUN-001, got %v", err)
	}
}
