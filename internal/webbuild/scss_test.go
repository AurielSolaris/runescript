package webbuild

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedSCSS(t *testing.T) {
	root := t.TempDir()
	for name, data := range map[string]string{
		Manifest:             `{"pages":{"index.html":"main.rsc"},"docs":{"content":"content","layout":"layout.html","title":"Docs","description":"Guide"},"styles":["styles/app.scss","styles/app.scss"]}`,
		"public/index.html":  "<head>" + marker + "</head>",
		"layout.html":        "<head>" + marker + "</head><article>{{.Content}}</article>",
		"main.rsc":           "package main; protected int main() { return 0; }",
		"content/intro.md":   "---\ntitle: Intro\ndescription: Guide\nsummary: Start here\nsection: Guide\norder: 1\nsource: intro.md\n---\n## Welcome\n",
		"styles/app.scss":    `@use "theme"; .card { @include theme.panel; &:hover { color: black; } }`,
		"styles/_theme.scss": `@use "sass:map"; $palette: (ink: #213547); @mixin panel { color: map.get($palette, ink); } .theme { color: black; }`,
	} {
		if err := writeOutput(root, name, []byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	for _, mapped := range []bool{false, true} {
		dist := filepath.Join(t.TempDir(), "dist")
		if ds, err := Build(root, dist, Options{SourceMap: mapped}); err != nil || len(ds) != 0 {
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
		css, err := filepath.Glob(filepath.Join(dist, generated, "scss-*.css"))
		if err != nil || len(css) != 1 {
			t.Fatalf("expected one CSS asset, got %v (%v)", css, err)
		}
		name := generated + "/" + filepath.Base(css[0])
		for _, page := range []string{"index.html", "docs.html", "docs/intro.html"} {
			if html := read(page); strings.Count(html, relativeURL(page, name)) != 1 {
				t.Errorf("%s does not link the shared stylesheet exactly once", page)
			}
		}
		source := read(name)
		if !strings.Contains(source, ".card:hover") || !strings.Contains(source, "#213547") {
			t.Fatalf("unexpected CSS %q", source)
		}
		_, mapName, hasMap := strings.Cut(source, "sourceMappingURL=")
		if hasMap != mapped {
			t.Fatalf("source map presence %t, want %t", hasMap, mapped)
		}
		if mapped {
			var sourceMap struct {
				Sources        []string
				SourcesContent []string
			}
			name := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(mapName), "*/"))
			if err := json.Unmarshal([]byte(read(generated+"/"+name)), &sourceMap); err != nil ||
				len(sourceMap.Sources) != 2 || len(sourceMap.SourcesContent) != 2 {
				t.Fatalf("SCSS source map: %+v (%v)", sourceMap, err)
			}
			for _, name := range sourceMap.Sources {
				if !strings.HasPrefix(name, "styles/") {
					t.Errorf("source map leaked a non-project-relative path: %q", name)
				}
			}
		}
	}
}

func TestSCSSFailureLeavesSiteUnpublished(t *testing.T) {
	root := t.TempDir()
	for name, data := range map[string]string{
		Manifest:            `{"pages":{"index.html":"main.rsc"},"styles":["broken.scss"]}`,
		"public/index.html": "<head>" + marker + "</head>",
		"broken.scss":       "a { color: $unknown; }",
	} {
		if err := writeOutput(root, name, []byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	dist := filepath.Join(root, "dist")
	if _, err := Build(root, dist, Options{}); err == nil || !strings.Contains(err.Error(), "broken.scss") || !strings.Contains(err.Error(), "Undefined variable") {
		t.Fatalf("expected a located Sass error, got %v", err)
	}
	if _, err := os.Stat(dist); !os.IsNotExist(err) {
		t.Fatalf("failed build published output: %v", err)
	}
}
