// Package runtime is a tree-walking interpreter for RuneScript programs.
//
// It runs the typed IR (internal/ir), the same verified representation the
// JavaScript back end compiles, so both targets start from one lowering of
// the checked program and the IR verifier guards both. Every expression in
// the IR has a concrete type, every name is resolved to its declaration,
// and every operation that can fail carries the source location its error
// is reported at.
package runtime

import (
	"cmp"
	"fmt"
	"io"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/AurielSolaris/runescript/internal/conv"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/ir"
	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// MaxCallDepth bounds recursion so runaway programs report an error instead
// of exhausting the Go stack. It is an implementation limit, not language
// semantics.
const MaxCallDepth = 10000

// Value is a runtime value. Integers and chars are held in Int, a uint64
// as its bit pattern (see types.Big), with Bits set to the integer type's
// Bits. A float is held as a float64; Bits is 32 for float32 values,
// which are always exact float32 values, and 0 otherwise (see
// types.FloatT).
type Value struct {
	Kind types.Kind
	Int  int64
	F    float64
	Bits int
	Str  string
	Bool bool
	Err  *ErrorValue
	// Fn is the function of a function value. Functions do not capture
	// variables, so no environment is needed.
	Fn *ir.Func
	// Arr is an array. Arrays are held by reference: copying the Value
	// shares the elements.
	Arr *ArrayValue
	// St is a struct. A StructValue is never changed after it is built
	// (a field assignment builds a changed copy), so sharing one between
	// Values is the same as copying it.
	St *StructValue
	// Opq is a value of an opaque type (types.Opaque), such as a
	// json.Value: whatever the standard library package made. It is
	// never changed, so sharing it is the same as copying it.
	Opq any
}

// StructValue holds the fields of a struct of type T.
type StructValue struct {
	T      *types.Type
	Fields []Value
}

// ArrayValue holds the elements of an array.
type ArrayValue struct{ Elems []Value }

func (v Value) String() string {
	switch v.Kind {
	case types.String:
		return v.Str
	case types.Bool:
		return strconv.FormatBool(v.Bool)
	case types.Float:
		return conv.FormatFloat(v.F, v.Bits)
	case types.Error:
		return v.Err.String()
	case types.Char:
		return string(rune(v.Int))
	case types.Unsigned:
		if v.Bits == 64 {
			return strconv.FormatUint(uint64(v.Int), 10)
		}
	case types.Struct:
		// Point{x: 1, name: "a", initial: 'b'}: strings and chars quoted.
		var b strings.Builder
		b.WriteString(v.St.T.Name + "{")
		for i, f := range v.St.Fields {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(v.St.T.Fields[i].Name + ": ")
			switch f.Kind {
			case types.String:
				b.WriteString(`"` + f.Str + `"`)
			case types.Char:
				b.WriteString("'" + f.String() + "'")
			default:
				b.WriteString(f.String())
			}
		}
		return b.String() + "}"
	}
	return strconv.FormatInt(v.Int, 10)
}

// big returns the exact value of v, an integer or a char.
func (v Value) big() *big.Int {
	if v.Kind == types.Unsigned && v.Bits == 64 {
		return new(big.Int).SetUint64(uint64(v.Int))
	}
	return big.NewInt(v.Int)
}

// ErrorValue is a RuneScript error: one created with `new`, or one the
// runtime raised for a failed operation.
type ErrorValue struct {
	Type    *types.Type
	Message string
	ID      int64
	// Code is the diagnostic an uncaught error is reported with: the
	// runtime error's own code, or RunUncaught for one created with new.
	Code diagnostics.Code
	// Site is where it was raised: the failed operation, or the first
	// throw statement that threw it. Set reports whether it is known.
	Site diagnostics.Span
	Set  bool
}

// String is how trace prints an error, as ActionScript 3's
// Error.toString does: "RangeError: message", or the name alone.
func (e *ErrorValue) String() string {
	if e.Message == "" {
		return e.Type.Name
	}
	return e.Type.Name + ": " + e.Message
}

// thrown carries an error being thrown through the interpreter.
type thrown struct{ err *ErrorValue }

// faultTypes are the error types of the runtime errors that can be caught.
// Exceeding the call depth (RUN-001) cannot be caught.
var faultTypes = map[diagnostics.Code]*types.Type{
	diagnostics.RunOverflow:     types.RangeErrorT,
	diagnostics.RunDivideByZero: types.RangeErrorT,
	diagnostics.RunNativeFailed: types.ErrorT,
	diagnostics.RunIndex:        types.RangeErrorT,
	diagnostics.RunShift:        types.RangeErrorT,
}

// ErrorID is the errorID of a runtime error: its code's number, e.g. 2 for
// ASX-RUN-002.
func ErrorID(code diagnostics.Code) int64 {
	n, _ := strconv.ParseInt(string(code)[strings.LastIndex(string(code), "-")+1:], 10, 64)
	return n
}

// Error is a runtime failure with the source location that caused it.
type Error struct {
	Diag *diagnostics.Diagnostic
}

func (e *Error) Error() string { return e.Diag.Error() }

// control signals non-local exits from statement execution.
type control int

const (
	ctlNone control = iota
	ctlBreak
	ctlContinue
	ctlReturn
)

// frame is one function call. Each variable declaration is its own
// *ir.Var, so one map holds every variable of the call.
type frame struct {
	vars map[*ir.Var]Value
	ret  Value
}

type interp struct {
	out   io.Writer
	env   *stdlib.Env
	depth int
	// raising is the error being thrown, from the throw (or failed
	// operation) until a catch clause catches it, and fatal is set once an
	// error that cannot be caught is raised. Try statements look at them
	// instead of recovering the panic in every frame: a finally block runs
	// while the panic goes on, and a catch clause recovers only an error it
	// catches. Each error is raised once, so unwinding is linear in the
	// call depth.
	raising *ErrorValue
	fatal   bool
	// olds are the old elements of the compound element assignments
	// (a[i] += v) being evaluated, innermost last (ir.IndexOld).
	olds []Value
}

// Run executes the program's entry point and returns main's result as the
// process exit code. The program must have passed
// ir.Verify, and must not call extern (JavaScript) functions or use
// rune:dom, which only the JavaScript target can run.
func Run(p *ir.Program, out io.Writer) (code int, err error) {
	return RunEnv(p, stdlib.NewEnv(out, io.Discard, nil, nil))
}

// RunEnv is Run with the program's environment: trace and io.print write
// to env.Stdout, io.error and io.warn to env.Stderr, and so on.
func RunEnv(p *ir.Program, env *stdlib.Env) (code int, err error) {
	out := env.Stdout
	if len(p.Externs) > 0 {
		// The driver reports this as ASX-TARGET-002 before calling Run.
		return 0, fmt.Errorf("the program calls extern function '%s'; only the JavaScript target can run it", p.Externs[0].Name)
	}
	in := &interp{out: out, env: env}
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case *Error:
				err = r
			case thrown:
				err = uncaught(r.err)
			default:
				panic(r)
			}
		}
	}()
	return int(in.call(p.Entry, nil, p.EntrySite).Int), nil
}

