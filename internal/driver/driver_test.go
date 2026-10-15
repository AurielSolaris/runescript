package driver

import (
	"strings"
	"testing"
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
