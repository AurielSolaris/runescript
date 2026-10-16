// Command runescript checks, runs, and compiles RuneScript programs.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	goruntime "runtime"
	"runtime/debug"
	"strings"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/jsgen"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

// Version is the language/implementation version.
const Version = "0.3.0-dev"

const usage = `usage:
  runescript run [flags] <file>     check and run a program
  runescript check [flags] <file>   check a program without running it
  runescript build [flags] [--target js] [-o <out.js>] <file>
                             compile a program to a JavaScript module
                             (default output: <file> with a .js extension)
  runescript version         print the version

flags:
  -Werror                    treat warnings as errors
  --allow <code>[,<code>]    do not report these warnings, e.g.
                             --allow ASX-WARN-002 (may be repeated)
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
		fmt.Fprintf(stdout, "runescript %s%s\n", Version, buildDetails())
		return 0
	}
	if len(args) >= 1 && args[0] == "build" {
		return build(args[1:], stderr)
	}
	if len(args) < 2 || (args[0] != "run" && args[0] != "check") {
		fmt.Fprint(stderr, usage)
		return exitUsage
	}
	fs, policy := flags(args[0], stderr)
	if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 1 || policy.err != nil {
		usageError(stderr, policy.err)
		return exitUsage
	}
	in, err := input(fs.Arg(0), stderr)
	in.Warnings = policy.WarningPolicy
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

// warningFlags holds the warning flags of a command.
type warningFlags struct {
	driver.WarningPolicy
	err error // a bad code given to --allow
}

// flags returns the flag set of command name, with the warning flags
// defined. Warnings are reported to stderr.
func flags(name string, stderr io.Writer) (*flag.FlagSet, *warningFlags) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	w := &warningFlags{WarningPolicy: driver.WarningPolicy{
		Allow:  map[diagnostics.Code]bool{},
		Report: func(ws []*diagnostics.Diagnostic) { report(stderr, ws) },
	}}
	fs.BoolVar(&w.AsErrors, "Werror", false, "")
	fs.Func("allow", "", func(list string) error {
		for _, s := range strings.Split(list, ",") {
			code, ok := diagnostics.ParseCode(s)
			if !ok || !strings.HasPrefix(string(code), "WARN-") {
				w.err = fmt.Errorf("--allow takes warning codes such as %s, not %q", diagnostics.WarnUnusedVar, strings.TrimSpace(s))
				return nil
			}
			w.Allow[code] = true
		}
		return nil
	})
	return fs, w
}

// usageError reports err, or prints the usage if err is nil.
func usageError(stderr io.Writer, err error) {
	if err != nil {
		fmt.Fprintf(stderr, "runescript: %v\n", err)
		return
	}
	fmt.Fprint(stderr, usage)
}

// input is the program whose entry file is path. Uses are resolved
// against the working directory. Any file extension is accepted.
func input(path string, notes io.Writer) (driver.Input, error) {
	dir, err := os.Getwd()
	return driver.Input{Path: path, Dir: dir, Notes: notes}, err
}

// buildDetails describes the build from what the Go toolchain recorded:
// " (commit 9609a35, go1.25.11 windows/amd64)" for a build from a checkout,
// or the module version for `go install ...@version`.
func buildDetails() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	settings := map[string]string{}
	for _, s := range info.Settings {
		settings[s.Key] = s.Value
	}
	var parts []string
	switch rev := settings["vcs.revision"]; {
	case rev != "":
		if len(rev) > 7 {
			rev = rev[:7]
		}
		if settings["vcs.modified"] == "true" {
			rev += ", modified"
		}
		parts = append(parts, "commit "+rev)
	case info.Main.Version != "" && info.Main.Version != "(devel)":
		parts = append(parts, info.Main.Version)
	}
	parts = append(parts, goruntime.Version()+" "+goruntime.GOOS+"/"+goruntime.GOARCH)
	return " (" + strings.Join(parts, ", ") + ")"
}

// build implements `runescript build`.
func build(args []string, stderr io.Writer) int {
	fs, policy := flags("build", stderr)
	target := fs.String("target", "js", "")
	out := fs.String("o", "", "")
	if err := fs.Parse(args); err != nil || fs.NArg() != 1 || policy.err != nil {
		usageError(stderr, policy.err)
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
	in, err := input(path, stderr)
	in.Warnings = policy.WarningPolicy
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
