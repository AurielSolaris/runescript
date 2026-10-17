// Package checker performs name resolution and static type checking.
//
// Where the specification is undecided, the checker rejects the program
// rather than choosing a semantics. Rejection is forward-compatible: a later
// decision can accept more programs without changing the meaning of
// programs that are accepted today.
//
// The checker is split into checker.go (scopes and declarations),
// stmts.go, and exprs.go.
package checker

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/dom"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// Info is the result of checking, consumed by the runtime.
type Info struct {
	// Program is the program that was checked.
	Program *loader.Program
	// Types holds the resolved type of every checked expression.
	Types map[ast.Expr]*types.Type
	// Consts holds the value of integer constant expressions that were
	// folded and converted to a concrete type. The runtime uses these
	// values instead of evaluating the expression.
	Consts map[ast.Expr]int64
	// Callees maps direct calls of named functions (top-level or nested)
	// to the function called.
	Callees map[*ast.CallExpr]*FuncSym
	// FuncRefs maps identifiers that name a function used as a value.
	FuncRefs map[*ast.Ident]*FuncSym
	// Natives maps calls to standard library functions to their
	// implementation.
	Natives map[*ast.CallExpr]*stdlib.Func
	// FloatConsts holds the value of floating-point constants that were
	// given a concrete float type, like Consts for integers. The value is
	// exactly representable in that type.
	FloatConsts map[ast.Expr]float64
	// CatchTypes holds the error type each catch clause catches.
	CatchTypes map[*ast.CatchClause]*types.Type
	// Builtins maps calls of the built-in functions isNaN and isFinite to
	// their name. (trace is recognised by name, as it cannot be hidden.)
	Builtins map[*ast.CallExpr]string
	// Externs maps calls of extern (JavaScript) functions to the function
	// called; ExternDecls lists every extern function declared.
	Externs     map[*ast.CallExpr]*ExternSym
	ExternDecls []*ExternSym
	// DOM reports whether the program uses rune:dom; DOMUses are those
	// use declarations. Signals maps calls that create, read, or set a
	// rune; DOMCalls maps calls of rune:dom functions to their name.
	DOM      bool
	DOMUses  []*ast.UseDecl
	Signals  map[*ast.CallExpr]SignalOp
	DOMCalls map[*ast.CallExpr]string
	// QualRefs maps pkg.name used as a value to the function it names.
	QualRefs map[*ast.SelectorExpr]*FuncSym
	// TypeRefs holds the type every written type resolved to.
	TypeRefs map[*ast.TypeRef]*types.Type
	// ArrayCalls maps calls of array methods (a.push(v), a.pop()) to the
	// method's name. The array is e.Pkg, or the X of the SelectorExpr
	// callee.
	ArrayCalls map[*ast.CallExpr]string
	// TemplateElems and TemplateAttrs map the elements and attributes of
	// templates to their entries in the element list.
	TemplateElems map[*ast.MElement]*dom.Element
	TemplateAttrs map[*ast.MAttr]*dom.Attr
	// Entry is the program entry point, `protected int main()`.
	Entry *ast.FuncDecl
	// Warnings holds the warnings found, in no particular order. They are
	// not in the diagnostics Check returns.
	Warnings []*diagnostics.Diagnostic
}

// FuncSym is a named function: top-level or nested, possibly declared by a
// prototype before its definition.
type FuncSym struct {
	Name  string
	Type  *types.Type
	Def   *ast.FuncDecl // nil until the definition is seen
	Proto *ast.FuncDecl // the prototype, if any
	// For top-level functions, the file and package declaring them; nil
	// for nested functions.
	File *loader.File
	Pkg  *loader.Package
}

// TopLevel reports whether f is a top-level function.
func (f *FuncSym) TopLevel() bool { return f.Pkg != nil }

// visibility is the access modifier of f's first declaration.
func (f *FuncSym) visibility() ast.Visibility { return firstDecl(f).Visibility }

