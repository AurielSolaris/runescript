// Package runtime is a tree-walking interpreter for checked RuneScript
// programs.
package runtime

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/conv"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// MaxCallDepth bounds recursion so runaway programs report an error instead
// of exhausting the Go stack. It is an implementation limit, not language
// semantics.
const MaxCallDepth = 10000

// Value is a runtime value. Integers of every current type fit in int64.
// A float is held as a float64; Bits is 32 for float32 values, which are
// always exact float32 values, and 0 otherwise (see types.FloatT).
type Value struct {
	Kind types.Kind
	Int  int64
	F    float64
	Bits int
	Str  string
	Bool bool
	Err  *ErrorValue
	// Fn is the function of a function value: *ast.FuncDecl or
	// *ast.LambdaExpr. Functions do not capture variables, so no
	// environment is needed.
	Fn ast.Node
}

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
	}
	return strconv.FormatInt(v.Int, 10)
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

type frame struct {
	scopes []map[string]*Value
	ret    Value
}

type interp struct {
	info  *checker.Info
	out   io.Writer
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
}

// Run executes the program's entry point and returns main's result as the
// process exit code (0 for a void main).
func Run(info *checker.Info, out io.Writer) (code int, err error) {
	if len(info.ExternDecls) > 0 {
		// The driver reports this as ASX-TARGET-002 before calling Run.
		return 0, fmt.Errorf("the program declares extern function '%s'; only the JavaScript target can run it", info.ExternDecls[0].Name)
	}
	in := &interp{info: info, out: out}
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
	v := in.call(info.Entry, nil, info.Entry.Name.Span)
	if info.EntryVoid {
		return 0, nil
	}
	return int(v.Int), nil
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

// call invokes a named function definition.
func (in *interp) call(fn *ast.FuncDecl, args []Value, site diagnostics.Span) Value {
	f := in.enter(fn.Params, args, site)
	defer func() { in.depth-- }()
	in.block(f, fn.Body)
	return f.ret
}

// callValue invokes a function value.
func (in *interp) callValue(v Value, args []Value, site diagnostics.Span) Value {
	switch fn := v.Fn.(type) {
	case *ast.FuncDecl:
		return in.call(fn, args, site)
	case *ast.LambdaExpr:
		f := in.enter(fn.Params, args, site)
		defer func() { in.depth-- }()
		if fn.Expr != nil {
			return in.expr(f, fn.Expr)
		}
		in.block(f, fn.Body)
		return f.ret
	}
	panic(fmt.Sprintf("runtime: calling non-function value %v", v)) // prevented by the checker
}

// enter starts a new call frame with the parameters bound. Each function
// has its own frame: there is no access to the caller's variables.
func (in *interp) enter(params []*ast.Param, args []Value, site diagnostics.Span) *frame {
	if in.depth >= MaxCallDepth {
		in.fail(diagnostics.RunCallDepth, site, "maximum call depth (%d) exceeded", MaxCallDepth)
	}
	in.depth++
	f := &frame{scopes: []map[string]*Value{{}}}
	for i, p := range params {
		v := args[i]
		f.scopes[0][p.Name.Name] = &v
	}
	return f
}

// ---- statements ----

func (in *interp) block(f *frame, b *ast.Block) control {
	f.scopes = append(f.scopes, map[string]*Value{})
	defer func() { f.scopes = f.scopes[:len(f.scopes)-1] }()
	for _, s := range b.Stmts {
		if c := in.stmt(f, s); c != ctlNone {
			return c
		}
	}
	return ctlNone
}

func (f *frame) lookup(name string) *Value {
	for i := len(f.scopes) - 1; i >= 0; i-- {
		if v, ok := f.scopes[i][name]; ok {
			return v
		}
	}
	panic("runtime: unresolved variable " + name) // prevented by the checker
}

func (in *interp) stmt(f *frame, s ast.Stmt) control {
	switch s := s.(type) {
	case *ast.FuncDecl:
		// Nested functions are resolved statically by the checker.
	case *ast.VarDecl:
		v := in.expr(f, s.Init)
		f.scopes[len(f.scopes)-1][s.Name.Name] = &v
	case *ast.AssignStmt:
		dst := f.lookup(s.Target.Name)
		v := in.expr(f, s.Value)
		switch {
		case s.Op == "=":
		case dst.Kind == types.Float:
			v = in.floatArith(s.Op[:1], *dst, v, floatType(dst.Bits))
		default:
			v = in.arith(s.Op[:1], *dst, v, s.Span)
		}
		*dst = v
	case *ast.IncDecStmt:
		dst := f.lookup(s.Target.Name)
		*dst = in.arith(s.Op[:1], *dst, Value{Kind: dst.Kind, Int: 1}, s.Span)
	case *ast.ExprStmt:
		in.expr(f, s.Call)
	case *ast.ReturnStmt:
		if s.Value != nil {
			f.ret = in.expr(f, s.Value)
		}
		return ctlReturn
	case *ast.IfStmt:
		if in.expr(f, s.Cond).Bool {
			return in.block(f, s.Then)
		}
		switch e := s.Else.(type) {
		case *ast.Block:
			return in.block(f, e)
		case *ast.IfStmt:
			return in.stmt(f, e)
		}
	case *ast.WhileStmt:
		for in.expr(f, s.Cond).Bool {
			if c := in.block(f, s.Body); c == ctlBreak {
				break
			} else if c == ctlReturn {
				return c
			}
		}
	case *ast.ForStmt:
		return in.forStmt(f, s)
	case *ast.BranchStmt:
		if s.Keyword == "break" {
			return ctlBreak
		}
		return ctlContinue
	case *ast.ThrowStmt:
		err := in.expr(f, s.Value).Err
		if !err.Set {
			err.Site, err.Set = s.Span, true
		}
		in.throw(err)
	case *ast.TryStmt:
		return in.tryStmt(f, s)
	}
	return ctlNone
}

// tryStmt runs try/catch/finally. The finally block runs however the try
// statement ends (normally, by return, break, or continue, or by an error)
// except when the program is stopping for an error that cannot be caught.
// A finally block that returns, breaks, or continues replaces that ending,
// discarding an error in flight, as in ActionScript 3 and Java.
func (in *interp) tryStmt(f *frame, s *ast.TryStmt) (c control) {
	if s.Finally != nil {
		depth, scopes := in.depth, len(f.scopes)
		defer func() {
			if in.fatal {
				return
			}
			// Any error in flight goes on after the block, unless the
			// block throws one of its own, which replaces it, or returns,
			// breaks, or continues, which discards it.
			raising := in.raising
			in.depth, f.scopes = depth, f.scopes[:scopes]
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

func (in *interp) tryCatch(f *frame, s *ast.TryStmt) control {
	if len(s.Catches) == 0 {
		return in.block(f, s.Body)
	}
	depth, scopes := in.depth, len(f.scopes)
	var clause *ast.CatchClause
	c, err := func() (c control, err *ErrorValue) {
		defer func() {
			e := in.raising
			if in.fatal || e == nil {
				return
			}
			for _, cl := range s.Catches {
				if types.IsSubtype(e.Type, in.info.CatchTypes[cl]) {
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
	in.depth, f.scopes = depth, f.scopes[:scopes]
	f.scopes = append(f.scopes, map[string]*Value{clause.Name.Name: {Kind: types.Error, Err: err}})
	defer func() { f.scopes = f.scopes[:scopes] }()
	return in.block(f, clause.Body)
}

func (in *interp) forStmt(f *frame, s *ast.ForStmt) control {
	f.scopes = append(f.scopes, map[string]*Value{})
	defer func() { f.scopes = f.scopes[:len(f.scopes)-1] }()
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

// ---- expressions ----

func (in *interp) expr(f *frame, e ast.Expr) Value {
	if c, ok := in.info.Consts[e]; ok {
		return Value{Kind: in.info.Types[e].Kind, Int: c}
	}
	if c, ok := in.info.FloatConsts[e]; ok {
		return Value{Kind: types.Float, F: c, Bits: in.info.Types[e].Bits}
	}
	switch e := e.(type) {
	case *ast.CastExpr:
		return in.convert(in.expr(f, e.X), in.info.Types[e.X], in.info.Types[e], e.Span)
	case *ast.NewExpr:
		err := &ErrorValue{Type: in.info.Types[e], Code: diagnostics.RunUncaught}
		if len(e.Args) > 0 {
			err.Message = in.expr(f, e.Args[0]).Str
		}
		if len(e.Args) > 1 {
			err.ID = in.expr(f, e.Args[1]).Int
		}
		return Value{Kind: types.Error, Err: err}
	case *ast.SelectorExpr:
		if fn, ok := in.info.QualRefs[e]; ok { // pkg.name used as a value
			return Value{Kind: types.Func, Fn: fn.Def}
		}
		err := in.expr(f, e.X).Err
		switch e.Name.Name {
		case "message":
			return Value{Kind: types.String, Str: err.Message}
		case "name":
			return Value{Kind: types.String, Str: err.Type.Name}
		}
		return Value{Kind: types.Signed, Int: err.ID}
	case *ast.StringLit:
		return Value{Kind: types.String, Str: e.Value}
	case *ast.BoolLit:
		return Value{Kind: types.Bool, Bool: e.Value}
	case *ast.ParenExpr:
		return in.expr(f, e.X)
	case *ast.Ident:
		if fn, ok := in.info.FuncRefs[e]; ok {
			return Value{Kind: types.Func, Fn: fn.Def}
		}
		return *f.lookup(e.Name)
	case *ast.LambdaExpr:
		return Value{Kind: types.Func, Fn: e}
	case *ast.CallExpr:
		return in.callExpr(f, e)
	case *ast.UnaryExpr:
		x := in.expr(f, e.X)
		if e.Op == "!" {
			return Value{Kind: types.Bool, Bool: !x.Bool}
		}
		if x.Kind == types.Float {
			x.F = -x.F
			return x
		}
		return in.arith("-", Value{Kind: x.Kind}, x, e.Span)
	case *ast.BinaryExpr:
		return in.binary(f, e)
	}
	panic(fmt.Sprintf("runtime: unexpected expression %T", e))
}

// callExpr evaluates the callee first, then the arguments left to right.
func (in *interp) callExpr(f *frame, e *ast.CallExpr) Value {
	var callee Value
	if e.Callee != nil {
		callee = in.expr(f, e.Callee)
	}
	args := make([]Value, len(e.Args))
	for i, a := range e.Args {
		args[i] = in.expr(f, a)
	}
	if nf, ok := in.info.Natives[e]; ok {
		return in.native(nf, args, e)
	}
	if e.Callee != nil {
		return in.callValue(callee, args, e.Span)
	}
	if fn, ok := in.info.Callees[e]; ok {
		return in.call(fn.Def, args, e.Span)
	}
	switch in.info.Builtins[e] {
	case "isNaN":
		return boolean(math.IsNaN(args[0].F))
	case "isFinite":
		return boolean(!math.IsNaN(args[0].F) && !math.IsInf(args[0].F, 0))
	}
	if e.Fn.Name != "trace" { // a variable holding a function value
		return in.callValue(*f.lookup(e.Fn.Name), args, e.Span)
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = a.String()
	}
	fmt.Fprintln(in.out, strings.Join(parts, " "))
	return Value{Kind: types.Void}
}

func (in *interp) binary(f *frame, e *ast.BinaryExpr) Value {
	x := in.expr(f, e.X)
	switch e.Op { // short-circuit
	case "&&":
		if !x.Bool {
			return x
		}
		return in.expr(f, e.Y)
	case "||":
		if x.Bool {
			return x
		}
		return in.expr(f, e.Y)
	}
	y := in.expr(f, e.Y)
	switch e.Op {
	case "==":
		return boolean(x == y)
	case "!=":
		return boolean(x != y)
	}
	if x.Kind == types.Float {
		switch e.Op {
		case "<":
			return boolean(x.F < y.F)
		case "<=":
			return boolean(x.F <= y.F)
		case ">":
			return boolean(x.F > y.F)
		case ">=":
			return boolean(x.F >= y.F)
		}
		return in.floatArith(e.Op, x, y, in.info.Types[e])
	}
	switch e.Op {
	case "<":
		return boolean(x.Int < y.Int)
	case "<=":
		return boolean(x.Int <= y.Int)
	case ">":
		return boolean(x.Int > y.Int)
	case ">=":
		return boolean(x.Int >= y.Int)
	}
	return in.arith(e.Op, x, y, e.OpPos)
}

func boolean(b bool) Value { return Value{Kind: types.Bool, Bool: b} }

// arith performs integer arithmetic. Operands are within their type's range,
// so int64 arithmetic is exact; a result outside the range is a runtime
// error because widening is not yet specified (OQ-NUM-01).
func (in *interp) arith(op string, x, y Value, span diagnostics.Span) Value {
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
	t := types.Int
	if x.Kind == types.Unsigned {
		t = types.Uint
	}
	if !types.InRange(t, r) {
		in.fail(diagnostics.RunOverflow, span, "integer overflow: %d %s %d = %d does not fit in '%s'", x.Int, op, y.Int, r, t)
	}
	return Value{Kind: x.Kind, Int: r}
}

// floatArith performs + - * / on floats of type t, with IEEE 754
// semantics as in ActionScript 3: overflow gives Infinity, and 0/0 gives
// NaN. float32 results are rounded to float32.
func (in *interp) floatArith(op string, x, y Value, t *types.Type) Value {
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

// floatType is the float type of a value with the given Bits.
func floatType(bits int) *types.Type {
	switch bits {
	case 32:
		return types.Float32
	case 64:
		return types.Float64
	}
	return types.FloatT
}

// convert performs the checked conversion (to) v, where v has type from.
// A value outside the destination's range throws a RangeError, and a
// string that is not a number (or bool) throws a TypeError.
func (in *interp) convert(v Value, from, to *types.Type, span diagnostics.Span) Value {
	outOfRange := func() {
		in.raise(types.RangeErrorT, diagnostics.RunConversion, span, fmt.Sprintf("cannot convert %s to '%s': out of range", quoted(v), to))
	}
	notValid := func(what string) {
		in.raise(types.TypeErrorT, diagnostics.RunConversion, span, fmt.Sprintf("cannot convert %s to '%s': not %s", quoted(v), to, what))
	}
	switch to.Kind {
	case types.String:
		return Value{Kind: types.String, Str: v.String()}
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
		var n int64
		switch v.Kind {
		case types.Signed, types.Unsigned:
			n = v.Int
		case types.Float:
			t := math.Trunc(v.F)
			if math.IsNaN(t) || t < math.MinInt64 || t >= math.MaxInt64 {
				outOfRange()
			}
			n = int64(t)
		case types.String:
			parsed, fits, ok := conv.ParseInt(v.Str)
			if !ok {
				notValid("an integer")
			}
			if !fits {
				outOfRange()
			}
			n = parsed
		}
		if !types.InRange(to, n) {
			outOfRange()
		}
		return Value{Kind: to.Kind, Int: n}
	case types.Float:
		var f float64
		switch v.Kind {
		case types.Signed, types.Unsigned:
			f = float64(v.Int)
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
	panic(fmt.Sprintf("runtime: conversion from %s to %s", from, to)) // prevented by the checker
}

// quoted renders a conversion's operand for an error message: strings in
// double quotes, as written, and other values as trace prints them.
func quoted(v Value) string {
	if v.Kind == types.String {
		return `"` + v.Str + `"`
	}
	return v.String()
}

// native calls a standard library function. Go panics (for example a
// negative count passed to string.repeat) become runtime errors, and integer
// results are range-checked against the declared result type.
func (in *interp) native(f *stdlib.Func, args []Value, e *ast.CallExpr) (result Value) {
	goArgs := make([]any, len(args))
	for i, a := range args {
		switch a.Kind {
		case types.String:
			goArgs[i] = a.Str
		case types.Bool:
			goArgs[i] = a.Bool
		default:
			goArgs[i] = a.Int
		}
	}
	name := e.Pkg.Name + "." + f.Name
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case *Error, thrown:
				panic(r)
			}
			in.fail(diagnostics.RunNativeFailed, e.Span, "%s failed: %v", name, r)
		}
	}()
	out := f.Impl(goArgs)
	switch f.Result.Kind {
	case types.Void:
		return Value{Kind: types.Void}
	case types.String:
		return Value{Kind: types.String, Str: out.(string)}
	case types.Bool:
		return Value{Kind: types.Bool, Bool: out.(bool)}
	}
	n := out.(int64)
	if !types.InRange(f.Result, n) {
		in.fail(diagnostics.RunOverflow, e.Span, "%s returned %d, which does not fit in '%s'", name, n, f.Result)
	}
	return Value{Kind: f.Result.Kind, Int: n}
}
