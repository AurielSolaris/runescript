package stdlib

import (
	"encoding/json"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// pathCase is one call of a path function, with the result Node.js gives.
type pathCase struct {
	Flavor string   `json:"flavor"` // "posix" or "win32"
	Fn     string   `json:"fn"`
	Args   []string `json:"args"`
	Want   any      `json:"want"`
}

// nodeScript computes each case's result with Node.js's path module, with
// process.cwd() replaced by the fake working directory.
const nodeScript = `
const path = require("node:path");
const fs = require("node:fs");
const [input, cwd] = process.argv.slice(2);
process.cwd = () => cwd;
const cases = JSON.parse(fs.readFileSync(input, "utf8"));
for (const c of cases) {
  const p = path[c.flavor];
  c.want = c.fn === "join" || c.fn === "resolve" ? p[c.fn](...c.args) : p[c.fn](...c.args);
}
process.stdout.write(JSON.stringify(cases));
`

// fakeCwd is the working directory both sides see, in the host's form.
func fakeCwd() string {
	if runtime.GOOS == "windows" {
		return `C:\work\dir`
	}
	return "/work/dir"
}

// randomPath builds a path from pieces that exercise the rules: both
// separators, ".", "..", drives, UNC roots, and extensions.
func randomPath(r *rand.Rand) string {
	pieces := []string{"a", "b", "file.txt", ".hidden", "x.y.z", "..", ".", "/", `\`, "//", `\\`, "C:", "c:", "D:",
		`\\server\share`, "//srv/vol", ":", "dir.", "é", "日本", " ", "..."}
	var b strings.Builder
	for n := r.Intn(6); n >= 0; n-- {
		b.WriteString(pieces[r.Intn(len(pieces))])
	}
	return b.String()
}

func TestPathMatchesNode(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		if os.Getenv("RUNESCRIPT_REQUIRE_NODE") != "" {
			t.Fatal("node not found")
		}
		t.Skip("node not found")
	}
	r := rand.New(rand.NewSource(5))
	var cases []*pathCase
	for i := 0; i < 6000; i++ {
		flavor := []string{"posix", "win32"}[r.Intn(2)]
		var c pathCase
		switch k := r.Intn(11); k {
		case 0, 1, 2, 3, 4, 5:
			fn := []string{"normalize", "isAbsolute", "dirname", "basename", "extname", "resolve"}[k]
			c = pathCase{Fn: fn, Args: []string{randomPath(r)}}
		case 6:
			c = pathCase{Fn: "basename", Args: []string{randomPath(r), []string{".txt", "z", "e", "a", "file.txt", ""}[r.Intn(6)]}}
		case 7, 8:
			args := make([]string, 1+r.Intn(4))
			for j := range args {
				args[j] = randomPath(r)
			}
			c = pathCase{Fn: []string{"join", "resolve"}[k-7], Args: args}
		case 9:
			c = pathCase{Fn: "relative", Args: []string{randomPath(r), randomPath(r)}}
		default:
			c = pathCase{Fn: "parse", Args: []string{randomPath(r)}}
		}
		c.Flavor = flavor
		cases = append(cases, &c)
	}
	dir := t.TempDir()
	input, script := filepath.Join(dir, "cases.json"), filepath.Join(dir, "paths.cjs")
	data, _ := json.Marshal(cases)
	os.WriteFile(input, data, 0o644)
	os.WriteFile(script, []byte(nodeScript), 0o644)
	out, err := exec.Command(node, script, input, fakeCwd()).Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	if err := json.Unmarshal(out, &cases); err != nil {
		t.Fatal(err)
	}

	env := &Env{Windows: runtime.GOOS == "windows", Getwd: func() (string, error) { return fakeCwd(), nil }}
	fails := 0
	for _, c := range cases {
		f := posixPath
		if c.Flavor == "win32" {
			f = win32Path
		}
		var got any
		switch c.Fn {
		case "normalize":
			got = f.normalize(c.Args[0])
		case "isAbsolute":
			got = f.isAbsolute(c.Args[0])
		case "dirname":
			got = f.dirname(c.Args[0])
		case "basename":
			suffix := ""
			if len(c.Args) > 1 {
				suffix = c.Args[1]
			}
			got = f.basename(c.Args[0], suffix)
		case "extname":
			got = f.extname(c.Args[0])
		case "join":
			got = f.join(c.Args)
		case "resolve":
			got = f.resolve(env, c.Args)
		case "relative":
			got = f.relative(env, c.Args[0], c.Args[1])
		case "parse":
			got = f.root(c.Args[0])
			c.Want = c.Want.(map[string]any)["root"]
		}
		if got != c.Want {
			fails++
			if fails <= 25 {
				t.Errorf("path.%s.%s(%q) = %q, Node.js gives %q", c.Flavor, c.Fn, c.Args, got, c.Want)
			}
		}
	}
	if fails > 25 {
		t.Errorf("... %d mismatches in all", fails)
	}
}
