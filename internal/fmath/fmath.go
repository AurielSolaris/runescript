// Package fmath holds the pure-Go implementations of the transcendental
// functions of Go's math package (sin, exp, log, pow, ...), copied from
// the Go source (see LICENSE-GO).
//
// RuneScript promises bit-identical results on the interpreter and the
// JavaScript target (OQ-MATH-01). Go's math may use architecture-specific
// assembly for these functions, and JavaScript's Math functions are
// implemented differently by every engine, so neither can be used
// directly. Instead, both targets run this one algorithm: the interpreter
// calls this package, and the JavaScript runtime (internal/jsgen,
// runtime.js) is a line-by-line port of it.
//
// One change was made to the copied code: every product that is an
// operand of a sum or difference is wrapped in float64(...), which the Go
// specification guarantees is rounded on its own. That stops the compiler
// from fusing x*y + z into one fused multiply-add on processors that have
// it (arm64, ppc64, s390x, riscv64), which would change the last bit.
// TestNoFusableProducts keeps it that way.
//
// The functions IEEE 754 defines exactly (Sqrt, Floor, Frexp, ...) give
// the same result however they are computed, so they are Go's.
package fmath

import "math"

const (
	uvnan    = 0x7FF8000000000001
	uvinf    = 0x7FF0000000000000
	uvneginf = 0xFFF0000000000000
	uvone    = 0x3FF0000000000000
	mask     = 0x7FF
	shift    = 64 - 11 - 1
	bias     = 1023
	signMask = 1 << 63
	fracMask = 1<<shift - 1
)

const (
	Pi                     = math.Pi
	Ln2                    = math.Ln2
	Log2E                  = math.Log2E
	Ln10                   = math.Ln10
	Sqrt2                  = math.Sqrt2
	MaxFloat64             = math.MaxFloat64
	SmallestNonzeroFloat64 = math.SmallestNonzeroFloat64
)

func Abs(x float64) float64               { return math.Abs(x) }
func Sqrt(x float64) float64              { return math.Sqrt(x) }
func Floor(x float64) float64             { return math.Floor(x) }
func Ceil(x float64) float64              { return math.Ceil(x) }
func Trunc(x float64) float64             { return math.Trunc(x) }
func Mod(x, y float64) float64            { return math.Mod(x, y) }
func Modf(f float64) (float64, float64)   { return math.Modf(f) }
func Frexp(f float64) (float64, int)      { return math.Frexp(f) }
func Ldexp(frac float64, exp int) float64 { return math.Ldexp(frac, exp) }
func IsNaN(x float64) bool                { return math.IsNaN(x) }
func IsInf(x float64, sign int) bool      { return math.IsInf(x, sign) }
func Inf(sign int) float64                { return math.Inf(sign) }
func NaN() float64                        { return math.NaN() }
func Copysign(f, sign float64) float64    { return math.Copysign(f, sign) }
func Signbit(x float64) bool              { return math.Signbit(x) }
func Float64bits(f float64) uint64        { return math.Float64bits(f) }
func Float64frombits(b uint64) float64    { return math.Float64frombits(b) }
