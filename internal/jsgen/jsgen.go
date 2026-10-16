// Package jsgen compiles RuneScript IR to a JavaScript module. The module
// uses only ES2015 plus BigInt, ??, and ES modules (OQ-JS-07).
//
// Type safety is preserved in two layers:
//
//   - Static: the checker accepted the program, and ir.Verify re-checks the
//     IR's types before anything is emitted. Types are then erased: a
//     well-typed program cannot store a value of the wrong type, so the
//     generated code needs no tag checks on variables or parameters.
//   - Dynamic: wherever JavaScript's semantics differ from RuneScript's,
//     the generated code calls the runtime (runtime.js) instead of using
//     the JavaScript operator: integer arithmetic is range-checked and
//     traps instead of producing doubles, division truncates, division by
//     zero is an error, call depth is limited, and std:string works in
//     UTF-8 byte offsets.
//   - Boundary: values returned by extern (JavaScript) functions are
//     checked against their declared types when they cross into
//     RuneScript, since JavaScript code is not type-checked.
//
// Runtime errors carry the same diagnostic codes, messages, and source
// positions as the interpreter's.
package jsgen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/AurielSolaris/runescript/internal/conv"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/ir"
	"github.com/AurielSolaris/runescript/internal/jshost"
	"github.com/AurielSolaris/runescript/internal/runtime"
	"github.com/AurielSolaris/runescript/internal/types"
)

//go:embed runtime.js
var runtimeJS string

// Options configures code generation.
type Options struct {
	// Version is the compiler version recorded in the output's header.
	Version string
}

// Generate compiles p to a JavaScript module. Diagnostics report programs
// the JavaScript target cannot represent; err reports an internal compiler
// error.
func Generate(p *ir.Program, opts Options) (out []byte, diags []*diagnostics.Diagnostic, err error) {
	if err := ir.Verify(p); err != nil {
		return nil, nil, err
	}
	g := &gen{indent: 1, sites: map[diagnostics.Pos]int{}, files: map[*diagnostics.Source]int{},
		helpers: map[string]bool{}, tables: map[string]bool{}}
	pkgs := map[string]bool{}
	for _, f := range p.Funcs {
		pkgs[f.Pkg] = true
	}
	g.name(p.Funcs, len(pkgs) > 1)
	g.externNames(p.Externs)
	for _, f := range p.Funcs {
		g.funcDecl(f)
	}
	if len(g.diags) > 0 {
		return nil, g.diags, nil
	}
	entrySite := g.site(p.EntrySite)
	return []byte(g.module(p, opts, entrySite)), nil, nil
}

// Runtime returns the runtime with every section and table present. It is
// used to test the runtime directly.
func Runtime() string {
	return fillRuntime(map[string]bool{"upper": true, "lower": true, "title": true, "fold": true},
		map[string]bool{"extern": true, "dom": true, "string": true})
}

// helperSections names the optional runtime section (runtime.js,
// `//@section`) each helper lives in. A program gets a section only if it
// uses one of its helpers, so scripts carry no DOM or extern code: the web
// support is a layer on top of the scripting language, not part of it.
var helperSections = map[string]string{
	"ext": "extern", "std": "string",
	"signal": "dom", "el": "dom", "text": "dom", "attr": "dom", "bindText": "dom",
	"bindNode": "dom", "bindAttr": "dom", "on": "dom", "when": "dom", "renderApp": "dom",
}

// sections removes the optional runtime sections not in keep, and the
// section markers themselves.
func sections(src string, keep map[string]bool) string {
	var out strings.Builder
	skip, open := false, false
	for _, line := range strings.SplitAfter(src, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "//@section "):
			if open {
				panic("jsgen: nested runtime section")
			}
			name := strings.TrimPrefix(t, "//@section ")
			if _, ok := sectionNames[name]; !ok {
				panic("jsgen: unknown runtime section " + name)
			}
			skip, open = !keep[name], true
		case t == "//@end":
			skip, open = false, false
		case !skip:
			out.WriteString(line)
		}
	}
	if open {
		panic("jsgen: unterminated runtime section")
	}
	return out.String()
}

var sectionNames = map[string]bool{"extern": true, "dom": true, "string": true}

type gen struct {
	b       strings.Builder
	indent  int
	sites   map[diagnostics.Pos]int
	order   []diagnostics.Pos
	files   map[*diagnostics.Source]int // index in the module's file table
	names   []string                    // the file table
	fnames  map[*ir.Func]string         // JavaScript names of top-level functions
	helpers map[string]bool             // runtime members used, e.g. "addI"
	tables  map[string]bool             // Unicode tables used, e.g. "upper"
	tmps    int                         // template node variables so far
	diags   []*diagnostics.Diagnostic

	// imports are the JavaScript modules the program imports, in order,
	// with the namespace each is imported as. externs maps each extern to
	// the expression naming its function.
	imports []jsImport
	externs map[*ir.Extern]string
	// adapters are the host-neutral modules used, with the namespace of
	// the host module their functions are written against.
	adapters []*adapter
}

