package stdlib

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"math"
	"reflect"
	"testing"
)

func TestMatrix(t *testing.T) {
	a := [][]float64{{1, 2}, {3, 4}}
	b := [][]float64{{5, 6}, {7, 8}}
	eq := func(name string, got, want any) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
	}
	eq("zeros", mZeros(2, 3), [][]float64{{0, 0, 0}, {0, 0, 0}})
	eq("zeros(0,0)", mZeros(0, 0), [][]float64{})
	eq("identity", mIdentity(2), [][]float64{{1, 0}, {0, 1}})
	eq("add", mAddSub(a, b, false), [][]float64{{6, 8}, {10, 12}})
	eq("sub", mAddSub(a, b, true), [][]float64{{-4, -4}, {-4, -4}})
	eq("scale", mScale(a, 0.5), [][]float64{{0.5, 1}, {1.5, 2}})
	eq("mul", mMul(a, b), [][]float64{{19, 22}, {43, 50}})
	eq("mul 2x3 3x1", mMul([][]float64{{1, 2, 3}, {4, 5, 6}}, [][]float64{{1}, {0}, {-1}}), [][]float64{{-2}, {-2}})
	eq("mul 2x0 0x0", mMul([][]float64{{}, {}}, [][]float64{}), [][]float64{{}, {}})
	eq("mulVec", mMulVec(a, []float64{1, 1}), []float64{3, 7})
	eq("transpose", mTranspose([][]float64{{1, 2, 3}, {4, 5, 6}}), [][]float64{{1, 4}, {2, 5}, {3, 6}})
	eq("det", mDet(a), -2.0)
	eq("det 0x0", mDet([][]float64{}), 1.0)
	eq("det singular", mDet([][]float64{{1, 2}, {2, 4}}), 0.0)
	eq("det 3x3", mDet([][]float64{{2, 0, 1}, {1, 3, 2}, {4, 1, 2}}), -3.0)
	eq("inverse", mInverse([][]float64{{4, 7}, {2, 6}}), [][]float64{{0.6000000000000001, -0.7000000000000001}, {-0.2, 0.4}})
	eq("solve", mSolve([][]float64{{2, 1}, {1, 3}}, []float64{3, 5}), []float64{0.8, 1.4})
	eq("format", mFormat([][]float64{{1, 2.5}, {-0.0, math.Inf(1)}}), "[1, 2.5]\n[0, Infinity]")
	eq("format 0x0", mFormat([][]float64{}), "")
	eq("a unchanged", a, [][]float64{{1, 2}, {3, 4}})

	faults := []struct {
		f    func()
		typ  string
		want string
	}{
		{func() { mSize(-1, "the number of rows") }, "ArgumentError", "the number of rows cannot be negative (-1)"},
		{func() { mShape([][]float64{{1, 2}, {3}}) }, "ArgumentError", "the matrix is not rectangular: row 1 has 1 column, row 0 has 2"},
		{func() { mAddSub(a, [][]float64{{1}}, false) }, "ArgumentError", "cannot add a 2x2 matrix and a 1x1 matrix"},
		{func() { mAddSub(a, [][]float64{{1}}, true) }, "ArgumentError", "cannot subtract a 2x2 matrix and a 1x1 matrix"},
		{func() { mMul(a, [][]float64{{1, 2}}) }, "ArgumentError", "cannot multiply a 2x2 matrix by a 1x2 matrix: the first has 2 columns, the second 1 row"},
		{func() { mMulVec(a, []float64{1}) }, "ArgumentError", "cannot multiply a 2x2 matrix by a vector of length 1"},
		{func() { mDet([][]float64{{1, 2}}) }, "ArgumentError", "the matrix must be square, not 1x2"},
		{func() { mInverse([][]float64{{1, 2}, {2, 4}}) }, "Error", "the matrix is singular"},
		{func() { mSolve([][]float64{{0, 0}, {0, 0}}, []float64{1, 2}) }, "Error", "the matrix is singular"},
		{func() { mSolve(a, []float64{1}) }, "ArgumentError", "cannot solve a 2x2 system with a right-hand side of length 1"},
	}
	for _, c := range faults {
		if f := fault(c.f); f.Type != c.typ || f.Message != c.want {
			t.Errorf("got %s %q, want %s %q", f.Type, f.Message, c.typ, c.want)
		}
	}
}

// TestMatrixHasNoFusableProducts checks that matrix.go wraps every float
// product that feeds an addition or subtraction (including += and -=) in
// float64(...): otherwise the Go compiler may fuse the two into one
// multiply-add on arm64, and the results would differ from JavaScript's.
func TestMatrixHasNoFusableProducts(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "matrix.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}
	conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil), Error: func(error) {}}
	conf.Check("stdlib", fset, []*ast.File{f}, info) // errors: the rest of the package is missing
	isFloatMul := func(e ast.Expr) bool {
		for {
			p, ok := e.(*ast.ParenExpr)
			if !ok {
				break
			}
			e = p.X
		}
		b, ok := e.(*ast.BinaryExpr)
		if !ok || b.Op != token.MUL {
			return false
		}
		basic, ok := info.Types[b].Type.(*types.Basic)
		return ok && basic.Info()&types.IsFloat != 0
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BinaryExpr:
			if (n.Op == token.ADD || n.Op == token.SUB) && (isFloatMul(n.X) || isFloatMul(n.Y)) {
				t.Errorf("%s: a float product feeds %s; wrap it in float64(...)", fset.Position(n.Pos()), n.Op)
			}
		case *ast.AssignStmt:
			if (n.Tok == token.ADD_ASSIGN || n.Tok == token.SUB_ASSIGN) && isFloatMul(n.Rhs[0]) {
				t.Errorf("%s: a float product feeds %s; wrap it in float64(...)", fset.Position(n.Pos()), n.Tok)
			}
		}
		return true
	})
}
