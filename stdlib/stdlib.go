// Package stdlib holds RuneScript's standard library packages, implemented
// natively in Go and made available with `use "std:<name>";`.
//
// Native functions exchange values with the runtime as plain Go values:
// int64 for every integer type, float64, string, bool, rune for char,
// []string for Array<string>, and Go functions for callbacks. The runtime
// converts arguments and range-checks integer results against the
// declared types.
package stdlib

import (
	"fmt"
	"math/big"
	"sort"

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
	// EnvImpl, set instead of Impl, also receives the program's Env.
	EnvImpl func(env *Env, args []any) any
	// JSName is the function's name in the JavaScript runtime's package
	// object: Name, except for overloads, which need names of their own.
	JSName string
}

// Const is a package constant such as math.pi. Like a literal, it is
// untyped: it takes the type of where it is used, if it fits.
type Const struct {
	Int   *big.Int // an integer constant, or nil
	Float float64  // a floating-point constant, when Int is nil
}

// Package is a standard library package such as std:string.
type Package struct {
	// Path is the text after "std:" in a use declaration.
	Path string
	// Qualifier is the name used to call the package's functions, e.g.
	// `string` in string.toUpper(s). It is the last segment of Path.
	Qualifier string
	// Funcs maps each name to its function, or to the first of its
	// Overloads.
	Funcs map[string]*Func
	// Overloads lists the signatures of a name that has several, in the
	// order calls try them (OQ-MATH-01: math.max on ints, then floats).
	Overloads map[string][]*Func
	Consts    map[string]Const
}

// All returns every function of p, each overload included, sorted by
// JSName (which is unique).
func (p *Package) All() []*Func {
	var out []*Func
	for name, f := range p.Funcs {
		if fs := p.Overloads[name]; len(fs) > 1 {
			out = append(out, fs...)
		} else {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].JSName < out[j].JSName })
	return out
}

// add registers fs in p. A name given more than once is overloaded.
func (p *Package) add(fs ...*Func) {
	if p.Funcs == nil {
		p.Funcs, p.Overloads, p.Consts = map[string]*Func{}, map[string][]*Func{}, map[string]Const{}
	}
	for _, f := range fs {
		if first, ok := p.Funcs[f.Name]; ok {
			if len(p.Overloads[f.Name]) == 0 {
				p.Overloads[f.Name] = []*Func{first}
			}
			p.Overloads[f.Name] = append(p.Overloads[f.Name], f)
			continue
		}
		p.Funcs[f.Name] = f
	}
}

var packages = map[string]*Package{}

func register(p *Package) {
	if p.Overloads == nil {
		p.Overloads, p.Consts = map[string][]*Func{}, map[string]Const{}
	}
	for _, f := range p.Funcs {
		f.Package = p.Path
	}
	for _, fs := range p.Overloads {
		for _, f := range fs {
			f.Package = p.Path
		}
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
	return &Func{Name: name, Params: params, Result: result, Impl: impl, JSName: name}
}
