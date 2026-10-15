// Command runescript checks, runs, and compiles RuneScript programs.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/jsgen"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

// Version is the language/implementation version.
const Version = "0.3.0-dev"

const usage = `usage:
  runescript run <file>      check and run a program
  runescript check <file>    check a program without running it
  runescript build [--target js] [-o <out.js>] <file>
                             compile a program to a JavaScript module
                             (default output: <file> with a .js extension)
  runescript version         print the version
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
	if len(args) >= 1 && args[0] == "build" {
		return build(args[1:], stderr)
	}
	if len(args) != 2 || (args[0] != "run" && args[0] != "check") {
		fmt.Fprint(stderr, usage)
		return exitUsage
	}
	in, err := input(args[1])
	if err != nil {
		fmt.Fprintf(stderr, "runescript: %v\n", err)
		return exitUsage
	}
	if args[0] == "check" {
		_, diags, err := driver.CheckInput(in)
		if err != nil {
			fmt.Fprintf(stderr, "runescript: %v\n", err)
			return exitUsage
		}
		if diags != nil {
			report(stderr, diags)
			return exitCompileError
		}
		return 0
	}
	code, diags, err := driver.RunInput(in, stdout)
	if diags != nil {
		report(stderr, diags)
		return exitCompileError
	}
	if err != nil {
		var re *runtime.Error
		if errors.As(err, &re) {
			fmt.Fprint(stderr, diagnostics.Render(re.Diag))
			return exitRuntimeError
		}
		fmt.Fprintf(stderr, "runescript: %v\n", err)
		return exitUsage
	}
	return code
}

// input is the program whose entry file is path. Uses are resolved
// against the working directory. Any file extension is accepted.
func input(path string) (driver.Input, error) {
	dir, err := os.Getwd()
	return driver.Input{Path: path, Dir: dir}, err
}

// build implements `runescript build`.
func build(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	target := fs.String("target", "js", "")
	out := fs.String("o", "", "")
	if err := fs.Parse(args); err != nil || fs.NArg() != 1 {
		fmt.Fprint(stderr, usage)
		return exitUsage
	}
	if *target != "js" {
		fmt.Fprintf(stderr, "runescript: unknown target %q (supported: js)\n", *target)
		return exitUsage
	}
	path := fs.Arg(0)
	if *out == "" {
		*out = strings.TrimSuffix(path, filepath.Ext(path)) + ".js"
	}
	if same(path, *out) {
		fmt.Fprintf(stderr, "runescript: the output file %s would overwrite the source; choose another with -o\n", *out)
		return exitUsage
	}
	in, err := input(path)
	if err != nil {
		fmt.Fprintf(stderr, "runescript: %v\n", err)
		return exitUsage
	}
	js, diags, err := driver.BuildJSInput(in, jsgen.Options{Version: Version})
	if diags != nil {
		report(stderr, diags)
		return exitCompileError
	}
	if err != nil {
		fmt.Fprintf(stderr, "runescript: %v\n", err)
		var pe *os.PathError
		if errors.As(err, &pe) {
			return exitUsage // unreadable entry file
		}
		return exitCompileError // internal compiler error
	}
	if err := os.WriteFile(*out, js, 0o644); err != nil {
		fmt.Fprintf(stderr, "runescript: %v\n", err)
		return exitUsage
	}
	return 0
}

// same reports whether paths a and b name the same file.
func same(a, b string) bool {
	if filepath.Clean(a) == filepath.Clean(b) {
		return true
	}
	ia, errA := os.Stat(a)
	ib, errB := os.Stat(b)
	return errA == nil && errB == nil && os.SameFile(ia, ib)
}

func report(w io.Writer, diags []*diagnostics.Diagnostic) {
	for _, d := range diags {
		fmt.Fprint(w, diagnostics.Render(d))
	}
}