// symbol is anything a name can refer to inside a function.
type symbol struct {
	typ     *types.Type // variable type
	isParam bool
	isCatch bool       // the error variable of a catch clause (read-only)
	fn      *FuncSym   // non-nil for named functions
	ext     *ExternSym // non-nil for extern functions (typ is their type)
	st      *StructSym // non-nil for struct types (typ is the struct type)
	decl    *ast.Ident // where a variable or parameter is declared
	read    bool       // the variable's value is used somewhere
	written bool       // the variable is assigned after its declaration
}

// scope is one lexical scope. A boundary scope is the outermost scope of a
// function or lambda body: variables declared outside it are not visible
// inside it (functions do not capture), but functions are.
type scope struct {
	syms     map[string]*symbol
	order    []string
	boundary bool
}

type checker struct {
	// structs are the declared structs, in declaration order; fileQuals
	// and fileBroken are each file's use qualifiers, bound once.
	structs    []*StructSym
	fileQuals  map[*loader.File]map[string]*binding
	fileBroken map[*loader.File]bool
	info       *Info
	diags      []*diagnostics.Diagnostic
	scopes     []*scope
	pkgs       map[*loader.Package]*scope // public and protected top-level functions
	files      map[*loader.File]*scope    // private top-level functions of each file
	// domUsed holds the files that use rune:dom's types or signal(), so
	// their use "rune:dom" is not reported as unused.
	domUsed map[*loader.File]bool

	// state of the file being checked
	file   *loader.File
	quals  map[string]*binding // use qualifiers
	broken bool                // a use with an unknown qualifier could not be loaded
	dom    bool                // the file uses rune:dom

	// state of the function or lambda being checked
	fnName       string
	result       *types.Type // nil while inferring a lambda's result type
	loopDepth    int
	lastTemplate *ast.TemplateStmt // the template that may end this function's body
}

// operand is the checked form of an expression.
type operand struct {
	typ    *types.Type
	konst  *big.Int // non-nil only for untyped integer constants
	fconst *float64 // non-nil only for untyped float constants
}

var invalid = operand{typ: types.InvalidT}

// binding is what a use qualifier refers to in one file: a standard
// library package or a RuneScript package (neither for a broken use).
type binding struct {
	std  *stdlib.Package
	pkg  *loader.Package
	decl *ast.UseDecl
	used bool   // the qualifier is referenced
	rune string // a rune: module, e.g. "dom"
}

// Check type-checks a loaded program. Every file must have parsed. It
// returns all diagnostics found.
func Check(prog *loader.Program) (*Info, []*diagnostics.Diagnostic) {
	c := &checker{
		info: &Info{
			Program:  prog,
			Types:    map[ast.Expr]*types.Type{},
			Consts:   map[ast.Expr]int64{},
			Callees:  map[*ast.CallExpr]*FuncSym{},
			FuncRefs: map[*ast.Ident]*FuncSym{},
			Natives:  map[*ast.CallExpr]*stdlib.Func{},

			FloatConsts: map[ast.Expr]float64{},
			CatchTypes:  map[*ast.CatchClause]*types.Type{},
			Builtins:    map[*ast.CallExpr]string{},
			Externs:     map[*ast.CallExpr]*ExternSym{},
			Signals:     map[*ast.CallExpr]SignalOp{},
			DOMCalls:    map[*ast.CallExpr]string{},
			QualRefs:    map[*ast.SelectorExpr]*FuncSym{},
			ArrayCalls:  map[*ast.CallExpr]string{},
			TypeRefs:    map[*ast.TypeRef]*types.Type{},

			TemplateElems: map[*ast.MElement]*dom.Element{},
			TemplateAttrs: map[*ast.MAttr]*dom.Attr{},
		},
		pkgs:       map[*loader.Package]*scope{},
		files:      map[*loader.File]*scope{},
		fileQuals:  map[*loader.File]map[string]*binding{},
		fileBroken: map[*loader.File]bool{},
		domUsed:    map[*loader.File]bool{},
	}
	for _, p := range prog.Packages {
		c.pkgs[p] = &scope{syms: map[string]*symbol{}}
	}
	for _, f := range prog.Files {
		c.files[f] = &scope{syms: map[string]*symbol{}}
	}

	// Each file's use qualifiers are bound once, first: signatures and
	// struct fields may name another package's struct (geom.Point).
	for _, f := range prog.Files {
		c.enterFile(f)
		c.useDecls(f)
		c.fileQuals[f], c.fileBroken[f] = c.quals, c.broken
	}
	// A package's structs and top-level functions are visible in all of
	// its files, so declare every one before checking any body: struct
	// names first, then their fields, then function signatures.
	for _, f := range prog.Files {
		c.enterFile(f)
		for _, d := range f.AST.Structs {
			c.declareStruct(d)
		}
	}
	c.structFields()
	sigs := map[*ast.FuncDecl]*types.Type{}
	for _, f := range prog.Files {
		c.enterFile(f)
		for _, fn := range f.AST.Funcs {
			sigs[fn] = c.declareTop(fn)
		}
	}
	for _, f := range prog.Files {
		c.enterFile(f)
		c.declareExterns(f.AST)
	}
	for _, f := range prog.Files {
		c.enterFile(f)
		for _, fn := range f.AST.Funcs {
			if fn.Body != nil {
				c.funcBody(fn, sigs[fn])
			}
		}
		c.reportUnusedUses()
	}
	for _, f := range prog.Files {
		c.enterFile(f)
		c.reportUndefined(c.files[f])
	}
	for _, p := range prog.Packages {
		c.reportUndefined(c.pkgs[p])
	}
	c.entry(prog)
	return c.info, c.diags
}