type jsImport struct{ module, ns string }

type adapter struct {
	module *jshost.Module
	name   string // the adapter object, e.g. "$r0"
	host   string // the imported host module's namespace, e.g. "$h0"
	funcs  []*jshost.Func
}

func (g *gen) line(format string, args ...any) {
	g.b.WriteString(strings.Repeat("  ", g.indent))
	fmt.Fprintf(&g.b, format, args...)
	g.b.WriteByte('\n')
}

// site returns the index of span's start in the site table.
func (g *gen) site(span diagnostics.Span) int {
	if i, ok := g.sites[span.Start]; ok {
		return i
	}
	i := len(g.order)
	g.sites[span.Start] = i
	g.order = append(g.order, span.Start)
	if f := span.Start.File; f != nil {
		if _, ok := g.files[f]; !ok {
			g.files[f] = len(g.names)
			g.names = append(g.names, f.Name)
		}
	}
	return i
}

// name chooses the JavaScript names of top-level functions. When the
// program has several packages they are named pkg$name (org$x$geom$area).
// Private functions of different files may share a name; the second one
// becomes name$2, and so on. RuneScript names never contain '$', so these
// names cannot collide with each other or with nested functions and
// variables.
func (g *gen) name(funcs []*ir.Func, qualify bool) {
	g.fnames = map[*ir.Func]string{}
	used := map[string]int{}
	for _, f := range funcs {
		name := jsName(f.Name)
		if qualify {
			name = strings.ReplaceAll(f.Pkg, ".", "$") + "$" + f.Name
		}
		used[name]++
		if n := used[name]; n > 1 {
			name += "$" + strconv.Itoa(n)
		}
		g.fnames[f] = name
	}
}

// funcName is f's JavaScript name.
func (g *gen) funcName(f *ir.Func) string {
	if name, ok := g.fnames[f]; ok {
		return name
	}
	return jsName(f.Name)
}

func (g *gen) use(helper string) string {
	g.helpers[helper] = true
	return "$" + helper
}

// ---- module ----

func (g *gen) module(p *ir.Program, opts Options, entrySite int) string {
	var m strings.Builder
	fmt.Fprintf(&m, "// Generated by runescript %s from %s. Do not edit.\n", opts.Version, entryFile(p))
	m.WriteString("// An ES module (ES2015 + BigInt and ??): `node <this file>` runs it;\n")
	m.WriteString("// browsers and other modules call the exported run().\n\n")
	for _, im := range g.imports {
		fmt.Fprintf(&m, "import * as %s from %s;\n", im.ns, quote(im.module))
	}
	if len(g.imports) > 0 {
		m.WriteString("\n")
	}
	keep := map[string]bool{}
	for h := range g.helpers {
		if sec, ok := helperSections[h]; ok {
			keep[sec] = true
		}
	}
	m.WriteString(fillRuntime(g.tables, keep))
	m.WriteString("\n// ---- program ----\n\n")
	m.WriteString("const $program = (($rt) => {\n")
	m.WriteString("  \"use strict\";\n")
	if len(g.helpers) > 0 {
		var names []string
		for h := range g.helpers {
			names = append(names, h+": $"+h)
		}
		sort.Strings(names)
		fmt.Fprintf(&m, "  const { %s } = $rt;\n", strings.Join(names, ", "))
	}
	for _, a := range g.adapters {
		fmt.Fprintf(&m, "  // %s on Node.js\n", a.module.Name)
		fmt.Fprintf(&m, "  const %s = {\n", a.name)
		for _, f := range a.funcs {
			fmt.Fprintf(&m, "    %s: %s,\n", f.Name, strings.ReplaceAll(f.Node, "$host", a.host))
		}
		m.WriteString("  };\n")
	}
	m.WriteString("\n")
	m.WriteString(g.b.String())
	m.WriteString("\n  return {\n")
	quoted := make([]string, len(g.names))
	for i, name := range g.names {
		quoted[i] = quote(name)
	}
	fmt.Fprintf(&m, "    files: [%s],\n", strings.Join(quoted, ", "))
	m.WriteString("    // Source positions of runtime errors, as [file, line, column].\n")
	m.WriteString("    sites: [")
	for i, pos := range g.order {
		if i > 0 {
			m.WriteString(", ")
		}
		fmt.Fprintf(&m, "[%d, %d, %d]", g.files[pos.File], pos.Line, pos.Col)
	}
	m.WriteString("],\n")
	fmt.Fprintf(&m, "    main: %s,\n", g.funcName(p.Entry))
	fmt.Fprintf(&m, "    entrySite: %d,\n", entrySite)
	fmt.Fprintf(&m, "    entryVoid: %t,\n", p.EntryVoid)
	m.WriteString("  };\n")
	m.WriteString("})($rt);\n\n")
	m.WriteString(moduleFooter)
	return m.String()
}

