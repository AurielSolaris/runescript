package tests

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/jsgen"
)

// Generated JavaScript is limited to ES2015 plus BigInt, `??`, and ES
// modules (import/export and import.meta) (owner decision, OQ-JS-07), so
// that the planned bundled runtime only has to implement that much.
//
// This is a lint, not a parser: it blanks comments and the contents of
// strings, templates, and regular expressions, then looks for syntax and
// built-ins added after ES2015. It cannot tell an array's includes
// (ES2016) from a string's (ES2015), and does not detect class fields.

type jsRule struct {
	re   *regexp.Regexp
	what string
}

// jsCodeRules are the constructs outside the allowed level, with the
// edition that added them.
var jsCodeRules = []jsRule{
	{regexp.MustCompile(`\?\?=|\|\|=|&&=`), "logical assignment (ES2021)"},
	{regexp.MustCompile(`\?\.[^0-9]`), "optional chaining (ES2020)"},
	{regexp.MustCompile(`\*\*`), "exponentiation (ES2016)"},
	{regexp.MustCompile(`\b(async|await)\b`), "async functions (ES2017)"},
	{regexp.MustCompile(`#[A-Za-z_$]`), "private class members (ES2022)"},
	{regexp.MustCompile(`\bstatic\s*\{`), "static blocks (ES2022)"},
	{regexp.MustCompile(`\bcatch\s*\{`), "catch without a binding (ES2019)"},
	{regexp.MustCompile(`\b\d+_\d`), "numeric separators (ES2021)"},
	// Only `{ ...x`: spread after a comma is usually a call's (ES2015).
	{regexp.MustCompile(`\{\s*\.\.\.`), "object spread or rest (ES2018)"},
	{regexp.MustCompile(`\bimport\s*\(`), "dynamic import (ES2020)"},
	{regexp.MustCompile(`\bexport\s*\*\s*as\b`), "export * as (ES2020)"},
	// Only globalThis.x or globalThis[x]: a RuneScript variable may be
	// called globalThis (tests/testdata/js/names.rsc).
	{regexp.MustCompile(`\bglobalThis\s*[.\[]`), "globalThis (ES2020)"},
	{regexp.MustCompile(`\b(structuredClone|WeakRef|FinalizationRegistry|AggregateError)\b`), "newer global (ES2021+)"},
	{regexp.MustCompile(`\bObject\.(entries|values|fromEntries|hasOwn|getOwnPropertyDescriptors)\b`), "newer Object function (ES2017+)"},
	{regexp.MustCompile(`\bPromise\.(allSettled|any)\b`), "newer Promise function (ES2020+)"},
	{regexp.MustCompile(`(?:^|[^.])\.(at|padStart|padEnd|trimStart|trimEnd|matchAll|replaceAll|flat|flatMap|findLast|findLastIndex)\(`), "newer method (ES2017+)"},
}

// jsRegexpRules apply to regular expression literals.
var jsRegexpRules = []jsRule{
	{regexp.MustCompile(`\(\?<`), "regular expression lookbehind or named group (ES2018)"},
	{regexp.MustCompile(`/[a-z]*[sd][a-z]*$`), "regular expression flag s or d (ES2018+)"},
}

// jsCode returns src with comments blanked and the contents of strings and
// templates blanked (the code inside ${...} is kept), and the regular
// expressions found. Newlines are kept so line numbers stay right.
func jsCode(src string) (code string, regexps []string) {
	var b strings.Builder
	blank := func(c byte) {
		if c == '\n' {
			b.WriteByte('\n')
		} else {
			b.WriteByte(' ')
		}
	}
	var templates []int // brace depth at each open ${, innermost last
	depth := 0
	prev := byte('(') // last significant character, for telling / apart
	inTemplate := func(i int) int {
		// scan template text from i; return the index after it ends or
		// after the ${ that interrupts it
		for i < len(src) {
			switch {
			case src[i] == '\\' && i+1 < len(src):
				blank(src[i])
				blank(src[i+1])
				i += 2
			case src[i] == '`':
				b.WriteByte('`')
				return i + 1
			case src[i] == '$' && i+1 < len(src) && src[i+1] == '{':
				b.WriteString("${")
				templates = append(templates, depth)
				depth++
				return i + 2
			default:
				blank(src[i])
				i++
			}
		}
		return i
	}
	for i := 0; i < len(src); {
		c := src[i]
		switch {
		case strings.HasPrefix(src[i:], "//"):
			for i < len(src) && src[i] != '\n' {
				blank(src[i])
				i++
			}
			continue
		case strings.HasPrefix(src[i:], "/*"):
			end := strings.Index(src[i+2:], "*/")
			if end < 0 {
				end = len(src) - i - 4
			}
			for _, ch := range []byte(src[i : i+end+4]) {
				blank(ch)
			}
			i += end + 4
			continue
		case c == '"' || c == '\'':
			b.WriteByte(c)
			j := i + 1
			for j < len(src) && src[j] != c && src[j] != '\n' {
				if src[j] == '\\' {
					blank(src[j])
					j++
				}
				blank(src[j])
				j++
			}
			b.WriteByte(c)
			i, prev = j+1, 'a'
			continue
		case c == '`':
			b.WriteByte('`')
			i, prev = inTemplate(i+1), 'a'
			continue
		case c == '}' && len(templates) > 0 && templates[len(templates)-1] == depth-1:
			templates = templates[:len(templates)-1]
			depth--
			b.WriteByte('}')
			i, prev = inTemplate(i+1), 'a'
			continue
		case c == '/' && strings.IndexByte("(,=:[!&|?{};+-*%<>~^", prev) >= 0:
			j, class := i+1, false
			for j < len(src) && src[j] != '\n' && (src[j] != '/' || class) {
				switch src[j] {
				case '\\':
					j++
				case '[':
					class = true
				case ']':
					class = false
				}
				j++
			}
			for j++; j < len(src) && src[j] >= 'a' && src[j] <= 'z'; j++ {
			}
			regexps = append(regexps, src[i:j])
			b.WriteString(strings.Repeat(" ", j-i))
			i, prev = j, 'a'
			continue
		case c == '{':
			depth++
		case c == '}':
			depth--
		}
		b.WriteByte(c)
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			prev = c
		}
		i++
	}
	return b.String(), regexps
}

