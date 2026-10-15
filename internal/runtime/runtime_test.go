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
		{"exit code", "package p;\nint main() { return 7; }", "", 7},
		{"void main", "package p;\nvoid main() { trace(\"hi\"); }", "hi\n", 0},
		{"trace joins with spaces", "package p;\nvoid main() { trace(\"a\", 1, -2, true, false); }", "a 1 -2 true false\n", 0},
		{"trace no args", "package p;\nvoid main() { trace(); }", "\n", 0},
		{"string escapes", "package p;\nvoid main() { trace(\"q\\\"\\\\\"); }", "q\"\\\n", 0},
		{"arithmetic", "package p;\nvoid main() { int x = 7; trace(x + 2, x - 9, x * 3, x / 2, x % 4, -x); }", "9 -2 21 3 3 -7\n", 0},
		{"truncating division", "package p;\nvoid main() { int x = -7; trace(x / 2, x % 2); }", "-3 -1\n", 0},
		{"precedence", "package p;\nvoid main() { trace(2 + 3 * 4, (2 + 3) * 4); }", "14 20\n", 0},
		{"comparisons", "package p;\nvoid main() { int a = 1; int b = 2; trace(a < b, a <= a, a > b, b >= a, a == b, a != b); }", "true true false true false true\n", 0},
		{"string equality", "package p;\nvoid main() { string s = \"x\"; trace(s == \"x\", s != \"x\"); }", "true false\n", 0},
		{"assignment", "package p;\nvoid main() { int x = 1; x = 5; x += 2; x -= 1; x *= 3; x /= 2; x %= 5; trace(x); }", "4\n", 0},
		{"inc dec", "package p;\nvoid main() { int x = 0; x++; x++; x--; trace(x); }", "1\n", 0},
		{"if else chain", "package p;\nvoid f(int n) { if (n < 0) { trace(\"neg\"); } else if (n == 0) { trace(\"zero\"); } else { trace(\"pos\"); } }\nvoid main() { f(-1); f(0); f(1); }", "neg\nzero\npos\n", 0},
		{"while", "package p;\nvoid main() { int i = 0; while (i < 3) { trace(i); i++; } }", "0\n1\n2\n", 0},
		{"for", "package p;\nvoid main() { for (int i = 3; i > 0; i--) { trace(i); } }", "3\n2\n1\n", 0},
		{"break continue", "package p;\nvoid main() { for (int i = 0; i < 10; i++) { if (i == 1) { continue; } if (i == 3) { break; } trace(i); } }", "0\n2\n", 0},
		{"for without clauses", "package p;\nvoid main() { int i = 0; for (;;) { i++; if (i == 4) { break; } } trace(i); }", "4\n", 0},
		{"while continue", "package p;\nvoid main() { int i = 0; while (i < 4) { i++; if (i % 2 == 0) { continue; } trace(i); } }", "1\n3\n", 0},
		{"nested loops break inner", "package p;\nvoid main() { for (int i = 0; i < 2; i++) { for (int j = 0; j < 5; j++) { if (j == 2) { break; } trace(i, j); } } }", "0 0\n0 1\n1 0\n1 1\n", 0},
		{"return from loop", "package p;\nint find() { for (int i = 0; i < 100; i++) { if (i * i > 50) { return i; } } return -1; }\nvoid main() { trace(find()); }", "8\n", 0},
		{"short circuit", "package p;\nbool boom() { trace(\"boom\"); return true; }\nvoid main() { trace(false && boom(), true || boom()); }", "false true\n", 0},
		{"recursion", "package p;\nint fact(int n) { if (n <= 1) { return 1; } return n * fact(n - 1); }\nint main() { trace(fact(10)); return 0; }", "3628800\n", 0},
		{"arguments are copied", "package p;\nint inc(int n) { int m = n; m++; return m; }\nvoid main() { int x = 1; trace(inc(x), x); }", "2 1\n", 0},
		{"scopes reset per iteration", "package p;\nvoid main() { for (int i = 0; i < 2; i++) { int y = i * 10; trace(y); } }", "0\n10\n", 0},
		{"uint max", "package p;\nvoid main() { uint u = 4294967295; trace(u); }", "4294967295\n", 0},
		{"int bounds", "package p;\nvoid main() { int lo = -2147483648; int hi = 2147483647; trace(lo, hi); }", "-2147483648 2147483647\n", 0},
		{"sint same as int", "package p;\nint main() { sint s = -3; int i = s; return i + 3; }", "", 0},
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
		{"int overflow", "package p;\nvoid main() {\n    int x = 2147483647;\n    x = x + 1;\n}", diagnostics.RunOverflow, "4:11"},
		{"int overflow increment", "package p;\nvoid main() {\n    int x = 2147483647;\n    x++;\n}", diagnostics.RunOverflow, "4:5"},
		{"int underflow", "package p;\nvoid main() {\n    int x = -2147483648;\n    x -= 1;\n}", diagnostics.RunOverflow, "4:5"},
		{"negate min int", "package p;\nvoid main() {\n    int x = -2147483648;\n    trace(-x);\n}", diagnostics.RunOverflow, "4:11"},
		{"mul overflow", "package p;\nvoid main() {\n    int x = 65536;\n    trace(x * x);\n}", diagnostics.RunOverflow, "4:13"},
		{"uint below zero", "package p;\nvoid main() {\n    uint u = 0;\n    u--;\n}", diagnostics.RunOverflow, "4:5"},
		{"uint overflow", "package p;\nvoid main() {\n    uint u = 4294967295;\n    trace(u + 1);\n}", diagnostics.RunOverflow, "4:13"},
		{"divide by zero", "package p;\nvoid main() {\n    int z = 0;\n    trace(1 / z);\n}", diagnostics.RunDivideByZero, "4:13"},
		{"modulo by zero", "package p;\nvoid main() {\n    int z = 0;\n    trace(1 % z);\n}", diagnostics.RunDivideByZero, "4:13"},
		{"infinite recursion", "package p;\nvoid f() {\n    f();\n}\nvoid main() { f(); }", diagnostics.RunCallDepth, "3:5"},
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
	out, _, err := run(t, "package p;\nvoid main() { trace(\"before\"); int z = 0; trace(1 / z); }")
	if err == nil || out != "before\n" {
		t.Errorf("got %q, %v", out, err)
	}
}
