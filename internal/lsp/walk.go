package lsp

import (
	"github.com/AurielSolaris/runescript/internal/ast"
)

// node is anything in a file's syntax tree the server looks at: an
// ast.Node, or one of the parts without a NodeSpan method (*ast.TypeRef,
// *ast.Param, *ast.FieldDecl, *ast.FieldInit, *ast.UseDecl, *ast.CatchClause).
type node = any

// spanOf returns n's span, and false for a nil node.
func spanOf(n node) (ast.Span, bool) {
	switch n := n.(type) {
	case *ast.Ident:
		if n == nil {
			return ast.Span{}, false
		}
		return n.Span, true
	case *ast.TypeRef:
		if n == nil {
			return ast.Span{}, false
		}
		return n.Span, true
	case *ast.UseDecl:
		return n.Span, true
	case *ast.CatchClause:
		return n.Span, true
	case *ast.Param:
		start, ok := spanOf(n.Type)
		if !ok {
			return ast.Span{}, false
		}
		if end, ok := spanOf(n.Name); ok {
			start.End = end.End
		}
		return start, true
	case *ast.FieldDecl:
		start, _ := spanOf(n.Type)
		end, _ := spanOf(n.Name)
		return ast.Span{Start: start.Start, End: end.End}, true
	case *ast.FieldInit:
		start, _ := spanOf(n.Name)
		if end, ok := spanOf(n.Value); ok {
			start.End = end.End
		}
		return start, true
	case ast.Node:
		if n == nil || isNilNode(n) {
			return ast.Span{}, false
		}
		return n.NodeSpan(), true
	}
	return ast.Span{}, false
}

// isNilNode reports a typed nil in an interface, such as a nil *ast.Block
// held as an ast.Stmt.
func isNilNode(n ast.Node) bool {
	switch n := n.(type) {
	case *ast.Block:
		return n == nil
	case *ast.FuncDecl:
		return n == nil
	case *ast.CallExpr:
		return n == nil
	case *ast.MElement:
		return n == nil
	}
	return false
}

// children returns n's direct children, in source order.
func children(n node) []node {
	var out []node
	add := func(xs ...node) {
		for _, x := range xs {
			if _, ok := spanOf(x); ok {
				out = append(out, x)
			}
		}
	}
	switch n := n.(type) {
	case *ast.File:
		for _, u := range n.Uses {
			add(u)
		}
		for _, d := range n.Structs {
			add(d)
		}
		for _, d := range n.Classes {
			add(d)
		}
		for _, d := range n.Externs {
			add(d)
		}
		for _, d := range n.Funcs {
			add(d)
		}
	case *ast.UseDecl:
		add(n.Alias)
	case *ast.StructDecl:
		add(n.Name)
		for _, f := range n.Fields {
			add(f)
		}
	case *ast.FieldDecl:
		add(n.Type, n.Name)
	case *ast.ClassDecl:
		add(n.Name)
		for _, f := range n.Fields {
			add(f)
		}
		for _, c := range n.Ctors {
			add(c)
		}
		for _, m := range n.Methods {
			add(m)
		}
	case *ast.MemberField:
		add(n.Type, n.Name, n.Init)
	case *ast.ExternDecl:
		add(n.Result, n.Name)
		for _, p := range n.Params {
			add(p)
		}
	case *ast.FuncDecl:
		add(n.Result, n.Name)
		for _, p := range n.Params {
			add(p)
		}
		if n.Body != nil {
			add(n.Body)
		}
	case *ast.Param:
		add(n.Type, n.Name)
	case *ast.TypeRef:
		add(n.Pkg)
		if n.Func != nil {
			add(n.Func.Result)
			for _, p := range n.Func.Params {
				add(p)
			}
		}
		for _, a := range n.Args {
			add(a)
		}
	case *ast.Block:
		for _, s := range n.Stmts {
			add(s)
		}
	case *ast.VarDecl:
		add(n.Type, n.Name, n.Init)
	case *ast.AssignStmt:
		add(n.Target, n.Value)
	case *ast.SetStmt:
		add(n.Target, n.Value)
	case *ast.IncDecStmt:
		add(n.Target)
	case *ast.ExprStmt:
		add(n.Call)
	case *ast.ReturnStmt:
		add(n.Value)
	case *ast.IfStmt:
		add(n.Cond, n.Then)
		if n.Else != nil {
			add(n.Else)
		}
	case *ast.WhileStmt:
		add(n.Cond, n.Body)
	case *ast.ForStmt:
		if n.Init != nil {
			add(n.Init)
		}
		add(n.Cond)
		if n.Post != nil {
			add(n.Post)
		}
		add(n.Body)
	case *ast.ForEachStmt:
		add(n.Type, n.Name, n.X, n.Body)
	case *ast.MForEach:
		add(n.Type, n.Name, n.X)
		for _, c := range n.Body {
			add(c)
		}
	case *ast.ThrowStmt:
		add(n.Value)
	case *ast.TryStmt:
		add(n.Body)
		for _, c := range n.Catches {
			add(c)
		}
		if n.Finally != nil {
			add(n.Finally)
		}
	case *ast.CatchClause:
		add(n.Type, n.Name, n.Body)
	case *ast.TemplateStmt:
		add(n.Root)
	case *ast.CallExpr:
		add(n.Pkg, n.Fn, n.Callee)
		for _, a := range n.Args {
			add(a)
		}
	case *ast.UnaryExpr:
		add(n.X)
	case *ast.BinaryExpr:
		add(n.X, n.Y)
	case *ast.LambdaExpr:
		for _, p := range n.Params {
			add(p)
		}
		add(n.Expr)
		if n.Body != nil {
			add(n.Body)
		}
	case *ast.ParenExpr:
		add(n.X)
	case *ast.CastExpr:
		add(n.Type, n.X)
	case *ast.NewExpr:
		add(n.Type)
		for _, a := range n.Args {
			add(a)
		}
	case *ast.IndexExpr:
		add(n.X, n.Index)
	case *ast.ArrayLit:
		for _, e := range n.Elems {
			add(e)
		}
	case *ast.StructLit:
		add(n.Pkg, n.Name)
		for _, f := range n.Fields {
			add(f)
		}
	case *ast.FieldInit:
		add(n.Name, n.Value)
	case *ast.SelectorExpr:
		add(n.X, n.Name)
	case *ast.InterpString:
		for _, e := range n.Exprs {
			add(e)
		}
	case *ast.MElement:
		for _, a := range n.Attrs {
			if a.Expr != nil {
				add(a.Expr)
			}
		}
		for _, c := range n.Children {
			add(c)
		}
	case *ast.MInterp:
		add(n.X)
	case *ast.MIf:
		for _, c := range n.Conds {
			add(c)
		}
		for _, b := range n.Branches {
			for _, c := range b {
				add(c)
			}
		}
		for _, c := range n.Else {
			add(c)
		}
	}
	return out
}

// pathAt returns the nodes of f containing the byte offset off, outermost
// first. A node contains the offsets from its start to its end inclusive,
// so a cursor just after a name is on the name.
func pathAt(f *ast.File, off int) []node {
	var path []node
	var n node = f
	for {
		var next node
		for _, c := range children(n) {
			sp, _ := spanOf(c)
			if sp.Start.Offset <= off && off <= sp.End.Offset {
				next = c
				// A later, tighter match wins: in `a.b`, the cursor at the
				// '.' is at the end of a and the start of b.
			}
		}
		if next == nil {
			return path
		}
		path = append(path, next)
		n = next
	}
}
