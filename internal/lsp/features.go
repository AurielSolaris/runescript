package lsp

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/lexer"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// at returns the document's current analysis, if its file parsed, and the
// byte offset of pos in it.
func (s *Server) at(uri string, pos Position) (*analysis, int) {
	d := s.docs[uri]
	if d == nil || d.last == nil || d.last.file == nil || d.last.file.AST == nil || d.last.info == nil {
		return nil, 0
	}
	return d.last, toOffset(d.last.file.Source.Text, pos, s.utf16)
}

// ---- hover ----

func (s *Server) hover(p *textDocumentPositionParams) *hover {
	a, off := s.at(p.TextDocument.URI, p.Position)
	if a == nil {
		return nil
	}
	path := pathAt(a.file.AST, off)
	if len(path) == 0 {
		return nil
	}
	text, sp := a.describe(path)
	if text == "" {
		return nil
	}
	r := s.toRange(sp, a.file.Source.Text)
	return &hover{Contents: markupContent{Kind: "markdown", Value: text}, Range: &r}
}

func code(lines ...string) string {
	return "```runescript\n" + strings.Join(lines, "\n") + "\n```"
}

// describe returns the hover text for the innermost node of path, and the
// span it describes.
func (a *analysis) describe(path []node) (string, ast.Span) {
	info := a.info
	n := path[len(path)-1]
	var parent, grand node
	if len(path) >= 2 {
		parent = path[len(path)-2]
	}
	if len(path) >= 3 {
		grand = path[len(path)-3]
	}
	sp, _ := spanOf(n)
	switch n := n.(type) {
	case *ast.Ident:
		return a.describeIdent(n, parent, grand), sp
	case *ast.TypeRef:
		return a.describeType(info.TypeRefs[n]), sp
	case *ast.UseDecl:
		return a.describeUse(n), sp
	case ast.Expr:
		if t := info.Types[n]; t != nil && t.Kind != types.Invalid {
			return code(t.String()), sp
		}
	}
	return "", sp
}

func (a *analysis) describeIdent(n *ast.Ident, parent, grand node) string {
	info := a.info
	switch p := parent.(type) {
	case *ast.CallExpr:
		if n == p.Fn {
			switch {
			case info.Callees[p] != nil:
				return code(funcSig(info.Callees[p]))
			case info.Natives[p] != nil:
				return a.nativeText(p.Pkg.Name, info.Natives[p])
			case info.Externs[p] != nil:
				return code(externSig(info.Externs[p]))
			case info.Builtins[p] != "":
				return code("bool " + info.Builtins[p] + "(float x)")
			case p.Pkg == nil && n.Name == "trace" && info.VarRefs[n] == nil:
				return code("void trace(...)") + "\n\nWrites its arguments, separated by spaces, and a line break."
			case info.ArrayCalls[p] != "":
				if t := info.Types[p]; t != nil {
					return code(t.String()+" "+info.ArrayCalls[p]+"(...)") + "\n\nArray method"
				}
			}
			if t := info.Types[n]; t != nil {
				return code(t.String() + " " + n.Name)
			}
		}
		if n == p.Pkg {
			if decl := info.VarRefs[n]; decl != nil { // an array: a.push(x)
				if t := info.Types[n]; t != nil {
					return code(t.String() + " " + n.Name)
				}
			}
			return a.describeQualifier(n.Name)
		}
	case *ast.SelectorExpr:
		if n == p.Name {
			if f := info.QualRefs[p]; f != nil {
				return code(funcSig(f))
			}
			if t := info.Types[p]; t != nil {
				return code(t.String() + " " + n.Name)
			}
		}
	case *ast.FuncDecl:
		if n == p.Name {
			return code(declSig(p))
		}
	case *ast.ExternDecl:
		if n == p.Name {
			return code(externDeclSig(p))
		}
	case *ast.StructDecl:
		return code(structText(p))
	case *ast.FieldDecl:
		return code(p.Type.String()+" "+p.Name.Name) + "\n\nField"
	case *ast.VarDecl:
		if n == p.Name {
			return code(typeText(info, p.Type) + " " + n.Name)
		}
	case *ast.Param:
		return code(typeText(info, p.Type)+" "+n.Name) + "\n\nParameter"
	case *ast.ForEachStmt:
		if n == p.Name {
			return code(typeText(info, p.Type) + " " + n.Name)
		}
	case *ast.MForEach:
		if n == p.Name {
			return code(typeText(info, p.Type) + " " + n.Name)
		}
	case *ast.CatchClause:
		if t := info.CatchTypes[p]; t != nil {
			return code(t.String() + " " + n.Name)
		}
	case *ast.StructLit:
		if n == p.Name {
			return a.describeType(info.Types[p])
		}
		return a.describeQualifier(n.Name)
	case *ast.FieldInit:
		if lit, ok := grand.(*ast.StructLit); ok {
			if t := info.Types[lit]; t != nil {
				for _, f := range t.Fields {
					if f.Name == n.Name {
						return code(f.Type.String()+" "+f.Name) + "\n\nField of " + t.String()
					}
				}
			}
		}
	case *ast.UseDecl, *ast.TypeRef:
		return a.describeQualifier(n.Name)
	}
	if f := info.FuncRefs[n]; f != nil {
		return code(funcSig(f))
	}
	if t := info.Types[n]; t != nil && t.Kind != types.Invalid {
		return code(t.String() + " " + n.Name)
	}
	return ""
}

