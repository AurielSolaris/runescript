package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, src string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "prog.rsc")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCLI(t *testing.T) {
	ok := write(t, "package main;\nprotected int main() {\n    trace(\"hello\");\n    return 3;\n}\n")
	bad := write(t, "package main;\npublic int main() {\n    sint s = -10;\n    uint u = s;\n    return 0;\n}\n")
	v02 := write(t, "package main;\nuse \"std:string\";\npublic int main() {\n    string(string) shout = (string s) => string.toUpper(s);\n    trace(shout(\"rune\"));\n    private int seven() { return 7; }\n    return seven();\n}\n")
	crash := write(t, "package main;\npublic int main() {\n    int z = 0;\n    return 1 / z;\n}\n")
	warn := write(t, "package main;\npublic int main() {\n    int unused = 1;\n    trace(\"ran\");\n    return 0;\n}\n")

	cases := []struct {
		name           string
		args           []string
		code           int
		stdout, stderr string // substrings
	}{
		{"run", []string{"run", ok}, 3, "hello\n", ""},
		{"check ok", []string{"check", ok}, 0, "", ""},
		{"v0.2 features", []string{"run", v02}, 7, "RUNE\n", ""},
		{"check bad", []string{"check", bad}, exitCompileError, "", "error ASX-TYPE-004"},
		{"run bad", []string{"run", bad}, exitCompileError, "", "prog.rsc:4:14"},
		{"runtime error", []string{"run", crash}, exitRuntimeError, "", "error ASX-RUN-003"},
		{"missing file", []string{"run", filepath.Join(t.TempDir(), "nope.rsc")}, exitUsage, "", "runescript:"},
		{"no args", nil, exitUsage, "", "usage:"},
		{"unknown command", []string{"frobnicate", ok}, exitUsage, "", "usage:"},
		{"build bad", []string{"build", bad}, exitCompileError, "", "error ASX-TYPE-004"},
		{"build unknown target", []string{"build", "--target", "wasm", ok}, exitUsage, "", "unknown target"},
		{"build no file", []string{"build"}, exitUsage, "", "usage:"},
		{"version", []string{"version"}, 0, "runescript " + Version, ""},
		{"warning runs", []string{"run", warn}, 0, "ran\n", "prog.rsc:3:9: warning ASX-WARN-001: variable 'unused' is never used"},
		{"warning check", []string{"check", warn}, 0, "", "warning ASX-WARN-001"},
		{"Werror run", []string{"run", "-Werror", warn}, exitCompileError, "", "error ASX-WARN-001: variable 'unused' is never used [-Werror]"},
		{"Werror build", []string{"build", "-Werror", "-o", filepath.Join(t.TempDir(), "w.js"), warn}, exitCompileError, "", "[-Werror]"},
		{"allow", []string{"run", "-Werror", "--allow", "ASX-WARN-001", warn}, 0, "ran\n", ""},
		{"allow list without prefix", []string{"check", "-Werror", "--allow=warn-002,WARN-001", warn}, 0, "", ""},
		{"allow bad code", []string{"check", "--allow", "ASX-TYPE-004", warn}, exitUsage, "", "--allow takes warning codes"},
		{"run too many files", []string{"run", ok, ok}, exitUsage, "", "usage:"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr strings.Builder
			code := cli(c.args, &stdout, &stderr)
			if code != c.code {
				t.Errorf("exit %d, want %d (stderr: %s)", code, c.code, stderr.String())
			}
			if !strings.Contains(stdout.String(), c.stdout) {
				t.Errorf("stdout %q does not contain %q", stdout.String(), c.stdout)
			}
			if !strings.Contains(stderr.String(), c.stderr) {
				t.Errorf("stderr %q does not contain %q", stderr.String(), c.stderr)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	src := write(t, "package main;\npublic int main() {\n    trace(\"hi\");\n    return 0;\n}\n")
	var stdout, stderr strings.Builder
	if code := cli([]string{"build", src}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	js, err := os.ReadFile(strings.TrimSuffix(src, ".rsc") + ".js")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"prog.rsc. Do not edit.", "export function run(", `$trace("hi")`} {
		if !strings.Contains(string(js), want) {
			t.Errorf("output lacks %q", want)
		}
	}
	out := filepath.Join(t.TempDir(), "custom.mjs")
	if code := cli([]string{"build", "--target", "js", "-o", out, src}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if _, err := os.Stat(out); err != nil {
		t.Error(err)
	}
}

// Programs can span files with any extension; uses are resolved against
// the working directory.
func TestMultiFileProgram(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "lib"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"tool.txt": `package app;
use "./lib/math.inc";
public int main() {
    trace(math.triple(3));
    return math.triple(1);
}
`,
		"lib/math.inc": `package math;
public int triple(int n) {
    return n * 3;
}
`,
		"broken.rsc": `package app;
use "./lib/nope.rsc";
public int main() {
    return 0;
}
`,
	}
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	if code := cli([]string{"run", "tool.txt"}, &stdout, &stderr); code != 3 || strings.TrimSpace(stdout.String()) != "9" {
		t.Errorf("run: exit %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
	}
	stderr.Reset()
	want := `broken.rsc:2:5: error ASX-USE-004: cannot find file "lib/nope.rsc"`
	if code := cli([]string{"check", "broken.rsc"}, &stdout, &stderr); code != exitCompileError || !strings.HasPrefix(stderr.String(), want) {
		t.Errorf("check: exit %d, stderr %q", code, stderr.String())
	}
	if code := cli([]string{"build", "tool.txt"}, &stdout, &stderr); code != 0 {
		t.Fatalf("build: exit %d: %s", code, stderr.String())
	}
	if _, err := os.Stat("tool.js"); err != nil {
		t.Error(err)
	}
}

// A UTF-16 source file is converted, with a note; a file in a legacy 8-bit
// encoding is rejected.
func TestSourceEncodings(t *testing.T) {
	dir := t.TempDir()
	src := "package main;\npublic int main() {\n    trace(\"héllo €\");\n    return 0;\n}\n"
	utf16le := []byte{0xFF, 0xFE}
	for _, r := range src { // the program is in the Basic Multilingual Plane
		utf16le = append(utf16le, byte(r), byte(r>>8))
	}
	wide := filepath.Join(dir, "wide.rsc")
	latin := filepath.Join(dir, "latin.rsc")
	if err := os.WriteFile(wide, utf16le, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(latin, []byte("package main;\npublic int main() {\n    trace(\"h\xe9llo\");\n    return 0;\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	if code := cli([]string{"run", wide}, &stdout, &stderr); code != 0 || stdout.String() != "héllo €\n" ||
		!strings.Contains(stderr.String(), "is UTF-16LE; it was read as UTF-8") {
		t.Errorf("UTF-16: exit %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := cli([]string{"run", latin}, &stdout, &stderr); code != exitCompileError ||
		!strings.Contains(stderr.String(), "latin.rsc:3:13: error ASX-LEX-007: the file is not valid UTF-8: byte 0xE9") {
		t.Errorf("Windows-1252: exit %d, stderr %q", code, stderr.String())
	}
}

// build refuses to overwrite its own source.
func TestBuildDoesNotOverwriteSource(t *testing.T) {
	src := filepath.Join(t.TempDir(), "app.js")
	text := `package main;
public int main() {
    return 0;
}
`
	if err := os.WriteFile(src, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	if code := cli([]string{"build", src}, &stdout, &stderr); code != exitUsage || !strings.Contains(stderr.String(), "would overwrite the source") {
		t.Errorf("exit %d, stderr %q", code, stderr.String())
	}
	if got, _ := os.ReadFile(src); string(got) != text {
		t.Error("source was modified")
	}
}
