package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/loader"
)

// withExterns is a program whose externs come before main; main's body
// starts on line 2 + the number of extern lines + 1.
func withExterns(externs, body string) string {
	return "package p;\n" + externs + "\npublic int main() {\n" + body + "\n    return 0;\n}\n"
}

func TestExternsAccepted(t *testing.T) {
	cases := map[string]string{
		"rune:fs":              withExterns(`extern "rune:fs" string readFile(string path);`, `    trace(readFile("a"));`),
		"param names optional": withExterns(`extern "rune:fs" void writeFile(string, string);`, `    writeFile("a", "b");`),
		"own module":           withExterns(`extern "./lib.mjs" float32 f(int a, uint b, float c, float64 d, bool e, string s);`, `    trace(f(1, 2, 0.5, 1.5, true, "s"));`),
		"errors":               withExterns(`extern "./lib.mjs" RangeError f(Error e);`, "    RangeError r = f(new TypeError(\"t\"));\n    trace(r.message);"),
		"void result":          withExterns(`extern "./lib.mjs" void f();`, `    f();`),
		"extern after main":    "package p;\npublic void main() { trace(f()); }\nextern \"./lib.mjs\" int f();\n",
		"wrapped in lambda":    withExterns(`extern "./lib.mjs" int f(int n);`, "    int(int) g = (int n) => f(n);\n    trace(g(1));"),
		"per-file name":        withExterns(`extern "./lib.mjs" int f();`, `    trace(f());`),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if diags := check(t, src); len(diags) != 0 {
				t.Errorf("unexpected diagnostics: %v", diags)
			}
		})
	}
}

func TestExternsRejected(t *testing.T) {
	cases := []struct {
		name string
		src  string
		code diagnostics.Code
		pos  string
	}{
		{"function type param", withExterns(`extern "./m.mjs" void f(int(int) g);`, ""), diagnostics.TypeExternType, "2:25"},
		{"function type result", withExterns(`extern "./m.mjs" int() f();`, ""), diagnostics.TypeExternType, "2:18"},
		{"void param", withExterns(`extern "./m.mjs" int f(void v);`, ""), diagnostics.TypeVoidValue, "2:24"},
		{"unknown rune module", withExterns(`extern "rune:net" int f();`, ""), diagnostics.UseExternUnknown, "2:8"},
		{"unknown rune function", withExterns(`extern "rune:fs" string read(string p);`, ""), diagnostics.UseExternUnknown, "2:25"},
		{"rune signature mismatch", withExterns(`extern "rune:fs" int readFile(string p);`, ""), diagnostics.TypeExternMismatch, "2:22"},
		{"rune param mismatch", withExterns(`extern "rune:fs" bool exists(int p);`, ""), diagnostics.TypeExternMismatch, "2:23"},
		{"duplicate extern", withExterns("extern \"./m.mjs\" int f();\nextern \"./n.mjs\" int f();", ""), diagnostics.NameRedeclared, "3:22"},
		{"clashes with function", "package p;\nextern \"./m.mjs\" int main();\npublic void main() { }\n", diagnostics.NameRedeclared, "2:22"},
		{"local shadows extern", withExterns(`extern "./m.mjs" int f();`, "    int f = 1;\n    trace(f);"), diagnostics.NameRedeclared, "4:9"},
		{"used as value", withExterns(`extern "./m.mjs" int f();`, "    int() g = f;\n    trace(g());"), diagnostics.NameNotAValue, "4:15"},
		{"assigned", withExterns(`extern "./m.mjs" int f();`, "    f = 1;"), diagnostics.NameAssignToParam, "4:5"},
		{"argument type", withExterns(`extern "./m.mjs" void f(uint u);`, "    int s = 1;\n    f(s);"), diagnostics.TypeSignedToUnsigned, "5:7"},
		{"argument count", withExterns(`extern "./m.mjs" void f(int a);`, "    f();"), diagnostics.TypeArgCount, "4:5"},
		{"redeclares trace", withExterns(`extern "./m.mjs" void trace(string s);`, ""), diagnostics.NameBuiltinReused, "2:23"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			diags := check(t, c.src)
			for _, d := range diags {
				if d.Code == c.code && d.Span.Start.String() == c.pos {
					return
				}
			}
			t.Errorf("got %v, want %s at %s", diags, c.code, c.pos)
		})
	}
}

// Externs are private to their file: another file of the package does
// not see them.
func TestExternIsFilePrivate(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) {
		t.Helper()
		if err := writeFile(dir, name, src); err != nil {
			t.Fatal(err)
		}
	}
	write("lib.rsm", "package p;\nextern \"./m.mjs\" int f();\nprotected int g() { return f(); }\n")
	prog, diags := loader.LoadSource(diagnostics.NewSource("main.rsc", []byte("package p;\nuse \"./lib.rsm\";\npublic int main() { return g() + f(); }\n")), dir)
	if len(diags) != 0 {
		t.Fatal(diags)
	}
	_, diags = Check(prog)
	if len(diags) != 1 || diags[0].Code != diagnostics.NameUndefined || !strings.Contains(diags[0].Message, "'f'") {
		t.Errorf("got %v, want one undefined 'f'", diags)
	}
}

func TestExternMessages(t *testing.T) {
	diags := check(t, withExterns(`extern "rune:fs" int readFile(string p);`, ""))
	if len(diags) != 1 {
		t.Fatalf("got %v", diags)
	}
	want := `'readFile' of "rune:fs" has type 'string(string)', but this declaration says 'int(string)'`
	if diags[0].Message != want || diags[0].Help != `declare it as: extern "rune:fs" string readFile(string);` {
		t.Errorf("got %q / %q", diags[0].Message, diags[0].Help)
	}
}

func TestHostModuleWarning(t *testing.T) {
	got := warnings(t, withExterns(`extern "node:fs" bool existsSync(string p);`, `    trace(existsSync("a"));`))
	if strings.Join(got, ",") != "2:8 ASX-WARN-006" {
		t.Errorf("got %v", got)
	}
	if got := warnings(t, withExterns(`extern "rune:fs" bool exists(string p);`, `    trace(exists("a"));`)); len(got) != 0 {
		t.Errorf("rune:fs warned: %v", got)
	}
}

func writeFile(dir, name, src string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644)
}
