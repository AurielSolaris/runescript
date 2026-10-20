package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// diagCase is a pair of invalid programs with the same mistake, one in
// each language, under bench/diagnostics/<key>.
type diagCase struct {
	Dir    string `json:"-"`
	ID     string `json:"id"`
	Title  string `json:"title"`
	RSLine int    `json:"rs_line"` // the line the primary error should point at
	TSLine int    `json:"ts_line"`
}

var (
	rsDiag = regexp.MustCompile(`:(\d+):(\d+): error (\S+): (.*)$`)
	tsDiag = regexp.MustCompile(`\((\d+),(\d+)\): error (TS\d+): (.*)$`)
)

type diagResult struct {
	errors    int
	line      int
	code, msg string
	full      string
	ms        float64
}

func parseDiags(out string, re *regexp.Regexp) diagResult {
	var r diagResult
	for _, l := range strings.Split(normalize(out), "\n") {
		m := re.FindStringSubmatch(strings.TrimRight(l, " "))
		if m == nil {
			continue
		}
		r.errors++
		if r.errors == 1 {
			fmt.Sscan(m[1], &r.line)
			r.code, r.msg = m[3], m[4]
		}
	}
	r.full = strings.TrimSpace(normalize(out))
	return r
}

var diagHeader = []string{"case", "id", "compiler", "compiler_version", "rejected", "errors", "first_line", "expected_line", "location_ok", "code", "message", "median_ms"}

func cmdDiag(fs *flag.FlagSet, args []string, w io.Writer) error {
	runs := fs.Int("runs", 3, "timed runs per case (median reported)")
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

	paths, _ := filepath.Glob(filepath.Join(tc.BenchRoot, "diagnostics", "*", "bench.json"))
	var rows [][]string
	var sb strings.Builder
	fmt.Fprintf(&sb, "## Diagnostics\n\n%s vs TypeScript %s. Each case is the same mistake in both languages; *location* says whether the first error points at the line with the mistake.\n\n", tc.RSVersion, orNA(tc.TSVersion))
	var detail strings.Builder
	sb.WriteString("| Case | RS errors | RS location | TS errors | TS location | RS ms | TS ms |\n|---|---:|---|---:|---|---:|---:|\n")
	for _, p := range paths {
		c := diagCase{Dir: filepath.Dir(p)}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &c); err != nil {
			return fmt.Errorf("%s: %v", p, err)
		}
		timed := func(argv []string, re *regexp.Regexp) (diagResult, bool) {
			var res diagResult
			var times []Sample
			rejected := true
			for i := 0; i < *runs; i++ {
				s, err := measure(argv, c.Dir)
				if err != nil {
					return res, false
				}
				rejected = s.Code != 0
				res = parseDiags(s.Stderr+s.Stdout, re)
				times = append(times, s)
			}
			res.ms = stats(times).Median
			return res, rejected
		}
		rs, rsRej := timed([]string{tc.RuneScript, "check", "--color=never", "bad.rsc"}, rsDiag)
		// want 0: the language has no equivalent check, and accepting the
		// program is the expected outcome.
		row := func(name, ver string, r diagResult, rej bool, want int) string {
			ok := r.line == want && rej == (want != 0)
			rows = append(rows, []string{filepath.Base(c.Dir), c.ID, name, ver, fmt.Sprint(rej), fmt.Sprint(r.errors),
				fmt.Sprint(r.line), fmt.Sprint(want), fmt.Sprint(ok), r.code, r.msg, f2(r.ms)})
			if !rej && want == 0 {
				return "accepted (no such check)"
			}
			if !rej {
				return "**accepted**"
			}
			if ok {
				return fmt.Sprintf("line %d ✓", r.line)
			}
			return fmt.Sprintf("line %d (want %d)", r.line, want)
		}
		rsCell := row("runescript", tc.RSVersion, rs, rsRej, c.RSLine)
		tsCell, tsErrs, tsMS := "n/a", "", ""
		var ts diagResult
		if tc.TSC != "" {
			argv := append([]string{tc.Node, tc.TSC}, TSFlags...)
			argv = append(argv, "--outDir", filepath.Join(tc.Work, "diag", c.ID), "bad.ts")
			var tsRej bool
			ts, tsRej = timed(argv, tsDiag)
			tsCell, tsErrs, tsMS = row("tsc", "typescript "+tc.TSVersion, ts, tsRej, c.TSLine), fmt.Sprint(ts.errors), f1(ts.ms)
		}
		fmt.Fprintf(&sb, "| %s %s | %d | %s | %s | %s | %s | %s |\n", c.ID, c.Title, rs.errors, rsCell, tsErrs, tsCell, f1(rs.ms), tsMS)
		show := func(s string) string {
			s = strings.ReplaceAll(s, filepath.ToSlash(c.Dir)+"/", "")
			if s == "" {
				return "(no errors)"
			}
			return s
		}
		fmt.Fprintf(&detail, "### %s %s\n\nRuneScript:\n\n```text\n%s\n```\n\nTypeScript:\n\n```text\n%s\n```\n\n", c.ID, c.Title, show(rs.full), show(ts.full))
	}
	sb.WriteString("\n" + detail.String())
	stamp := time.Now().Format("20060102-150405")
	csvPath, err := writeCSV(tc, *outDir, "diag-"+stamp+".csv", diagHeader, rows)
	if err != nil {
		return err
	}
	mdPath, err := writeFile(tc, *outDir, "diag-"+stamp+".md", sb.String())
	if err != nil {
		return err
	}
	fmt.Fprint(w, sb.String())
	fmt.Fprintf(w, "\nwrote %s and %s\n", csvPath, mdPath)
	return nil
}
