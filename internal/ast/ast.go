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
	Externs []*ExternDecl
}

// ExternDecl is `extern "module" Type name(Type p, ...);`: a function
// implemented in JavaScript by the named module. It is visible only in
// the file that declares it.
type ExternDecl struct {
	Module     string
	ModuleSpan Span
	Result     *TypeRef
	Name       *Ident
	Params     []*Param // names are optional
	Span       Span
}

func (d *ExternDecl) NodeSpan() Span { return d.Span }

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
	// Args are the type arguments of a named type, e.g. int in Signal<int>.
	Args []*TypeRef
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
		if len(t.Args) == 0 {
			return t.Name
		}
		s := t.Name + "<"
		for i, a := range t.Args {
			if i > 0 {
				s += ", "
			}
			s += a.String()
		}
		return s + ">"
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

// ThrowStmt is `throw value;`.
type ThrowStmt struct {
	Value Expr
	Span  Span
}

// TryStmt is `try { } catch (T e) { } ... finally { }`. It has at least
// one catch clause or a finally block.
type TryStmt struct {
	Body    *Block
	Catches []*CatchClause
	Finally *Block // nil without finally
	Span    Span
}

// CatchClause is `catch (Type name) { }`, or `catch (name) { }`, which
// catches every error.
type CatchClause struct {
	Type *TypeRef // nil for catch (name)
	Name *Ident
	Body *Block
	Span Span
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
func (s *ThrowStmt) NodeSpan() Span  { return s.Span }
func (s *TryStmt) NodeSpan() Span    { return s.Span }

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
func (*ThrowStmt) stmt()  {}
func (*TryStmt) stmt()    {}

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

// FloatLit is a non-negative floating-point literal. Value is the nearest
// float64 to the literal as written.
type FloatLit struct {
	Value float64
	Text  string
	Span  Span
}

// StringLit is a string literal with escapes already decoded.
type StringLit struct {
	Value string
	Span  Span
}

// InterpString is `$"text {x} text"`. Parts are the decoded text pieces
// around the holes, so len(Parts) == len(Exprs)+1.
type InterpString struct {
	Parts []string
	Exprs []Expr
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

// CastExpr is the checked conversion `(Type) x`.
type CastExpr struct {
	Type *TypeRef
	X    Expr
	Span Span
}

// NewExpr is `new Type(args)`. Only the built-in error types can be
// constructed so far.
type NewExpr struct {
	Type *TypeRef
	Args []Expr
	Span Span
}

// SelectorExpr is `x.name`, a field of a value such as e.message.
type SelectorExpr struct {
	X    Expr
	Name *Ident
	Span Span
}

func (e *IntLit) NodeSpan() Span       { return e.Span }
func (e *StringLit) NodeSpan() Span    { return e.Span }
func (e *InterpString) NodeSpan() Span { return e.Span }
func (e *BoolLit) NodeSpan() Span      { return e.Span }
func (e *Ident) NodeSpan() Span        { return e.Span }
func (e *CallExpr) NodeSpan() Span     { return e.Span }
func (e *UnaryExpr) NodeSpan() Span    { return e.Span }
func (e *BinaryExpr) NodeSpan() Span   { return e.Span }
func (e *ParenExpr) NodeSpan() Span    { return e.Span }
func (e *FloatLit) NodeSpan() Span     { return e.Span }
func (e *CastExpr) NodeSpan() Span     { return e.Span }
func (e *NewExpr) NodeSpan() Span      { return e.Span }
func (e *SelectorExpr) NodeSpan() Span { return e.Span }
func (e *LambdaExpr) NodeSpan() Span   { return e.Span }

func (*IntLit) expr()       {}
func (*StringLit) expr()    {}
func (*InterpString) expr() {}
func (*BoolLit) expr()      {}
func (*Ident) expr()        {}
func (*CallExpr) expr()     {}
func (*UnaryExpr) expr()    {}
func (*BinaryExpr) expr()   {}
func (*ParenExpr) expr()    {}
func (*FloatLit) expr()     {}
func (*CastExpr) expr()     {}
func (*NewExpr) expr()      {}
func (*SelectorExpr) expr() {}
func (*LambdaExpr) expr()   {}

// ---- templates (rune:dom) ----

// TemplateStmt is `template { markup }`, the last statement of a function
// returning Element: the element the markup describes is the result.
// Root is the template's one root element.
type TemplateStmt struct {
	Root *MElement
	Span Span
}

func (s *TemplateStmt) NodeSpan() Span { return s.Span }
func (*TemplateStmt) stmt()            {}

// Markup is a node of a template.
type Markup interface {
	Node
	markup()
}

// MElement is an element: `<name attrs>children</name>` or `<name attrs/>`.
type MElement struct {
	Name     string
	NameSpan Span
	Attrs    []*MAttr
	Children []Markup
	Span     Span
}

// MAttr is an attribute: `name`, `name="value"`, or `name={{ expr }}`.
// Value is nil for a bare name and for an interpolation (Expr).
type MAttr struct {
	Name      string
	NameSpan  Span
	Value     *string
	ValueSpan Span
	Expr      Expr
	Span      Span
}

// MText is text, with character references decoded and white space
// containing a line break removed at its ends and collapsed inside it.
type MText struct {
	Text string
	Span Span
}

// MInterp is `{{ expr }}` in content.
type MInterp struct {
	X    Expr
	Span Span
}

// MIf is `$if (c) … $elseif (c) … $else … $endif`. Conds[i] guards
// Branches[i]; Else is nil without $else.
type MIf struct {
	Conds    []Expr
	Branches [][]Markup
	Else     []Markup
	HasElse  bool
	Span     Span
}

func (n *MElement) NodeSpan() Span { return n.Span }
func (n *MText) NodeSpan() Span    { return n.Span }
func (n *MInterp) NodeSpan() Span  { return n.Span }
func (n *MIf) NodeSpan() Span      { return n.Span }
func (*MElement) markup()          {}
func (*MText) markup()             {}
func (*MInterp) markup()           {}
func (*MIf) markup()               {}
