package tests

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/driver"
)

// The benchmark suite (bench/, run with `go run ./bench/rsbench`) compares
// RuneScript with TypeScript 1.0. These tests keep its RuneScript programs
// honest under plain `go test`: every benchmark is well formed, and at its
// "test" size prints expected/test.out on the interpreter and on the
// JavaScript target. The TypeScript side is checked by `rsbench verify`,
// which needs the pinned compiler from `npm install` in bench/.

type benchManifest struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Categories []string          `json:"categories"`
	Sizes      map[string]string `json:"sizes"`
	Notes      string            `json:"notes"`
}

func benchDirs(t *testing.T) []string {
	t.Helper()
	dirs, _ := filepath.Glob("../bench/suite/*")
	if len(dirs) == 0 {
		t.Fatal("no benchmarks found in bench/suite")
	}
	return dirs
}

func readManifest(t *testing.T, dir string) benchManifest {
	t.Helper()
	var m benchManifest
	if err := json.Unmarshal(read(t, filepath.Join(dir, "bench.json")), &m); err != nil {
		t.Fatalf("bench.json: %v", err)
	}
	return m
}

func TestBenchSuiteWellFormed(t *testing.T) {
	ids := map[string]string{}
	for _, dir := range benchDirs(t) {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			m := readManifest(t, dir)
			if m.ID == "" || m.Name == "" || len(m.Categories) == 0 {
				t.Fatal("bench.json needs id, name, and categories")
			}
			if other, dup := ids[m.ID]; dup {
				t.Errorf("id %s is also used by %s", m.ID, other)
			}
			ids[m.ID] = dir
			if m.Sizes["test"] == "" {
				t.Error(`bench.json needs a "test" size`)
			}
			for _, c := range m.Categories {
				switch c {
				case "runtime", "size", "conformance":
				default:
					t.Errorf("unknown category %q", c)
				}
				if c == "runtime" && m.Sizes["small"] == "" {
					t.Error(`a runtime benchmark needs a "small" size`)
				}
			}
			for _, f := range []string{"main.rsc", "main.ts", "expected/test.out"} {
				if !fileExists(filepath.Join(dir, f)) {
					t.Errorf("missing %s", f)
				}
			}
			// A TypeScript-specific expected output records an intended
			// divergence, which the notes must explain.
			if fileExists(filepath.Join(dir, "expected", "test.ts.out")) && m.Notes == "" {
				t.Error("expected/test.ts.out needs notes in bench.json explaining the divergence")
			}
			// TypeScript 1.0 sources: no ES2015 syntax the baseline compiler rejects.
			ts, _ := filepath.Glob(filepath.Join(dir, "*.ts"))
			for _, f := range ts {
				src := string(read(t, f))
				for _, bad := range []string{"let ", "const ", "`"} {
					if strings.Contains(src, bad) {
						t.Errorf("%s uses %q, which TypeScript 1.0 does not support", filepath.Base(f), bad)
					}
				}
			}
		})
	}
}

func TestBenchProgramsInterpreter(t *testing.T) {
	for _, dir := range benchDirs(t) {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			m := readManifest(t, dir)
			want := normalize(string(read(t, filepath.Join(dir, "expected", "test.out"))))
			got := runInterpreterIn(t, read(t, filepath.Join(dir, "main.rsc")), []string{m.Sizes["test"]}, "")
			if got.code != 0 || got.stdout != want {
				t.Errorf("exit %d, stderr %q\ngot:\n%s\nwant:\n%s", got.code, got.stderr, got.stdout, want)
			}
		})
	}
}

func TestBenchProgramsJS(t *testing.T) {
	nodePath := node(t)
	for _, dir := range benchDirs(t) {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			m := readManifest(t, dir)
			want := normalize(string(read(t, filepath.Join(dir, "expected", "test.out"))))
			got := runModuleIn(t, nodePath, read(t, filepath.Join(dir, "main.rsc")), []string{m.Sizes["test"]}, "")
			if got.code != 0 || got.stdout != want {
				t.Errorf("exit %d, stderr %q\ngot:\n%s\nwant:\n%s", got.code, got.stderr, got.stdout, want)
			}
		})
	}
}

// TestBenchDiagnostics checks that each invalid program in
// bench/diagnostics is rejected, with its first error on rs_line.
func TestBenchDiagnostics(t *testing.T) {
	dirs, _ := filepath.Glob("../bench/diagnostics/*")
	if len(dirs) == 0 {
		t.Fatal("no cases found in bench/diagnostics")
	}
	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			var c struct {
				RSLine int `json:"rs_line"`
			}
			if err := json.Unmarshal(read(t, filepath.Join(dir, "bench.json")), &c); err != nil {
				t.Fatal(err)
			}
			abs, err := filepath.Abs(dir)
			if err != nil {
				t.Fatal(err)
			}
			_, diags, err := driver.CheckInput(driver.Input{Path: filepath.Join(abs, "bad.rsc"), Dir: abs})
			if err != nil {
				t.Fatal(err)
			}
			if len(diags) == 0 {
				t.Fatal("bad.rsc was accepted")
			}
			if line := diags[0].Span.Start.Line; line != c.RSLine {
				t.Errorf("first error on line %d, want %d: %s", line, c.RSLine, diags[0].Error())
			}
		})
	}
}
