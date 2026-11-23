package tests

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/ir"
	"github.com/AurielSolaris/runescript/internal/jsgen"
)

func TestNativeGraphicsExamples(t *testing.T) {
	for _, name := range []string{"click-circle.rsc", "keep-up.rsc"} {
		t.Run(name, func(t *testing.T) {
			in := driver.Input{Path: filepath.Join("..", "examples", "native", name), Dir: "."}
			info, ds, err := driver.CheckInput(in)
			if err != nil || len(ds) != 0 {
				t.Fatalf("check: %v %v", ds, err)
			}
			if err := ir.Verify(ir.Lower(info)); err != nil {
				t.Fatal(err)
			}
			if _, ds, err := driver.CheckWASMInput(in); err != nil || len(ds) != 0 {
				t.Fatalf("WASM check: %v %v", ds, err)
			}
			_, ds, err = driver.BuildJSInput(in, jsgen.Options{})
			if err != nil || len(ds) == 0 || ds[0].Code != diagnostics.TargetNeedsNative {
				t.Fatalf("JS build diagnostics: %v %v", ds, err)
			}
		})
	}
}

func TestGraphicsValueTypes(t *testing.T) {
	in := driver.Memory("values.rsc", []byte(`package game;
use "rune.gfx" as drawing;
protected int main() {
    drawing.Color original = drawing.rgba(20, 30, 40, 255);
    drawing.Color copy = original;
    copy.r = 80;
    drawing.Vector2 p = drawing.vector(12.5, 6.25);
    trace(original.r, copy.r, p.x, p.y);
    return 0;
}`))
	var out strings.Builder
	code, ds, err := driver.RunInput(in, &out)
	if err != nil || len(ds) != 0 || code != 0 || out.String() != "20 80 12.5 6.25\n" {
		t.Fatalf("exit %d, output %q, diagnostics %v, error %v", code, out.String(), ds, err)
	}
}
