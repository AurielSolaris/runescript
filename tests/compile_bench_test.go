package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/ir"
	"github.com/AurielSolaris/runescript/internal/jsgen"
	"github.com/AurielSolaris/runescript/internal/loader"
)

// BenchmarkCompilerStages measures the front end and JavaScript generator
// in-process. LoadParse includes filesystem access and import resolution.
// EmitJS includes Verify because the generator verifies its input.
func BenchmarkCompilerStages(b *testing.B) {
	for _, workload := range []string{"MostlyDead", "AllLive", "ManyFiles"} {
		b.Run(workload, func(b *testing.B) {
			dir := b.TempDir()
			entry := writeCompilerWorkload(b, dir, workload)
			load := func(b *testing.B) *loader.Program {
				p, ds, err := loader.Load(entry, dir)
				if err != nil || len(ds) != 0 {
					b.Fatalf("load: %v %v", err, ds)
				}
				return p
			}
			p := load(b)
			check := func(b *testing.B) *checker.Info {
				info, ds := checker.Check(p)
				if len(ds) != 0 {
					b.Fatalf("check: %v", ds)
				}
				return info
			}
			info := check(b)
			program := ir.LowerWithoutStatementSpans(info)
			mapped := ir.Lower(info)
			measure := func(stage string, run func(*testing.B)) {
				b.Run(stage, func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						run(b)
					}
				})
			}
			measure("LoadParse", func(b *testing.B) { load(b) })
			measure("Check", func(b *testing.B) { check(b) })
			measure("Lower", func(*testing.B) { ir.LowerWithoutStatementSpans(info) })
			measure("LowerWithMap", func(*testing.B) { ir.Lower(info) })
			measure("Verify", func(b *testing.B) {
				if err := ir.Verify(program); err != nil {
					b.Fatal(err)
				}
			})
			emit := func(stage string, p *ir.Program, opts jsgen.Options) {
				measure(stage, func(b *testing.B) {
					if _, ds, err := jsgen.GenerateMap(p, opts); err != nil || len(ds) != 0 {
						b.Fatalf("emit: %v %v", err, ds)
					}
				})
			}
			emit("EmitJS", program, jsgen.Options{})
			emit("EmitJSWithMap", mapped, jsgen.Options{SourceMap: true, MapURL: "main.js.map"})
		})
	}
}

func writeCompilerWorkload(b *testing.B, dir, workload string) string {
	b.Helper()
	write := func(name, source string) {
		b.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			b.Fatal(err)
		}
	}
	const body = "int s = a * 7 + b; if (s > 1000) { s = s % 1000; } else { s = s + 5; } " +
		"for (int i = 0; i < 2; i++) { s = (s * 7 + i) % 10007; } return s;"
	var source, calls strings.Builder
	source.WriteString("package main;\n")
	calls.WriteString("protected int main() {\n    int sum = 0;\n")
	if workload == "ManyFiles" {
		for i := range 100 {
			name := fmt.Sprintf("m%d", i)
			write(name+".rsc", "package "+name+";\npublic int step(int a, int b) { "+body+" }\n")
			fmt.Fprintf(&source, "use \"./%s.rsc\";\n", name)
			fmt.Fprintf(&calls, "    sum = %s.step(sum, %d);\n", name, i)
		}
	} else {
		for i := range 2000 {
			fmt.Fprintf(&source, "public int f%d(int a, int b) { %s }\n", i, body)
			if workload == "AllLive" || i%97 == 0 {
				fmt.Fprintf(&calls, "    sum = f%d(sum, %d);\n", i, i)
			}
		}
	}
	calls.WriteString("    return sum;\n}\n")
	source.WriteString(calls.String())
	write("main.rsc", source.String())
	return filepath.Join(dir, "main.rsc")
}