const moduleFooter = `export { RuneScriptError };

// RSError is the class of RuneScript error values. JavaScript called
// through extern functions receives errors as RSError objects, and may
// return or throw them: new RSError("RangeError", "message", errorID).
export const RSError = $rt.RSError;

// run executes main and returns the exit code: main's int result, or 0 for
// a void main. A runtime error is thrown as a RuneScriptError. options.trace
// receives each line written by trace (default: console.log).
export function run(options = {}) {
  return $rt.run($program, options);
}

if (import.meta.main) {
  try {
    process.exitCode = run({ trace: (line) => process.stdout.write(line + "\n") });
  } catch (e) {
    if (!(e instanceof RuneScriptError)) throw e;
    process.stderr.write(e.message + "\n");
    process.exitCode = 1;
  }
}
`

func fillRuntime(tables, keep map[string]bool) string {
	codes, _ := json.Marshal(map[string]string{
		"callDepth":     diagnostics.RunCallDepth.String(),
		"overflow":      diagnostics.RunOverflow.String(),
		"divideByZero":  diagnostics.RunDivideByZero.String(),
		"nativeFailed":  diagnostics.RunNativeFailed.String(),
		"uncaught":      diagnostics.RunUncaught.String(),
		"conversion":    diagnostics.RunConversion.String(),
		"externResult":  diagnostics.RunExternResult.String(),
		"externThrew":   diagnostics.RunExternThrew.String(),
		"noMountTarget": diagnostics.RunNoMountTarget.String(),
		"domValue":      diagnostics.RunDOMValue.String(),
		"index":         diagnostics.RunIndex.String(),
		"shift":         diagnostics.RunShift.String(),
	})
	table := func(name string, runs func() []int64) string {
		if !tables[name] {
			return "[]"
		}
		return ints(runs())
	}
	return strings.NewReplacer(
		"/*@CODES@*/", string(codes),
		"/*@MAX_DEPTH@*/", strconv.Itoa(runtime.MaxCallDepth),
		"/*@UPPER@*/", table("upper", upperRuns),
		"/*@LOWER@*/", table("lower", lowerRuns),
		"/*@TITLE@*/", table("title", titleRuns),
		"/*@FOLD@*/", table("fold", foldRuns),
		"/*@SPACE@*/", ints(spaces()),
		"/*@F32STR@*/", conv.Float32JS,
	).Replace(sections(runtimeJS, keep))
}

