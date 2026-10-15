// Package stdlib holds RuneScript's standard library packages, implemented
// natively in Go and made available with `use "std:<name>";`.
//
// Native functions exchange values with the runtime as plain Go values:
// int64 for every integer type, string, and bool. The runtime converts
// arguments and range-checks integer results against the declared types.
package stdlib

import "github.com/AurielSolaris/runescript/internal/types"

// Func is a natively implemented function.
type Func struct {
	Name   string
	Params []*types.Type
	Result *types.Type
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

// fn builds a Func; it is a small helper for package tables.
func fn(name string, result *types.Type, impl func(args []any) any, params ...*types.Type) *Func {
	return &Func{Name: name, Params: params, Result: result, Impl: impl}
}
