package jsgen

import (
	"regexp"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/ir"
	"github.com/AurielSolaris/runescript/internal/loader"
)

func generate(t *testing.T, src string) (string, []*diagnostics.Diagnostic) {
	t.Helper()
	prog, diags := loader.LoadSource(diagnostics.NewSource("t.rsc", []byte(src)), t.TempDir())
	if diags != nil {
		t.Fatal(diags)
	}
	info, diags := checker.Check(prog)
	if diags != nil {
		t.Fatal(diags)
	}
	out, diags, err := Generate(ir.Lower(info), Options{Version: "test"})
	if err != nil {
		t.Fatal(err)
	}
	return string(out), diags
}

func program(js string) string {
	_, after, _ := strings.Cut(js, "// ---- program ----")
	body, _, _ := strings.Cut(after, "export { RuneScriptError }")
	return body
}

// No integer operation reaches JavaScript's own arithmetic operators: each
// goes through a range-checked runtime helper with its source position.
func TestArithmeticIsChecked(t *testing.T) {
	js, _ := generate(t, `package p;
public uint f(uint a, uint b) { return a * b - a / b + a % b; }
public int main() {
    int x = 3;
    x += 4;
    x++;
    return -x * 2;
}
`)
	body := program(js)
	if m := regexp.MustCompile(`[a-z0-9)] [-+*/%] `).FindString(body); m != "" {
		t.Errorf("raw arithmetic operator %q in generated code:\n%s", m, body)
	}
	for _, want := range []string{"$mulU(a, b,", "$divU(a, b,", "$modU(a, b,", "x = $addI(x, 4,", "$subI(0, x,"} {
		if !strings.Contains(body, want) {
			t.Errorf("generated code lacks %q:\n%s", want, body)
		}
	}
}

func TestReservedNamesAreRenamed(t *testing.T) {
	js, _ := generate(t, `package p;
public int delete(int this) { return this; }
public int main() { int var = delete(1); return var; }
`)
	for _, want := range []string{"function delete$(this$, $site)", "let var$ = delete$(1,", "return var$;"} {
		if !strings.Contains(js, want) {
			t.Errorf("generated code lacks %q", want)
		}
	}
}

func TestQuote(t *testing.T) {
	cases := map[string]string{
		"plain":         `"plain"`,
		"q\"b\n\n\t\r":  `"q\"b\n\n\t\r"`,
		"\x00\x1f\x7f":  `"\u0000\u001f\u007f"`,
		"é😀":            `"é😀"`,
		"line\u2028sep": `"line\u2028sep"`,
	}
	for in, want := range cases {
		if got := quote(in); got != want {
			t.Errorf("quote(%q) = %s, want %s", in, got, want)
		}
	}
}

// Unicode tables are only included when the program uses them.
func TestTablesOnlyWhenUsed(t *testing.T) {
	plain, _ := generate(t, "package p;\npublic void main() { trace(1); }\n")
	upper, _ := generate(t, "package p;\nuse \"std:string\";\npublic void main() { trace(string.toUpper(\"a\")); }\n")
	if !strings.Contains(plain, "upper = lazy([])") || strings.Contains(upper, "upper = lazy([])") {
		t.Error("toUpper table inclusion is wrong")
	}
	if !strings.Contains(upper, "lower = lazy([])") {
		t.Error("unused toLower table was included")
	}
}
