package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

func loadSelected(fs *flag.FlagSet, args []string, defaultCategory string) ([]*Bench, *string, error) {
	category := fs.String("category", defaultCategory, "only benchmarks in this category (runtime, size, conformance; empty for all)")
	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}
	root, err := findRoot()
	if err != nil {
		return nil, nil, err
	}
	suite, err := loadSuite(filepath.Join(root, "bench"))
	if err != nil {
		return nil, nil, err
	}
	benches, err := selectBenches(suite, *category, fs.Args())
	return benches, category, err
}

func cmdList(fs *flag.FlagSet, args []string, w io.Writer) error {
	benches, _, err := loadSelected(fs, args, "")
	if err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tDIRECTORY\tCATEGORIES\tIMPLEMENTATIONS\tSIZES\tTITLE")
	for _, b := range benches {
		var sizes []string
		for _, s := range b.SizeNames() {
			sizes = append(sizes, s+"="+b.Sizes[s])
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", b.ID, b.Key(), strings.Join(b.Categories, ","),
			strings.Join(b.Impls(), ","), strings.Join(sizes, " "), b.Title)
	}
	return tw.Flush()
}

// check runs p once at a size and compares the output with the expected
// output. It returns "pass", or a short reason for the failure.
func check(b *Bench, p *Program, size string) (string, Sample) {
	if p.Err != nil {
		return "unavailable: " + firstLine(p.Err.Error()), Sample{}
	}
	s, err := measure(append(p.Argv, b.Sizes[size]), p.Dir)
	if err != nil {
		return "error: " + err.Error(), s
	}
	if s.Code != 0 {
		return fmt.Sprintf("exit %d: %s", s.Code, firstLine(s.Stderr)), s
	}
	want, ok := b.Expected(size, p.Impl)
	if !ok {
		return "no expected/" + size + ".out (run rsbench bless)", s
	}
	if s.Stdout != want {
		return "wrong output: " + firstDiff(s.Stdout, want), s
	}
	return "pass", s
}

func firstDiff(got, want string) string {
	g, x := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(g) || i < len(x); i++ {
		var a, b string
		if i < len(g) {
			a = g[i]
		}
		if i < len(x) {
			b = x[i]
		}
		if a != b {
			return fmt.Sprintf("line %d: got %q, want %q", i+1, a, b)
		}
	}
	return ""
}

func cmdVerify(fs *flag.FlagSet, args []string, w io.Writer) error {
	size := fs.String("size", "test", "the size to verify (needs expected/<size>.out)")
	benches, _, err := loadSelected(fs, args, "")
	if err != nil {
		return err
	}
	tc, err := setup(true)
	if err != nil {
		return err
	}
	defer tc.Close()
	tc.warnMissingTS(func(f string, a ...any) { fmt.Fprintf(os.Stderr, f, a...) })

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tBENCHMARK\tIMPL\tRESULT")
	failed := 0
	for _, b := range benches {
		if _, ok := b.Sizes[*size]; !ok {
			continue
		}
		for _, p := range tc.prepare(b) {
			res, _ := check(b, p, *size)
			if res != "pass" && !strings.HasPrefix(res, "unavailable") {
				failed++
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", b.ID, b.Key(), p.Impl, res)
		}
	}
	tw.Flush()
	if failed > 0 {
		return fmt.Errorf("%d check(s) failed", failed)
	}
	return nil
}

// cmdBless records the interpreter's output as the expected output. The
// other implementations are then held to it by verify. With -from ts it
// records TypeScript's output as expected/<size>.ts.out instead: an
// intended semantic difference, which the benchmark's notes must explain.
func cmdBless(fs *flag.FlagSet, args []string, w io.Writer) error {
	size := fs.String("size", "test", "the size to record")
	from := fs.String("from", "rs-vm", "the implementation to record: rs-vm, or ts for an intended TypeScript divergence")
	benches, _, err := loadSelected(fs, args, "")
	if err != nil {
		return err
	}
	if *from != "rs-vm" && *from != "ts" {
		return fmt.Errorf("-from must be rs-vm or ts")
	}
	if *from == "ts" && len(fs.Args()) == 0 {
		return fmt.Errorf("-from ts records a divergence: name the benchmarks explicitly")
	}
	tc, err := setup(*from == "ts")
	if err != nil {
		return err
	}
	defer tc.Close()
	for _, b := range benches {
		arg, ok := b.Sizes[*size]
		if !ok {
			continue
		}
		if *from == "ts" && b.Notes == "" {
			return fmt.Errorf("%s: explain the TypeScript divergence in bench.json notes first", b.Key())
		}
		var p *Program
		for _, q := range tc.prepare(b) {
			if q.Impl == *from {
				p = q
			}
		}
		if p == nil || p.Err != nil {
			return fmt.Errorf("%s: %s is unavailable", b.Key(), *from)
		}
		s, err := measure(append(p.Argv, arg), p.Dir)
		if err != nil {
			return err
		}
		if s.Code != 0 {
			return fmt.Errorf("%s: %s exited with %d: %s", b.Key(), *from, s.Code, firstLine(s.Stderr))
		}
		name := *size + ".out"
		if *from == "ts" {
			name = *size + ".ts.out"
		}
		path := filepath.Join(b.Dir, "expected", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(s.Stdout), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(w, "wrote %s\n", filepath.ToSlash(must(filepath.Rel(tc.Root, path))))
	}
	return nil
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
