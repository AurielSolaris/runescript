package ir

import (
	"fmt"
	"math"

	"github.com/AurielSolaris/runescript/internal/types"
)

// Verify re-checks the typing rules on p, independently of the checker:
// every stored, passed, or returned value has exactly the destination's
// type, integer operations never mix signedness, and conditions are bool.
// A non-nil error is an internal compiler error. Back ends call Verify
// before generating code, so a checker bug cannot turn into unguarded
// output.
func Verify(p *Program) (err error) {
	v := &verifier{seen: map[*Func]bool{}}
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(verifyError); ok {
				err = e
				return
			}
			panic(r)
		}
	}()
	if p.Entry == nil {
		v.fail("program has no entry point")
	}
	for _, f := range p.Funcs {
		v.fn(f)
	}
	return nil
}

type verifyError struct{ msg string }

func (e verifyError) Error() string { return "ir: internal compiler error: " + e.msg }

type verifier struct {
	seen   map[*Func]bool
	result *types.Type // result type of the function being verified
	loops  int
	// wideOK is set while verifying an operand that may be a wide
	// intermediate (Arith.Wide).
	wideOK bool
}

func (v *verifier) fail(format string, args ...any) {
	panic(verifyError{fmt.Sprintf(format, args...)})
}

func (v *verifier) fn(f *Func) {
	if v.seen[f] {
		return
	}
	v.seen[f] = true
	if f.Type == nil || f.Type.Kind != types.Func || f.Body == nil {
		v.fail("function %q is incomplete", f.Name)
	}
	if len(f.Params) != len(f.Type.Params) {
		v.fail("function %q has %d parameters but type %s", f.Name, len(f.Params), f.Type)
	}
	for i, p := range f.Params {
		v.storable(p.Type, "parameter "+p.Name)
		if !types.Identical(p.Type, f.Type.Params[i]) {
			v.fail("parameter %q of %q has type %s, signature says %s", p.Name, f.Name, p.Type, f.Type.Params[i])
		}
	}
	savedResult, savedLoops := v.result, v.loops
	v.result, v.loops = f.Type.Result, 0
	v.block(f.Body)
	v.result, v.loops = savedResult, savedLoops
}

// storable rejects types no value can have.
func (v *verifier) storable(t *types.Type, what string) {
	switch t.Kind {
	case types.Signed, types.Unsigned, types.Bool, types.String, types.Func, types.Float, types.Error,
		types.Element, types.Signal, types.Char, types.Array, types.Struct, types.Opaque, types.Class:
		return
	}
	v.fail("%s has type %s, which cannot hold a value", what, t)
}

// same requires e to have type want, or, for errors, a subtype of it.
// Error values of every type share one representation, so a subtype
// needs no conversion; every other value has exactly its type.
func (v *verifier) same(want *types.Type, e Expr, what string) {
	got := v.expr(e)
	if !types.IsSubtype(got, want) {
		v.fail("%s: expected %s, found %s", what, want, got)
	}
}

func (v *verifier) block(b *Block) {
	for _, s := range b.Stmts {
		v.stmt(s)
	}
}