func ints(v []int64) string {
	parts := make([]string, len(v))
	for i, n := range v {
		parts[i] = strconv.FormatInt(n, 10)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// ---- functions and statements ----

// funcDecl emits a named function as a declaration. Function declarations
// in blocks are block-scoped in strict mode, matching nested functions.
func (g *gen) funcDecl(f *ir.Func) {
	g.line("function %s(%s) {", g.funcName(f), params(f))
	g.indent++
	g.funcBody(f)
	g.indent--
	g.line("}")
}

func params(f *ir.Func) string {
	var ps []string
	for _, p := range f.Params {
		ps = append(ps, jsName(p.Name))
	}
	return strings.Join(append(ps, "$site"), ", ")
}

// funcBody wraps the body with call-depth accounting. $site is the caller's
// call site, passed as a trailing argument.
func (g *gen) funcBody(f *ir.Func) {
	g.line("%s($site);", g.use("enter"))
	g.line("try {")
	g.indent++
	g.stmts(f.Body.Stmts)
	g.indent--
	g.line("} catch ($e) {")
	g.line("  throw %s($e, $site);", g.use("unwind"))
	g.line("} finally {")
	g.line("  %s();", g.use("leave"))
	g.line("}")
}

func (g *gen) stmts(list []ir.Stmt) {
	for _, s := range list {
		g.stmt(s)
	}
}

func (g *gen) block(b *ir.Block) {
	g.indent++
	g.stmts(b.Stmts)
	g.indent--
}

func (g *gen) stmt(s ir.Stmt) {
	switch s := s.(type) {
	case *ir.FuncStmt:
		g.funcDecl(s.Func)
	case *ir.Decl, *ir.Assign, *ir.Eval, *ir.IndexSet:
		g.line("%s;", g.simple(s))
	case *ir.Return:
		if s.Value == nil {
			g.line("return;")
		} else {
			g.line("return %s;", g.expr(s.Value, precLowest))
		}
	case *ir.If:
		g.line("if (%s) {", g.expr(s.Cond, precLowest))
		g.ifRest(s)
	case *ir.While:
		g.line("while (%s) {", g.expr(s.Cond, precLowest))
		g.block(s.Body)
		g.line("}")
	case *ir.For:
		var init, cond, post string
		if s.Init != nil {
			init = g.simple(s.Init)
		}
		if s.Cond != nil {
			cond = " " + g.expr(s.Cond, precLowest)
		}
		if s.Post != nil {
			post = " " + g.simple(s.Post)
		}
		g.line("for (%s;%s;%s) {", init, cond, post)
		g.block(s.Body)
		g.line("}")
	case *ir.Break:
		g.line("break;")
	case *ir.Continue:
		g.line("continue;")
	case *ir.Throw:
		g.line("throw %s(%s, %d);", g.use("thrown"), g.expr(s.X, precAssign), g.site(s.Site))
	case *ir.Try:
		g.try(s)
	default:
		panic(fmt.Sprintf("jsgen: unexpected statement %T", s))
	}
}

// try emits try/catch/finally. The caught value goes through $caught,
// which turns runtime errors into RuneScript errors and lets errors that
// cannot be caught pass. Catch clauses become an if chain on the error's
// type. A finally block is skipped while the program is stopping for an
// error that cannot be caught, as in the interpreter.
func (g *gen) try(s *ir.Try) {
	g.line("try {")
	g.block(s.Body)
	if len(s.Catches) > 0 {
		g.line("} catch ($e) {")
		g.indent++
		g.line("const $x = %s($e, $site);", g.use("caught"))
		for i, c := range s.Catches {
			kw := "} else if"
			if i == 0 {
				kw = "if"
			}
			g.line("%s (%s($x, %s)) {", kw, g.use("isA"), quote(c.T.Name))
			g.indent++
			g.line("const %s = $x;", jsName(c.Var.Name))
			g.stmts(c.Body.Stmts)
			g.indent--
		}
		g.line("} else {")
		g.line("  throw $x;")
		g.line("}")
		g.indent--
	}
	if s.Finally != nil {
		g.line("} finally {")
		g.line("  if (!%s()) {", g.use("fatal"))
		g.indent += 2
		g.stmts(s.Finally.Stmts)
		g.indent -= 2
		g.line("  }")
	}
	g.line("}")
}

// ifRest emits the body of an if whose header is already written, and its
// else-if chain.
func (g *gen) ifRest(s *ir.If) {
	g.block(s.Then)
	switch e := s.Else.(type) {
	case nil:
		g.line("}")
	case *ir.Block:
		g.line("} else {")
		g.block(e)
		g.line("}")
	case *ir.If:
		g.line("} else if (%s) {", g.expr(e.Cond, precLowest))
		g.ifRest(e)
	}
}

// simple renders a declaration, assignment, or call without its
// semicolon, so it can also appear in a for header.
func (g *gen) simple(s ir.Stmt) string {
	switch s := s.(type) {
	case *ir.Decl:
		return fmt.Sprintf("let %s = %s", jsName(s.Var.Name), g.expr(s.Init, precAssign))
	case *ir.Assign:
		return fmt.Sprintf("%s = %s", jsName(s.Var.Name), g.expr(s.Value, precAssign))
	case *ir.Eval:
		return g.expr(s.X, precLowest)
	case *ir.IndexSet:
		site := g.site(s.Site)
		if s.Compound {
			// $o is the old element (ir.IndexOld) in the value.
			return fmt.Sprintf("%s(%s, %s, ($o) => %s, %d)", g.use("aupd"), g.expr(s.X, precAssign),
				g.expr(s.I, precAssign), g.expr(s.Value, precAssign), site)
		}
		return fmt.Sprintf("%s(%s, %s, %s, %d)", g.use("aset"), g.expr(s.X, precAssign),
			g.expr(s.I, precAssign), g.expr(s.Value, precAssign), site)
	}
	panic(fmt.Sprintf("jsgen: unexpected simple statement %T", s))
}

// ---- expressions ----

// JavaScript precedence levels used here, lowest first.
const (
	precLowest = iota
	precAssign // also arrow functions
	precOr
	precAnd
	precEquality
	precRelational
	precUnary
	precCall // calls and primary expressions
)

var compareOps = map[string]struct {
	js   string
	prec int
}{
	"==": {"===", precEquality}, "!=": {"!==", precEquality},
	"<": {"<", precRelational}, "<=": {"<=", precRelational},
	">": {">", precRelational}, ">=": {">=", precRelational},
}

var arithHelpers = map[string]string{"+": "add", "-": "sub", "*": "mul", "/": "div", "%": "mod"}

// expr renders e, parenthesized if its precedence is below min.
func (g *gen) expr(e ir.Expr, min int) string {
	s, prec := g.exprPrec(e)
	if prec < min {
		return "(" + s + ")"
	}
	return s
}

func (g *gen) exprPrec(e ir.Expr) (string, int) {
	switch e := e.(type) {
	case *ir.IntConst:
		s := types.Big(e.T, e.Value).String()
		if types.IsWide(e.T) {
			s += "n" // a BigInt
		}
		if strings.HasPrefix(s, "-") {
			return s, precUnary
		}
		return s, precCall
	case *ir.StringConst:
		if !utf8.ValidString(e.Value) {
			// The lexer accepts only UTF-8 source, and no operation makes
			// a string invalid.
			panic(fmt.Sprintf("jsgen: internal compiler error: string at %s is not valid UTF-8", e.Span.Start))
		}
		return quote(e.Value), precCall
	case *ir.BoolConst:
		return strconv.FormatBool(e.Value), precCall
	case *ir.ArrayLit:
		parts := make([]string, len(e.Elems))
		for i, x := range e.Elems {
			parts[i] = g.expr(x, precAssign)
		}
		return "[" + strings.Join(parts, ", ") + "]", precCall
	case *ir.Index:
		return fmt.Sprintf("%s(%s, %s, %d)", g.use("aget"), g.expr(e.X, precAssign), g.expr(e.I, precAssign), g.site(e.Site)), precCall
	case *ir.IndexOld:
		return "$o", precCall
	case *ir.ArrayLen:
		return g.expr(e.X, precCall) + ".length", precCall
	case *ir.ArrayPush:
		return g.expr(e.X, precCall) + ".push(" + g.expr(e.V, precAssign) + ")", precCall
	case *ir.ArrayPop:
		return fmt.Sprintf("%s(%s, %d)", g.use("apop"), g.expr(e.X, precAssign), g.site(e.Site)), precCall
	case *ir.CharConst:
		// A char is its code point; the comment shows the character.
		return fmt.Sprintf("%d /* %s */", e.Value, charComment(e.Value)), precCall
	case *ir.FloatConst:
		var s string
		switch {
		case math.IsNaN(e.Value):
			return "NaN", precCall
		case math.IsInf(e.Value, 1):
			return "Infinity", precCall
		case math.IsInf(e.Value, -1):
			s = "-Infinity"
		default:
			s = strconv.FormatFloat(e.Value, 'g', -1, 64) // exact: the shortest decimal that reads back
		}
		if strings.HasPrefix(s, "-") {
			return s, precUnary
		}
		return s, precCall
	case *ir.Neg:
		x := g.expr(e.X, precUnary)
		if strings.HasPrefix(x, "-") {
			x = "(" + x + ")"
		}
		return "-" + x, precUnary
	case *ir.Convert:
		return fmt.Sprintf("%s(%s, %s, %s, %s, %d)", g.use("cast"), g.expr(e.X, precAssign),
			quote(kindCode(e.X.Type())), quote(kindCode(e.To)), quote(e.To.Name), g.site(e.Site)), precCall
	case *ir.NewError:
		msg, id := `""`, "0"
		if e.Message != nil {
			msg = g.expr(e.Message, precAssign)
		}
		if e.ID != nil {
			id = g.expr(e.ID, precAssign)
		}
		return fmt.Sprintf("%s(%s, %s, %s)", g.use("newError"), quote(e.T.Name), msg, id), precCall
	case *ir.FloatTest:
		// Through the runtime: a user function may be called isNaN.
		return g.use(e.Name) + "(" + g.expr(e.X, precAssign) + ")", precCall
	case *ir.Field:
		return g.expr(e.X, precCall) + "." + e.Name, precCall
	case *ir.Load:
		return jsName(e.Var.Name), precCall
	case *ir.FuncRef:
		return g.funcName(e.Func), precCall
	case *ir.Lambda:
		return g.lambda(e.Func), precAssign
	case *ir.Arith:
		if e.T.Kind == types.Float {
			// JavaScript's own IEEE arithmetic; float32 rounds the result.
			op := g.expr(e.X, precUnary) + " " + e.Op + " " + g.expr(e.Y, precUnary)
			if e.T.Bits == 32 {
				return "Math.fround(" + op + ")", precCall
			}
			return "(" + op + ")", precCall
		}
		if e.T.Bits != 0 { // an explicit width: numbers, or BigInts for 64 bits
			h := "arN"
			if types.IsWide(e.T) {
				h = "arW"
			}
			return fmt.Sprintf("%s(%s, %s, %s, %s, %d)", g.use(h), quote(e.Op), g.expr(e.X, precAssign), g.expr(e.Y, precAssign),
				quote(kindCode(e.T)), g.site(e.Site)), precCall
		}
		suffix := "I"
		if e.T.Kind == types.Unsigned {
			suffix = "U"
		}
		h := g.use(arithHelpers[e.Op] + suffix)
		return fmt.Sprintf("%s(%s, %s, %d)", h, g.expr(e.X, precAssign), g.expr(e.Y, precAssign), g.site(e.Site)), precCall
	case *ir.Bits:
		// & | ^ on numbers work on 32-bit two's complement, which keeps
		// signed values of every width in range; unsigned 32-bit results
		// need >>> 0. BigInts (64-bit) are exact.
		s := g.expr(e.X, precUnary) + " " + e.Op + " " + g.expr(e.Y, precUnary)
		if e.T.Kind == types.Unsigned && types.IntBits(e.T) == 32 {
			return "((" + s + ") >>> 0)", precCall
		}
		return "(" + s + ")", precCall
	case *ir.Complement:
		t := e.X.Type()
		x := "~" + g.expr(e.X, precUnary)
		switch bits := types.IntBits(t); {
		case t.Kind == types.Signed:
			return x, precUnary
		case bits == 64:
			return "(" + x + " & 0xFFFFFFFFFFFFFFFFn)", precCall
		case bits == 32:
			return "(" + x + " >>> 0)", precCall
		default:
			return fmt.Sprintf("(%s & 0x%X)", x, 1<<bits-1), precCall
		}
	case *ir.Shift:
		return fmt.Sprintf("%s(%s, %s, %s, %s, %d)", g.use("shift"), quote(e.Op), g.expr(e.X, precAssign), g.expr(e.Count, precAssign),
			quote(kindCode(e.X.Type())), g.site(e.Site)), precCall
	case *ir.Concat:
		// Every part is a string, so JavaScript's + concatenates.
		parts := make([]string, len(e.Parts))
		for i, p := range e.Parts {
			parts[i] = g.expr(p, precUnary)
		}
		if len(parts) == 1 {
			return parts[0], precUnary
		}
		return "(" + strings.Join(parts, " + ") + ")", precCall
	case *ir.Compare:
		op := compareOps[e.Op]
		// Both sides use op.prec+1, so a chain such as a == b == c keeps
		// explicit parentheses.
		return g.expr(e.X, op.prec+1) + " " + op.js + " " + g.expr(e.Y, op.prec+1), op.prec
	case *ir.Logic:
		prec := precAnd
		if e.Op == "||" {
			prec = precOr
		}
		return g.expr(e.X, prec) + " " + e.Op + " " + g.expr(e.Y, prec+1), prec
	case *ir.Not:
		return "!" + g.expr(e.X, precUnary), precUnary
	case *ir.CallFunc:
		return g.funcName(e.Func) + g.args(e.Args, e.Site), precCall
	case *ir.CallValue:
		return g.expr(e.Callee, precCall) + g.args(e.Args, e.Site), precCall
	case *ir.CallNative:
		g.native(e.Func.Name)
		result := "null"
		switch e.Func.Result.Kind {
		case types.Signed:
			result = `"int"`
		case types.Unsigned:
			result = `"uint"`
		}
		parts := []string{strconv.Itoa(g.site(e.Site)), quote(e.Pkg + "." + e.Func.Name), result,
			g.use("std") + "." + e.Func.Package + "." + e.Func.Name}
		site := g.site(e.Site)
		for _, a := range e.Args {
			if a.Type().Kind == types.Func {
				// A RuneScript function value takes a trailing call site;
				// the library calls it with the character only.
				parts = append(parts, fmt.Sprintf("($c) => %s($c, %d)", g.expr(a, precCall), site))
				continue
			}
			parts = append(parts, g.expr(a, precAssign))
		}
		return g.use("native") + "(" + strings.Join(parts, ", ") + ")", precCall
	case *ir.CallExtern:
		x := e.Extern
		kind := "v"
		switch r := x.Type.Result; r.Kind {
		case types.Void:
		case types.Error:
			kind = r.Name
		default:
			kind = kindCode(r)
		}
		parts := []string{strconv.Itoa(g.site(e.Site)), quote(x.Name), quote(x.Module), quote(kind), quote(x.Type.Result.Name), g.externs[x]}
		for _, a := range e.Args {
			parts = append(parts, g.expr(a, precAssign))
		}
		return g.use("ext") + "(" + strings.Join(parts, ", ") + ")", precCall
	case *ir.Template:
		return g.template(e), precCall
	case *ir.NewSignal:
		return g.use("signal") + "(" + g.expr(e.Init, precAssign) + ")", precCall
	case *ir.SignalGet:
		return g.expr(e.X, precCall) + ".get()", precCall
	case *ir.SignalSet:
		return g.expr(e.X, precCall) + ".set(" + g.expr(e.Value, precAssign) + ")", precCall
	case *ir.RenderApp:
		return g.use("renderApp") + "(" + g.expr(e.Component, precAssign) + ", " + g.expr(e.Selector, precAssign) + ", " +
			strconv.Itoa(g.site(e.Site)) + ")", precCall
	case *ir.Trace:
		parts := make([]string, len(e.Args))
		for i, a := range e.Args {
			parts[i] = g.expr(a, precAssign)
			if t := a.Type(); t.Kind == types.Float && t.Bits == 32 {
				parts[i] = g.use("f32str") + "(" + parts[i] + ")"
			} else if t.Kind == types.Char {
				parts[i] = "String.fromCodePoint(" + g.expr(a, precAssign) + ")"
			}
		}
		return g.use("trace") + "(" + strings.Join(parts, ", ") + ")", precCall
	}
	panic(fmt.Sprintf("jsgen: unexpected expression %T", e))
}

// args renders a RuneScript call's argument list: the arguments, then the
// call site.
func (g *gen) args(args []ir.Expr, site ir.Span) string {
	parts := make([]string, 0, len(args)+1)
	for _, a := range args {
		parts = append(parts, g.expr(a, precAssign))
	}
	parts = append(parts, strconv.Itoa(g.site(site)))
	return "(" + strings.Join(parts, ", ") + ")"
}

// native records which Unicode tables a standard library call needs.
func (g *gen) native(name string) {
	switch name {
	case "toUpper":
		g.tables["upper"] = true
	case "toLower":
		g.tables["lower"] = true
	case "toTitle":
		g.tables["title"] = true
	case "equalFold":
		g.tables["fold"] = true
	}
}

// lambda renders an arrow function. Its body is emitted at the current
// indentation, so the arrow spans several lines.
func (g *gen) lambda(f *ir.Func) string {
	saved := g.b
	g.b = strings.Builder{}
	g.indent++
	g.funcBody(f)
	g.indent--
	body := g.b.String()
	g.b = saved
	return fmt.Sprintf("(%s) => {\n%s%s}", params(f), body, strings.Repeat("  ", g.indent))
}

// ---- names and literals ----

// reserved lists JavaScript reserved words and the names strict mode does
// not allow as bindings. RuneScript identifiers cannot contain '$', so
// appending one never collides with another RuneScript name or with the
// generator's own '$'-prefixed names.
var reserved = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`await break case catch class const continue
		debugger default delete do else enum export extends false finally for
		function if implements import in instanceof interface let new null
		package private protected public return static super switch this throw
		true try typeof var void while with yield arguments eval undefined NaN
		Infinity`) {
		reserved[w] = true
	}
}

