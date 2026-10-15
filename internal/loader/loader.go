// Package loader finds and parses every source file of a program: the
// entry file and everything it reaches through `use` declarations.
//
// A use names one of three things (owner decisions, v0.3):
//
//   - "std:name": a standard library package, resolved by the checker.
//   - a path, written with at least one '/' ("./geom.rsc", "lib/x.txt"):
//     that one file, whatever its extension.
//   - a package name ("geom", "org.x.geom"): every .rsm file directly in
//     the directory org/x/geom. Each must declare `package org.x.geom;`.
//
// Paths and package directories are relative to the working directory,
// not to the file containing the use. Each file is loaded once, however
// many times it is used. Packages are identified by name: every loaded
// file that declares `package geom;` belongs to package geom.
package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/lexer"
	"github.com/AurielSolaris/runescript/internal/parser"
	"github.com/AurielSolaris/runescript/internal/srcenc"
)

// ModuleExt is the extension of the files that make up a package
// directory. Files named directly can have any extension.
const ModuleExt = ".rsm"

// Kind classifies a use declaration.
type Kind int

const (
	UseStd         Kind = iota // "std:name"
	UsePackage                 // "a.b"
	UseFile                    // a path containing '/'
	UseUnsupported             // another "scheme:" form, e.g. "go:strings"
)

// Program is every file of a program, in load order.
type Program struct {
	Entry *File
	// Files[i].Source.Index == i; the entry file is first.
	Files []*File
	// Packages are in order of first appearance.
	Packages []*Package
	// Dir is the working directory that paths were resolved against.
	Dir string
}

// Package is a named package and the loaded files that declare it.
type Package struct {
	Name  string
	Files []*File
}

// File is one source file.
type File struct {
	Source *diagnostics.Source
	// AST and Package are nil if the file failed to parse.
	AST     *ast.File
	Package *Package
	// Uses has one entry per use declaration, in source order.
	Uses []*Use
}

// Use is a resolved use declaration.
type Use struct {
	Decl *ast.UseDecl
	Kind Kind
	// Name is the package name for Std ("string") and Package ("a.b").
	Name string
	// Target is the package a Package or File use refers to. It is nil
	// when the target could not be loaded; that is already reported.
	Target *Package
	// Qualifier is the name calls are qualified with: the alias, or the
	// last segment of the target package's name. It is "" for a Std use
	// without an alias (the checker knows std qualifiers) and for a
	// broken use whose package name is unknown.
	Qualifier string
}

// Self reports whether u refers to the package of the file it is in.
// Such a use loads files but binds no qualifier: their functions are
// already in scope.
func (u *Use) Self(in *File) bool { return u.Target != nil && u.Target == in.Package }

// Broken reports whether u is a RuneScript package or file use whose
// target could not be loaded.
func (u *Use) Broken() bool {
	return (u.Kind == UsePackage || u.Kind == UseFile) && u.Target == nil
}

// Load reads the entry file at path and everything it uses. dir is the
// working directory. err reports only an unreadable entry file; every
// other problem is a diagnostic.
func Load(path, dir string) (*Program, []*diagnostics.Diagnostic, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	l := newLoader(dir)
	l.load(diagnostics.NewSource(entryName(path, dir), text), info)
	return l.finish()
}

// entryName is how the entry file is named in diagnostics: relative to the
// working directory when it is inside it, otherwise as given.
func entryName(path, dir string) string {
	abs, err1 := filepath.Abs(path)
	base, err2 := filepath.Abs(dir)
	if err1 != nil || err2 != nil {
		return path
	}
	rel, err := filepath.Rel(base, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}
	return filepath.ToSlash(rel)
}

// LoadSource loads a program whose entry file is already in memory.
func LoadSource(src *diagnostics.Source, dir string) (*Program, []*diagnostics.Diagnostic) {
	l := newLoader(dir)
	l.load(src, nil)
	prog, diags, _ := l.finish()
	return prog, diags
}

type loader struct {
	prog  *Program
	diags []*diagnostics.Diagnostic
	pkgs  map[string]*Package
	infos []os.FileInfo // per file, for os.SameFile; nil for in-memory files
	dirs  map[string]*Package
	queue []*File
}