// enterFile makes f the file being checked.
func (c *checker) enterFile(f *loader.File) {
	c.file = f
	c.scopes = []*scope{c.pkgs[f.Package], c.files[f]}
	c.quals, c.broken = c.fileQuals[f], c.fileBroken[f]
	if c.quals == nil {
		c.quals = map[string]*binding{}
	}
	c.dom = false
	for _, u := range f.Uses {
		if u.Kind == loader.UseRune && u.Name == "dom" {
			c.dom = true
		}
	}
}

func (c *checker) errorf(code diagnostics.Code, span diagnostics.Span, format string, args ...any) *diagnostics.Diagnostic {
	d := diagnostics.New(code, span, format, args...)
	c.diags = append(c.diags, d)
	return d
}

// ---- use declarations ----

// useDecls binds the qualifiers of f's use declarations. The loader has
// already resolved RuneScript packages and files and reported the uses it
// could not resolve.
func (c *checker) useDecls(f *loader.File) {
	for _, u := range f.Uses {
		switch u.Kind {
		case loader.UseStd:
			pkg := stdlib.Lookup(u.Name)
			if pkg == nil {
				paths := stdlib.Paths()
				sort.Strings(paths)
				c.errorf(diagnostics.UseUnknownPackage, u.Decl.PathSpan, "unknown standard library package %q", u.Decl.Path).
					WithHelp("available: std:" + strings.Join(paths, ", std:"))
				continue
			}
			q := u.Qualifier
			if q == "" {
				q = pkg.Qualifier
			}
			c.bind(q, &binding{std: pkg, decl: u.Decl})
		case loader.UseRune:
			if !runeModules[u.Name] {
				c.runeModuleError(u.Name, u.Decl)
				continue
			}
			c.info.DOM = true
			c.info.DOMUses = append(c.info.DOMUses, u.Decl)
			c.bind(u.Qualifier, &binding{rune: u.Name, decl: u.Decl})
		case loader.UsePackage, loader.UseFile:
			switch {
			case u.Self(f):
				// Its functions are already in scope.
			case u.Target == nil && u.Qualifier != "":
				c.bind(u.Qualifier, &binding{decl: u.Decl})
			case u.Target == nil:
				// A missing file: its package name, and so its qualifier,
				// is unknown.
				c.broken = true
			default:
				c.bind(u.Qualifier, &binding{pkg: u.Target, decl: u.Decl})
			}
		}
	}
}

