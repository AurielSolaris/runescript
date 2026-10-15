// Package driver connects the compiler stages: parse, check, and run.
package driver

import (
	"io"

	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/parser"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

// Check parses and type-checks src. Diagnostics are returned in source
// order of discovery; info is nil when there are any.
func Check(src []byte) (*checker.Info, []*diagnostics.Diagnostic) {
	file, diags := parser.Parse(src)
	if len(diags) > 0 {
		return nil, diags
	}
	info, diags := checker.Check(file)
	if len(diags) > 0 {
		return nil, diags
	}
	return info, nil
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
