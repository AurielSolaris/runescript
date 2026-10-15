// Command runescript checks and runs RuneScript programs.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

// Version is the language/implementation version.
const Version = "0.1.0-dev"

const usage = `usage:
  runescript run <file>     check and run a program
  runescript check <file>   check a program without running it
  runescript version        print the version
`

// Exit codes used by the CLI itself. A successful `run` exits with the
// value returned by main.
const (
	exitCompileError = 1
	exitRuntimeError = 1
	exitUsage        = 2
)

func main() {
	os.Exit(cli(os.Args[1:], os.Stdout, os.Stderr))
}

func cli(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "version" {
		fmt.Fprintf(stdout, "runescript %s\n", Version)
		return 0
	}
	if len(args) != 2 || (args[0] != "run" && args[0] != "check") {
		fmt.Fprint(stderr, usage)
		return exitUsage
	}
	path := args[1]
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "runescript: %v\n", err)
		return exitUsage
	}
	if args[0] == "check" {
		if _, diags := driver.Check(src); diags != nil {
			report(stderr, path, src, diags)
			return exitCompileError
		}
		return 0
	}
	code, diags, err := driver.Run(src, stdout)
	if diags != nil {
		report(stderr, path, src, diags)
		return exitCompileError
	}
	if err != nil {
		var re *runtime.Error
		if errors.As(err, &re) {
			fmt.Fprint(stderr, diagnostics.Format(re.Diag, path, src))
		} else {
			fmt.Fprintf(stderr, "runescript: %v\n", err)
		}
		return exitRuntimeError
	}
	return code
}

func report(w io.Writer, path string, src []byte, diags []*diagnostics.Diagnostic) {
	for _, d := range diags {
		fmt.Fprint(w, diagnostics.Format(d, path, src))
	}
}