func newLoader(dir string) *loader {
	return &loader{
		prog: &Program{Dir: dir},
		pkgs: map[string]*Package{},
		dirs: map[string]*Package{},
	}
}

func (l *loader) errorf(code diagnostics.Code, span diagnostics.Span, format string, args ...any) *diagnostics.Diagnostic {
	d := diagnostics.New(code, span, format, args...)
	l.diags = append(l.diags, d)
	return d
}

func (l *loader) finish() (*Program, []*diagnostics.Diagnostic, error) {
	for len(l.queue) > 0 {
		f := l.queue[0]
		l.queue = l.queue[1:]
		l.uses(f)
	}
	l.cycles()
	return l.prog, l.diags, nil
}

// load parses src and adds it to the program.
func (l *loader) load(src *diagnostics.Source, info os.FileInfo) *File {
	src.Text, src.ConvertedFrom = srcenc.Decode(src.Text)
	src.Index = len(l.prog.Files)
	f := &File{Source: src}
	l.prog.Files = append(l.prog.Files, f)
	l.infos = append(l.infos, info)
	if l.prog.Entry == nil {
		l.prog.Entry = f
	}
	file, diags := parser.ParseSource(src)
	if len(diags) > 0 {
		l.diags = append(l.diags, diags...)
		return f
	}
	f.AST = file
	f.Package = l.pkg(strings.Join(file.Package.Name, "."))
	f.Package.Files = append(f.Package.Files, f)
	l.queue = append(l.queue, f)
	return f
}

func (l *loader) pkg(name string) *Package {
	p := l.pkgs[name]
	if p == nil {
		p = &Package{Name: name}
		l.pkgs[name] = p
		l.prog.Packages = append(l.prog.Packages, p)
	}
	return p
}

// open returns the loaded file at path, loading it if needed. A missing
// or unreadable file is reported at span.
func (l *loader) open(path, name string, span diagnostics.Span) (*File, bool) {
	abs := path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(l.prog.Dir, path)
	}
	info, err := os.Stat(abs)
	if err == nil && info.IsDir() {
		l.errorf(diagnostics.UseNotFound, span, "%q is a directory, not a file", name).
			WithHelp(`to use a package, write its name: use "a.b"; loads the .rsm files in a/b`)
		return nil, false
	}
	if err != nil {
		l.errorf(diagnostics.UseNotFound, span, "cannot find file %q", name).
			WithHelp(fmt.Sprintf("paths are relative to the working directory (%s)", l.prog.Dir))
		return nil, false
	}
	for i, other := range l.infos {
		if other != nil && os.SameFile(other, info) {
			return l.prog.Files[i], true
		}
	}
	text, err := os.ReadFile(abs)
	if err != nil {
		l.errorf(diagnostics.UseNotFound, span, "cannot read %q: %v", name, err)
		return nil, false
	}
	return l.load(diagnostics.NewSource(name, text), info), true
}

// Classify reports what kind of thing a use path names. For Std and
// Package it also returns the package name.
func Classify(path string) (Kind, string) {
	if name, ok := strings.CutPrefix(path, "std:"); ok {
		return UseStd, name
	}
	if strings.Contains(path, "/") {
		return UseFile, ""
	}
	if strings.Contains(path, ":") {
		return UseUnsupported, ""
	}
	return UsePackage, path
}