func (a *analysis) describeType(t *types.Type) string {
	if t == nil || t.Kind == types.Invalid {
		return ""
	}
	if st := a.info.Structs[t]; st != nil {
		return code(structText(st.Decl))
	}
	return code(t.String())
}

// describeUse describes the package or file a use declaration loads.
func (a *analysis) describeUse(u *ast.UseDecl) string {
	kind, name := loader.Classify(u.Path)
	switch kind {
	case loader.UseStd:
		return a.stdText(name)
	case loader.UsePackage, loader.UseFile:
		for _, lu := range a.file.Uses {
			if lu.Decl == u && lu.Target != nil {
				var files []string
				for _, f := range lu.Target.Files {
					files = append(files, "- "+f.Source.Name)
				}
				return code("package "+lu.Target.Name) + "\n\n" + strings.Join(files, "\n")
			}
		}
	}
	return ""
}

// qualifier is the name u binds: its alias, or the last part of the
// package's name (the loader leaves it empty for std: packages).
func qualifier(u *loader.Use) string {
	switch {
	case u.Qualifier != "":
		return u.Qualifier
	case u.Decl.Alias != nil:
		return u.Decl.Alias.Name
	case u.Kind == loader.UseStd:
		return u.Name[strings.LastIndex(u.Name, ".")+1:]
	}
	return ""
}

// describeQualifier describes the package a qualifier names in this file.
func (a *analysis) describeQualifier(q string) string {
	for _, u := range a.file.Uses {
		if qualifier(u) != q {
			continue
		}
		switch {
		case u.Kind == loader.UseStd:
			return a.stdText(u.Name)
		case u.Target != nil:
			return code("package " + u.Target.Name)
		}
		return code("use \"" + u.Decl.Path + "\"")
	}
	return ""
}

func (a *analysis) stdText(name string) string {
	pkg := stdlib.Lookup(name)
	if pkg == nil {
		return code("use \"std:" + name + "\"")
	}
	var names []string
	for n := range pkg.Funcs {
		names = append(names, n)
	}
	sort.Strings(names)
	return code("use \"std:"+name+"\"") + "\n\nStandard library package: " + strings.Join(names, ", ")
}

func (a *analysis) nativeText(qual string, f *stdlib.Func) string {
	pkg := stdlib.Lookup(f.Package)
	fs := []*stdlib.Func{f}
	if pkg != nil && len(pkg.Overloads[f.Name]) > 1 {
		fs = pkg.Overloads[f.Name]
	}
	var lines []string
	for _, g := range fs {
		lines = append(lines, nativeSig(qual, g))
	}
	return code(lines...) + "\n\nstd:" + f.Package
}

func nativeSig(qual string, f *stdlib.Func) string {
	var ps []string
	for _, p := range f.Params {
		ps = append(ps, p.String())
	}
	return fmt.Sprintf("%s %s.%s(%s)", f.Result, qual, f.Name, strings.Join(ps, ", "))
}

