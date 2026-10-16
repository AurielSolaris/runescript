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

## v0.2: Standard library and functions

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
- [x] Enforce `private` (file) and `protected` (package) once multi-file programs exist (done in v0.3)

### Tests and examples ✅
- [x] Fuzz tests for the parser/checker and the interpreter (no crashes on arbitrary input)
- [x] Golden error programs for functions and `use`; runtime-error golden programs with expected partial output
- [x] Diagnostics are reported in source order
- [x] Examples: `callbacks.rsc`, `primes.rsc`, `strings.rsc`, `boundary_error.rsc`, plus `examples/README.md`

### Settle v0.1 provisional rules
- [x] Final source file extension: `.rsc` (`.rs` was Rust's)
- [ ] Final diagnostic prefix (currently `ASX`)
- [ ] Integer widening model, overflow behaviour, and the width of `uint`
- [x] Explicit checked conversion syntax: C-style `(T) x` (v0.3)
- [ ] Unsigned → signed and mixed-signedness rules
- [ ] Entry-point rules and visibility semantics for top-level functions
- [ ] Confirm or relax: required initializers, read-only parameters, no shadowing, mandatory braces

### JavaScript target ✅
- [x] Typed IR (`internal/ir`) lowered from the checked AST, with an independent verifier
- [x] `internal/jsgen`: ES module output with checked integer arithmetic, call-depth limit, and located runtime errors
- [x] `std:string` in JavaScript with UTF-8 byte offsets and Go-derived Unicode tables
- [x] `runescript build [--target js] [-o out.js]`; modules run with `node` or `import { run }`
- [x] Differential tests (interpreter vs. Node.js) for every example and runtime program, and randomized `std:string` conformance
- [ ] Source maps
- [ ] JavaScript interop: moved to "JavaScript target, part 2" below
- [ ] Measure generated-code performance before optimizing (e.g. inlining range checks)

### Warnings ✅ (requested by the owner, 2026-10-16)
- [x] Warning severity in diagnostics (`warning ASX-WARN-…`), reported only for programs without errors
- [x] `-Werror` and `--allow <codes>` on `run`, `check`, and `build`
- [x] Unused locals and parameters (`ASX-WARN-001`/`002`), unused `use` (`003`), unreachable code (`004`), discarded results (`005`)
- [x] Docs (guide chapter 8, in all translations) and tests

### JavaScript target, part 2 (owner decisions, 2026-10-16)
- [x] Generated code limited to ES2015 plus `BigInt`, `??`, and ES modules (`import.meta` included); `tests/jslevel_test.go` lints every generated program and the runtime (removed `??=` and BigInt `**`)
- [x] No `process` dependence beyond running the module directly: `process` is only used under `import.meta.main`
- [x] `extern "<module>" Type name(Type p, …);` declarations for JavaScript functions (no `use "js:…"`, owner decision)
- [x] Host-neutral `rune:fs` (readFile, writeFile, appendFile, exists, remove) mapped to `node:fs`; `node:`/`bun:`/`deno:` modules allowed with `ASX-WARN-006`
- [x] Boundary checks on every result (scalars and errors; float32 rounded): `ASX-RUN-007` `TypeError`; JavaScript exceptions are `ASX-RUN-008` `Error`; error values cross as the exported `RSError` class
- [x] The interpreter rejects programs with externs (`ASX-TARGET-002`); Node.js tests in `tests/testdata/extern`
- [ ] `rune:process` (arguments, environment, exit code)
- [ ] Objects, arrays, and callbacks across the boundary (need classes and arrays first)
- [ ] Source maps

### Bundled JavaScript runtime (planned, not started)
- [ ] A small ES2015-compatible JavaScript engine shipped with RuneScript, so compiled programs run without Node.js or Bun. It supports the extras above (`BigInt`, `??`, modules)
- [ ] Conformance: its output is compared against Node.js (ES2022), the reference host

## v0.3: `use` for packages and files (current)

Syntax decided by the owner: `use "<package_name>";` and `use "<file_path>";`,
with paths relative to the current working directory. (`std:` works since v0.2.)

- [x] Decisions: a path contains `/`; a package is the `.rsm` files of its
  directory; qualifier = last segment or `as` alias; a file use binds the
  file's package; cycles are errors; exactly one `main`, in any file; file
  extensions are a convention (any extension works when named directly)
- [x] `internal/loader`: load and parse every used file once; detect cycles
- [x] Checker: package scopes, per-file private scopes, qualified calls,
  `public`/`protected`/`private` enforced (`ASX-NAME-011`)
- [x] Diagnostics and runtime errors name the file; the JavaScript target
  bundles every file into one module
- [x] Tests: `tests/testdata/multi` (missing files, cycles, mismatched
  packages, clashes, visibility, duplicates, bad names, a runtime error in
  a used file, running from a different working directory), all also run on
  Node.js
- [ ] Qualified function values (`geom.area` without a call)

### Floats, conversions, and errors ✅ (requested by the owner during v0.3)
- [x] `float` (float32 by default, widening exactly, per CLAUDE.md §4.4),
  `float32`, `float64`; literals `64.3`, `1e10`; IEEE 754 (`NaN`,
  `Infinity`, `isNaN`, `isFinite`), as in ActionScript 3
- [x] Float formatting as in ActionScript 3, identical on both targets
  (`internal/conv`, tested against Node.js on random values)
- [x] Checked conversions `(T) x` between numbers, strings, and bools;
  compile-time checks for constants
- [x] `Error`, `RangeError`, `TypeError`, `ArgumentError`; `new`, fields
  `message`/`name`/`errorID`
- [x] `throw`, `try`/`catch`/`finally`, untyped `catch (e)`; `return`,
  `break`, and `continue` in `finally` as in AS3; runtime errors are
  catchable (except the call-depth limit)
- [x] Interpreter and JavaScript target behave identically (differential
  tests), and error unwinding stays linear in the call depth
- [ ] `std:math`

### Installation and source text ✅
- [x] `install.sh` (Linux, macOS) and `install.ps1` (Windows): check Go
  (found, ≥ 1.25, working), build the CLI, install it per user, add it to
  PATH and set `RUNESCRIPT_HOME`, verify with a test program; uninstall;
  tested in scratch locations
- [x] `runescript version` reports the commit or module version and the Go
  toolchain
- [x] Source encodings: UTF-16/UTF-32 (and a UTF-8 byte-order mark) are
  converted; other non-UTF-8 files are rejected (`ASX-LEX-007`) on every
  target
- [ ] `runescript` flag to resolve uses from a directory other than the
  working directory, if wanted

## v0.4: Strings and output

- [ ] String concatenation / interpolation (design first)
- [ ] `string` length and indexing; `char` design
- [ ] Escape sequences: Unicode escapes

## v0.5: Numbers

- [ ] Explicit-width integers (`int8`…`int64`, `uint8`…`uint64`)
- [ ] Widening as decided in v0.2, with boundary and non-contraction tests
- [x] `float`, `float32`, `float64` (v0.3); widening is exact, so
  representation only matters once floats are stored in arrays/structs
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
