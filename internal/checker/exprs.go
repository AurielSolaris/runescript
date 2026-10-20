package checker

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/conv"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// ---- assignment compatibility ----

var verbs = map[string]string{"assign": "assign", "pass": "pass", "return": "return"}

// assign checks that e can be stored in a destination of type dst.
// verb is "assign", "pass", or "return" and only affects wording.
func (c *checker) assign(dst *types.Type, e ast.Expr, verb string) {
	var src operand
	if call, ok := unparen(e).(*ast.CallExpr); ok && dst.Kind == types.Signal && c.isSignalBuiltin(call) {
		src = operand{typ: c.signalNew(call, dst.Elem)}
		c.info.Types[e] = src.typ
	} else if lit, ok := unparen(e).(*ast.ArrayLit); ok && dst.Kind == types.Invalid {
		c.exprs(lit.Elems) // the destination's type was already reported
		return
	} else if lit, ok := unparen(e).(*ast.ArrayLit); ok && dst.Kind == types.Array {
		src = operand{typ: c.arrayLit(lit, dst)}
		c.info.Types[e] = src.typ
	} else if lam, ok := unparen(e).(*ast.LambdaExpr); ok && dst.Kind == types.Func {
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
		if !types.IsNumeric(dst) {
			c.errorf(diagnostics.TypeMismatch, span, "Cannot %s integer constant to '%s'.", verbs[verb], dst)
			return
		}
		c.convertConst(e, src.konst, dst)
	case src.typ.Kind == types.UntypedFloat:
		if dst.Kind != types.Float {
			d := c.errorf(diagnostics.TypeMismatch, span, "Cannot %s float constant to '%s'.", verbs[verb], dst)
			if types.IsInteger(dst) {
				d.WithHelp(fmt.Sprintf("convert it explicitly, e.g. (%s) %s, which truncates toward zero", dst, conv.FormatFloat(*src.fconst, 64)))
			}
			return
		}
		c.convertFloatConst(e, *src.fconst, dst)
	case src.typ.Kind == types.Void:
		c.errorf(diagnostics.TypeVoidValue, span, "expression has type 'void' and cannot be used as a value")
	case types.IsSubtype(src.typ, dst):
	case src.typ.Kind == types.Signed && dst.Kind == types.Unsigned:
		c.errorf(diagnostics.TypeSignedToUnsigned, span, "Cannot %s signed integer '%s' to unsigned integer '%s'.", verbs[verb], src.typ, dst).
			WithHelp(fmt.Sprintf("Use an explicit checked conversion, (%s) x, or change the destination type.", dst))
	case src.typ.Kind == types.Unsigned && dst.Kind == types.Signed:
		c.errorf(diagnostics.TypeUnsignedToSigned, span, "Cannot %s unsigned integer '%s' to signed integer '%s'.", verbs[verb], src.typ, dst).
			WithHelp(fmt.Sprintf("Use an explicit checked conversion, (%s) x, or change the destination type.", dst))
	case types.IsNumeric(src.typ) && types.IsNumeric(dst):
		c.errorf(diagnostics.TypeMismatch, span, "Cannot %s '%s' to '%s'.", verbs[verb], src.typ, dst).
			WithHelp(fmt.Sprintf("numbers are not converted implicitly; write (%s) x", dst))
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

// convertConst gives an untyped integer constant the concrete numeric type
// dst, recording its value for the runtime.
func (c *checker) convertConst(e ast.Expr, v *big.Int, dst *types.Type) bool {
	span := e.NodeSpan()
	if dst.Kind == types.Float {
		f, acc := new(big.Float).SetInt(v).Float64()
		if acc != big.Exact {
			c.errorf(diagnostics.TypeInexactFloat, span, "constant %s cannot be represented exactly as '%s'", v, dst).
				WithHelp(fmt.Sprintf("write (%s) %s to round it", dst, v))
			return false
		}
		return c.convertFloatConst(e, f, dst)
	}
	if dst.Kind == types.Unsigned && v.Sign() < 0 {
		c.errorf(diagnostics.TypeNegativeToUnsigned, span, "Cannot convert negative constant %s to unsigned integer '%s'.", v, dst).
			WithHelp("An unsigned integer cannot hold a negative value; change the destination type.")
		return false
	}
	if !types.Fits(dst, v) {
		c.errorf(diagnostics.TypeConstOutOfRange, span, "constant %s does not fit in '%s'", v, dst).
			WithHelp(rangeHelp(dst))
		return false
	}
	c.info.Types[e] = dst
	c.info.Consts[e] = types.FromBig(dst, v)
	return true
}

func rangeHelp(t *types.Type) string {
	lo, hi := types.IntRange(t)
	if t.Bits == 0 {
		return fmt.Sprintf("'%s' holds %s to %s; results computed along the way may be wider, but a value of type '%s' must fit (use int64 for more)", t, lo, hi, t)
	}
	return fmt.Sprintf("'%s' holds %s to %s", t, lo, hi)
}

// convertFloatConst gives an untyped float constant (or an integer
// constant's exact float64 value) the float type dst. A float32 must hold
// it exactly: rounding needs an explicit (float32) conversion.
func (c *checker) convertFloatConst(e ast.Expr, v float64, dst *types.Type) bool {
	span := e.NodeSpan()
	if dst.Bits == 32 && !math.IsNaN(v) {
		r := conv.ToFloat32(v)
		if math.IsInf(r, 0) && !math.IsInf(v, 0) {
			c.errorf(diagnostics.TypeConstOutOfRange, span, "constant %s does not fit in 'float32'", conv.FormatFloat(v, 64)).
				WithHelp("the largest float32 is about 3.4028235e+38")
			return false
		}
		if r != v {
			c.errorf(diagnostics.TypeInexactFloat, span, "constant %s cannot be represented exactly as 'float32'", conv.FormatFloat(v, 64)).
				WithHelp(fmt.Sprintf("write (float32) %s to round it to the nearest float32, %s", conv.FormatFloat(v, 64), conv.FormatFloat(r, 32)))
			return false
		}
	}
	c.info.Types[e] = dst
	c.info.FloatConsts[e] = v
	return true
}

// convertNum gives an untyped constant operand the numeric type dst.
func (c *checker) convertNum(e ast.Expr, op operand, dst *types.Type) bool {
	if op.typ.Kind == types.UntypedFloat {
		return c.convertFloatConst(e, *op.fconst, dst)
	}
	return c.convertConst(e, op.konst, dst)
}

func isUntyped(t *types.Type) bool { return t.Kind == types.UntypedInt || t.Kind == types.UntypedFloat }

func isFloat(t *types.Type) bool { return t.Kind == types.Float || t.Kind == types.UntypedFloat }

// defaulted checks e and converts an untyped constant to its default type:
// 'int' for integers, 'float' for floats.
func (c *checker) defaulted(e ast.Expr) operand {
	return c.defaultOp(e, c.expr(e))
}

// defaultOp gives op, the checked form of e, its default type if it is an
// untyped constant.
func (c *checker) defaultOp(e ast.Expr, op operand) operand {
	switch op.typ.Kind {
	case types.UntypedInt:
		if !c.convertConst(e, op.konst, types.Int) {
			return invalid
		}
		return operand{typ: types.Int}
	case types.UntypedFloat:
		if !c.convertFloatConst(e, *op.fconst, types.FloatT) {
			return invalid
		}
		return operand{typ: types.FloatT}
	}
	return op
}

// ---- expressions ----

func (c *checker) expr(e ast.Expr) operand {
	op := c.exprInner(e)
	if !isUntyped(op.typ) {
		c.info.Types[e] = op.typ
	}
	return op
}

func (c *checker) exprInner(e ast.Expr) operand {
	switch e := e.(type) {
	case *ast.IntLit:
		return operand{typ: types.UntypedConst, konst: new(big.Int).SetUint64(e.Value)}
	case *ast.FloatLit:
		v := e.Value
		return operand{typ: types.UntypedFloatConst, fconst: &v}
	case *ast.CastExpr:
		return operand{typ: c.cast(e)}
	case *ast.NewExpr:
		return operand{typ: c.newExpr(e)}
	case *ast.SelectorExpr:
		if k, ok := c.stdConst(e); ok {
			return k
		}
		return operand{typ: c.selector(e)}
	case *ast.StringLit:
		return operand{typ: types.StringT}
	case *ast.InterpString:
		c.interpString(e)
		return operand{typ: types.StringT}
	case *ast.CharLit:
		return operand{typ: types.CharT}
	case *ast.IndexExpr:
		return operand{typ: c.index(e)}
	case *ast.ArrayLit:
		return operand{typ: c.arrayLit(e, nil)}
	case *ast.StructLit:
		return operand{typ: c.structLit(e)}
	case *ast.BoolLit:
		return operand{typ: types.BoolT}
	case *ast.ParenExpr:
		return c.expr(e.X)
	case *ast.Ident:
		return c.identOperand(e)
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

// floatConstants are the predeclared constants NaN and Infinity. A
// variable or function of the same name hides them.
var floatConstants = map[string]float64{"NaN": math.NaN(), "Infinity": math.Inf(1)}

// identOperand checks a name used as a value. NaN and Infinity are untyped
// float constants unless something else has their name.
func (c *checker) identOperand(e *ast.Ident) operand {
	if v, ok := floatConstants[e.Name]; ok {
		if sym, _ := c.lookup(e.Name); sym == nil {
			return operand{typ: types.UntypedFloatConst, fconst: &v}
		}
	}
	return operand{typ: c.ident(e)}
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
		if !c.privateError(c.file.Package, e.Name, e.Span) {
			d := c.errorf(diagnostics.NameUndefined, e.Span, "undefined name '%s'", e.Name)
			switch e.Name {
			case "null", "undefined", "nil", "NULL":
				d.WithHelp("RuneScript has no null: every value is a real value of its type. To say \"no value\", keep a bool that is false until there is one (guide, chapter 17)")
			}
		}
	case sym.fn != nil:
		c.info.FuncRefs[e] = sym.fn
		return sym.fn.Type
	case sym.st != nil:
		c.errorf(diagnostics.NameNotAValue, e.Span, "'%s' is a struct type, not a value", e.Name).
			WithHelp(fmt.Sprintf("build one with a literal, e.g. %s{...}", e.Name))
	case sym.ext != nil:
		c.externValueError(e)
	case crossed && !isRune(sym):
		c.boundaryError(e)
	default:
		sym.read = true
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
	case e.Op == "-" && x.typ.Kind == types.UntypedFloat:
		v := -*x.fconst
		return operand{typ: types.UntypedFloatConst, fconst: &v}
	case e.Op == "-" && (x.typ.Kind == types.Signed || x.typ.Kind == types.Float):
		return x
	case e.Op == "-" && x.typ.Kind == types.Unsigned:
		c.errorf(diagnostics.TypeBadOperand, e.Span, "cannot negate unsigned integer '%s'", x.typ)
		return invalid
	case e.Op == "!" && x.typ.Kind == types.Bool:
		return x
	case e.Op == "~" && x.typ.Kind == types.UntypedInt:
		return operand{typ: types.UntypedConst, konst: new(big.Int).Not(x.konst)}
	case e.Op == "~" && (x.typ.Kind == types.Signed || x.typ.Kind == types.Unsigned):
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
	case "<<", ">>":
		return c.shift(op, x, y, xe, ye, opSpan)
	case "&", "|", "^":
		if !types.IsInteger(x.typ) || !types.IsInteger(y.typ) {
			d := c.errorf(diagnostics.TypeBadOperand, opSpan, "operator '%s' is not defined on '%s' and '%s'", op, x.typ, y.typ)
			switch {
			case x.typ.Kind == types.Bool || y.typ.Kind == types.Bool:
				if op == "^" {
					d.WithHelp("bitwise operators work on integers; for bools, a ^ b is a != b")
				} else {
					d.WithHelp(fmt.Sprintf("bitwise operators work on integers; for bools use %s%s, or parenthesize, e.g. (x %s mask) == 0", op, op, op))
				}
			default:
				d.WithHelp("bitwise operators work on integers only")
			}
			return invalid
		}
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
		if x.typ.Kind == types.Struct || y.typ.Kind == types.Struct {
			switch {
			case !types.Identical(x.typ, y.typ):
				c.errorf(diagnostics.TypeBadOperand, opSpan, "operator '%s' compares '%s' with '%s'", op, x.typ, y.typ)
			case !types.Comparable(x.typ):
				c.errorf(diagnostics.TypeBadOperand, opSpan, "structs of type %s cannot be compared", x.typ).
					WithHelp("== compares every field, and arrays, functions, and errors cannot be compared; compare the fields you need")
			default:
				return operand{typ: types.BoolT}
			}
			return invalid
		}
	case "+":
		// Concatenation joins two strings; nothing is converted implicitly
		// (OQ-STR-01).
		if x.typ.Kind == types.String && y.typ.Kind == types.String {
			return operand{typ: types.StringT}
		}
	}
	if x.typ.Kind == types.Char && y.typ.Kind == types.Char {
		switch op {
		case "==", "!=", "<", "<=", ">", ">=":
			return operand{typ: types.BoolT} // by code point
		}
	}
	if isFloat(x.typ) || isFloat(y.typ) {
		if types.IsNumeric(x.typ) && types.IsNumeric(y.typ) {
			return c.floatBinary(op, x, y, xe, ye, opSpan)
		}
	}
	if !types.IsInteger(x.typ) || !types.IsInteger(y.typ) {
		help := ""
		switch {
		case op == "+" && (x.typ.Kind == types.String || y.typ.Kind == types.String):
			help = `'+' joins two strings and converts nothing; write (string) x, or use an interpolated string such as $"n={x}"`
		case x.typ.Kind == types.Func || y.typ.Kind == types.Func:
			help = "function values cannot be compared or used in arithmetic"
		case x.typ.Kind == types.Char || y.typ.Kind == types.Char:
			help = "a char is not a number; convert it, e.g. (int) c, or compare two chars"
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
	case !types.Identical(x.typ, y.typ):
		c.errorf(diagnostics.TypeBadOperand, opSpan, "operator '%s' mixes '%s' and '%s'", op, x.typ, y.typ).
			WithHelp(fmt.Sprintf("integers of different types are not converted implicitly; convert one side, e.g. (%s) x", wider(x.typ, y.typ)))
		return invalid
	}
	if comparison {
		return operand{typ: types.BoolT}
	}
	return operand{typ: t}
}

// floatBinary checks an operator with at least one float operand; both are
// numeric. Untyped constants take the other operand's float type, or
// 'float' if both are constants. Typed operands must have identical types:
// mixing an int with a float, or two float types, needs a conversion.
func (c *checker) floatBinary(op string, x, y operand, xe, ye ast.Expr, opSpan diagnostics.Span) operand {
	comparison := op == "==" || op == "!=" || op == "<" || op == "<=" || op == ">" || op == ">="
	if op == "%" {
		c.errorf(diagnostics.TypeBadOperand, opSpan, "operator '%%' is not defined on floating-point numbers").
			WithHelp("'%' is defined on integers only")
		return invalid
	}
	var t *types.Type
	switch {
	case isUntyped(x.typ) && isUntyped(y.typ):
		t = types.FloatT
		if !c.convertNum(xe, x, t) || !c.convertNum(ye, y, t) {
			return invalid
		}
	case isUntyped(x.typ) && y.typ.Kind == types.Float:
		t = y.typ
		if !c.convertNum(xe, x, t) {
			return invalid
		}
	case isUntyped(y.typ) && x.typ.Kind == types.Float:
		t = x.typ
		if !c.convertNum(ye, y, t) {
			return invalid
		}
	case types.Identical(x.typ, y.typ):
		t = x.typ
	default:
		c.errorf(diagnostics.TypeBadOperand, opSpan, "operator '%s' mixes %s and %s", op, describe(x.typ), describe(y.typ)).
			WithHelp(fmt.Sprintf("numbers are not converted implicitly; convert one side, e.g. (%s) x", floatOf(x.typ, y.typ)))
		return invalid
	}
	if comparison {
		return operand{typ: types.BoolT}
	}
	return operand{typ: t}
}

func describe(t *types.Type) string {
	switch t.Kind {
	case types.UntypedInt:
		return "an integer constant"
	case types.UntypedFloat:
		return "a float constant"
	}
	return "'" + t.Name + "'"
}

// floatOf names the float type among a and b, to suggest a conversion.
func floatOf(a, b *types.Type) string {
	for _, t := range []*types.Type{a, b} {
		if t.Kind == types.Float {
			return t.Name
		}
	}
	return "float"
}

// literalHint suggests a literal for struct type t: "x: ..., y: ...".
func literalHint(t *types.Type) string {
	parts := make([]string, len(t.Fields))
	for i, f := range t.Fields {
		parts[i] = f.Name + ": ..."
	}
	return strings.Join(parts, ", ")
}

// wider names the wider of two integer types of the same signedness, to
// suggest a conversion that cannot lose the value.
func wider(a, b *types.Type) *types.Type {
	if types.IntBits(b) > types.IntBits(a) {
		return b
	}
	return a
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
	case "&":
		return r.And(x, y)
	case "|":
		return r.Or(x, y)
	case "^":
		return r.Xor(x, y)
	}
	panic("unknown operator " + op)
}

// maxConstShift bounds the shift count of a constant expression, which is
// computed exactly; no integer type is wider than 64 bits.
const maxConstShift = 64

// shift checks x << n and x >> n (OQ-OP-02). x is an integer, and its type
// is the result's; n is any integer, and must be from 0 to x's width minus
// 1. A constant count outside that range is a compile-time error; at run
// time it throws a RangeError. << is checked like multiplication: losing
// bits is an overflow error, never a silent wrap.
func (c *checker) shift(op string, x, y operand, xe, ye ast.Expr, opSpan diagnostics.Span) operand {
	if !types.IsInteger(x.typ) || !types.IsInteger(y.typ) {
		c.errorf(diagnostics.TypeBadOperand, opSpan, "operator '%s' is not defined on '%s' and '%s'", op, x.typ, y.typ).
			WithHelp("shifts take an integer and an integer count, e.g. x << 3")
		return invalid
	}
	if y.typ.Kind == types.UntypedInt {
		switch {
		case y.konst.Sign() < 0:
			c.errorf(diagnostics.TypeConstOutOfRange, ye.NodeSpan(), "shift count %s is negative", y.konst)
			return invalid
		case x.typ.Kind == types.UntypedInt:
			if y.konst.Cmp(big.NewInt(maxConstShift)) > 0 {
				c.errorf(diagnostics.TypeConstOutOfRange, ye.NodeSpan(), "shift count %s is too large", y.konst).
					WithHelp("no integer type is wider than 64 bits")
				return invalid
			}
			n := uint(y.konst.Uint64())
			if op == "<<" {
				return operand{typ: types.UntypedConst, konst: new(big.Int).Lsh(x.konst, n)}
			}
			return operand{typ: types.UntypedConst, konst: new(big.Int).Rsh(x.konst, n)}
		case y.konst.Cmp(big.NewInt(int64(types.IntBits(x.typ)))) >= 0:
			c.errorf(diagnostics.TypeConstOutOfRange, ye.NodeSpan(), "shift count %s is out of range for '%s'", y.konst, x.typ).
				WithHelp(fmt.Sprintf("a '%s' can be shifted by 0 to %d bits", x.typ, types.IntBits(x.typ)-1))
			return invalid
		}
		if !c.convertConst(ye, y.konst, types.Int) {
			return invalid
		}
	}
	if x.typ.Kind == types.UntypedInt {
		// A constant shifted by a variable count is an int.
		if !c.convertConst(xe, x.konst, types.Int) {
			return invalid
		}
		return operand{typ: types.Int}
	}
	return operand{typ: x.typ}
}

// ---- calls ----

// call checks a call and returns its result type.
func (c *checker) call(e *ast.CallExpr) *types.Type {
	switch {
	case e.Pkg != nil:
		return c.qualifiedCall(e)
	case e.Fn == nil:
		if sel, ok := e.Callee.(*ast.SelectorExpr); ok && !c.isPackageRef(sel) {
			// x.name(args): a method of the value x
			t := c.defaulted(sel.X).typ
			if t.Kind == types.Array {
				return c.arrayMethod(e, sel.Name, t)
			}
			ft := c.selectorOf(sel, t)
			c.info.Types[sel] = ft
			return c.callValue(ft, e, fmt.Sprintf("a value of type '%s'", ft))
		}
		t := c.expr(e.Callee).typ
		return c.callValue(t, e, fmt.Sprintf("a value of type '%s'", t))
	case e.Fn.Name == "trace":
		for _, a := range e.Args {
			if t := c.defaulted(a).typ; t.Kind == types.Array {
				c.errorf(diagnostics.TypeBadOperand, a.NodeSpan(), "arrays cannot be traced").
					WithHelp(`join the elements first, e.g. trace(string.join(names, ", "))`)
			} else if t.Kind == types.Struct && !types.Traceable(t) {
				c.errorf(diagnostics.TypeBadOperand, a.NodeSpan(), "structs of type %s cannot be traced: a field is an array or a function", t).
					WithHelp("trace the fields you need")
			} else if t.Kind == types.Void {
				c.errorf(diagnostics.TypeVoidValue, a.NodeSpan(), "expression has type 'void' and cannot be traced")
			} else if t.Kind == types.Func {
				c.errorf(diagnostics.TypeBadOperand, a.NodeSpan(), "function values cannot be traced")
			} else if t.Kind == types.Opaque {
				c.errorf(diagnostics.TypeBadOperand, a.NodeSpan(), "values of type %s cannot be traced", t).
					WithHelp("turn it into text with its package's functions, e.g. trace(json.stringify(v))")
			}
		}
		return types.VoidT
	}
	name := e.Fn.Name
	sym, crossed := c.lookup(name)
	switch {
	case sym == nil && (name == "isNaN" || name == "isFinite"):
		return c.floatTest(e)
	case sym == nil && c.isSignalBuiltin(e):
		c.domUsed[c.file] = true
		return c.signalNew(e, nil)
	case sym == nil:
		if !c.privateError(c.file.Package, name, e.Fn.Span) {
			c.errorf(diagnostics.NameUndefined, e.Fn.Span, "undefined function '%s'", name)
		}
	case sym.fn != nil:
		c.info.Callees[e] = sym.fn
		return c.checkArgs(sym.fn.Type, e, "function '"+name+"'")
	case sym.st != nil:
		c.errorf(diagnostics.NameNotCallable, e.Fn.Span, "'%s' is a struct type, not a function", name).
			WithHelp(fmt.Sprintf("build a struct with a literal naming its fields, e.g. %s{%s}", name, literalHint(sym.st.Type)))
	case sym.ext != nil:
		c.info.Externs[e] = sym.ext
		return c.checkArgs(sym.ext.Type, e, "extern function '"+name+"'")
	case crossed && !isRune(sym):
		c.boundaryError(e.Fn)
	case isRune(sym):
		sym.read = true
		return c.signalCall(e, sym.typ)
	case sym.typ.Kind == types.Func:
		sym.read = true
		c.info.Types[e.Fn] = sym.typ
		return c.checkArgs(sym.typ, e, "'"+name+"'")
	case sym.typ.Kind == types.Invalid:
		sym.read = true
	default:
		sym.read = true
		c.errorf(diagnostics.NameNotCallable, e.Fn.Span, "'%s' is a variable of type '%s', not a function", name, sym.typ)
	}
	c.exprs(e.Args)
	return types.InvalidT
}

// interpString checks the holes of $"…{x}…". Each value is converted to
// text as (string) x converts it, so it must be a type that conversion
// accepts.
func (c *checker) interpString(e *ast.InterpString) {
	for _, x := range e.Exprs {
		t := c.defaulted(x).typ
		switch {
		case t.Kind == types.Invalid || convertible(t, types.StringT):
		case t.Kind == types.Void:
			c.errorf(diagnostics.TypeVoidValue, x.NodeSpan(), "expression has type 'void' and cannot be shown in a string")
		default:
			c.errorf(diagnostics.TypeBadConversion, x.NodeSpan(), "a value of type '%s' cannot be shown in a string", t).
				WithHelp("interpolation shows numbers, bools, and strings, as (string) x converts them")
		}
	}
}

// floatTest checks isNaN(x) or isFinite(x), the built-in tests of a float
// of any float type. A function of the same name hides them.
func (c *checker) floatTest(e *ast.CallExpr) *types.Type {
	if len(e.Args) != 1 {
		c.errorf(diagnostics.TypeArgCount, e.Span, "%s expects 1 argument, found %d", e.Fn.Name, len(e.Args))
		c.exprs(e.Args)
		return types.BoolT
	}
	x := c.expr(e.Args[0])
	switch {
	case isUntyped(x.typ):
		c.convertNum(e.Args[0], x, types.FloatT)
	case x.typ.Kind != types.Float && x.typ.Kind != types.Invalid:
		c.errorf(diagnostics.TypeMismatch, e.Args[0].NodeSpan(), "%s takes a float, found '%s'", e.Fn.Name, x.typ).
			WithHelp("convert it first, e.g. (float) x")
	}
	c.info.Builtins[e] = e.Fn.Name
	return types.BoolT
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

// qualifiedCall checks a call such as string.toUpper(s) or geom.area(2).
func (c *checker) qualifiedCall(e *ast.CallExpr) *types.Type {
	b := c.quals[e.Pkg.Name]
	if b != nil {
		b.used = true
	}
	switch {
	case b == nil && c.isVariable(e.Pkg.Name):
		t := c.ident(e.Pkg)
		c.info.Types[e.Pkg] = t
		if t.Kind == types.Signal {
			return c.signalMethod(e, t)
		}
		if t.Kind == types.Array {
			return c.arrayMethod(e, e.Fn, t)
		}
		c.errorf(diagnostics.TypeNoField, e.Fn.Span, "values of type '%s' have no methods", t).
			WithHelp(fieldHelp(t))
	case b == nil:
		if !c.broken {
			d := c.errorf(diagnostics.NamePackageNotUsed, e.Pkg.Span, "package '%s' is not in scope", e.Pkg.Name)
			if p := stdlib.Lookup(e.Pkg.Name); p != nil {
				d.WithHelp(fmt.Sprintf(`add use "std:%s"; after the package declaration`, p.Path))
			}
		}
	case b.rune != "":
		return c.domCall(e)
	case b.std != nil:
		return c.nativeCall(e, b.std)
	case b.pkg != nil:
		return c.packageCall(e, b.pkg)
	}
	// Unknown qualifier, or a use that could not be loaded (reported).
	c.exprs(e.Args)
	return types.InvalidT
}

// packageCall checks a call of a function of another RuneScript package.
func (c *checker) packageCall(e *ast.CallExpr, pkg *loader.Package) *types.Type {
	sym := c.pkgs[pkg].syms[e.Fn.Name]
	if sym == nil || sym.fn == nil {
		if !c.privateError(pkg, e.Fn.Name, e.Fn.Span) {
			c.errorf(diagnostics.NameNotInPackage, e.Fn.Span, "package %s has no function '%s'", pkg.Name, e.Fn.Name)
		}
		c.exprs(e.Args)
		return types.InvalidT
	}
	c.checkProtected(sym.fn, e.Fn.Span)
	c.info.Callees[e] = sym.fn
	return c.checkArgs(sym.fn.Type, e, fmt.Sprintf("function '%s.%s'", e.Pkg.Name, e.Fn.Name))
}

// stdConst returns the value of a standard library constant such as
// math.pi: an untyped constant, like a literal.
func (c *checker) stdConst(e *ast.SelectorExpr) (operand, bool) {
	id, ok := e.X.(*ast.Ident)
	if !ok || c.isVariable(id.Name) || c.quals[id.Name] == nil || c.quals[id.Name].std == nil {
		return operand{}, false
	}
	k, ok := c.quals[id.Name].std.Consts[e.Name.Name]
	if !ok {
		return operand{}, false
	}
	c.quals[id.Name].used = true
	if k.Int != nil {
		return operand{typ: types.UntypedConst, konst: new(big.Int).Set(k.Int)}, true
	}
	v := k.Float
	return operand{typ: types.UntypedFloatConst, fconst: &v}, true
}

// overloadCall checks a call of an overloaded library function (such as
// math.max, on ints or on float64s). The first signature whose parameters
// accept every argument is called: an integer constant fits an integer or
// a float parameter, a float constant only a float one, and a typed value
// only a parameter of its exact type.
func (c *checker) overloadCall(e *ast.CallExpr, fs []*stdlib.Func) *types.Type {
	ops := make([]operand, len(e.Args))
	for i, a := range e.Args {
		ops[i] = c.expr(a)
		if ops[i].typ.Kind == types.Invalid {
			c.exprs(e.Args[i+1:])
			return types.InvalidT
		}
	}
	accepts := func(p *types.Type, o operand) bool {
		switch o.typ.Kind {
		case types.UntypedInt:
			return types.IsNumeric(p)
		case types.UntypedFloat:
			return p.Kind == types.Float
		}
		return types.Identical(p, o.typ)
	}
	for _, f := range fs {
		ok := len(f.Params) == len(e.Args)
		for i := 0; ok && i < len(f.Params); i++ {
			ok = accepts(f.Params[i], ops[i])
		}
		if !ok {
			continue
		}
		for i, p := range f.Params {
			if isUntyped(ops[i].typ) {
				c.convertNum(e.Args[i], ops[i], p)
			}
		}
		c.info.Natives[e] = f
		return f.Result
	}
	sigs := make([]string, len(fs))
	for i, f := range fs {
		sigs[i] = types.NewFunc(f.Params, f.Result).String()
	}
	found := make([]string, len(ops))
	for i, o := range ops {
		found[i] = describe(o.typ)
	}
	c.errorf(diagnostics.TypeMismatch, e.Span, "no version of %s.%s takes (%s)", e.Pkg.Name, e.Fn.Name, strings.Join(found, ", ")).
		WithHelp(fmt.Sprintf("its versions are %s; convert the arguments, e.g. (float64) x", strings.Join(sigs, " and ")))
	return types.InvalidT
}

// nativeCall checks a call of a standard library function.
func (c *checker) nativeCall(e *ast.CallExpr, pkg *stdlib.Package) *types.Type {
	if fs := pkg.Overloads[e.Fn.Name]; len(fs) > 1 {
		return c.overloadCall(e, fs)
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
	return c.checkArgs(types.NewFunc(f.Params, f.Result), e, fmt.Sprintf("function '%s.%s'", e.Pkg.Name, f.Name))
}

// isVariable reports whether name is a visible variable.
func (c *checker) isVariable(name string) bool {
	sym, _ := c.lookup(name)
	return sym != nil && sym.fn == nil && sym.ext == nil
}

// ---- conversions, errors, and fields ----

// cast checks the conversion (T) x. Conversions are checked at run time
// (failures throw a RangeError or TypeError); conversions of constants are
// done, and checked, at compile time.
func (c *checker) cast(e *ast.CastExpr) *types.Type {
	to := c.resolveType(e.Type)
	x := c.expr(e.X)
	if x.typ.Kind == types.Invalid || to.Kind == types.Invalid {
		return types.InvalidT
	}
	span := e.X.NodeSpan()
	switch {
	case x.typ.Kind == types.UntypedInt && types.IsInteger(to):
		if !c.convertConst(e.X, x.konst, to) {
			return types.InvalidT
		}
		c.info.Consts[e] = c.info.Consts[e.X]
		return to
	case x.typ.Kind == types.UntypedInt && to.Kind == types.Float:
		f, _ := new(big.Float).SetInt(x.konst).Float64() // nearest
		return c.floatConstCast(e, f, to, x.konst.String())
	case x.typ.Kind == types.UntypedFloat && types.IsInteger(to):
		v := math.Trunc(*x.fconst)
		text := conv.FormatFloat(*x.fconst, 64)
		var exact *big.Int
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			exact, _ = big.NewFloat(v).Int(nil)
		}
		switch {
		case to.Kind == types.Unsigned && v < 0:
			c.errorf(diagnostics.TypeNegativeToUnsigned, span, "Cannot convert negative constant %s to unsigned integer '%s'.", text, to).
				WithHelp("An unsigned integer cannot hold a negative value; change the destination type.")
			return types.InvalidT
		case exact == nil || !types.Fits(to, exact):
			c.errorf(diagnostics.TypeConstOutOfRange, span, "constant %s does not fit in '%s'", text, to).
				WithHelp(rangeHelp(to))
			return types.InvalidT
		}
		c.info.Types[e] = to
		c.info.Consts[e] = types.FromBig(to, exact)
		return to
	case x.typ.Kind == types.UntypedFloat && to.Kind == types.Float:
		return c.floatConstCast(e, *x.fconst, to, conv.FormatFloat(*x.fconst, 64))
	}
	from := c.defaultOp(e.X, x).typ
	if from.Kind == types.Invalid {
		return types.InvalidT
	}
	if !convertible(from, to) {
		d := c.errorf(diagnostics.TypeBadConversion, e.Span, "cannot convert '%s' to '%s'", from, to)
		switch {
		case from.Kind == types.Bool && types.IsNumeric(to):
			d.WithHelp("bool is not a number; compare instead, e.g. use if (b) { n = 1; }")
		case types.IsNumeric(from) && to.Kind == types.Bool:
			d.WithHelp("numbers are not truth values; compare instead, e.g. n != 0")
		case from.Kind == types.Error:
			d.WithHelp("use the error's fields, e.g. e.message")
		}
		return types.InvalidT
	}
	return to
}

// floatConstCast converts a constant to float type to at compile time.
// Unlike an implicit conversion, an explicit one may round, and a value
// beyond float32's range becomes Infinity.
func (c *checker) floatConstCast(e *ast.CastExpr, v float64, to *types.Type, _ string) *types.Type {
	if to.Bits == 32 {
		v = conv.ToFloat32(v)
	}
	c.info.Types[e] = to
	c.info.FloatConsts[e] = v
	return to
}

// convertible reports whether (to) x is allowed for x of type from.
// Numbers convert to each other; numbers, bools, chars, and strings convert
// to string; strings convert to numbers, bool, and char; chars and
// integers convert to each other.
func convertible(from, to *types.Type) bool {
	switch {
	case from.Kind == types.Char:
		return to.Kind == types.Char || to.Kind == types.String || types.IsInteger(to)
	case to.Kind == types.Char:
		return types.IsInteger(from) || from.Kind == types.String
	case types.IsNumeric(from) && types.IsNumeric(to):
		return true
	case to.Kind == types.String:
		return types.IsNumeric(from) || from.Kind == types.Bool || from.Kind == types.String
	case from.Kind == types.String:
		return types.IsNumeric(to) || to.Kind == types.Bool
	case from.Kind == types.Bool && to.Kind == types.Bool:
		return true
	}
	return false
}

// newExpr checks `new T(message, id)`. Only the built-in error types can
// be constructed so far. As in ActionScript 3, both arguments are
// optional: the message defaults to "" and the id to 0.
func (c *checker) newExpr(e *ast.NewExpr) *types.Type {
	t := c.resolveType(e.Type)
	if t.Kind == types.Invalid {
		c.exprs(e.Args)
		return types.InvalidT
	}
	if t.Kind != types.Error {
		c.errorf(diagnostics.TypeBadNew, e.Type.Span, "'new' can only create error values so far, not '%s'", t).
			WithHelp("classes are planned; the error types are Error, RangeError, TypeError, and ArgumentError")
		c.exprs(e.Args)
		return types.InvalidT
	}
	params := []*types.Type{types.StringT, types.Int}
	if len(e.Args) > len(params) {
		c.errorf(diagnostics.TypeArgCount, e.Span, "new %s takes at most 2 arguments (message, errorID), found %d", t, len(e.Args))
		c.exprs(e.Args)
		return t
	}
	for i, a := range e.Args {
		c.assign(params[i], a, "pass")
	}
	return t
}

// errorFields are the fields of every error type.
var errorFields = map[string]*types.Type{"message": types.StringT, "name": types.StringT, "errorID": types.Int}

func fieldHelp(t *types.Type) string {
	if t.Kind == types.Error {
		return "errors have the fields message, name, and errorID"
	}
	return ""
}

// isPackageRef reports whether e is pkg.name for a package qualifier.
func (c *checker) isPackageRef(e *ast.SelectorExpr) bool {
	id, ok := e.X.(*ast.Ident)
	return ok && !c.isVariable(id.Name) && c.quals[id.Name] != nil
}

// selector checks x.name: a field of an error value or an array, or a
// package function used as a value.
func (c *checker) selector(e *ast.SelectorExpr) *types.Type {
	if c.isPackageRef(e) {
		id := e.X.(*ast.Ident)
		c.quals[id.Name].used = true
		return c.packageFuncValue(e, c.quals[id.Name])
	}
	return c.selectorOf(e, c.defaulted(e.X).typ)
}

// selectorOf checks the field name of a value of type t.
func (c *checker) selectorOf(e *ast.SelectorExpr, t *types.Type) *types.Type {
	switch {
	case t.Kind == types.Invalid:
		return types.InvalidT
	case t.Kind == types.Array:
		return c.arrayField(e, t)
	case t.Kind == types.Struct:
		return c.structField(e, t)
	case t.Kind != types.Error:
		c.errorf(diagnostics.TypeNoField, e.Name.Span, "values of type '%s' have no fields", t)
		return types.InvalidT
	}
	f := errorFields[e.Name.Name]
	if f == nil {
		c.errorf(diagnostics.TypeNoField, e.Name.Span, "'%s' has no field '%s'", t, e.Name.Name).WithHelp(fieldHelp(t))
		return types.InvalidT
	}
	return f
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
		c.lastTemplate = templateFor(e.Body.Stmts)
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
