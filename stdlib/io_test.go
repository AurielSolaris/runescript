package stdlib

import (
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDecodeUTF8MatchesTextDecoder decodes random byte strings, rich in
// broken and overlong UTF-8, here and with the JavaScript target's
// TextDecoder: both must give the same text, so files and standard input
// read the same on both targets.
func TestDecodeUTF8MatchesTextDecoder(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		if os.Getenv("RUNESCRIPT_REQUIRE_NODE") != "" {
			t.Fatal("node not found")
		}
		t.Skip("node not found")
	}
	r := rand.New(rand.NewSource(9))
	interesting := []byte{0x00, 0x41, 0x7F, 0x80, 0xBF, 0xC0, 0xC1, 0xC2, 0xDF, 0xE0, 0xED, 0xEF, 0xF0, 0xF4, 0xF5, 0xFF, 0x9F, 0xA0, 0x8F, 0x90}
	var inputs []string // hex
	for i := 0; i < 3000; i++ {
		b := make([]byte, r.Intn(8))
		for j := range b {
			if r.Intn(3) == 0 {
				b[j] = byte(r.Intn(256))
			} else {
				b[j] = interesting[r.Intn(len(interesting))]
			}
		}
		inputs = append(inputs, hex.EncodeToString(b))
	}
	inputs = append(inputs, hex.EncodeToString(append([]byte{0xEF, 0xBB, 0xBF}, "keeps the BOM"...)))
	dir := t.TempDir()
	in, script := filepath.Join(dir, "in.json"), filepath.Join(dir, "decode.cjs")
	data, _ := json.Marshal(inputs)
	os.WriteFile(in, data, 0o644)
	os.WriteFile(script, []byte(`
const fs = require("node:fs");
const d = new TextDecoder("utf-8", { ignoreBOM: true });
const out = JSON.parse(fs.readFileSync(process.argv[2], "utf8")).map((h) => d.decode(Buffer.from(h, "hex")));
process.stdout.write(JSON.stringify(out));
`), 0o644)
	out, err := exec.Command(node, script, in).Output()
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatal(err)
	}
	for i, h := range inputs {
		b, _ := hex.DecodeString(h)
		if got := decodeUTF8(b); got != want[i] {
			t.Errorf("decodeUTF8(% x) = %q, TextDecoder gives %q", b, got, want[i])
		}
	}
}

func TestErrorAndWarningLines(t *testing.T) {
	for _, color := range []bool{false, true} {
		var stderr strings.Builder
		env := NewEnv(nil, &stderr, nil, nil)
		env.Color = color
		Lookup("io").Funcs["error"].EnvImpl(env, []any{"bad"})
		Lookup("io").Funcs["warn"].EnvImpl(env, []any{"careful"})
		want := "bad\ncareful\n"
		if color {
			want = "\x1b[31mbad\x1b[0m\n\x1b[33mcareful\x1b[0m\n"
		}
		if stderr.String() != want {
			t.Errorf("color %t: got %q, want %q", color, stderr.String(), want)
		}
	}
}