// fail raises a runtime error. Those with an error type are thrown and can
// be caught; the others stop the program.
func (in *interp) fail(code diagnostics.Code, span diagnostics.Span, format string, args ...any) {
	if t, ok := faultTypes[code]; ok {
		in.raise(t, code, span, fmt.Sprintf(format, args...))
	}
	in.fatal = true
	panic(&Error{diagnostics.New(code, span, format, args...)})
}

// raise throws an error of type t raised by the runtime at span.
func (in *interp) raise(t *types.Type, code diagnostics.Code, span diagnostics.Span, msg string) {
	in.throw(&ErrorValue{Type: t, Message: msg, ID: ErrorID(code), Code: code, Site: span, Set: true})
}

func (in *interp) throw(err *ErrorValue) {
	in.raising = err
	panic(thrown{err})
}

// uncaught is the runtime error reported for an error that nothing caught.
func uncaught(e *ErrorValue) *Error {
	if e.Code != diagnostics.RunUncaught {
		return &Error{diagnostics.New(e.Code, e.Site, "%s", e.Message)}
	}
	return &Error{diagnostics.New(e.Code, e.Site, "uncaught %s", e.String())}
}

// ice reports IR the interpreter cannot run. ir.Verify and the driver rule
// it out, so reaching it is a compiler bug.
func ice(format string, args ...any) {
	panic("runtime: internal compiler error: " + fmt.Sprintf(format, args...))
}

// call invokes f, a named function or a lambda, from site. Each call has
// its own frame: there is no access to the caller's variables.
func (in *interp) call(f *ir.Func, args []Value, site diagnostics.Span) Value {
	if in.depth >= MaxCallDepth {
		in.fail(diagnostics.RunCallDepth, site, "maximum call depth (%d) exceeded", MaxCallDepth)
	}
	in.depth++
	defer func() { in.depth-- }()
	fr := &frame{vars: make(map[*ir.Var]Value, len(f.Params))}
	for i, p := range f.Params {
		fr.vars[p] = args[i]
	}
	in.block(fr, f.Body)
	return fr.ret
}

// ---- statements ----

func (in *interp) block(f *frame, b *ir.Block) control {
	for _, s := range b.Stmts {
		if c := in.stmt(f, s); c != ctlNone {
			return c
		}
	}
	return ctlNone
}

func (in *interp) stmt(f *frame, s ir.Stmt) control {
	switch s := s.(type) {
	case *ir.FuncStmt:
		// Nested functions are referred to directly by the IR.
	case *ir.Decl:
		f.vars[s.Var] = in.expr(f, s.Init)
	case *ir.Assign:
		f.vars[s.Var] = in.expr(f, s.Value)
	case *ir.Eval:
		in.expr(f, s.X)
	case *ir.IndexSet:
		in.indexSet(f, s)
	case *ir.Return:
		if s.Value != nil {
			f.ret = in.expr(f, s.Value)
		}
		return ctlReturn
	case *ir.If:
		if in.expr(f, s.Cond).Bool {
			return in.block(f, s.Then)
		}
		switch e := s.Else.(type) {
		case *ir.Block:
			return in.block(f, e)
		case *ir.If:
			return in.stmt(f, e)
		}
	case *ir.While:
		for in.expr(f, s.Cond).Bool {
			if c := in.block(f, s.Body); c == ctlBreak {
				break
			} else if c == ctlReturn {
				return c
			}
		}
	case *ir.For:
		return in.forStmt(f, s)
	case *ir.Block:
		return in.block(f, s)
	case *ir.Break:
		return ctlBreak
	case *ir.Continue:
		return ctlContinue
	case *ir.Throw:
		err := in.expr(f, s.X).Err
		if !err.Set {
			err.Site, err.Set = s.Site, true
		}
		in.throw(err)
	case *ir.Try:
		return in.tryStmt(f, s)
	default:
		ice("unexpected statement %T", s)
	}
	return ctlNone
}