func funcSig(f *checker.FuncSym) string {
	if f.Def != nil {
		return declSig(f.Def)
	}
	if f.Proto != nil {
		return declSig(f.Proto)
	}
	return f.Type.String() + " " + f.Name
}

func declSig(d *ast.FuncDecl) string {
	var ps []string
	for _, p := range d.Params {
		if p.Name != nil {
			ps = append(ps, p.Type.String()+" "+p.Name.Name)
		} else {
			ps = append(ps, p.Type.String())
		}
	}
	return fmt.Sprintf("%s %s %s(%s)", d.Visibility, d.Result, d.Name.Name, strings.Join(ps, ", "))
}

func externSig(e *checker.ExternSym) string {
	if e.Decl != nil {
		return externDeclSig(e.Decl)
	}
	return e.Type.String() + " " + e.Name
}

func externDeclSig(d *ast.ExternDecl) string {
	var ps []string
	for _, p := range d.Params {
		if p.Name != nil {
			ps = append(ps, p.Type.String()+" "+p.Name.Name)
		} else {
			ps = append(ps, p.Type.String())
		}
	}
	return fmt.Sprintf("extern %q %s %s(%s)", d.Module, d.Result, d.Name.Name, strings.Join(ps, ", "))
}

func structText(d *ast.StructDecl) string {
	lines := []string{fmt.Sprintf("%s struct %s {", d.Visibility, d.Name.Name)}
	for _, f := range d.Fields {
		lines = append(lines, "    "+f.Type.String()+" "+f.Name.Name+";")
	}
	return strings.Join(append(lines, "}"), "\n")
}

// typeText is the resolved type of a written type, or the type as written.
func typeText(info *checker.Info, t *ast.TypeRef) string {
	if t == nil {
		return ""
	}
	if r := info.TypeRefs[t]; r != nil && r.Kind != types.Invalid {
		return r.String()
	}
	return t.String()
}

// ---- go to definition ----

func (s *Server) definition(p *textDocumentPositionParams) []Location {
	a, off := s.at(p.TextDocument.URI, p.Position)
	if a == nil {
		return nil
	}
	path := pathAt(a.file.AST, off)
	if len(path) == 0 {
		return nil
	}
	var spans []ast.Span
	switch n := path[len(path)-1].(type) {
	case *ast.Ident:
		var parent, grand node
		if len(path) >= 2 {
			parent = path[len(path)-2]
		}
		if len(path) >= 3 {
			grand = path[len(path)-3]
		}
		spans = a.identTargets(n, parent, grand)
	case *ast.TypeRef:
		if sp, ok := a.structSpan(a.info.TypeRefs[n]); ok {
			spans = append(spans, sp)
		}
	case *ast.UseDecl:
		spans = a.useTargets(n)
	}
	var locs []Location
	for _, sp := range spans {
		if loc, ok := s.location(a, sp); ok {
			locs = append(locs, loc)
		}
	}
	return locs
}

func (a *analysis) identTargets(n *ast.Ident, parent, grand node) []ast.Span {
	info := a.info
	one := func(sp ast.Span, ok bool) []ast.Span {
		if !ok {
			return nil
		}
		return []ast.Span{sp}
	}
	if decl := info.VarRefs[n]; decl != nil {
		return []ast.Span{decl.Span}
	}
	if f := info.FuncRefs[n]; f != nil {
		return one(funcSpan(f))
	}
	switch p := parent.(type) {
	case *ast.CallExpr:
		if n == p.Fn {
			if f := info.Callees[p]; f != nil {
				return one(funcSpan(f))
			}
			if e := info.Externs[p]; e != nil && e.Decl != nil {
				return []ast.Span{e.Decl.Name.Span}
			}
		}
		if n == p.Pkg {
			return a.qualifierTargets(n.Name)
		}
	case *ast.SelectorExpr:
		if n == p.Name {
			if f := info.QualRefs[p]; f != nil {
				return one(funcSpan(f))
			}
			if x, ok := p.X.(ast.Expr); ok {
				return one(a.fieldSpan(info.Types[x], n.Name))
			}
		}
	case *ast.StructLit:
		if n == p.Name {
			return one(a.structSpan(info.Types[p]))
		}
		return a.qualifierTargets(n.Name)
	case *ast.FieldInit:
		if lit, ok := grand.(*ast.StructLit); ok {
			return one(a.fieldSpan(info.Types[lit], n.Name))
		}
	case *ast.TypeRef, *ast.UseDecl:
		return a.qualifierTargets(n.Name)
	}
	return nil
}

