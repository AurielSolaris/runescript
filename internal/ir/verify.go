package ir

import (
	"fmt"

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
	case types.Signed, types.Unsigned, types.Bool, types.String, types.Func:
		return
	}
	v.fail("%s has type %s, which cannot hold a value", what, t)
}

// same requires e to have exactly type want.
func (v *verifier) same(want *types.Type, e Expr, what string) {
	got := v.expr(e)
	if !types.Identical(want, got) {
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
	case *StringConst, *BoolConst, *Load:
	case *FuncRef:
		if e.Func.Type == nil {
			v.fail("reference to undefined function %q", e.Func.Name)
		}
	case *Lambda:
		v.fn(e.Func)
	case *Arith:
		if e.T.Kind != types.Signed && e.T.Kind != types.Unsigned {
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
		case x.Kind == types.Signed || x.Kind == types.Unsigned:
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
	case *Trace:
		for _, a := range e.Args {
			switch t := v.expr(a); t.Kind {
			case types.Signed, types.Unsigned, types.Bool, types.String:
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
