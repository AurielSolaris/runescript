package stdlib

import (
	"math"

	"github.com/AurielSolaris/runescript/internal/types"
)

// std:math mirrors Go's math package (OQ-MATH-01): camelCase names on
// float64, untyped constants (math.pi), and the int helpers abs, min, max,
// and clamp. Other float types convert explicitly: math.sqrt((float64) f).
//
// Every function here must give bit-identical results on both targets, so
// the JavaScript runtime implements each one exactly as Go does, never
// with a host function that merely approximates the same value. The
// functions whose results IEEE 754 defines exactly (sqrt, floor, round,
// mod, ...) are here; the transcendental ones (sin, exp, log, pow, ...)
// need one shared implementation for both targets and come next.
//
// A NaN's sign is not portable across hosts, so it is not observable:
// signbit(NaN) is false, and copysign treats a NaN sign as positive.
func init() {
	F, I, B := types.Float64, types.Int, types.BoolT
	f := func(a any) float64 { return a.(float64) }
	n := func(a any) int64 { return a.(int64) }
	f1 := func(name string, g func(float64) float64) *Func {
		return fn(name, F, func(a []any) any { return g(f(a[0])) }, F)
	}
	f2 := func(name string, g func(float64, float64) float64) *Func {
		return fn(name, F, func(a []any) any { return g(f(a[0]), f(a[1])) }, F, F)
	}
	// The int helpers overload abs, min, and max: calls try the int
	// signature first, so math.max(3, 7) is an int.
	intFn := func(name string, impl func(a []any) any, params ...*types.Type) *Func {
		g := fn(name, I, impl, params...)
		g.JSName = name + "Int"
		return g
	}

	p := &Package{Path: "math", Qualifier: "math"}
	p.add(
		intFn("abs", func(a []any) any {
			if v := n(a[0]); v < 0 {
				return -v // abs(minimum int) does not fit: the runtime reports it
			} else {
				return v
			}
		}, I),
		intFn("min", func(a []any) any { return min(n(a[0]), n(a[1])) }, I, I),
		intFn("max", func(a []any) any { return max(n(a[0]), n(a[1])) }, I, I),
		intFn("clamp", func(a []any) any {
			x, lo, hi := n(a[0]), n(a[1]), n(a[2])
			if lo > hi {
				panic(Fault{Type: "ArgumentError", Message: "lo is greater than hi"})
			}
			return min(max(x, lo), hi)
		}, I, I, I),

		f1("abs", math.Abs),
		f1("ceil", math.Ceil),
		f1("floor", math.Floor),
		f1("trunc", math.Trunc),
		f1("round", math.Round),
		f1("roundToEven", math.RoundToEven),
		f1("sqrt", math.Sqrt),
		f2("min", math.Min),
		f2("max", math.Max),
		f2("mod", math.Mod),
		f2("remainder", math.Remainder),
		f2("dim", math.Dim),
		f2("nextafter", math.Nextafter),
		f2("copysign", func(x, sign float64) float64 {
			if math.IsNaN(sign) {
				sign = 1
			}
			return math.Copysign(x, sign)
		}),
		fn("signbit", B, func(a []any) any { return !math.IsNaN(f(a[0])) && math.Signbit(f(a[0])) }, F),
		fn("isInf", B, func(a []any) any { return math.IsInf(f(a[0]), int(n(a[1]))) }, F, I),
		fn("inf", F, func(a []any) any { return math.Inf(int(n(a[0]))) }, I),
		fn("nan", F, func(a []any) any { return math.NaN() }),
	)

	flt := func(v float64) Const { return Const{Float: v} }
	ints := func(t *types.Type) (Const, Const) {
		lo, hi := types.IntRange(t)
		return Const{Int: lo}, Const{Int: hi}
	}
	for name, v := range map[string]float64{
		"e": math.E, "pi": math.Pi, "phi": math.Phi,
		"sqrt2": math.Sqrt2, "sqrtE": math.SqrtE, "sqrtPi": math.SqrtPi, "sqrtPhi": math.SqrtPhi,
		"ln2": math.Ln2, "log2E": math.Log2E, "ln10": math.Ln10, "log10E": math.Log10E,
		"maxFloat64": math.MaxFloat64, "smallestNonzeroFloat64": math.SmallestNonzeroFloat64,
		"maxFloat32": math.MaxFloat32, "smallestNonzeroFloat32": math.SmallestNonzeroFloat32,
	} {
		p.Consts[name] = flt(v)
	}
	for name, t := range map[string]*types.Type{
		"Int": types.Int, "Int8": types.Int8, "Int16": types.Int16, "Int32": types.Int32, "Int64": types.Int64,
	} {
		p.Consts["min"+name], p.Consts["max"+name] = ints(t)
	}
	for name, t := range map[string]*types.Type{
		"Uint": types.Uint, "Uint8": types.Uint8, "Uint16": types.Uint16, "Uint32": types.Uint32, "Uint64": types.Uint64,
	} {
		_, p.Consts["max"+name] = ints(t)
	}
	register(p)
}
