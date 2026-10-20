package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type sizes struct {
	File      string `json:"file"`
	Raw       int64  `json:"raw"`
	RawGzip   int64  `json:"raw_gzip"`
	RawBrotli int64  `json:"raw_brotli"`
	Min       int64  `json:"min"`
	MinGzip   int64  `json:"min_gzip"`
	MinBrotli int64  `json:"min_brotli"`
}

// measureSizes runs tools/sizes.mjs on files.
func measureSizes(tc *Toolchain, files []string) (map[string]sizes, error) {
	args := append([]string{filepath.Join(tc.BenchRoot, "tools", "sizes.mjs")}, files...)
	cmd := exec.Command(tc.Node, args...)
	cmd.Dir = tc.BenchRoot // so that esbuild resolves from bench/node_modules
	out, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return nil, fmt.Errorf("sizes.mjs: %v\n%s", err, exit.Stderr)
		}
		return nil, err
	}
	m := map[string]sizes{}
	for _, line := range strings.Split(strings.TrimSpace(normalize(string(out))), "\n") {
		var s sizes
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			return nil, fmt.Errorf("sizes.mjs output %q: %v", line, err)
		}
		m[s.File] = s
	}
	return m, nil
}

var sizeHeader = []string{
	"benchmark", "id", "implementation", "compiler_version",
	"raw_bytes", "raw_gzip_bytes", "raw_brotli_bytes",
	"minified_bytes", "minified_gzip_bytes", "minified_brotli_bytes",
}

func cmdSize(fs *flag.FlagSet, args []string, w io.Writer) error {
	outDir := fs.String("out", "", "directory for the CSV and Markdown results (default bench/results)")
	benches, _, err := loadSelected(fs, args, "size")
	if err != nil {
		return err
	}
	tc, err := setup(true)
	if err != nil {
		return err
	}
	defer tc.Close()
	tc.warnMissingTS(func(f string, a ...any) { fmt.Fprintf(os.Stderr, f, a...) })
	if !tc.Esbuild {
		fmt.Fprintln(os.Stderr, "note: esbuild is not installed; minified sizes are left empty (npm install in bench/)")
	}

	type out struct {
		b    *Bench
		impl string
		file string
	}
	var outs []out
	var files []string
	for _, b := range benches {
		dir := filepath.Join(tc.Work, b.Key())
		rs := filepath.Join(dir, "rs", "main.mjs")
		if err := tc.BuildRS(filepath.Join(b.Dir, "main.rsc"), rs); err != nil {
			return err
		}
		outs = append(outs, out{b, "rs-js", rs})
		files = append(files, rs)
		variants := b.TSVariants()
		for _, impl := range b.Impls()[2:] {
			js, err := tc.CompileTS(variants[impl], filepath.Join(dir, impl))
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s/%s unavailable: %s\n", b.Key(), impl, firstLine(err.Error()))
				continue
			}
			outs = append(outs, out{b, impl, js})
			files = append(files, js)
		}
	}
	m, err := measureSizes(tc, files)
	if err != nil {
		return err
	}

	var rows [][]string
	var sb strings.Builder
	fmt.Fprintf(&sb, "## Generated JavaScript size\n\n%s vs TypeScript %s; minified with esbuild, gzip level 9, Brotli quality 11. ", tc.RSVersion, orNA(tc.TSVersion))
	sb.WriteString("RuneScript modules inline the runtime helpers they use; TypeScript 1.0 output has no runtime. Ratio = RS / TS (minified + gzip); Δ = absolute difference.\n\n")
	sb.WriteString("| Benchmark | Impl | Raw | Raw gzip | Min | Min gzip | Min brotli | Ratio | Δ min gzip |\n|---|---|---:|---:|---:|---:|---:|---:|---:|\n")
	tsMinGz := map[*Bench]int64{}
	for _, o := range outs {
		if o.impl == "ts" {
			tsMinGz[o.b] = pick(m[o.file].MinGzip, m[o.file].RawGzip)
		}
	}
	for _, o := range outs {
		s := m[o.file]
		rows = append(rows, []string{o.b.Key(), o.b.ID, o.impl, compilerVersion(tc, o.impl),
			fmt.Sprint(s.Raw), fmt.Sprint(s.RawGzip), fmt.Sprint(s.RawBrotli),
			opt(s.Min), opt(s.MinGzip), opt(s.MinBrotli)})
		rat, diff := "", ""
		if t, ok := tsMinGz[o.b]; ok && o.impl == "rs-js" {
			r := pick(s.MinGzip, s.RawGzip)
			diff = fmt.Sprintf("%+d", r-t)
			// A ratio against a near-empty file is meaningless: only Δ then.
			if t >= 64 {
				rat = fmt.Sprintf("%.2f", float64(r)/float64(t))
			}
		}
		fmt.Fprintf(&sb, "| %s %s | %s | %d | %d | %s | %s | %s | %s | %s |\n", o.b.ID, o.b.Name, o.impl,
			s.Raw, s.RawGzip, opt(s.Min), opt(s.MinGzip), opt(s.MinBrotli), rat, diff)
	}
	stamp := time.Now().Format("20060102-150405")
	csvPath, err := writeCSV(tc, *outDir, "size-"+stamp+".csv", sizeHeader, rows)
	if err != nil {
		return err
	}
	mdPath, err := writeFile(tc, *outDir, "size-"+stamp+".md", sb.String())
	if err != nil {
		return err
	}
	fmt.Fprint(w, sb.String())
	fmt.Fprintf(w, "\nwrote %s and %s\n", csvPath, mdPath)
	return nil
}

func pick(v, fallback int64) int64 {
	if v < 0 {
		return fallback
	}
	return v
}

func opt(v int64) string {
	if v < 0 {
		return ""
	}
	return fmt.Sprint(v)
}
