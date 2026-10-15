// Package checker performs name resolution and static type checking.
//
// Where the specification is undecided, the checker rejects the program
// rather than choosing a semantics. Rejection
// is forward-compatible: a later decision can accept more programs without
// changing the meaning of programs that are accepted today.
package checker

import (
	"math/big"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/types"
)

// Info is the result of checking, consumed by the runtime.
type Info struct {
	// Types holds the resolved type of every checked expression.
	Types map[ast.Expr]*types.Type
	// Consts holds the value of integer constant expressions that were
	// folded and converted to a concrete type. The runtime uses these
	// values instead of evaluating the expression.
	Consts map[ast.Expr]int64
	// Funcs maps function names to declarations.
	Funcs map[string]*ast.FuncDecl
	// Entry is the program entry point.
	Entry *ast.FuncDecl
}

type variable struct {
	typ     *types.Type
	isParam bool
}

type checker struct {
	info      *Info
	diags     []*diagnostics.Diagnostic
	scopes    []map[string]*variable
	fn        *ast.FuncDecl
	result    *types.Type
	loopDepth int
}

// operand is the checked form of an expression.
type operand struct {
	typ   *types.Type
	konst *big.Int // non-nil only for untyped integer constants
}

var invalid = operand{typ: types.InvalidT}

// Check type-checks file. It returns all diagnostics found.
func Check(file *ast.File) (*Info, []*diagnostics.Diagnostic) {
	c := &checker{info: &Info{
		Types:  map[ast.Expr]*types.Type{},
		Consts: map[ast.Expr]int64{},
		Funcs:  map[string]*ast.FuncDecl{},
	}}
	c.collect(file)
	for _, fn := range file.Funcs {
		c.funcBody(fn)
	}
	c.entry(file)
	return c.info, c.diags
}

func (c *checker) errorf(code diagnostics.Code, span diagnostics.Span, format string, args ...any) *diagnostics.Diagnostic {
	d := diagnostics.New(code, span, format, args...)
	c.diags = append(c.diags, d)
	return d
}

// ---- declarations ----

func (c *checker) collect(file *ast.File) {
	for _, fn := range file.Funcs {
		name := fn.Name.Name
		if name == "trace" {
			c.errorf(diagnostics.NameBuiltinReused, fn.Name.Span, "'trace' is a built-in function and cannot be redeclared")
			continue
		}
		if prev, dup := c.info.Funcs[name]; dup {
			c.errorf(diagnostics.NameRedeclared, fn.Name.Span, "function '%s' is already declared at %s", name, prev.Name.Span.Start).
				WithHelp("function overloading is not yet specified")
			continue
		}
		c.info.Funcs[name] = fn
	}
}

func (c *checker) resolveType(ref *ast.TypeRef) *types.Type {
	t := types.Lookup(ref.Name)
	if t == nil {
		c.errorf(diagnostics.TypeUnknown, ref.Span, "unknown type '%s'", ref.Name)
		return types.InvalidT
	}
	return t
}

func (c *checker) funcBody(fn *ast.FuncDecl) {
	c.fn = fn
	c.result = c.resolveType(fn.Result)
	c.scopes = []map[string]*variable{{}}
	for _, p := range fn.Params {
		t := c.resolveType(p.Type)
		if t.Kind == types.Void {
			c.errorf(diagnostics.TypeVoidValue, p.Type.Span, "parameter '%s' cannot have type 'void'", p.Name.Name)
			t = types.InvalidT
		}
		c.declare(p.Name, t, true)
	}
	c.stmts(fn.Body.Stmts)
	if c.result.Kind != types.Void && c.result.Kind != types.Invalid && !terminates(fn.Body.Stmts) {
		end := fn.Body.Span.End
		start := end
		start.Offset--
		start.Col--
		c.errorf(diagnostics.TypeMissingReturn, diagnostics.Span{Start: start, End: end},
			"function '%s' must return a value of type '%s' on every path", fn.Name.Name, c.result).
			WithHelp("add a return statement at the end of the function")
	}
}

// terminates reports whether stmts always end in a return. Loops are never
// considered terminating in v0.1.
func terminates(stmts []ast.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	switch s := stmts[len(stmts)-1].(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.IfStmt:
		return ifTerminates(s)
	}
	return false
}