func (in *interp) forStmt(f *frame, s *ir.For) control {
	if s.Init != nil {
		in.stmt(f, s.Init)
	}
	for s.Cond == nil || in.expr(f, s.Cond).Bool {
		c := in.block(f, s.Body)
		if c == ctlBreak {
			break
		}
		if c == ctlReturn {
			return c
		}
		if s.Post != nil {
			in.stmt(f, s.Post)
		}
	}
	return ctlNone
}

// tryStmt runs try/catch/finally. The finally block runs however the try
// statement ends (normally, by return, break, or continue, or by an error)
// except when the program is stopping for an error that cannot be caught.
// A finally block that returns, breaks, or continues replaces that ending,
// discarding an error in flight, as in ActionScript 3 and Java.
func (in *interp) tryStmt(f *frame, s *ir.Try) (c control) {
	if s.Finally != nil {
		depth, olds := in.depth, len(in.olds)
		defer func() {
			if in.fatal {
				return
			}
			// Any error in flight goes on after the block, unless the
			// block throws one of its own, which replaces it, or returns,
			// breaks, or continues, which discards it.
			raising := in.raising
			in.depth, in.olds = depth, in.olds[:olds]
			in.raising = nil
			if fc := in.block(f, s.Finally); fc != ctlNone {
				if raising != nil {
					recover()
				}
				c = fc
				return
			}
			in.raising = raising
		}()
	}
	return in.tryCatch(f, s)
}

func (in *interp) tryCatch(f *frame, s *ir.Try) control {
	if len(s.Catches) == 0 {
		return in.block(f, s.Body)
	}
	depth, olds := in.depth, len(in.olds)
	var clause *ir.Catch
	c, err := func() (c control, err *ErrorValue) {
		defer func() {
			e := in.raising
			if in.fatal || e == nil {
				return
			}
			for _, cl := range s.Catches {
				if types.IsSubtype(e.Type, cl.T) {
					recover()
					in.raising, clause, err = nil, cl, e
					return
				}
			}
		}()
		return in.block(f, s.Body), nil
	}()
	if err == nil {
		return c
	}
	in.depth, in.olds = depth, in.olds[:olds]
	f.vars[clause.Var] = Value{Kind: types.Error, Err: err}
	return in.block(f, clause.Body)
}

// ---- expressions ----

