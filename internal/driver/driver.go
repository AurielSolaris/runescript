// Package driver connects the compiler stages: load, check, and run or
// compile.
package driver

import (
	"io"
	"sort"

	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/ir"
	"github.com/AurielSolaris/runescript/internal/jsgen"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

// Input is a program's entry file and the working directory its uses are
// resolved against. Set either Path (read from disk) or Source (already in
// memory).
type Input struct {
	Path   string
	Source *diagnostics.Source
	Dir    string
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
	code, err := runtime.Run(info, out)
	return code, nil, err
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
