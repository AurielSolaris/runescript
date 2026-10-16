package driver

import (
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/jsgen"
)

// Diagnostics come back in source order even when the checker finds them
// out of order (an undefined prototype is only known at the end of its scope).
func TestDiagnosticsAreSortedByPosition(t *testing.T) {
	_, diags := Check([]byte("package p;\npublic int f(int);\npublic void main() {\n    trace(x);\n}\n"))
	if len(diags) != 2 {
		t.Fatalf("got %v", diags)
	}
	if diags[0].Span.Start.Line != 2 || diags[1].Span.Start.Line != 4 {
		t.Errorf("not sorted: %v", diags)
	}
}

func TestRunReturnsExitCodeAndOutput(t *testing.T) {
	var out strings.Builder
	code, diags, err := Run([]byte("package p;\npublic int main() {\n    int(int) f = (int x) => x + 1;\n    trace(\"ok\");\n    return f(41);\n}\n"), &out)
	if diags != nil || err != nil || code != 42 || out.String() != "ok\n" {
		t.Errorf("code=%d out=%q diags=%v err=%v", code, out.String(), diags, err)
	}
}

// Extern functions run only on the JavaScript target: the program checks,
// but the interpreter refuses it with one diagnostic at the first extern.
func TestInterpreterRejectsExterns(t *testing.T) {
	src := []byte("package p;\npublic void main() { trace(f(), g()); }\nextern \"./m.mjs\" int f();\nextern \"./m.mjs\" int g();\n")
	if _, diags := Check(src); diags != nil {
		t.Fatalf("check: %v", diags)
	}
	var out strings.Builder
	_, diags, err := Run(src, &out)
	if err != nil || len(diags) != 1 || diags[0].Code.String() != "ASX-TARGET-002" || diags[0].Span.Start.Line != 3 ||
		!strings.Contains(diags[0].Message, "(1 more extern function(s) are declared)") {
		t.Errorf("diags=%v err=%v", diags, err)
	}
	if js, diags, err := BuildJS("p.rsc", src, jsgen.Options{}); err != nil || diags != nil || !strings.Contains(string(js), `import * as $m0 from "./m.mjs";`) {
		t.Errorf("build: diags=%v err=%v", diags, err)
	}
}