func (in *interp) expr(f *frame, e ir.Expr) Value {
	switch e := e.(type) {
	case *ir.IntConst:
		return Value{Kind: e.T.Kind, Int: e.Value, Bits: e.T.Bits}
	case *ir.FloatConst:
		return Value{Kind: types.Float, F: e.Value, Bits: e.T.Bits}
	case *ir.StringConst:
		return Value{Kind: types.String, Str: e.Value}
	case *ir.BoolConst:
		return boolean(e.Value)
	case *ir.CharConst:
		return Value{Kind: types.Char, Int: int64(e.Value)}
	case *ir.Load:
		v, ok := f.vars[e.Var]
		if !ok {
			ice("variable %s read before its declaration", e.Var.Name)
		}
		return v
	case *ir.FuncRef:
		return Value{Kind: types.Func, Fn: e.Func}
	case *ir.Lambda:
		return Value{Kind: types.Func, Fn: e.Func}
	case *ir.Neg:
		x := in.expr(f, e.X)
		x.F = -x.F
		return x
	case *ir.Not:
		return boolean(!in.expr(f, e.X).Bool)
	case *ir.Arith:
		x, y := in.expr(f, e.X), in.expr(f, e.Y)
		switch {
		case e.T.Kind == types.Float:
			return floatArith(e.Op, x, y, e.T)
		case e.T.Bits == 0:
			return in.genericArith(e, x, y)
		}
		return in.arith(e.Op, e.T, x, y, e.Site)
	case *ir.Bits:
		x, y := in.expr(f, e.X), in.expr(f, e.Y)
		// On two's-complement int64 values (uint64 bit patterns included),
		// & | ^ give an in-range result of the same type.
		var r int64
		switch e.Op {
		case "&":
			r = x.Int & y.Int
		case "|":
			r = x.Int | y.Int
		default:
			r = x.Int ^ y.Int
		}
		return Value{Kind: e.T.Kind, Int: r, Bits: e.T.Bits}
	case *ir.Complement:
		x := in.expr(f, e.X)
		r := ^x.Int
		if bits := types.IntBits(e.X.Type()); x.Kind == types.Unsigned && bits < 64 {
			r &= 1<<bits - 1
		}
		x.Int = r
		return x
	case *ir.Shift:
		return in.shift(e, in.expr(f, e.X), in.expr(f, e.Count))
	case *ir.Concat:
		var b strings.Builder
		for _, p := range e.Parts {
			b.WriteString(in.expr(f, p).Str)
		}
		return Value{Kind: types.String, Str: b.String()}
	case *ir.Compare:
		return boolean(compare(e.Op, in.expr(f, e.X), in.expr(f, e.Y)))
	case *ir.Logic:
		x := in.expr(f, e.X)
		if x.Bool == (e.Op == "||") { // short-circuit
			return x
		}
		return in.expr(f, e.Y)
	case *ir.Convert:
		return in.convert(in.expr(f, e.X), e.To, e.Site)
	case *ir.NewError:
		err := &ErrorValue{Type: e.T, Code: diagnostics.RunUncaught}
		if e.Message != nil {
			err.Message = in.expr(f, e.Message).Str
		}
		if e.ID != nil {
			err.ID = in.expr(f, e.ID).Int
		}
		return Value{Kind: types.Error, Err: err}
	case *ir.Field:
		err := in.expr(f, e.X).Err
		switch e.Name {
		case "message":
			return Value{Kind: types.String, Str: err.Message}
		case "name":
			return Value{Kind: types.String, Str: err.Type.Name}
		}
		return Value{Kind: types.Signed, Int: err.ID}
	case *ir.FloatTest:
		x := in.expr(f, e.X).F
		if e.Name == "isNaN" {
			return boolean(math.IsNaN(x))
		}
		return boolean(!math.IsNaN(x) && !math.IsInf(x, 0))
	case *ir.CallFunc:
		return in.call(e.Func, in.exprs(f, e.Args), e.Site)
	case *ir.CallValue:
		callee := in.expr(f, e.Callee) // before the arguments
		return in.call(callee.Fn, in.exprs(f, e.Args), e.Site)
	case *ir.CallNative:
		return in.native(e, in.exprs(f, e.Args))
	case *ir.Trace:
		parts := make([]string, len(e.Args))
		for i, a := range e.Args {
			parts[i] = in.expr(f, a).String()
		}
		fmt.Fprintln(in.out, strings.Join(parts, " "))
		return Value{Kind: types.Void}
	case *ir.ArrayLit:
		return Value{Kind: types.Array, Arr: &ArrayValue{Elems: in.exprs(f, e.Elems)}}
	case *ir.StructLit:
		st := &StructValue{T: e.T, Fields: make([]Value, len(e.T.Fields))}
		for _, init := range e.Inits {
			st.Fields[init.Index] = in.expr(f, init.Value)
		}
		return Value{Kind: types.Struct, St: st}
	case *ir.GetField:
		return in.expr(f, e.X).St.Fields[e.Index]
	case *ir.WithField:
		old := in.expr(f, e.X).St
		v := in.expr(f, e.Value)
		st := &StructValue{T: old.T, Fields: append([]Value(nil), old.Fields...)}
		st.Fields[e.Index] = v
		return Value{Kind: types.Struct, St: st}
	case *ir.Index:
		a := in.expr(f, e.X).Arr
		i := in.expr(f, e.I)
		return a.Elems[in.checkIndex(a, i, e.Site)]
	case *ir.IndexOld:
		return in.olds[len(in.olds)-1]
	case *ir.ArrayLen:
		return Value{Kind: types.Signed, Int: int64(len(in.expr(f, e.X).Arr.Elems))}
	case *ir.ArrayPush:
		a := in.expr(f, e.X).Arr
		v := in.expr(f, e.V)
		a.Elems = append(a.Elems, v)
		return Value{Kind: types.Void}
	case *ir.ArrayPop:
		a := in.expr(f, e.X).Arr
		if len(a.Elems) == 0 {
			in.fail(diagnostics.RunIndex, e.Site, "pop from an empty array")
		}
		last := a.Elems[len(a.Elems)-1]
		a.Elems = a.Elems[:len(a.Elems)-1]
		return last
	case *ir.ArrayCall:
		return in.arrayCall(f, e)
	}
	// Templates, runes, renderApp, and extern calls need the JavaScript
	// target; the driver rejects such programs (ASX-TARGET-002).
	ice("unexpected expression %T", e)
	return Value{}
}

// exprs evaluates a list left to right.
func (in *interp) exprs(f *frame, list []ir.Expr) []Value {
	out := make([]Value, len(list))
	for i, x := range list {
		out[i] = in.expr(f, x)
	}
	return out
}

func boolean(b bool) Value { return Value{Kind: types.Bool, Bool: b} }

