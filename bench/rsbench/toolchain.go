package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TSFlags are the flags every TypeScript program is compiled with. The
// baseline is TypeScript 1.0 RC (npm typescript@0.9.7), which has no
// strict mode and no ES2015 output: ES5 and --noImplicitAny are the
// closest settings to RuneScript's checking.
var TSFlags = []string{"--target", "ES5", "--noImplicitAny", "--module", "commonjs"}

// Toolchain locates and describes the tools a run uses.
type Toolchain struct {
	Root      string // repository root
	BenchRoot string // Root/bench
	Work      string // scratch directory for this run

	RuneScript string // the runescript binary, built from Root
	Node       string
	TSC        string // bench/node_modules/typescript/bin/tsc.js, or ""
	Esbuild    bool   // bench/node_modules/esbuild is installed

	RSVersion, NodeVersion, TSVersion string
}

func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(data, []byte("module github.com/AurielSolaris/runescript\n")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("run rsbench inside the RuneScript repository")
		}
		dir = parent
	}
}

// setup builds the runescript CLI into a fresh scratch directory and
// finds Node.js and the pinned TypeScript compiler. needNode reports an
// error if node is missing; a missing TypeScript install is not an error
// (the TypeScript implementations are reported as unavailable).
func setup(needNode bool) (*Toolchain, error) {
	root, err := findRoot()
	if err != nil {
		return nil, err
	}
	work, err := os.MkdirTemp("", "rsbench-")
	if err != nil {
		return nil, err
	}
	tc := &Toolchain{Root: root, BenchRoot: filepath.Join(root, "bench"), Work: work}

	exe := filepath.Join(work, "runescript")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	build := exec.Command("go", "build", "-o", exe, "./cmd/runescript")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("building runescript: %v\n%s", err, out)
	}
	tc.RuneScript = exe
	tc.RSVersion = firstLine(output(exe, "version"))

	if p, err := exec.LookPath("node"); err == nil {
		tc.Node = p
		tc.NodeVersion = firstLine(output(p, "--version"))
	} else if needNode {
		return nil, errors.New("node is not on PATH")
	}
	tsc := filepath.Join(tc.BenchRoot, "node_modules", "typescript", "bin", "tsc.js")
	if tc.Node != "" && exists(tsc) {
		tc.TSC = tsc
		tc.TSVersion = strings.TrimPrefix(firstLine(output(tc.Node, tsc, "-v")), "Version ")
	}
	tc.Esbuild = exists(filepath.Join(tc.BenchRoot, "node_modules", "esbuild"))
	return tc, nil
}

func (tc *Toolchain) Close() { os.RemoveAll(tc.Work) }

// warnMissingTS explains how to install the TypeScript baseline.
func (tc *Toolchain) warnMissingTS(w func(string, ...any)) {
	if tc.TSC == "" {
		w("note: TypeScript is not installed; run `npm install` in bench/ to compare against TypeScript 1.0 RC\n")
	}
}

// CompileTS compiles one TypeScript source to dir/<name>.js and returns
// the output path.
func (tc *Toolchain) CompileTS(src, dir string) (string, error) {
	if tc.TSC == "" {
		return "", errors.New("TypeScript is not installed (npm install in bench/)")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	args := append([]string{tc.TSC}, TSFlags...)
	args = append(args, "--outDir", dir, src)
	cmd := exec.Command(tc.Node, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("tsc %s: %v\n%s", filepath.Base(src), err, out)
	}
	return filepath.Join(dir, strings.TrimSuffix(filepath.Base(src), ".ts")+".js"), nil
}

// BuildRS compiles a RuneScript program to a JavaScript module.
func (tc *Toolchain) BuildRS(src, out string) error {
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	cmd := exec.Command(tc.RuneScript, "build", "--color=never", "-o", out, src)
	if o, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("runescript build %s: %v\n%s", filepath.Base(src), err, o)
	}
	return nil
}

// Program is a ready-to-run implementation of a benchmark: a command
// whose last argument is the size.
type Program struct {
	Impl string
	Argv []string
	Dir  string
	Err  error // set when the implementation could not be prepared
}

// prepare compiles every implementation of b once and returns the
// commands that run them.
func (tc *Toolchain) prepare(b *Bench) []*Program {
	dir := filepath.Join(tc.Work, b.Key())
	src := filepath.Join(b.Dir, "main.rsc")
	var progs []*Program

	progs = append(progs, &Program{Impl: "rs-vm", Argv: []string{tc.RuneScript, "run", "--color=never", src}, Dir: b.Dir})

	rsjs := &Program{Impl: "rs-js"}
	out := filepath.Join(dir, "rs", "main.mjs")
	if tc.Node == "" {
		rsjs.Err = errors.New("node is not on PATH")
	} else if err := tc.BuildRS(src, out); err != nil {
		rsjs.Err = err
	} else {
		rsjs.Argv, rsjs.Dir = []string{tc.Node, out}, dir
	}
	progs = append(progs, rsjs)

	variants := b.TSVariants()
	for _, impl := range b.Impls()[2:] {
		p := &Program{Impl: impl}
		js, err := tc.CompileTS(variants[impl], filepath.Join(dir, impl))
		if err != nil {
			p.Err = err
		} else {
			p.Argv, p.Dir = []string{tc.Node, js}, dir
		}
		progs = append(progs, p)
	}
	return progs
}

// Sample is one process run.
type Sample struct {
	Wall    time.Duration
	PeakRSS int64 // bytes; -1 when the platform doesn't report it
	Stdout  string
	Stderr  string
	Code    int
}

// measure runs argv once in dir and times it from start to exit.
func measure(argv []string, dir string) (Sample, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	start := time.Now()
	err := cmd.Run()
	s := Sample{Wall: time.Since(start), PeakRSS: -1, Stdout: normalize(o.String()), Stderr: normalize(e.String())}
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		s.Code = exit.ExitCode()
	case err != nil:
		return s, err
	}
	if cmd.ProcessState != nil {
		s.PeakRSS = peakRSS(cmd.ProcessState)
	}
	return s, nil
}

func output(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "unknown"
	}
	return string(out)
}

func firstLine(s string) string {
	s, _, _ = strings.Cut(strings.TrimSpace(normalize(s)), "\n")
	return s
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
