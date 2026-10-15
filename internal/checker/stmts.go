package checker

import (
	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/types"
)

func (c *checker) stmts(list []ast.Stmt) {
	for _, s := range list {
		c.stmt(s)
	}
}

func (c *checker) block(b *ast.Block) {
	c.push(false)
	c.stmts(b.Stmts)
	c.pop()
}

func (c *checker) stmt(s ast.Stmt) {
	switch s := s.(type) {
	case *ast.FuncDecl:
		// Nested function: visible from here to the end of the enclosing
		// block (and inside its own body, for recursion).
		t := c.declareFunc(s)
		if s.Body != nil {
			c.funcBody(s, t)
		}
	case *ast.VarDecl:
		t := c.resolveType(s.Type)
		if t.Kind == types.Void {
			c.errorf(diagnostics.TypeVoidValue, s.Type.Span, "variable '%s' cannot have type 'void'", s.Name.Name)
			t = types.InvalidT
		}
		c.assign(t, s.Init, "assign")
		c.declareVar(s.Name, t, false)
	case *ast.AssignStmt:
		c.assignStmt(s)
	case *ast.IncDecStmt:
		if v := c.target(s.Target); v != nil && v.typ.Kind != types.Invalid && !types.IsInteger(v.typ) {
			c.errorf(diagnostics.TypeBadOperand, s.Span, "operator '%s' requires an integer variable, but '%s' has type '%s'", s.Op, s.Target.Name, v.typ)
		}
	case *ast.ExprStmt:
		c.call(s.Call)
	case *ast.ReturnStmt:
		c.returnStmt(s)
	case *ast.IfStmt:
		c.cond(s.Cond)
		c.block(s.Then)
		if s.Else != nil {
			if b, ok := s.Else.(*ast.Block); ok {
				c.block(b)
			} else {
				c.stmt(s.Else)
			}
		}
	case *ast.WhileStmt:
		c.cond(s.Cond)
		c.loopBody(s.Body)
	case *ast.ForStmt:
		c.push(false)
		if s.Init != nil {
			c.stmt(s.Init)
		}
		if s.Cond != nil {
			c.cond(s.Cond)
		}
		if s.Post != nil {
			c.stmt(s.Post)
		}
		c.loopBody(s.Body)
		c.pop()
	case *ast.BranchStmt:
		if c.loopDepth == 0 {
			c.errorf(diagnostics.FlowBranchOutsideLoop, s.Span, "'%s' is only allowed inside a loop", s.Keyword).
				WithHelp("break and continue do not cross function boundaries")
		}
	}
}

func (c *checker) loopBody(b *ast.Block) {
	c.loopDepth++
	c.block(b)
	c.loopDepth--
}

// target resolves an assignment target. Parameters and functions are
// read-only (OQ-VAR-02); enclosing functions' variables are not visible.
func (c *checker) target(id *ast.Ident) *symbol {
	v, crossed := c.lookup(id.Name)
	switch {
	case v == nil:
		c.errorf(diagnostics.NameUndefined, id.Span, "undefined variable '%s'", id.Name)
	case v.fn != nil:
		c.errorf(diagnostics.NameAssignToParam, id.Span, "cannot assign to function '%s'", id.Name).
			WithHelp("declare a variable of a function type to hold function values")
	case crossed:
		c.boundaryError(id)
	case v.isParam:
		c.errorf(diagnostics.NameAssignToParam, id.Span, "cannot assign to parameter '%s'", id.Name).
			WithHelp("parameters are read-only; copy the value into a local variable")
	default:
		return v
	}
	return nil
}

func (c *checker) boundaryError(id *ast.Ident) {
	c.errorf(diagnostics.NameAcrossBoundary, id.Span, "'%s' belongs to an enclosing function and is not visible here", id.Name).
		WithHelp("nested functions and lambdas do not capture outer variables; pass '" + id.Name + "' as a parameter")
}

func (c *checker) assignStmt(s *ast.AssignStmt) {
	v := c.target(s.Target)
	if v == nil {
		c.expr(s.Value)
		return
	}
	if s.Op == "=" {
		c.assign(v.typ, s.Value, "assign")
		return
	}
	// x op= y is checked as x = x op y.
	op := s.Op[:1]
	y := c.expr(s.Value)
	t := c.binaryType(op, operand{typ: v.typ}, y, s.Target, s.Value, s.Span)
	if t.typ.Kind != types.Invalid && !types.Identical(t.typ, v.typ) {
		c.errorf(diagnostics.TypeMismatch, s.Span, "operator '%s' produces '%s', which cannot be stored in '%s'", s.Op, t.typ, v.typ)
	}
}

func (c *checker) returnStmt(s *ast.ReturnStmt) {
	if c.result == nil { // first return of a lambda whose result is inferred
		c.result = types.VoidT
		if s.Value != nil {
			op := c.defaulted(s.Value)
			if op.typ.Kind == types.Void {
				c.errorf(diagnostics.TypeVoidValue, s.Value.NodeSpan(), "expression has type 'void' and cannot be returned")
			}
			c.result = op.typ
		}
		return
	}
	switch {
	case c.result.Kind == types.Void && s.Value != nil:
		c.expr(s.Value)
		c.errorf(diagnostics.TypeBadReturn, s.Value.NodeSpan(), "%s returns 'void' and cannot return a value", c.fnName)
	case c.result.Kind != types.Void && s.Value == nil:
		if c.result.Kind != types.Invalid {
			c.errorf(diagnostics.TypeBadReturn, s.Span, "%s must return a value of type '%s'", c.fnName, c.result)
		}
	case s.Value != nil:
		c.assign(c.result, s.Value, "return")
	}
}

func (c *checker) cond(e ast.Expr) {
	op := c.defaulted(e)
	if op.typ.Kind != types.Bool && op.typ.Kind != types.Invalid {
		c.errorf(diagnostics.TypeCondition, e.NodeSpan(), "condition must be 'bool', found '%s'", op.typ).
			WithHelp("integers are not implicitly converted to bool; compare explicitly, e.g. `x != 0`")
	}
}
