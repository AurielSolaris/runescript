package stdlib

import (
	"fmt"
	"math"
	"strings"

	"github.com/AurielSolaris/runescript/internal/conv"
	"github.com/AurielSolaris/runescript/internal/types"
)

// std:matrix is a library of functions over Array<Array<float64>>, an
// array of rows (OQ-TYPE-03a: no matrix type, no operators). Every
// function checks that its matrices are rectangular (all rows the same
// length) and that their sizes fit the operation, and throws an
// ArgumentError otherwise. Results are new arrays; arguments are never
// changed. A matrix may have no rows (0x0) or rows with no columns (nx0).
//
// Both targets must give bit-identical results, so every function is one
// fixed sequence of IEEE 754 operations, written the same way in the
// JavaScript twin (section matrix in internal/jsgen/runtime.js): sums run
// from index 0 up, starting at 0, and every product added to a sum is
// wrapped in float64(...) so the Go compiler cannot fuse the two into one
// multiply-add (matrix_test.go checks this). det, inverse, and solve use
// Gaussian elimination with partial pivoting: the pivot is the first
// entry of largest absolute value at or below the diagonal.
func init() {
	F, I, S := types.Float64, types.Int, types.StringT
	V := types.NewArray(F)
	M := types.NewArray(V)
	m := func(a any) [][]float64 { return a.([][]float64) }
	p := &Package{Path: "matrix", Qualifier: "matrix"}
	p.add(
		fn("zeros", M, func(a []any) any {
			return mZeros(mSize(a[0].(int64), "the number of rows"), mSize(a[1].(int64), "the number of columns"))
		}, I, I),
		fn("identity", M, func(a []any) any { return mIdentity(mSize(a[0].(int64), "the size")) }, I),
		fn("rows", I, func(a []any) any { r, _ := mShape(m(a[0])); return int64(r) }, M),
		fn("cols", I, func(a []any) any { _, c := mShape(m(a[0])); return int64(c) }, M),
		fn("add", M, func(a []any) any { return mAddSub(m(a[0]), m(a[1]), false) }, M, M),
		fn("sub", M, func(a []any) any { return mAddSub(m(a[0]), m(a[1]), true) }, M, M),
		fn("scale", M, func(a []any) any { return mScale(m(a[0]), a[1].(float64)) }, M, F),
		fn("mul", M, func(a []any) any { return mMul(m(a[0]), m(a[1])) }, M, M),
		fn("mulVec", V, func(a []any) any { return mMulVec(m(a[0]), a[1].([]float64)) }, M, V),
		fn("transpose", M, func(a []any) any { return mTranspose(m(a[0])) }, M),
		fn("det", F, func(a []any) any { return mDet(m(a[0])) }, M),
		fn("inverse", M, func(a []any) any { return mInverse(m(a[0])) }, M),
		fn("solve", V, func(a []any) any { return mSolve(m(a[0]), a[1].([]float64)) }, M, V),
		fn("format", S, func(a []any) any { return mFormat(m(a[0])) }, M),
	)
	register(p)
}

func argFault(format string, args ...any) {
	panic(Fault{Type: "ArgumentError", Message: fmt.Sprintf(format, args...)})
}

// mSize checks a size argument.
func mSize(n int64, what string) int {
	if n < 0 {
		argFault("%s cannot be negative (%d)", what, n)
	}
	return int(n)
}