// bind makes qualifier q refer to b in the current file.
func (c *checker) bind(q string, b *binding) {
	prev := c.quals[q]
	switch {
	case prev == nil:
		c.quals[q] = b
	case prev.std != nil && prev.std == b.std, prev.pkg != nil && prev.pkg == b.pkg, prev.rune != "" && prev.rune == b.rune:
		c.errorf(diagnostics.UseDuplicate, b.decl.Span, "%q is already used at %s", b.decl.Path, prev.decl.Span.Start)
	default:
		span := b.decl.PathSpan
		if b.decl.Alias != nil {
			span = b.decl.Alias.Span
		}
		c.errorf(diagnostics.UseQualifierClash, span, "'%s' already refers to %q, used at %s", q, prev.decl.Path, prev.decl.Span.Start).
			WithHelp(fmt.Sprintf("give one of them another name, e.g. use %q as other;", b.decl.Path))
	}
}

// ---- scopes ----

func (c *checker) push(boundary bool) {
	c.scopes = append(c.scopes, &scope{syms: map[string]*symbol{}, boundary: boundary})
}

func (c *checker) pop() {
	c.reportUndefined(c.scopes[len(c.scopes)-1])
	c.reportUnused(c.scopes[len(c.scopes)-1])
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *checker) top() *scope { return c.scopes[len(c.scopes)-1] }

// lookup finds name in the visible scopes. crossed reports whether the
// symbol lies outside the current function boundary.
func (c *checker) lookup(name string) (sym *symbol, crossed bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if s, ok := c.scopes[i].syms[name]; ok {
			return s, crossed
		}
		if c.scopes[i].boundary {
			crossed = true
		}
	}
	return nil, false
}

func (s *scope) add(name string, sym *symbol) {
	if _, exists := s.syms[name]; !exists {
		s.order = append(s.order, name)
	}
	s.syms[name] = sym
}

// conflict reports whether declaring name would reuse a visible name.
// Variables of enclosing functions are not visible across a function
// boundary, so they do not conflict. Shadowing is otherwise rejected
// (OQ-SCOPE-01).
func (c *checker) conflict(id *ast.Ident) bool {
	if id.Name == "trace" {
		c.errorf(diagnostics.NameBuiltinReused, id.Span, "'trace' is a built-in function and cannot be redeclared")
		return true
	}
	prev, crossed := c.lookup(id.Name)
	switch {
	case prev == nil, crossed && prev.fn == nil && prev.ext == nil && !isRune(prev):
		return false
	case prev.ext != nil:
		c.errorf(diagnostics.NameRedeclared, id.Span, "'%s' is already declared as an extern function", id.Name).
			WithHelp("shadowing is not allowed; choose a different name")
	case prev.fn != nil:
		c.errorf(diagnostics.NameRedeclared, id.Span, "'%s' is already declared as a function", id.Name).
			WithHelp("shadowing is not allowed; choose a different name")
	default:
		c.errorf(diagnostics.NameRedeclared, id.Span, "'%s' is already declared in this function", id.Name).
			WithHelp("shadowing is not allowed; choose a different name")
	}
	return true
}

func (c *checker) declareVar(id *ast.Ident, t *types.Type, isParam bool) {
	if !c.shadows(id) {
		c.conflict(id)
	}
	c.top().add(id.Name, &symbol{typ: t, isParam: isParam, decl: id})
}

// shadows reports whether declaring the variable id in a nested block
// hides a variable or parameter of an enclosing block of the same
// function. That is allowed, with a warning (owner decision, OQ-SCOPE-01);
// reusing a name in the same block, or the name of a function, struct, or
// extern, is still an error (conflict).
func (c *checker) shadows(id *ast.Ident) bool {
	if _, here := c.top().syms[id.Name]; here || c.top().boundary {
		return false
	}
	prev, crossed := c.lookup(id.Name)
	if prev == nil || crossed || prev.decl == nil || prev.fn != nil || prev.ext != nil || prev.st != nil {
		return false
	}
	what := "variable"
	if prev.isParam {
		what = "parameter"
	}
	c.warnf(diagnostics.WarnShadow, id.Span, "'%s' hides the %s declared at %s", id.Name, what, prev.decl.Span.Start.Where(c.file.Source)).
		WithHelp("the outer one cannot be used in this block; rename one of them if that is not intended")
	return true
}