// compare applies a comparison to two values of the same type.
func compare(op string, x, y Value) bool {
	var c int // -1, 0, or 1
	switch x.Kind {
	case types.Float:
		switch {
		case math.IsNaN(x.F) || math.IsNaN(y.F):
			return op == "!=" // NaN is unordered, and equal to nothing
		case x.F < y.F:
			c = -1
		case x.F > y.F:
			c = 1
		}
	case types.String:
		c = strings.Compare(x.Str, y.Str)
	case types.Struct:
		// == and != only: equal when every field is.
		for i := range x.St.Fields {
			if !compare("==", x.St.Fields[i], y.St.Fields[i]) {
				c = 1
				break
			}
		}
	case types.Bool:
		if x.Bool != y.Bool {
			c = 1
		}
	case types.Unsigned:
		if x.Bits != 64 {
			// Narrower values, and wide uint intermediates (which may be
			// negative), compare as int64s.
			c = cmp.Compare(x.Int, y.Int)
			break
		}
		c = cmp.Compare(uint64(x.Int), uint64(y.Int)) // uint64 bit patterns
	default: // signed integers and chars
		switch {
		case x.Int < y.Int:
			c = -1
		case x.Int > y.Int:
			c = 1
		}
	}
	switch op {
	case "==":
		return c == 0
	case "!=":
		return c != 0
	case "<":
		return c < 0
	case "<=":
		return c <= 0
	case ">":
		return c > 0
	}
	return c >= 0
}

// arith performs integer arithmetic of type t. Operands are within t's
// range, so for types up to 32 bits int64 arithmetic is exact, and 64-bit
// types are computed exactly with big.Int. A result outside the range is a
// runtime error: integers never wrap (OQ-NUM-01).
func (in *interp) arith(op string, t *types.Type, x, y Value, span diagnostics.Span) Value {
	if types.IntBits(t) == 64 || (op == "*" && t.Kind == types.Unsigned) {
		// Exact with big.Int: 64-bit operands, and products of two 32-bit
		// unsigned values, which can exceed int64.
		return in.wideArith(op, t, x, y, span)
	}
	var r int64
	switch op {
	case "+":
		r = x.Int + y.Int
	case "-":
		r = x.Int - y.Int
	case "*":
		r = x.Int * y.Int
	case "/", "%":
		if y.Int == 0 {
			in.fail(diagnostics.RunDivideByZero, span, "integer division by zero")
		}
		if op == "/" {
			r = x.Int / y.Int
		} else {
			r = x.Int % y.Int
		}
	}
	if !types.InRange(t, r) {
		in.fail(diagnostics.RunOverflow, span, "integer overflow: %d %s %d = %d does not fit in '%s'", x.Int, op, y.Int, r, typeName(t))
	}
	return Value{Kind: t.Kind, Int: r, Bits: t.Bits}
}

// genericArith is arithmetic on the generic int or uint (OQ-NUM-01,
// option A). Its operands may be wide intermediates (ir.Arith.Wide), any
// int64; the result is exact, and must fit int64 if it is itself wide, or
// e.T otherwise. A wide value keeps e.T's kind: a uint intermediate may be
// negative, as in `a - b + c`.
func (in *interp) genericArith(e *ir.Arith, x, y Value) Value {
	a, b := x.Int, y.Int
	const small = 1 << 31
	if a > -small && a < small && b > -small && b < small && e.Op != "/" && e.Op != "%" {
		// Both operands below 2^31 in magnitude: exact in int64.
		var r int64
		switch e.Op {
		case "+":
			r = a + b
		case "-":
			r = a - b
		default:
			r = a * b
		}
		if e.Wide || types.InRange(e.T, r) {
			return Value{Kind: e.T.Kind, Int: r}
		}
		in.fail(diagnostics.RunOverflow, e.Site, "integer overflow: %d %s %d = %d does not fit in '%s'", a, e.Op, b, r, typeName(e.T))
	}
	ba, bb, r := big.NewInt(a), big.NewInt(b), new(big.Int)
	switch e.Op {
	case "+":
		r.Add(ba, bb)
	case "-":
		r.Sub(ba, bb)
	case "*":
		r.Mul(ba, bb)
	default:
		if b == 0 {
			in.fail(diagnostics.RunDivideByZero, e.Site, "integer division by zero")
		}
		if e.Op == "/" {
			r.Quo(ba, bb) // truncates toward zero
		} else {
			r.Rem(ba, bb) // takes the dividend's sign
		}
	}
	switch {
	case e.Wide && !r.IsInt64():
		in.fail(diagnostics.RunOverflow, e.Site, "integer overflow: %d %s %d = %s is beyond the int64 range of intermediate results", a, e.Op, b, r)
	case !e.Wide && !types.Fits(e.T, r):
		in.fail(diagnostics.RunOverflow, e.Site, "integer overflow: %d %s %d = %s does not fit in '%s'", a, e.Op, b, r, typeName(e.T))
	}
	return Value{Kind: e.T.Kind, Int: r.Int64()}
}

// wideArith is arith computed exactly with big.Int, for any integer type.
func (in *interp) wideArith(op string, t *types.Type, x, y Value, span diagnostics.Span) Value {
	a, b, r := x.big(), y.big(), new(big.Int)
	switch op {
	case "+":
		r.Add(a, b)
	case "-":
		r.Sub(a, b)
	case "*":
		r.Mul(a, b)
	case "/", "%":
		if b.Sign() == 0 {
			in.fail(diagnostics.RunDivideByZero, span, "integer division by zero")
		}
		if op == "/" {
			r.Quo(a, b) // truncates toward zero
		} else {
			r.Rem(a, b) // takes the dividend's sign
		}
	}
	if !types.Fits(t, r) {
		in.fail(diagnostics.RunOverflow, span, "integer overflow: %s %s %s = %s does not fit in '%s'", a, op, b, r, typeName(t))
	}
	return Value{Kind: t.Kind, Int: types.FromBig(t, r), Bits: t.Bits}
}

