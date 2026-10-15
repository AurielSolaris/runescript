// Package ir defines RuneScript's typed intermediate representation.
//
// The IR is produced from a checked AST (see Lower) and consumed by code
// generators such as internal/jsgen. Every expression carries its concrete
// type: there are no untyped constants, no invalid types, and no names left
// to resolve. Integer arithmetic is a distinct node (Arith) with the
// source location a runtime error is reported at, so a back end cannot emit
// an integer operation without also emitting its range check.
//
// Verify re-checks the typing rules on the IR itself, independently of the
// checker, before any code is generated.
package ir

import (
	"github.com/AurielSolaris/runescript/internal/diagnostics"
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
)

func (*FuncStmt) stmt() {}
func (*Decl) stmt()     {}
func (*Assign) stmt()   {}
func (*Eval) stmt()     {}
func (*Return) stmt()   {}
func (*If) stmt()       {}
func (*While) stmt()    {}
func (*For) stmt()      {}
func (*Break) stmt()    {}
func (*Continue) stmt() {}
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
	// Load reads a variable.
	Load struct{ Var *Var }
	// FuncRef is a named function used as a value.
	FuncRef struct{ Func *Func }
	// Lambda is an anonymous function value.
	Lambda struct{ Func *Func }
	// Arith is a range-checked integer operation: + - * / %. X and Y have
	// type T. A result outside T's range, or a zero divisor, is a runtime
	// error reported at Site. Unary minus is Arith{"-", 0, x}.
	Arith struct {
		Op   string
		T    *types.Type
		X, Y Expr
		Site Span
	}
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
	// Trace is the built-in trace(...).
	Trace struct{ Args []Expr }
)

func (e *IntConst) Type() *types.Type    { return e.T }
func (e *StringConst) Type() *types.Type { return types.StringT }
func (e *BoolConst) Type() *types.Type   { return types.BoolT }
func (e *Load) Type() *types.Type        { return e.Var.Type }
func (e *FuncRef) Type() *types.Type     { return e.Func.Type }
func (e *Lambda) Type() *types.Type      { return e.Func.Type }
func (e *Arith) Type() *types.Type       { return e.T }
func (e *Compare) Type() *types.Type     { return types.BoolT }
func (e *Logic) Type() *types.Type       { return types.BoolT }
func (e *Not) Type() *types.Type         { return types.BoolT }
func (e *CallFunc) Type() *types.Type    { return e.Func.Type.Result }
func (e *CallValue) Type() *types.Type   { return e.Callee.Type().Result }
func (e *CallNative) Type() *types.Type  { return e.Func.Result }
func (e *Trace) Type() *types.Type       { return types.VoidT }
