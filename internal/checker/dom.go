package checker

import (
	"fmt"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/dom"
	"github.com/AurielSolaris/runescript/internal/types"
)

// rune:dom (OQ-JS-09, OQ-DOM-01). `use "rune:dom";` makes the types
// Element and Signal<T> and the function signal(...) available without a
// qualifier, and dom.renderApp through it. Programs that use rune:dom run
// only on the JavaScript target.

// SignalOp is what a call involving a rune does.
type SignalOp struct {
	Op string // "new", "get", or "set"
	// T is the rune's type, Signal<T>.
	T *types.Type
}

// runeModules are the rune: modules a use declaration can name. rune:fs
// is reached through extern declarations instead (internal/jshost).
var runeModules = map[string]bool{"dom": true}

// isRune reports whether sym is a variable or parameter holding a rune.
// Nested functions and lambdas may use the runes of the functions around
// them (owner decision, OQ-DOM-01); other variables stay invisible
// (OQ-FN-03).
func isRune(sym *symbol) bool {
	return sym.fn == nil && sym.ext == nil && sym.typ != nil && sym.typ.Kind == types.Signal
}

// domType resolves Element and Signal<T>, which need `use "rune:dom";`.
func (c *checker) domType(ref *ast.TypeRef) *types.Type {
	if !c.dom {
		c.errorf(diagnostics.TypeUnknown, ref.Span, "type '%s' needs rune:dom", ref.Name).
			WithHelp(`add use "rune:dom"; after the package declaration`)
		return types.InvalidT
	}
	c.domUsed[c.file] = true
	if ref.Name == "Element" {
		if len(ref.Args) > 0 {
			c.errorf(diagnostics.TypeArgs, ref.Span, "type 'Element' takes no type arguments")
			return types.InvalidT
		}
		return types.ElementT
	}
	if len(ref.Args) != 1 {
		c.errorf(diagnostics.TypeArgs, ref.Span, "'Signal' takes one type argument, the type of its value, e.g. Signal<int>")
		return types.InvalidT
	}
	elem := c.resolveType(ref.Args[0])
	if elem.Kind != types.Invalid && !types.IsScalar(elem) {
		c.errorf(diagnostics.TypeArgs, ref.Args[0].Span, "a rune holds a number, bool, or string, not '%s'", elem).
			WithHelp("runes of other types need generics, which are not designed yet")
		return types.InvalidT
	}
	if elem.Kind == types.Invalid {
		return types.InvalidT
	}
	return types.NewSignal(elem)
}

// isSignalBuiltin reports whether call is rune:dom's signal(...).
func (c *checker) isSignalBuiltin(call *ast.CallExpr) bool {
	if call.Pkg != nil || call.Fn == nil || call.Fn.Name != "signal" || !c.dom {
		return false
	}
	sym, _ := c.lookup("signal")
	return sym == nil
}

// signalNew checks signal(initial). With a destination type Signal<T>,
// initial must be assignable to T; without one, T is initial's type.
func (c *checker) signalNew(e *ast.CallExpr, elem *types.Type) *types.Type {
	if len(e.Args) != 1 {
		c.errorf(diagnostics.TypeArgCount, e.Span, "signal expects 1 argument, the initial value, found %d", len(e.Args))
		c.exprs(e.Args)
		return types.InvalidT
	}
	if elem == nil {
		elem = c.defaulted(e.Args[0]).typ
		switch {
		case elem.Kind == types.Invalid:
			return types.InvalidT
		case !types.IsScalar(elem):
			c.errorf(diagnostics.TypeArgs, e.Args[0].NodeSpan(), "a rune holds a number, bool, or string, not '%s'", elem)
			return types.InvalidT
		}
	} else {
		c.assign(elem, e.Args[0], "assign")
	}
	t := types.NewSignal(elem)
	c.info.Signals[e] = SignalOp{Op: "new", T: t}
	return t
}

// signalCall checks count(), which reads the rune count.
func (c *checker) signalCall(e *ast.CallExpr, t *types.Type) *types.Type {
	if len(e.Args) != 0 {
		c.errorf(diagnostics.TypeArgCount, e.Span, "reading a rune takes no arguments, found %d", len(e.Args)).
			WithHelp(fmt.Sprintf("to change it, write %s.set(value)", e.Fn.Name))
		c.exprs(e.Args)
	}
	c.info.Signals[e] = SignalOp{Op: "get", T: t}
	return t.Elem
}