// shift runs x << n or x >> n (ir.Shift).
func (in *interp) shift(e *ir.Shift, x, n Value) Value {
	t := e.X.Type()
	bits := types.IntBits(t)
	count := n.big()
	if count.Sign() < 0 || count.Cmp(big.NewInt(int64(bits))) >= 0 {
		in.fail(diagnostics.RunShift, e.Site, "shift count %s is out of range for '%s' (0 to %d)", count, typeName(t), bits-1)
	}
	c := uint(count.Uint64())
	if e.Op == ">>" {
		if t.Kind == types.Unsigned {
			x.Int = int64(uint64(x.Int) >> c)
		} else {
			x.Int >>= c // arithmetic: rounds toward negative infinity
		}
		return x
	}
	r := new(big.Int).Lsh(x.big(), c)
	if !types.Fits(t, r) {
		in.fail(diagnostics.RunOverflow, e.Site, "integer overflow: %s << %d = %s does not fit in '%s'", x.big(), c, r, typeName(t))
	}
	x.Int = types.FromBig(t, r)
	return x
}

// typeName names an integer type in a message: the generic types by
// their main spelling ('int' for sint too).
func typeName(t *types.Type) *types.Type {
	switch {
	case t.Bits != 0:
		return t
	case t.Kind == types.Unsigned:
		return types.Uint
	}
	return types.Int
}

// floatArith performs + - * / on floats of type t, with IEEE 754
// semantics as in ActionScript 3: overflow gives Infinity, and 0/0 gives
// NaN. float32 results are rounded to float32.
func floatArith(op string, x, y Value, t *types.Type) Value {
	var r float64
	switch op {
	case "+":
		r = x.F + y.F
	case "-":
		r = x.F - y.F
	case "*":
		r = x.F * y.F
	case "/":
		r = x.F / y.F
	}
	if t.Bits == 32 {
		r = conv.ToFloat32(r)
	}
	return Value{Kind: types.Float, F: r, Bits: t.Bits}
}

// convert performs the checked conversion (to) v. A value outside the
// destination's range throws a RangeError, and a string that is not a
// number (or bool, or one character) throws a TypeError.
func (in *interp) convert(v Value, to *types.Type, span diagnostics.Span) Value {
	outOfRange := func() {
		in.raise(types.RangeErrorT, diagnostics.RunConversion, span, fmt.Sprintf("cannot convert %s to '%s': out of range", quoted(v), to))
	}
	notValid := func(what string) {
		in.raise(types.TypeErrorT, diagnostics.RunConversion, span, fmt.Sprintf("cannot convert %s to '%s': not %s", quoted(v), to, what))
	}
	switch to.Kind {
	case types.String:
		return Value{Kind: types.String, Str: v.String()}
	case types.Char:
		switch v.Kind {
		case types.Char:
			return v
		case types.String:
			r, size := utf8.DecodeRuneInString(v.Str)
			if v.Str == "" || size != len(v.Str) {
				notValid("a single character")
			}
			return Value{Kind: types.Char, Int: int64(r)}
		}
		if !types.ValidChar(v.Int) {
			in.raise(types.RangeErrorT, diagnostics.RunConversion, span, fmt.Sprintf("cannot convert %s to 'char': not a Unicode scalar value", quoted(v)))
		}
		return Value{Kind: types.Char, Int: v.Int}
	case types.Bool:
		if v.Kind == types.Bool {
			return v
		}
		switch v.Str {
		case "true", "false":
			return boolean(v.Str == "true")
		}
		notValid("'true' or 'false'")
	case types.Signed, types.Unsigned:
		var n *big.Int // the exact value
		switch v.Kind {
		case types.Signed, types.Unsigned, types.Char:
			n = v.big()
		case types.Float:
			t := math.Trunc(v.F)
			if math.IsNaN(t) || math.IsInf(t, 0) {
				outOfRange()
			}
			n, _ = big.NewFloat(t).Int(nil)
		case types.String:
			if _, _, ok := conv.ParseInt(v.Str); !ok {
				notValid("an integer")
			}
			n, _ = new(big.Int).SetString(strings.TrimPrefix(v.Str, "+"), 10)
		}
		if !types.Fits(to, n) {
			outOfRange()
		}
		return Value{Kind: to.Kind, Int: types.FromBig(to, n), Bits: to.Bits}
	case types.Float:
		var f float64
		switch v.Kind {
		case types.Signed:
			f = float64(v.Int)
		case types.Unsigned:
			f = float64(v.Int) // a wide uint intermediate may be negative
			if v.Bits == 64 {
				f = float64(uint64(v.Int)) // a uint64 bit pattern
			}
		case types.Float:
			f = v.F
		case types.String:
			parsed, ok := conv.ParseFloat(v.Str)
			if !ok {
				notValid("a number")
			}
			f = parsed
		}
		if to.Bits == 32 {
			f = conv.ToFloat32(f)
		}
		return Value{Kind: types.Float, F: f, Bits: to.Bits}
	}
	ice("conversion from %v to %s", v.Kind, to)
	return Value{}
}

