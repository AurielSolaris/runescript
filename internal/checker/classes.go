package checker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/types"
)

// Classes (OQ-CLASS-01, owner decisions 2026-11-01) are reference types:
// assigning, passing, or storing an object shares it, and == compares
// identity. A class is declared at the top level with an access modifier;
// its members are fields, methods, and at most one constructor, each with
// its own modifier, meaning what it means at the top level: public is
// everywhere, protected is this package, private is this file. Objects
// are made with `new Name(args)`. Inside methods and the constructor the
// object is `this`, which is required to reach its fields and methods.
// There is no inheritance, no static member, and no nested class yet, and
// no null: a variable of a class type must be initialized.

// MemberSet holds the members of a declared struct or class.
type MemberSet struct {
	Type        *types.Type
	Name        *ast.Ident
	Span        ast.Span
	Visibility  ast.Visibility
	MethodDecls []*ast.FuncDecl
	CtorDecls   []*ast.FuncDecl
	File        *loader.File
	Pkg         *loader.Package
	// Fields are the field declarations, in the order of Type.Fields.
	Fields     []*ast.MemberField
	FieldDecls []*ast.MemberField
	// Methods are the methods by name. Each one's Type leaves out `this`.
	Methods map[string]*FuncSym
	// Ctor is the constructor; its Def is nil for the implicit one a
	// class without a constructor has, which takes no arguments and has
	// the class's own access modifier (CtorVis).
	Ctor    *FuncSym
	CtorVis ast.Visibility
	// CtorPrefix is the number of statements the constructor's body
	// starts with that only set fields, `this.f = value;` with no other
	// use of `this`. They run before the object exists: a field whose
	// type has no zero value may be given its first value there.
	CtorPrefix int
}

// MethodCall is a call x.name(args) on an object: of a method, or of a
// field holding a function (Method nil, Field its index).
type MethodCall struct {
	Recv   ast.Expr
	Method *FuncSym
	Field  int
}

// declareClass declares d's name like a struct's.
func (c *checker) declareClass(d *ast.ClassDecl) {
	name := d.Name.Name
	if types.Lookup(name) != nil || name == "Element" || name == "Signal" || name == "Array" {
		c.errorf(diagnostics.NameRedeclared, d.Name.Span, "'%s' is a built-in type and cannot be redeclared", name)
		return
	}
	s := c.pkgs[c.file.Package]
	if d.Visibility == ast.VisPrivate {
		s = c.files[c.file]
	}
	if c.pkgs[c.file.Package].syms[name] != nil || c.files[c.file].syms[name] != nil {
		c.errorf(diagnostics.NameRedeclared, d.Name.Span, "'%s' is already declared in package %s", name, c.file.Package.Name)
		return
	}
	cs := &MemberSet{Type: types.NewClass(name), Name: d.Name, Span: d.Span, Visibility: d.Visibility, MethodDecls: d.Methods, CtorDecls: d.Ctors, File: c.file, Pkg: c.file.Package, Methods: map[string]*FuncSym{}}
	c.info.Members[cs.Type] = cs
	s.add(name, &symbol{cl: cs, typ: cs.Type})
	c.aggregates = append(c.aggregates, cs)
	cs.FieldDecls = d.Fields
}

func (c *checker) classFields(cs *MemberSet) {
	seen := map[string]bool{}

	for _, fd := range cs.FieldDecls {
		t := c.resolveType(fd.Type)
		switch {
		case t.Kind == types.Void:
			c.errorf(diagnostics.TypeVoidValue, fd.Type.Span, "field '%s' cannot have type 'void'", fd.Name.Name)
			t = types.InvalidT
		case t.Kind != types.Invalid && !types.IsElemType(t):
			c.errorf(diagnostics.TypeMismatch, fd.Type.Span, "a class cannot hold a value of type '%s'", t)
			t = types.InvalidT
		}
		if seen[fd.Name.Name] {
			c.errorf(diagnostics.NameRedeclared, fd.Name.Span, "class %s already has a field %q", cs.Type, fd.Name.Name)
			continue
		}
		seen[fd.Name.Name] = true
		cs.Type.Fields = append(cs.Type.Fields, &types.Field{Name: fd.Name.Name, Type: t})
		cs.Fields = append(cs.Fields, fd)
	}
}

