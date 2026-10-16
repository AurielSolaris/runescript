// Package jshost describes the host-neutral JavaScript modules ("rune:fs",
// ...) that extern declarations may name, and how each is provided on a
// JavaScript host.
//
// A program that only uses rune: modules does not depend on one host:
// the code generator maps each function to the host's own API. Today the
// only host is Node.js; the planned bundled runtime (OQ-JS-08) will be a
// second one. The function lists are provisional (OQ-JS-02).
package jshost

import (
	"sort"
	"strings"

	"github.com/AurielSolaris/runescript/internal/types"
)

// Prefix starts the name of every host-neutral module.
const Prefix = "rune:"

// Func is one function of a host-neutral module.
type Func struct {
	Name   string
	Params []*types.Type
	Result *types.Type
	Doc    string
	// Node is the function's implementation on Node.js: a JavaScript
	// arrow function that may refer to the namespace imported from
	// Module.Node as $host.
	Node string
}

// Type is the function's RuneScript type.
func (f *Func) Type() *types.Type { return types.NewFunc(f.Params, f.Result) }

// Module is a host-neutral module.
type Module struct {
	Name  string // e.g. "rune:fs"
	Funcs map[string]*Func
	// Node is the Node.js module the implementations use, e.g. "node:fs".
	Node string
}

var (
	str = types.StringT
	bln = types.BoolT
	vd  = types.VoidT
)

var modules = map[string]*Module{}

func register(m *Module, funcs ...*Func) {
	m.Funcs = map[string]*Func{}
	for _, f := range funcs {
		m.Funcs[f.Name] = f
	}
	modules[m.Name] = m
}

func init() {
	// Files are read and written as UTF-8. Invalid UTF-8 read from a file
	// becomes U+FFFD, so the result is always a valid RuneScript string.
	register(&Module{Name: "rune:fs", Node: "node:fs"},
		&Func{Name: "readFile", Params: []*types.Type{str}, Result: str,
			Doc: "the contents of a file, as UTF-8", Node: `(path) => $host.readFileSync(path, "utf8")`},
		&Func{Name: "writeFile", Params: []*types.Type{str, str}, Result: vd,
			Doc: "creates or replaces a file", Node: `(path, data) => { $host.writeFileSync(path, data); }`},
		&Func{Name: "appendFile", Params: []*types.Type{str, str}, Result: vd,
			Doc: "adds to the end of a file, creating it if needed", Node: `(path, data) => { $host.appendFileSync(path, data); }`},
		&Func{Name: "exists", Params: []*types.Type{str}, Result: bln,
			Doc: "whether a file or directory exists", Node: `(path) => $host.existsSync(path)`},
		&Func{Name: "remove", Params: []*types.Type{str}, Result: vd,
			Doc: "deletes a file", Node: `(path) => { $host.unlinkSync(path); }`},
	)
}

// Lookup returns the host-neutral module called name ("rune:fs"), or nil.
func Lookup(name string) *Module { return modules[name] }

// Names returns the names of every host-neutral module, sorted.
func Names() []string {
	var names []string
	for n := range modules {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// FuncNames returns m's function names, sorted.
func (m *Module) FuncNames() []string {
	var names []string
	for n := range m.Funcs {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// HostSpecific reports whether module names the built-in modules of one
// JavaScript host, such as "node:fs".
func HostSpecific(module string) bool {
	for _, p := range []string{"node:", "bun:", "deno:"} {
		if strings.HasPrefix(module, p) {
			return true
		}
	}
	return false
}