// mCount is "1 row", "2 rows", and so on.
func mCount(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// mShape returns the size of m, which must be rectangular.
func mShape(m [][]float64) (rows, cols int) {
	if len(m) == 0 {
		return 0, 0
	}
	cols = len(m[0])
	for i, row := range m {
		if len(row) != cols {
			argFault("the matrix is not rectangular: row %d has %s, row 0 has %d", i, mCount(len(row), "column"), cols)
		}
	}
	return len(m), cols
}

// mSquare returns the size of m, which must be square.
func mSquare(m [][]float64) int {
	r, c := mShape(m)
	if r != c {
		argFault("the matrix must be square, not %dx%d", r, c)
	}
	return r
}

func mZeros(rows, cols int) [][]float64 {
	out := make([][]float64, rows)
	for i := range out {
		out[i] = make([]float64, cols)
	}
	return out
}

func mIdentity(n int) [][]float64 {
	out := mZeros(n, n)
	for i := range out {
		out[i][i] = 1
	}
	return out
}

func mCopy(m [][]float64) [][]float64 {
	out := make([][]float64, len(m))
	for i, row := range m {
		out[i] = append([]float64{}, row...)
	}
	return out
}

func mAddSub(a, b [][]float64, sub bool) [][]float64 {
	ar, ac := mShape(a)
	br, bc := mShape(b)
	if ar != br || ac != bc {
		verb := "add"
		if sub {
			verb = "subtract"
		}
		argFault("cannot %s a %dx%d matrix and a %dx%d matrix", verb, ar, ac, br, bc)
	}
	out := mZeros(ar, ac)
	for i := range out {
		for j := range out[i] {
			if sub {
				out[i][j] = a[i][j] - b[i][j]
			} else {
				out[i][j] = a[i][j] + b[i][j]
			}
		}
	}
	return out
}

func mScale(m [][]float64, k float64) [][]float64 {
	r, c := mShape(m)
	out := mZeros(r, c)
	for i := range out {
		for j := range out[i] {
			out[i][j] = m[i][j] * k
		}
	}
	return out
}

func mMul(a, b [][]float64) [][]float64 {
	ar, ac := mShape(a)
	br, bc := mShape(b)
	if ac != br {
		argFault("cannot multiply a %dx%d matrix by a %dx%d matrix: the first has %s, the second %s", ar, ac, br, bc, mCount(ac, "column"), mCount(br, "row"))
	}
	out := mZeros(ar, bc)
	for i := 0; i < ar; i++ {
		for j := 0; j < bc; j++ {
			s := 0.0
			for k := 0; k < ac; k++ {
				s += float64(a[i][k] * b[k][j])
			}
			out[i][j] = s
		}
	}
	return out
}

func mMulVec(m [][]float64, v []float64) []float64 {
	r, c := mShape(m)
	if len(v) != c {
		argFault("cannot multiply a %dx%d matrix by a vector of length %d", r, c, len(v))
	}
	out := make([]float64, r)
	for i := 0; i < r; i++ {
		s := 0.0
		for k := 0; k < c; k++ {
			s += float64(m[i][k] * v[k])
		}
		out[i] = s
	}
	return out
}

func mTranspose(m [][]float64) [][]float64 {
	r, c := mShape(m)
	out := mZeros(c, r)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			out[j][i] = m[i][j]
		}
	}
	return out
}

// mPivot returns the row, from k down, whose entry in column k has the
// largest absolute value (the first such row).
func mPivot(a [][]float64, k int) int {
	p := k
	for i := k + 1; i < len(a); i++ {
		if math.Abs(a[i][k]) > math.Abs(a[p][k]) {
			p = i
		}
	}
	return p
}

func mDet(m [][]float64) float64 {
	n := mSquare(m)
	a := mCopy(m)
	det := 1.0
	for k := 0; k < n; k++ {
		p := mPivot(a, k)
		if a[p][k] == 0 {
			return 0
		}
		if p != k {
			a[p], a[k] = a[k], a[p]
			det = -det
		}
		det *= a[k][k]
		for i := k + 1; i < n; i++ {
			f := a[i][k] / a[k][k]
			for j := k + 1; j < n; j++ {
				a[i][j] -= float64(f * a[k][j])
			}
		}
	}
	return det
}

func singular() {
	panic(Fault{Type: "Error", Message: "the matrix is singular"})
}

func mInverse(m [][]float64) [][]float64 {
	n := mSquare(m)
	a := mCopy(m)
	inv := mIdentity(n)
	for k := 0; k < n; k++ {
		p := mPivot(a, k)
		if a[p][k] == 0 {
			singular()
		}
		a[p], a[k] = a[k], a[p]
		inv[p], inv[k] = inv[k], inv[p]
		piv := a[k][k]
		for j := 0; j < n; j++ {
			a[k][j] /= piv
			inv[k][j] /= piv
		}
		for i := 0; i < n; i++ {
			if i == k {
				continue
			}
			f := a[i][k]
			for j := 0; j < n; j++ {
				a[i][j] -= float64(f * a[k][j])
				inv[i][j] -= float64(f * inv[k][j])
			}
		}
	}
	return inv
}

func mSolve(m [][]float64, b []float64) []float64 {
	n := mSquare(m)
	if len(b) != n {
		argFault("cannot solve a %dx%d system with a right-hand side of length %d", n, n, len(b))
	}
	a := mCopy(m)
	y := append([]float64{}, b...)
	for k := 0; k < n; k++ {
		p := mPivot(a, k)
		if a[p][k] == 0 {
			singular()
		}
		a[p], a[k] = a[k], a[p]
		y[p], y[k] = y[k], y[p]
		for i := k + 1; i < n; i++ {
			f := a[i][k] / a[k][k]
			for j := k + 1; j < n; j++ {
				a[i][j] -= float64(f * a[k][j])
			}
			y[i] -= float64(f * y[k])
		}
	}
	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		s := y[i]
		for j := i + 1; j < n; j++ {
			s -= float64(a[i][j] * x[j])
		}
		x[i] = s / a[i][i]
	}
	return x
}

// mFormat writes one row per line, as [1, 2.5, -3].
func mFormat(m [][]float64) string {
	mShape(m)
	var b strings.Builder
	for i, row := range m {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteByte('[')
		for j, x := range row {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(conv.FormatFloat(x, 64))
		}
		b.WriteByte(']')
	}
	return b.String()
}