// quoted renders a conversion's operand for an error message: strings in
// double quotes, as written, and other values as trace prints them.
func quoted(v Value) string {
	if v.Kind == types.String {
		return `"` + v.Str + `"`
	}
	return v.String()
}

// ---- standard library ----

// native calls a standard library function. Go panics (for example a
// negative count passed to string.repeat) become runtime errors, a
// stdlib.Fault becomes an error of its type, and integer results are
// range-checked against the declared result type.
func (in *interp) native(e *ir.CallNative, args []Value) Value {
	f := e.Func
	goArgs := make([]any, len(args))
	for i, a := range args {
		switch a.Kind {
		case types.Opaque:
			goArgs[i] = a.Opq
		case types.Array: // Array<string> as []string, Array<uint8> as []byte, Array<opaque> as []any
			switch f.Params[i].Elem.Kind {
			case types.Opaque:
				vals := make([]any, len(a.Arr.Elems))
				for j, v := range a.Arr.Elems {
					vals[j] = v.Opq
				}
				goArgs[i] = vals
				continue
			}
			if f.Params[i].Elem.Kind == types.String {
				strs := make([]string, len(a.Arr.Elems))
				for j, v := range a.Arr.Elems {
					strs[j] = v.Str
				}
				goArgs[i] = strs
				break
			}
			b := make([]byte, len(a.Arr.Elems))
			for j, v := range a.Arr.Elems {
				b[j] = byte(v.Int)
			}
			goArgs[i] = b
		case types.String:
			goArgs[i] = a.Str
		case types.Bool:
			goArgs[i] = a.Bool
		case types.Char:
			goArgs[i] = rune(a.Int)
		case types.Float:
			goArgs[i] = a.F
		case types.Func:
			goArgs[i] = in.goFunc(a, f.Params[i], e.Site)
		default:
			goArgs[i] = a.Int
		}
	}
	name := e.Pkg + "." + f.Name
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case *Error, thrown:
				panic(r)
			case stdlib.Fault:
				in.raise(types.Lookup(r.Type), diagnostics.RunNativeFailed, e.Site, name+": "+r.Message)
			}
			in.fail(diagnostics.RunNativeFailed, e.Site, "%s failed: %v", name, r)
		}
	}()
	var out any
	if f.EnvImpl != nil {
		out = f.EnvImpl(in.env, goArgs)
	} else {
		out = f.Impl(goArgs)
	}
	switch f.Result.Kind {
	case types.Void:
		return Value{Kind: types.Void}
	case types.String:
		return Value{Kind: types.String, Str: out.(string)}
	case types.Bool:
		return boolean(out.(bool))
	case types.Char:
		return Value{Kind: types.Char, Int: int64(out.(rune))}
	case types.Float:
		return Value{Kind: types.Float, F: out.(float64), Bits: f.Result.Bits}
	case types.Opaque:
		return Value{Kind: types.Opaque, Opq: out}
	case types.Array: // Array<string> from []string, Array<uint8> from []byte, Array<opaque> from []any
		if vals, ok := out.([]any); ok {
			arr := &ArrayValue{Elems: make([]Value, len(vals))}
			for i, x := range vals {
				arr.Elems[i] = Value{Kind: types.Opaque, Opq: x}
			}
			return Value{Kind: types.Array, Arr: arr}
		}
		if b, ok := out.([]byte); ok {
			arr := &ArrayValue{Elems: make([]Value, len(b))}
			for i, x := range b {
				arr.Elems[i] = Value{Kind: types.Unsigned, Int: int64(x), Bits: 8}
			}
			return Value{Kind: types.Array, Arr: arr}
		}
		strs := out.([]string)
		arr := &ArrayValue{Elems: make([]Value, len(strs))}
		for i, s := range strs {
			arr.Elems[i] = Value{Kind: types.String, Str: s}
		}
		return Value{Kind: types.Array, Arr: arr}
	}
	n := out.(int64)
	if !types.InRange(f.Result, n) {
		in.fail(diagnostics.RunOverflow, e.Site, "%s returned %d, which does not fit in '%s'", name, n, f.Result)
	}
	return Value{Kind: f.Result.Kind, Int: n, Bits: f.Result.Bits}
}

// goFunc wraps the function value v, of type t, as the Go function a
// standard library function expects: func(rune) bool or func(rune) rune.
// Each call is a RuneScript call from site, so errors and the call-depth
// limit work as usual.
func (in *interp) goFunc(v Value, t *types.Type, site diagnostics.Span) any {
	switch {
	case len(t.Params) == 1 && t.Params[0].Kind == types.Char && t.Result.Kind == types.Bool:
		return func(r rune) bool {
			return in.call(v.Fn, []Value{{Kind: types.Char, Int: int64(r)}}, site).Bool
		}
	case len(t.Params) == 1 && t.Params[0].Kind == types.Char && t.Result.Kind == types.Char:
		return func(r rune) rune {
			return rune(in.call(v.Fn, []Value{{Kind: types.Char, Int: int64(r)}}, site).Int)
		}
	}
	ice("no Go form for function type %s", t)
	return nil
}

