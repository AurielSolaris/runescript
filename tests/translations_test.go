package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Translations (localization/<lang>/) mirror the English README and guide.
// They translate the prose only: every code block must be identical to the
// English one, in the same order, so a translation can never show a program
// that differs from the tested English docs. (TestDocExamples also compiles
// the programs in the translated guides.)

var (
	sourceRE = regexp.MustCompile(`^<!-- source: (\S+) @ ([0-9a-f]{7,40}) -->`)
	blockRE  = regexp.MustCompile("(?ms)^[ \t]*```[^\n]*\n.*?^[ \t]*```")
)

func codeBlocks(text string) []string {
	blocks := blockRE.FindAllString(strings.ReplaceAll(text, "\r\n", "\n"), -1)
	for i, b := range blocks {
		blocks[i] = strings.TrimSpace(b)
	}
	return blocks
}

// translatedFiles are the English files every translation covers.
func translatedFiles(t *testing.T) []string {
	t.Helper()
	guide, err := filepath.Glob("../docs/guide/*.md")
	if err != nil || len(guide) == 0 {
		t.Fatalf("no guide pages: %v", err)
	}
	files := []string{"README.md"}
	for _, g := range guide {
		files = append(files, strings.TrimPrefix(filepath.ToSlash(g), "../"))
	}
	return files
}

func TestTranslationsKeepCodeBlocks(t *testing.T) {
	langs, _ := filepath.Glob("../localization/*")
	var dirs []string
	for _, l := range langs {
		if info, err := os.Stat(l); err == nil && info.IsDir() {
			dirs = append(dirs, l)
		}
	}
	if len(dirs) == 0 {
		t.Skip("no translations")
	}
	for _, dir := range dirs {
		lang := filepath.Base(dir)
		for _, src := range translatedFiles(t) {
			t.Run(lang+"/"+src, func(t *testing.T) {
				path := filepath.Join(dir, filepath.FromSlash(src))
				tr, err := os.ReadFile(path)
				if os.IsNotExist(err) {
					// Not translated yet: readers get the English page
					// (localization/README.md lists the coverage).
					t.Skipf("%s is not translated yet", src)
				}
				if err != nil {
					t.Fatal(err)
				}
				m := sourceRE.FindSubmatch(tr)
				if m == nil || string(m[1]) != src {
					t.Fatalf("first line must be <!-- source: %s @ <commit> -->", src)
				}
				en := codeBlocks(string(read(t, "../"+src)))
				got := codeBlocks(string(tr))
				if len(got) != len(en) {
					t.Fatalf("%d code blocks, English has %d", len(got), len(en))
				}
				for i := range en {
					if got[i] != en[i] {
						t.Errorf("code block %d differs from the English one:\n%s\n--- English:\n%s", i+1, got[i], en[i])
					}
				}
			})
		}
	}
}