func (v *verifier) stmt(s Stmt) {
	switch s := s.(type) {
	case *FuncStmt:
		v.fn(s.Func)
	case *Decl:
		v.storable(s.Var.Type, "variable "+s.Var.Name)
		v.same(s.Var.Type, s.Init, "initializer of "+s.Var.Name)
	case *Assign:
		if s.Var.Param {
			v.fail("assignment to parameter %q", s.Var.Name)
		}
		v.same(s.Var.Type, s.Value, "assignment to "+s.Var.Name)
	case *Eval:
		v.expr(s.X)
	case *IndexSet:
		t := v.array(s.X, "element assignment")
		v.index(s.I)
		v.same(t.Elem, s.Value, "element assignment")
	case *ObjectSet:
		t := v.object(s.X, s.Index)
		v.same(t.Fields[s.Index].Type, s.Value, "field "+t.Fields[s.Index].Name)
	case *Return:
		if s.Value == nil {
			if v.result.Kind != types.Void {
				v.fail("bare return from a function returning %s", v.result)
			}
			return
		}
		if v.result.Kind == types.Void {
			v.fail("void function returns a value")
		}
		v.same(v.result, s.Value, "return value")
	case *If:
		v.same(types.BoolT, s.Cond, "if condition")
		v.block(s.Then)
		if s.Else != nil {
			v.stmt(s.Else)
		}
	case *Block:
		v.block(s)
	case *While:
		v.same(types.BoolT, s.Cond, "while condition")
		v.loop(s.Body)
	case *For:
		if s.Init != nil {
			v.stmt(s.Init)
		}
		if s.Cond != nil {
			v.same(types.BoolT, s.Cond, "for condition")
		}
		if s.Post != nil {
			v.stmt(s.Post)
		}
		v.loop(s.Body)
	case *Break, *Continue:
		if v.loops == 0 {
			v.fail("break or continue outside a loop")
		}
	case *Throw:
		if t := v.expr(s.X); t.Kind != types.Error {
			v.fail("throw of %s", t)
		}
	case *Try:
		v.block(s.Body)
		for _, c := range s.Catches {
			if c.T.Kind != types.Error || !types.Identical(c.Var.Type, c.T) {
				v.fail("catch clause of type %s into %s", c.T, c.Var.Type)
			}
			v.block(c.Body)
		}
		if s.Finally != nil {
			v.block(s.Finally)
		}
	default:
		v.fail("unexpected statement %T", s)
	}
}

func (v *verifier) loop(b *Block) {
	v.loops++
	v.block(b)
	v.loops--
}

