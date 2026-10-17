package tests

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

// Programs in testdata/env use the program's environment: std:io,
// std:process, std:fs, and std:path. Each runs in a new, empty working
// directory, with the arguments of its `// args:` comment, the standard
// input of its `// stdin:` comment (with \n and \r for line ends), and the
// environment variable RUNESCRIPT_TEST=yes. It runs in the interpreter and
// as a Node.js module run directly (`node prog.mjs args...`), which uses
// the module's own wiring of Node's streams, arguments, and fs. Both must
// print the same standard output and standard error and exit alike, and
// standard output must match <name>.out (and standard error <name>.err,
// if present).

var (
	argsRE  = regexp.MustCompile(`(?m)^// args: (.*)$`)
	stdinRE = regexp.MustCompile(`(?m)^// stdin: (.*)$`)
)

type envRun struct {
	stdout, stderr string
	code           int
}

func TestEnvironmentPrograms(t *testing.T) {
	nodePath := node(t)
	files, _ := filepath.Glob("testdata/env/*.rsc")
	if len(files) == 0 {
		t.Fatal("no programs in testdata/env")
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			src := read(t, path)
			var args []string
			if m := argsRE.FindSubmatch(src); m != nil {
				args = strings.Fields(string(m[1]))
			}
			stdin := ""
			if m := stdinRE.FindSubmatch(src); m != nil {
				stdin = strings.NewReplacer(`\n`, "\n", `\r`, "\r").Replace(string(m[1]))
			}
			base, _ := filepath.Abs(strings.TrimSuffix(path, ".rsc"))
			t.Setenv("RUNESCRIPT_TEST", "yes")
			interp := runInterpreterIn(t, src, args, stdin)
			js := runModuleIn(t, nodePath, src, args, stdin)
			if interp != js {
				t.Fatalf("the targets differ\ninterpreter: %+v\nnode:        %+v", interp, js)
			}
			want := string(read(t, base+".out"))
			if interp.stdout != normalize(want) {
				t.Errorf("stdout:\n%s\nwant:\n%s", interp.stdout, want)
			}
			if errFile := base + ".err"; fileExists(errFile) {
				if want := normalize(string(read(t, errFile))); interp.stderr != want {
					t.Errorf("stderr:\n%s\nwant:\n%s", interp.stderr, want)
				}
			}
		})
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// runInterpreterIn runs src in the interpreter, in a new working
// directory.
func runInterpreterIn(t *testing.T, src []byte, args []string, stdin string) envRun {
	t.Helper()
	t.Chdir(t.TempDir())
	var out, errOut strings.Builder
	in := driver.Memory("prog.rsc", src)
	in.Args, in.Stdin, in.Stderr = args, strings.NewReader(stdin), &errOut
	code, diags, err := driver.RunInput(in, &out)
	if diags != nil {
		t.Fatalf("compile errors: %v", diags)
	}
	var re *runtime.Error
	if errors.As(err, &re) {
		errOut.WriteString(re.Diag.Error() + "\n")
		code = 1
	} else if err != nil {
		t.Fatal(err)
	}
	return envRun{normalize(out.String()), normalize(errOut.String()), code}
}

// runModuleIn compiles src and runs the module with node, in a new working
// directory.
func runModuleIn(t *testing.T, nodePath string, src []byte, args []string, stdin string) envRun {
	t.Helper()
	module := compileJS(t, src)
	cmd := exec.Command(nodePath, append([]string{module}, args...)...)
	cmd.Dir = t.TempDir()
	cmd.Stdin = strings.NewReader(stdin)
	var o, e strings.Builder
	cmd.Stdout, cmd.Stderr = &o, &e
	code := 0
	err := cmd.Run()
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		code = exit.ExitCode()
	case err != nil:
		t.Fatal(err)
	}
	return envRun{normalize(o.String()), normalize(e.String()), code}
}
