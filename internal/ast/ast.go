// Package ast defines the RuneScript abstract syntax tree.
package ast

import "github.com/AurielSolaris/runescript/internal/diagnostics"

// Span is re-exported for brevity.
type Span = diagnostics.Span

// Node is implemented by every AST node.
type Node interface{ NodeSpan() Span }

// File is one parsed source file.
type File struct {
	Package *PackageDecl
	Uses    []*UseDecl
	Funcs   []*FuncDecl
}

// UseDecl is `use "path";` or `use "path" as name;`. The path names a
// standard library package ("std:string"), a RuneScript package
// ("org.x.geom"), or a file ("./lib/geom.rsc"); see internal/loader.
type UseDecl struct {
	Path     string
	PathSpan Span
	Alias    *Ident // nil without `as`
	Span     Span
}

// PackageDecl is `package a.b.c;`.
type PackageDecl struct {
	Name []string
	Span Span
}

// Visibility is a function's access modifier. Every function declaration
// has one. Decided meaning: public = everywhere, private = this file,
// protected = this package. The checker enforces it on top-level
// functions; on nested functions it has no effect yet.
type Visibility int

const (
	VisNone Visibility = iota
	VisPublic
	VisProtected
	VisPrivate
)

func (v Visibility) String() string {
	return [...]string{"", "public", "protected", "private"}[v]
}

// FuncDecl is `visibility Type name(Type p, ...) { ... }`, or a prototype
// `visibility Type name(Type p, ...);` when Body is nil. It appears at top
// level and, as a statement, inside blocks (nested functions).
type FuncDecl struct {
	Visibility Visibility
	Result     *TypeRef
	Name       *Ident
	Params     []*Param
	Body       *Block // nil for a prototype
	Span       Span
}

// Param is a single function parameter. Name is nil only in prototypes.
type Param struct {
	Type *TypeRef
	Name *Ident
}

// TypeRef is a type as written in source: a named type (Name) or a function
// type (Func), e.g. `int(int, int)`.
type TypeRef struct {
	Name string
	Func *FuncTypeRef
	Span Span
}

// FuncTypeRef is `Result(Params...)`.
type FuncTypeRef struct {
	Result *TypeRef
	Params []*TypeRef
}

// String renders the type as written, e.g. "int(int, int)".
func (t *TypeRef) String() string {
	if t.Func == nil {
		return t.Name
	}
	s := t.Func.Result.String() + "("
	for i, p := range t.Func.Params {
		if i > 0 {
			s += ", "
		}
		s += p.String()
	}
	return s + ")"
}

// Ident is a name occurrence.
type Ident struct {
	Name string
	Span Span
}

// ---- statements ----

// Stmt is any statement.
type Stmt interface {
	Node
	stmt()
}

// Block is `{ stmts }`; it introduces a scope.
type Block struct {
	Stmts []Stmt
	Span  Span
}

// VarDecl is `Type name = init;`. An initializer is required (OQ-VAR-01).
type VarDecl struct {
	Type *TypeRef
	Name *Ident
	Init Expr
	Span Span
}

// AssignStmt is `name op value;` where op is one of = += -= *= /= %=.
type AssignStmt struct {
	Target *Ident
	Op     string
	Value  Expr
	Span   Span
}

// IncDecStmt is `name++;` or `name--;` (statement form only, OQ-OP-03).
type IncDecStmt struct {
	Target *Ident
	Op     string // "++" or "--"
	Span   Span
}

// ExprStmt is an expression evaluated for its effect; only calls are allowed.
type ExprStmt struct {
	Call *CallExpr
}

// ReturnStmt is `return;` or `return value;`.
type ReturnStmt struct {
	Value Expr // nil for a bare return
	Span  Span
}

// IfStmt is `if (cond) { } [else { } | else if ...]`.
type IfStmt struct {
	Cond Expr
	Then *Block
	Else Stmt // nil, *Block, or *IfStmt
	Span Span
}

// WhileStmt is `while (cond) { }`.
type WhileStmt struct {
	Cond Expr
	Body *Block
	Span Span
}

