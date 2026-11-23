package webbuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildStaticDocs(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "static-docs")
	dist := filepath.Join(t.TempDir(), "dist")
	if ds, err := Build(root, dist, Options{SourceMap: true}); err != nil || len(ds) > 0 {
		t.Fatalf("build: %v %v", ds, err)
	}
	for name, fragments := range map[string][]string{
		"docs.html":                 {`<html lang="en">`, `href="/docs/getting-started"`},
		"docs/getting-started.html": {`href="#your-first-program"`, `id="your-first-program"`, `class="language-runescript"`, `Hello, RuneScript!`},
	} {
		b, err := os.ReadFile(filepath.Join(dist, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range append(fragments, `href="/docs.css"`) {
			if !strings.Contains(string(b), fragment) {
				t.Errorf("%s: missing %s", name, fragment)
			}
		}
		if strings.Contains(string(b), "<script") {
			t.Errorf("%s loads JavaScript", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dist, generated)); !os.IsNotExist(err) {
		t.Errorf("static docs emitted a runtime or JS entry: %v", err)
	}
}

func TestStaticDocsRejectInvalidFrontMatterAndRouteAliases(t *testing.T) {
	const original = "title: Getting started\ndescription: Docs\nsummary: Start here\nsection: Guide\norder: 1\nsource: guide.md"
	for _, test := range []struct {
		name, content, publicPath, want string
	}{
		{"quoted order", strings.Replace(original, "order: 1", `order: "1"`, 1), "", "order must be a signed 32-bit integer"},
		{"fractional order", strings.Replace(original, "order: 1", "order: 1.0", 1), "", "order must be a signed 32-bit integer"},
		{"wide order", strings.Replace(original, "order: 1", "order: 2147483648", 1), "", "order must be a signed 32-bit integer"},
		{"null title", strings.Replace(original, "title: Getting started", "title: null", 1), "", "title must be a non-empty string"},
		{"duplicate title", original + "\ntitle: Second", "", "expected a unique front-matter field name"},
		{"index alias", original, "docs/index.html", "docs route /docs conflicts"},
		{"chapter alias", original, "docs/getting-started/index.html", "docs route /docs/getting-started conflicts"},
		{"case alias", original, "DOCS.HTML", "docs route /docs conflicts"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			files := map[string]string{
				Manifest:                     `{"docs":{"content":"content","layout":"layout.html","title":"Docs","description":"Guide"}}`,
				"layout.html":                "{{.Content}}",
				"content/getting-started.md": "---\n" + test.content + "\n---\n## Install\n",
				"public/style.css":           "",
			}
			if test.publicPath != "" {
				files[filepath.ToSlash(filepath.Join("public", test.publicPath))] = ""
			}
			for name, content := range files {
				if err := writeOutput(root, name, []byte(content)); err != nil {
					t.Fatal(err)
				}
			}
			dist := filepath.Join(root, "dist")
			if ds, err := Build(root, dist, Options{}); err == nil || !strings.Contains(err.Error(), test.want) || len(ds) != 0 {
				t.Fatalf("build: %v %v, want %q", ds, err, test.want)
			}
			if _, err := os.Stat(dist); !os.IsNotExist(err) {
				t.Fatalf("failed build published output: %v", err)
			}
		})
	}
}