func (v *verifier) expr(e Expr) *types.Type {
	if a, ok := e.(*Arith); ok && a.Wide && !v.wideOK {
		v.fail("wide intermediate %s used outside integer arithmetic, a comparison, or a conversion", a.Op)
	}
	v.wideOK = false
	switch e := e.(type) {
	case *IntConst:
		if !types.InRange(e.T, e.Value) && !(e.T.Kind == types.Unsigned && types.IntBits(e.T) == 64) {
			v.fail("constant %d does not fit in %s", e.Value, e.T)
		}
	case *FloatConst:
		if e.T.Kind != types.Float || (e.T.Bits == 32 && !math.IsNaN(e.Value) && float64(float32(e.Value)) != e.Value) {
			v.fail("constant %v does not fit in %s", e.Value, e.T)
		}
	case *Neg:
		if t := v.expr(e.X); t.Kind != types.Float {
			v.fail("float negation of %s", t)
		}
	case *Convert:
		v.wideOK = true
		if from := v.expr(e.X); !Convertible(from, e.To) {
			v.fail("conversion from %s to %s", from, e.To)
		}
	case *NewError:
		if e.T.Kind != types.Error {
			v.fail("new %s", e.T)
		}
		if e.Message != nil {
			v.same(types.StringT, e.Message, "error message")
		}
		if e.ID != nil {
			v.same(types.Int, e.ID, "errorID")
		}
	case *FloatTest:
		if t := v.expr(e.X); t.Kind != types.Float || (e.Name != "isNaN" && e.Name != "isFinite") {
			v.fail("%s of %s", e.Name, t)
		}
	case *Field:
		want := map[string]*types.Type{"message": types.StringT, "name": types.StringT, "errorID": types.Int}[e.Name]
		if t := v.expr(e.X); t.Kind != types.Error || want == nil || !types.Identical(want, e.T) {
			v.fail("field %s of %s", e.Name, t)
		}
	case *StructLit:
		if e.T.Kind != types.Struct || len(e.Inits) != len(e.T.Fields) {
			v.fail("struct literal of %s with %d fields", e.T, len(e.Inits))
		}
		given := map[int]bool{}
		for _, in := range e.Inits {
			if in.Index < 0 || in.Index >= len(e.T.Fields) || given[in.Index] {
				v.fail("struct literal of %s sets field %d badly", e.T, in.Index)
			}
			given[in.Index] = true
			v.same(e.T.Fields[in.Index].Type, in.Value, "field "+e.T.Fields[in.Index].Name)
		}
	case *GetField:
		v.field(e.X, e.Index)
	case *ObjectNew:
		if e.T.Kind != types.Class || len(e.Fields) != len(e.T.Fields) {
			v.fail("new object of %s with %d fields", e.T, len(e.Fields))
		}
		for i, x := range e.Fields {
			v.same(e.T.Fields[i].Type, x, "field "+e.T.Fields[i].Name)
		}
	case *ObjectGet:
		v.object(e.X, e.Index)
	case *WithField:
		t := v.field(e.X, e.Index)
		v.same(t.Fields[e.Index].Type, e.Value, "field "+t.Fields[e.Index].Name)
	case *ArrayLit:
		if e.T.Kind != types.Array || !types.IsElemType(e.T.Elem) {
			v.fail("array literal of type %s", e.T)
		}
		for _, x := range e.Elems {
			v.same(e.T.Elem, x, "array element")
		}
	case *Index:
		v.array(e.X, "indexing")
		v.index(e.I)
	case *IndexOld:
	case *ArrayLen:
		v.array(e.X, "length")
	case *ArrayPush:
		t := v.array(e.X, "push")
		v.same(t.Elem, e.V, "pushed element")
	case *ArrayPop:
		v.array(e.X, "pop")
	case *ArrayCall:
		v.arrayCall(e)
	case *CharConst:
		if !types.ValidChar(int64(e.Value)) {
			v.fail("char constant %d is not a Unicode scalar value", e.Value)
		}
	case *StringConst, *BoolConst, *Load:
	case *FuncRef:
		if e.Func.Type == nil {
			v.fail("reference to undefined function %q", e.Func.Name)
		}
	case *Lambda:
		v.fn(e.Func)
	case *Arith:
		switch {
		case e.T.Kind == types.Float && e.Op == "%":
			v.fail("%% on %s", e.T)
		case e.T.Kind != types.Signed && e.T.Kind != types.Unsigned && e.T.Kind != types.Float:
			v.fail("arithmetic on %s", e.T)
		case e.Wide && !generic(e.T):
			v.fail("wide arithmetic on %s", e.T)
		}
		v.wideOK = generic(e.T)
		v.same(e.T, e.X, "left operand of "+e.Op)
		v.wideOK = generic(e.T)
		v.same(e.T, e.Y, "right operand of "+e.Op)
	case *Bits:
		if e.T.Kind != types.Signed && e.T.Kind != types.Unsigned {
			v.fail("bitwise %s on %s", e.Op, e.T)
		}
		v.same(e.T, e.X, "left operand of "+e.Op)
		v.same(e.T, e.Y, "right operand of "+e.Op)
	case *Complement:
		if t := v.expr(e.X); t.Kind != types.Signed && t.Kind != types.Unsigned {
			v.fail("~ on %s", t)
		}
	case *Shift:
		if t := v.expr(e.X); t.Kind != types.Signed && t.Kind != types.Unsigned {
			v.fail("shift of %s", t)
		}
		v.index(e.Count) // an integer
	case *Concat:
		if len(e.Parts) == 0 {
			v.fail("empty string join")
		}
		for _, p := range e.Parts {
			v.same(types.StringT, p, "part of a string join")
		}
	case *Compare:
		v.wideOK = true
		x := v.expr(e.X)
		v.wideOK = true
		y := v.expr(e.Y)
		if !types.Identical(x, y) {
			v.fail("comparison %s of %s and %s", e.Op, x, y)
		}
		ordered := e.Op != "==" && e.Op != "!="
		switch {
		case x.Kind == types.Signed || x.Kind == types.Unsigned || x.Kind == types.Float || x.Kind == types.Char:
		case !ordered && x.Kind == types.Struct && types.Comparable(x):
		case !ordered && x.Kind == types.Class:
		case !ordered && (x.Kind == types.Bool || x.Kind == types.String):
		default:
			v.fail("operator %s is not defined on %s", e.Op, x)
		}
	case *Logic:
		v.same(types.BoolT, e.X, "operand of "+e.Op)
		v.same(types.BoolT, e.Y, "operand of "+e.Op)
	case *Not:
		v.same(types.BoolT, e.X, "operand of !")
	case *CallFunc:
		v.args(e.Func.Type, e.Args, "call of "+e.Func.Name)
	case *CallValue:
		t := v.expr(e.Callee)
		if t.Kind != types.Func {
			v.fail("call of a value of type %s", t)
		}
		v.args(t, e.Args, "call of a function value")
	case *CallNative:
		v.args(types.NewFunc(e.Func.Params, e.Func.Result), e.Args, "call of "+e.Pkg+"."+e.Func.Name)
	case *CallExtern:
		for _, t := range e.Extern.Type.Params {
			v.boundary(t, false, e.Extern.Name)
		}
		v.boundary(e.Extern.Type.Result, true, e.Extern.Name)
		v.args(e.Extern.Type, e.Args, "call of extern "+e.Extern.Name)
	case *NewSignal:
		if e.T.Kind != types.Signal || !types.IsScalar(e.T.Elem) {
			v.fail("rune of type %s", e.T)
		}
		v.same(e.T.Elem, e.Init, "initial value of a rune")
	case *SignalGet:
		if t := v.expr(e.X); t.Kind != types.Signal {
			v.fail("reading a %s as a rune", t)
		}
	case *SignalSet:
		t := v.expr(e.X)
		if t.Kind != types.Signal {
			v.fail("setting a %s as a rune", t)
		}
		v.same(t.Elem, e.Value, "value of a rune")
	case *Template:
		v.telement(e.Root)
	case *RenderApp:
		v.same(types.NewFunc(nil, types.ElementT), e.Component, "component passed to renderApp")
		v.same(types.StringT, e.Selector, "selector passed to renderApp")
	case *Trace:
		for _, a := range e.Args {
			switch t := v.expr(a); t.Kind {
			case types.Signed, types.Unsigned, types.Bool, types.String, types.Float, types.Error, types.Char, types.Class:
			case types.Struct:
				if !types.Traceable(t) {
					v.fail("trace of %s", t)
				}
			default:
				v.fail("trace of %s", t)
			}
		}
	default:
		v.fail("unexpected expression %T", e)
	}
	return e.Type()
}

