package checker

import (
	"fmt"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/types"
)

// Warnings are collected in Info.Warnings, apart from errors: they never
// stop a program from checking. The driver decides whether to show them
// and whether to treat them as errors.

func (c *checker) warnf(code diagnostics.Code, span diagnostics.Span, format string, args ...any) *diagnostics.Diagnostic {
	d := diagnostics.NewWarning(code, span, format, args...)
	c.info.Warnings = append(c.info.Warnings, d)
	return d
}

// reportUnused warns about the variables and parameters of s that are
// never read. Assigning to a variable is not reading it. The error
// variable of a catch clause is never reported: the syntax requires it.
func (c *checker) reportUnused(s *scope) {
	for _, name := range s.order {
		sym := s.syms[name]
		if sym.fn != nil || sym.isCatch || sym.read || sym.decl == nil || sym.typ.Kind == types.Invalid {
			continue
		}
		switch {
		case sym.isParam:
			c.warnf(diagnostics.WarnUnusedParam, sym.decl.Span, "parameter '%s' is never used", name).
				WithHelp("remove the parameter, or allow this warning with --allow " + diagnostics.WarnUnusedParam.String())
		case sym.written:
			c.warnf(diagnostics.WarnUnusedVar, sym.decl.Span, "variable '%s' is assigned but never read", name).
				WithHelp("remove the variable and the assignments to it")
		default:
			c.warnf(diagnostics.WarnUnusedVar, sym.decl.Span, "variable '%s' is never used", name).
				WithHelp("remove the declaration")
		}
	}
}

// reportUnusedUses warns about the current file's use declarations whose
// qualifier is never referenced.
func (c *checker) reportUnusedUses() {
	for q, b := range c.quals {
		if b.used || (b.std == nil && b.pkg == nil && b.rune == "") || b.rune == "dom" && c.domUsed[c.file] {
			continue
		}
		c.warnf(diagnostics.WarnUnusedUse, b.decl.Span, "%q is used but '%s' is never referenced", b.decl.Path, q).
			WithHelp("remove the use declaration")
	}
}

// unreachable warns once about the statements of a block that follow one
// that never finishes normally.
func (c *checker) unreachable(list []ast.Stmt) {
	for i, s := range list[:max(len(list)-1, 0)] {
		if jumps(s) {
			span := diagnostics.Span{Start: list[i+1].NodeSpan().Start, End: list[len(list)-1].NodeSpan().End}
			c.warnf(diagnostics.WarnUnreachable, span, "unreachable code: the %s at %s never lets execution reach this", jumpWord(s), s.NodeSpan().Start).
				WithHelp("remove the unreachable statements")
			return
		}
	}
}

// jumps reports whether s always leaves the enclosing statement list: by
// return, throw, break, or continue, on every path. Loops never count
// (OQ-FLOW-01).
func jumps(s ast.Stmt) bool {
	switch s := s.(type) {
	case *ast.ReturnStmt, *ast.ThrowStmt, *ast.BranchStmt:
		return true
	case *ast.Block:
		return listJumps(s.Stmts)
	case *ast.IfStmt:
		if !listJumps(s.Then.Stmts) || s.Else == nil {
			return false
		}
		return jumps(s.Else)
	case *ast.TryStmt:
		if s.Finally != nil && listJumps(s.Finally.Stmts) {
			return true
		}
		if !listJumps(s.Body.Stmts) {
			return false
		}
		for _, cl := range s.Catches {
			if !listJumps(cl.Body.Stmts) {
				return false
			}
		}
		return true
	}
	return false
}

func listJumps(list []ast.Stmt) bool {
	for _, s := range list {
		if jumps(s) {
			return true
		}
	}
	return false
}

func jumpWord(s ast.Stmt) string {
	switch s := s.(type) {
	case *ast.ReturnStmt:
		return "return"
	case *ast.ThrowStmt:
		return "throw"
	case *ast.BranchStmt:
		return s.Keyword
	case *ast.TryStmt:
		return "try statement"
	}
	return "if statement"
}

// discarded warns about a call statement whose result is not void.
func (c *checker) discarded(call *ast.CallExpr, t *types.Type) {
	if t.Kind == types.Void || t.Kind == types.Invalid {
		return
	}
	what := "this call"
	switch {
	case call.Pkg != nil:
		what = fmt.Sprintf("'%s.%s'", call.Pkg.Name, call.Fn.Name)
	case call.Fn != nil:
		what = "'" + call.Fn.Name + "'"
	}
	c.warnf(diagnostics.WarnDiscardResult, call.Span, "the result of %s, of type '%s', is discarded", what, t).
		WithHelp("store it in a variable, or allow this warning with --allow " + diagnostics.WarnDiscardResult.String())
}
