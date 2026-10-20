package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// project is a generated compiler workload, written in both languages.
type project struct {
	ID, Name, Title string
	gen             func(n int) (rs, ts map[string]string) // file name -> source; the entry is main.*
	n               int
}

// TypeScript 1.0 has no `process` typings; programs declare it.
const tsPrelude = "declare var process: any;\n"

func projects(large bool) []project {
	files, funcs, depth := 100, 2000, 300
	ps := []project{
		{ID: "P01", Name: "empty", Title: "Empty program", gen: genEmpty},
		{ID: "P02", Name: "small", Title: "Small program (A03 sieve)", gen: nil},
		{ID: "P03", Name: "files", Title: "Many small files", gen: genFiles, n: files},
		{ID: "P05", Name: "large_file", Title: "One large generated file", gen: genLargeFile, n: funcs},
		{ID: "P06", Name: "deep_nesting", Title: "Deeply nested expression", gen: genDeep, n: depth},
		{ID: "P07", Name: "declarations", Title: "Many local declarations", gen: genDecls, n: funcs},
	}
	if large {
		ps = append(ps,
			project{ID: "P04", Name: "files_1000", Title: "1,000 small files", gen: genFiles, n: 1000},
			project{ID: "P15", Name: "very_large_file", Title: "Very large generated file", gen: genLargeFile, n: 20000},
		)
	}
	return ps
}

func genEmpty(int) (map[string]string, map[string]string) {
	return map[string]string{"main.rsc": "package main;\n\nprotected int main() {\n    return 0;\n}\n"},
		map[string]string{"main.ts": tsPrelude + "process.exitCode = 0;\n"}
}

func genFiles(n int) (map[string]string, map[string]string) {
	rs, ts := map[string]string{}, map[string]string{}
	var rsMain, tsMain strings.Builder
	rsMain.WriteString("package main;\n\n")
	tsMain.WriteString(tsPrelude)
	for k := 0; k < n; k++ {
		rs[fmt.Sprintf("m%d.rsc", k)] = fmt.Sprintf(`package m%d;

public int step(int x) {
    if (x %% 2 == 0) {
        return x / 2;
    }
    return x * 3 + %d;
}

public int run(int x, int times) {
    int v = x;
    for (int i = 0; i < times; i++) {
        v = step(v) %% 1000;
    }
    return v;
}
`, k, k%7+1)
		ts[fmt.Sprintf("m%d.ts", k)] = fmt.Sprintf(`export function step(x: number): number {
    if (x %% 2 == 0) {
        return Math.floor(x / 2);
    }
    return x * 3 + %d;
}

export function run(x: number, times: number): number {
    var v = x;
    for (var i = 0; i < times; i++) {
        v = step(v) %% 1000;
    }
    return v;
}
`, k%7+1)
		fmt.Fprintf(&rsMain, "use \"./m%d.rsc\";\n", k)
		fmt.Fprintf(&tsMain, "import m%d = require(\"./m%d\");\n", k, k)
	}
	rsMain.WriteString("\nprotected int main() {\n    int total = 0;\n")
	tsMain.WriteString("\nvar total = 0;\n")
	for k := 0; k < n; k++ {
		fmt.Fprintf(&rsMain, "    total = (total + m%d.run(%d, 3)) %% 1000003;\n", k, k)
		fmt.Fprintf(&tsMain, "total = (total + m%d.run(%d, 3)) %% 1000003;\n", k, k)
	}
	rsMain.WriteString("    trace(total);\n    return 0;\n}\n")
	tsMain.WriteString("process.stdout.write(total + \"\\n\");\n")
	rs["main.rsc"], ts["main.ts"] = rsMain.String(), tsMain.String()
	return rs, ts
}

func genLargeFile(n int) (map[string]string, map[string]string) {
	var r, t strings.Builder
	r.WriteString("package main;\n\n")
	t.WriteString(tsPrelude)
	for k := 0; k < n; k++ {
		fmt.Fprintf(&r, `public int f%d(int a, int b) {
    int s = a * %d + b;
    if (s > 1000) {
        s = s %% 1000;
    } else {
        s = s + %d;
    }
    for (int i = 0; i < 2; i++) {
        s = (s * 7 + i) %% 10007;
    }
    return s;
}

`, k, k%13+1, k%5)
		fmt.Fprintf(&t, `function f%d(a: number, b: number): number {
    var s = a * %d + b;
    if (s > 1000) {
        s = s %% 1000;
    } else {
        s = s + %d;
    }
    for (var i = 0; i < 2; i++) {
        s = (s * 7 + i) %% 10007;
    }
    return s;
}

`, k, k%13+1, k%5)
	}
	r.WriteString("protected int main() {\n    int total = 0;\n")
	t.WriteString("var total = 0;\n")
	for k := 0; k < n; k += 97 {
		fmt.Fprintf(&r, "    total = (total + f%d(%d, total)) %% 1000003;\n", k, k)
		fmt.Fprintf(&t, "total = (total + f%d(%d, total)) %% 1000003;\n", k, k)
	}
	r.WriteString("    trace(total);\n    return 0;\n}\n")
	t.WriteString("process.stdout.write(total + \"\\n\");\n")
	return map[string]string{"main.rsc": r.String()}, map[string]string{"main.ts": t.String()}
}