// ---- arrays ----

// checkIndex returns i if it is an index of a, and throws a RangeError
// otherwise.
func (in *interp) checkIndex(a *ArrayValue, i Value, span diagnostics.Span) int64 {
	if n := i.big(); !n.IsInt64() || n.Int64() < 0 || n.Int64() >= int64(len(a.Elems)) {
		in.fail(diagnostics.RunIndex, span, "index %s is out of range for an array of length %d", n, len(a.Elems))
	}
	return i.Int
}

// indexSet runs a[i] = v or a[i] op= v (ir.IndexSet): the array, then the
// index, are evaluated once. For a compound assignment, the old element
// is read (checking the index) before the value is evaluated; the index is
// checked again when storing, since evaluating the value may have
// shortened the array.
func (in *interp) indexSet(f *frame, s *ir.IndexSet) {
	a := in.expr(f, s.X).Arr
	i := in.expr(f, s.I)
	var v Value
	if s.Compound {
		in.olds = append(in.olds, a.Elems[in.checkIndex(a, i, s.Site)])
		v = in.expr(f, s.Value)
		in.olds = in.olds[:len(in.olds)-1]
	} else {
		v = in.expr(f, s.Value)
	}
	a.Elems[in.checkIndex(a, i, s.Site)] = v
}

// arrayCall runs an array method (ir.ArrayCall).
func (in *interp) arrayCall(f *frame, e *ir.ArrayCall) Value {
	a := in.expr(f, e.X).Arr
	args := in.exprs(f, e.Args)
	void := Value{Kind: types.Void}
	switch e.Method {
	case "insert":
		n := args[0].big()
		if !n.IsInt64() || n.Int64() < 0 || n.Int64() > int64(len(a.Elems)) {
			in.fail(diagnostics.RunIndex, e.Site, "index %s is out of range for inserting into an array of length %d", n, len(a.Elems))
		}
		i := n.Int64()
		a.Elems = append(a.Elems, Value{})
		copy(a.Elems[i+1:], a.Elems[i:])
		a.Elems[i] = args[1]
		return void
	case "removeAt":
		i := in.checkIndex(a, args[0], e.Site)
		v := a.Elems[i]
		a.Elems = append(a.Elems[:i], a.Elems[i+1:]...)
		return v
	case "slice":
		lo, hi := args[0].big(), args[1].big()
		n := int64(len(a.Elems))
		if !lo.IsInt64() || !hi.IsInt64() || lo.Int64() < 0 || lo.Int64() > hi.Int64() || hi.Int64() > n {
			in.fail(diagnostics.RunIndex, e.Site, "slice [%s, %s) is out of range for an array of length %d", lo, hi, n)
		}
		return Value{Kind: types.Array, Arr: &ArrayValue{Elems: append([]Value(nil), a.Elems[lo.Int64():hi.Int64()]...)}}
	case "indexOf", "contains":
		at := -1
		for i, x := range a.Elems {
			if compare("==", x, args[0]) {
				at = i
				break
			}
		}
		if e.Method == "contains" {
			return boolean(at >= 0)
		}
		return Value{Kind: types.Signed, Int: int64(at)}
	case "reverse":
		for i, j := 0, len(a.Elems)-1; i < j; i, j = i+1, j-1 {
			a.Elems[i], a.Elems[j] = a.Elems[j], a.Elems[i]
		}
		return void
	case "sort":
		less := func(x, y Value) bool { return compare("<", x, y) }
		if len(args) == 1 {
			fn := args[0].Fn
			less = func(x, y Value) bool { return in.call(fn, []Value{x, y}, e.Site).Bool }
		}
		work := append([]Value(nil), a.Elems...)
		mergeSort(work, make([]Value, len(work)), 0, len(work), less)
		a.Elems = work
		return void
	}
	ice("unknown array method %q", e.Method)
	return Value{}
}

// mergeSort sorts a[lo:hi] stably with a top-down merge sort, using tmp
// (as long as a) for merging. The JavaScript runtime (msort in runtime.js)
// runs the same steps, so a sort function is called with the same
// arguments in the same order on both targets. Sorting works on a copy of
// the array, which replaces its contents only when sorting finishes: an
// error thrown by the sort function leaves the array unchanged.
func mergeSort(a, tmp []Value, lo, hi int, less func(x, y Value) bool) {
	if hi-lo < 2 {
		return
	}
	mid := lo + (hi-lo)/2
	mergeSort(a, tmp, lo, mid, less)
	mergeSort(a, tmp, mid, hi, less)
	i, j, k := lo, mid, lo
	for i < mid && j < hi {
		if less(a[j], a[i]) {
			tmp[k] = a[j]
			j++
		} else {
			tmp[k] = a[i]
			i++
		}
		k++
	}
	k += copy(tmp[k:], a[i:mid])
	copy(tmp[k:], a[j:hi])
	copy(a[lo:hi], tmp[lo:hi])
}
