package checker

import (
	"sort"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/loader"
)

// warnings checks src, which must have no errors, and returns its warnings
// as "line:col CODE" strings in source order.
func warnings(t *testing.T, src string) []string {
	t.Helper()
	prog, diags := loader.LoadSource(diagnostics.NewSource("t.rsc", []byte(src)), t.TempDir())
	if len(diags) != 0 {
		t.Fatalf("load: %v", diags)
	}
	info, diags := Check(prog)
	if len(diags) != 0 {
		t.Fatalf("errors: %v", diags)
	}
	ws := info.Warnings
	sort.SliceStable(ws, func(i, j int) bool { return ws[i].Span.Start.Before(ws[j].Span.Start) })
	var out []string
	for _, w := range ws {
		if w.Severity != diagnostics.Warning {
			t.Errorf("%v has severity %v", w, w.Severity)
		}
		out = append(out, w.Span.Start.String()+" "+w.Code.String())
	}
	return out
}

func TestWarnings(t *testing.T) {
	cases := []struct {
		name, src string
		want      []string
	}{
		{"unused local", program("    int x = 1;", ""), []string{"3:9 ASX-WARN-001"}},
		{"assigned, never read", program("    int x = 1;\n    x = 2;\n    x += 1;\n    x++;", ""), []string{"3:9 ASX-WARN-001"}},
		{"read local", program("    int x = 1;\n    trace(x);", ""), nil},
		{"read in condition", program("    bool b = true;\n    if (b) { }", ""), nil},
		{"called function variable", program("    void() f = () => trace(1);\n    f();", ""), nil},
		{"unused for variable", program("    for (int i = 0; ; ) { break; }", ""), []string{"3:14 ASX-WARN-001"}},
		{"unused parameter", "package p;\npublic void f(int a, int b) { trace(b); }\nprotected int main() { f(1, 2); return 0; }", []string{"2:19 ASX-WARN-002"}},
		{"unused lambda parameter", program("    int(int) f = (int n) => 0;\n    trace(f(1));", ""), []string{"3:23 ASX-WARN-002"}},
		{"prototype parameters", "package p;\npublic int f(int);\npublic int f(int n) { return n; }\nprotected int main() { trace(f(1)); return 0; }", nil},
		{"caught error not reported", program("    try { } catch (Error e) { }", ""), nil},
		{"unused use", "package p;\nuse \"std:string\";\nprotected int main() { return 0; }", []string{"2:1 ASX-WARN-003"}},
		{"used use", "package p;\nuse \"std:string\";\nprotected int main() { trace(string.toUpper(\"a\")); return 0; }", nil},
		{"after return", "package p;\nprotected int main() {\n    return 0;\n    trace(1);\n    trace(2);\n    return 0;\n}", []string{"4:5 ASX-WARN-004"}},
		{"after throw", "package p;\nprotected int main() {\n    throw new Error(\"x\");\n    trace(1);\n    return 0;\n}", []string{"4:5 ASX-WARN-004"}},
		{"after break", program("    while (true) {\n        break;\n        trace(1);\n    }", ""), []string{"5:9 ASX-WARN-004"}},
		{"after if that always returns", "package p;\nprotected int main() {\n    if (true) { return 0; } else { return 0; }\n    trace(1);\n    return 0;\n}", []string{"4:5 ASX-WARN-004"}},
		{"after if without else", "package p;\nprotected int main() {\n    if (true) { return 1; }\n    return 0;\n}", nil},
		{"after a loop", "package p;\nprotected int main() {\n    while (true) { }\n    return 0;\n}", nil},
		{"discarded result", program("    one();", "public int one() { return 1; }"), []string{"3:5 ASX-WARN-005"}},
		{"discarded std result", "package p;\nuse \"std:string\";\nprotected int main() { string.toUpper(\"a\"); return 0; }", []string{"3:24 ASX-WARN-005"}},
		{"void call", program("    trace(1);\n    v();", "public void v() { }"), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := warnings(t, tc.src)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWarningMessages(t *testing.T) {
	prog, _ := loader.LoadSource(diagnostics.NewSource("t.rsc", []byte(program("    int x = 1;\n    x = 2;\n    one();", "public int one() { return 1; }"))), t.TempDir())
	info, _ := Check(prog)
	var msgs []string
	for _, w := range info.Warnings {
		msgs = append(msgs, w.Message)
	}
	got := strings.Join(msgs, "\n")
	for _, want := range []string{"variable 'x' is assigned but never read", "the result of 'one', of type 'int', is discarded"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}