// aggregateMembers resolves method and constructor signatures after fields.
func (c *checker) aggregateMembers() {
	for _, cs := range c.aggregates {
		c.enterFile(cs.File)
		if cs.Type.Kind == types.Class {
			c.classFields(cs)
		}
		names := map[string]diagnostics.Span{}
		for _, fd := range cs.Fields {
			names[fd.Name.Name] = fd.Name.Span
		}
		for _, md := range cs.MethodDecls {
			t := c.funcType(md)
			if md.Name.Name == cs.Type.Name {
				c.errorf(diagnostics.NameRedeclared, md.Name.Span, "a method cannot have its type's name").
					WithHelp(fmt.Sprintf("the constructor has no result type: `%s %s(...) { ... }`", md.Visibility, cs.Type.Name))
				continue
			}
			if prev, ok := names[md.Name.Name]; ok {
				c.errorf(diagnostics.NameRedeclared, md.Name.Span, "%s %s already has a member '%s', at %s", aggregateKind(cs.Type), cs.Type, md.Name.Name, prev.Start.Where(c.file.Source)).
					WithHelp("fields and methods share one set of names; methods cannot be overloaded")
				continue
			}
			names[md.Name.Name] = md.Name.Span
			cs.Methods[md.Name.Name] = &FuncSym{Name: md.Name.Name, Type: t, Def: md, File: cs.File, Pkg: cs.Pkg}
		}
		for _, extra := range cs.CtorDecls[min(1, len(cs.CtorDecls)):] {
			c.errorf(diagnostics.NameRedeclared, extra.Name.Span, "%s %s already has a constructor, at %s", aggregateKind(cs.Type), cs.Type, cs.CtorDecls[0].Name.Span.Start.Where(c.file.Source)).
				WithHelp("a type has at most one constructor; constructors cannot be overloaded")
		}
		if len(cs.CtorDecls) == 0 {
			if cs.Type.Kind == types.Struct {
				continue
			}
			cs.Ctor = &FuncSym{Name: cs.Type.Name, Type: types.NewFunc(nil, cs.Type), File: cs.File, Pkg: cs.Pkg}
			cs.CtorVis = cs.Visibility
			continue
		}
		ctor := cs.CtorDecls[0]
		t := c.funcType(ctor)
		cs.Ctor = &FuncSym{Name: cs.Type.Name, Type: types.NewFunc(t.Params, cs.Type), Def: ctor, File: cs.File, Pkg: cs.Pkg}
		cs.CtorVis = ctor.Visibility
	}
	for _, cs := range c.aggregates {
		c.enterFile(cs.File)
		if cs.Ctor != nil {
			c.firstValues(cs)
		}
	}
}

// firstValues finds the constructor's prefix that sets fields before the
// object is used (MemberSet.CtorPrefix), and reports every field that has
// no value then: no initializer, no zero value, and not set there. There
// is no null, so an object never has a field without a value.
func (c *checker) firstValues(cs *MemberSet) {
	set := map[string]bool{}
	if def := cs.Ctor.Def; def != nil {
		for _, st := range def.Body.Stmts {
			s, ok := st.(*ast.SetStmt)
			if !ok || s.Op != "=" || usesThis(s.Value) {
				break
			}
			sel, ok := unparen(s.Target).(*ast.SelectorExpr)
			if !ok {
				break
			}
			if _, ok := unparen(sel.X).(*ast.ThisExpr); !ok {
				break
			}
			set[sel.Name.Name] = true
			cs.CtorPrefix++
		}
	}
	for i, fd := range cs.Fields {
		t := cs.Type.Fields[i].Type
		if fd.Init != nil || set[fd.Name.Name] || t.Kind == types.Invalid || types.HasZero(t) {
			continue
		}
		c.errorf(diagnostics.TypeNoZeroValue, fd.Name.Span, "field '%s' of type '%s' has no zero value: give it an initial value", fd.Name.Name, t).
			WithHelp(fmt.Sprintf("write `%s %s %s = ...;`, or set it at the start of the constructor, before any other use of this: `this.%s = ...;`",
				fd.Visibility, fd.Type, fd.Name.Name, fd.Name.Name))
	}
}

// usesThis reports whether e mentions `this`. Lambdas cannot (they do not
// capture it), so they are not searched.
func usesThis(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.ThisExpr:
		return true
	case *ast.ParenExpr:
		return usesThis(e.X)
	case *ast.UnaryExpr:
		return usesThis(e.X)
	case *ast.BinaryExpr:
		return usesThis(e.X) || usesThis(e.Y)
	case *ast.CastExpr:
		return usesThis(e.X)
	case *ast.SelectorExpr:
		return usesThis(e.X)
	case *ast.IndexExpr:
		return usesThis(e.X) || usesThis(e.Index)
	case *ast.CallExpr:
		return (e.Callee != nil && usesThis(e.Callee)) || anyUsesThis(e.Args)
	case *ast.NewExpr:
		return anyUsesThis(e.Args)
	case *ast.ArrayLit:
		return anyUsesThis(e.Elems)
	case *ast.InterpString:
		return anyUsesThis(e.Exprs)
	case *ast.StructLit:
		for _, f := range e.Fields {
			if usesThis(f.Value) {
				return true
			}
		}
	}
	return false
}

