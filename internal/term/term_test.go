package term

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMode(t *testing.T) {
	for s, want := range map[string]Mode{"auto": Auto, "always": Always, "never": Never} {
		if m, err := ParseMode(s); err != nil || m != want {
			t.Errorf("ParseMode(%q) = %v, %v", s, m, err)
		}
	}
	if _, err := ParseMode("yes"); err == nil {
		t.Error("ParseMode accepted \"yes\"")
	}
}

func TestColor(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), "out.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	switch {
	case !Color(Always, nil), !Color(Always, file):
		t.Error("Always did not color")
	case Color(Never, os.Stderr):
		t.Error("Never colored")
	case Color(Auto, nil):
		t.Error("Auto colored a nil file")
	case Color(Auto, file):
		t.Error("Auto colored a regular file")
	case IsTerminal(file):
		t.Error("a regular file is a terminal")
	}
	t.Setenv("NO_COLOR", "1")
	if Color(Auto, os.Stderr) {
		t.Error("Auto colored with NO_COLOR set")
	}
}
