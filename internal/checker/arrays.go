package checker

import (
	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/types"
)

// Arrays (OQ-TYPE-02a): Array<T> is a growable list held by reference.
// Elements are read with a[i] and written with a[i] = v; an index out of
// range throws a RangeError at run time. a.length is the number of
// elements; a.push(v) appends and a.pop() removes and returns the last.

// arrayType resolves Array<T>.
func (c *checker) arrayType(ref *ast.TypeRef) *types.Type {
	if len(ref.Args) != 1 {
		c.errorf(diagnostics.TypeArgs, ref.Span, "Array takes one type argument, the element type, e.g. Array<string>")
		return types.InvalidT
	}
	elem := c.resolveType(ref.Args[0])
	switch {
	case elem.Kind == types.Invalid:
		return types.InvalidT
	case !types.IsElemType(elem):
		c.errorf(diagnostics.TypeArgs, ref.Args[0].Span, "an array cannot hold values of type '%s'", elem)
		return types.InvalidT
	}
	return types.NewArray(elem)
}

// arrayLit checks [a, b, c]. With an expected array type, each element must
// be assignable to its element type; without one, the element type is the
// first element's (an empty literal then needs a declared type).
func (c *checker) arrayLit(e *ast.ArrayLit, expected *types.Type) *types.Type {
	t := expected
	if t == nil {
		if len(e.Elems) == 0 {
			c.errorf(diagnostics.TypeMismatch, e.Span, "the element type of an empty array literal is unknown").
				WithHelp("declare it, e.g. Array<string> names = [];")
			return types.InvalidT
		}
		first := c.defaulted(e.Elems[0]).typ
		switch {
		case first.Kind == types.Invalid:
			c.exprs(e.Elems[1:])
			return types.InvalidT
		case !types.IsElemType(first):
			c.errorf(diagnostics.TypeMismatch, e.Elems[0].NodeSpan(), "an array cannot hold values of type '%s'", first)
			c.exprs(e.Elems[1:])
			return types.InvalidT
		}
		t = types.NewArray(first)
		for _, x := range e.Elems[1:] {
			c.assign(first, x, "assign")
		}
	} else {
		for _, x := range e.Elems {
			c.assign(t.Elem, x, "assign")
		}
	}
	c.info.Types[e] = t
	return t
}

// index checks a[i] and returns the element type.
func (c *checker) index(e *ast.IndexExpr) *types.Type {
	t := c.defaulted(e.X).typ
	c.indexValue(e.Index)
	switch t.Kind {
	case types.Invalid:
		return types.InvalidT
	case types.Array:
		return t.Elem
	case types.String:
		c.errorf(diagnostics.TypeBadOperand, e.Span, "strings cannot be indexed").
			WithHelp("use string.slice(s, start, end), with byte offsets")
		return types.InvalidT
	}
	c.errorf(diagnostics.TypeBadOperand, e.Span, "values of type '%s' cannot be indexed", t)
	return types.InvalidT
}

// indexValue checks an index: an int or uint (a constant becomes an int).
func (c *checker) indexValue(i ast.Expr) {
	op := c.expr(i)
	switch {
	case op.typ.Kind == types.UntypedInt:
		c.convertConst(i, op.konst, types.Int)
	case op.typ.Kind == types.Signed || op.typ.Kind == types.Unsigned || op.typ.Kind == types.Invalid:
	default:
		c.errorf(diagnostics.TypeMismatch, i.NodeSpan(), "an index must be an integer, found '%s'", describe(op.typ))
	}
}

// indexAssign checks a[i] op= v.
func (c *checker) indexAssign(s *ast.IndexAssignStmt) {
	elem := c.index(s.Target)
	if elem.Kind == types.Invalid {
		c.expr(s.Value)
		return
	}
	c.info.Types[s.Target] = elem
	if s.Op == "=" {
		c.assign(elem, s.Value, "assign")
		return
	}
	op := s.Op[:1]
	y := c.expr(s.Value)
	t := c.binaryType(op, operand{typ: elem}, y, s.Target, s.Value, s.Span)
	if t.typ.Kind != types.Invalid && !types.Identical(t.typ, elem) {
		c.errorf(diagnostics.TypeMismatch, s.Span, "operator '%s' produces '%s', which cannot be stored in '%s'", s.Op, t.typ, elem)
	}
}

// arrayField checks a.length.
func (c *checker) arrayField(e *ast.SelectorExpr, t *types.Type) *types.Type {
	if e.Name.Name == "length" {
		return types.Int
	}
	c.errorf(diagnostics.TypeNoField, e.Name.Span, "'%s' has no field '%s'", t, e.Name.Name).
		WithHelp("arrays have the field length and the methods push and pop")
	return types.InvalidT
}

// arrayMethod checks a.push(v) and a.pop() on an array of type t.
func (c *checker) arrayMethod(e *ast.CallExpr, name *ast.Ident, t *types.Type) *types.Type {
	c.info.ArrayCalls[e] = name.Name
	switch name.Name {
	case "push":
		if len(e.Args) != 1 {
			c.errorf(diagnostics.TypeArgCount, e.Span, "push expects 1 argument, the element to append, found %d", len(e.Args))
			c.exprs(e.Args)
			return types.VoidT
		}
		c.assign(t.Elem, e.Args[0], "pass")
		return types.VoidT
	case "pop":
		if len(e.Args) != 0 {
			c.errorf(diagnostics.TypeArgCount, e.Span, "pop takes no arguments, found %d", len(e.Args))
			c.exprs(e.Args)
		}
		return t.Elem
	}
	delete(c.info.ArrayCalls, e)
	c.errorf(diagnostics.TypeNoField, name.Span, "'%s' has no method '%s'", t, name.Name).
		WithHelp("arrays have the field length and the methods push and pop")
	c.exprs(e.Args)
	return types.InvalidT
}
