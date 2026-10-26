package lsp

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/parser"
)

// analysis is the result of checking the program around one document.
type analysis struct {
	prog *loader.Program
	// info is the checker's result; nil if a file did not parse. It is
	// kept even when the program has type errors, so hover and
	// go-to-definition work in a program being edited.
	info *checker.Info
	// file is the document's file in prog, or nil.
	file *loader.File
	// diags are the diagnostics in the document.
	diags []*diagnostics.Diagnostic
}

// checkPackage is the package name of the entry the server makes up to
// check a module (.rsm) file together with the rest of its package.
const checkPackage = "runescript_lsp_check"

// analyze checks the program doc belongs to. Files that are open in the
// editor are read from their buffers, the rest from the disk.
func (s *Server) analyze(doc *document) *analysis {
	dir := s.workDir(doc.path)
	opts := loader.Options{Overlay: s.overlay}
	var prog *loader.Program
	var diags []*diagnostics.Diagnostic

	if doc.path != "" && strings.EqualFold(filepath.Ext(doc.path), loader.ModuleExt) {
		// A module is checked as part of its package, through a made-up
		// entry that uses the package, when it sits in its package's
		// directory: lib/geom/a.rsm is package lib.geom.
		if name := packageName(doc.text); name != "" && s.inPackageDir(doc.path, dir, name) {
			src := diagnostics.NewSource("<runescript lsp>", []byte("package "+checkPackage+";\nuse \""+name+"\";\n"))
			prog, diags = loader.LoadSourceWith(src, dir, opts)
		}
	}
	if prog == nil {
		if _, err := os.Stat(doc.path); doc.path != "" && err == nil {
			var err error
			prog, diags, err = loader.LoadWith(doc.path, dir, opts)
			if err != nil {
				prog = nil
			}
		}
	}
	if prog == nil { // a new, unsaved document
		name := filepath.Base(doc.path)
		if doc.path == "" {
			name = "untitled.rsc"
		}
		prog, diags = loader.LoadSourceWith(diagnostics.NewSource(name, doc.text), dir, opts)
	}

	a := &analysis{prog: prog}
	for _, f := range prog.Files {
		if s.sameFile(prog, f, doc) {
			a.file = f
		}
	}
	parsed := true
	for _, f := range prog.Files {
		if f.AST == nil {
			parsed = false
		}
	}
	if parsed {
		info, more := checker.Check(prog)
		a.info = info
		diags = append(diags, more...)
	}
	errors := false
	for _, d := range diags {
		// A file without main is fine while editing: it may be a helper
		// that another file uses.
		if d.Code == diagnostics.EntryMissing {
			continue
		}
		errors = errors || d.Severity == diagnostics.Error
		a.add(d)
	}
	// Warnings are reported only for programs without errors, as the
	// compiler does: errors often cause follow-on warnings.
	if !errors && a.info != nil {
		for _, w := range a.info.Warnings {
			a.add(w)
		}
	}
	return a
}

// add keeps d if it is in the document.
func (a *analysis) add(d *diagnostics.Diagnostic) {
	if a.file != nil && d.Span.Start.File == a.file.Source {
		a.diags = append(a.diags, d)
	}
}

// workDir is the directory a document's uses are resolved against, as
// `runescript run` resolves them against the working directory: the
// workspace folder containing the document, or else its directory.
func (s *Server) workDir(path string) string {
	if path == "" {
		if s.root != "" {
			return s.root
		}
		return "."
	}
	for _, root := range s.roots {
		rel, err := filepath.Rel(root, path)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return root
		}
	}
	return filepath.Dir(path)
}

// inPackageDir reports whether path is directly in the directory of
// package name under dir.
func (s *Server) inPackageDir(path, dir, name string) bool {
	want := filepath.Join(append([]string{dir}, strings.Split(name, ".")...)...)
	return pathKey(filepath.Clean(filepath.Dir(path))) == pathKey(filepath.Clean(want))
}

// packageName returns the package a source text declares, or "".
func packageName(text []byte) string {
	f, diags := parser.ParseSource(diagnostics.NewSource("x", text))
	if len(diags) > 0 || f == nil || f.Package == nil {
		return ""
	}
	return strings.Join(f.Package.Name, ".")
}

// filePath returns the absolute path of a loaded file, or "" for one that
// is not on the disk.
func filePath(prog *loader.Program, src *diagnostics.Source) string {
	name := src.Name
	if strings.HasPrefix(name, "<") {
		return ""
	}
	if !filepath.IsAbs(name) {
		name = filepath.Join(prog.Dir, filepath.FromSlash(name))
	}
	abs, err := filepath.Abs(name)
	if err != nil {
		return ""
	}
	return abs
}

// sameFile reports whether f is doc's file.
func (s *Server) sameFile(prog *loader.Program, f *loader.File, doc *document) bool {
	if doc.path == "" {
		return f.Source.Name == "untitled.rsc"
	}
	p := filePath(prog, f.Source)
	return p != "" && pathKey(p) == pathKey(doc.path)
}

// toDiagnostic converts d, whose span is in text, to an LSP diagnostic.
func (s *Server) toDiagnostic(d *diagnostics.Diagnostic, text []byte) Diagnostic {
	r := s.toRange(d.Span, text)
	msg := d.Message
	if d.Help != "" {
		msg += "\nhelp: " + d.Help
	}
	sev := severityError
	if d.Severity == diagnostics.Warning {
		sev = severityWarning
	}
	return Diagnostic{Range: r, Severity: sev, Code: d.Code.String(), Source: "runescript", Message: msg}
}

// toRange converts a span in text to an LSP range. An empty span covers
// the character it starts at, so the editor has something to underline.
func (s *Server) toRange(sp diagnostics.Span, text []byte) Range {
	start, end := sp.Start.Offset, sp.End.Offset
	if sp.End.Line == 0 || end < start {
		end = start
	}
	if end == start && end < len(text) && text[end] != '\n' {
		end++
		for end < len(text) && text[end]&0xC0 == 0x80 {
			end++
		}
	}
	return Range{Start: toPosition(text, start, s.utf16), End: toPosition(text, end, s.utf16)}
}
