package runtime

import (
	"github.com/AurielSolaris/runescript/internal/conv"
	"github.com/AurielSolaris/runescript/internal/ir"
	"github.com/AurielSolaris/runescript/internal/types"
)

type statement func(*interp, *frame) control
type expression func(*interp, *frame) Value
type floatExpression func(*interp, *frame) float64

func noStatement(*interp, *frame) control { return ctlNone }

func compileBlock(b *ir.Block) statement {
	body := make([]statement, 0, len(b.Stmts))
	for _, s := range b.Stmts {
		if _, nested := s.(*ir.FuncStmt); !nested {
			body = append(body, compileStmt(s))
		}
	}
	return func(in *interp, f *frame) control {
		for _, exec := range body {
			if c := exec(in, f); c != ctlNone {
				return c
			}
		}
		return ctlNone
	}
}

func compileStore(slot int, value expression) statement {
	return func(in *interp, f *frame) control {
		f.vars[slot] = value(in, f)
		return ctlNone
	}
}

func compileStmt(s ir.Stmt) statement {
	switch s := s.(type) {
	case nil:
		return noStatement
	case *ir.Decl:
		return compileStore(s.Var.Slot, compileExpr(s.Init))
	case *ir.Assign:
		return compileStore(s.Var.Slot, compileExpr(s.Value))
	case *ir.Eval:
		expr := compileExpr(s.X)
		return func(in *interp, f *frame) control {
			expr(in, f)
			return ctlNone
		}
	case *ir.Return:
		if s.Value == nil {
			return func(*interp, *frame) control { return ctlReturn }
		}
		value := compileExpr(s.Value)
		return func(in *interp, f *frame) control {
			f.ret = value(in, f)
			return ctlReturn
		}
	case *ir.ConstructorBody:
		body := compileBlock(s.Body)
		return func(in *interp, f *frame) control {
			body(in, f)
			return ctlNone
		}
	case *ir.ConstructorReturn:
		return func(*interp, *frame) control { return ctlReturn }
	case *ir.If:
		cond, yes, no := compileExpr(s.Cond), compileBlock(s.Then), compileStmt(s.Else)
		return func(in *interp, f *frame) control {
			if cond(in, f).Bool {
				return yes(in, f)
			}
			return no(in, f)
		}
	case *ir.Block:
		return compileBlock(s)
	case *ir.While:
		return compileLoop(nil, s.Cond, nil, s.Body)
	case *ir.For:
		return compileLoop(s.Init, s.Cond, s.Post, s.Body)
	case *ir.Break:
		return func(*interp, *frame) control { return ctlBreak }
	case *ir.Continue:
		return func(*interp, *frame) control { return ctlContinue }
	default:
		// Keep exception unwinding and complex assignments on the same
		// implementation as the reference evaluator.
		return func(in *interp, f *frame) control { return in.stmt(f, s) }
	}
}

func compileLoop(init ir.Stmt, condition ir.Expr, post ir.Stmt, block *ir.Block) statement {
	start, next := compileStmt(init), compileStmt(post)
	body := compileBlock(block)
	var cond expression
	if condition != nil {
		cond = compileExpr(condition)
	}
	return func(in *interp, f *frame) control {
		start(in, f)
		for cond == nil || cond(in, f).Bool {
			switch body(in, f) {
			case ctlBreak:
				return ctlNone
			case ctlReturn:
				return ctlReturn
			}
			next(in, f)
		}
		return ctlNone
	}
}