func ifTerminates(s *ast.IfStmt) bool {
	if !terminates(s.Then.Stmts) {
		return false
	}
	switch e := s.Else.(type) {
	case *ast.Block:
		return terminates(e.Stmts)
	case *ast.IfStmt:
		return ifTerminates(e)
	}
	return false
}

func (c *checker) entry(file *ast.File) {
	fn, ok := c.info.Funcs["main"]
	if !ok {
		c.errorf(diagnostics.EntryMissing, file.Package.Span, "no entry point: function 'main' is not declared").
			WithHelp("declare `int main() { return 0; }`")
		return
	}
	res := types.Lookup(fn.Result.Name)
	if len(fn.Params) != 0 || res == nil || (res.Kind != types.Signed && res.Kind != types.Void) {
		c.errorf(diagnostics.EntryBadSignature, fn.Name.Span, "'main' must take no parameters and return 'int' or 'void'").
			WithHelp("entry-point parameters (e.g. command-line arguments) are not yet specified")
		return
	}
	c.info.Entry = fn
}

// ---- scopes ----

func (c *checker) push() { c.scopes = append(c.scopes, map[string]*variable{}) }
func (c *checker) pop()  { c.scopes = c.scopes[:len(c.scopes)-1] }

func (c *checker) lookup(name string) *variable {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if v, ok := c.scopes[i][name]; ok {
			return v
		}
	}
	return nil
}

// declare adds a variable to the innermost scope. Shadowing any visible
// variable or function is rejected in v0.1.
func (c *checker) declare(id *ast.Ident, t *types.Type, isParam bool) {
	switch {
	case c.lookup(id.Name) != nil:
		c.errorf(diagnostics.NameRedeclared, id.Span, "'%s' is already declared in this function", id.Name).
			WithHelp("shadowing is not allowed in v0.1; choose a different name")
	case c.info.Funcs[id.Name] != nil || id.Name == "trace":
		c.errorf(diagnostics.NameRedeclared, id.Span, "'%s' is already declared as a function", id.Name)
	}
	c.scopes[len(c.scopes)-1][id.Name] = &variable{typ: t, isParam: isParam}
}

// ---- statements ----

func (c *checker) stmts(list []ast.Stmt) {
	for _, s := range list {
		c.stmt(s)
	}
}

func (c *checker) block(b *ast.Block) {
	c.push()
	c.stmts(b.Stmts)
	c.pop()
}

func (c *checker) stmt(s ast.Stmt) {
	switch s := s.(type) {
	case *ast.VarDecl:
		t := c.resolveType(s.Type)
		if t.Kind == types.Void {
			c.errorf(diagnostics.TypeVoidValue, s.Type.Span, "variable '%s' cannot have type 'void'", s.Name.Name)
			t = types.InvalidT
		}
		c.assign(t, s.Init, "assign")
		c.declare(s.Name, t, false)
	case *ast.AssignStmt:
		c.assignStmt(s)
	case *ast.IncDecStmt:
		if v := c.target(s.Target); v != nil && v.typ.Kind != types.Invalid && !types.IsInteger(v.typ) {
			c.errorf(diagnostics.TypeBadOperand, s.Span, "operator '%s' requires an integer variable, but '%s' has type '%s'", s.Op, s.Target.Name, v.typ)
		}
	case *ast.ExprStmt:
		c.call(s.Call)
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
		c.push()
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
	case *ast.BranchStmt:
		if c.loopDepth == 0 {
			c.errorf(diagnostics.FlowBranchOutsideLoop, s.Span, "'%s' is only allowed inside a loop", s.Keyword)
		}
	}
}

func (c *checker) loopBody(b *ast.Block) {
	c.loopDepth++
	c.block(b)
	c.loopDepth--
}