// kindCode names a conversion operand or destination for the runtime's
// cast: i(nt), u(int), f(loat or float64), f32, s(tring), or b(ool).
func kindCode(t *types.Type) string {
	switch t.Kind {
	case types.Signed:
		if t.Bits != 0 {
			return "i" + strconv.Itoa(t.Bits)
		}
		return "i"
	case types.Unsigned:
		if t.Bits != 0 {
			return "u" + strconv.Itoa(t.Bits)
		}
		return "u"
	case types.Float:
		if t.Bits == 32 {
			return "f32"
		}
		return "f"
	case types.String:
		return "s"
	case types.Bool:
		return "b"
	case types.Char:
		return "c"
	}
	panic("jsgen: conversion of " + t.Name)
}

// charComment renders a char constant for a comment: the character
// itself if it is printable and cannot end the comment, else U+XXXX.
func charComment(r rune) string {
	if unicode.IsPrint(r) && r != '*' && r != '/' {
		return "'" + string(r) + "'"
	}
	return fmt.Sprintf("U+%04X", r)
}

// entryFile is the name of the file declaring main.
func entryFile(p *ir.Program) string {
	if f := p.EntrySite.Start.File; f != nil {
		return f.Name
	}
	return "<input>"
}

func jsName(name string) string {
	if reserved[name] {
		return name + "$"
	}
	return name
}

