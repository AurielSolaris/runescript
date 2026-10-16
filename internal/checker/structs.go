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

// Structs (OQ-STRUCT-01) are value types: assigning, passing, returning,
// or storing a struct copies it, so a change through one name is never
// seen through another. They are declared at the top level with an access
// modifier, built with a literal naming every field, and read and written
// field by field. Two structs compare equal when every field does.

// StructSym is a declared struct type.
type StructSym struct {
	Type *types.Type
	Decl *ast.StructDecl
	File *loader.File
	Pkg  *loader.Package
}

// declareStruct declares d's name in the current file's scope (private
// structs) or its package's scope, like a top-level function.
func (c *checker) declareStruct(d *ast.StructDecl) {
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
	st := &StructSym{Type: types.NewStruct(name), Decl: d, File: c.file, Pkg: c.file.Package}
	s.add(name, &symbol{st: st, typ: st.Type})
	c.structs = append(c.structs, st)
}

// structFields resolves the field types of every declared struct, then
// rejects structs that contain themselves.
func (c *checker) structFields() {
	for _, st := range c.structs {
		c.enterFile(st.File)
		seen := map[string]bool{}
		for _, fd := range st.Decl.Fields {
			t := c.resolveType(fd.Type)
			switch {
			case seen[fd.Name.Name]:
				c.errorf(diagnostics.NameRedeclared, fd.Name.Span, "struct %s already has a field '%s'", st.Type, fd.Name.Name)
				continue
			case t.Kind == types.Void:
				c.errorf(diagnostics.TypeVoidValue, fd.Type.Span, "field '%s' cannot have type 'void'", fd.Name.Name)
				t = types.InvalidT
			case t.Kind != types.Invalid && !types.IsElemType(t):
				c.errorf(diagnostics.TypeMismatch, fd.Type.Span, "a struct cannot hold a value of type '%s'", t)
				t = types.InvalidT
			}
			seen[fd.Name.Name] = true
			st.Type.Fields = append(st.Type.Fields, &types.Field{Name: fd.Name.Name, Type: t})
		}
	}
	// A struct that contains itself, directly or through other structs,
	// would be infinitely large. Through an array it is fine: arrays are
	// references.
	reported := map[string]bool{} // each cycle once, at its first struct
	for _, st := range c.structs {
		path := c.cycle(st.Type, st.Type, nil)
		if path == nil || reported[path[0]] {
			continue
		}
		for _, step := range path {
			reported[step] = true
		}
		c.errorf(diagnostics.TypeRecursiveStruct, st.Decl.Name.Span, "struct %s contains itself: %s", st.Type, strings.Join(path, " -> ")).
			WithHelp(fmt.Sprintf("hold the inner value in an array instead, e.g. Array<%s>, which is a reference", st.Type))
	}
}

// cycle returns the field path by which t contains target by value, or nil.
func (c *checker) cycle(t, target *types.Type, path []string) []string {
	if len(path) > 64 {
		return nil
	}
	for _, f := range t.Fields {
		if f.Type.Kind != types.Struct {
			continue
		}
		p := append(append([]string{}, path...), t.Name+"."+f.Name)
		if f.Type == target {
			return p
		}
		if r := c.cycle(f.Type, target, p); r != nil {
			return r
		}
	}
	return nil
}

// structType resolves a struct name as a type: Name, or pkg.Name for a
// struct of another package. It returns nil if ref names no struct.
func (c *checker) structType(ref *ast.TypeRef) *types.Type {
	st := c.findStruct(ref.Pkg, ref.Name, ref.Span)
	if st == nil {
		return nil
	}
	return st.Type
}

