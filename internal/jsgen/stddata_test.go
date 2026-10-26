package jsgen

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// The std:csv, std:matrix, and std:semver parity tests: random calls run
// through the Go package and its JavaScript twin, and every result or
// error must be the same. Floats are compared bit for bit.

// call is one standard library call: a function and its arguments in JSON
// form. Floats are written as their IEEE 754 bits in hex ("f:..."), so
// that -0, NaN, and the infinities survive JSON.
type call struct {
	Pkg  string `json:"pkg"`
	Name string `json:"name"`
	Args []any  `json:"args"`
}

// goCall runs c in Go and summarizes the result like the JavaScript side.
func goCall(c call) string {
	f := stdlib.Lookup(c.Pkg).Funcs[c.Name]
	// Round-trip through JSON, as the arguments reach JavaScript.
	var raw []any
	data, _ := json.Marshal(c.Args)
	json.Unmarshal(data, &raw)
	args := make([]any, len(raw))
	for i, a := range raw {
		args[i] = goArg(a, f.Params[i])
	}
	var out any
	err := func() (msg string) {
		defer func() {
			if r := recover(); r != nil {
				ft, ok := r.(stdlib.Fault)
				if !ok {
					panic(r)
				}
				msg = ft.Type + ": " + ft.Message
			}
		}()
		out = f.Impl(args)
		return ""
	}()
	if err != "" {
		return err
	}
	b, _ := json.Marshal(toJSON(out))
	return string(b)
}

// goArg converts a JSON argument to the Go value of type t.
func goArg(a any, t *types.Type) any {
	switch {
	case t.Kind == types.String:
		return a.(string)
	case t.Kind == types.Char:
		return []rune(a.(string)[2:])[0]
	case t.Kind == types.Float:
		return hexFloat(a.(string))
	case t.Kind == types.Signed:
		return int64(a.(float64))
	case t.Kind == types.Array && t.Elem.Kind == types.String:
		out := []string{}
		for _, x := range a.([]any) {
			out = append(out, x.(string))
		}
		return out
	case t.Kind == types.Array && t.Elem.Kind == types.Float:
		out := []float64{}
		for _, x := range a.([]any) {
			out = append(out, hexFloat(x.(string)))
		}
		return out
	case t.Kind == types.Array && t.Elem.Kind == types.Array && t.Elem.Elem.Kind == types.String:
		out := [][]string{}
		for _, x := range a.([]any) {
			out = append(out, goArg(x, t.Elem).([]string))
		}
		return out
	case t.Kind == types.Array && t.Elem.Kind == types.Array && t.Elem.Elem.Kind == types.Float:
		out := [][]float64{}
		for _, x := range a.([]any) {
			out = append(out, goArg(x, t.Elem).([]float64))
		}
		return out
	}
	panic("goArg: " + t.String())
}

func hexFloat(s string) float64 {
	var bits uint64
	fmt.Sscanf(s[2:], "%x", &bits)
	return math.Float64frombits(bits)
}

// argHex writes an argument's bits, NaN included.
func argHex(f float64) string { return fmt.Sprintf("f:%016x", math.Float64bits(f)) }

// floatHex writes a result's bits. Every NaN is written the same: a NaN's sign
// and payload are not portable across hosts, so programs cannot observe
// them (see std:math).
func floatHex(f float64) string {
	if math.IsNaN(f) {
		return "f:NaN"
	}
	return fmt.Sprintf("f:%016x", math.Float64bits(f))
}

// toJSON writes a result with floats as hex bits.
func toJSON(v any) any {
	switch v := v.(type) {
	case float64:
		return floatHex(v)
	case int64:
		return v
	case []float64:
		out := make([]any, len(v))
		for i, x := range v {
			out[i] = floatHex(x)
		}
		return out
	case [][]float64:
		out := make([]any, len(v))
		for i, x := range v {
			out[i] = toJSON(x)
		}
		return out
	}
	return v
}

