package checker

import (
	"fmt"

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

// arrayField checks a.length.
func (c *checker) arrayField(e *ast.SelectorExpr, t *types.Type) *types.Type {
	if e.Name.Name == "length" {
		return types.Int
	}
	c.errorf(diagnostics.TypeNoField, e.Name.Span, "'%s' has no field '%s'", t, e.Name.Name).
		WithHelp(arrayHelp)
	return types.InvalidT
}

// arrayHelp lists what arrays have, for diagnostics.
const arrayHelp = "arrays have the field length and the methods push, pop, insert, removeAt, slice, indexOf, contains, reverse, and sort"

// arrayMethod checks a method call on an array of type t:
//
//	push(v)          append v                       void
//	pop()            remove and return the last     T
//	insert(i, v)     insert v before index i        void (0 <= i <= length)
//	removeAt(i)      remove and return element i    T
//	slice(start, end) copy of [start, end)           Array<T>
//	indexOf(v)       first index of v, or -1        int (T comparable)
//	contains(v)      whether v is an element        bool (T comparable)
//	reverse()        reverse in place               void
//	sort()           sort in place, ascending       void (T a number, char, or string)
//	sort(less)       sort in place by less          void (less is bool(T, T))
func (c *checker) arrayMethod(e *ast.CallExpr, name *ast.Ident, t *types.Type) *types.Type {
	c.info.ArrayCalls[e] = name.Name
	argc := func(n int, what string) bool {
		if len(e.Args) == n {
			return true
		}
		c.errorf(diagnostics.TypeArgCount, e.Span, "%s expects %s, found %d", name.Name, what, len(e.Args))
		c.exprs(e.Args)
		return false
	}
	comparable := func() bool {
		if types.Comparable(t.Elem) {
			return true
		}
		c.errorf(diagnostics.TypeBadOperand, name.Span, "%s needs elements that can be compared with ==, but '%s' cannot", name.Name, t.Elem).
			WithHelp("search with a loop instead")
		c.exprs(e.Args)
		return false
	}
	switch name.Name {
	case "push":
		if argc(1, "1 argument, the element to append") {
			c.assign(t.Elem, e.Args[0], "pass")
		}
		return types.VoidT
	case "pop", "reverse":
		argc(0, "no arguments")
		if name.Name == "pop" {
			return t.Elem
		}
		return types.VoidT
	case "insert":
		if argc(2, "2 arguments, an index and the element") {
			c.indexValue(e.Args[0])
			c.assign(t.Elem, e.Args[1], "pass")
		}
		return types.VoidT
	case "removeAt":
		if argc(1, "1 argument, the index of the element to remove") {
			c.indexValue(e.Args[0])
		}
		return t.Elem
	case "slice":
		if argc(2, "2 arguments, the start and end indexes") {
			c.indexValue(e.Args[0])
			c.indexValue(e.Args[1])
		}
		return t
	case "indexOf", "contains":
		if argc(1, "1 argument, the element to look for") && comparable() {
			c.assign(t.Elem, e.Args[0], "pass")
		}
		if name.Name == "indexOf" {
			return types.Int
		}
		return types.BoolT
	case "sort":
		switch len(e.Args) {
		case 0:
			if !types.IsNumeric(t.Elem) && t.Elem.Kind != types.Char && t.Elem.Kind != types.String {
				c.errorf(diagnostics.TypeBadOperand, name.Span, "sort() needs numbers, chars, or strings, but the elements have type '%s'", t.Elem).
					WithHelp(fmt.Sprintf("pass a function that orders them: a.sort((%s x, %s y) => ...)", t.Elem, t.Elem))
			}
		case 1:
			c.assign(types.NewFunc([]*types.Type{t.Elem, t.Elem}, types.BoolT), e.Args[0], "pass")
		default:
			argc(1, "no arguments, or 1, a function telling whether one element goes before another")
		}
		return types.VoidT
	}
	delete(c.info.ArrayCalls, e)
	c.errorf(diagnostics.TypeNoField, name.Span, "'%s' has no method '%s'", t, name.Name).WithHelp(arrayHelp)
	c.exprs(e.Args)
	return types.InvalidT
}

// forEach checks `for each (T x in a) { }`. a must be an array whose
// elements can be assigned to T; x is a new variable of the loop body.
func (c *checker) forEach(s *ast.ForEachStmt) {
	t := c.eachHeader("for each", s.Type, s.Name, s.X)
	c.push(false)
	c.declareVar(s.Name, t, false)
	c.loopBody(s.Body)
	c.pop()
}

// eachHeader checks the header `T name in x` of a for-each loop or a
// template $foreach and returns the loop variable's type.
func (c *checker) eachHeader(what string, typ *ast.TypeRef, name *ast.Ident, x ast.Expr) *types.Type {
	at := c.defaulted(x).typ
	t := c.resolveType(typ)
	if t.Kind == types.Void {
		c.errorf(diagnostics.TypeVoidValue, typ.Span, "variable '%s' cannot have type 'void'", name.Name)
		t = types.InvalidT
	}
	switch {
	case at.Kind == types.Invalid || t.Kind == types.Invalid:
	case at.Kind != types.Array:
		d := c.errorf(diagnostics.TypeMismatch, x.NodeSpan(), "%s needs an array, found '%s'", what, describe(at))
		if at.Kind == types.String {
			d.WithHelp("split the string first, e.g. string.split(s, \"\")")
		}
	case !types.IsSubtype(at.Elem, t):
		c.errorf(diagnostics.TypeMismatch, typ.Span, "the elements of '%s' have type '%s', not '%s'", at, at.Elem, t).
			WithHelp(fmt.Sprintf("declare the loop variable as %s %s", at.Elem, name.Name))
	}
	return t
}