func genDeep(depth int) (map[string]string, map[string]string) {
	rng := rand.New(rand.NewSource(1))
	expr := "x"
	for i := 0; i < depth; i++ {
		op := []string{"+", "-", "*"}[rng.Intn(3)]
		if op == "*" {
			expr = fmt.Sprintf("((%s) %% 1000 * %d)", expr, rng.Intn(9)+1)
		} else {
			expr = fmt.Sprintf("(%s %s %d)", expr, op, rng.Intn(100))
		}
	}
	return map[string]string{"main.rsc": "package main;\n\nprotected int main() {\n    int x = 3;\n    int y = " + expr + ";\n    trace(y);\n    return 0;\n}\n"},
		map[string]string{"main.ts": tsPrelude + "var x = 3;\nvar y: number = " + expr + ";\nprocess.stdout.write(y + \"\\n\");\n"}
}

func genDecls(n int) (map[string]string, map[string]string) {
	var r, t strings.Builder
	r.WriteString("package main;\n\nprotected int main() {\n    int v0 = 1;\n")
	t.WriteString(tsPrelude + "function main(): number {\n    var v0 = 1;\n")
	for k := 1; k < n; k++ {
		fmt.Fprintf(&r, "    int v%d = (v%d * 3 + %d) %% 10007;\n", k, k-1, k)
		fmt.Fprintf(&t, "    var v%d: number = (v%d * 3 + %d) %% 10007;\n", k, k-1, k)
	}
	fmt.Fprintf(&r, "    trace(v%d);\n    return 0;\n}\n", n-1)
	fmt.Fprintf(&t, "    process.stdout.write(v%d + \"\\n\");\n    return 0;\n}\nmain();\n", n-1)
	return map[string]string{"main.rsc": r.String()}, map[string]string{"main.ts": t.String()}
}

var compileHeader = []string{
	"project", "id", "task", "compiler_version", "files", "source_bytes",
	"warmup_runs", "measured_runs", "median_ms", "min_ms", "max_ms", "stddev_ms",
	"peak_memory_mb", "ok", "os", "arch",
}