func anyUsesThis(list []ast.Expr) bool {
	for _, e := range list {
		if usesThis(e) {
			return true
		}
	}
	return false
}

// classBodies checks the field initializers, the methods, and the
// constructor of the current file's classes.
func (c *checker) aggregateBodies() {
	for _, cs := range c.aggregates {
		if cs.File != c.file {
			continue
		}
		for i, fd := range cs.Fields {
			if fd.Init == nil {
				continue
			}
			restore := c.saveFuncState()
			c.fnName = fmt.Sprintf("the initial value of field '%s'", fd.Name.Name)
			c.result = types.VoidT
			c.push(true)
			if ft := cs.Type.Fields[i].Type; ft.Kind != types.Invalid {
				c.assign(ft, fd.Init, "assign")
			} else {
				c.expr(fd.Init)
			}
			c.pop()
			restore()
		}
		for _, md := range cs.MethodDecls {
			if m := cs.Methods[md.Name.Name]; m != nil && m.Def == md {
				c.memberBody(md, m.Type, cs, fmt.Sprintf("method '%s.%s'", cs.Type, md.Name.Name))
			}
		}
		if cs.Ctor != nil && cs.Ctor.Def != nil {
			c.memberBody(cs.Ctor.Def, types.NewFunc(cs.Ctor.Type.Params, types.VoidT), cs, fmt.Sprintf("the constructor of %s", cs.Type))
		}
	}
}

// memberBody checks the body of a method or constructor of cs, in which
// `this` is the object.
func (c *checker) memberBody(fn *ast.FuncDecl, t *types.Type, cs *MemberSet, what string) {
	defer c.saveFuncState()()
	c.fnName = what
	c.result = t.Result
	c.this = cs
	c.thisWritable = cs.Type.Kind == types.Struct && cs.Ctor != nil && cs.Ctor.Def == fn
	c.push(true)
	for i, p := range fn.Params {
		c.declareVar(p.Name, t.Params[i], true)
	}
	c.stmts(fn.Body.Stmts)
	c.pop()
	c.checkReturns(fn.Body)
}

// thisExpr checks `this`.
func (c *checker) thisExpr(e *ast.ThisExpr) *types.Type {
	switch {
	case c.this != nil:
		return c.this.Type
	case c.thisHidden:
		c.errorf(diagnostics.NameAcrossBoundary, e.Span, "'this' is not visible in a nested function or lambda").
			WithHelp("nested functions and lambdas do not capture; pass the object as a parameter")
	default:
		c.errorf(diagnostics.NameUndefined, e.Span, "'this' is only available in the methods and the constructor of a class or struct")
	}
	return types.InvalidT
}

// memberVisible reports, at span, a member of cs that the current file
// may not use, and reports whether it may.
func (c *checker) memberVisible(cs *MemberSet, vis ast.Visibility, what, name string, span diagnostics.Span) bool {
	switch {
	case vis == ast.VisPrivate && c.file != cs.File:
		c.errorf(diagnostics.NameNotVisible, span, "%s '%s' of %s %s is private to %s", what, name, aggregateKind(cs.Type), cs.Type, cs.File.Source.Name).
			WithHelp(fmt.Sprintf("declare it protected to share it within package %s, or public to share it with every package", cs.Pkg.Name))
		return false
	case vis == ast.VisProtected && c.file.Package != cs.Pkg:
		c.errorf(diagnostics.NameNotVisible, span, "%s '%s' of %s %s is protected: only package %s can use it", what, name, aggregateKind(cs.Type), cs.Type, cs.Pkg.Name).
			WithHelp("declare it public to use it from other packages")
		return false
	}
	return true
}

// classField checks x.name on an object of class t: a field.
func (c *checker) classField(e *ast.SelectorExpr, t *types.Type) *types.Type {
	cs := c.info.Members[t]
	i := types.FieldIndex(t, e.Name.Name)
	if i < 0 {
		if cs.Methods[e.Name.Name] != nil {
			c.errorf(diagnostics.NameNotAValue, e.Name.Span, "method '%s' of class %s can only be called", e.Name.Name, t).
				WithHelp(fmt.Sprintf("call it: %s(...); to pass it on, wrap the call in a lambda", e.Name.Name))
			return types.InvalidT
		}
		c.errorf(diagnostics.TypeNoField, e.Name.Span, "class %s has no field '%s'", t, e.Name.Name).
			WithHelp(memberHelp(cs))
		return types.InvalidT
	}
	c.memberVisible(cs, cs.Fields[i].Visibility, "field", e.Name.Name, e.Name.Span)
	return t.Fields[i].Type
}

