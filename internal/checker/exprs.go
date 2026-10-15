package checker

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// ---- assignment compatibility ----

var verbs = map[string]string{"assign": "assign", "pass": "pass", "return": "return"}

// assign checks that e can be stored in a destination of type dst.
// verb is "assign", "pass", or "return" and only affects wording.
func (c *checker) assign(dst *types.Type, e ast.Expr, verb string) {
	var src operand
	if lam, ok := unparen(e).(*ast.LambdaExpr); ok && dst.Kind == types.Func {
		// A lambda takes its result type from the function type it is
		// assigned, passed, or returned to.
		src = operand{typ: c.lambda(lam, dst)}
		c.info.Types[e] = src.typ
	} else {
		src = c.expr(e)
	}
	if src.typ.Kind == types.Invalid || dst.Kind == types.Invalid || hasInvalid(src.typ) {
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

// hasInvalid reports whether a function type mentions an invalid type, so
// that one bad parameter type is not reported again as a mismatch.
func hasInvalid(t *types.Type) bool {
	if t.Kind == types.Invalid {
		return true
	}
	if t.Kind != types.Func {
		return false
	}
	for _, p := range t.Params {
		if hasInvalid(p) {
			return true
		}
	}
	return hasInvalid(t.Result)
}

func unparen(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
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
		return "'uint' is currently 32 bits (0 to 4294967295); widening is not yet implemented"
	}
	return "'int' is currently 32 bits (-2147483648 to 2147483647); widening is not yet implemented"
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
		return operand{typ: c.ident(e)}
	case *ast.CallExpr:
		return operand{typ: c.call(e)}
	case *ast.LambdaExpr:
		return operand{typ: c.lambda(e, nil)}
	case *ast.UnaryExpr:
		return c.unary(e)
	case *ast.BinaryExpr:
		x := c.expr(e.X)
		y := c.expr(e.Y)
		return c.binaryType(e.Op, x, y, e.X, e.Y, e.OpPos)
	}
	return invalid
}

