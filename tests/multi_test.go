package tests

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/jsgen"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

// Each directory in testdata/multi is a program whose entry file is
// main.rsc, run with the directory as the working directory.
//
//   - If its files contain `// expect: CODE file:line:col` comments, the
//     program must fail to compile with exactly those diagnostics.
//   - Otherwise it must print expect.out. A `// expect runtime error: CODE
//     file:line:col` comment means it must then stop with that error.
//
// Programs that compile are also run on the JavaScript target, which must
// behave identically.

var (
	multiExpectRE  = regexp.MustCompile(`// expect: (\S+) (\S+:\d+:\d+)`)
	multiRuntimeRE = regexp.MustCompile(`// expect runtime error: (\S+) (\S+:\d+:\d+)`)
)

func TestMultiFilePrograms(t *testing.T) {
	dirs, _ := filepath.Glob("testdata/multi/*")
	if len(dirs) == 0 {
		t.Fatal("no programs found")
	}
	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			compileErrs, runtimeErr := expectations(t, dir)
			in := driver.Input{Path: filepath.Join(dir, "main.rsc"), Dir: dir}
			if compileErrs != nil {
				_, diags, err := driver.CheckInput(in)
				if err != nil {
					t.Fatal(err)
				}
				var got []string
				for _, d := range diags {
					got = append(got, d.Code.String()+" "+where(d))
				}
				sort.Strings(got)
				if strings.Join(got, "\n") != strings.Join(compileErrs, "\n") {
					t.Errorf("diagnostics mismatch\ngot:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(compileErrs, "\n"))
				}
				return
			}
			want := result{out: normalize(string(read(t, filepath.Join(dir, "expect.out"))))}
			if runtimeErr != "" {
				want.code = 1
			}
			got := interpretInput(t, in)
			if got.out != want.out || got.code != want.code || (runtimeErr != "") != (got.err != "") ||
				(runtimeErr != "" && !strings.HasPrefix(got.err, runtimeErr)) {
				t.Fatalf("interpreter: got %+v, want output %q, error %q", got, want.out, runtimeErr)
			}
			if js := runJSInput(t, in); js != got {
				t.Errorf("JavaScript target differs from the interpreter\ngot:  %+v\nwant: %+v", js, got)
			}
		})
	}
}

// expectations collects the expectation comments of every file under dir.
// compileErrs is sorted, as "CODE file:line:col"; runtimeErr is
// "file:line:col: error CODE".
func expectations(t *testing.T, dir string) (compileErrs []string, runtimeErr string) {
	t.Helper()
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		src := string(read(t, path))
		for _, m := range multiExpectRE.FindAllStringSubmatch(src, -1) {
			compileErrs = append(compileErrs, m[1]+" "+m[2])
		}
		if m := multiRuntimeRE.FindStringSubmatch(src); m != nil {
			runtimeErr = m[2] + ": error " + m[1]
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(compileErrs)
	return compileErrs, runtimeErr
}

func where(d *diagnostics.Diagnostic) string {
	return d.Span.Start.File.Name + ":" + d.Span.Start.String()
}

// interpretInput runs a program with the interpreter. A runtime error is
// recorded as "file:line:col: error CODE: message".
func interpretInput(t *testing.T, in driver.Input) result {
	t.Helper()
	var out strings.Builder
	code, diags, err := driver.RunInput(in, &out)
	if diags != nil {
		t.Fatalf("compile errors: %v", diags)
	}
	r := result{out: out.String(), code: code}
	var re *runtime.Error
	if errors.As(err, &re) {
		r.err, r.code = re.Diag.Span.Start.File.Name+":"+re.Diag.Error(), 1
	} else if err != nil {
		t.Fatal(err)
	}
	return r
}

// runJSInput compiles a program to JavaScript and runs it with Node.js.
func runJSInput(t *testing.T, in driver.Input) result {
	t.Helper()
	nodePath := node(t)
	js, diags, err := driver.BuildJSInput(in, jsgen.Options{Version: "test"})
	if diags != nil || err != nil {
		t.Fatalf("BuildJS: diags=%v err=%v", diags, err)
	}
	tmp := t.TempDir()
	module := filepath.Join(tmp, "prog.mjs")
	h := filepath.Join(tmp, "harness.mjs")
	if err := os.WriteFile(module, js, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(h, []byte(harness), 0o644); err != nil {
		t.Fatal(err)
	}
	out, stderr, code := runNode(t, nodePath, h, module)
	return result{out: out, code: code, err: stderr}
}

// Paths and package names are relative to the working directory, not to
// the entry file: the same program run from its parent directory cannot
// find its packages.
func TestUsesAreRelativeToTheWorkingDirectory(t *testing.T) {
	parent := filepath.Join("testdata", "multi")
	in := driver.Input{Path: filepath.Join(parent, "ok_packages", "main.rsc"), Dir: parent}
	_, diags, err := driver.CheckInput(in)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, d := range diags {
		got = append(got, d.Code.String()+" "+where(d))
	}
	want := []string{
		"ASX-USE-004 ok_packages/main.rsc:5:5",    // use "geom";
		"ASX-USE-004 ok_packages/main.rsc:6:5",    // use "./util/text.txt" as t;
		"ASX-USE-004 ok_packages/main.rsc:7:5",    // use "./helpers.rsc";
		"ASX-NAME-001 ok_packages/main.rsc:12:11", // twice, from helpers.rsc
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
