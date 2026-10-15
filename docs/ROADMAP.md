# RuneScript Roadmap

Milestones are small and independently testable. A feature counts as done
only when tests show it working (`go test ./...`). A milestone that depends
on an undecided language rule waits until the owner decides it.

## v0.1: Core language ✅

Scope set by the owner: keep it very simple. Variables and assignment,
conditionals, loops, and basic functions.

- [x] Project layout: `cmd/`, `internal/{lexer,parser,ast,checker,types,runtime,diagnostics,driver}`, `tests/`
- [x] Lexer with source positions; comments; decimal integer, string, and bool literals
- [x] Diagnostics with code, `line:col`, source excerpt, caret, and help text
- [x] `package a.b;` declaration
- [x] Top-level functions: `[visibility] Type name(Type p, …) { … }`, calls in any order, recursion
- [x] Types `int`/`sint`, `uint`, `bool`, `string`, `void`
- [x] Variable declarations with initializers, `=`, `+= -= *= /= %=`, `++`/`--`
- [x] Arithmetic, comparison, and logical operators with C-style precedence; short-circuit `&&`/`||`
- [x] `if`/`else if`/`else`, `while`, `for`, `break`, `continue`, `return`
- [x] Static checking: undefined names, arity, assignability, definite return, block scoping
- [x] Signed → unsigned assignment rejected at compile time (`ASX-TYPE-004`), along with negative constants → `uint` and mixed signedness
- [x] Exact integer arithmetic; overflow and division by zero are located runtime errors (never silent wraparound)
- [x] `trace(...)`
- [x] Tree-walking interpreter; `main()` result becomes the exit code
- [x] CLI: `runescript run|check|version` (`build` added in v0.2)
- [x] Examples (`examples/`), with golden-output tests
- [x] Logo: `assets/runescript.svg` and an icon variant, rasterized to PNG via headless Edge (`scripts/render-logo.sh`)
- [x] Docs: specification, user guide, roadmap

## v0.2: Standard library and functions (current)

### Standard library, part 1 ✅
- [x] `use "std:<name>";` declarations and qualified calls (`string.toUpper(s)`)
- [x] Native-package registry in `stdlib/`: Go-implemented functions with checked RuneScript signatures
- [x] `std:string`: all 25 functions from Go's `strings` package expressible with
  `string`/`int`/`bool`, camelCase names, byte offsets
- [x] Docs (`docs/guide/10-std-string.md`) and an example (`examples/strings.rsc`)

### Functions ✅
- [x] Required access modifier on every function: `public|private|protected <type> <name>(…)`
- [x] Prototypes (forward declarations), with signature checking
- [x] Nested functions, including recursion and mutual recursion through prototypes
- [x] Function types `int(int, int)`, function values, callbacks, functions returning functions
- [x] Lambdas `(int a) => a + 1` and `(int a) => { … }`
- [x] Function boundaries: no capture of enclosing variables
- [x] Docs (guide chapter 6) and an example (`examples/callbacks.rsc`)
- [ ] Enforce `private` (file) and `protected` (package) once multi-file programs exist (v0.3)

### Tests and examples ✅
- [x] Fuzz tests for the parser/checker and the interpreter (no crashes on arbitrary input)
- [x] Golden error programs for functions and `use`; runtime-error golden programs with expected partial output
- [x] Diagnostics are reported in source order
- [x] Examples: `callbacks.rsc`, `primes.rsc`, `strings.rsc`, `boundary_error.rsc`, plus `examples/README.md`

