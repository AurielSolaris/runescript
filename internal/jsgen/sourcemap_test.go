package jsgen

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/ir"
	"github.com/AurielSolaris/runescript/internal/loader"
)

func generateMap(t *testing.T, src string, opts Options) Output {
	t.Helper()
	prog, diags := loader.LoadSource(diagnostics.NewSource("t.rsc", []byte(src)), t.TempDir())
	if diags != nil {
		t.Fatal(diags)
	}
	info, diags := checker.Check(prog)
	if diags != nil {
		t.Fatal(diags)
	}
	out, diags, err := GenerateMap(ir.Lower(info), opts)
	if err != nil || diags != nil {
		t.Fatal(err, diags)
	}
	return out
}

// segment is a decoded mapping: generated line and column, and the
// source line and column (all 0-based); src is -1 when unmapped.
type segment struct{ line, col, src, srcLine, srcCol int }

// decodeMappings decodes a source map's mappings independently of the
// encoder in sourcemap.go.
func decodeMappings(t *testing.T, s string) []segment {
	t.Helper()
	var out []segment
	src, srcLine, srcCol := 0, 0, 0
	for line, group := range strings.Split(s, ";") {
		col := 0
		if group == "" {
			continue
		}
		for _, seg := range strings.Split(group, ",") {
			var fields []int
			for i := 0; i < len(seg); {
				v, shift := 0, 0
				for {
					d := strings.IndexByte(base64Digits, seg[i])
					if d < 0 {
						t.Fatalf("bad base64 digit %q in %q", seg[i], seg)
					}
					i++
					v |= (d & 31) << shift
					shift += 5
					if d&32 == 0 {
						break
					}
				}
				if v&1 != 0 {
					v = -(v >> 1)
				} else {
					v >>= 1
				}
				fields = append(fields, v)
			}
			col += fields[0]
			switch len(fields) {
			case 1:
				out = append(out, segment{line, col, -1, 0, 0})
			case 4:
				src += fields[1]
				srcLine += fields[2]
				srcCol += fields[3]
				out = append(out, segment{line, col, src, srcLine, srcCol})
			default:
				t.Fatalf("segment %q has %d fields", seg, len(fields))
			}
		}
	}
	return out
}

func TestSourceMap(t *testing.T) {
	src := `package p;

public int twice(int x) {
    return x * 2;
}

protected int main() {
    int(int) f = (int a) => {
        return a + 1;
    };
    Array<int> xs = [1, 2];
    for each (int x in xs) {
        trace(twice(x));
    }
    string s = "é😀"; trace(s);
    return f(0);
}
`
	out := generateMap(t, src, Options{Version: "test", SourceMap: true, File: "t.js", MapURL: "t.js.map"})
	js := string(out.JS)
	if strings.ContainsAny(js, string(markOpen)+string(markClose)) {
		t.Fatal("markers left in the output")
	}
	if !strings.HasSuffix(js, "\n//# sourceMappingURL=t.js.map\n") {
		t.Errorf("no sourceMappingURL comment at the end:\n%s", js[len(js)-200:])
	}
	var m sourceMapJSON
	if err := json.Unmarshal(out.Map, &m); err != nil {
		t.Fatal(err)
	}
	if m.Version != 3 || m.File != "t.js" || len(m.Sources) != 1 || m.Sources[0] != "t.rsc" || m.SourcesContent[0] != src {
		t.Errorf("map header: %+v", m)
	}
	segs := decodeMappings(t, m.Mappings)
	jsLines := strings.Split(js, "\n")
	srcLines := strings.Split(src, "\n")
	programStart := 0
	for i, l := range jsLines {
		if strings.Contains(l, "// ---- program ----") {
			programStart = i
		}
	}
	// at returns the source text a generated line containing code maps to.
	at := func(code string) string {
		t.Helper()
		for i, l := range jsLines {
			if i < programStart || !strings.Contains(l, code) {
				continue
			}
			var best *segment
			for j := range segs {
				if segs[j].line == i {
					best = &segs[j]
				}
			}
			if best == nil || best.src < 0 {
				t.Errorf("%q (generated line %d) is not mapped", code, i+1)
				return ""
			}
			if want := len(l) - len(strings.TrimLeft(l, " ")); best.col != want {
				t.Errorf("%q: generated column %d, want %d", code, best.col, want)
			}
			return string(utf16.Decode(utf16.Encode([]rune(srcLines[best.srcLine]))[best.srcCol:]))
		}
		t.Fatalf("no generated line contains %q", code)
		return ""
	}
	for code, want := range map[string]string{
		"function twice(":   "public int twice(int x) {",
		"function main(":    "protected int main() {",
		"let f = ":          "int(int) f = (int a) => {",
		"return $addI(a, 1": "return a + 1;",
		"$trace(twice(x":    "trace(twice(x));",
		"let x = $aget(":    "for each (int x in xs) {",
		"let s = ":          `string s = "é😀"; trace(s);`,
		"$trace(s)":         "trace(s);", // UTF-16 column after a non-ASCII string
		"return f(0":        "return f(0);",
	} {
		if got := at(code); got != want {
			t.Errorf("%q maps to %q, want %q", code, got, want)
		}
	}
	// The module's footer comes from no source: its first line is unmapped.
	foot := -1
	for i, l := range jsLines {
		if strings.HasPrefix(l, "export { RuneScriptError }") {
			foot = i
		}
	}
	found := false
	for _, s := range segs {
		if s.line == foot && s.src == -1 {
			found = true
		}
	}
	if !found {
		t.Errorf("the footer (line %d) is not marked unmapped", foot+1)
	}
}

func TestSourceMapInlineAndOff(t *testing.T) {
	src := "package p;\nprotected int main() {\n    trace(1);\n    return 0;\n}\n"
	plain := generateMap(t, src, Options{Version: "test"})
	if plain.Map != nil || strings.Contains(string(plain.JS), "sourceMappingURL") {
		t.Error("a map without SourceMap")
	}
	inline := generateMap(t, src, Options{Version: "test", SourceMap: true})
	js := string(inline.JS)
	_, url, ok := strings.Cut(js, "//# sourceMappingURL=data:application/json;charset=utf-8;base64,")
	if !ok {
		t.Fatalf("no inline map:\n%s", js[len(js)-200:])
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(url))
	if err != nil || string(data) != string(inline.Map) {
		t.Errorf("inline map differs from Map: %v", err)
	}
	// Apart from the comment, the code is the same with or without a map.
	if code, _, _ := strings.Cut(js, "//# sourceMappingURL="); code != string(plain.JS) {
		t.Error("the source map changed the generated code")
	}
}

func TestVLQ(t *testing.T) {
	for n, want := range map[int]string{0: "A", 1: "C", -1: "D", 15: "e", 16: "gB", -16: "hB", 1000: "w+B"} {
		var b strings.Builder
		vlq(&b, n)
		if b.String() != want {
			t.Errorf("vlq(%d) = %q, want %q", n, b.String(), want)
		}
	}
}