// quote renders s as a JavaScript string literal. s is valid UTF-8, and
// non-ASCII characters are written as themselves.
func quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7F || r == 0x2028 || r == 0x2029 || r == utf8.RuneError:
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// externNames decides how each extern function is reached: a function of
// an imported module ($m0["readFileSync"]), or of the adapter object of a
// host-neutral module ($r0.readFile), whose functions use the host's own
// module.
func (g *gen) externNames(externs []*ir.Extern) {
	g.externs = map[*ir.Extern]string{}
	nss := map[string]string{}
	importNS := func(module, prefix string) string {
		if ns, ok := nss[module]; ok {
			return ns
		}
		ns := prefix + strconv.Itoa(len(g.imports))
		nss[module] = ns
		g.imports = append(g.imports, jsImport{module: module, ns: ns})
		return ns
	}
	adapters := map[*jshost.Module]*adapter{}
	for _, x := range externs {
		if x.Host == nil {
			g.externs[x] = importNS(x.Module, "$m") + "[" + quote(x.Name) + "]"
			continue
		}
		m := jshost.Lookup(x.Module)
		a := adapters[m]
		if a == nil {
			a = &adapter{module: m, name: "$r" + strconv.Itoa(len(g.adapters)), host: importNS(m.Node, "$h")}
			adapters[m] = a
			g.adapters = append(g.adapters, a)
		}
		a.funcs = append(a.funcs, x.Host)
		g.externs[x] = a.name + "." + x.Host.Name
	}
}

