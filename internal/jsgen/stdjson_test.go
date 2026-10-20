package jsgen

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/conv"
	"github.com/AurielSolaris/runescript/stdlib"
)

// TestStdJSONMatchesGo runs random JSON texts, valid and broken, through
// std:json on both targets. Each text is parsed, then summarized: the
// compact and indented output, and every number read with asFloat, asInt,
// and asInt64. A failure is summarized by its error type and message, so
// the byte offsets in parse errors must agree too.
func TestStdJSONMatchesGo(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		if os.Getenv("RUNESCRIPT_REQUIRE_NODE") != "" {
			t.Fatal("node is not on PATH and RUNESCRIPT_REQUIRE_NODE is set")
		}
		t.Skip("node is not on PATH")
	}
	r := rand.New(rand.NewSource(7))
	var texts []string
	for i := 0; i < 3000; i++ {
		text := randomJSON(r, 0)
		if i%3 == 1 {
			text = corrupt(r, text)
		}
		texts = append(texts, text)
	}
	texts = append(texts, jsonEdgeCases...)

	var want []string
	for _, text := range texts {
		want = append(want, goJSONSummary(text))
	}

	data, _ := json.Marshal(texts)
	dir := t.TempDir()
	script := Runtime() + `
const texts = JSON.parse(require("node:fs").readFileSync(process.argv[2], "utf8"));
const J = $rt.std.json;
const attempt = (f) => {
  try {
    return String(f());
  } catch (e) {
    if (e && e.type) return e.type + ": " + e.message;
    throw e;
  }
};
function summary(v) {
  switch (J.kind(v)) {
    case "number":
      return "<" + J.stringify(v) + " " + attempt(() => J.asFloat(v)) + " " + attempt(() => J.asInt(v)) + " " + attempt(() => J.asInt64(v)) + ">";
    case "array":
      return "[" + J.items(v).map(summary).join(",") + "]";
    case "object":
      return "{" + J.keys(v).map((k) => J.stringify(J.fromString(k)) + "=" + summary(J.get(v, k))).join(",") + "}";
  }
  return J.kind(v) + ":" + J.stringify(v);
}
const results = texts.map((text) => attempt(() => {
  const v = J.parse(text);
  return J.stringify(v) + "\n" + J.stringifyIndent(v, "\t") + "\n" + summary(v);
}));
process.stdout.write(JSON.stringify(results));
`
	textsPath := filepath.Join(dir, "texts.json")
	scriptPath := filepath.Join(dir, "json.cjs")
	if err := os.WriteFile(textsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(nodePath, scriptPath, textsPath).Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var got []string
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	failures, parsed := 0, 0
	for i, text := range texts {
		if !strings.Contains(want[i], ": ") || strings.Contains(want[i], "\n") {
			parsed++
		}
		if got[i] != want[i] {
			failures++
			if failures <= 10 {
				t.Errorf("json %q:\nJavaScript: %s\nGo:         %s", text, got[i], want[i])
			}
		}
	}
	if failures > 0 {
		t.Errorf("%d of %d texts differ", failures, len(texts))
	}
	if parsed < len(texts)/2 {
		t.Errorf("only %d of %d texts parsed: the generator is too broken to test much", parsed, len(texts))
	}
}

