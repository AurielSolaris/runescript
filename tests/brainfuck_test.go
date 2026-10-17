package tests

import (
	"path/filepath"
	"testing"
)

// TestBrainfuckExample runs examples/brainfuck.rsc on the programs in
// examples/bf and on broken ones, in the interpreter and on Node.js.
func TestBrainfuckExample(t *testing.T) {
	nodePath := node(t)
	src := read(t, "../examples/brainfuck.rsc")
	bf := func(name string) string {
		p, err := filepath.Abs(filepath.Join("../examples/bf", name))
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	cases := []struct {
		name   string
		args   []string
		stdin  string
		stdout string
		stderr string
		code   int
	}{
		{"built-in", nil, "", "Hello World!\n", "", 0},
		{"hello", []string{bf("hello.bf")}, "", "Hello World!\n", "", 0},
		{"rot13", []string{bf("rot13.bf")}, "Hello, World!\nnaïve café ✓\n", "Uryyb, Jbeyq!\nanïir pnsé ✓\n", "", 0},
		{"cat", []string{bf("cat.bf")}, "cat me\nüñí 日本", "cat me\nüñí 日本", "", 0},
		{"alphabet", []string{bf("alphabet.bf")}, "", "ABCDEFGHIJKLMNOPQRSTUVWXYZ\n", "", 0},
		{"inline", []string{"-e", "++++++++[>++++++++<-]>+.+.+."}, "", "ABC", "", 0},
		{"all 32768 cells", []string{"-e", "+[>+]"}, "", "",
			"-e:1:3: the data pointer moved off the tape (to cell 32768; the cells are 0 to 32767)\n", 1},
		{"before the first cell", []string{"-e", "+<"}, "", "",
			"-e:1:2: the data pointer moved off the tape (to cell -1; the cells are 0 to 32767)\n", 1},
		{"cells wrap", []string{"-e", "-.+."}, "", "�\x00", "", 0},
		{"input end leaves the cell", []string{"-e", "+++++++++++++++++++++++++++++++++++++++++++++++++,."}, "", "1", "", 0},
		{"unmatched ]", []string{"-e", "+]"}, "", "", "-e:1:2: this ']' has no matching '['\n", 1},
		{"unmatched [", []string{"-e", "+\n [ [-]"}, "", "", "-e:2:2: this '[' has no matching ']'\n", 1},
		{"missing file", []string{"nope.bf"}, "", "", "fs.readFile: cannot read nope.bf: no such file or directory\n", 1},
		{"usage", []string{"a", "b"}, "", "", "usage: brainfuck.rsc [<file.bf> | -e <code>]\n", 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			want := envRun{c.stdout, c.stderr, c.code}
			if got := runInterpreterIn(t, src, c.args, c.stdin); got != want {
				t.Errorf("interpreter: %+v\nwant         %+v", got, want)
			}
			if got := runModuleIn(t, nodePath, src, c.args, c.stdin); got != want {
				t.Errorf("node: %+v\nwant  %+v", got, want)
			}
		})
	}
}
