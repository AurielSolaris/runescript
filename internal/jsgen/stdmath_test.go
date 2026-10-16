package jsgen

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// floatEdges are inputs where float functions disagree most easily.
var floatEdges = []float64{0, math.Copysign(0, -1), 0.5, -0.5, 1.5, -1.5, 2.5, -2.5, 0.49999999999999994,
	1, -1, math.Inf(1), math.Inf(-1), math.NaN(), math.MaxFloat64, -math.MaxFloat64,
	math.SmallestNonzeroFloat64, 2.2250738585072014e-308, 4503599627370495.5, 9007199254740993, 1e300, 7, -7}

func randomFloat(r *rand.Rand) float64 {
	switch r.Intn(4) {
	case 0:
		return floatEdges[r.Intn(len(floatEdges))]
	case 1:
		return float64(r.Intn(21)-10) / 2 // halves
	case 2:
		return (r.Float64()*2 - 1) * 1000
	}
	return math.Ldexp(r.Float64()*2-1, r.Intn(2100)-1050)
}

// TestStdMathMatchesGo calls every std:math function in the JavaScript
// runtime with random and edge-case arguments and requires bit-identical
// results to the interpreter's Go implementation. Floats cross as their
// bit patterns, so -0 and NaN compare exactly (any NaN equals any NaN:
// a NaN's sign is not observable).
func TestStdMathMatchesGo(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		if os.Getenv("RUNESCRIPT_REQUIRE_NODE") != "" {
			t.Fatal("node is not on PATH and RUNESCRIPT_REQUIRE_NODE is set")
		}
		t.Skip("node is not on PATH")
	}
	type mathCase struct {
		Fn    string   `json:"fn"`
		Args  []string `json:"args"` // floats as bits, ints as decimal
		Kinds []string `json:"kinds"`
		Want  string   `json:"want"`
		Fails bool     `json:"fails"`
	}
	enc := func(v any) string {
		switch v := v.(type) {
		case float64:
			if math.IsNaN(v) {
				return "NaN"
			}
			return "f" + strconv.FormatUint(math.Float64bits(v), 10)
		case bool:
			return strconv.FormatBool(v)
		case int64:
			return strconv.FormatInt(v, 10)
		}
		panic("unexpected value")
	}
	r := rand.New(rand.NewSource(3))
	var cases []mathCase
	for _, f := range stdlib.Lookup("math").All() {
		for i := 0; i < 400; i++ {
			c := mathCase{Fn: f.JSName}
			args := make([]any, len(f.Params))
			for j, p := range f.Params {
				if p.Kind == types.Float {
					args[j] = randomFloat(r)
					c.Kinds = append(c.Kinds, "f")
				} else {
					args[j] = []int64{math.MinInt32, math.MaxInt32, 0, -1, 1, 5, -5, 100}[r.Intn(8)]
					c.Kinds = append(c.Kinds, "i")
				}
				c.Args = append(c.Args, enc(args[j]))
			}
			func() {
				defer func() {
					if recover() != nil {
						c.Fails = true
					}
				}()
				c.Want = enc(f.Impl(args))
			}()
			cases = append(cases, c)
		}
	}
	data, _ := json.Marshal(cases)
	dir := t.TempDir()
	script := Runtime() + `
const F = new Float64Array(1), B = new BigUint64Array(F.buffer);
const dec = (s, k) => k === "i" ? Number(s) : s === "NaN" ? NaN : (B[0] = BigInt(s.slice(1)), F[0]);
const enc = (v) => typeof v === "boolean" ? String(v) : Number.isNaN(v) ? "NaN" : (F[0] = v, "f" + B[0]);
const cases = JSON.parse(require("node:fs").readFileSync(process.argv[2], "utf8"));
process.stdout.write(JSON.stringify(cases.map((c) => {
  try {
    const r = $rt.std.math[c.fn](...(c.args || []).map((a, i) => dec(a, c.kinds[i])));
    return { got: c.fn.endsWith("Int") ? String(r) : enc(r) };
  } catch (e) {
    return { fails: true };
  }
})));
`
	casesPath := filepath.Join(dir, "cases.json")
	scriptPath := filepath.Join(dir, "math.cjs")
	os.WriteFile(casesPath, data, 0o644)
	os.WriteFile(scriptPath, []byte(script), 0o644)
	out, err := exec.Command(nodePath, scriptPath, casesPath).Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var results []struct {
		Got   string `json:"got"`
		Fails bool   `json:"fails"`
	}
	if err := json.Unmarshal(out, &results); err != nil {
		t.Fatal(err)
	}
	failures := 0
	for i, c := range cases {
		got := results[i]
		// The int helpers' results are range-checked by the caller, not
		// here, so a Go result outside int still compares as a number.
		if got.Fails != c.Fails || (!c.Fails && got.Got != c.Want) {
			failures++
			if failures <= 20 {
				t.Errorf("math.%s%v: JavaScript gave %s (fails=%v), Go gave %s (fails=%v)", c.Fn, c.Args, got.Got, got.Fails, c.Want, c.Fails)
			}
		}
	}
	if failures > 0 {
		t.Errorf("%d of %d cases differ", failures, len(cases))
	}
}