// uses resolves f's use declarations.
func (l *loader) uses(f *File) {
	for _, d := range f.AST.Uses {
		u := &Use{Decl: d}
		u.Kind, u.Name = Classify(d.Path)
		f.Uses = append(f.Uses, u)
		switch u.Kind {
		case UseUnsupported:
			l.errorf(diagnostics.UseUnsupported, d.PathSpan, "cannot use %q: only \"std:\" packages, RuneScript packages, and files are supported", d.Path).
				WithHelp(`using Go packages ("go:...") is planned`)
		case UsePackage:
			if strings.Contains(d.Path, `\`) {
				l.errorf(diagnostics.UseBadName, d.PathSpan, "%q is not a package name or a path", d.Path).
					WithHelp(`separate path segments with '/', e.g. use "./lib/geom.rsc";`)
				break
			}
			if !validPackageName(u.Name) {
				l.errorf(diagnostics.UseBadName, d.PathSpan, "%q is not a valid package name", d.Path).
					WithHelp(`a package name is dot-separated identifiers, e.g. "org.x.geom"; to use a file, write a path such as "./geom.rsc"`)
				break
			}
			u.Target = l.packageDir(u.Name, d.PathSpan)
		case UseFile:
			if g, ok := l.open(d.Path, displayName(d.Path), d.PathSpan); ok {
				u.Target = g.Package
			}
		}
		switch {
		case d.Alias != nil:
			u.Qualifier = d.Alias.Name
		case u.Target != nil:
			u.Qualifier = lastSegment(u.Target.Name)
		case u.Kind == UsePackage:
			u.Qualifier = lastSegment(u.Name)
		}
	}
}

func validPackageName(name string) bool {
	for _, seg := range strings.Split(name, ".") {
		if !lexer.IsName(seg) {
			return false
		}
	}
	return true
}

func lastSegment(name string) string { return name[strings.LastIndex(name, ".")+1:] }

// displayName is how a used file is named in diagnostics: the path as
// written, cleaned, with '/' separators.
func displayName(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

// packageDir loads the .rsm files of package name. It returns nil if none
// could be loaded.
func (l *loader) packageDir(name string, span diagnostics.Span) *Package {
	if p, ok := l.dirs[name]; ok {
		return p
	}
	l.dirs[name] = nil
	rel := filepath.Join(strings.Split(name, ".")...)
	entries, _ := os.ReadDir(filepath.Join(l.prog.Dir, rel))
	var files []string
	for _, e := range entries {
		if e.Type().IsRegular() && filepath.Ext(e.Name()) == ModuleExt {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		l.errorf(diagnostics.UseNotFound, span, "cannot find package %q: there are no %s files in %s", name, ModuleExt, filepath.ToSlash(rel)+"/").
			WithHelp(fmt.Sprintf("package directories are relative to the working directory (%s); to use a single file, write a path such as \"./%s.rsc\"", l.prog.Dir, lastSegment(name)))
		return nil
	}
	var found bool
	for _, base := range files {
		path := filepath.Join(rel, base)
		f, ok := l.open(path, displayName(path), span)
		if !ok || f.AST == nil {
			continue
		}
		if f.Package.Name != name {
			l.errorf(diagnostics.UsePackageMismatch, f.AST.Package.Span, "%s is in the directory of package %q but declares package %q", f.Source.Name, name, f.Package.Name).
				WithHelp("change its package declaration, or move the file out of " + filepath.ToSlash(rel) + "/")
			continue
		}
		found = true
	}
	if !found {
		return nil
	}
	p := l.pkg(name)
	l.dirs[name] = p
	return p
}

// cycles reports packages that use each other in a cycle.
func (l *loader) cycles() {
	type edge struct {
		to  *Package
		use *Use
	}
	edges := map[*Package][]edge{}
	for _, f := range l.prog.Files {
		for _, u := range f.Uses {
			if u.Target != nil && !u.Self(f) {
				edges[f.Package] = append(edges[f.Package], edge{u.Target, u})
			}
		}
	}
	const (
		unvisited = iota
		active
		done
	)
	state := map[*Package]int{}
	var stack []*Package
	var visit func(p *Package)
	visit = func(p *Package) {
		state[p] = active
		stack = append(stack, p)
		for _, e := range edges[p] {
			switch state[e.to] {
			case unvisited:
				visit(e.to)
			case active:
				var names []string
				for i := len(stack) - 1; i >= 0; i-- {
					names = append([]string{stack[i].Name}, names...)
					if stack[i] == e.to {
						break
					}
				}
				names = append(names, e.to.Name)
				l.errorf(diagnostics.UseCycle, e.use.Decl.PathSpan, "import cycle: %s", strings.Join(names, " → ")).
					WithHelp("packages cannot use each other in a cycle; move the functions they share into a separate package")
			}
		}
		stack = stack[:len(stack)-1]
		state[p] = done
	}
	for _, p := range l.prog.Packages {
		if state[p] == unvisited {
			visit(p)
		}
	}
}
