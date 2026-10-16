// Package ir defines RuneScript's typed intermediate representation.
//
// The IR is the boundary between semantic analysis and the back ends: it
// is produced from a checked AST (see Lower) and consumed by both the
// interpreter (internal/runtime) and the JavaScript generator
// (internal/jsgen). Every expression carries its concrete
// type: there are no untyped constants, no invalid types, and no names left
// to resolve. Integer arithmetic is a distinct node (Arith) with the
// source location a runtime error is reported at, so a back end cannot emit
// an integer operation without also emitting its range check.
//
// Verify re-checks the typing rules on the IR itself, independently of the
// checker, before either back end consumes it, so a checker or lowering
// bug is reported as an internal compiler error instead of running or
// generating questionable code.
package ir

import (
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/jshost"
	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// Span is re-exported for brevity.
type Span = diagnostics.Span

// Program is a whole checked program.
type Program struct {
	// Funcs are the top-level function definitions of every file, in load
	// order, then source order.
	Funcs []*Func
	// Entry is main; EntryVoid reports whether it returns void.
	Entry     *Func
	EntryVoid bool
	// EntrySite is where a runtime error raised on entering main is
	// reported (main's name, as in the interpreter).
	EntrySite Span
	// Externs are the extern functions called anywhere in the program, in
	// order of first call.
	Externs []*Extern
}

// Extern is a function implemented in JavaScript by Module. Its argument
// and result types are numbers, bool, string, errors, or (result only)
// void; results are checked when they cross into RuneScript.
type Extern struct {
	Name   string
	Module string
	Type   *types.Type
	// Host is the host-neutral function of a rune: module, or nil.
	Host *jshost.Func
}

// Func is a named function (top-level or nested) or a lambda.
type Func struct {
	Name   string // "" for a lambda
	Pkg    string // the package of a top-level function; "" otherwise
	Type   *types.Type
	Params []*Var
	Body   *Block
	Span   Span
}

// Var is a parameter or local variable. Each declaration has its own Var,
// so two variables with the same name in different functions are distinct.
type Var struct {
	Name  string
	Type  *types.Type
	Param bool
}

// Block is a sequence of statements that forms a scope.
type Block struct {
	Stmts []Stmt
}

// Stmt is any statement.
type Stmt interface{ stmt() }

type (
	// FuncStmt defines a nested function at this point of a block.
	FuncStmt struct{ Func *Func }
	// Decl declares and initializes a local variable.
	Decl struct {
		Var  *Var
		Init Expr
	}
	// Assign stores Value in Var. Compound assignments and ++/-- are
	// lowered to Assign of an Arith.
	Assign struct {
		Var   *Var
		Value Expr
	}
	// Eval evaluates an expression for its effect.
	Eval struct{ X Expr }
	// IndexSet stores Value in element I of the array X. X, then I, are
	// evaluated once. For a compound assignment (a[i] += v), Value reads
	// the old element through IndexOld: the element is read, checking the
	// index, before the rest of Value is evaluated. The index is checked
	// again when storing. A bad index is a RangeError reported at Site.
	IndexSet struct {
		X, I     Expr
		Value    Expr
		Compound bool
		Site     Span
	}
	// Return returns Value, or nothing when Value is nil.
	Return struct{ Value Expr }
	// If runs Then when Cond holds, otherwise Else (nil, *Block, or *If).
	If struct {
		Cond Expr
		Then *Block
		Else Stmt
	}
	// While loops while Cond holds.
	While struct {
		Cond Expr
		Body *Block
	}
	// For is a C-style loop. Init, Cond, and Post may be nil. Init is a
	// *Decl, *Assign, or *Eval; Post is an *Assign or *Eval.
	For struct {
		Init Stmt
		Cond Expr
		Post Stmt
		Body *Block
	}
	// Break and Continue apply to the innermost loop.
	Break    struct{}
	Continue struct{}
	// Throw throws X, an error. Site is the throw statement; it becomes
	// the error's location if the error has none yet.
	Throw struct {
		X    Expr
		Site Span
	}
	// Try runs Body. An error thrown in it is caught by the first of
	// Catches whose type it has; Finally (nil if absent) runs however the
	// statement ends, except when the program stops for an error that
	// cannot be caught. A return, break, or continue in Finally replaces
	// that ending.
	Try struct {
		Body    *Block
		Catches []*Catch
		Finally *Block
	}
)

// Catch is one catch clause: it catches errors of type T (or a subtype)
// into Var.
type Catch struct {
	T    *types.Type
	Var  *Var
	Body *Block
}

func (*FuncStmt) stmt() {}
func (*Decl) stmt()     {}
func (*Assign) stmt()   {}
func (*Eval) stmt()     {}
func (*IndexSet) stmt() {}
func (*Return) stmt()   {}
func (*If) stmt()       {}
func (*While) stmt()    {}
func (*For) stmt()      {}
func (*Break) stmt()    {}
func (*Continue) stmt() {}
func (*Throw) stmt()    {}
func (*Try) stmt()      {}
func (*Block) stmt()    {}

// Expr is any expression. Type is never invalid or untyped.
type Expr interface{ Type() *types.Type }

type (
	// IntConst is an integer constant of a concrete integer type. Value is
	// within the type's range.
	IntConst struct {
		T     *types.Type
		Value int64
	}
	// StringConst is a string constant. Span locates the literal.
	StringConst struct {
		Value string
		Span  Span
	}
	BoolConst struct{ Value bool }
	// CharConst is a char constant, a Unicode scalar value.
	CharConst struct{ Value rune }
	// ArrayLit creates an array of type T holding Elems.
	ArrayLit struct {
		T     *types.Type
		Elems []Expr
	}
	// Index reads element I of the array X; a bad index is a RangeError
	// reported at Site.
	Index struct {
		X, I Expr
		Site Span
	}
	// IndexOld is the old element in the Value of a compound IndexSet.
	IndexOld struct{ T *types.Type }
	// ArrayLen is X.length.
	ArrayLen struct{ X Expr }
	// ArrayPush appends V to the array X.
	ArrayPush struct{ X, V Expr }
	// ArrayPop removes and returns the last element of X; an empty array
	// is a RangeError reported at Site.
	ArrayPop struct {
		X    Expr
		Site Span
	}
	// FloatConst is a floating-point constant of float type T, exactly
	// representable in T.
	FloatConst struct {
		T     *types.Type
		Value float64
	}
	// Neg is float negation. Integer negation is an Arith.
	Neg struct{ X Expr }
	// Convert is the checked conversion (To) X. A failure throws a
	// RangeError or TypeError reported at Site.
	Convert struct {
		X    Expr
		To   *types.Type
		Site Span
	}
	// NewError creates an error of type T. Message and ID may be nil
	// (they default to "" and 0).
	NewError struct {
		T           *types.Type
		Message, ID Expr
	}
	// FloatTest is the built-in isNaN(X) or isFinite(X).
	FloatTest struct {
		Name string
		X    Expr
	}
	// Field reads a field of an error: message, name, or errorID.
	Field struct {
		X    Expr
		Name string
		T    *types.Type
	}
	// Load reads a variable.
	Load struct{ Var *Var }
	// FuncRef is a named function used as a value.
	FuncRef struct{ Func *Func }
	// Lambda is an anonymous function value.
	Lambda struct{ Func *Func }
	// Arith is a range-checked integer operation (+ - * / %), or an IEEE
	// float operation (+ - * /), rounded to float32 for float32. X and Y
	// have type T. For integers, a result outside T's range or a zero
	// divisor is a runtime error reported at Site. Integer unary minus is
	// Arith{"-", 0, x}.
	Arith struct {
		Op   string
		T    *types.Type
		X, Y Expr
		Site Span
	}
	// Bits is the bitwise & | or ^ of X and Y, integers of type T. It
	// cannot overflow.
	Bits struct {
		Op   string
		T    *types.Type
		X, Y Expr
	}
	// Complement is ~X, the bitwise complement of an integer in its type's
	// width.
	Complement struct{ X Expr }
	// Shift is X << Count or X >> Count. X has the result's integer type;
	// Count is any integer. A count outside 0 to the width minus 1 throws a
	// RangeError, and so does a << that loses bits (an overflow), both
	// reported at Site. >> is arithmetic for signed types.
	Shift struct {
		Op       string
		X, Count Expr
		Site     Span
	}
	// Concat joins strings: string + string, += on a string, and
	// interpolated strings (whose holes are lowered to conversions).
	Concat struct{ Parts []Expr }
	// Compare is == != < <= > >=. Its operands have identical types.
	Compare struct {
		Op   string
		X, Y Expr
	}
	// Logic is the short-circuit && or ||.
	Logic struct {
		Op   string
		X, Y Expr
	}
	Not struct{ X Expr }
	// CallFunc calls a named function directly.
	CallFunc struct {
		Func *Func
		Args []Expr
		Site Span
	}
	// CallValue calls a function value. Callee is evaluated first.
	CallValue struct {
		Callee Expr
		Args   []Expr
		Site   Span
	}
	// CallNative calls a standard library function. Pkg is the qualifier
	// it was called with (an alias, or the package's own qualifier); the
	// package itself is Func.Package.
	CallNative struct {
		Pkg  string
		Func *stdlib.Func
		Args []Expr
		Site Span
	}
	// CallExtern calls an extern function. A result of the wrong type, or
	// a JavaScript exception, is a runtime error reported at Site.
	CallExtern struct {
		Extern *Extern
		Args   []Expr
		Site   Span
	}
	// Trace is the built-in trace(...).
	Trace struct{ Args []Expr }
	// NewSignal creates a rune of type T (a Signal) holding Init.
	NewSignal struct {
		T    *types.Type
		Init Expr
	}
	// SignalGet reads the rune X. Inside a template binding, reading a
	// rune subscribes the binding to it.
	SignalGet struct{ X Expr }
	// SignalSet stores Value in the rune X, updating everything that
	// read it.
	SignalSet struct{ X, Value Expr }
	// RenderApp is rune:dom's renderApp: it calls Component and puts the
	// element it returns into the first element Selector matches. No match
	// is a runtime error reported at Site.
	RenderApp struct {
		Component, Selector Expr
		Site                Span
	}
)

func (e *IntConst) Type() *types.Type    { return e.T }
func (e *StringConst) Type() *types.Type { return types.StringT }
func (e *BoolConst) Type() *types.Type   { return types.BoolT }
func (e *CharConst) Type() *types.Type   { return types.CharT }
func (e *ArrayLit) Type() *types.Type    { return e.T }
func (e *Index) Type() *types.Type       { return e.X.Type().Elem }
func (e *IndexOld) Type() *types.Type    { return e.T }
func (e *ArrayLen) Type() *types.Type    { return types.Int }
func (e *ArrayPush) Type() *types.Type   { return types.VoidT }
func (e *ArrayPop) Type() *types.Type    { return e.X.Type().Elem }
func (e *Load) Type() *types.Type        { return e.Var.Type }
func (e *FuncRef) Type() *types.Type     { return e.Func.Type }
func (e *Lambda) Type() *types.Type      { return e.Func.Type }
func (e *Arith) Type() *types.Type       { return e.T }
func (e *Compare) Type() *types.Type     { return types.BoolT }
func (e *Concat) Type() *types.Type      { return types.StringT }
func (e *Bits) Type() *types.Type        { return e.T }
func (e *Complement) Type() *types.Type  { return e.X.Type() }
func (e *Shift) Type() *types.Type       { return e.X.Type() }
func (e *Logic) Type() *types.Type       { return types.BoolT }
func (e *Not) Type() *types.Type         { return types.BoolT }
func (e *CallFunc) Type() *types.Type    { return e.Func.Type.Result }
func (e *CallValue) Type() *types.Type   { return e.Callee.Type().Result }
func (e *CallNative) Type() *types.Type  { return e.Func.Result }
func (e *Trace) Type() *types.Type       { return types.VoidT }
func (e *CallExtern) Type() *types.Type  { return e.Extern.Type.Result }
func (e *NewSignal) Type() *types.Type   { return e.T }
func (e *SignalGet) Type() *types.Type   { return e.X.Type().Elem }
func (e *SignalSet) Type() *types.Type   { return types.VoidT }
func (e *RenderApp) Type() *types.Type   { return types.VoidT }
func (e *FloatConst) Type() *types.Type  { return e.T }
func (e *Neg) Type() *types.Type         { return e.X.Type() }
func (e *Convert) Type() *types.Type     { return e.To }
func (e *NewError) Type() *types.Type    { return e.T }
func (e *Field) Type() *types.Type       { return e.T }
func (e *FloatTest) Type() *types.Type   { return types.BoolT }

// ---- templates (rune:dom) ----

// Template builds the element a template describes; a component returns
// it. Dynamic parts are bindings: they update when a rune they read
// changes.
type Template struct{ Root *TElement }

func (e *Template) Type() *types.Type { return types.ElementT }

// AttrMode is how a target sets an attribute at run time. It is all a
// back end needs to know about an attribute: the HTML/SVG metadata that
// the checker validates templates against (internal/dom) stays in the
// language tooling, and Lower translates it into a mode.
type AttrMode int

const (
	// AttrText sets the attribute to its value shown as text (a number or
	// a string).
	AttrText AttrMode = iota
	// AttrBool adds the attribute when its bool value is true and removes
	// it when false.
	AttrBool
	// AttrURL sets a string that must not run script (javascript: and
	// similar URLs are a runtime error).
	AttrURL
	// AttrEvent listens for the event named by the attribute (without its
	// "on" prefix); the value is a void() handler.
	AttrEvent
)

// TNode is a node of a template.
type TNode interface{ tnode() }

type (
	// TElement is an element. SVG elements are created in the SVG
	// namespace.
	TElement struct {
		Tag      string
		SVG      bool
		Attrs    []*TAttr
		Children []TNode
	}
	// TAttr is an attribute: static (Value nil) or a binding to Value.
	// For an event, Value is the handler, a void() function value.
	TAttr struct {
		Name string
		Mode AttrMode
		// Static is the value of a static attribute; Present marks a
		// static boolean attribute.
		Static  string
		Present bool
		Value   Expr
		// Values, when not nil, are the only keywords a binding may set;
		// any other value is a runtime error.
		Values []string
		Site   Span
	}
	// TText is static text.
	TText struct{ Text string }
	// TInterp is {{ X }} in content: text, or a child element.
	TInterp struct{ X Expr }
	// TIf shows the first branch whose condition holds, else Else.
	TIf struct {
		Conds    []Expr
		Branches [][]TNode
		Else     []TNode
	}
)

func (*TElement) tnode() {}
func (*TText) tnode()    {}
func (*TInterp) tnode()  {}
func (*TIf) tnode()      {}
