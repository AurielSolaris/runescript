package fmath

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fusable returns the float64 products in the package's files that are a
// direct operand of + or -: the expressions the Go compiler may turn into
// a fused multiply-add on some processors.
func fusable(t *testing.T) []string {
	t.Helper()
	fset := token.NewFileSet()
	names, _ := filepath.Glob("*.go")
	var files []*ast.File
	for _, n := range names {
		if strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, n, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}
	conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil)}
	if _, err := conf.Check("fmath", fset, files, info); err != nil {
		t.Fatal(err)
	}
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
		tv := info.Types[b]
		basic, ok := tv.Type.Underlying().(*types.Basic)
		return ok && basic.Info()&types.IsFloat != 0 && tv.Value == nil // constants are exact
	}
	var found []string
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.BinaryExpr:
				if n.Op == token.ADD || n.Op == token.SUB {
					for _, op := range []ast.Expr{n.X, n.Y} {
						if isFloatMul(op) {
							found = append(found, fset.Position(op.Pos()).String())
						}
					}
				}
			case *ast.AssignStmt:
				// x += y*z and x -= y*z fuse too.
				if n.Tok == token.ADD_ASSIGN || n.Tok == token.SUB_ASSIGN {
					if isFloatMul(n.Rhs[0]) {
						found = append(found, fset.Position(n.Rhs[0].Pos()).String())
					}
				}
			}
			return true
		})
	}
	return found
}

// TestNoFusableProducts keeps the results identical on every processor:
// each float product that feeds a sum must be wrapped in float64(...).
func TestNoFusableProducts(t *testing.T) {
	if _, err := os.Stat("sin.go"); err != nil {
		t.Skip("run from the package directory")
	}
	for _, p := range fusable(t) {
		t.Errorf("%s: a product feeds a sum; wrap it in float64(...) so it cannot be fused", p)
	}
}
