package tests

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/types"
)

// TestIntegerWidthsMatch runs random arithmetic, comparisons, and
// conversions on every integer type, with operands near each type's
// bounds, in the interpreter and on Node.js, and requires identical
// output: the same results, and the same overflow, division, and
// conversion errors.
func TestIntegerWidthsMatch(t *testing.T) {
	nodePath := node(t)
	ints := []*types.Type{types.Int, types.Uint, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint8, types.Uint16, types.Uint32, types.Uint64}
	r := rand.New(rand.NewSource(7))
	pick := func(ty *types.Type) string {
		lo, hi := types.IntRange(ty)
		switch r.Intn(5) {
		case 0:
			return lo.String()
		case 1:
			return hi.String()
		case 2:
			return fmt.Sprint(r.Intn(7) - 3)
		}
		// anywhere in range
		span := new(big.Int).Sub(hi, lo)
		v := new(big.Int).Rand(r, span.Add(span, big.NewInt(1)))
		return v.Add(v, lo).String()
	}
	var b strings.Builder
	b.WriteString("package fuzz;\n\nprotected int main() {\n")
	n := 0
	step := func(stmt string) {
		n++
		fmt.Fprintf(&b, "    try {\n        %s\n    } catch (Error e) {\n        trace(%d, e.name, e.message);\n    }\n", stmt, n)
	}
	for _, ty := range ints {
		for i := 0; i < 40; i++ {
			x, y := pick(ty), pick(ty)
			// Operands come from strings, so no constant is folded.
			decl := fmt.Sprintf("%s x = (%s) %q; %s y = (%s) %q;", ty, ty, x, ty, ty, y)
			op := []string{"+", "-", "*", "/", "%"}[r.Intn(5)]
			step(fmt.Sprintf("%s trace(%d, x %s y, x < y, x == y);", decl, n+1, op))
			bop := []string{"&", "|", "^"}[r.Intn(3)]
			step(fmt.Sprintf("%s trace(%d, x %s y, ~x, ~y);", decl, n+1, bop))
			// Counts from -1 to the width, so out-of-range ones are covered.
			count := r.Intn(types.IntBits(ty)+2) - 1
			sop := []string{"<<", ">>"}[r.Intn(2)]
			step(fmt.Sprintf("%s int k = (int) %q; trace(%d, x %s k, y >> (uint8) 1);", decl, fmt.Sprint(count), n+1, sop))
			to := ints[r.Intn(len(ints))]
			step(fmt.Sprintf("%s trace(%d, (%s) x, (float64) x, (float32) y, (string) x);", decl, n+1, to))
		}
	}
	b.WriteString("    return 0;\n}\n")
	src := []byte(b.String())

	want := interpret(t, src)
	got := runJS(t, nodePath, src)
	if got != want {
		t.Fatalf("JavaScript differs from the interpreter\ninterpreter:\n%s\nnode:\n%s", want.out, got.out)
	}
	if strings.Count(want.out, "\n") < 700 {
		t.Fatalf("only %d lines of output; the program did not run as intended:\n%s", strings.Count(want.out, "\n"), want.out)
	}
}
