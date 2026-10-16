package jsgen

import (
	"encoding/json"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"
	"unicode/utf8"

	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

// alphabet mixes ASCII with characters whose handling differs between
// JavaScript and Go: multi-byte UTF-8, astral code points (surrogate pairs
// in JavaScript), special casing, case-fold orbits, and white space that
// only one of the two languages treats as space.
var alphabet = []string{
	"a", "b", "A", "S", "s", "k", "K", "\u212a", " ", "\t", "\n", "-", "\u00e9", "\u00c9", "\u00df", "\u1e9e",
	"\u017f", "\u01c5", "\u01c6", "\u01c4", "\u0130", "\u0131", "i", "\u03a3", "\u03c3", "\u03c2", "\ufb01", "\u20ac",
	"\U0001f600", "\U00010400", "\U00010428", "\u0085", "\u00a0", "\u2028", "\u3000", "\ufeff", "\ufffd",
	"\uffff", "\ue000",
}

func randomString(r *rand.Rand) string {
	s := ""
	for n := r.Intn(6); n > 0; n-- {
		s += alphabet[r.Intn(len(alphabet))]
	}
	return s
}

// callbacks are the char callbacks the random cases pass, written in Go
// and, identically, in JavaScript (callbacksJS).
var callbacks = map[string]any{
	"lowerASCII": func(r rune) bool { return r >= 'a' && r <= 'z' },
	"isA":        func(r rune) bool { return r == 'a' },
	"nonASCII":   func(r rune) bool { return r > 127 },
	"always":     func(rune) bool { return true },
	"never":      func(rune) bool { return false },
	"upperASCII": func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 32
		}
		return r
	},
	"astral": func(r rune) rune {
		if r == 'a' {
			return 0x1F600
		}
		return r
	},
	"same": func(r rune) rune { return r },
}

const callbacksJS = `const callbacks = {
  lowerASCII: (c) => c >= 97 && c <= 122,
  isA: (c) => c === 97,
  nonASCII: (c) => c > 127,
  always: () => true,
  never: () => false,
  upperASCII: (c) => (c >= 97 && c <= 122 ? c - 32 : c),
  astral: (c) => (c === 97 ? 0x1F600 : c),
  same: (c) => c,
};
`

var predicates = []string{"lowerASCII", "isA", "nonASCII", "always", "never"}
var mappers = []string{"upperASCII", "astral", "same"}

type stdCase struct {
	Fn   string `json:"fn"`
	Args []any  `json:"args"`
	// Want is the interpreter's result; Fails means it raised an error.
	Want  any  `json:"want"`
	Fails bool `json:"fails"`
}

// TestStdStringMatchesGo calls every std:string function in the JavaScript
// runtime with random arguments and compares each result with the
// interpreter's Go implementation.
func TestStdStringMatchesGo(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		if os.Getenv("RUNESCRIPT_REQUIRE_NODE") != "" {
			t.Fatal("node is not on PATH and RUNESCRIPT_REQUIRE_NODE is set")
		}
		t.Skip("node is not on PATH")
	}
	pkg := stdlib.Lookup("string")
	var names []string
	for name := range pkg.Funcs {
		names = append(names, name)
	}
	sort.Strings(names)

	r := rand.New(rand.NewSource(1))
	var cases []stdCase
	for _, name := range names {
		f := pkg.Funcs[name]
		for i := 0; i < 300; i++ {
			args := make([]any, len(f.Params))   // for JSON
			goArgs := make([]any, len(f.Params)) // for Impl
			for j, p := range f.Params {
				switch p.Kind {
				case types.String:
					args[j] = randomString(r)
				case types.Array: // Array<string>
					list := []string{}
					for n := r.Intn(4); n > 0; n-- {
						list = append(list, randomString(r))
					}
					args[j] = list
				case types.Char:
					c, _ := utf8.DecodeRuneInString(alphabet[r.Intn(len(alphabet))])
					args[j], goArgs[j] = c, c
					continue
				case types.Func:
					names := predicates
					if p.Result.Kind == types.Char {
						names = mappers
					}
					name := names[r.Intn(len(names))]
					args[j], goArgs[j] = map[string]string{"callback": name}, callbacks[name]
					continue
				default:
					args[j] = int64(r.Intn(6) - 2) // -2 .. 3
				}
				goArgs[j] = args[j]
			}
			cases = append(cases, goCase(f, args, goArgs))
		}
	}

	data, _ := json.Marshal(cases)
	dir := t.TempDir()
	script := Runtime() + callbacksJS + `
const cases = JSON.parse(require("node:fs").readFileSync(process.argv[2], "utf8"));
const arg = (a) => (a !== null && typeof a === "object" && !Array.isArray(a) ? callbacks[a.callback] : a);
const results = cases.map((c) => {
  try {
    return { got: $rt.std.string[c.fn](...c.args.map(arg)) };
  } catch (e) {
    return { fails: true, message: String(e) };
  }
});
process.stdout.write(JSON.stringify(results));
`
	casesPath := filepath.Join(dir, "cases.json")
	scriptPath := filepath.Join(dir, "conformance.cjs")
	if err := os.WriteFile(casesPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(nodePath, scriptPath, casesPath).Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var results []struct {
		Got   any  `json:"got"`
		Fails bool `json:"fails"`
	}
	if err := json.Unmarshal(out, &results); err != nil {
		t.Fatal(err)
	}
	failures := 0
	for i, c := range cases {
		got := results[i]
		// Compare as JSON: numbers decode as float64, arrays as []any.
		g, _ := json.Marshal(got.Got)
		w, _ := json.Marshal(c.Want)
		if got.Fails != c.Fails || (!c.Fails && string(g) != string(w)) {
			failures++
			if failures <= 20 {
				t.Errorf("string.%s%q: JavaScript gave %#v (fails=%v), Go gave %#v (fails=%v)",
					c.Fn, c.Args, got.Got, got.Fails, c.Want, c.Fails)
			}
		}
	}
	if failures > 0 {
		t.Errorf("%d of %d cases differ", failures, len(cases))
	}
}

func goCase(f *stdlib.Func, args, goArgs []any) (c stdCase) {
	c = stdCase{Fn: f.Name, Args: args}
	defer func() {
		if recover() != nil {
			c.Fails, c.Want = true, nil
		}
	}()
	c.Want = f.Impl(goArgs)
	return c
}
