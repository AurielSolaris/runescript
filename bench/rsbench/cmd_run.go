package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

// baselineKey is the empty program: its time is each implementation's
// fixed startup cost, subtracted to give net times.
const baselineKey = "s01_empty"

type job struct {
	b       *Bench
	p       *Program
	size    string
	samples []Sample
	first   *Sample // the first run, used for the correctness check
	err     error
}

func cmdRun(fs *flag.FlagSet, args []string, w io.Writer) error {
	size := fs.String("size", "small", "input size: test, small, medium, or large")
	warmup := fs.Int("warmup", 2, "warm-up runs per implementation (not measured)")
	runs := fs.Int("runs", 10, "measured runs per implementation")
	implFlag := fs.String("impl", "", "comma-separated implementations to run (default: all)")
	seed := fs.Int64("seed", 0, "seed for the run order (0: from the clock)")
	outDir := fs.String("out", "", "directory for the CSV and Markdown results (default bench/results)")
	benches, _, err := loadSelected(fs, args, "runtime")
	if err != nil {
		return err
	}
	if *runs < 1 {
		return fmt.Errorf("-runs must be at least 1")
	}
	tc, err := setup(true)
	if err != nil {
		return err
	}
	defer tc.Close()
	tc.warnMissingTS(func(f string, a ...any) { fmt.Fprintf(os.Stderr, f, a...) })

	// Always measure the empty program, for the startup baseline.
	if !containsKey(benches, baselineKey) {
		root, _ := findRoot()
		suite, _ := loadSuite(root + "/bench")
		if base, _ := selectBenches(suite, "", []string{baselineKey}); len(base) == 1 {
			benches = append([]*Bench{base[0]}, benches...)
		}
	}
	wanted := map[string]bool{}
	for _, i := range strings.Split(*implFlag, ",") {
		if i = strings.TrimSpace(i); i != "" {
			wanted[i] = true
		}
	}

	var jobs []*job
	for _, b := range benches {
		sz := *size
		if _, ok := b.Sizes[sz]; !ok {
			if b.Key() != baselineKey {
				fmt.Fprintf(os.Stderr, "skipping %s: no %q size\n", b.Key(), sz)
				continue
			}
			sz = "test"
		}
		fmt.Fprintf(os.Stderr, "preparing %s\n", b.Key())
		for _, p := range tc.prepare(b) {
			if len(wanted) > 0 && !wanted[p.Impl] {
				continue
			}
			j := &job{b: b, p: p, size: sz, err: p.Err}
			if p.Err != nil {
				fmt.Fprintf(os.Stderr, "  %s unavailable: %s\n", p.Impl, firstLine(p.Err.Error()))
			}
			jobs = append(jobs, j)
		}
	}

	// Rounds: each round runs every job once, in a new random order, so
	// that background load and thermal drift spread over all of them.
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(*seed))
	total := *warmup + *runs
	for round := 0; round < total; round++ {
		fmt.Fprintf(os.Stderr, "\rround %d/%d", round+1, total)
		order := rng.Perm(len(jobs))
		for _, k := range order {
			j := jobs[k]
			if j.err != nil {
				continue
			}
			s, err := measure(append(j.p.Argv, j.b.Sizes[j.size]), j.p.Dir)
			if err == nil && s.Code != 0 {
				err = fmt.Errorf("exit %d: %s", s.Code, firstLine(s.Stderr))
			}
			if err != nil {
				j.err = err
				fmt.Fprintf(os.Stderr, "\n%s/%s failed: %v\n", j.b.Key(), j.p.Impl, err)
				continue
			}
			if j.first == nil {
				j.first = &s
			}
			if round >= *warmup {
				j.samples = append(j.samples, s)
			}
		}
	}
	fmt.Fprintln(os.Stderr)

	rows := runtimeRows(tc, jobs, *warmup)
	stamp := time.Now().Format("20060102-150405")
	csvPath, err := writeCSV(tc, *outDir, "runtime-"+stamp+".csv", runtimeHeader, rows)
	if err != nil {
		return err
	}
	md := runtimeMarkdown(tc, jobs, *size, *warmup, *runs, *seed)
	mdPath, err := writeFile(tc, *outDir, "runtime-"+stamp+".md", md)
	if err != nil {
		return err
	}
	fmt.Fprint(w, md)
	fmt.Fprintf(w, "\nwrote %s and %s\n", csvPath, mdPath)
	return nil
}