### Settle v0.1 provisional rules
- [x] Final source file extension: `.rsc` (`.rs` was Rust's)
- [ ] Final diagnostic prefix (currently `ASX`)
- [ ] Integer widening model, overflow behaviour, and the width of `uint`
- [ ] Explicit checked conversion syntax
- [ ] Unsigned → signed and mixed-signedness rules
- [ ] Entry-point rules and visibility semantics for top-level functions
- [ ] Confirm or relax: required initializers, read-only parameters, no shadowing, mandatory braces

### JavaScript target ✅
- [x] Typed IR (`internal/ir`) lowered from the checked AST, with an independent verifier
- [x] `internal/jsgen`: ES2022 module output with checked integer arithmetic, call-depth limit, and located runtime errors
- [x] `std:string` in JavaScript with UTF-8 byte offsets and Go-derived Unicode tables
- [x] `runescript build [--target js] [-o out.js]`; modules run with `node` or `import { run }`
- [x] Differential tests (interpreter vs. Node.js) for every example and runtime program, and randomized `std:string` conformance
- [ ] Source maps
- [ ] JavaScript interop: `use "js:<module>";` with typed prototypes, checked at the boundary
- [ ] Measure generated-code performance before optimizing (e.g. inlining range checks)

## v0.3: `use` for packages and files

Syntax decided by the owner: `use "<package_name>";` and `use "<file_path>";`,
with paths relative to the current working directory. (`std:` works since v0.2.)

- [ ] Decide how a package name is told apart from a file path, aliases, and
  what visibility exposes
- [ ] Multi-file programs: load, parse, and check every used file once; detect cycles
- [ ] Tests: missing file, cycles, name clashes, running from a different working directory

## v0.4: Strings and output

- [ ] String concatenation / interpolation (design first)
- [ ] `string` length and indexing; `char` design
- [ ] Escape sequences: Unicode escapes

## v0.5: Numbers

- [ ] Explicit-width integers (`int8`…`int64`, `uint8`…`uint64`)
- [ ] Widening as decided in v0.2, with boundary and non-contraction tests
- [ ] `float`, `float32`, `float64`; precision-aware widening
- [ ] Additional literal forms (hex, separators) if approved
- [ ] `std:math`: Go's `math` package in full, once floats exist

## v0.6: Aggregates

- [ ] `array` / `Array<T>` (needs a generics syntax decision)
- [ ] `struct`: value semantics, copy rules
- [ ] `matrix` (design first)

## v0.7: Graphics package (native window, drawing, input, UI)

Games were AS3's main use, so native graphics are a first-class goal. These
are native-desktop graphics: no browser, no Flash Player.

- [ ] Confirm the backend (recommended: raylib via `raylib-go`)
- [ ] Design a small RuneScript-facing API (window, frame loop, shapes,
  text, textures, keyboard/mouse, audio later) instead of a 1:1 binding,
  so the backend can be replaced later
- [ ] Expose it as a standard package, e.g. `use "rune.gfx";` (name to be decided)
- [ ] Immediate-mode UI widgets (raygui) as a follow-up package
- [ ] Headless CI tests for the binding layer; example games in `examples/`

Depends on: `use` (v0.3), floats (v0.5), and structs for colours and vectors (v0.6).

## v0.8: Classes

- [ ] `class` declarations, fields, methods, constructors
- [ ] Visibility enforcement
- [ ] Inheritance / interfaces (design first)
- [ ] `std:string` additions that need arrays, multiple returns, or classes
  (`split`, `fields`, `join`, `cut`, builder/replacer)

## v0.9: `std:symrs` (symbolic mathematics)

A built-in port of the owner's SympJS library: symbols, expression trees,
simplification, differentiation, integration, equation solving, matrices,
complex numbers, and Taylor and Fourier series.

- [ ] Design the RuneScript API (expression values with chained methods need classes or an opaque `Expr` type)
- [ ] Port in layers: expressions and simplification → `diff`/`int` → solving → trig → complex → matrices → Taylor/Fourier
- [ ] Output as text and LaTeX instead of SympJS's HTML renderer (there's no browser)

Depends on: floats (v0.5), arrays (v0.6), classes or opaque native types (v0.8), function values (v0.2).

## Later

- [ ] Standard library modules (`stdlib/`)
- [ ] Better parser error recovery (report more than one syntax error)
- [ ] Performance work (bytecode/VM) only after semantics are stable and measured