func compileExpr(e ir.Expr) expression {
	switch e := e.(type) {
	case *ir.Load:
		slot := e.Var.Slot
		return func(_ *interp, f *frame) Value { return f.vars[slot] }
	case *ir.IntConst:
		value := Value{Kind: e.T.Kind, Int: e.Value, Bits: e.T.Bits}
		return func(*interp, *frame) Value { return value }
	case *ir.BoolConst:
		value := boolean(e.Value)
		return func(*interp, *frame) Value { return value }
	case *ir.Compare:
		x, y, op := compileExpr(e.X), compileExpr(e.Y), e.Op
		return func(in *interp, f *frame) Value { return boolean(compare(op, x(in, f), y(in, f))) }
	case *ir.Logic:
		x, y, short := compileExpr(e.X), compileExpr(e.Y), e.Op == ir.OpOr
		return func(in *interp, f *frame) Value {
			value := x(in, f)
			if value.Bool == short {
				return value
			}
			return y(in, f)
		}
	case *ir.Not:
		x := compileExpr(e.X)
		return func(in *interp, f *frame) Value { return boolean(!x(in, f).Bool) }
	case *ir.Arith:
		if e.T.Kind == types.Float {
			calc := compileFloat(e)
			return func(in *interp, f *frame) Value {
				return Value{Kind: types.Float, F: calc(in, f), Bits: e.T.Bits}
			}
		}
		x, y := compileExpr(e.X), compileExpr(e.Y)
		if e.T.Bits == 0 {
			return func(in *interp, f *frame) Value { return in.genericArith(e, x(in, f), y(in, f)) }
		}
		return func(in *interp, f *frame) Value { return in.arith(e.Op, e.T, x(in, f), y(in, f), e.Site) }
	case *ir.Convert:
		x := compileExpr(e.X)
		return func(in *interp, f *frame) Value { return in.convert(x(in, f), e.To, e.Site) }
	case *ir.CallFunc:
		args := compileArgs(e.Args)
		return func(in *interp, f *frame) Value { return in.call(e.Func, args(in, f), e.Site) }
	case *ir.CallValue:
		call, args := compileExpr(e.Callee), compileArgs(e.Args)
		return func(in *interp, f *frame) Value { return in.call(call(in, f).Fn, args(in, f), e.Site) }
	case *ir.CallNative:
		args := compileArgs(e.Args)
		return func(in *interp, f *frame) Value { return in.native(e, args(in, f)) }
	default:
		return func(in *interp, f *frame) Value { return in.expr(f, e) }
	}
}

func compileArgs(args []ir.Expr) func(*interp, *frame) []Value {
	if len(args) == 0 {
		return func(*interp, *frame) []Value { return nil }
	}
	values := make([]expression, len(args))
	for i, e := range args {
		values[i] = compileExpr(e)
	}
	return func(in *interp, f *frame) []Value {
		result := make([]Value, len(values))
		for i, value := range values {
			result[i] = value(in, f)
		}
		return result
	}
}

// compileFloat keeps intermediate arithmetic in float64, with the same
// explicit rounding points as the IR evaluator. Leaves load scalar values
// directly from slots when possible.
func compileFloat(e ir.Expr) floatExpression {
	switch e := e.(type) {
	case *ir.FloatConst:
		return func(*interp, *frame) float64 { return e.Value }
	case *ir.Load:
		slot := e.Var.Slot
		return func(_ *interp, f *frame) float64 { return f.vars[slot].F }
	case *ir.Neg:
		x := compileFloat(e.X)
		return func(in *interp, f *frame) float64 { return -x(in, f) }
	case *ir.Arith:
		x, y := compileFloat(e.X), compileFloat(e.Y)
		var result floatExpression
		switch e.Op {
		case ir.OpAdd:
			result = func(in *interp, f *frame) float64 { return x(in, f) + y(in, f) }
		case ir.OpSub:
			result = func(in *interp, f *frame) float64 { return x(in, f) - y(in, f) }
		case ir.OpMul:
			result = func(in *interp, f *frame) float64 { return x(in, f) * y(in, f) }
		case ir.OpDiv:
			result = func(in *interp, f *frame) float64 { return x(in, f) / y(in, f) }
		default:
			ice("invalid float operator %s", e.Op)
		}
		if e.T.Bits == 32 {
			return func(in *interp, f *frame) float64 { return conv.ToFloat32(result(in, f)) }
		}
		return result
	default:
		value := compileExpr(e)
		return func(in *interp, f *frame) float64 { return value(in, f).F }
	}
}
