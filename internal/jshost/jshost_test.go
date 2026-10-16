package jshost

import (
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/types"
)

func TestModules(t *testing.T) {
	if Lookup("rune:fs") == nil || Lookup("rune:nope") != nil || Lookup("node:fs") != nil {
		t.Fatal("lookup")
	}
	for _, name := range Names() {
		m := Lookup(name)
		if !strings.HasPrefix(name, Prefix) || m.Node == "" || len(m.Funcs) == 0 {
			t.Errorf("%s: incomplete module", name)
		}
		for fname, f := range m.Funcs {
			if f.Name != fname || f.Doc == "" || !strings.Contains(f.Node, "$host.") || !strings.Contains(f.Node, "=>") {
				t.Errorf("%s.%s: incomplete function", name, fname)
			}
			// Only types that can cross the JavaScript boundary.
			for _, p := range append(append([]*types.Type{}, f.Params...), f.Result) {
				switch p.Kind {
				case types.Signed, types.Unsigned, types.Float, types.Bool, types.String, types.Error, types.Void:
				default:
					t.Errorf("%s.%s: type %s", name, fname, p)
				}
			}
		}
	}
}

func TestHostSpecific(t *testing.T) {
	for m, want := range map[string]bool{"node:fs": true, "bun:sqlite": true, "deno:x": true, "rune:fs": false, "./lib.mjs": false, "lodash": false} {
		if HostSpecific(m) != want {
			t.Errorf("HostSpecific(%q) = %v", m, !want)
		}
	}
}