// ident resolves a name used as a value: a variable, or a named function
// (which yields a function value).
func (c *checker) ident(e *ast.Ident) *types.Type {
	sym, crossed := c.lookup(e.Name)
	switch {
	case sym == nil && e.Name == "trace":
		c.errorf(diagnostics.NameNotAValue, e.Span, "the built-in 'trace' cannot be used as a value").
			WithHelp("wrap it in a lambda, e.g. (string s) => trace(s)")
	case sym == nil:
		c.errorf(diagnostics.NameUndefined, e.Span, "undefined name '%s'", e.Name)
	case sym.fn != nil:
		c.info.FuncRefs[e] = sym.fn
		return sym.fn.Type
	case crossed:
		c.boundaryError(e)
	default:
		return sym.typ
	}
	return types.InvalidT
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
		switch {
		case op == "+" && (x.typ.Kind == types.String || y.typ.Kind == types.String):
			help = "string concatenation is not yet specified; pass several arguments to trace instead"
		case x.typ.Kind == types.Func || y.typ.Kind == types.Func:
			help = "function values cannot be compared or used in arithmetic"
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

// ---- calls ----

// call checks a call and returns its result type.
func (c *checker) call(e *ast.CallExpr) *types.Type {
	switch {
	case e.Pkg != nil:
		return c.nativeCall(e)
	case e.Fn == nil:
		t := c.expr(e.Callee).typ
		return c.callValue(t, e, fmt.Sprintf("a value of type '%s'", t))
	case e.Fn.Name == "trace":
		for _, a := range e.Args {
			op := c.defaulted(a)
			switch op.typ.Kind {
			case types.Void:
				c.errorf(diagnostics.TypeVoidValue, a.NodeSpan(), "expression has type 'void' and cannot be traced")
			case types.Func:
				c.errorf(diagnostics.TypeBadOperand, a.NodeSpan(), "function values cannot be traced")
			}
		}
		return types.VoidT
	}
	name := e.Fn.Name
	sym, crossed := c.lookup(name)
	switch {
	case sym == nil:
		c.errorf(diagnostics.NameUndefined, e.Fn.Span, "undefined function '%s'", name)
	case sym.fn != nil:
		c.info.Callees[e] = sym.fn
		return c.checkArgs(sym.fn.Type, e, "function '"+name+"'")
	case crossed:
		c.boundaryError(e.Fn)
	case sym.typ.Kind == types.Func:
		c.info.Types[e.Fn] = sym.typ
		return c.checkArgs(sym.typ, e, "'"+name+"'")
	case sym.typ.Kind == types.Invalid:
	default:
		c.errorf(diagnostics.NameNotCallable, e.Fn.Span, "'%s' is a variable of type '%s', not a function", name, sym.typ)
	}
	c.exprs(e.Args)
	return types.InvalidT
}

// callValue checks a call of a computed function value, e.g. f(1)(2).
func (c *checker) callValue(t *types.Type, e *ast.CallExpr, what string) *types.Type {
	switch t.Kind {
	case types.Func:
		return c.checkArgs(t, e, what)
	case types.Invalid:
	default:
		c.errorf(diagnostics.TypeNotCallable, e.Callee.NodeSpan(), "cannot call %s", what)
	}
	c.exprs(e.Args)
	return types.InvalidT
}

func (c *checker) exprs(list []ast.Expr) {
	for _, a := range list {
		c.expr(a)
	}
}

// checkArgs checks call arguments against function type ft and returns the
// result type.
func (c *checker) checkArgs(ft *types.Type, e *ast.CallExpr, what string) *types.Type {
	if len(e.Args) != len(ft.Params) {
		c.errorf(diagnostics.TypeArgCount, e.Span, "%s expects %d argument(s), found %d", what, len(ft.Params), len(e.Args))
		c.exprs(e.Args)
		return ft.Result
	}
	for i, a := range e.Args {
		if ft.Params[i].Kind == types.Invalid {
			c.expr(a)
			continue
		}
		c.assign(ft.Params[i], a, "pass")
	}
	return ft.Result
}

// nativeCall checks a qualified call such as string.toUpper(s).
func (c *checker) nativeCall(e *ast.CallExpr) *types.Type {
	pkg := c.uses[e.Pkg.Name]
	if pkg == nil {
		d := c.errorf(diagnostics.NamePackageNotUsed, e.Pkg.Span, "package '%s' is not in scope", e.Pkg.Name)
		if p := stdlib.Lookup(e.Pkg.Name); p != nil {
			d.WithHelp(fmt.Sprintf(`add use "std:%s"; after the package declaration`, p.Path))
		}
		c.exprs(e.Args)
		return types.InvalidT
	}
	f := pkg.Funcs[e.Fn.Name]
	if f == nil {
		d := c.errorf(diagnostics.NameNotInPackage, e.Fn.Span, "package std:%s has no function '%s'", pkg.Path, e.Fn.Name)
		for name := range pkg.Funcs {
			if strings.EqualFold(name, e.Fn.Name) {
				d.WithHelp(fmt.Sprintf("did you mean '%s'? Standard library names are camelCase", name))
			}
		}
		c.exprs(e.Args)
		return types.InvalidT
	}
	c.info.Natives[e] = f
	return c.checkArgs(types.NewFunc(f.Params, f.Result), e, fmt.Sprintf("function '%s.%s'", pkg.Qualifier, f.Name))
}

// ---- lambdas ----

// lambda checks a lambda and returns its function type. When expected is a
// function type with the same parameter types, the lambda's result type is
// expected's result type; otherwise it is inferred from the body.
func (c *checker) lambda(e *ast.LambdaExpr, expected *types.Type) *types.Type {
	params := make([]*types.Type, len(e.Params))
	for i, p := range e.Params {
		params[i] = c.resolveType(p.Type)
		if params[i].Kind == types.Void {
			c.errorf(diagnostics.TypeVoidValue, p.Type.Span, "parameter '%s' cannot have type 'void'", p.Name.Name)
			params[i] = types.InvalidT
		}
	}
	var result *types.Type // nil: infer from the body
	if expected != nil && expected.Kind == types.Func && types.Identical(types.NewFunc(params, types.VoidT), types.NewFunc(expected.Params, types.VoidT)) {
		result = expected.Result
	}

	defer c.saveFuncState()()
	c.fnName = "lambda"
	c.result = result
	c.push(true)
	for i, p := range e.Params {
		c.declareVar(p.Name, params[i], true)
	}
	if e.Expr != nil {
		switch {
		case result == nil:
			result = c.defaulted(e.Expr).typ
		case result.Kind == types.Void:
			c.defaulted(e.Expr) // evaluated for its effect
		default:
			c.assign(result, e.Expr, "return")
		}
	} else {
		c.stmts(e.Body.Stmts)
		if c.result == nil {
			c.result = types.VoidT
		}
		result = c.result
		c.checkReturns(e.Body)
	}
	c.pop()
	t := types.NewFunc(params, result)
	c.info.Types[e] = t
	return t
}
