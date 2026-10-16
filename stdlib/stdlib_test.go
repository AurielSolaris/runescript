package stdlib

import (
	"testing"

	"github.com/AurielSolaris/runescript/internal/types"
)

// Every native function must return the Go type matching its declared
// result type when called with zero values; a mismatch would panic in the
// runtime's conversion.
func TestSignaturesMatchImplementations(t *testing.T) {
	for _, path := range Paths() {
		for name, f := range Lookup(path).Funcs {
			if name != f.Name {
				t.Errorf("std:%s: registered as %q but named %q", path, name, f.Name)
			}
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
			out := f.Impl(args)
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