// reportUndefined reports prototypes in s that never got a definition.
func (c *checker) reportUndefined(s *scope) {
	for _, name := range s.order {
		if f := s.syms[name].fn; f != nil && f.Def == nil && f.Proto != nil {
			c.errorf(diagnostics.NameNeverDefined, f.Proto.Name.Span, "function '%s' is declared but never defined", name).
				WithHelp("add a definition with a body in the same scope")
		}
	}
}

// ---- functions ----

func (c *checker) resolveType(ref *ast.TypeRef) *types.Type {
	t := c.resolveTypeRef(ref)
	c.info.TypeRefs[ref] = t
	return t
}

func (c *checker) resolveTypeRef(ref *ast.TypeRef) *types.Type {
	if ref.Pkg != nil {
		if t := c.structType(ref); t != nil {
			return t
		}
		return types.InvalidT
	}
	if ref.Func != nil {
		params := make([]*types.Type, len(ref.Func.Params))
		for i, p := range ref.Func.Params {
			params[i] = c.resolveType(p)
			if params[i].Kind == types.Void {
				c.errorf(diagnostics.TypeVoidValue, p.Span, "function type parameters cannot be 'void'")
				params[i] = types.InvalidT
			}
		}
		return types.NewFunc(params, c.resolveType(ref.Func.Result))
	}
	if ref.Name == "Element" || ref.Name == "Signal" {
		return c.domType(ref)
	}
	if ref.Name == "Array" {
		return c.arrayType(ref)
	}
	if len(ref.Args) > 0 {
		c.errorf(diagnostics.TypeArgs, ref.Span, "type '%s' takes no type arguments", ref.Name)
		return types.InvalidT
	}
	t := types.Lookup(ref.Name)
	if t == nil {
		if st := c.structType(ref); st != nil {
			return st
		}
		d := c.errorf(diagnostics.TypeUnknown, ref.Span, "unknown type '%s'", ref.Name)
		switch ref.Name {
		case "any", "Object", "object", "dynamic", "var", "auto", "unknown", "Any":
			d.WithHelp("RuneScript has no catch-all type: every variable, parameter, and result has one exact type (guide, chapter 17)")
		case "T", "U", "K", "V":
			d.WithHelp("RuneScript has no user-defined generics: write the function for each concrete type (guide, chapter 17)")
		}
		return types.InvalidT
	}
	return t
}

// funcType resolves a declaration's signature.
func (c *checker) funcType(fn *ast.FuncDecl) *types.Type {
	params := make([]*types.Type, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = c.resolveType(p.Type)
		if params[i].Kind == types.Void {
			c.errorf(diagnostics.TypeVoidValue, p.Type.Span, "parameter '%s' cannot have type 'void'", paramName(p)).
				WithHelp("use an empty parameter list for a function that takes nothing")
			params[i] = types.InvalidT
		}
	}
	return types.NewFunc(params, c.resolveType(fn.Result))
}

func paramName(p *ast.Param) string {
	if p.Name == nil {
		return "_"
	}
	return p.Name.Name
}

// declareTop declares a top-level function. Private functions go in the
// file's scope and the others in the package's, so two files may each
// have a private function of the same name.
func (c *checker) declareTop(fn *ast.FuncDecl) *types.Type {
	name := fn.Name.Name
	fileScope, pkgScope := c.files[c.file], c.pkgs[c.file.Package]
	switch {
	case fileScope.syms[name] != nil:
		return c.declareIn(fileScope, fn, true)
	case pkgScope.syms[name] != nil:
		return c.declareIn(pkgScope, fn, true)
	case fn.Visibility == ast.VisPrivate:
		return c.declareIn(fileScope, fn, true)
	}
	if other := c.privateElsewhere(c.file.Package, name); other != nil {
		c.errorf(diagnostics.NameRedeclared, fn.Name.Span, "function '%s' is already declared, as private, at %s", name, firstDecl(other).Name.Span.Start.Where(c.file.Source)).
			WithHelp("only private functions of different files may share a name")
		return c.funcType(fn)
	}
	return c.declareIn(pkgScope, fn, true)
}

