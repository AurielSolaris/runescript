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
	"github.com/AurielSolaris/runescript/internal/runtime"
)

// The JavaScript target must behave exactly like the interpreter. These
// tests compile programs to JavaScript, run them with Node.js, and compare
// stdout, the exit code, and any runtime error (code, position, message)
// with the interpreter's.
//
// They are skipped when `node` is not on PATH, unless RUNESCRIPT_REQUIRE_NODE
// is set, in which case a missing Node.js fails the test run.

// harness runs a generated module in a worker thread with a large stack, so
// that the host stack never runs out before the language's call-depth limit
// and RUN-001 is reported exactly where the interpreter reports it.
const harness = `import { Worker, isMainThread, parentPort, workerData } from "node:worker_threads";
import { pathToFileURL } from "node:url";
if (isMainThread) {
  const w = new Worker(new URL(import.meta.url), {
    workerData: pathToFileURL(process.argv[2]).href,
    resourceLimits: { stackSizeMb: 256 },
  });
  w.on("message", (m) => {
    process.stdout.write(m.out);
    if (m.err) process.stderr.write(m.err + "\n");
    process.exitCode = m.code;
  });
} else {
  const { run, RuneScriptError } = await import(workerData);
  let out = "";
  try {
    const code = run({
      trace: (line) => { out += line + "\n"; },
      cwd: () => process.cwd(),
      platform: process.platform,
      fs: await import("node:fs"),
    });
    parentPort.postMessage({ out, code });
  } catch (e) {
    if (!(e instanceof RuneScriptError)) throw e;
    parentPort.postMessage({ out, err: e.message, code: 1 });
  }
}
`

func node(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("node")
	if err != nil {
		if os.Getenv("RUNESCRIPT_REQUIRE_NODE") != "" {
			t.Fatal("node is not on PATH and RUNESCRIPT_REQUIRE_NODE is set")
		}
		t.Skip("node is not on PATH; skipping JavaScript target tests")
	}
	return path
}

type result struct {
	out  string
	code int
	err  string // "line:col: error CODE: message", or ""
}

func interpret(t *testing.T, src []byte) result {
	t.Helper()
	var out strings.Builder
	code, diags, err := driver.Run(src, &out)
	if diags != nil {
		t.Fatalf("compile errors: %v", diags)
	}
	r := result{out: out.String(), code: code}
	var re *runtime.Error
	if errors.As(err, &re) {
		r.err, r.code = re.Diag.Error(), 1
	} else if err != nil {
		t.Fatal(err)
	}
	return r
}

// compileJS compiles src and writes the module to a temporary directory.
func compileJS(t *testing.T, src []byte) string {
	t.Helper()
	js, diags, err := driver.BuildJS("prog.rsc", src, jsgen.Options{Version: "test"})
	if diags != nil || err != nil {
		t.Fatalf("BuildJS: diags=%v err=%v", diags, err)
	}
	path := filepath.Join(t.TempDir(), "prog.mjs")
	if err := os.WriteFile(path, js, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// runNode runs args with node and returns stdout, the first line of
// stderr, and the exit code.
func runNode(t *testing.T, nodePath string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(nodePath, args...)
	var o, e strings.Builder
	cmd.Stdout, cmd.Stderr = &o, &e
	err := cmd.Run()
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		code = exit.ExitCode()
	case err != nil:
		t.Fatal(err)
	}
	stderr, _, _ = strings.Cut(normalize(e.String()), "\n")
	return normalize(o.String()), stderr, code
}

func runJS(t *testing.T, nodePath string, src []byte) result {
	t.Helper()
	module := compileJS(t, src)
	h := filepath.Join(filepath.Dir(module), "harness.mjs")
	if err := os.WriteFile(h, []byte(harness), 0o644); err != nil {
		t.Fatal(err)
	}
	out, stderr, code := runNode(t, nodePath, h, module)
	return result{out: out, code: code, err: strings.TrimPrefix(stderr, "prog.rsc:")}
}

func TestJSTargetMatchesInterpreter(t *testing.T) {
	nodePath := node(t)
	var files []string
	for _, glob := range []string{"../examples/*.rsc", "testdata/runtime/*.rsc", "testdata/js/*.rsc"} {
		matches, _ := filepath.Glob(glob)
		for _, m := range matches {
			if !strings.HasSuffix(m, "_error.rsc") {
				files = append(files, m)
			}
		}
	}
	if len(files) == 0 {
		t.Fatal("no programs found")
	}
	for _, path := range files {
		t.Run(filepath.ToSlash(path), func(t *testing.T) {
			t.Parallel()
			src := read(t, path)
			want := interpret(t, src)
			got := runJS(t, nodePath, src)
			if got != want {
				t.Errorf("JavaScript target differs from the interpreter\ngot:  %+v\nwant: %+v", got, want)
			}
		})
	}
}

// TestJSPrograms checks testdata/js programs against their expected output
// on the JavaScript target (the interpreter is held to the same output by
// TestJSTargetMatchesInterpreter).
func TestJSPrograms(t *testing.T) {
	nodePath := node(t)
	files, _ := filepath.Glob("testdata/js/*.rsc")
	if len(files) == 0 {
		t.Fatal("no programs found")
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			got := runJS(t, nodePath, read(t, path))
			want := normalize(string(read(t, strings.TrimSuffix(path, ".rsc")+".out")))
			if got.out != want || got.err != "" {
				t.Errorf("got output:\n%s\nerror: %s\nwant:\n%s", got.out, got.err, want)
			}
		})
	}
}

// TestJSHostStackExhaustion runs a recursive program on a deliberately
// small host stack: the host runs out before the language's call-depth
// limit, and the program must still stop with RUN-001, not a raw
// JavaScript RangeError.
func TestJSHostStackExhaustion(t *testing.T) {
	nodePath := node(t)
	src := []byte(`package p;
public int down(int n) {
    if (n == 0) {
        return 0;
    }
    return down(n - 1) + 1;
}
protected int main() {
    trace("start");
    trace(down(9000));
    return 0;
}
`)
	if got := interpret(t, src); got.out != "start\n9000\n" {
		t.Fatalf("interpreter: %+v", got)
	}
	out, stderr, code := runNode(t, nodePath, "--stack-size=150", compileJS(t, src))
	if out != "start\n" || code != 1 {
		t.Errorf("stdout %q, exit %d", out, code)
	}
	if !strings.HasPrefix(stderr, "prog.rsc:6:12: error ASX-RUN-001: maximum call depth exceeded: the JavaScript host ran out of stack") {
		t.Errorf("stderr %q", stderr)
	}
}

// TestJSDirectRun checks that a generated module runs on its own with
// `node <module>`, setting the exit code from main.
func TestJSDirectRun(t *testing.T) {
	nodePath := node(t)
	module := compileJS(t, []byte("package p;\nprotected int main() {\n    trace(\"direct\", 1, true);\n    return 4;\n}\n"))
	out, stderr, code := runNode(t, nodePath, module)
	if out != "direct 1 true\n" || stderr != "" || code != 4 {
		t.Errorf("stdout %q, stderr %q, exit %d", out, stderr, code)
	}
}
