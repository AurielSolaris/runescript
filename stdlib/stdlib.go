// Package stdlib holds RuneScript's standard library packages, implemented
// natively in Go and made available with `use "std:<name>";`.
//
// Native functions exchange values with the runtime as plain Go values:
// int64 for every integer type, string, and bool. The runtime converts
// arguments and range-checks integer results against the declared types.
package stdlib

import (
	"fmt"

	"github.com/AurielSolaris/runescript/internal/types"
)

// Func is a natively implemented function.
type Func struct {
	Name string
	// Package is the path of the package the function belongs to, e.g.
	// "string". It is set when the package is registered.
	Package string
	Params  []*types.Type
	Result  *types.Type
	// Impl receives one Go value per parameter and returns the result
	// (nil for void). A panic inside Impl becomes a RuneScript runtime error.
	Impl func(args []any) any
}

// Package is a standard library package such as std:string.
type Package struct {
	// Path is the text after "std:" in a use declaration.
	Path string
	// Qualifier is the name used to call the package's functions, e.g.
	// `string` in string.toUpper(s). It is the last segment of Path.
	Qualifier string
	Funcs     map[string]*Func
}

var packages = map[string]*Package{}

func register(p *Package) {
	for _, f := range p.Funcs {
		f.Package = p.Path
	}
	packages[p.Path] = p
}

// Lookup returns the std package with the given path (without "std:").
func Lookup(path string) *Package { return packages[path] }

// Paths lists the registered package paths.
func Paths() []string {
	var out []string
	for p := range packages {
		out = append(out, p)
	}
	return out
}

// Fault is a deliberate failure of a native function: Impl panics with a
// Fault to throw a RuneScript error of type Type (RangeError, TypeError,
// ArgumentError, or Error) with the message "<qualifier>.<name>: Message".
// Any other panic becomes an Error. The JavaScript runtime has the same
// mechanism (Fault in runtime.js).
type Fault struct {
	Type    string
	Message string
}

func rangeFault(format string, args ...any) {
	panic(Fault{Type: "RangeError", Message: fmt.Sprintf(format, args...)})
}

// fn builds a Func; it is a small helper for package tables.
func fn(name string, result *types.Type, impl func(args []any) any, params ...*types.Type) *Func {
	return &Func{Name: name, Params: params, Result: result, Impl: impl}
}