// privateElsewhere returns a private function called name in a file of
// pkg other than the current one.
func (c *checker) privateElsewhere(pkg *loader.Package, name string) *FuncSym {
	for _, f := range pkg.Files {
		if f == c.file {
			continue
		}
		if sym := c.files[f].syms[name]; sym != nil && sym.fn != nil {
			return sym.fn
		}
	}
	return nil
}

// declareFunc declares a nested function (definition or prototype) in the
// innermost scope and returns its signature.
func (c *checker) declareFunc(fn *ast.FuncDecl) *types.Type {
	return c.declareIn(c.top(), fn, false)
}

func (c *checker) declareIn(s *scope, fn *ast.FuncDecl, topLevel bool) *types.Type {
	t := c.funcType(fn)
	name := fn.Name.Name
	if existing, ok := s.syms[name]; ok {
		f := existing.fn
		switch {
		case f != nil && f.Def == nil && fn.Body != nil && f.File != nil && f.File != c.file:
			c.errorf(diagnostics.NameRedeclared, fn.Name.Span, "function '%s' has a prototype in %s; define it in that file", name, f.File.Source.Name).
				WithHelp("a package's functions are visible in all of its files without prototypes")
		case f != nil && f.Def == nil && fn.Body != nil:
			if !types.Identical(f.Type, t) {
				c.errorf(diagnostics.TypeProtoMismatch, fn.Name.Span,
					"definition of '%s' has type '%s', but its prototype at %s declares '%s'",
					name, t, f.Proto.Name.Span.Start.Where(c.file.Source), f.Type)
			}
			f.Def = fn
		case f != nil:
			c.errorf(diagnostics.NameRedeclared, fn.Name.Span, "function '%s' is already declared at %s", name, firstDecl(f).Name.Span.Start.Where(c.file.Source)).
				WithHelp("functions cannot be overloaded; a prototype may appear once, before the definition")
		case existing.st != nil:
			c.errorf(diagnostics.NameRedeclared, fn.Name.Span, "'%s' is already declared as a struct", name)
		default:
			c.errorf(diagnostics.NameRedeclared, fn.Name.Span, "'%s' is already declared in this function", name)
		}
		return t
	}
	if c.conflict(fn.Name) {
		return t
	}
	f := &FuncSym{Name: name, Type: t}
	if topLevel {
		f.File, f.Pkg = c.file, c.file.Package
	}
	if fn.Body != nil {
		f.Def = fn
	} else {
		f.Proto = fn
	}
	s.add(name, &symbol{fn: f})
	return t
}

func firstDecl(f *FuncSym) *ast.FuncDecl {
	if f.Proto != nil {
		return f.Proto
	}
	return f.Def
}

// funcBody checks a function definition's body inside a new boundary scope.
func (c *checker) funcBody(fn *ast.FuncDecl, t *types.Type) {
	defer c.saveFuncState()()
	c.fnName = "function '" + fn.Name.Name + "'"
	c.result = t.Result
	c.push(true)
	for i, p := range fn.Params {
		c.declareVar(p.Name, t.Params[i], true)
	}
	c.lastTemplate = templateFor(fn.Body.Stmts)
	c.stmts(fn.Body.Stmts)
	c.pop()
	c.checkReturns(fn.Body)
}

// saveFuncState saves the per-function state and returns a function that
// restores it, for checking nested functions and lambdas.
func (c *checker) saveFuncState() func() {
	name, result, loops, tmpl := c.fnName, c.result, c.loopDepth, c.lastTemplate
	c.loopDepth, c.lastTemplate = 0, nil
	return func() { c.fnName, c.result, c.loopDepth, c.lastTemplate = name, result, loops, tmpl }
}

// checkReturns reports a non-void body that can finish without returning.
func (c *checker) checkReturns(body *ast.Block) {
	if c.result == nil || c.result.Kind == types.Void || c.result.Kind == types.Invalid || terminates(body.Stmts) {
		return
	}
	end := body.Span.End
	start := end
	start.Offset--
	start.Col--
	c.errorf(diagnostics.TypeMissingReturn, diagnostics.Span{Start: start, End: end},
		"%s must return a value of type '%s' on every path", c.fnName, c.result).
		WithHelp("add a return statement at the end")
}