func funcSpan(f *checker.FuncSym) (ast.Span, bool) {
	switch {
	case f.Def != nil:
		return f.Def.Name.Span, true
	case f.Proto != nil:
		return f.Proto.Name.Span, true
	}
	return ast.Span{}, false
}

func (a *analysis) structSpan(t *types.Type) (ast.Span, bool) {
	if t == nil {
		return ast.Span{}, false
	}
	if st := a.info.Structs[t]; st != nil {
		return st.Decl.Name.Span, true
	}
	return ast.Span{}, false
}

func (a *analysis) fieldSpan(t *types.Type, name string) (ast.Span, bool) {
	if t == nil {
		return ast.Span{}, false
	}
	if st := a.info.Structs[t]; st != nil {
		for _, f := range st.Decl.Fields {
			if f.Name.Name == name {
				return f.Name.Span, true
			}
		}
	}
	return ast.Span{}, false
}

// qualifierTargets are the files of the package a qualifier names.
func (a *analysis) qualifierTargets(q string) []ast.Span {
	for _, u := range a.file.Uses {
		if qualifier(u) == q {
			return a.useTargets(u.Decl)
		}
	}
	return nil
}

// useTargets are the package declarations of the files a use loads.
func (a *analysis) useTargets(d *ast.UseDecl) []ast.Span {
	for _, u := range a.file.Uses {
		if u.Decl != d || u.Target == nil {
			continue
		}
		var out []ast.Span
		for _, f := range u.Target.Files {
			if f.AST != nil && f.AST.Package != nil {
				out = append(out, f.AST.Package.Span)
			}
		}
		return out
	}
	return nil
}

// location converts a span, in any file of the program, to an LSP
// location.
func (s *Server) location(a *analysis, sp ast.Span) (Location, bool) {
	src := sp.Start.File
	if src == nil {
		return Location{}, false
	}
	path := filePath(a.prog, src)
	if path == "" {
		return Location{}, false
	}
	return Location{URI: pathToURI(path), Range: s.toRange(sp, src.Text)}, true
}

// ---- completion ----

