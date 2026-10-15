package tests

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/runtime"
)

var (
	fenceRE        = regexp.MustCompile("(?s)```actionscript\r?\n(.*?)```")
	compileErrRE   = regexp.MustCompile(`// error: (ASX-[A-Z]+-\d+)`)
	runtimeErrRE   = regexp.MustCompile(`// runtime error: (ASX-RUN-\d+)`)
	docSourceGlobs = []string{"../README.md", "../docs/*.md", "../docs/guide/*.md"}
)

// TestDocExamples keeps the documentation honest: every ```actionscript
// block that is a complete program (starts with `package`) must compile,
// unless it is annotated with `// error: CODE` (must fail with CODE first)
// or `// runtime error: CODE` (must compile, then fail at run time).
func TestDocExamples(t *testing.T) {
	total := 0
	for _, glob := range docSourceGlobs {
		files, _ := filepath.Glob(glob)
		for _, path := range files {
			if filepath.Base(path) == "OPEN_QUESTIONS.md" {
				continue // local design notes, not user documentation
			}
			text := string(read(t, path))
			for i, m := range fenceRE.FindAllStringSubmatch(text, -1) {
				src := m[1]
				if !strings.HasPrefix(strings.TrimSpace(src), "package") {
					continue
				}
				total++
				name := filepath.Base(path) + "#" + string(rune('A'+i%26))
				t.Run(name, func(t *testing.T) { checkDocProgram(t, src) })
			}
		}
	}
	if total < 20 {
		t.Errorf("only %d doc programs found; did the fence format change?", total)
	}
}

func checkDocProgram(t *testing.T, src string) {
	if m := compileErrRE.FindStringSubmatch(src); m != nil {
		_, diags := driver.Check([]byte(src))
		if len(diags) == 0 || diags[0].Code.String() != m[1] {
			t.Errorf("expected %s, got %v\n%s", m[1], diags, src)
		}
		return
	}
	_, diags, err := driver.Run([]byte(src), io.Discard)
	if diags != nil {
		t.Fatalf("doc program does not compile: %v\n%s", diags, src)
	}
	if m := runtimeErrRE.FindStringSubmatch(src); m != nil {
		var re *runtime.Error
		if !errors.As(err, &re) || re.Diag.Code.String() != m[1] {
			t.Errorf("expected runtime error %s, got %v\n%s", m[1], err, src)
		}
		return
	}
	if err != nil {
		t.Errorf("doc program failed at run time: %v\n%s", err, src)
	}
}

// Guard against the docs drifting from the code list in diagnostics.
func TestDiagnosticsReferenceIsComplete(t *testing.T) {
	ref := string(read(t, "../docs/guide/08-diagnostics.md"))
	src, err := os.ReadFile("../internal/diagnostics/diagnostics.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range regexp.MustCompile(`Code = "([A-Z]+-\d+)"`).FindAllStringSubmatch(string(src), -1) {
		if !strings.Contains(ref, "ASX-"+m[1]) {
			t.Errorf("docs/guide/08-diagnostics.md does not document ASX-%s", m[1])
		}
	}
}