// findStruct finds the struct name, qualified by pkg if not nil, and
// checks that it is visible here. It reports an error and returns nil
// when there is no such struct.
func (c *checker) findStruct(pkg *ast.Ident, name string, span diagnostics.Span) *StructSym {
	if pkg == nil {
		sym, _ := c.lookup(name)
		if sym == nil || sym.st == nil {
			if st := c.privateStructElsewhere(c.file.Package, name); st != nil {
				c.errorf(diagnostics.NameNotVisible, span, "struct %s is private to %s", name, st.File.Source.Name).
					WithHelp(fmt.Sprintf("declare it protected to share it within package %s", st.Pkg.Name))
				return &StructSym{Type: types.InvalidT}
			}
			return nil
		}
		return sym.st
	}
	b := c.quals[pkg.Name]
	switch {
	case b == nil && !c.broken:
		c.errorf(diagnostics.NamePackageNotUsed, pkg.Span, "package '%s' is not in scope", pkg.Name)
		return &StructSym{Type: types.InvalidT}
	case b == nil || (b.pkg == nil && b.std == nil && b.rune == ""):
		return &StructSym{Type: types.InvalidT} // a use that could not be loaded (reported)
	case b.pkg == nil:
		c.errorf(diagnostics.NameNotInPackage, span, "%s has no struct %s", pkg.Name, name)
		return &StructSym{Type: types.InvalidT}
	}
	b.used = true
	sym := c.pkgs[b.pkg].syms[name]
	if sym == nil || sym.st == nil {
		if st := c.privateStructElsewhere(b.pkg, name); st != nil {
			c.errorf(diagnostics.NameNotVisible, span, "struct %s is private to %s", name, st.File.Source.Name).
				WithHelp("declare it public to use it from other packages")
		} else {
			c.errorf(diagnostics.NameNotInPackage, span, "package %s has no struct %s", b.pkg.Name, name)
		}
		return &StructSym{Type: types.InvalidT}
	}
	if sym.st.Decl.Visibility == ast.VisProtected && sym.st.Pkg != c.file.Package {
		c.errorf(diagnostics.NameNotVisible, span, "struct %s is protected: only package %s can use it", name, sym.st.Pkg.Name).
			WithHelp("declare it public to use it from other packages")
	}
	return sym.st
}

// privateStructElsewhere returns a private struct named name in another
// file of pkg, or nil.
func (c *checker) privateStructElsewhere(pkg *loader.Package, name string) *StructSym {
	for _, f := range pkg.Files {
		if f == c.file {
			continue
		}
		if sym := c.files[f].syms[name]; sym != nil && sym.st != nil {
			return sym.st
		}
	}
	return nil
}

// structLit checks Name{field: value, ...}: every field exactly once, each
// value assignable to its field's type.
func (c *checker) structLit(e *ast.StructLit) *types.Type {
	st := c.findStruct(e.Pkg, e.Name.Name, e.Name.Span)
	if st == nil {
		c.errorf(diagnostics.TypeUnknown, e.Name.Span, "unknown struct '%s'", e.Name.Name)
	}
	if st == nil || st.Type.Kind == types.Invalid {
		for _, fi := range e.Fields {
			c.expr(fi.Value)
		}
		return types.InvalidT
	}
	t := st.Type
	given := map[string]bool{}
	for _, fi := range e.Fields {
		i := types.FieldIndex(t, fi.Name.Name)
		switch {
		case i < 0:
			c.errorf(diagnostics.TypeStructLit, fi.Name.Span, "struct %s has no field '%s'", t, fi.Name.Name).
				WithHelp("its fields are " + fieldList(t))
			c.expr(fi.Value)
			continue
		case given[fi.Name.Name]:
			c.errorf(diagnostics.TypeStructLit, fi.Name.Span, "field '%s' is given twice", fi.Name.Name)
			c.expr(fi.Value)
			continue
		}
		given[fi.Name.Name] = true
		if ft := t.Fields[i].Type; ft.Kind != types.Invalid {
			c.assign(ft, fi.Value, "assign")
		} else {
			c.expr(fi.Value)
		}
	}
	var missing []string
	for _, f := range t.Fields {
		if !given[f.Name] {
			missing = append(missing, f.Name)
		}
	}
	if len(missing) > 0 {
		c.errorf(diagnostics.TypeStructLit, e.Span, "%s{...} is missing %s", t, plural("field", missing)).
			WithHelp("a struct literal gives every field a value; there are no default values")
	}
	c.info.Types[e] = t
	return t
}

