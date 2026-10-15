# RuneScript Roadmap

Milestones are small and independently testable. A feature counts as done
only when tests show it working (`go test ./...`). A milestone that depends
on an undecided language rule waits until the owner decides it.

## v0.1: Core language ✅ (current)

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
- [x] CLI: `runescript run|check|version`
- [x] Examples (`examples/`), with golden-output tests
- [x] Logo: `assets/runescript.svg` and an icon variant, rasterized to PNG via headless Edge (`scripts/render-logo.sh`)
- [x] Docs: specification, user guide, roadmap

## v0.2: Settle the v0.1 provisional rules

No new syntax until these are decided. Each decision gets an update to the
spec and the tests.

- [ ] Final source file extension (`.rs` conflicts with Rust)
- [ ] Final diagnostic prefix (currently `ASX`)
- [ ] Integer widening model (variable storage vs. expression results), overflow behaviour, and the width of `uint`
- [ ] Explicit checked conversion syntax, then implement it so the `ASX-TYPE-004` help text has something to point to
- [ ] Unsigned → signed and mixed-signedness rules
- [ ] Entry-point rules and visibility semantics for top-level functions
- [ ] Confirm or relax the other v0.1 restrictions: required initializers, read-only parameters, no shadowing, mandatory braces

## v0.3: `use` (packages and files)

Syntax decided by the owner: `use "<package_name>";` and `use "<file_path>";`,
with paths relative to the current working directory.

- [ ] Decide how a package name is told apart from a file path, how used
  names are qualified, and what visibility exposes
- [ ] Multi-file programs: load, parse, and check every used file once; detect cycles
- [ ] A native-package mechanism: Go-implemented functions registered under a package name
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

## Later

- [ ] Standard library modules (`stdlib/`)
- [ ] Better parser error recovery (report more than one syntax error)
- [ ] Performance work (bytecode/VM) only after semantics are stable and measured
