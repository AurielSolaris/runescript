package jsgen

import (
	"sync"
	"unicode"
)

// Unicode tables for the JavaScript runtime, computed from Go's unicode
// package so that std:string behaves exactly as it does in the interpreter
// (JavaScript's own toUpperCase, trim, etc. use different rules).
//
// A table is a list of runs [start, count, step, delta, ...]: code point
// start+k*step maps to itself plus delta, for k < count.

var (
	upperRuns = sync.OnceValue(func() []int64 { return mappingRuns(unicode.ToUpper) })
	lowerRuns = sync.OnceValue(func() []int64 { return mappingRuns(unicode.ToLower) })
	titleRuns = sync.OnceValue(func() []int64 { return mappingRuns(unicode.ToTitle) })
	// foldRuns maps each code point to the smallest member of its
	// unicode.SimpleFold orbit; two code points are equal under case
	// folding (strings.EqualFold) exactly when they map to the same one.
	foldRuns = sync.OnceValue(func() []int64 {
		return mappingRuns(func(r rune) rune {
			least := r
			for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
				least = min(least, f)
			}
			return least
		})
	})
)

// spaces lists the code points unicode.IsSpace accepts (strings.TrimSpace).
func spaces() []int64 {
	var out []int64
	for r := rune(0); r <= 0xFFFF; r++ {
		if unicode.IsSpace(r) {
			out = append(out, int64(r))
		}
	}
	return out
}

func mappingRuns(f func(rune) rune) []int64 {
	type entry struct{ r, d int64 }
	var entries []entry
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue // surrogates never occur in RuneScript strings
		}
		if m := f(r); m != r {
			entries = append(entries, entry{int64(r), int64(m - r)})
		}
	}
	var runs []int64
	for i := 0; i < len(entries); {
		start, d := entries[i].r, entries[i].d
		step := int64(1)
		if i+1 < len(entries) && entries[i+1].d == d && entries[i+1].r-start <= 2 {
			step = entries[i+1].r - start
		}
		count := int64(1)
		for i+int(count) < len(entries) {
			e := entries[i+int(count)]
			if e.d != d || e.r != start+count*step {
				break
			}
			count++
		}
		runs = append(runs, start, count, step, d)
		i += int(count)
	}
	return runs
}