// target resolves an assignment target; parameters are read-only.
func (c *checker) target(id *ast.Ident) *variable {
	v := c.lookup(id.Name)
	if v == nil {
		c.errorf(diagnostics.NameUndefined, id.Span, "undefined variable '%s'", id.Name)
		return nil
	}
	if v.isParam {
		c.errorf(diagnostics.NameAssignToParam, id.Span, "cannot assign to parameter '%s'", id.Name).
			WithHelp("parameters are read-only in v0.1; copy the value into a local variable")
		return nil
	}
	return v
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
	op := s.Op[:1]
	y := c.expr(s.Value)
	t := c.binaryType(op, operand{typ: v.typ}, y, s.Target, s.Value, s.Span)
	if t.typ.Kind != types.Invalid && !types.Identical(t.typ, v.typ) {
		c.errorf(diagnostics.TypeMismatch, s.Span, "operator '%s' produces '%s', which cannot be stored in '%s'", s.Op, t.typ, v.typ)
	}
}

func (c *checker) returnStmt(s *ast.ReturnStmt) {
	switch {
	case c.result.Kind == types.Void && s.Value != nil:
		c.expr(s.Value)
		c.errorf(diagnostics.TypeBadReturn, s.Value.NodeSpan(), "function '%s' returns 'void' and cannot return a value", c.fn.Name.Name)
	case c.result.Kind != types.Void && s.Value == nil:
		if c.result.Kind != types.Invalid {
			c.errorf(diagnostics.TypeBadReturn, s.Span, "function '%s' must return a value of type '%s'", c.fn.Name.Name, c.result)
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

// ---- assignment compatibility ----

var verbs = map[string]string{"assign": "assign", "pass": "pass", "return": "return"}

// assign checks that e can be stored in a destination of type dst.
// verb is "assign", "pass", or "return" and only affects wording.
func (c *checker) assign(dst *types.Type, e ast.Expr, verb string) {
	src := c.expr(e)
	if src.typ.Kind == types.Invalid || dst.Kind == types.Invalid {
		return
	}
	span := e.NodeSpan()
	switch {
	case src.typ.Kind == types.UntypedInt:
		if !types.IsInteger(dst) {
			c.errorf(diagnostics.TypeMismatch, span, "Cannot %s integer constant to '%s'.", verbs[verb], dst)
			return
		}
		c.convertConst(e, src.konst, dst)
	case src.typ.Kind == types.Void:
		c.errorf(diagnostics.TypeVoidValue, span, "expression has type 'void' and cannot be used as a value")
	case types.Identical(src.typ, dst):
	case src.typ.Kind == types.Signed && dst.Kind == types.Unsigned:
		c.errorf(diagnostics.TypeSignedToUnsigned, span, "Cannot %s signed integer '%s' to unsigned integer '%s'.", verbs[verb], src.typ, dst).
			WithHelp("Use an explicit checked conversion (syntax not yet specified), or change the destination type.")
	case src.typ.Kind == types.Unsigned && dst.Kind == types.Signed:
		c.errorf(diagnostics.TypeUnsignedToSigned, span, "Cannot %s unsigned integer '%s' to signed integer '%s'.", verbs[verb], src.typ, dst).
			WithHelp("Unsigned-to-signed conversion rules are not yet specified; change the destination type.")
	default:
		c.errorf(diagnostics.TypeMismatch, span, "Cannot %s '%s' to '%s'.", verbs[verb], src.typ, dst)
	}
}

// convertConst gives an untyped constant the concrete integer type dst,
// recording its value for the runtime.
func (c *checker) convertConst(e ast.Expr, v *big.Int, dst *types.Type) bool {
	span := e.NodeSpan()
	if dst.Kind == types.Unsigned && v.Sign() < 0 {
		c.errorf(diagnostics.TypeNegativeToUnsigned, span, "Cannot convert negative constant %s to unsigned integer '%s'.", v, dst).
			WithHelp("An unsigned integer cannot hold a negative value; change the destination type.")
		return false
	}
	if !v.IsInt64() || !types.InRange(dst, v.Int64()) {
		c.errorf(diagnostics.TypeConstOutOfRange, span, "constant %s does not fit in '%s'", v, dst).
			WithHelp(rangeHelp(dst))
		return false
	}
	c.info.Types[e] = dst
	c.info.Consts[e] = v.Int64()
	return true
}

func rangeHelp(t *types.Type) string {
	if t.Kind == types.Unsigned {
		return "v0.1 represents 'uint' as 32 bits (0 to 4294967295); widening is not yet implemented"
	}
	return "v0.1 represents 'int' as 32 bits (-2147483648 to 2147483647); widening is not yet implemented"
}

// defaulted checks e and converts an untyped constant to 'int'.
func (c *checker) defaulted(e ast.Expr) operand {
	op := c.expr(e)
	if op.typ.Kind == types.UntypedInt {
		if !c.convertConst(e, op.konst, types.Int) {
			return invalid
		}
		return operand{typ: types.Int}
	}
	return op
}

// ---- expressions ----

func (c *checker) expr(e ast.Expr) operand {
	op := c.exprInner(e)
	if op.typ.Kind != types.UntypedInt {
		c.info.Types[e] = op.typ
	}
	return op
}

func (c *checker) exprInner(e ast.Expr) operand {
	switch e := e.(type) {
	case *ast.IntLit:
		return operand{typ: types.UntypedConst, konst: new(big.Int).SetUint64(e.Value)}
	case *ast.StringLit:
		return operand{typ: types.StringT}
	case *ast.BoolLit:
		return operand{typ: types.BoolT}
	case *ast.ParenExpr:
		return c.expr(e.X)
	case *ast.Ident:
		if v := c.lookup(e.Name); v != nil {
			return operand{typ: v.typ}
		}
		if c.info.Funcs[e.Name] != nil || e.Name == "trace" {
			c.errorf(diagnostics.NameNotAValue, e.Span, "function '%s' cannot be used as a value", e.Name).
				WithHelp("did you mean to call it, e.g. " + e.Name + "()?")
			return invalid
		}
		c.errorf(diagnostics.NameUndefined, e.Span, "undefined name '%s'", e.Name)
		return invalid
	case *ast.CallExpr:
		return operand{typ: c.call(e)}
	case *ast.UnaryExpr:
		return c.unary(e)
	case *ast.BinaryExpr:
		x := c.expr(e.X)
		y := c.expr(e.Y)
		return c.binaryType(e.Op, x, y, e.X, e.Y, e.OpPos)
	}
	return invalid
}

func (c *checker) unary(e *ast.UnaryExpr) operand {
	x := c.expr(e.X)
	switch {
	case x.typ.Kind == types.Invalid:
		return invalid
	case e.Op == "-" && x.typ.Kind == types.UntypedInt:
		return operand{typ: types.UntypedConst, konst: new(big.Int).Neg(x.konst)}
	case e.Op == "-" && x.typ.Kind == types.Signed:
		return x
	case e.Op == "-" && x.typ.Kind == types.Unsigned:
		c.errorf(diagnostics.TypeBadOperand, e.Span, "cannot negate unsigned integer '%s'", x.typ)
		return invalid
	case e.Op == "!" && x.typ.Kind == types.Bool:
		return x
	}
	c.errorf(diagnostics.TypeBadOperand, e.Span, "operator '%s' is not defined on '%s'", e.Op, x.typ)
	return invalid
}

// binaryType checks `x op y`. xe and ye are the operand expressions (used to
// record constant conversions); opSpan is where errors are reported.
func (c *checker) binaryType(op string, x, y operand, xe, ye ast.Expr, opSpan diagnostics.Span) operand {
	if x.typ.Kind == types.Invalid || y.typ.Kind == types.Invalid {
		return invalid
	}
	switch op {
	case "&&", "||":
		if x.typ.Kind != types.Bool || y.typ.Kind != types.Bool {
			c.errorf(diagnostics.TypeBadOperand, opSpan, "operator '%s' requires 'bool' operands, found '%s' and '%s'", op, x.typ, y.typ)
			return invalid
		}
		return operand{typ: types.BoolT}
	case "==", "!=":
		if x.typ.Kind == y.typ.Kind && (x.typ.Kind == types.Bool || x.typ.Kind == types.String) {
			return operand{typ: types.BoolT}
		}
	}
	if !types.IsInteger(x.typ) || !types.IsInteger(y.typ) {
		help := ""
		if op == "+" && (x.typ.Kind == types.String || y.typ.Kind == types.String) {
			help = "string concatenation is not yet specified; pass several arguments to trace instead"
		}
		c.errorf(diagnostics.TypeBadOperand, opSpan, "operator '%s' is not defined on '%s' and '%s'", op, x.typ, y.typ).WithHelp(help)
		return invalid
	}
	comparison := op == "==" || op == "!=" || op == "<" || op == "<=" || op == ">" || op == ">="

	// Two constants: fold arithmetic exactly; compare as 'int'.
	if x.typ.Kind == types.UntypedInt && y.typ.Kind == types.UntypedInt {
		if comparison {
			if !c.convertConst(xe, x.konst, types.Int) || !c.convertConst(ye, y.konst, types.Int) {
				return invalid
			}
			return operand{typ: types.BoolT}
		}
		if (op == "/" || op == "%") && y.konst.Sign() == 0 {
			c.errorf(diagnostics.TypeConstDivByZero, opSpan, "constant division by zero")
			return invalid
		}
		return operand{typ: types.UntypedConst, konst: fold(op, x.konst, y.konst)}
	}

	// One constant: it takes the other operand's type.
	t := x.typ
	switch {
	case x.typ.Kind == types.UntypedInt:
		t = y.typ
		if !c.convertConst(xe, x.konst, t) {
			return invalid
		}
	case y.typ.Kind == types.UntypedInt:
		if !c.convertConst(ye, y.konst, t) {
			return invalid
		}
	case x.typ.Kind != y.typ.Kind:
		c.errorf(diagnostics.TypeMixedSignedness, opSpan, "operator '%s' mixes signed '%s' and unsigned '%s'", op, signedOf(x.typ, y.typ), unsignedOf(x.typ, y.typ)).
			WithHelp("mixed-signedness arithmetic and comparison are not yet specified")
		return invalid
	}
	if comparison {
		return operand{typ: types.BoolT}
	}
	return operand{typ: t}
}

func signedOf(a, b *types.Type) *types.Type {
	if a.Kind == types.Signed {
		return a
	}
	return b
}

func unsignedOf(a, b *types.Type) *types.Type {
	if a.Kind == types.Unsigned {
		return a
	}
	return b
}

// fold evaluates a constant arithmetic operation exactly. Division
// truncates toward zero, matching the runtime.
func fold(op string, x, y *big.Int) *big.Int {
	r := new(big.Int)
	switch op {
	case "+":
		return r.Add(x, y)
	case "-":
		return r.Sub(x, y)
	case "*":
		return r.Mul(x, y)
	case "/":
		return r.Quo(x, y)
	case "%":
		return r.Rem(x, y)
	}
	panic("unknown operator " + op)
}

// call checks a call and returns its result type.
func (c *checker) call(e *ast.CallExpr) *types.Type {
	name := e.Fn.Name
	if c.lookup(name) != nil {
		c.errorf(diagnostics.NameNotCallable, e.Fn.Span, "'%s' is a variable, not a function", name)
		return types.InvalidT
	}
	if name == "trace" {
		for _, a := range e.Args {
			op := c.defaulted(a)
			if op.typ.Kind == types.Void {
				c.errorf(diagnostics.TypeVoidValue, a.NodeSpan(), "expression has type 'void' and cannot be traced")
			}
		}
		return types.VoidT
	}
	fn := c.info.Funcs[name]
	if fn == nil {
		c.errorf(diagnostics.NameUndefined, e.Fn.Span, "undefined function '%s'", name)
		for _, a := range e.Args {
			c.expr(a)
		}
		return types.InvalidT
	}
	if len(e.Args) != len(fn.Params) {
		c.errorf(diagnostics.TypeArgCount, e.Span, "function '%s' expects %d argument(s), found %d", name, len(fn.Params), len(e.Args))
		for _, a := range e.Args {
			c.expr(a)
		}
	} else {
		for i, a := range e.Args {
			pt := types.Lookup(fn.Params[i].Type.Name)
			if pt == nil || pt.Kind == types.Void {
				c.expr(a)
				continue
			}
			c.assign(pt, a, "pass")
		}
	}
	if rt := types.Lookup(fn.Result.Name); rt != nil {
		return rt
	}
	return types.InvalidT
}
