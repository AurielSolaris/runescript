package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		path string
		kind Kind
		name string
	}{
		{"std:string", UseStd, "string"},
		{"geom", UsePackage, "geom"},
		{"org.x.geom", UsePackage, "org.x.geom"},
		{"geom.rsc", UsePackage, "geom.rsc"}, // a package name, and an invalid one
		{"./geom.rsc", UseFile, ""},
		{"lib/geom", UseFile, ""},
		{"../shared/x.txt", UseFile, ""},
		{"go:strings", UseUnsupported, ""},
	}
	for _, c := range cases {
		if kind, name := Classify(c.path); kind != c.kind || name != c.name {
			t.Errorf("Classify(%q) = %v, %q; want %v, %q", c.path, kind, name, c.kind, c.name)
		}
	}
}

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A file is loaded once however its path is spelled, and only .rsm files
// make up a package directory.
func TestFilesAreLoadedOnce(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "lib", "x.rsc"), "package lib;\n")
	write(t, filepath.Join(dir, "lib", "y.rsm"), "package lib;\n")
	write(t, filepath.Join(dir, "lib", "notes.rsc"), "this is not RuneScript\n")
	src := diagnostics.NewSource("main.rsc", []byte(`package app;
use "./lib/x.rsc";
use "lib/../lib/x.rsc" as again;
use "lib";
use "lib" as alsoLib;
`))
	prog, diags := LoadSource(src, dir)
	if diags != nil {
		t.Fatal(diags)
	}
	var names []string
	for _, f := range prog.Files {
		names = append(names, f.Source.Name)
	}
	if got := len(prog.Files); got != 3 {
		t.Fatalf("loaded %v, want main.rsc, lib/x.rsc, lib/y.rsm", names)
	}
	if len(prog.Packages) != 2 || len(prog.Packages[1].Files) != 2 {
		t.Errorf("packages: %+v", prog.Packages)
	}
	u := prog.Entry.Uses
	if u[0].Target != u[2].Target || u[0].Qualifier != "lib" || u[1].Qualifier != "again" {
		t.Errorf("uses: %+v %+v %+v", u[0], u[1], u[2])
	}
	for i, f := range prog.Files {
		if f.Source.Index != i {
			t.Errorf("%s has index %d, want %d", f.Source.Name, f.Source.Index, i)
		}
	}
}

func TestEntryName(t *testing.T) {
	dir := t.TempDir()
	if got := entryName(filepath.Join(dir, "a", "main.rsc"), dir); got != "a/main.rsc" {
		t.Errorf("inside: got %q", got)
	}
	outside := filepath.Join(filepath.Dir(dir), "main.rsc")
	if got := entryName(outside, dir); got != outside {
		t.Errorf("outside: got %q", got)
	}
}