// terminates reports whether stmts always end in a return or a throw.
// Loops are never considered terminating (OQ-FLOW-01).
func terminates(stmts []ast.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	switch s := stmts[len(stmts)-1].(type) {
	case *ast.ReturnStmt, *ast.ThrowStmt, *ast.TemplateStmt:
		return true
	case *ast.IfStmt:
		return ifTerminates(s)
	case *ast.TryStmt:
		if s.Finally != nil && terminates(s.Finally.Stmts) {
			return true
		}
		if !terminates(s.Body.Stmts) {
			return false
		}
		for _, cl := range s.Catches {
			if !terminates(cl.Body.Stmts) {
				return false
			}
		}
		return true
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

// checkProtected reports a call, through a use qualifier, of a protected
// function of another package.
func (c *checker) checkProtected(f *FuncSym, span diagnostics.Span) {
	if f.visibility() == ast.VisProtected && f.Pkg != c.file.Package {
		c.errorf(diagnostics.NameNotVisible, span, "function '%s' is protected: only package %s can use it", f.Name, f.Pkg.Name).
			WithHelp("declare it public to use it from other packages")
	}
}

// privateError reports name, at span, if it is a private function of
// another file of pkg, and reports whether it was.
func (c *checker) privateError(pkg *loader.Package, name string, span diagnostics.Span) bool {
	f := c.privateElsewhere(pkg, name)
	if f == nil {
		return false
	}
	help := fmt.Sprintf("declare it protected to share it within package %s, or public to share it with every package", f.Pkg.Name)
	if pkg != c.file.Package {
		help = "declare it public to use it from other packages"
	}
	c.errorf(diagnostics.NameNotVisible, span, "function '%s' is private to %s", name, f.File.Source.Name).WithHelp(help)
	return true
}

// entry finds the program's entry point: the one top-level function named
// main, in any file (owner decision, v0.3).
func (c *checker) entry(prog *loader.Program) {
	var mains []*FuncSym
	var scopes []*scope
	for _, p := range prog.Packages {
		scopes = append(scopes, c.pkgs[p])
	}
	for _, f := range prog.Files {
		scopes = append(scopes, c.files[f])
	}
	for _, s := range scopes {
		if sym := s.syms["main"]; sym != nil && sym.fn != nil && sym.fn.Def != nil {
			mains = append(mains, sym.fn)
		}
	}
	sort.SliceStable(mains, func(i, j int) bool { return mains[i].Def.Span.Start.Before(mains[j].Def.Span.Start) })
	if len(mains) == 0 {
		c.errorf(diagnostics.EntryMissing, prog.Entry.AST.Package.Span, "no entry point: function 'main' is not defined").
			WithHelp("define `protected int main() { return 0; }`")
		return
	}
	for _, m := range mains[1:] {
		c.errorf(diagnostics.EntryDuplicate, m.Def.Name.Span, "more than one entry point: 'main' is also defined in package %s at %s",
			mains[0].Pkg.Name, mains[0].Def.Name.Span.Start.Where(m.File.Source)).
			WithHelp("a program has exactly one function named 'main'; rename the others")
	}
	// The entry point is exactly `protected int main()` (owner decision,
	// OQ-ENTRY-01): its int result is the exit code, and command-line
	// arguments come from std:process.
	main := mains[0]
	t := main.Type
	if t.Result.Kind == types.Invalid {
		return
	}
	var problem, help string
	switch {
	case len(t.Params) != 0:
		problem, help = "takes no parameters", "read command-line arguments with process.args() from std:process"
	case t.Result.Kind != types.Signed || t.Result.Bits != 0:
		problem, help = "returns 'int'", "return the exit code, e.g. `return 0;` for success"
	case main.visibility() != ast.VisProtected:
		problem, help = "is protected", fmt.Sprintf("write `protected int main()` instead of `%s`", main.visibility())
	}
	if problem != "" {
		c.errorf(diagnostics.EntryBadSignature, main.Def.Name.Span, "the entry point must be declared `protected int main()`: it %s", problem).
			WithHelp(help)
		return
	}
	c.info.Entry = main.Def
}