// jsLevelViolations returns one line per construct in src outside the
// allowed language level.
func jsLevelViolations(src string) []string {
	code, regexps := jsCode(src)
	var out []string
	for n, line := range strings.Split(code, "\n") {
		for _, r := range jsCodeRules {
			if m := r.re.FindString(line); m != "" {
				out = append(out, r.what+": "+strings.TrimSpace(m)+" on line "+strconv.Itoa(n+1))
			}
		}
	}
	for _, re := range regexps {
		for _, r := range jsRegexpRules {
			if r.re.MatchString(re) {
				out = append(out, r.what+": "+re)
			}
		}
	}
	return out
}

func TestJSLevelLint(t *testing.T) {
	bad := map[string]string{
		"x ??= 1;":                   "logical assignment",
		"a.b?.c":                     "optional chaining",
		"2 ** 3":                     "exponentiation",
		"async function f() {}":      "async",
		"class A { #x = 1; }":        "private",
		"class A { static { } }":     "static blocks",
		"try { } catch { }":          "catch without",
		"const n = 1_000;":           "numeric separators",
		"const o = { ...p };":        "object spread",
		"import(\"x\")":              "dynamic import",
		"globalThis.x":               "globalThis",
		"Object.entries(o)":          "Object",
		"s.replaceAll(\"a\", \"b\")": "newer method",
		"const r = /(?<y>\\d+)/;":    "lookbehind or named group",
		"const r = /a.b/s;":          "flag s",
		"`a ${ x ?? y ** 2 } b`":     "exponentiation",
		"f(`${`${o?.p}`}`)":          "optional chaining",
	}
	for src, want := range bad {
		got := strings.Join(jsLevelViolations(src), "\n")
		if !strings.Contains(got, want) {
			t.Errorf("%q: got %q, want a %q violation", src, got, want)
		}
	}
	good := []string{
		"const a = b ?? c; const n = 10n; BigInt(3);",
		"import { x } from \"y\"; export { z }; if (import.meta.main) { }",
		"f(...args); function g(...rest) { } f(a, ...at(b));",
		"let globalThis = 1; trace(process, globalThis);",
		"const pow = BigInt(\"1\" + \"0\".repeat(3));",
		"const s = \"a ?? b ** c #x async\"; // ?.x ** 2",
		"/* x ??= 1 */ const t = `a ** b`;",
		"const r = /^[+-]?[0-9]+$/; const q = x / y / z;",
		"s.includes(t); Array.from(s); s.codePointAt(0);",
		"const o = { contains: (s) => s, replaceAll: (s) => s };",
		"x?.5:1",
	}
	for _, src := range good {
		if v := jsLevelViolations(src); len(v) != 0 {
			t.Errorf("%q: unexpected %v", src, v)
		}
	}
}

// TestGeneratedJSLanguageLevel compiles every example and test program and
// checks the modules, runtime included, against the allowed level.
func TestGeneratedJSLanguageLevel(t *testing.T) {
	var files []string
	for _, glob := range []string{"../examples/*.rsc", "testdata/runtime/*.rsc", "testdata/js/*.rsc", "testdata/extern/*.rsc", "testdata/dom/*.rsc"} {
		m, _ := filepath.Glob(glob)
		files = append(files, m...)
	}
	if len(files) == 0 {
		t.Fatal("no programs found")
	}
	for _, path := range files {
		js, diags, err := driver.BuildJS("prog.rsc", read(t, path), jsgen.Options{Version: "test"})
		if diags != nil || err != nil {
			continue // programs that do not compile are tested elsewhere
		}
		for _, v := range jsLevelViolations(string(js)) {
			t.Errorf("%s: %s", filepath.ToSlash(path), v)
		}
	}
	for _, v := range jsLevelViolations(jsgen.Runtime()) {
		t.Errorf("runtime: %s", v)
	}
}