// jsonEdgeCases are texts the random ones are unlikely to hit.
var jsonEdgeCases = []string{
	"", " ", "nul", "nulls", "tru", "-", "-0", "0", "01", "1.", ".5", "1e", "1e+", "1E-7", "-0.0e0",
	"2147483647", "2147483648", "-2147483648", "-2147483649", "9223372036854775807", "9223372036854775808",
	"-9223372036854775808", "1e400", "-1e400", "1e-400", "4.9e-324", "0.1", "1.7976931348623157e308",
	"123456789012345678901234567890", `"\ud83d\ude00"`, `"\ud83d"`, `"\ude00"`, `"\ud83dx"`, `"\ud83d\u0041"`,
	`"\ud83d\ud83d\ude00"`, `"\u00"`, `"\u00g0"`, `"\x"`, `"a`, `"\`, "\"\t\"", "\"\x7f\"", `"\u0000\u001f"`,
	"[", "[1,]", "[,1]", "{", `{"a"}`, `{"a":}`, `{"a":1,}`, `{a:1}`, `{"a":1 "b":2}`, `{"a":1,"a":2,"b":3}`,
	"\ufeff1", "1 2", "[1]x", "  [ ]  ", "{ }", "\u00e9", "[\"\u2028\u2029\"]", `"\/"`,
	strings.Repeat("[", 1000) + strings.Repeat("]", 1000), strings.Repeat("[", 1001) + strings.Repeat("]", 1001),
	strings.Repeat(`{"a":`, 1001) + "1" + strings.Repeat("}", 1001),
}

func goJSONSummary(text string) string {
	pkg := stdlib.Lookup("json")
	call := func(name string, args ...any) any { return pkg.Funcs[name].Impl(args) }
	attempt := func(f func() string) (out string) {
		defer func() {
			if r := recover(); r != nil {
				out = r.(stdlib.Fault).Type + ": " + r.(stdlib.Fault).Message
			}
		}()
		return f()
	}
	str := func(f func() any) string {
		return attempt(func() string {
			switch x := f().(type) {
			case float64:
				return fmtNumber(x)
			case int64:
				return fmt.Sprint(x)
			default:
				return fmt.Sprint(x)
			}
		})
	}
	var summary func(v any) string
	summary = func(v any) string {
		switch call("kind", v) {
		case "number":
			return "<" + call("stringify", v).(string) + " " + str(func() any { return call("asFloat", v) }) + " " +
				str(func() any { return call("asInt", v) }) + " " + str(func() any { return call("asInt64", v) }) + ">"
		case "array":
			var parts []string
			for _, e := range call("items", v).([]any) {
				parts = append(parts, summary(e))
			}
			return "[" + strings.Join(parts, ",") + "]"
		case "object":
			var parts []string
			for _, k := range call("keys", v).([]string) {
				parts = append(parts, call("stringify", call("fromString", k)).(string)+"="+summary(call("get", v, k)))
			}
			return "{" + strings.Join(parts, ",") + "}"
		}
		return call("kind", v).(string) + ":" + call("stringify", v).(string)
	}
	return attempt(func() string {
		v := call("parse", text)
		return call("stringify", v).(string) + "\n" + call("stringifyIndent", v, "\t").(string) + "\n" + summary(v)
	})
}

// fmtNumber is JavaScript's String(x) for a float64.
func fmtNumber(x float64) string { return conv.FormatFloat(x, 64) }

// randomJSON writes a random JSON value, with random spacing and string
// escapes, as a person or another program might.
func randomJSON(r *rand.Rand, depth int) string {
	sp := func() string { return []string{"", "", " ", "\n  ", "\t", "\r\n"}[r.Intn(6)] }
	n := r.Intn(10)
	if depth == 0 && r.Intn(10) < 7 {
		n = 6 + r.Intn(4) // mostly an array or an object at the top
	}
	if depth > 3 {
		n = r.Intn(6)
	}
	switch n {
	case 0:
		return "null"
	case 1:
		return []string{"true", "false"}[r.Intn(2)]
	case 2, 3:
		return randomNumber(r)
	case 4, 5:
		return randomJSONString(r)
	case 6, 7:
		var parts []string
		for k := r.Intn(4); k > 0; k-- {
			parts = append(parts, sp()+randomJSON(r, depth+1)+sp())
		}
		return "[" + strings.Join(parts, ",") + sp() + "]"
	default:
		var parts []string
		for k := r.Intn(4); k > 0; k-- {
			key := randomJSONString(r)
			if r.Intn(4) == 0 {
				key = `"a"` // repeated keys
			}
			parts = append(parts, sp()+key+sp()+":"+sp()+randomJSON(r, depth+1)+sp())
		}
		return "{" + strings.Join(parts, ",") + sp() + "}"
	}
}

func randomNumber(r *rand.Rand) string {
	var b strings.Builder
	if r.Intn(3) == 0 {
		b.WriteByte('-')
	}
	switch r.Intn(4) {
	case 0:
		b.WriteByte('0')
	case 1:
		fmt.Fprint(&b, r.Intn(100))
	case 2:
		fmt.Fprint(&b, r.Int63())
	default:
		fmt.Fprintf(&b, "%d%d", r.Int63(), r.Int63()) // beyond int64
	}
	if r.Intn(3) == 0 {
		fmt.Fprintf(&b, ".%d", r.Intn(100000))
	}
	if r.Intn(4) == 0 {
		fmt.Fprintf(&b, "%s%s%d", []string{"e", "E"}[r.Intn(2)], []string{"", "+", "-"}[r.Intn(3)], r.Intn(400))
	}
	return b.String()
}

func randomJSONString(r *rand.Rand) string {
	pieces := []string{"a", "Z", " ", "é", "€", "😀", "\u2028", `\"`, `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`,
		`\u0041`, `\u00e9`, `\u20AC`, `\ud83d\ude00`, `\ud800`, `\udfff`, `\u0000`, `\u001F`, "\x7f"}
	var b strings.Builder
	b.WriteByte('"')
	for k := r.Intn(5); k > 0; k-- {
		b.WriteString(pieces[r.Intn(len(pieces))])
	}
	b.WriteByte('"')
	return b.String()
}

// corrupt breaks text a little: a character removed, doubled, or replaced
// by one that often appears in JSON.
func corrupt(r *rand.Rand, text string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return "x"
	}
	i := r.Intn(len(runes))
	switch r.Intn(3) {
	case 0:
		return string(runes[:i]) + string(runes[i+1:])
	case 1:
		return string(runes[:i+1]) + string(runes[i:])
	}
	repl := []rune(`,:[]{}"\ 0-.eé` + "\n")
	return string(runes[:i]) + string(repl[r.Intn(len(repl))]) + string(runes[i+1:])
}
