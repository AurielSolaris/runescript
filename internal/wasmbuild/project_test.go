package wasmbuild

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"testing"

	"github.com/AurielSolaris/runescript/internal/loader"
)

func TestProjectArchiveContainsLoadedSourcesAndSelectedAssets(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"main.rsc":          "package app;\nuse \"geom\";\nprotected int main() { return geom.zero(); }\n",
		"geom/shape.rsm":    "package geom;\npublic semver version = 1.0.0;\npublic int zero() { return 0; }\n",
		"assets/sprite.png": "test image bytes",
		"unrelated.txt":     "leave out",
	}
	for name, text := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	prog, diags, err := loader.Load(filepath.Join(dir, "main.rsc"), dir)
	if err != nil || len(diags) != 0 {
		t.Fatalf("load: %v %v", diags, err)
	}
	var data bytes.Buffer
	tw := tar.NewWriter(&data)
	entry, err := writeProject(tw, prog, []string{"assets/sprite.png", "assets/sprite.png"})
	if err := errors.Join(err, tw.Close()); err != nil || entry != "main.rsc" {
		t.Fatalf("archive entry %q: %v", entry, err)
	}
	filesInTar := map[string]string{}
	tr := tar.NewReader(&data)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if _, found := filesInTar[h.Name]; found {
			t.Fatalf("duplicate archive path %q", h.Name)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		filesInTar[h.Name] = string(body)
	}
	delete(files, "unrelated.txt")
	if !maps.Equal(filesInTar, files) {
		t.Fatalf("archive files: %v", filesInTar)
	}
	for _, name := range []string{"..", "../unrelated.txt", "."} {
		if _, err := projectName(dir, name); err == nil {
			t.Errorf("invalid project path %q was accepted", name)
		}
	}
}