func containsKey(bs []*Bench, key string) bool {
	for _, b := range bs {
		if b.Key() == key {
			return true
		}
	}
	return false
}

// Stats summarizes a job's measured runs, in milliseconds.
type Stats struct {
	N                                int
	Median, Min, Max, Mean, SD, CI95 float64
	PeakMB                           float64 // -1 when unknown
}

func stats(samples []Sample) Stats {
	st := Stats{N: len(samples), PeakMB: -1}
	if st.N == 0 {
		return st
	}
	ms := make([]float64, st.N)
	var peak int64 = -1
	for i, s := range samples {
		ms[i] = float64(s.Wall) / float64(time.Millisecond)
		if s.PeakRSS > peak {
			peak = s.PeakRSS
		}
	}
	sort.Float64s(ms)
	st.Min, st.Max = ms[0], ms[st.N-1]
	if st.N%2 == 1 {
		st.Median = ms[st.N/2]
	} else {
		st.Median = (ms[st.N/2-1] + ms[st.N/2]) / 2
	}
	for _, v := range ms {
		st.Mean += v
	}
	st.Mean /= float64(st.N)
	if st.N > 1 {
		var ss float64
		for _, v := range ms {
			ss += (v - st.Mean) * (v - st.Mean)
		}
		st.SD = math.Sqrt(ss / float64(st.N-1))
		st.CI95 = 1.96 * st.SD / math.Sqrt(float64(st.N))
	}
	if peak >= 0 {
		st.PeakMB = float64(peak) / (1 << 20)
	}
	return st
}

// correctness compares a job's first output with expected/<size>.out,
// or, without one, with the interpreter's output for the same size.
func correctness(j *job, jobs []*job) string {
	if j.first == nil {
		return "fail"
	}
	if want, ok := j.b.Expected(j.size, j.p.Impl); ok {
		return passFail(j.first.Stdout == want)
	}
	for _, o := range jobs {
		if o.b == j.b && o.p.Impl == "rs-vm" && o.first != nil {
			if strings.HasPrefix(j.p.Impl, "ts") && j.b.Notes != "" && j.first.Stdout != o.first.Stdout {
				return "differs (see notes)"
			}
			return passFail(j.first.Stdout == o.first.Stdout)
		}
	}
	return "unchecked"
}

func passFail(ok bool) string {
	if ok {
		return "pass"
	}
	return "fail"
}

var runtimeHeader = []string{
	"benchmark", "id", "implementation", "runtime_version", "compiler_version",
	"input_size", "size_arg", "warmup_runs", "measured_runs",
	"median_ms", "min_ms", "max_ms", "mean_ms", "stddev_ms", "ci95_ms", "net_median_ms",
	"peak_memory_mb", "correctness", "os", "arch",
}

func compilerVersion(tc *Toolchain, impl string) string {
	if strings.HasPrefix(impl, "ts") {
		return "typescript " + tc.TSVersion
	}
	return tc.RSVersion
}

func runtimeVersion(tc *Toolchain, impl string) string {
	if impl == "rs-vm" {
		return tc.RSVersion + " (" + runtime.Version() + ")"
	}
	return "node " + tc.NodeVersion
}

func baselines(jobs []*job) map[string]float64 {
	base := map[string]float64{}
	for _, j := range jobs {
		if j.b.Key() == baselineKey && len(j.samples) > 0 {
			base[j.p.Impl] = stats(j.samples).Median
		}
	}
	return base
}

