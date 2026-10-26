package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/lexer"
)

// The editor packages in editors/ list RuneScript's keywords for
// highlighting; a keyword added to the lexer must be added to them too.
func TestEditorsKnowEveryKeyword(t *testing.T) {
	files := []string{
		"../editors/vscode/syntaxes/runescript.tmLanguage.json",
		"../editors/vim/syntax/runescript.vim",
		"../editors/emacs/runescript-mode.el",
	}
	for _, path := range files {
		text := string(read(t, path))
		for _, k := range append(lexer.Keywords(), "Array", "each", "in", "as", "template") {
			if !regexp.MustCompile(`(^|[^A-Za-z0-9_])` + k + `($|[^A-Za-z0-9_])`).MatchString(text) {
				t.Errorf("%s does not mention the keyword %q", path, k)
			}
		}
	}
}

func TestVSCodeExtensionManifest(t *testing.T) {
	var pkg struct {
		Main        string `json:"main"`
		Contributes struct {
			Languages []struct {
				ID            string   `json:"id"`
				Extensions    []string `json:"extensions"`
				Configuration string   `json:"configuration"`
			} `json:"languages"`
			Grammars []struct {
				ScopeName string `json:"scopeName"`
				Path      string `json:"path"`
			} `json:"grammars"`
		} `json:"contributes"`
	}
	if err := json.Unmarshal(read(t, "../editors/vscode/package.json"), &pkg); err != nil {
		t.Fatal(err)
	}
	lang := pkg.Contributes.Languages
	if len(lang) != 1 || lang[0].ID != "runescript" || strings.Join(lang[0].Extensions, " ") != ".rsc .rsm" {
		t.Errorf("languages: %+v", lang)
	}
	for _, p := range []string{pkg.Main, lang[0].Configuration, pkg.Contributes.Grammars[0].Path} {
		var v any
		if strings.HasSuffix(p, ".json") {
			if err := json.Unmarshal(read(t, filepath.Join("../editors/vscode", p)), &v); err != nil {
				t.Errorf("%s: %v", p, err)
			}
		} else if _, err := os.Stat(filepath.Join("../editors/vscode", p)); err != nil {
			t.Error(err)
		}
	}
	var grammar struct {
		ScopeName string `json:"scopeName"`
	}
	json.Unmarshal(read(t, "../editors/vscode/syntaxes/runescript.tmLanguage.json"), &grammar)
	if grammar.ScopeName != pkg.Contributes.Grammars[0].ScopeName {
		t.Errorf("grammar scope %q, manifest %q", grammar.ScopeName, pkg.Contributes.Grammars[0].ScopeName)
	}
}

// TestVimPlugin loads the Vim plugin in a headless Vim, if there is one,
// and checks the file type, a few highlight groups, and indentation.
func TestVimPlugin(t *testing.T) {
	vim, err := exec.LookPath("vim")
	if err != nil {
		t.Skip("vim is not on PATH")
	}
	dir := t.TempDir()
	plugin, _ := filepath.Abs("../editors/vim")
	src := filepath.Join(dir, "main.rsc")
	write := func(name, text string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("main.rsc", "package main;\npublic struct Point { int x; }\nprotected int main() {\nstring s = $\"{s} {{\";\nif (true) {\ntrace(0x1F); // TODO\n}\nreturn 0;\n}\n")
	write("check.vim", `let out = [&filetype]
for [l, c] in [[1, 1], [2, 15], [3, 1], [3, 15], [4, 14], [4, 15], [4, 18], [6, 1], [6, 7], [6, 17]]
  call add(out, synIDattr(synID(l, c, 1), 'name'))
endfor
normal! gg=G
call extend(out, getline(1, '$'))
call writefile(out, 'out.txt')
qa!
`)
	cmd := exec.Command(vim, "-Nu", "NONE", "-i", "NONE", "-es", "--cmd", "set rtp^="+filepath.ToSlash(plugin),
		"--cmd", "filetype plugin indent on", "--cmd", "syntax on", "-S", "check.vim", src)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("vim: %v\n%s", err, out)
	}
	got := strings.TrimSpace(string(read(t, filepath.Join(dir, "out.txt"))))
	want := strings.Join([]string{
		"runescript",
		"runescriptInclude", "runescriptStructName", "runescriptStorage", "runescriptFunction",
		"runescriptHoleDelim", "runescriptHole", "runescriptBraceEscape",
		"runescriptBuiltin", "runescriptNumber", "runescriptTodo",
		"package main;",
		"public struct Point { int x; }",
		"protected int main() {",
		"    string s = $\"{s} {{\";",
		"    if (true) {",
		"        trace(0x1F); // TODO",
		"    }",
		"    return 0;",
		"}",
	}, "\n")
	if strings.ReplaceAll(got, "\r\n", "\n") != want {
		t.Errorf("vim gave:\n%s\nwant:\n%s", got, want)
	}
}
