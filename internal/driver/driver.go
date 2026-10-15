// Package driver connects the compiler stages: parse, check, and run or
// compile.
package driver

import (
	"io"
	"sort"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/ir"
	"github.com/AurielSolaris/runescript/internal/jsgen"
	"github.com/AurielSolaris/runescript/internal/parser"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

// Check parses and type-checks src. Diagnostics are returned sorted by
// source position; info is nil when there are any.
func Check(src []byte) (*checker.Info, []*diagnostics.Diagnostic) {
	_, info, diags := parse(src)
	return info, diags
}

func parse(src []byte) (*ast.File, *checker.Info, []*diagnostics.Diagnostic) {
	file, diags := parser.Parse(src)
	if len(diags) > 0 {
		return nil, nil, sorted(diags)
	}
	info, diags := checker.Check(file)
	if len(diags) > 0 {
		return nil, nil, sorted(diags)
	}
	return file, info, nil
}

func sorted(diags []*diagnostics.Diagnostic) []*diagnostics.Diagnostic {
	sort.SliceStable(diags, func(i, j int) bool {
		return diags[i].Span.Start.Offset < diags[j].Span.Start.Offset
	})
	return diags
}

// Run checks and executes src, writing program output to out. It returns
// the program's exit code, compile-time diagnostics, or a runtime error.
func Run(src []byte, out io.Writer) (int, []*diagnostics.Diagnostic, error) {
	info, diags := Check(src)
	if diags != nil {
		return 0, diags, nil
	}
	code, err := runtime.Run(info, out)
	return code, nil, err
}

// BuildJS checks src and compiles it to a JavaScript module. It returns
// the module, compile-time diagnostics, or an internal compiler error.
func BuildJS(src []byte, opts jsgen.Options) ([]byte, []*diagnostics.Diagnostic, error) {
	file, info, diags := parse(src)
	if diags != nil {
		return nil, diags, nil
	}
	out, diags, err := jsgen.Generate(ir.Lower(file, info), opts)
	return out, sorted(diags), err
}