func runtimeRows(tc *Toolchain, jobs []*job, warmup int) [][]string {
	base := baselines(jobs)
	var rows [][]string
	for _, j := range jobs {
		st := stats(j.samples)
		net := ""
		if b, ok := base[baseImpl(j.p.Impl)]; ok && st.N > 0 && j.b.Key() != baselineKey {
			net = f2(st.Median - b)
		}
		corr := correctness(j, jobs)
		if j.err != nil {
			corr = "error: " + firstLine(j.err.Error())
		}
		rows = append(rows, []string{
			j.b.Key(), j.b.ID, j.p.Impl, runtimeVersion(tc, j.p.Impl), compilerVersion(tc, j.p.Impl),
			j.size, j.b.Sizes[j.size], fmt.Sprint(warmup), fmt.Sprint(st.N),
			f2(st.Median), f2(st.Min), f2(st.Max), f2(st.Mean), f2(st.SD), f2(st.CI95), net,
			mb(st.PeakMB), corr, runtime.GOOS, runtime.GOARCH,
		})
	}
	return rows
}

func runtimeMarkdown(tc *Toolchain, jobs []*job, size string, warmup, runs int, seed int64) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## Runtime (%s inputs)\n\n", size)
	fmt.Fprintf(&sb, "%s; node %s; TypeScript %s; %s/%s; %d warm-up + %d measured runs per cell, randomized order (seed %d).\n\n",
		tc.RSVersion, tc.NodeVersion, orNA(tc.TSVersion), runtime.GOOS, runtime.GOARCH, warmup, runs, seed)
	sb.WriteString("Times are median wall-clock milliseconds per process, startup included. *Net* subtracts the same implementation's empty-program median. ")
	sb.WriteString("R = TS net / RS-JS net (above 1: RuneScript's JavaScript was faster). S = RS-VM net / RS-JS net (interpreter slowdown).\n\n")

	impls := []string{"rs-vm", "rs-js", "ts", "ts-checked"}
	sb.WriteString("| Benchmark | RS-VM | RS-JS | TS | TS-checked | R | S | Correct |\n|---|---:|---:|---:|---:|---:|---:|---|\n")
	base := baselines(jobs)
	byBench := map[*Bench]map[string]*job{}
	var order []*Bench
	for _, j := range jobs {
		if byBench[j.b] == nil {
			byBench[j.b] = map[string]*job{}
			order = append(order, j.b)
		}
		byBench[j.b][j.p.Impl] = j
	}
	for _, b := range order {
		m := byBench[b]
		net := map[string]float64{}
		cells := []string{fmt.Sprintf("%s %s", b.ID, b.Name)}
		allOK := true
		for _, impl := range impls {
			j := m[impl]
			if j == nil {
				cells = append(cells, "")
				continue
			}
			st := stats(j.samples)
			if st.N == 0 {
				cells = append(cells, "n/a")
				allOK = false
				continue
			}
			cell := f1(st.Median)
			if bm, ok := base[baseImpl(impl)]; ok && b.Key() != baselineKey {
				net[impl] = st.Median - bm
				cell += fmt.Sprintf(" (net %s)", f1(net[impl]))
			}
			cells = append(cells, cell)
			if c := correctness(j, jobs); c != "pass" && c != "differs (see notes)" {
				allOK = false
			}
		}
		cells = append(cells, ratio(net["ts"], net["rs-js"]), ratio(net["rs-vm"], net["rs-js"]))
		if allOK {
			cells = append(cells, "yes")
		} else {
			cells = append(cells, "**no**")
		}
		sb.WriteString("| " + strings.Join(cells, " | ") + " |\n")
	}
	return sb.String()
}

// ratio is a/b, or empty when either time is too small to mean anything
// (within a millisecond of the startup baseline).
func ratio(a, b float64) string {
	if a < 1 || b < 1 {
		return ""
	}
	return fmt.Sprintf("%.2f", a/b)
}

func f1(v float64) string { return fmt.Sprintf("%.1f", v) }
func f2(v float64) string { return fmt.Sprintf("%.2f", v) }

func mb(v float64) string {
	if v < 0 {
		return ""
	}
	return fmt.Sprintf("%.1f", v)
}

func orNA(s string) string {
	if s == "" {
		return "not installed"
	}
	return s
}

// baseImpl is the implementation whose empty-program time is impl's
// startup baseline: TypeScript variants (ts-checked, ...) share ts's.
func baseImpl(impl string) string {
	if strings.HasPrefix(impl, "ts-") {
		return "ts"
	}
	return impl
}
