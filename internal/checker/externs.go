package checker

import (
	"fmt"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/jshost"
	"github.com/AurielSolaris/runescript/internal/types"
)

// ExternSym is a function implemented in JavaScript (an extern
// declaration). Values crossing the boundary are checked at run time.
type ExternSym struct {
	Name   string
	Module string
	Type   *types.Type
	Decl   *ast.ExternDecl
	// Host is the host-neutral function for a rune: module; nil for any
	// other module.
	Host *jshost.Func
}

// declareExterns declares the current file's extern functions in its file
// scope. Every top-level function is already declared, so a clash with
// one is found whichever comes first in the source.
func (c *checker) declareExterns(f *ast.File) {
	fileScope, pkgScope := c.files[c.file], c.pkgs[c.file.Package]
	for _, d := range f.Externs {
		t := c.externType(d)
		name := d.Name.Name
		switch prev := fileScope.syms[name]; {
		case name == "trace":
			c.errorf(diagnostics.NameBuiltinReused, d.Name.Span, "'trace' is a built-in function and cannot be redeclared")
			continue
		case prev != nil && prev.ext != nil:
			c.errorf(diagnostics.NameRedeclared, d.Name.Span, "extern function '%s' is already declared at %s", name, prev.ext.Decl.Name.Span.Start).
				WithHelp("declare each extern function once per file")
			continue
		case prev != nil:
			c.errorf(diagnostics.NameRedeclared, d.Name.Span, "'%s' is already declared as a function at %s", name, firstDecl(prev.fn).Name.Span.Start.Where(c.file.Source)).
				WithHelp("functions cannot be overloaded; rename one of them")
			continue
		case pkgScope.syms[name] != nil:
			fn := pkgScope.syms[name].fn
			c.errorf(diagnostics.NameRedeclared, d.Name.Span, "'%s' is already declared as a function of package %s at %s", name, c.file.Package.Name, firstDecl(fn).Name.Span.Start.Where(c.file.Source)).
				WithHelp("functions cannot be overloaded; rename one of them")
			continue
		}
		sym := &ExternSym{Name: name, Module: d.Module, Type: t, Decl: d}
		c.externModule(sym)
		fileScope.add(name, &symbol{typ: t, ext: sym})
		c.info.ExternDecls = append(c.info.ExternDecls, sym)
	}
}

// externType resolves an extern declaration's signature. Only values
// that can be checked when they cross from JavaScript are allowed:
// numbers, bool, string, and errors (and void as the result).
func (c *checker) externType(d *ast.ExternDecl) *types.Type {
	params := make([]*types.Type, len(d.Params))
	for i, p := range d.Params {
		params[i] = c.boundaryType(p.Type, false)
	}
	return types.NewFunc(params, c.boundaryType(d.Result, true))
}

func (c *checker) boundaryType(ref *ast.TypeRef, result bool) *types.Type {
	t := c.resolveType(ref)
	switch {
	case (t.Kind == types.Signed || t.Kind == types.Unsigned) && t.Bits != 0:
		// The explicit widths, and 64-bit BigInts in particular, have no
		// checked JavaScript representation yet (OQ-NUM-07).
		c.errorf(diagnostics.TypeExternType, ref.Span, "type '%s' cannot cross the JavaScript boundary yet", t).
			WithHelp("use int or uint at the boundary and convert, e.g. (int16) n")
		return types.InvalidT
	}
	switch t.Kind {
	case types.Invalid, types.Signed, types.Unsigned, types.Float, types.Bool, types.String, types.Error:
		return t
	case types.Void:
		if result {
			return t
		}
		c.errorf(diagnostics.TypeVoidValue, ref.Span, "extern function parameters cannot be 'void'")
	default:
		c.errorf(diagnostics.TypeExternType, ref.Span, "type '%s' cannot cross the JavaScript boundary", t).
			WithHelp("extern functions take and return int, uint, float, float32, float64, bool, string, and error types")
	}
	return types.InvalidT
}

// externModule checks sym's module: a rune: function must exist and match
// its declaration exactly; a host's own module is allowed with a warning.
func (c *checker) externModule(sym *ExternSym) {
	d := sym.Decl
	switch {
	case strings.HasPrefix(d.Module, jshost.Prefix):
		m := jshost.Lookup(d.Module)
		if m == nil {
			c.errorf(diagnostics.UseExternUnknown, d.ModuleSpan, "unknown host-neutral module %q", d.Module).
				WithHelp("available: " + strings.Join(jshost.Names(), ", "))
			return
		}
		f := m.Funcs[d.Name.Name]
		if f == nil {
			c.errorf(diagnostics.UseExternUnknown, d.Name.Span, "module %q has no function '%s'", d.Module, d.Name.Name).
				WithHelp("its functions are " + strings.Join(m.FuncNames(), ", "))
			return
		}
		if !hasInvalid(sym.Type) && !types.Identical(sym.Type, f.Type()) {
			c.errorf(diagnostics.TypeExternMismatch, d.Name.Span, "'%s' of %q has type '%s', but this declaration says '%s'", f.Name, d.Module, f.Type(), sym.Type).
				WithHelp(fmt.Sprintf("declare it as: extern %q %s;", d.Module, signature(f)))
			return
		}
		sym.Host = f
	case jshost.HostSpecific(d.Module):
		c.warnf(diagnostics.WarnHostModule, d.ModuleSpan, "%q ties the program to one JavaScript host", d.Module).
			WithHelp("use a host-neutral module where one exists (" + strings.Join(jshost.Names(), ", ") + "), or allow this warning with --allow " + diagnostics.WarnHostModule.String())
	}
}

// signature renders f as it is declared, e.g. "string readFile(string path)".
func signature(f *jshost.Func) string {
	params := make([]string, len(f.Params))
	for i, p := range f.Params {
		params[i] = p.String()
	}
	return fmt.Sprintf("%s %s(%s)", f.Result, f.Name, strings.Join(params, ", "))
}

// externValueError reports an extern function used other than by calling it.
func (c *checker) externValueError(id *ast.Ident) {
	c.errorf(diagnostics.NameNotAValue, id.Span, "extern function '%s' can only be called", id.Name).
		WithHelp("wrap it in a lambda to use it as a value")
}
