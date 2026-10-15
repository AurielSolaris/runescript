package tests

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/AurielSolaris/runescript/internal/driver"
)

// FuzzCheck feeds arbitrary source text to the parser and checker. Any input
// must produce a result or diagnostics, never a panic. Seeds are the
// examples and test programs; `go test` runs only the seeds, and
// `go test -fuzz=FuzzCheck ./tests` explores further.
func FuzzCheck(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, src string) {
		driver.Check([]byte(src))
	})
}

// FuzzRun also executes programs that check successfully. Runtime errors
// are fine; panics are not. Sources containing loops are skipped because an
// accepted infinite loop would hang the fuzzer; unbounded recursion is
// stopped by the interpreter's call-depth limit.
func FuzzRun(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, src string) {
		if containsLoop(src) {
			return // an accepted infinite loop would hang the fuzzer
		}
		driver.Run([]byte(src), io.Discard)
	})
}

func addSeeds(f *testing.F) {
	for _, glob := range []string{"../examples/*.rsc", "testdata/errors/*.rsc"} {
		files, _ := filepath.Glob(glob)
		for _, path := range files {
			b, err := os.ReadFile(path)
			if err != nil {
				f.Fatal(err)
			}
			f.Add(string(b))
		}
	}
	f.Add("")
	f.Add("package p;")
	f.Add("package p;\npublic void main() { ((int a) => a)(1)(2); }")
	f.Add("package p;\npublic int(int)(int main() {")
}

func containsLoop(src string) bool {
	for i := 0; i+3 <= len(src); i++ {
		switch {
		case i+5 <= len(src) && src[i:i+5] == "while":
			return true
		case src[i:i+3] == "for":
			return true
		}
	}
	return false
}