// ForStmt is `for (init; cond; post) { }`. Init, Cond, and Post may be nil.
type ForStmt struct {
	Init Stmt // *VarDecl, *AssignStmt, *IncDecStmt, *ExprStmt, or nil
	Cond Expr
	Post Stmt // *AssignStmt, *IncDecStmt, *ExprStmt, or nil
	Body *Block
	Span Span
}

// BranchStmt is `break;` or `continue;`.
type BranchStmt struct {
	Keyword string
	Span    Span
}

func (s *FuncDecl) NodeSpan() Span   { return s.Span }
func (s *Block) NodeSpan() Span      { return s.Span }
func (s *VarDecl) NodeSpan() Span    { return s.Span }
func (s *AssignStmt) NodeSpan() Span { return s.Span }
func (s *IncDecStmt) NodeSpan() Span { return s.Span }
func (s *ExprStmt) NodeSpan() Span   { return s.Call.Span }
func (s *ReturnStmt) NodeSpan() Span { return s.Span }
func (s *IfStmt) NodeSpan() Span     { return s.Span }
func (s *WhileStmt) NodeSpan() Span  { return s.Span }
func (s *ForStmt) NodeSpan() Span    { return s.Span }
func (s *BranchStmt) NodeSpan() Span { return s.Span }

func (*FuncDecl) stmt()   {}
func (*Block) stmt()      {}
func (*VarDecl) stmt()    {}
func (*AssignStmt) stmt() {}
func (*IncDecStmt) stmt() {}
func (*ExprStmt) stmt()   {}
func (*ReturnStmt) stmt() {}
func (*IfStmt) stmt()     {}
func (*WhileStmt) stmt()  {}
func (*ForStmt) stmt()    {}
func (*BranchStmt) stmt() {}

// ---- expressions ----

// Expr is any expression.
type Expr interface {
	Node
	expr()
}

// IntLit is a non-negative decimal integer literal. Negative constants are
// UnaryExpr{"-", IntLit}.
type IntLit struct {
	Value uint64
	Text  string
	Span  Span
}

// StringLit is a string literal with escapes already decoded.
type StringLit struct {
	Value string
	Span  Span
}

// BoolLit is `true` or `false`.
type BoolLit struct {
	Value bool
	Span  Span
}

// CallExpr is `name(args)`, `pkg.name(args)`, or `expr(args)`.
// Exactly one of Fn and Callee is set; Pkg is set only with Fn.
type CallExpr struct {
	Pkg    *Ident // package qualifier, nil for unqualified calls
	Fn     *Ident // named callee
	Callee Expr   // any other callee, e.g. makeAdder(1)(2)
	Args   []Expr
	Span   Span
}

// UnaryExpr is `-x` or `!x`.
type UnaryExpr struct {
	Op   string
	X    Expr
	Span Span
}

// BinaryExpr is `x op y`.
type BinaryExpr struct {
	Op    string
	X, Y  Expr
	OpPos Span
	Span  Span
}

// LambdaExpr is `(Type a, ...) => expr` or `(Type a, ...) => { ... }`.
// Exactly one of Expr and Body is set.
type LambdaExpr struct {
	Params []*Param
	Expr   Expr
	Body   *Block
	Span   Span
}

// ParenExpr is `(x)`.
type ParenExpr struct {
	X    Expr
	Span Span
}

func (e *IntLit) NodeSpan() Span     { return e.Span }
func (e *StringLit) NodeSpan() Span  { return e.Span }
func (e *BoolLit) NodeSpan() Span    { return e.Span }
func (e *Ident) NodeSpan() Span      { return e.Span }
func (e *CallExpr) NodeSpan() Span   { return e.Span }
func (e *UnaryExpr) NodeSpan() Span  { return e.Span }
func (e *BinaryExpr) NodeSpan() Span { return e.Span }
func (e *ParenExpr) NodeSpan() Span  { return e.Span }
func (e *LambdaExpr) NodeSpan() Span { return e.Span }

func (*IntLit) expr()     {}
func (*StringLit) expr()  {}
func (*BoolLit) expr()    {}
func (*Ident) expr()      {}
func (*CallExpr) expr()   {}
func (*UnaryExpr) expr()  {}
func (*BinaryExpr) expr() {}
func (*ParenExpr) expr()  {}
func (*LambdaExpr) expr() {}
