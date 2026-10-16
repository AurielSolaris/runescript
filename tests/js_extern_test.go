package tests

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/jsgen"
)

// TestJSExterns runs the programs in testdata/extern with Node.js. Each
// is compiled to prog.mjs in an empty directory, next to copies of the
// JavaScript helper modules there, and run with that directory as the
// working directory. Its output must match the .out file and, if there
// is a .err file, the first line of stderr must match it.
func TestJSExterns(t *testing.T) {
	nodePath := node(t)
	programs, _ := filepath.Glob("testdata/extern/*.rsc")
	helpers, _ := filepath.Glob("testdata/extern/*.mjs")
	if len(programs) == 0 || len(helpers) == 0 {
		t.Fatal("no programs found")
	}
	for _, path := range programs {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, h := range helpers {
				if err := os.WriteFile(filepath.Join(dir, filepath.Base(h)), read(t, h), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			js, diags, err := driver.BuildJS("prog.rsc", read(t, path), jsgen.Options{Version: "test"})
			if diags != nil || err != nil {
				t.Fatalf("BuildJS: diags=%v err=%v", diags, err)
			}
			if err := os.WriteFile(filepath.Join(dir, "prog.mjs"), js, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "harness.mjs"), []byte(harness), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(nodePath, "harness.mjs", "prog.mjs")
			cmd.Dir = dir
			var stdout, stderr strings.Builder
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			code := 0
			var exit *exec.ExitError
			if err := cmd.Run(); errors.As(err, &exit) {
				code = exit.ExitCode()
			} else if err != nil {
				t.Fatal(err)
			}

			base := strings.TrimSuffix(path, ".rsc")
			if want := normalize(string(read(t, base+".out"))); normalize(stdout.String()) != want {
				t.Errorf("stdout:\n%s\nwant:\n%s", stdout.String(), want)
			}
			wantErr, err := os.ReadFile(base + ".err")
			gotErr, _, _ := strings.Cut(normalize(stderr.String()), "\n")
			switch {
			case err == nil:
				if gotErr != strings.TrimSpace(normalize(string(wantErr))) || code != 1 {
					t.Errorf("exit %d, stderr %q, want exit 1 and %q", code, gotErr, strings.TrimSpace(string(wantErr)))
				}
			case code != 0 || gotErr != "":
				t.Errorf("exit %d, stderr %q", code, stderr.String())
			}
		})
	}
}
