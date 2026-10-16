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
	case types.Signed, types.Unsigned, types.Bool, types.String, types.Func, types.Float, types.Error:
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
	switch e := e.(type) {
	case *IntConst:
		if !types.InRange(e.T, e.Value) {
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
		}
		v.same(e.T, e.X, "left operand of "+e.Op)
		v.same(e.T, e.Y, "right operand of "+e.Op)
	case *Compare:
		x, y := v.expr(e.X), v.expr(e.Y)
		if !types.Identical(x, y) {
			v.fail("comparison %s of %s and %s", e.Op, x, y)
		}
		ordered := e.Op != "==" && e.Op != "!="
		switch {
		case x.Kind == types.Signed || x.Kind == types.Unsigned || x.Kind == types.Float:
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
	case *Trace:
		for _, a := range e.Args {
			switch t := v.expr(a); t.Kind {
			case types.Signed, types.Unsigned, types.Bool, types.String, types.Float, types.Error:
			default:
				v.fail("trace of %s", t)
			}
		}
	default:
		v.fail("unexpected expression %T", e)
	}
	return e.Type()
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
	switch {
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