var (
	useRE       = regexp.MustCompile(`\buse\s+"([^"]*)$`)
	qualifiedRE = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\.[A-Za-z0-9_]*$`)
)

// contextual words are not keywords but begin constructs.
var contextual = []string{"Array", "each", "in", "as", "template", "NaN", "Infinity", "isNaN", "isFinite", "trace"}

func (s *Server) completion(p *textDocumentPositionParams) []completionItem {
	items := []completionItem{}
	d := s.docs[p.TextDocument.URI]
	if d == nil {
		return items
	}
	off := toOffset(d.text, p.Position, s.utf16)
	lineStart := strings.LastIndexByte(string(d.text[:off]), '\n') + 1
	prefix := string(d.text[lineStart:off])

	a := d.last
	if a == nil || a.file == nil || a.file.AST == nil || a.info == nil {
		a = d.good // the last version that parsed
	}
	seen := map[string]bool{}
	add := func(label string, kind int, detail string) {
		if !seen[label] {
			seen[label] = true
			items = append(items, completionItem{Label: label, Kind: kind, Detail: detail})
		}
	}

	if m := useRE.FindStringSubmatch(prefix); m != nil {
		paths := stdlib.Paths()
		sort.Strings(paths)
		for _, name := range paths {
			if strings.HasPrefix(m[1], "std:") {
				add(name, kindModule, "std:"+name)
			} else {
				add("std:"+name, kindModule, "standard library")
			}
		}
		return items
	}
	if m := qualifiedRE.FindStringSubmatch(prefix); m != nil {
		if a != nil && a.file != nil {
			a.memberItems(m[1], off, add)
		}
		return items
	}

	for _, k := range lexer.Keywords() {
		add(k, kindKeyword, "")
	}
	for _, w := range contextual {
		add(w, kindKeyword, "")
	}
	if a == nil || a.file == nil || a.file.AST == nil {
		return items
	}
	for _, u := range a.file.Uses {
		if q := qualifier(u); q != "" {
			add(q, kindModule, `use "`+u.Decl.Path+`"`)
		}
	}
	if a.file.Package != nil {
		for _, f := range a.file.Package.Files {
			if f.AST == nil {
				continue
			}
			for _, fd := range f.AST.Funcs {
				if f == a.file || fd.Visibility != ast.VisPrivate {
					add(fd.Name.Name, kindFunction, declSig(fd))
				}
			}
			for _, sd := range f.AST.Structs {
				if f == a.file || sd.Visibility != ast.VisPrivate {
					add(sd.Name.Name, kindStruct, "struct")
				}
			}
		}
	}
	for _, e := range a.file.AST.Externs {
		add(e.Name.Name, kindFunction, externDeclSig(e))
	}
	for _, v := range a.locals(off) {
		add(v.name, kindVariable, v.typ)
	}
	return items
}

// memberItems completes after q.: a package's functions, an array's
// methods, a struct's fields, an error's fields.
func (a *analysis) memberItems(q string, off int, add func(string, int, string)) {
	for _, u := range a.file.Uses {
		if qualifier(u) != q {
			continue
		}
		if u.Kind == loader.UseStd {
			if pkg := stdlib.Lookup(u.Name); pkg != nil {
				for _, f := range pkg.All() {
					add(f.Name, kindFunction, nativeSig(q, f))
				}
				for name := range pkg.Consts {
					add(name, kindConstant, "constant")
				}
				for name := range pkg.Types {
					add(name, kindStruct, "type")
				}
			}
			return
		}
		if u.Target != nil {
			same := a.file.Package == u.Target
			for _, f := range u.Target.Files {
				if f.AST == nil {
					continue
				}
				for _, fd := range f.AST.Funcs {
					if fd.Visibility == ast.VisPublic || (same && fd.Visibility == ast.VisProtected) {
						add(fd.Name.Name, kindFunction, declSig(fd))
					}
				}
				for _, sd := range f.AST.Structs {
					if sd.Visibility == ast.VisPublic || (same && sd.Visibility == ast.VisProtected) {
						add(sd.Name.Name, kindStruct, "struct")
					}
				}
			}
		}
		return
	}
	if a.info == nil {
		return
	}
	for _, v := range a.locals(off) {
		if v.name != q || v.t == nil {
			continue
		}
		switch t := v.t; {
		case t.Kind == types.Array:
			add("length", kindProperty, "int")
			for _, m := range []string{"push", "pop", "insert", "removeAt", "slice", "indexOf", "contains", "reverse", "sort"} {
				add(m, kindMethod, "Array method")
			}
		case t.Kind == types.Struct:
			for _, f := range t.Fields {
				add(f.Name, kindField, f.Type.String())
			}
		case t.Kind == types.Error:
			add("message", kindField, "string")
			add("name", kindField, "string")
			add("errorID", kindField, "int")
		}
		return
	}
}

// local is a variable or parameter visible at an offset.
type local struct {
	name, typ string
	t         *types.Type
}

// locals returns the variables and parameters visible at off, innermost
// last: the parameters of the enclosing functions and lambdas, and the
// variables declared before off in the enclosing blocks.
func (a *analysis) locals(off int) []local {
	var out []local
	info := a.info
	resolved := func(t *ast.TypeRef) *types.Type {
		if info == nil || t == nil {
			return nil
		}
		return info.TypeRefs[t]
	}
	addVar := func(id *ast.Ident, t *ast.TypeRef) {
		if id == nil {
			return
		}
		rt := resolved(t)
		text := ""
		if t != nil {
			text = typeText(info, t)
		}
		out = append(out, local{name: id.Name, typ: text, t: rt})
	}
	before := func(n node) bool {
		sp, ok := spanOf(n)
		return ok && sp.End.Offset <= off
	}
	for _, n := range pathAt(a.file.AST, off) {
		switch n := n.(type) {
		case *ast.FuncDecl:
			for _, p := range n.Params {
				addVar(p.Name, p.Type)
			}
		case *ast.LambdaExpr:
			for _, p := range n.Params {
				addVar(p.Name, p.Type)
			}
		case *ast.ForEachStmt:
			addVar(n.Name, n.Type)
		case *ast.MForEach:
			addVar(n.Name, n.Type)
		case *ast.ForStmt:
			if vd, ok := n.Init.(*ast.VarDecl); ok {
				addVar(vd.Name, vd.Type)
			}
		case *ast.CatchClause:
			if info != nil && n.Name != nil {
				if t := info.CatchTypes[n]; t != nil {
					out = append(out, local{name: n.Name.Name, typ: t.String(), t: t})
				}
			}
		case *ast.Block:
			for _, st := range n.Stmts {
				if !before(st) {
					break
				}
				switch st := st.(type) {
				case *ast.VarDecl:
					addVar(st.Name, st.Type)
				case *ast.FuncDecl:
					out = append(out, local{name: st.Name.Name, typ: declSig(st)})
				}
			}
		}
	}
	// Innermost declarations win: keep the last of each name.
	last := map[string]int{}
	for i, v := range out {
		last[v.name] = i
	}
	var kept []local
	for i, v := range out {
		if last[v.name] == i {
			kept = append(kept, v)
		}
	}
	return kept
}

// ---- document symbols ----

func (s *Server) documentSymbols(uri string) []documentSymbol {
	out := []documentSymbol{}
	d := s.docs[uri]
	if d == nil {
		return out
	}
	a := d.last
	if a == nil || a.file == nil || a.file.AST == nil {
		a = d.good
	}
	if a == nil || a.file == nil || a.file.AST == nil {
		return out
	}
	text := a.file.Source.Text
	sym := func(name, detail string, kind int, whole, sel diagnostics.Span) documentSymbol {
		return documentSymbol{Name: name, Detail: detail, Kind: kind, Range: s.toRange(whole, text), SelectionRange: s.toRange(sel, text)}
	}
	var fn func(fd *ast.FuncDecl) documentSymbol
	fn = func(fd *ast.FuncDecl) documentSymbol {
		ds := sym(fd.Name.Name, declSig(fd), symbolFunction, fd.Span, fd.Name.Span)
		if fd.Body != nil {
			for _, st := range fd.Body.Stmts {
				if nested, ok := st.(*ast.FuncDecl); ok {
					ds.Children = append(ds.Children, fn(nested))
				}
			}
		}
		return ds
	}
	f := a.file.AST
	for _, sd := range f.Structs {
		ds := sym(sd.Name.Name, "struct", symbolStruct, sd.Span, sd.Name.Span)
		for _, fl := range sd.Fields {
			fsp, _ := spanOf(fl)
			ds.Children = append(ds.Children, sym(fl.Name.Name, fl.Type.String(), symbolField, fsp, fl.Name.Span))
		}
		out = append(out, ds)
	}
	for _, cd := range f.Classes {
		ds := sym(cd.Name.Name, "class", symbolClass, cd.Span, cd.Name.Span)
		for _, fl := range cd.Fields {
			ds.Children = append(ds.Children, sym(fl.Name.Name, fl.Type.String(), symbolField, fl.Span, fl.Name.Span))
		}
		for _, c := range cd.Ctors {
			ds.Children = append(ds.Children, sym(c.Name.Name, "constructor", symbolCtor, c.Span, c.Name.Span))
		}
		for _, m := range cd.Methods {
			ds.Children = append(ds.Children, sym(m.Name.Name, declSig(m), symbolMethod, m.Span, m.Name.Span))
		}
		out = append(out, ds)
	}
	for _, e := range f.Externs {
		out = append(out, sym(e.Name.Name, externDeclSig(e), symbolFunction, e.Span, e.Name.Span))
	}
	for _, fd := range f.Funcs {
		out = append(out, fn(fd))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Range.Start.Line < out[j].Range.Start.Line })
	return out
}
