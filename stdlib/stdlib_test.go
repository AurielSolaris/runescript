package stdlib

import (
	"io"
	"testing"

	"github.com/AurielSolaris/runescript/internal/types"
)

// Every native function must return the Go type matching its declared
// result type when called with zero values; a mismatch would panic in the
// runtime's conversion.
func TestSignaturesMatchImplementations(t *testing.T) {
	for _, path := range Paths() {
		for _, f := range Lookup(path).All() {
			name := f.Name
			args := make([]any, len(f.Params))
			for i, p := range f.Params {
				switch p.Kind {
				case types.String:
					args[i] = ""
				case types.Bool:
					args[i] = false
				case types.Signed, types.Unsigned:
					args[i] = int64(0)
				case types.Char:
					args[i] = rune('a')
				case types.Float:
					args[i] = 0.0
				case types.Array:
					args[i] = []string{}
				case types.Func:
					if p.Result.Kind == types.Char {
						args[i] = func(r rune) rune { return r }
					} else {
						args[i] = func(rune) bool { return false }
					}
				default:
					t.Fatalf("std:%s.%s: unsupported parameter type %s", path, name, p)
				}
			}
			out, faulted := call(t, f, args)
			if faulted {
				continue // e.g. fs.readFile(""): a deliberate failure
			}
			ok := false
			switch f.Result.Kind {
			case types.String:
				_, ok = out.(string)
			case types.Bool:
				_, ok = out.(bool)
			case types.Signed, types.Unsigned:
				_, ok = out.(int64)
			case types.Char:
				_, ok = out.(rune)
			case types.Float:
				_, ok = out.(float64)
			case types.Array:
				_, ok = out.([]string)
			case types.Void:
				ok = out == nil
			}
			if !ok {
				t.Errorf("std:%s.%s: declared %s but returned %T", path, name, f.Result, out)
			}
		}
	}
}

// call runs f with args in a test environment (an empty standard input,
// and output discarded). faulted reports a Fault, a deliberate failure.
func call(t *testing.T, f *Func, args []any) (out any, faulted bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(Fault); !ok {
				panic(r)
			}
			faulted = true
		}
	}()
	if f.EnvImpl != nil {
		env := NewEnv(io.Discard, io.Discard, nil, nil)
		dir := t.TempDir()
		env.Getwd = func() (string, error) { return dir, nil }
		return f.EnvImpl(env, args), false
	}
	return f.Impl(args), false
}

func TestStringPackageIsComplete(t *testing.T) {
	// The functions of Go's strings package expressible so far, plus len,
	// runeCount, and slice, which stand in for Go's len(s),
	// utf8.RuneCountInString, and s[i:j] (OQ-STR-03).
	want := []string{
		"len", "runeCount", "slice",
		"containsFunc", "containsRune", "indexFunc", "indexRune", "lastIndexFunc", "map",
		"trimFunc", "trimLeftFunc", "trimRightFunc",
		"fields", "fieldsFunc", "join", "split", "splitAfter", "splitAfterN", "splitN",
		"clone", "compare", "contains", "containsAny", "count", "equalFold",
		"hasPrefix", "hasSuffix", "index", "indexAny", "lastIndex", "lastIndexAny",
		"repeat", "replace", "replaceAll", "toLower", "toTitle", "toUpper",
		"toValidUTF8", "trim", "trimLeft", "trimPrefix", "trimRight", "trimSpace", "trimSuffix",
	}
	p := Lookup("string")
	if p == nil || p.Qualifier != "string" {
		t.Fatal("std:string is not registered")
	}
	if len(p.Funcs) != len(want) {
		t.Errorf("std:string has %d functions, want %d", len(p.Funcs), len(want))
	}
	for _, name := range want {
		if p.Funcs[name] == nil {
			t.Errorf("std:string.%s is missing", name)
		}
	}
}
