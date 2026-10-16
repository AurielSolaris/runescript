package fmath

import (
	"math"
	"math/rand"
	"sort"
	"testing"
)

// The copies must compute what Go's own math computes, up to a few units in
// the last place: on amd64, Go's math uses assembly for Exp and Log (and
// so for the functions built on them), which is exactly why RuneScript
// cannot use it. A copying mistake would be far larger.
func TestMatchesGoMath(t *testing.T) {
	f1 := map[string][2]func(float64) float64{
		"Sin": {Sin, math.Sin}, "Cos": {Cos, math.Cos}, "Tan": {Tan, math.Tan},
		"Asin": {Asin, math.Asin}, "Acos": {Acos, math.Acos}, "Atan": {Atan, math.Atan},
		"Sinh": {Sinh, math.Sinh}, "Cosh": {Cosh, math.Cosh}, "Tanh": {Tanh, math.Tanh},
		"Asinh": {Asinh, math.Asinh}, "Acosh": {Acosh, math.Acosh}, "Atanh": {Atanh, math.Atanh},
		"Exp": {Exp, math.Exp}, "Exp2": {Exp2, math.Exp2}, "Expm1": {Expm1, math.Expm1},
		"Log": {Log, math.Log}, "Log10": {Log10, math.Log10}, "Log2": {Log2, math.Log2},
		"Log1p": {Log1p, math.Log1p}, "Cbrt": {Cbrt, math.Cbrt},
	}
	names := make([]string, 0, len(f1))
	for name := range f1 {
		names = append(names, name)
	}
	sort.Strings(names) // a fixed order, so the inputs are the same on every run
	r := rand.New(rand.NewSource(1))
	for _, name := range names {
		f := f1[name]
		exact, ulp := 0, 0
		for i := 0; i < 5000; i++ {
			x := math.Ldexp(r.Float64()*2-1, r.Intn(80)-40)
			a, b := f[0](x), f[1](x)
			switch {
			case a == b || (math.IsNaN(a) && math.IsNaN(b)):
				exact++
			case math.Abs(a-b) <= 4*math.Abs(math.Nextafter(b, math.Inf(1))-b):
				ulp++
			case math.IsInf(b, 0) && math.Abs(a) > 1e307:
				// Near the overflow threshold, Go's assembly Exp overflows
				// a little early; the pure-Go result is the finite one.
				ulp++
			default:
				t.Fatalf("%s(%v) = %v, Go's math gives %v", name, x, a, b)
			}
		}
		if ulp > 0 {
			t.Logf("%s: %d of 5000 differ from Go's math in the last bit", name, ulp)
		}
	}
	close := func(a, b float64) bool {
		return a == b || (math.IsNaN(a) && math.IsNaN(b)) || math.Abs(a-b) <= 4*math.Abs(math.Nextafter(b, math.Inf(1))-b)
	}
	for i := 0; i < 5000; i++ {
		x, y := r.Float64()*20-10, r.Float64()*20-10
		if a, b := Pow(x, y), math.Pow(x, y); !close(a, b) {
			t.Fatalf("Pow(%v, %v) = %v, Go gives %v", x, y, a, b)
		}
		if a, b := Atan2(x, y), math.Atan2(x, y); a != b {
			t.Fatalf("Atan2(%v, %v) = %v, Go gives %v", x, y, a, b)
		}
		if a, b := Hypot(x, y), math.Hypot(x, y); a != b {
			t.Fatalf("Hypot(%v, %v) = %v, Go gives %v", x, y, a, b)
		}
	}
}