// object checks that x is an object with a field i, and returns its class.
func (v *verifier) object(x Expr, i int) *types.Type {
	t := v.expr(x)
	if t.Kind != types.Class || i < 0 || i >= len(t.Fields) {
		v.fail("field %d of a %s", i, t)
	}
	return t
}

func (v *verifier) field(x Expr, i int) *types.Type {
	t := v.expr(x)
	if t.Kind != types.Struct || i < 0 || i >= len(t.Fields) {
		v.fail("field %d of a %s", i, t)
	}
	return t
}

func (v *verifier) array(x Expr, what string) *types.Type {
	t := v.expr(x)
	if t.Kind != types.Array {
		v.fail("%s of a %s", what, t)
	}
	return t
}

func (v *verifier) index(i Expr) {
	if t := v.expr(i); t.Kind != types.Signed && t.Kind != types.Unsigned {
		v.fail("index of type %s", t)
	}
}

func (v *verifier) args(ft *types.Type, args []Expr, what string) {
	if len(args) != len(ft.Params) {
		v.fail("%s: %d arguments for %d parameters", what, len(args), len(ft.Params))
	}
	for i, a := range args {
		v.same(ft.Params[i], a, fmt.Sprintf("%s, argument %d", what, i+1))
	}
}

// Convertible reports whether (to) x is allowed for x of type from:
// numbers convert to each other; numbers, bools, and strings convert to
// string; strings convert to numbers and bool. It restates the checker's
// rule independently.
func Convertible(from, to *types.Type) bool {
	num := func(t *types.Type) bool {
		return t.Kind == types.Signed || t.Kind == types.Unsigned || t.Kind == types.Float
	}
	integer := func(t *types.Type) bool { return t.Kind == types.Signed || t.Kind == types.Unsigned }
	switch {
	case from.Kind == types.Char:
		return to.Kind == types.Char || to.Kind == types.String || integer(to)
	case to.Kind == types.Char:
		return integer(from) || from.Kind == types.String
	case num(from) && num(to):
		return true
	case to.Kind == types.String:
		return num(from) || from.Kind == types.Bool || from.Kind == types.String
	case from.Kind == types.String:
		return num(to) || to.Kind == types.Bool
	}
	return from.Kind == types.Bool && to.Kind == types.Bool
}

