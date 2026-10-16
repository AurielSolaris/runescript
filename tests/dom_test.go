package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/jsgen"
)

// rune:dom programs run on Node.js against a small fake DOM
// (testdata/dom/fakedom.mjs). Each program in testdata/dom is compiled,
// mounted on <div id="app">, and its HTML printed; then every
// `// click: selector` comment in the program clicks that element, and
// the HTML is printed again. The output must match the .out file. A
// runtime error is printed as "error: message" and ends the run.

const domDriver = `import { install, serialize } from "./fakedom.mjs";
const document = install();
const app = document.createElement("div");
app.setAttribute("id", "app");
document.body.appendChild(app);
const { run } = await import("./prog.mjs");
const clicks = JSON.parse(process.argv[2]);
const show = (label) => console.log(label + ": " + serialize(app));
try {
  run();
  show("render");
  for (const sel of clicks) {
    const e = document.querySelector(sel);
    if (!e) throw new Error("no element matches " + sel);
    e.click();
    show("click " + sel);
  }
} catch (e) {
  console.log("error: " + (e.detail !== undefined ? e.code + ": " + e.detail : e.message));
}
`

var clickRE = regexp.MustCompile(`(?m)^// click: (\S+)\s*$`)

// buildDOM compiles a rune:dom program into dir, next to the fake DOM
// and the driver.
func buildDOM(t *testing.T, dir string, in driver.Input) {
	t.Helper()
	js, diags, err := driver.BuildJSInput(in, jsgen.Options{Version: "test"})
	if diags != nil || err != nil {
		t.Fatalf("BuildJS: diags=%v err=%v", diags, err)
	}
	files := map[string][]byte{"prog.mjs": js, "driver.mjs": []byte(domDriver), "fakedom.mjs": read(t, "testdata/dom/fakedom.mjs")}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func runDOM(t *testing.T, nodePath string, in driver.Input, src []byte) string {
	t.Helper()
	dir := t.TempDir()
	buildDOM(t, dir, in)
	var clicks []string
	for _, m := range clickRE.FindAllStringSubmatch(string(src), -1) {
		clicks = append(clicks, m[1])
	}
	arg, _ := json.Marshal(append([]string{}, clicks...))
	cmd := exec.Command(nodePath, "driver.mjs", string(arg))
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, out)
	}
	return normalize(string(out))
}

func TestDOMPrograms(t *testing.T) {
	nodePath := node(t)
	files, _ := filepath.Glob("testdata/dom/*.rsc")
	if len(files) == 0 {
		t.Fatal("no programs found")
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			src := read(t, path)
			got := runDOM(t, nodePath, driver.Memory("prog.rsc", src), src)
			want := normalize(string(read(t, strings.TrimSuffix(path, ".rsc")+".out")))
			if got != want {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

// TestWebExamples runs every example in examples/web on the fake DOM. An
// example is a folder whose entry file is main.rsc, or else app.rsc; it is
// built from its folder, so its uses resolve there, as when it is run
// with `runescript dev` from that folder. The output must match
// testdata/web/<folder>.out.
func TestWebExamples(t *testing.T) {
	nodePath := node(t)
	dirs, _ := filepath.Glob("../examples/web/*")
	if len(dirs) == 0 {
		t.Fatal("no web examples found")
	}
	for _, dir := range dirs {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			entry := filepath.Join(dir, "main.rsc")
			if _, err := os.Stat(entry); err != nil {
				entry = filepath.Join(dir, "app.rsc")
			}
			abs, _ := filepath.Abs(dir)
			src := read(t, entry)
			in := driver.Input{Path: filepath.Join(abs, filepath.Base(entry)), Dir: abs}
			var warned []string
			in.Warnings.Report = func(ws []*diagnostics.Diagnostic) {
				for _, w := range ws {
					warned = append(warned, w.Error())
				}
			}
			got := runDOM(t, nodePath, in, src)
			if len(warned) > 0 {
				t.Errorf("the example has warnings: %v", warned)
			}
			want := normalize(string(read(t, "testdata/web/"+name+".out")))
			if got != want {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}
