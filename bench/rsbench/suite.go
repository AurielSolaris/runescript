package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Bench is one benchmark: a directory under bench/suite holding
// bench.json, main.rsc, main.ts, optional TypeScript variants, and
// expected outputs.
type Bench struct {
	Dir string `json:"-"`

	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Title      string            `json:"title"`
	Categories []string          `json:"categories"`
	Measures   string            `json:"measures"`
	Sizes      map[string]string `json:"sizes"`
	// Notes records semantic differences between the implementations,
	// so a result is never read as a like-for-like comparison when it
	// isn't one.
	Notes string `json:"notes,omitempty"`
}

// sizeOrder is the order sizes are listed in; "test" is the tiny size
// used by verify and by go test.
var sizeOrder = []string{"test", "small", "medium", "large"}

func loadSuite(benchRoot string) ([]*Bench, error) {
	paths, err := filepath.Glob(filepath.Join(benchRoot, "suite", "*", "bench.json"))
	if err != nil {
		return nil, err
	}
	var suite []*Bench
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		b := &Bench{Dir: filepath.Dir(p)}
		if err := json.Unmarshal(data, b); err != nil {
			return nil, fmt.Errorf("%s: %v", p, err)
		}
		if b.ID == "" || b.Name == "" || len(b.Categories) == 0 || b.Sizes["test"] == "" {
			return nil, fmt.Errorf("%s: id, name, categories, and sizes.test are required", p)
		}
		suite = append(suite, b)
	}
	sort.Slice(suite, func(i, j int) bool { return suite[i].Key() < suite[j].Key() })
	return suite, nil
}

// Key is the benchmark's directory name, e.g. "a03_sieve".
func (b *Bench) Key() string { return filepath.Base(b.Dir) }

func (b *Bench) Has(category string) bool {
	for _, c := range b.Categories {
		if c == category {
			return true
		}
	}
	return false
}

// SizeNames returns the benchmark's sizes in sizeOrder.
func (b *Bench) SizeNames() []string {
	var names []string
	for _, s := range sizeOrder {
		if _, ok := b.Sizes[s]; ok {
			names = append(names, s)
		}
	}
	return names
}

// TSVariants maps implementation names to TypeScript sources: main.ts is
// "ts", and <variant>.ts is "ts-<variant>" (e.g. checked.ts is
// "ts-checked", the same algorithm with RuneScript's overflow checks).
func (b *Bench) TSVariants() map[string]string {
	v := map[string]string{}
	files, _ := filepath.Glob(filepath.Join(b.Dir, "*.ts"))
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".ts")
		if name == "main" {
			v["ts"] = f
		} else {
			v["ts-"+name] = f
		}
	}
	return v
}

// Impls lists the implementations this benchmark has, in report order.
func (b *Bench) Impls() []string {
	impls := []string{"rs-vm", "rs-js"}
	var ts []string
	for name := range b.TSVariants() {
		ts = append(ts, name)
	}
	sort.Strings(ts)
	return append(impls, ts...)
}

// Expected returns the expected output for a size and implementation.
// expected/<size>.out applies to every implementation; a TypeScript
// implementation uses expected/<size>.ts.out instead when it exists,
// which documents an intended semantic difference (see Notes).
func (b *Bench) Expected(size, impl string) (string, bool) {
	if strings.HasPrefix(impl, "ts") {
		if data, err := os.ReadFile(filepath.Join(b.Dir, "expected", size+".ts.out")); err == nil {
			return normalize(string(data)), true
		}
	}
	data, err := os.ReadFile(filepath.Join(b.Dir, "expected", size+".out"))
	if err != nil {
		return "", false
	}
	return normalize(string(data)), true
}

// selectBenches picks benchmarks by name, ID, or directory; with no
// names it picks every benchmark in category ("" for all).
func selectBenches(suite []*Bench, category string, names []string) ([]*Bench, error) {
	if len(names) == 0 {
		var out []*Bench
		for _, b := range suite {
			if category == "" || b.Has(category) {
				out = append(out, b)
			}
		}
		return out, nil
	}
	var out []*Bench
	for _, n := range names {
		found := false
		for _, b := range suite {
			if strings.EqualFold(n, b.ID) || n == b.Name || n == b.Key() {
				out = append(out, b)
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("no benchmark named %q (see rsbench list)", n)
		}
	}
	return out, nil
}

func normalize(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
