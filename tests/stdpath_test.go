package tests

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// rsQuote writes s as a RuneScript string literal.
func rsQuote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

// TestStdPathMatches runs random std:path calls, with POSIX and Windows
// rules, in the interpreter and on Node.js: the JavaScript runtime's port
// of stdlib/path.go must give the same results. (stdlib's
// TestPathMatchesNode checks the Go code against Node.js's path module.)
func TestStdPathMatches(t *testing.T) {
	nodePath := node(t)
	r := rand.New(rand.NewSource(3))
	pieces := []string{"a", "b", "file.txt", ".hidden", "x.y.z", "..", ".", "/", `\`, "//", `\\`, "C:", "c:", "D:",
		`\\server\share`, "//srv/vol", ":", "dir.", "é", "日本", " ", "..."}
	random := func() string {
		var b strings.Builder
		for n := r.Intn(6); n >= 0; n-- {
			b.WriteString(pieces[r.Intn(len(pieces))])
		}
		return rsQuote(b.String())
	}
	var b strings.Builder
	b.WriteString("package fuzz;\n\nuse \"std:path\";\n\nprotected int main() {\n")
	for i := 0; i < 1500; i++ {
		flavor := []string{"posix", "win32"}[r.Intn(2)]
		var call string
		switch k := r.Intn(10); k {
		case 0, 1, 2, 3, 4, 5, 6:
			fn := []string{"Normalize", "IsAbsolute", "Dirname", "Basename", "Extname", "Resolve", "Root"}[k]
			call = fmt.Sprintf("path.%s%s(%s)", flavor, fn, random())
		case 7:
			call = fmt.Sprintf("path.%sBasename(%s, %s)", flavor, random(), rsQuote([]string{".txt", "z", "e", "file.txt"}[r.Intn(4)]))
		case 8:
			call = fmt.Sprintf("path.%sJoinAll([%s, %s, %s])", flavor, random(), random(), random())
		default:
			call = fmt.Sprintf("path.%sRelative(%s, %s)", flavor, random(), random())
		}
		fmt.Fprintf(&b, "    trace(%d, %s);\n", i, call)
	}
	b.WriteString("    return 0;\n}\n")
	src := []byte(b.String())
	want := interpret(t, src)
	got := runJS(t, nodePath, src)
	if got != want {
		wl, gl := strings.Split(want.out, "\n"), strings.Split(got.out, "\n")
		for i := range min(len(wl), len(gl)) {
			if wl[i] != gl[i] {
				t.Fatalf("first difference, line %d:\ninterpreter: %s\nnode:        %s", i, wl[i], gl[i])
			}
		}
		t.Fatalf("JavaScript differs from the interpreter: %+v vs %+v", want, got)
	}
}