func cmdCompile(fs *flag.FlagSet, args []string, w io.Writer) error {
	warmup := fs.Int("warmup", 1, "warm-up runs per task")
	runs := fs.Int("runs", 5, "measured runs per task")
	large := fs.Bool("large", false, "also run the large projects (1,000 files, 20,000 functions)")
	outDir := fs.String("out", "", "directory for the CSV and Markdown results (default bench/results)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	tc, err := setup(true)
	if err != nil {
		return err
	}
	defer tc.Close()
	tc.warnMissingTS(func(f string, a ...any) { fmt.Fprintf(os.Stderr, f, a...) })

	type task struct {
		p      project
		name   string
		argv   []string
		dir    string
		files  int
		bytes  int64
		sample []Sample
		err    error
	}
	var tasks []*task
	for _, p := range projects(*large) {
		var rs, ts map[string]string
		if p.gen == nil { // P02: a real program from the suite
			rs = map[string]string{"main.rsc": readOr(filepath.Join(tc.BenchRoot, "suite", "a03_sieve", "main.rsc"))}
			ts = map[string]string{"main.ts": readOr(filepath.Join(tc.BenchRoot, "suite", "a03_sieve", "main.ts"))}
		} else {
			rs, ts = p.gen(p.n)
		}
		rsDir := filepath.Join(tc.Work, "compile", p.ID, "rs")
		tsDir := filepath.Join(tc.Work, "compile", p.ID, "ts")
		rsBytes, err := writeAll(rsDir, rs)
		if err != nil {
			return err
		}
		tsBytes, err := writeAll(tsDir, ts)
		if err != nil {
			return err
		}
		tasks = append(tasks,
			&task{p: p, name: "rs-check", dir: rsDir, files: len(rs), bytes: rsBytes,
				argv: []string{tc.RuneScript, "check", "--color=never", "main.rsc"}},
			&task{p: p, name: "rs-build", dir: rsDir, files: len(rs), bytes: rsBytes,
				argv: []string{tc.RuneScript, "build", "--color=never", "-o", "out/main.mjs", "main.rsc"}})
		if tc.TSC != "" {
			argv := append([]string{tc.Node, tc.TSC}, TSFlags...)
			argv = append(argv, "--outDir", "out", "main.ts")
			tasks = append(tasks, &task{p: p, name: "tsc", dir: tsDir, files: len(ts), bytes: tsBytes, argv: argv})
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	total := *warmup + *runs
	for round := 0; round < total; round++ {
		fmt.Fprintf(os.Stderr, "\rround %d/%d", round+1, total)
		for _, k := range rng.Perm(len(tasks)) {
			t := tasks[k]
			if t.err != nil {
				continue
			}
			// Every build is a clean build into an empty out directory.
			os.RemoveAll(filepath.Join(t.dir, "out"))
			if err := os.MkdirAll(filepath.Join(t.dir, "out"), 0o755); err != nil {
				return err
			}
			s, err := measure(t.argv, t.dir)
			if err == nil && s.Code != 0 {
				err = fmt.Errorf("exit %d: %s", s.Code, firstLine(s.Stderr+s.Stdout))
			}
			if err != nil {
				t.err = err
				fmt.Fprintf(os.Stderr, "\n%s/%s failed: %v\n", t.p.ID, t.name, err)
				continue
			}
			if round >= *warmup {
				t.sample = append(t.sample, s)
			}
		}
	}
	fmt.Fprintln(os.Stderr)

	var rows [][]string
	var sb strings.Builder
	fmt.Fprintf(&sb, "## Compiler (clean builds)\n\n%s vs TypeScript %s (`tsc %s`); node %s; %d warm-up + %d measured runs, randomized order.\n\n",
		tc.RSVersion, orNA(tc.TSVersion), strings.Join(TSFlags, " "), tc.NodeVersion, *warmup, *runs)
	sb.WriteString("TypeScript 1.0 has no check-only mode, so `tsc` always type-checks and emits; compare it with `rs-build` (check + IR + JavaScript). Times are median wall-clock milliseconds, process startup included.\n\n")
	sb.WriteString("| Project | Files | rs-check | rs-build | tsc | tsc / rs-build |\n|---|---:|---:|---:|---:|---:|\n")
	cur := map[string]map[string]Stats{}
	var order []project
	for _, t := range tasks {
		st := stats(t.sample)
		ok := t.err == nil
		rows = append(rows, []string{t.p.Name, t.p.ID, t.name, compilerVersion(tc, strings.Replace(t.name, "tsc", "ts", 1)),
			fmt.Sprint(t.files), fmt.Sprint(t.bytes), fmt.Sprint(*warmup), fmt.Sprint(st.N),
			f2(st.Median), f2(st.Min), f2(st.Max), f2(st.SD), mb(st.PeakMB), fmt.Sprint(ok), goos(), goarch()})
		if cur[t.p.ID] == nil {
			cur[t.p.ID] = map[string]Stats{}
			order = append(order, t.p)
		}
		if ok {
			cur[t.p.ID][t.name] = st
		}
	}
	for _, p := range order {
		m := cur[p.ID]
		cell := func(n string) string {
			if s, ok := m[n]; ok && s.N > 0 {
				return f1(s.Median)
			}
			return "n/a"
		}
		r := ""
		if a, b := m["tsc"], m["rs-build"]; a.N > 0 && b.N > 0 {
			r = fmt.Sprintf("%.2f", a.Median/b.Median)
		}
		files := 1
		if p.Name == "files" || p.Name == "files_1000" {
			files = p.n + 1
		}
		fmt.Fprintf(&sb, "| %s %s | %d | %s | %s | %s | %s |\n", p.ID, p.Title, files, cell("rs-check"), cell("rs-build"), cell("tsc"), r)
	}
	stamp := time.Now().Format("20060102-150405")
	csvPath, err := writeCSV(tc, *outDir, "compile-"+stamp+".csv", compileHeader, rows)
	if err != nil {
		return err
	}
	mdPath, err := writeFile(tc, *outDir, "compile-"+stamp+".md", sb.String())
	if err != nil {
		return err
	}
	fmt.Fprint(w, sb.String())
	fmt.Fprintf(w, "\nwrote %s and %s\n", csvPath, mdPath)
	return nil
}

func writeAll(dir string, files map[string]string) (int64, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}
	var n int64
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			return 0, err
		}
		n += int64(len(src))
	}
	return n, nil
}

func readOr(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
