// Package tests holds end-to-end golden tests that run whole programs.
//
// Every examples/*.rsc file is either run and compared against
// testdata/examples/<name>.out, or, if its name ends in "_error", checked
// for the diagnostics listed in its `// expect: CODE line:col` comments.
// testdata/errors/*.rsc files follow the same "expect" convention, and
// testdata/runtime/*.rsc files must fail at run time (see
// TestRuntimeErrorPrograms).
package tests

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

var expectRE = regexp.MustCompile(`// expect: (\S+) (\d+:\d+)`)

func TestExamples(t *testing.T) {
	files, err := filepath.Glob("../examples/*.rsc")
	if err != nil || len(files) == 0 {
		t.Fatalf("no examples found: %v", err)
	}
	for _, path := range files {
		name := strings.TrimSuffix(filepath.Base(path), ".rsc")
		t.Run(name, func(t *testing.T) {
			src := read(t, path)
			if strings.HasSuffix(name, "_error") {
				expectDiagnostics(t, src)
				return
			}
			want := read(t, filepath.Join("testdata", "examples", name+".out"))
			var out strings.Builder
			code, diags, err := driver.Run(src, &out)
			if diags != nil || err != nil {
				t.Fatalf("diags=%v err=%v", diags, err)
			}
			if code != 0 {
				t.Errorf("exit code %d", code)
			}
			if got := out.String(); got != normalize(string(want)) {
				t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestErrorPrograms(t *testing.T) {
	files, _ := filepath.Glob("testdata/errors/*.rsc")
	if len(files) == 0 {
		t.Fatal("no error programs found")
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			expectDiagnostics(t, read(t, path))
		})
	}
}

func expectDiagnostics(t *testing.T, src []byte) {
	t.Helper()
	var want []string
	for _, m := range expectRE.FindAllStringSubmatch(string(src), -1) {
		want = append(want, m[1]+" "+m[2])
	}
	if len(want) == 0 {
		t.Fatal("no `// expect:` comments in error program")
	}
	_, diags := driver.Check(src)
	var got []string
	for _, d := range diags {
		got = append(got, d.Code.String()+" "+d.Span.Start.String())
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("diagnostics mismatch\ngot:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// normalize tolerates CRLF checkouts of the expected-output files.
func normalize(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }

var runtimeErrRE2 = regexp.MustCompile(`// expect runtime error: (ASX-RUN-\d+) (\d+:\d+)`)

// TestRuntimeErrorPrograms runs testdata/runtime/*.rsc. Each must compile,
// print exactly <name>.out, then stop with the runtime error named in its
// `// expect runtime error: CODE line:col` comment.
func TestRuntimeErrorPrograms(t *testing.T) {
	files, _ := filepath.Glob("testdata/runtime/*.rsc")
	if len(files) == 0 {
		t.Fatal("no runtime error programs found")
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			src := read(t, path)
			m := runtimeErrRE2.FindStringSubmatch(string(src))
			if m == nil {
				t.Fatal("missing `// expect runtime error:` comment")
			}
			var out strings.Builder
			_, diags, err := driver.Run(src, &out)
			if diags != nil {
				t.Fatalf("compile errors: %v", diags)
			}
			var re *runtime.Error
			if !errors.As(err, &re) {
				t.Fatalf("expected runtime error %s, got %v", m[1], err)
			}
			if got := re.Diag.Code.String() + " " + re.Diag.Span.Start.String(); got != m[1]+" "+m[2] {
				t.Errorf("got %s, want %s %s", got, m[1], m[2])
			}
			want := normalize(string(read(t, strings.TrimSuffix(path, ".rsc")+".out")))
			if out.String() != want {
				t.Errorf("output before the error:\ngot:\n%s\nwant:\n%s", out.String(), want)
			}
		})
	}
}
