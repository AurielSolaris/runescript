package tests

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/jsgen"
)

// RuneScript is a scripting language first; web support (rune:dom, the dev
// server) is a layer on top of it. These tests keep it that way: the web
// layer lives in known files, the scripting core never depends on it, and
// a script compiled to JavaScript carries none of its runtime.

const module = "github.com/AurielSolaris/runescript/"

// webPackages are the packages that exist only for web support.
var webPackages = []string{"internal/dom", "internal/devserver"}

// webImporters are the only files allowed to import a web package. Each is
// the rune:dom part of its stage, or the CLI wiring of `runescript dev`.
var webImporters = map[string]bool{
	"internal/parser/markup.go":       true,
	"internal/checker/checker.go":     true, // Info's template tables
	"internal/checker/dom.go":         true,
	"internal/checker/template.go":    true,
	"internal/ir/ir.go":               true, // template IR nodes
	"internal/ir/lower.go":            true,
	"internal/ir/verify.go":           true,
	"internal/jsgen/jsgen.go":         true,
	"internal/devserver/devserver.go": true,
	"cmd/runescript/watch.go":         true, // runescript dev
}

func TestWebSupportStaysALayer(t *testing.T) {
	files, err := filepath.Glob("../*/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	more, _ := filepath.Glob("../*/*.go")
	files = append(files, more...)
	fset := token.NewFileSet()
	for _, path := range files {
		rel := strings.TrimPrefix(filepath.ToSlash(path), "../")
		if strings.HasSuffix(rel, "_test.go") || strings.HasPrefix(rel, "tests/") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, im := range f.Imports {
			p, _ := strconv.Unquote(im.Path.Value)
			for _, web := range webPackages {
				if p == module+web && !webImporters[rel] && !strings.HasPrefix(rel, web+"/") {
					t.Errorf("%s imports %s; web support must stay in the files listed in webImporters", rel, web)
				}
			}
		}
	}
	for file := range webImporters {
		if _, err := os.Stat("../" + file); err != nil {
			t.Errorf("webImporters lists %s, which does not exist", file)
		}
	}
}

// TestScriptsCarryNoWebRuntime checks that the JavaScript runtime's
// optional sections (extern, dom, string) appear only in programs that
// use them.
func TestScriptsCarryNoWebRuntime(t *testing.T) {
	cases := []struct {
		name, src string
		want      []string // sections that must be present
	}{
		{"plain script", "package p;\npublic void main() { trace(1 + 2); }", nil},
		{"std:string", "package p;\nuse \"std:string\";\npublic void main() { trace(string.toUpper(\"a\")); }", []string{"string"}},
		{"extern", "package p;\nextern \"./x.mjs\" int f();\npublic void main() { trace(f()); }", []string{"extern"}},
		{"rune:dom", "package p;\nuse \"rune:dom\";\npublic Element app() {\n    template { <p>hi</p> }\n}\npublic void main() { dom.renderApp(app, \"#app\"); }", []string{"dom"}},
	}
	markers := map[string]string{
		"extern": "---- extern (JavaScript) functions",
		"dom":    "---- runes and the DOM",
		"string": "---- std:string ----",
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, diags, err := driver.BuildJS("p.rsc", []byte(c.src), jsgen.Options{})
			if err != nil || len(diags) > 0 {
				t.Fatalf("build failed: %v %v", err, diags)
			}
			js := string(out)
			if strings.Contains(js, "//@section") || strings.Contains(js, "//@end") {
				t.Error("section markers left in the output")
			}
			for sec, marker := range markers {
				want := false
				for _, w := range c.want {
					want = want || w == sec
				}
				if got := strings.Contains(js, marker); got != want {
					t.Errorf("section %s: present=%v, want %v", sec, got, want)
				}
			}
		})
	}
}