func plural(word string, names []string) string {
	q := make([]string, len(names))
	for i, n := range names {
		q[i] = "'" + n + "'"
	}
	if len(q) == 1 {
		return word + " " + q[0]
	}
	return word + "s " + strings.Join(q, ", ")
}

func fieldList(t *types.Type) string {
	var names []string
	for _, f := range t.Fields {
		names = append(names, f.Name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// structField checks x.name on a struct value.
func (c *checker) structField(e *ast.SelectorExpr, t *types.Type) *types.Type {
	i := types.FieldIndex(t, e.Name.Name)
	if i < 0 {
		c.errorf(diagnostics.TypeNoField, e.Name.Span, "struct %s has no field '%s'", t, e.Name.Name).
			WithHelp("its fields are " + fieldList(t))
		return types.InvalidT
	}
	return t.Fields[i].Type
}

// setStmt checks place op= value, where place is an element or a field.
func (c *checker) setStmt(s *ast.SetStmt) {
	t := c.place(s.Target)
	if t.Kind == types.Invalid {
		c.expr(s.Value)
		return
	}
	c.info.Types[s.Target] = t
	if s.Op == "=" {
		c.assign(t, s.Value, "assign")
		return
	}
	op := strings.TrimSuffix(s.Op, "=")
	y := c.expr(s.Value)
	r := c.binaryType(op, operand{typ: t}, y, s.Target, s.Value, s.Span)
	if r.typ.Kind != types.Invalid && !types.Identical(r.typ, t) {
		c.errorf(diagnostics.TypeMismatch, s.Span, "operator '%s' produces '%s', which cannot be stored in '%s'", s.Op, r.typ, t)
	}
}

// place checks an expression assigned to and returns its type: a variable,
// an array element (of any array: arrays are references), or a field of a
// place. A field of a temporary value, such as f().x, is not a place: the
// change would be lost.
func (c *checker) place(e ast.Expr) *types.Type {
	switch e := unparen(e).(type) {
	case *ast.Ident:
		v := c.target(e)
		if v == nil {
			return types.InvalidT
		}
		c.info.Types[e] = v.typ
		return v.typ
	case *ast.IndexExpr:
		t := c.index(e)
		c.info.Types[e] = t
		return t
	case *ast.SelectorExpr:
		if c.isPackageRef(e) {
			c.errorf(diagnostics.NameAssignToParam, e.Span, "cannot assign to %s.%s", e.X.(*ast.Ident).Name, e.Name.Name)
			return types.InvalidT
		}
		var xt *types.Type
		switch unparen(e.X).(type) {
		case *ast.Ident, *ast.IndexExpr, *ast.SelectorExpr:
			xt = c.place(e.X)
		default:
			xt = c.defaulted(e.X).typ
			if xt.Kind == types.Struct {
				c.errorf(diagnostics.NameAssignToParam, e.Span, "cannot assign to a field of a temporary value").
					WithHelp("store the struct in a variable first, then change its field")
				return types.InvalidT
			}
		}
		switch xt.Kind {
		case types.Invalid:
			return types.InvalidT
		case types.Struct:
			t := c.structField(e, xt)
			c.info.Types[e] = t
			return t
		case types.Array:
			c.errorf(diagnostics.NameAssignToParam, e.Name.Span, "an array's length cannot be assigned").
				WithHelp("use push and pop to change it")
		case types.Error:
			c.errorf(diagnostics.NameAssignToParam, e.Name.Span, "the fields of an error cannot be changed")
		default:
			c.errorf(diagnostics.TypeNoField, e.Name.Span, "values of type '%s' have no fields", xt)
		}
		return types.InvalidT
	}
	c.errorf(diagnostics.ParseUnexpected, e.NodeSpan(), "cannot assign to this expression")
	return types.InvalidT
}