// ---- templates (rune:dom) ----

// template emits an expression that builds a template's element. Static
// parts are created once; every {{ }}, bound attribute, and $if becomes a
// binding (see the runtime) that updates only its own part of the DOM.
func (g *gen) template(t *ir.Template) string {
	var b strings.Builder
	pad := strings.Repeat("  ", g.indent)
	b.WriteString("(() => {\n")
	root := g.telement(&b, t.Root, "", pad+"  ")
	fmt.Fprintf(&b, "%s  return %s;\n%s})()", pad, root, pad)
	return b.String()
}

func (g *gen) tmp() string {
	g.tmps++
	return "$n" + strconv.Itoa(g.tmps)
}

// telement emits the creation of element e, appended to parent unless
// parent is "", and returns the variable holding it.
func (g *gen) telement(b *strings.Builder, e *ir.TElement, parent, pad string) string {
	v := g.tmp()
	fmt.Fprintf(b, "%sconst %s = %s(%t, %s);\n", pad, v, g.use("el"), e.SVG, quote(e.Tag))
	for _, a := range e.Attrs {
		site := strconv.Itoa(g.site(a.Site))
		switch {
		case a.Mode == ir.AttrEvent:
			fmt.Fprintf(b, "%s%s(%s, %s, %s, %s);\n", pad, g.use("on"), v, quote(strings.TrimPrefix(a.Name, "on")), g.expr(a.Value, precAssign), site)
		case a.Value != nil:
			values := "null"
			if a.Values != nil {
				q := make([]string, len(a.Values))
				for i, s := range a.Values {
					q[i] = quote(s)
				}
				values = "[" + strings.Join(q, ", ") + "]"
			}
			fmt.Fprintf(b, "%s%s(%s, %s, () => %s, %s, %s, %s);\n", pad, g.use("bindAttr"), v, quote(a.Name),
				g.expr(a.Value, precAssign), quote(attrKind(a)), values, site)
		case a.Present:
			fmt.Fprintf(b, "%s%s(%s, %s, true);\n", pad, g.use("attr"), v, quote(a.Name))
		default:
			fmt.Fprintf(b, "%s%s(%s, %s, %s);\n", pad, g.use("attr"), v, quote(a.Name), quote(a.Static))
		}
	}
	g.tnodes(b, e.Children, v, pad)
	if parent != "" {
		fmt.Fprintf(b, "%s%s.appendChild(%s);\n", pad, parent, v)
	}
	return v
}

