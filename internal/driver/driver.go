// Package driver connects the compiler stages: load, check, and run or
// compile.
package driver

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/ir"
	"github.com/AurielSolaris/runescript/internal/jsgen"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/runtime"
	"github.com/AurielSolaris/runescript/stdlib"
)

// Input is a program's entry file and the working directory its uses are
// resolved against. Set either Path (read from disk) or Source (already in
// memory).
type Input struct {
	Path   string
	Source *diagnostics.Source
	Dir    string
	// Notes, if set, receives a line for each file that was converted to
	// UTF-8 from another encoding.
	Notes io.Writer
	// Warnings says how warnings are reported.
	Warnings WarningPolicy
	// Files, if set, receives the absolute paths of the files the program
	// was loaded from, even when it has errors (watch mode uses them).
	Files func(paths []string)

	// What a running program reads and writes besides its standard
	// output: its standard input (default: empty), standard error
	// (default: discarded), command-line arguments, and whether io.error
	// and io.warn color their lines.
	Stdin  io.Reader
	Stderr io.Writer
	Args   []string
	Color  bool
}

// WarningPolicy controls warnings. Warnings are only reported for programs
// without errors: a program with errors often has follow-on warnings that
// disappear once the errors are fixed.
type WarningPolicy struct {
	// Allow lists warning codes that are not reported.
	Allow map[diagnostics.Code]bool
	// AsErrors turns every reported warning into an error (-Werror).
	AsErrors bool
	// Report, if set, receives the warnings of a program that checks.
	// It is not called when AsErrors is set: the warnings are then
	// returned as errors.
	Report func([]*diagnostics.Diagnostic)
}

// Memory is an in-memory entry file named name, with uses resolved
// against the process's working directory.
func Memory(name string, src []byte) Input {
	return Input{Source: diagnostics.NewSource(name, src), Dir: "."}
}

// CheckInput loads and type-checks a program. Diagnostics are returned in
// load order, then source order; info is nil when there are any. err
// reports an unreadable entry file.
func CheckInput(in Input) (*checker.Info, []*diagnostics.Diagnostic, error) {
	var (
		prog  *loader.Program
		diags []*diagnostics.Diagnostic
	)
	if in.Source != nil {
		prog, diags = loader.LoadSource(in.Source, in.Dir)
	} else {
		var err error
		if prog, diags, err = loader.Load(in.Path, in.Dir); err != nil {
			return nil, nil, err
		}
	}
	if in.Files != nil {
		paths := make([]string, 0, len(prog.Files))
		for _, f := range prog.Files {
			p := f.Source.Name
			if !filepath.IsAbs(p) {
				p = filepath.Join(prog.Dir, filepath.FromSlash(p))
			}
			if abs, err := filepath.Abs(p); err == nil {
				p = abs
			}
			paths = append(paths, p)
		}
		in.Files(paths)
	}
	for _, f := range prog.Files {
		// A UTF-8 byte-order mark is dropped silently: it changes nothing.
		if in.Notes != nil && f.Source.ConvertedFrom != "" && !strings.HasPrefix(f.Source.ConvertedFrom, "UTF-8") {
			fmt.Fprintf(in.Notes, "runescript: note: %s is %s; it was read as UTF-8\n", f.Source.Name, f.Source.ConvertedFrom)
		}
	}
	for _, f := range prog.Files {
		if f.AST == nil {
			// Syntax errors come first; checking a partial program would
			// report confusing follow-on errors.
			return nil, sorted(diags), nil
		}
	}
	info, more := checker.Check(prog)
	diags = append(diags, more...)
	if len(diags) > 0 {
		return nil, sorted(diags), nil
	}
	var warns []*diagnostics.Diagnostic
	for _, w := range info.Warnings {
		if !in.Warnings.Allow[w.Code] {
			warns = append(warns, w)
		}
	}
	switch {
	case len(warns) == 0:
	case in.Warnings.AsErrors:
		for _, w := range warns {
			w.Severity, w.Promoted = diagnostics.Error, true
		}
		return nil, sorted(warns), nil
	case in.Warnings.Report != nil:
		in.Warnings.Report(sorted(warns))
	}
	return info, nil, nil
}