// jsCalls runs the calls in the JavaScript runtime and returns each
// summary.
func jsCalls(t *testing.T, calls []call) []string {
	t.Helper()
	nodePath, err := exec.LookPath("node")
	if err != nil {
		if os.Getenv("RUNESCRIPT_REQUIRE_NODE") != "" {
			t.Fatal("node is not on PATH and RUNESCRIPT_REQUIRE_NODE is set")
		}
		t.Skip("node is not on PATH")
	}
	data, _ := json.Marshal(calls)
	dir := t.TempDir()
	script := Runtime() + `
const calls = JSON.parse(require("node:fs").readFileSync(process.argv[2], "utf8"));
const dv = new DataView(new ArrayBuffer(8));
const fromHex = (s) => { dv.setBigUint64(0, BigInt("0x" + s.slice(2))); return dv.getFloat64(0); };
const toHex = (f) => { if (Number.isNaN(f)) return "f:NaN"; dv.setFloat64(0, f); return "f:" + dv.getBigUint64(0).toString(16).padStart(16, "0"); };
const arg = (a) => {
  if (typeof a === "string") {
    if (a.startsWith("f:")) return fromHex(a);
    if (a.startsWith("c:")) return a.codePointAt(2);
    return a;
  }
  if (Array.isArray(a)) return a.map(arg);
  return a;
};
const floats = (pkg, name) => pkg === "matrix" && !["rows", "cols", "format"].includes(name);
const out = (v, f) => (Array.isArray(v) ? v.map((x) => out(x, f)) : f && typeof v === "number" ? toHex(v) : v);
const results = calls.map((c) => {
  try {
    return JSON.stringify(out($rt.std[c.pkg][c.name](...c.args.map(arg)), floats(c.pkg, c.name)));
  } catch (e) {
    if (e && e.type) return e.type + ": " + e.message;
    throw e;
  }
});
process.stdout.write(JSON.stringify(results));
`
	in := filepath.Join(dir, "calls.json")
	js := filepath.Join(dir, "calls.cjs")
	os.WriteFile(in, data, 0o644)
	os.WriteFile(js, []byte(script), 0o644)
	raw, err := exec.Command(nodePath, js, in).Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var got []string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

// compareCalls runs calls on both targets and reports differences. It
// returns the number of calls that failed (on both targets alike).
func compareCalls(t *testing.T, calls []call) int {
	t.Helper()
	got := jsCalls(t, calls)
	diffs, faults := 0, 0
	for i, c := range calls {
		want := goCall(c)
		if !strings.HasPrefix(want, "[") && !strings.HasPrefix(want, `"`) && strings.Contains(want, ": ") {
			faults++
		}
		if got[i] != want {
			diffs++
			if diffs <= 10 {
				a, _ := json.Marshal(c.Args)
				t.Errorf("%s.%s(%s):\nJavaScript: %s\nGo:         %s", c.Pkg, c.Name, a, got[i], want)
			}
		}
	}
	if diffs > 0 {
		t.Errorf("%d of %d calls differ", diffs, len(calls))
	}
	return faults
}

func TestStdCSVMatchesGo(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	pieces := []string{"a", "b", "xy", "é", "→", ",", ",", ";", "\t", `"`, `"`, `""`, "\n", "\r\n", "\r", " ", string(rune(0xFEFF))}
	text := func() string {
		var b strings.Builder
		if r.Intn(8) == 0 {
			b.WriteString(string(rune(0xFEFF)))
		}
		for n := r.Intn(25); n > 0; n-- {
			b.WriteString(pieces[r.Intn(len(pieces))])
		}
		return b.String()
	}
	seps := []string{"c:,", "c:;", "c:\t", "c:→", `c:"`, "c:\n"}
	var calls []call
	for i := 0; i < 3000; i++ {
		calls = append(calls, call{"csv", "parse", []any{text()}})
		calls = append(calls, call{"csv", "parseWith", []any{text(), seps[r.Intn(len(seps))]}})
		var rows []any
		for n := r.Intn(4); n > 0; n-- {
			var rec []any
			for k := r.Intn(4); k > 0; k-- {
				rec = append(rec, strings.ReplaceAll(text(), string(rune(0xFEFF)), ""))
			}
			if rec == nil {
				rec = []any{}
			}
			rows = append(rows, rec)
		}
		if len(rows) > 0 {
			if r.Intn(2) == 0 {
				calls = append(calls, call{"csv", "format", []any{rows}})
			} else {
				calls = append(calls, call{"csv", "formatWith", []any{rows, seps[r.Intn(len(seps))]}})
			}
		}
	}
	if faults := compareCalls(t, calls); faults < 200 || faults > len(calls)-200 {
		t.Errorf("%d of %d calls failed: the generator should give plenty of both", faults, len(calls))
	}
}

func TestStdMatrixMatchesGo(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	special := []float64{0, math.Copysign(0, -1), 1, -1, 0.1, 1e300, -1e-300, math.Inf(1), math.NaN(), 3}
	num := func() any {
		if r.Intn(10) == 0 {
			return argHex(special[r.Intn(len(special))])
		}
		return argHex(float64(r.Intn(21)-10) + r.Float64())
	}
	mat := func(rows, cols int) []any {
		m := []any{}
		for i := 0; i < rows; i++ {
			row := []any{}
			for j := 0; j < cols; j++ {
				row = append(row, num())
			}
			m = append(m, row)
		}
		if r.Intn(25) == 0 && rows > 1 { // not rectangular
			m[1] = append(m[1].([]any), num())
		}
		return m
	}
	vec := func(n int) []any {
		v := []any{}
		for i := 0; i < n; i++ {
			v = append(v, num())
		}
		return v
	}
	var calls []call
	for i := 0; i < 2000; i++ {
		n, m, k := r.Intn(6), r.Intn(6), r.Intn(6)
		if r.Intn(3) > 0 {
			m = n
		}
		a := mat(n, m)
		calls = append(calls,
			call{"matrix", "det", []any{a}},
			call{"matrix", "inverse", []any{a}},
			call{"matrix", "solve", []any{a, vec(n + r.Intn(2))}},
			call{"matrix", "mul", []any{a, mat(m+r.Intn(5)/4, k)}},
			call{"matrix", "mulVec", []any{a, vec(m)}},
			call{"matrix", "add", []any{a, mat(n, m)}},
			call{"matrix", "sub", []any{a, mat(n, m+r.Intn(2))}},
			call{"matrix", "scale", []any{a, num()}},
			call{"matrix", "transpose", []any{a}},
			call{"matrix", "format", []any{a}},
			call{"matrix", "rows", []any{a}},
			call{"matrix", "cols", []any{a}},
			call{"matrix", "zeros", []any{r.Intn(4) - 1, r.Intn(4)}},
			call{"matrix", "identity", []any{r.Intn(4) - 1}},
		)
	}
	compareCalls(t, calls)
}

func TestStdSemverMatchesGo(t *testing.T) {
	r := rand.New(rand.NewSource(9))
	num := func() string {
		switch r.Intn(15) {
		case 0:
			return "01"
		case 1:
			return "2147483648"
		case 2:
			return "123456789012345678901234567890"
		}
		return fmt.Sprint(r.Intn(4))
	}
	ids := []string{"alpha", "beta", "rc", "0", "1", "11", "x-y", "-", "01", ""}
	version := func() string {
		v := num() + "." + num() + "." + num()
		if r.Intn(12) == 0 {
			v = num() + "." + num()
		}
		if r.Intn(2) == 0 {
			var p []string
			for n := 1 + r.Intn(3); n > 0; n-- {
				p = append(p, ids[r.Intn(len(ids))])
			}
			v += "-" + strings.Join(p, ".")
		}
		if r.Intn(5) == 0 {
			v += "+b." + fmt.Sprint(r.Intn(9))
		}
		return v
	}
	ops := []string{"", "", "=", "<", "<=", ">", ">=", "~", "^", "~>", "~ "}
	partial := func() string {
		parts := []string{}
		for n := 1 + r.Intn(3); n > 0; n-- {
			if r.Intn(5) == 0 {
				parts = append(parts, []string{"x", "X", "*"}[r.Intn(3)])
			} else {
				parts = append(parts, fmt.Sprint(r.Intn(3)))
			}
		}
		s := strings.Join(parts, ".")
		if len(parts) == 3 && r.Intn(3) == 0 {
			s += "-rc." + fmt.Sprint(r.Intn(3))
		}
		return s
	}
	rng := func() string {
		var sets []string
		for n := 1 + r.Intn(2); n > 0; n-- {
			if r.Intn(6) == 0 {
				sets = append(sets, partial()+" - "+partial())
				continue
			}
			var cs []string
			for k := 1 + r.Intn(2); k > 0; k-- {
				cs = append(cs, ops[r.Intn(len(ops))]+partial())
			}
			sets = append(sets, strings.Join(cs, " "))
		}
		s := strings.Join(sets, " || ")
		if r.Intn(30) == 0 {
			s += " v"
		}
		return s
	}
	var calls []call
	for i := 0; i < 4000; i++ {
		a, b := version(), version()
		name := []string{"compare", "eq", "neq", "gt", "gte", "lt", "lte"}[r.Intn(7)]
		calls = append(calls,
			call{"semver", "valid", []any{a}},
			call{"semver", name, []any{a, b}},
			call{"semver", "satisfies", []any{a, rng()}},
			call{"semver", "validRange", []any{rng()}},
			call{"semver", []string{"major", "minor", "patch", "prerelease", "build"}[r.Intn(5)], []any{a}},
		)
	}
	compareCalls(t, calls)
}
