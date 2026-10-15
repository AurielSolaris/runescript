// Package runtime is a tree-walking interpreter for checked RuneScript
// programs.
package runtime

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/types"
)

// MaxCallDepth bounds recursion so runaway programs report an error instead
// of exhausting the Go stack. It is an implementation limit, not language
// semantics.
const MaxCallDepth = 10000

// Value is a runtime value. Integers of every v0.1 type fit in int64.
type Value struct {
	Kind types.Kind
	Int  int64
	Str  string
	Bool bool
}

func (v Value) String() string {
	switch v.Kind {
	case types.String:
		return v.Str
	case types.Bool:
		return strconv.FormatBool(v.Bool)
	}
	return strconv.FormatInt(v.Int, 10)
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
}

// Run executes the program's entry point and returns main's result as the
// process exit code (0 for a void main).
func Run(info *checker.Info, out io.Writer) (code int, err error) {
	in := &interp{info: info, out: out}
	defer func() {
		if r := recover(); r != nil {
			re, ok := r.(*Error)
			if !ok {
				panic(r)
			}
			err = re
		}
	}()
	v := in.call(info.Entry, nil, info.Entry.Name.Span)
	if info.Entry.Result.Name == "void" {
		return 0, nil
	}
	return int(v.Int), nil
}

func fail(code diagnostics.Code, span diagnostics.Span, format string, args ...any) {
	panic(&Error{diagnostics.New(code, span, format, args...)})
}

func (in *interp) call(fn *ast.FuncDecl, args []Value, site diagnostics.Span) Value {
	if in.depth >= MaxCallDepth {
		fail(diagnostics.RunCallDepth, site, "maximum call depth (%d) exceeded", MaxCallDepth)
	}
	in.depth++
	defer func() { in.depth-- }()
	f := &frame{scopes: []map[string]*Value{{}}}
	for i, p := range fn.Params {
		v := args[i]
		f.scopes[0][p.Name.Name] = &v
	}
	in.block(f, fn.Body)
	return f.ret
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
	case *ast.VarDecl:
		v := in.expr(f, s.Init)
		f.scopes[len(f.scopes)-1][s.Name.Name] = &v
	case *ast.AssignStmt:
		dst := f.lookup(s.Target.Name)
		v := in.expr(f, s.Value)
		if s.Op != "=" {
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
	}
	return ctlNone
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
	switch e := e.(type) {
	case *ast.StringLit:
		return Value{Kind: types.String, Str: e.Value}
	case *ast.BoolLit:
		return Value{Kind: types.Bool, Bool: e.Value}
	case *ast.ParenExpr:
		return in.expr(f, e.X)
	case *ast.Ident:
		return *f.lookup(e.Name)
	case *ast.CallExpr:
		return in.callExpr(f, e)
	case *ast.UnaryExpr:
		x := in.expr(f, e.X)
		if e.Op == "!" {
			return Value{Kind: types.Bool, Bool: !x.Bool}
		}
		return in.arith("-", Value{Kind: x.Kind}, x, e.Span)
	case *ast.BinaryExpr:
		return in.binary(f, e)
	}
	panic(fmt.Sprintf("runtime: unexpected expression %T", e))
}

func (in *interp) callExpr(f *frame, e *ast.CallExpr) Value {
	args := make([]Value, len(e.Args))
	for i, a := range e.Args {
		args[i] = in.expr(f, a)
	}
	if e.Fn.Name == "trace" {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = a.String()
		}
		fmt.Fprintln(in.out, strings.Join(parts, " "))
		return Value{Kind: types.Void}
	}
	return in.call(in.info.Funcs[e.Fn.Name], args, e.Span)
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
			fail(diagnostics.RunDivideByZero, span, "integer division by zero")
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
		fail(diagnostics.RunOverflow, span, "integer overflow: %d %s %d = %d does not fit in '%s'", x.Int, op, y.Int, r, t)
	}
	return Value{Kind: x.Kind, Int: r}
}