func sorted(diags []*diagnostics.Diagnostic) []*diagnostics.Diagnostic {
	sort.SliceStable(diags, func(i, j int) bool {
		return diags[i].Span.Start.Before(diags[j].Span.Start)
	})
	return diags
}

// RunInput checks and executes a program, writing its output to out. It
// returns the program's exit code, compile-time diagnostics, or an error
// (a *runtime.Error, or an unreadable entry file).
func RunInput(in Input, out io.Writer) (int, []*diagnostics.Diagnostic, error) {
	info, diags, err := CheckInput(in)
	if diags != nil || err != nil {
		return 0, diags, err
	}
	if diags := needsJS(info); diags != nil {
		return 0, diags, nil
	}
	// The interpreter runs the same verified IR the JavaScript back end
	// compiles.
	p := ir.Lower(info)
	if err := ir.Verify(p); err != nil {
		return 0, nil, err
	}
	stderr := in.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	env := stdlib.NewEnv(out, stderr, in.Stdin, in.Args)
	env.Color = in.Color
	code, err := runtime.RunEnv(p, env)
	return code, nil, err
}

// needsJS reports, once, that a program with extern declarations can only
// run on the JavaScript target (OQ-JS-02). It returns nil for any other
// program.
func needsJS(info *checker.Info) []*diagnostics.Diagnostic {
	if info.DOM {
		d := info.DOMUses[0]
		for _, u := range info.DOMUses[1:] {
			if u.Span.Start.Before(d.Span.Start) {
				d = u
			}
		}
		return []*diagnostics.Diagnostic{diagnostics.New(diagnostics.TargetNeedsJS, d.Span,
			"this program uses rune:dom, which needs a browser: the interpreter cannot run it").
			WithHelp("serve it with `runescript dev`, or compile it with `runescript build`")}
	}
	if len(info.ExternDecls) == 0 {
		return nil
	}
	first := info.ExternDecls[0]
	for _, x := range info.ExternDecls[1:] {
		if x.Decl.Span.Start.Before(first.Decl.Span.Start) {
			first = x
		}
	}
	msg := fmt.Sprintf("extern function '%s' is implemented in JavaScript, so the interpreter cannot run this program", first.Name)
	if n := len(info.ExternDecls) - 1; n > 0 {
		msg += fmt.Sprintf(" (%d more extern function(s) are declared)", n)
	}
	return []*diagnostics.Diagnostic{diagnostics.New(diagnostics.TargetNeedsJS, first.Decl.Span, "%s", msg).
		WithHelp("compile with `runescript build` and run the module with node")}
}

// BuildJSInput checks a program and compiles it to one JavaScript module.
func BuildJSInput(in Input, opts jsgen.Options) ([]byte, []*diagnostics.Diagnostic, error) {
	info, diags, err := CheckInput(in)
	if diags != nil || err != nil {
		return nil, diags, err
	}
	js, diags, err := jsgen.Generate(ir.Lower(info), opts)
	return js, sorted(diags), err
}

// BuildJSMapInput is BuildJSInput, also returning the source map when
// opts.SourceMap is set.
func BuildJSMapInput(in Input, opts jsgen.Options) (jsgen.Output, []*diagnostics.Diagnostic, error) {
	info, diags, err := CheckInput(in)
	if diags != nil || err != nil {
		return jsgen.Output{}, diags, err
	}
	out, diags, err := jsgen.GenerateMap(ir.Lower(info), opts)
	return out, sorted(diags), err
}

// Check type-checks a single in-memory file.
func Check(src []byte) (*checker.Info, []*diagnostics.Diagnostic) {
	info, diags, _ := CheckInput(Memory("", src))
	return info, diags
}

// Run checks and executes a single in-memory file.
func Run(src []byte, out io.Writer) (int, []*diagnostics.Diagnostic, error) {
	return RunInput(Memory("", src), out)
}

// BuildJS compiles a single in-memory file named name.
func BuildJS(name string, src []byte, opts jsgen.Options) ([]byte, []*diagnostics.Diagnostic, error) {
	return BuildJSInput(Memory(name, src), opts)
}