// boundary checks that t can cross the JavaScript boundary.
func (v *verifier) boundary(t *types.Type, result bool, name string) {
	switch t.Kind {
	case types.Signed, types.Unsigned, types.Float, types.Bool, types.String, types.Error:
	case types.Void:
		if !result {
			v.fail("extern %s has a void parameter", name)
		}
	default:
		v.fail("extern %s has type %s at the JavaScript boundary", name, t)
	}
}

func (v *verifier) telement(e *TElement) {
	for _, a := range e.Attrs {
		if a.Value == nil {
			continue
		}
		// The checker applies each attribute's exact rules (internal/dom);
		// here, only what a back end relies on for the mode.
		t := v.expr(a.Value)
		ok := false
		switch a.Mode {
		case AttrEvent:
			ok = types.Identical(t, types.NewFunc(nil, types.VoidT))
		case AttrBool:
			ok = t.Kind == types.Bool
		case AttrURL:
			ok = t.Kind == types.String
		default:
			ok = t.Kind == types.String || t.Kind == types.Signed || t.Kind == types.Unsigned || t.Kind == types.Float
		}
		if a.Values != nil && t.Kind != types.String {
			ok = false
		}
		if !ok {
			v.fail("attribute %s of <%s> bound to a %s", a.Name, e.Tag, t)
		}
	}
	v.tnodes(e.Children)
}

func (v *verifier) tnodes(nodes []TNode) {
	for _, n := range nodes {
		switch n := n.(type) {
		case *TElement:
			v.telement(n)
		case *TInterp:
			if t := v.expr(n.X); !types.IsScalar(t) && t.Kind != types.Element {
				v.fail("{{ }} of a %s", t)
			}
		case *TIf:
			for i, c := range n.Conds {
				v.same(types.BoolT, c, "$if condition")
				v.tnodes(n.Branches[i])
			}
			v.tnodes(n.Else)
		case *TForEach:
			t := v.array(n.X, "$foreach")
			if t.Kind == types.Array && !types.IsSubtype(t.Elem, n.Var.Type) {
				v.fail("$foreach variable %s: expected %s, found %s", n.Var.Name, n.Var.Type, t.Elem)
			}
			v.storable(n.Var.Type, "variable "+n.Var.Name)
			v.tnodes(n.Body)
		}
	}
}

// arrayCall checks the receiver and arguments of an array method.
func (v *verifier) arrayCall(e *ArrayCall) {
	t := v.array(e.X, e.Method)
	want := func(n int) {
		if len(e.Args) != n {
			v.fail("%s: %d arguments, expected %d", e.Method, len(e.Args), n)
		}
	}
	switch e.Method {
	case "insert":
		want(2)
		v.index(e.Args[0])
		v.same(t.Elem, e.Args[1], "inserted element")
	case "removeAt":
		want(1)
		v.index(e.Args[0])
	case "slice":
		want(2)
		v.index(e.Args[0])
		v.index(e.Args[1])
	case "indexOf", "contains":
		want(1)
		if !types.Comparable(t.Elem) {
			v.fail("%s on %s, whose elements are not comparable", e.Method, t)
		}
		v.same(t.Elem, e.Args[0], e.Method+" argument")
	case "reverse":
		want(0)
	case "sort":
		if len(e.Args) == 0 {
			if !types.IsNumeric(t.Elem) && t.Elem.Kind != types.Char && t.Elem.Kind != types.String {
				v.fail("sort() on %s, whose elements are not ordered", t)
			}
			return
		}
		want(1)
		less := v.expr(e.Args[0])
		if less.Kind != types.Func || len(less.Params) != 2 || !types.Identical(less.Params[0], t.Elem) ||
			!types.Identical(less.Params[1], t.Elem) || less.Result.Kind != types.Bool {
			v.fail("sort function of type %s for %s", less, t)
		}
	default:
		v.fail("unknown array method %q", e.Method)
	}
}
