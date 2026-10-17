package tests

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// TestWideningMatches runs random nested int and uint expressions, whose
// intermediate results often leave the type's range (OQ-NUM-01, option A),
// in the interpreter and on Node.js, and requires identical output: the
// same results, and the same overflow and division errors, including
// intermediates beyond int64.
func TestWideningMatches(t *testing.T) {
	nodePath := node(t)
	r := rand.New(rand.NewSource(11))
	operand := map[string][]string{
		"int":  {"2147483647", "-2147483648", "65536", "-3", "0", "7", "1000000"},
		"uint": {"4294967295", "65536", "0", "3", "1000000"},
	}
	var expr func(ty string, depth int) string
	expr = func(ty string, depth int) string {
		if depth == 0 || r.Intn(4) == 0 {
			vals := operand[ty]
			return fmt.Sprintf("((%s) %q)", ty, vals[r.Intn(len(vals))]) // not folded
		}
		op := []string{"+", "-", "*", "/", "%"}[r.Intn(5)]
		e := "(" + expr(ty, depth-1) + " " + op + " " + expr(ty, depth-1) + ")"
		if ty == "int" && r.Intn(6) == 0 {
			e = "(-" + e + ")"
		}
		return e
	}
	var b strings.Builder
	b.WriteString("package fuzz;\n\nprotected int main() {\n")
	for i := 0; i < 400; i++ {
		ty := []string{"int", "uint"}[r.Intn(2)]
		e := expr(ty, 1+r.Intn(4))
		var stmt string
		switch r.Intn(4) {
		case 0:
			stmt = fmt.Sprintf("%s v = %s; trace(%d, v);", ty, e, i)
		case 1:
			stmt = fmt.Sprintf("trace(%d, (int64) %s);", i, e)
		case 2:
			stmt = fmt.Sprintf("trace(%d, %s < %s, %s == %s);", i, e, expr(ty, 2), e, e)
		default:
			stmt = fmt.Sprintf("trace(%d, (float64) %s, (string) %s);", i, e, e)
		}
		fmt.Fprintf(&b, "    try {\n        %s\n    } catch (Error e) {\n        trace(%d, e.name, e.message);\n    }\n", stmt, i)
	}
	b.WriteString("    return 0;\n}\n")
	src := []byte(b.String())

	want := interpret(t, src)
	got := runJS(t, nodePath, src)
	if got != want {
		t.Fatalf("JavaScript differs from the interpreter\ninterpreter:\n%s\nnode:\n%s", want.out, got.out)
	}
	if n := strings.Count(want.out, "\n"); n < 400 {
		t.Fatalf("only %d lines of output:\n%s", n, want.out)
	}
	if !strings.Contains(want.out, "int64 range") || !strings.Contains(want.out, "does not fit") {
		t.Errorf("the random program never overflowed an intermediate and a result; adjust the operands")
	}
}
