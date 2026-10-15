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
	"math/big"
	"sort"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// Info is the result of checking, consumed by the runtime.
type Info struct {
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
	// Entry is the program entry point; EntryVoid reports whether it
	// returns void.
	Entry     *ast.FuncDecl
	EntryVoid bool
}

// FuncSym is a named function: top-level or nested, possibly declared by a
// prototype before its definition.
type FuncSym struct {
	Name  string
	Type  *types.Type
	Def   *ast.FuncDecl // nil until the definition is seen
	Proto *ast.FuncDecl // the prototype, if any
}

// symbol is anything a name can refer to inside a function.
type symbol struct {
	typ     *types.Type // variable type
	isParam bool
	fn      *FuncSym // non-nil for named functions
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
	info   *Info
	diags  []*diagnostics.Diagnostic
	scopes []*scope
	uses   map[string]*stdlib.Package // by qualifier

	// state of the function or lambda being checked
	fnName    string
	result    *types.Type // nil while inferring a lambda's result type
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
	c := &checker{
		info: &Info{
			Types:    map[ast.Expr]*types.Type{},
			Consts:   map[ast.Expr]int64{},
			Callees:  map[*ast.CallExpr]*FuncSym{},
			FuncRefs: map[*ast.Ident]*FuncSym{},
			Natives:  map[*ast.CallExpr]*stdlib.Func{},
		},
		uses:   map[string]*stdlib.Package{},
		scopes: []*scope{{syms: map[string]*symbol{}}},
	}
	c.useDecls(file)

	// Top-level functions are visible throughout the file, so declare them
	// all before checking any body.
	sigs := make([]*types.Type, len(file.Funcs))
	for i, fn := range file.Funcs {
		sigs[i] = c.declareFunc(fn)
	}
	for i, fn := range file.Funcs {
		if fn.Body != nil {
			c.funcBody(fn, sigs[i])
		}
	}
	c.reportUndefined(c.scopes[0])
	c.entry(file)
	return c.info, c.diags
}

func (c *checker) errorf(code diagnostics.Code, span diagnostics.Span, format string, args ...any) *diagnostics.Diagnostic {
	d := diagnostics.New(code, span, format, args...)
	c.diags = append(c.diags, d)
	return d
}

// ---- use declarations ----

// useDecls resolves `use` declarations. Only "std:" packages exist so far;
// other packages, files, and "go:" packages are planned.
func (c *checker) useDecls(file *ast.File) {
	for _, u := range file.Uses {
		name, isStd := strings.CutPrefix(u.Path, "std:")
		if !isStd {
			c.errorf(diagnostics.UseUnsupported, u.PathSpan, "cannot use %q: only standard library packages (\"std:...\") are supported so far", u.Path).
				WithHelp("using other RuneScript packages, source files, and Go packages is planned")
			continue
		}
		pkg := stdlib.Lookup(name)
		if pkg == nil {
			paths := stdlib.Paths()
			sort.Strings(paths)
			c.errorf(diagnostics.UseUnknownPackage, u.PathSpan, "unknown standard library package %q", u.Path).
				WithHelp("available: std:" + strings.Join(paths, ", std:"))
			continue
		}
		if _, dup := c.uses[pkg.Qualifier]; dup {
			c.errorf(diagnostics.UseDuplicate, u.Span, "%q is already used", u.Path)
			continue
		}
		c.uses[pkg.Qualifier] = pkg
	}
}

// ---- scopes ----

func (c *checker) push(boundary bool) {
	c.scopes = append(c.scopes, &scope{syms: map[string]*symbol{}, boundary: boundary})
}

func (c *checker) pop() {
	c.reportUndefined(c.scopes[len(c.scopes)-1])
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
	case prev == nil, crossed && prev.fn == nil:
		return false
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
	c.conflict(id)
	c.top().add(id.Name, &symbol{typ: t, isParam: isParam})
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
	t := types.Lookup(ref.Name)
	if t == nil {
		c.errorf(diagnostics.TypeUnknown, ref.Span, "unknown type '%s'", ref.Name)
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

// declareFunc declares a top-level or nested function (definition or
// prototype) in the innermost scope and returns its signature.
func (c *checker) declareFunc(fn *ast.FuncDecl) *types.Type {
	t := c.funcType(fn)
	name := fn.Name.Name
	if existing, ok := c.top().syms[name]; ok {
		f := existing.fn
		switch {
		case f != nil && f.Def == nil && fn.Body != nil:
			if !types.Identical(f.Type, t) {
				c.errorf(diagnostics.TypeProtoMismatch, fn.Name.Span,
					"definition of '%s' has type '%s', but its prototype at %s declares '%s'",
					name, t, f.Proto.Name.Span.Start, f.Type)
			}
			f.Def = fn
		case f != nil:
			c.errorf(diagnostics.NameRedeclared, fn.Name.Span, "function '%s' is already declared at %s", name, firstDecl(f).Name.Span.Start).
				WithHelp("functions cannot be overloaded; a prototype may appear once, before the definition")
		default:
			c.errorf(diagnostics.NameRedeclared, fn.Name.Span, "'%s' is already declared in this function", name)
		}
		return t
	}
	if c.conflict(fn.Name) {
		return t
	}
	f := &FuncSym{Name: name, Type: t}
	if fn.Body != nil {
		f.Def = fn
	} else {
		f.Proto = fn
	}
	c.top().add(name, &symbol{fn: f})
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
	c.stmts(fn.Body.Stmts)
	c.pop()
	c.checkReturns(fn.Body)
}

// saveFuncState saves the per-function state and returns a function that
// restores it, for checking nested functions and lambdas.
func (c *checker) saveFuncState() func() {
	name, result, loops := c.fnName, c.result, c.loopDepth
	c.loopDepth = 0
	return func() { c.fnName, c.result, c.loopDepth = name, result, loops }
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

// terminates reports whether stmts always end in a return. Loops are never
// considered terminating (OQ-FLOW-01).
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
	sym := c.scopes[0].syms["main"]
	if sym == nil || sym.fn == nil || sym.fn.Def == nil {
		c.errorf(diagnostics.EntryMissing, file.Package.Span, "no entry point: function 'main' is not defined").
			WithHelp("define `public int main() { return 0; }`")
		return
	}
	t := sym.fn.Type
	if len(t.Params) != 0 || (t.Result.Kind != types.Signed && t.Result.Kind != types.Void) {
		if t.Result.Kind != types.Invalid {
			c.errorf(diagnostics.EntryBadSignature, sym.fn.Def.Name.Span, "'main' must take no parameters and return 'int' or 'void'").
				WithHelp("entry-point parameters (e.g. command-line arguments) are not yet specified")
		}
		return
	}
	c.info.Entry = sym.fn.Def
	c.info.EntryVoid = t.Result.Kind == types.Void
}