// methodCall checks recv.name(args) on an object of class t: a method, or
// a field holding a function.
func (c *checker) methodCall(e *ast.CallExpr, recv ast.Expr, name *ast.Ident, t *types.Type) *types.Type {
	cs := c.info.Members[t]
	if m := cs.Methods[name.Name]; m != nil {
		c.memberVisible(cs, m.Def.Visibility, "method", name.Name, name.Span)
		c.info.MethodCalls[e] = &MethodCall{Recv: recv, Method: m}
		return c.checkArgs(m.Type, e, fmt.Sprintf("method '%s.%s'", t, name.Name))
	}
	if i := types.FieldIndex(t, name.Name); i >= 0 {
		c.memberVisible(cs, cs.Fields[i].Visibility, "field", name.Name, name.Span)
		ft := t.Fields[i].Type
		if ft.Kind == types.Func {
			c.info.MethodCalls[e] = &MethodCall{Recv: recv, Field: i}
			return c.checkArgs(ft, e, fmt.Sprintf("field '%s.%s'", t, name.Name))
		}
		if ft.Kind != types.Invalid {
			c.errorf(diagnostics.TypeNotCallable, name.Span, "field '%s' of %s %s has type '%s', not a function type", name.Name, aggregateKind(t), t, ft)
		}
		c.exprs(e.Args)
		return types.InvalidT
	}
	c.errorf(diagnostics.TypeNoField, name.Span, "%s %s has no method '%s'", aggregateKind(t), t, name.Name).WithHelp(memberHelp(cs))
	c.exprs(e.Args)
	return types.InvalidT
}

// newObject checks `new Name(args)` for class t.
func (c *checker) newObject(e *ast.NewExpr, t *types.Type) *types.Type {
	cs := c.info.Members[t]
	if cs.Ctor.Def != nil { // the implicit one is as visible as the class, already checked
		c.memberVisible(cs, cs.CtorVis, "constructor", t.Name, e.Type.Span)
	}
	params := cs.Ctor.Type.Params
	if len(e.Args) != len(params) {
		c.errorf(diagnostics.TypeArgCount, e.Span, "the constructor of %s expects %d argument(s), found %d", t, len(params), len(e.Args))
		c.exprs(e.Args)
		return t
	}
	for i, a := range e.Args {
		if params[i].Kind == types.Invalid {
			c.expr(a)
			continue
		}
		c.assign(params[i], a, "pass")
	}
	return t
}

// memberHelp lists the members of cs that diagnostics can suggest.
func memberHelp(cs *MemberSet) string {
	var fields, methods []string
	for _, f := range cs.Type.Fields {
		fields = append(fields, f.Name)
	}
	for name := range cs.Methods {
		methods = append(methods, name+"()")
	}
	sort.Strings(methods)
	switch {
	case len(fields) == 0 && len(methods) == 0:
		return "it has no fields or methods"
	case len(methods) == 0:
		return "its fields are " + strings.Join(fields, ", ")
	case len(fields) == 0:
		return "its methods are " + strings.Join(methods, ", ")
	}
	return "its fields are " + strings.Join(fields, ", ") + "; its methods are " + strings.Join(methods, ", ")
}

// peekType returns the type of x, a chain of names, fields, and elements
// such as a.b[i].c, without checking it, or nil if that is not simple.
// place uses it to tell whether the object holding a field is a struct,
// which must itself be assignable, or an object, which is a reference.
func (c *checker) peekType(x ast.Expr) *types.Type {
	switch x := unparen(x).(type) {
	case *ast.Ident:
		if sym, _ := c.lookup(x.Name); sym != nil && sym.fn == nil && sym.ext == nil && sym.st == nil && sym.cl == nil {
			return sym.typ
		}
	case *ast.ThisExpr:
		if c.this != nil {
			return c.this.Type
		}
	case *ast.IndexExpr:
		if t := c.peekType(x.X); t != nil && t.Kind == types.Array {
			return t.Elem
		}
	case *ast.SelectorExpr:
		if t := c.peekType(x.X); t != nil && (t.Kind == types.Struct || t.Kind == types.Class) {
			if i := types.FieldIndex(t, x.Name.Name); i >= 0 {
				return t.Fields[i].Type
			}
		}
	}
	return nil
}

func aggregateKind(t *types.Type) string {
	if t.Kind == types.Struct {
		return "struct"
	}
	return "class"
}
