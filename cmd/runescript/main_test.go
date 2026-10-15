package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, src string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "prog.rs")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCLI(t *testing.T) {
	ok := write(t, "package main;\nprotected int main() {\n    trace(\"hello\");\n    return 3;\n}\n")
	bad := write(t, "package main;\nint main() {\n    sint s = -10;\n    uint u = s;\n    return 0;\n}\n")
	crash := write(t, "package main;\nint main() {\n    int z = 0;\n    return 1 / z;\n}\n")

	cases := []struct {
		name           string
		args           []string
		code           int
		stdout, stderr string // substrings
	}{
		{"run", []string{"run", ok}, 3, "hello\n", ""},
		{"check ok", []string{"check", ok}, 0, "", ""},
		{"check bad", []string{"check", bad}, exitCompileError, "", "error ASX-TYPE-004"},
		{"run bad", []string{"run", bad}, exitCompileError, "", "prog.rs:4:14"},
		{"runtime error", []string{"run", crash}, exitRuntimeError, "", "error ASX-RUN-003"},
		{"missing file", []string{"run", filepath.Join(t.TempDir(), "nope.rs")}, exitUsage, "", "runescript:"},
		{"no args", nil, exitUsage, "", "usage:"},
		{"unknown command", []string{"build", ok}, exitUsage, "", "usage:"},
		{"version", []string{"version"}, 0, "runescript " + Version, ""},
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
