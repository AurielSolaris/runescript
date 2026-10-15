package conv

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var point1, point2 = 0.1, 0.2 // variables, so the sum is not folded exactly

func TestFormatFloat(t *testing.T) {
	cases := []struct {
		f    float64
		bits int
		want string
	}{
		{0, 64, "0"},
		{math.Copysign(0, -1), 64, "0"},
		{64, 64, "64"},
		{64.3, 64, "64.3"},
		{-64.7, 64, "-64.7"},
		{point1 + point2, 64, "0.30000000000000004"},
		{1e21, 64, "1e+21"},
		{123456789012345680000, 64, "123456789012345680000"},
		{1e-6, 64, "0.000001"},
		{1e-7, 64, "1e-7"},
		{1.5e-7, 64, "1.5e-7"},
		{math.MaxFloat64, 64, "1.7976931348623157e+308"},
		{ToFloat32(0.1), 32, "0.1"},
		{ToFloat32(0.1), 64, "0.10000000149011612"},
		{ToFloat32(16777217), 32, "16777216"},
		{math.MaxFloat32, 32, "3.4028235e+38"},
		{math.Ldexp(1, 87), 32, "1.5474251e+26"},  // a power of two: the closest 8 digits don't round-trip
		{math.Ldexp(1, -12), 32, "0.00024414062"}, // an exact tie between ...62 and ...63: even wins
		{math.Ldexp(1, -149), 32, "1e-45"},        // the smallest float32
		{math.NaN(), 64, "NaN"},
		{math.Inf(1), 32, "Infinity"},
		{math.Inf(-1), 64, "-Infinity"},
		{math.NaN(), 64, "NaN"},
		{math.Inf(1), 32, "Infinity"},
		{math.Inf(-1), 64, "-Infinity"},
	}
	for _, c := range cases {
		if got := FormatFloat(c.f, c.bits); got != c.want {
			t.Errorf("FormatFloat(%v, %d) = %q, want %q", c.f, c.bits, got, c.want)
		}
	}
}

func TestParse(t *testing.T) {
	for s, want := range map[string]bool{"42": true, "-7": true, "+3": true, "": false, " 1": false, "1.0": false, "0x10": false, "1_000": false} {
		if _, _, ok := ParseInt(s); ok != want {
			t.Errorf("ParseInt(%q) ok = %v", s, ok)
		}
	}
	if _, fits, ok := ParseInt("99999999999999999999"); !ok || fits {
		t.Errorf("huge integer: fits=%v ok=%v", fits, ok)
	}
	for s, want := range map[string]bool{"64.7": true, "-1e5": true, "42": true, "1.5E+3": true,
		".5": false, "5.": false, "inf": false, "NaN": true, "-Infinity": true, "nan": false, "-NaN": false, "0x1p3": false, "1e": false, " 1": false} {
		if _, ok := ParseFloat(s); ok != want {
			t.Errorf("ParseFloat(%q) ok = %v", s, ok)
		}
	}
	if v, ok := ParseFloat("1e400"); !ok || !math.IsInf(v, 1) {
		t.Errorf("1e400 = %v, %v", v, ok)
	}
}

// TestFormatMatchesJavaScript compares FormatFloat with JavaScript's
// String(x), and with the float32 formatting of the JavaScript runtime
// (internal/jsgen), on random values.
func TestFormatMatchesJavaScript(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		if os.Getenv("RUNESCRIPT_REQUIRE_NODE") != "" {
			t.Fatal("node is not on PATH and RUNESCRIPT_REQUIRE_NODE is set")
		}
		t.Skip("node is not on PATH")
	}
	r := rand.New(rand.NewSource(1))
	var vals []float64
	for i := 0; i < 4000; i++ {
		var f float64
		switch i % 4 {
		case 0:
			f = math.Float64frombits(r.Uint64())
		case 1:
			f = r.NormFloat64() * math.Pow(10, float64(r.Intn(60)-30))
		case 2:
			f = math.Ldexp(1, r.Intn(250)-125) // powers of two: asymmetric rounding gaps
		default:
			f = float64(r.Intn(2000000)-1000000) / 1000
		}
		if math.IsNaN(f) || math.IsInf(f, 0) {
			continue
		}
		vals = append(vals, f)
	}
	data, _ := json.Marshal(vals)
	dir := t.TempDir()
	in := filepath.Join(dir, "vals.json")
	script := filepath.Join(dir, "fmt.cjs")
	if err := os.WriteFile(in, data, 0o644); err != nil {
		t.Fatal(err)
	}
	// f32 mirrors $rt.f32str in internal/jsgen/runtime.js.
	js := `
const vals = JSON.parse(require("node:fs").readFileSync(process.argv[2], "utf8"));
` + Float32JS + `
const out = vals.map((x) => {
  const y = Math.fround(x);
  return [String(x), isFinite(y) ? f32str(y) : null];
});
process.stdout.write(JSON.stringify(out));
`
	if err := os.WriteFile(script, []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(nodePath, script, in).Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var got [][2]*string
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	bad := 0
	for i, f := range vals {
		if want := FormatFloat(f, 64); *got[i][0] != want {
			bad++
			if bad < 10 {
				t.Errorf("float64 %v: JavaScript %q, Go %q", f, *got[i][0], want)
			}
		}
		y := ToFloat32(f)
		if math.IsInf(y, 0) {
			continue
		}
		if want := FormatFloat(y, 32); got[i][1] == nil || *got[i][1] != want {
			bad++
			if bad < 10 {
				js := "<nil>"
				if got[i][1] != nil {
					js = *got[i][1]
				}
				t.Errorf("float32 %v: JavaScript %q, Go %q", y, js, want)
			}
		}
	}
	if bad > 0 {
		t.Errorf("%d mismatches", bad)
	}
}