// signalMethod checks count.get() and count.set(value).
func (c *checker) signalMethod(e *ast.CallExpr, t *types.Type) *types.Type {
	switch e.Fn.Name {
	case "get":
		if len(e.Args) != 0 {
			c.errorf(diagnostics.TypeArgCount, e.Span, "get takes no arguments, found %d", len(e.Args))
			c.exprs(e.Args)
		}
		c.info.Signals[e] = SignalOp{Op: "get", T: t}
		return t.Elem
	case "set":
		if len(e.Args) != 1 {
			c.errorf(diagnostics.TypeArgCount, e.Span, "set expects 1 argument, the new value, found %d", len(e.Args))
			c.exprs(e.Args)
			return types.VoidT
		}
		c.assign(t.Elem, e.Args[0], "assign")
		c.info.Signals[e] = SignalOp{Op: "set", T: t}
		return types.VoidT
	}
	c.errorf(diagnostics.TypeNoField, e.Fn.Span, "a rune has no method '%s'", e.Fn.Name).
		WithHelp("read a rune with count() or count.get(), and change it with count.set(value)")
	c.exprs(e.Args)
	return types.InvalidT
}

// domCall checks a call of a rune:dom function through its qualifier.
func (c *checker) domCall(e *ast.CallExpr) *types.Type {
	if e.Fn.Name != "renderApp" {
		c.errorf(diagnostics.NameNotInPackage, e.Fn.Span, "rune:dom has no function '%s'", e.Fn.Name).
			WithHelp("rune:dom provides renderApp(component, selector); Element, Signal, and signal are used without a qualifier")
		c.exprs(e.Args)
		return types.InvalidT
	}
	component := types.NewFunc(nil, types.ElementT)
	t := c.checkArgs(types.NewFunc([]*types.Type{component, types.StringT}, types.VoidT), e, "function 'dom.renderApp'")
	if len(e.Args) == 2 {
		if lit, ok := unparen(e.Args[1]).(*ast.StringLit); ok && !dom.IsIDSelector(lit.Value) {
			c.warnf(diagnostics.WarnMountSelector, lit.Span, "%q is not an id selector: it may match several elements, or none", lit.Value).
				WithHelp(`mount the app on one element by id, e.g. "#app"; renderApp uses the first match`)
		}
	}
	c.info.DOMCalls[e] = e.Fn.Name
	return t
}

// packageFuncValue resolves pkg.name used as a value (not called), e.g.
// app.App passed to dom.renderApp.
func (c *checker) packageFuncValue(e *ast.SelectorExpr, b *binding) *types.Type {
	q := e.X.(*ast.Ident).Name
	if b.pkg == nil {
		c.errorf(diagnostics.NameNotAValue, e.Span, "'%s.%s' can only be called; library functions cannot be used as values yet", q, e.Name.Name).
			WithHelp(fmt.Sprintf("wrap it in a lambda, e.g. (int x) => %s.%s(x)", q, e.Name.Name))
		return types.InvalidT
	}
	sym := c.pkgs[b.pkg].syms[e.Name.Name]
	if sym == nil || sym.fn == nil {
		if !c.privateError(b.pkg, e.Name.Name, e.Name.Span) {
			c.errorf(diagnostics.NameNotInPackage, e.Name.Span, "package %s has no function '%s'", b.pkg.Name, e.Name.Name)
		}
		return types.InvalidT
	}
	c.checkProtected(sym.fn, e.Name.Span)
	c.info.QualRefs[e] = sym.fn
	return sym.fn.Type
}

// runeModuleError reports a use of an unknown rune: module.
func (c *checker) runeModuleError(name string, d *ast.UseDecl) {
	help := `the rune: modules are rune:dom; rune:fs is reached through extern declarations, e.g. extern "rune:fs" string readFile(string path);`
	if !strings.Contains(name, "fs") {
		help = "the rune: modules are rune:dom (and rune:fs, through extern declarations)"
	}
	c.errorf(diagnostics.UseUnknownPackage, d.PathSpan, "unknown module %q", d.Path).WithHelp(help)
}
