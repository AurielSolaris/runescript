package webbuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildMultiPage(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "multi-page")
	dist := filepath.Join(t.TempDir(), "dist")
	ds, err := Build(root, dist, Options{SourceMap: true, ExtractCSS: true})
	if err != nil || len(ds) > 0 {
		t.Fatalf("build: %v %v", ds, err)
	}
	read := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(dist, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	mods, err := filepath.Glob(filepath.Join(dist, generated, "runtime-*.js"))
	if err != nil || len(mods) != 1 {
		t.Fatalf("expected one shared runtime, got %v (%v)", mods, err)
	}
	styles, err := filepath.Glob(filepath.Join(dist, generated, "styles-*.css"))
	if err != nil || len(styles) != 1 {
		t.Fatalf("expected one shared stylesheet, got %v (%v)", styles, err)
	}
	style := generated + "/" + filepath.Base(styles[0])
	for _, tc := range []struct{ page, prefix, body, runtimePrefix string }{
		{"index", "./", "Home", "../"},
		{"about/index", "../", "About", "../../"},
	} {
		t.Run(tc.page, func(t *testing.T) {
			base := generated + "/pages/" + tc.page
			html := read(tc.page + ".html")
			if !strings.Contains(html, tc.prefix+base+".js") ||
				!strings.Contains(html, tc.prefix+style) || strings.Contains(html, marker) {
				t.Fatalf("HTML assets were not linked correctly: %s", html)
			}
			js, css := read(base+".js"), read(style)
			if !strings.Contains(js, tc.runtimePrefix+filepath.Base(mods[0])) ||
				!strings.Contains(js, tc.body) || !strings.Contains(js, "data-runescript-style") ||
				strings.Contains(js, "readStdin") ||
				strings.Contains(js, "background-color: #edf4ff") ||
				!strings.Contains(css, "background-color: #edf4ff") {
				t.Fatal("JS and CSS were not separated as expected")
			}
			read(base + ".js.map")
		})
	}
	if _, err := Build(root, dist, Options{}); err == nil {
		t.Fatal("an existing output directory was overwritten")
	}
}

func TestFailedPageDoesNotPublishPartialSite(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		Manifest:             `{"pages":{"index.html":"ok.rsc","second.html":"bad.rsc"}}`,
		"public/index.html":  "<head>" + marker + "</head>",
		"public/second.html": "<head>" + marker + "</head>",
		"ok.rsc":             "package app; protected int main() { return 0; }",
		"bad.rsc":            "package app; protected int main() { return missing; }",
	}
	if err := os.Mkdir(filepath.Join(dir, "public"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(name)), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(dir, "dist")
	ds, err := Build(dir, out, Options{})
	if err != nil || len(ds) == 0 {
		t.Fatalf("invalid page: %v %v", ds, err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("failed build published output: %v", err)
	}
}
