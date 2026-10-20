// Command rsbench runs the RuneScript benchmark and conformance suite:
// RuneScript on the interpreter (rs-vm), RuneScript compiled to JavaScript
// (rs-js), and the same programs written in TypeScript (ts, plus variants
// such as ts-checked), all on the same machine and inputs.
//
// Usage: go run ./bench/rsbench <command> [flags] [benchmark...]
//
// See bench/README.md for the commands, the layout of a benchmark, and the
// measurement protocol.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const usage = `usage: go run ./bench/rsbench <command> [flags] [benchmark...]

commands:
  list      list the benchmarks (filter with -category)
  verify    run every benchmark at its "test" size on every implementation
            and compare the output with expected/test.out (correctness)
  bless     write expected/test.out from the interpreter (review the diff!)
  run       runtime benchmarks: warm-up and measured runs, median/min/max,
            runtime ratios (category "runtime")
  size      generated JavaScript size: raw, minified, gzip, brotli
            (category "size")
  compile   compiler benchmarks on generated projects: runescript check and
            build vs tsc --noEmit and tsc
  diag      diagnostics: paired invalid programs, first-error location,
            error count, and both compilers' messages side by side

A benchmark is selected by its directory name, ID (e.g. A03), or name.
With none given, every benchmark in the command's category is used.
Run "go run ./bench/rsbench <command> -h" for a command's flags.
`

func main() {
	os.Exit(cli(os.Args[1:], os.Stdout, os.Stderr))
}

func cli(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	cmds := map[string]func(*flag.FlagSet, []string, io.Writer) error{
		"list":    cmdList,
		"verify":  cmdVerify,
		"bless":   cmdBless,
		"run":     cmdRun,
		"size":    cmdSize,
		"compile": cmdCompile,
		"diag":    cmdDiag,
	}
	run, ok := cmds[args[0]]
	if !ok {
		fmt.Fprintf(stderr, "rsbench: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
	fs := flag.NewFlagSet("rsbench "+args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := run(fs, args[1:], stdout); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		fmt.Fprintf(stderr, "rsbench %s: %v\n", args[0], err)
		return 1
	}
	return 0
}
