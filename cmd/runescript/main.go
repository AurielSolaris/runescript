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
const Version = "0.2.0-dev"

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
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "runescript: %v\n", err)
		return exitUsage
	}
	if *out == "" {
		*out = strings.TrimSuffix(path, filepath.Ext(path)) + ".js"
	}
	js, diags, err := driver.BuildJS(src, jsgen.Options{File: filepath.Base(path), Version: Version})
	if diags != nil {
		report(stderr, path, src, diags)
		return exitCompileError
	}
	if err != nil {
		fmt.Fprintf(stderr, "runescript: %v\n", err)
		return exitCompileError
	}
	if err := os.WriteFile(*out, js, 0o644); err != nil {
		fmt.Fprintf(stderr, "runescript: %v\n", err)
		return exitUsage
	}
	return 0
}

func report(w io.Writer, path string, src []byte, diags []*diagnostics.Diagnostic) {
	for _, d := range diags {
		fmt.Fprint(w, diagnostics.Format(d, path, src))
	}
}
