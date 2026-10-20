package checker

import (
	"fmt"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/types"
)

func (c *checker) stmts(list []ast.Stmt) {
	c.unreachable(list)
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
		if s.Init != nil {
			c.assign(t, s.Init, "assign")
		} else if t.Kind != types.Invalid && !types.HasZero(t) {
			c.errorf(diagnostics.TypeNoZeroValue, s.Name.Span, "variable '%s' needs an initial value: type '%s' has no zero value", s.Name.Name, t).
				WithHelp(zeroHelp(t))
		}
		c.declareVar(s.Name, t, false)
	case *ast.AssignStmt:
		c.assignStmt(s)
	case *ast.SetStmt:
		c.setStmt(s)
	case *ast.IncDecStmt:
		if v := c.target(s.Target); v != nil && v.typ.Kind != types.Invalid && !types.IsInteger(v.typ) {
			c.errorf(diagnostics.TypeBadOperand, s.Span, "operator '%s' requires an integer variable, but '%s' has type '%s'", s.Op, s.Target.Name, v.typ)
		}
	case *ast.ExprStmt:
		c.discarded(s.Call, c.call(s.Call))
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
	case *ast.ForEachStmt:
		c.forEach(s)
	case *ast.BranchStmt:
		if c.loopDepth == 0 {
			c.errorf(diagnostics.FlowBranchOutsideLoop, s.Span, "'%s' is only allowed inside a loop", s.Keyword).
				WithHelp("break and continue do not cross function boundaries")
		}
	case *ast.ThrowStmt:
		if t := c.expr(s.Value).typ; t.Kind != types.Error && t.Kind != types.Invalid {
			c.errorf(diagnostics.TypeNotError, s.Value.NodeSpan(), "only errors can be thrown, found '%s'", t).
				WithHelp(`throw an error value, e.g. throw new Error("message");`)
		}
	case *ast.TryStmt:
		c.tryStmt(s)
	case *ast.TemplateStmt:
		c.templateStmt(s)
	}
}

// tryStmt checks try/catch/finally. Catch clauses are tried in order, so a
// clause whose type an earlier clause already catches can never run.
func (c *checker) tryStmt(s *ast.TryStmt) {
	c.block(s.Body)
	var earlier []*ast.CatchClause
	for _, cl := range s.Catches {
		t := types.ErrorT // catch (e) catches every error
		if cl.Type != nil {
			t = c.resolveType(cl.Type)
		}
		switch {
		case t.Kind == types.Invalid:
		case t.Kind != types.Error:
			c.errorf(diagnostics.TypeNotError, cl.Type.Span, "a catch clause catches an error type, found '%s'", t).
				WithHelp("the error types are Error, RangeError, TypeError, and ArgumentError")
			t = types.InvalidT
		default:
			for _, prev := range earlier {
				if p := c.info.CatchTypes[prev]; types.IsSubtype(t, p) {
					c.errorf(diagnostics.TypeUnreachableCatch, catchSpan(cl), "this catch clause can never run: the clause at %s already catches '%s'", catchSpan(prev).Start, p).
						WithHelp("put catch clauses for specific error types before the ones for more general types")
					break
				}
			}
		}
		c.info.CatchTypes[cl] = t
		earlier = append(earlier, cl)
		c.push(false)
		c.conflict(cl.Name)
		c.top().add(cl.Name.Name, &symbol{typ: t, isCatch: true})
		c.block(cl.Body)
		c.pop()
	}
	if s.Finally != nil {
		// As in ActionScript 3 and Java, a finally block may return, break,
		// or continue; that replaces however the try statement ended.
		c.block(s.Finally)
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
	case v.fn != nil || v.ext != nil:
		c.errorf(diagnostics.NameAssignToParam, id.Span, "cannot assign to function '%s'", id.Name).
			WithHelp("declare a variable of a function type to hold function values")
	case crossed:
		c.boundaryError(id)
	case v.isParam:
		c.errorf(diagnostics.NameAssignToParam, id.Span, "cannot assign to parameter '%s'", id.Name).
			WithHelp("parameters are read-only; copy the value into a local variable")
	case v.isCatch:
		c.errorf(diagnostics.NameAssignToParam, id.Span, "cannot assign to caught error '%s'", id.Name).
			WithHelp("the error variable of a catch clause is read-only")
	default:
		v.written = true
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
	op := strings.TrimSuffix(s.Op, "=")
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

// catchSpan is where a catch clause's type is written, or its name for
// catch (e).
func catchSpan(cl *ast.CatchClause) diagnostics.Span {
	if cl.Type != nil {
		return cl.Type.Span
	}
	return cl.Name.Span
}

// zeroHelp suggests an initial value for a variable of type t, which has
// no zero value.
func zeroHelp(t *types.Type) string {
	switch t.Kind {
	case types.Func:
		return "initialize it with a function or a lambda, e.g. `= (int n) => n;`"
	case types.Error:
		return fmt.Sprintf("initialize it, e.g. `= new %s(\"\");`", t.Name)
	case types.Signal:
		return "initialize it with signal(...)"
	case types.Struct:
		return "one of its fields has a type without a zero value; initialize it with a struct literal"
	case types.Opaque:
		return "initialize it with a value made by one of its package's functions"
	}
	return "initialize it when declaring it"
}