// attrKind tells the runtime how to show a bound attribute's value.
func attrKind(a *ir.TAttr) string {
	switch {
	case a.Mode == ir.AttrBool:
		return "bool"
	case a.Mode == ir.AttrURL:
		return "url"
	case a.Value.Type().Kind == types.Float && a.Value.Type().Bits == 32:
		return "f32"
	}
	return ""
}

func (g *gen) tnodes(b *strings.Builder, nodes []ir.TNode, parent, pad string) {
	for _, n := range nodes {
		switch n := n.(type) {
		case *ir.TElement:
			g.telement(b, n, parent, pad)
		case *ir.TText:
			fmt.Fprintf(b, "%s%s(%s, %s);\n", pad, g.use("text"), parent, quote(n.Text))
		case *ir.TInterp:
			x := g.expr(n.X, precAssign)
			switch t := n.X.Type(); {
			case t.Kind == types.Element:
				fmt.Fprintf(b, "%s%s(%s, () => %s);\n", pad, g.use("bindNode"), parent, x)
			case t.Kind == types.Float && t.Bits == 32:
				fmt.Fprintf(b, "%s%s(%s, () => %s, \"f32\");\n", pad, g.use("bindText"), parent, x)
			default:
				fmt.Fprintf(b, "%s%s(%s, () => %s, \"\");\n", pad, g.use("bindText"), parent, x)
			}
		case *ir.TIf:
			// test returns the index of the branch to show, or -1.
			test := "-1"
			if n.Else != nil {
				test = strconv.Itoa(len(n.Conds))
			}
			for i := len(n.Conds) - 1; i >= 0; i-- {
				test = g.expr(n.Conds[i], precOr) + " ? " + strconv.Itoa(i) + " : " + test
			}
			fmt.Fprintf(b, "%s%s(%s, () => (%s), (i, $p) => {\n", pad, g.use("when"), parent, test)
			branches := append(append([][]ir.TNode{}, n.Branches...), n.Else)
			for i, br := range branches {
				if len(br) == 0 {
					continue
				}
				fmt.Fprintf(b, "%s  if (i === %d) {\n", pad, i)
				g.tnodes(b, br, "$p", pad+"    ")
				fmt.Fprintf(b, "%s  }\n", pad)
			}
			fmt.Fprintf(b, "%s});\n", pad)
		}
	}
}
