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
  `string`/`int`/`bool`, camelCase names, byte offsets (now 43, with `char`,
  callbacks, and arrays)
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

### Settle v0.1 provisional rules ✅ (owner decisions, 2026-10-17)
- [x] Final source file extension: `.rsc` (`.rs` was Rust's)
- [x] Final diagnostic prefix: `ASX` stays
- [x] Integer widening model: intermediate results widen, variables keep their type (option A, done in v0.5); `uint` stays 32 bits
- [x] Explicit checked conversion syntax: C-style `(T) x` (v0.3)
- [x] Unsigned → signed and mixed signedness: always an explicit, checked conversion; mixed arithmetic stays an error
- [x] Entry point: exactly `protected int main()` (`ASX-ENTRY-002` otherwise); arguments come from `std:process`
- [x] Relaxed: `T x;` takes the zero value (`ASX-TYPE-028` for types without one); shadowing is allowed in nested blocks, with `ASX-WARN-009`. Kept: read-only parameters, mandatory braces

### JavaScript target ✅
- [x] Typed IR (`internal/ir`) lowered from the checked AST, with an independent verifier
- [x] `internal/jsgen`: ES module output with checked integer arithmetic, call-depth limit, and located runtime errors
- [x] `std:string` in JavaScript with UTF-8 byte offsets and Go-derived Unicode tables
- [x] `runescript build [--target js] [-o out.js]`; modules run with `node` or `import { run }`
- [x] Differential tests (interpreter vs. Node.js) for every example and runtime program, and randomized `std:string` conformance
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
- [x] Arguments and environment: `std:process` on both targets instead of a JavaScript-only `rune:process` (2026-10-17)
- [ ] Objects, arrays, and callbacks across the boundary (classes and arrays exist now; design the crossing first)
- [x] Source maps: `runescript build --source-map` writes `<out>.map` (statements and functions mapped, sources included); `runescript dev` inlines one (2026-11-01)

### DOM interop (owner decisions, 2026-10-16; started)
- [x] Internal list of HTML4 and HTML5 elements (`internal/dom`): void and obsolete elements with hints, typed attributes per element (string, url, bool, int, uint, float, keywords, id references, tokens, events), global attributes and event handlers, `data-*`/`aria-*`, and obsolete HTML4 presentational attributes; cross-checked against independent element lists
- [x] SVG 1.1 and SVG 2 elements in their own namespace (case-sensitive names; SVG 1.1 elements SVG 2 removed are obsolete with hints), typed attributes including lengths, presentation attributes with CSS-wide keywords, `href` with obsolete `xlink:href`
- [x] Declarative, component-based UI like React's, but without JSX and without a virtual DOM: state changes update the affected DOM nodes directly
- [x] Templates in `.rsc` files, in the style of Vue/Svelte components and Laravel Blade: `$if`/`$elseif`/`$else`/`$endif`, `{{ }}` interpolation that escapes HTML by default; `Element` is a type imported with `use "rune:dom";` (2026-10-16)
- [x] `$foreach (T x in a) … $endforeach`, header checked like `for each`; the list is rebuilt when a rune its array expression read changes (2026-11-01)
- [ ] Keyed `$foreach` updates (rebuild only the items that changed)
- [x] Reactivity through explicit SolidJS-style signals, called runes: `Signal<int> count = signal(0);`, read `count()` or `count.get()`, written `count.set(n)`; setting one updates only the DOM nodes that read it; nested functions and lambdas may use the runes around them
- [x] Elements: HTML4 (obsolete ones warn), HTML5, and SVG; attributes typed per element, checked for static values and `{{ }}` types; bound keyword and URL values checked at run time
- [x] Mounting without a router: `use "./App.rsc";` then `dom.renderApp(app.App, "#app");` renders one component into the element the selector names; any CSS selector works, with a warning unless it is an id selector
- [x] Templates switch to SVG inside `<svg>` (and back to HTML inside `foreignObject`)
- [x] Tests: lexer, parser, checker, and fake-DOM runs on Node.js of test programs and every web example; counter and SVG checked in Edge
- [x] Docs: guide chapter 15 (and translations), diagnostics, spec section 13, web examples
- [ ] Component tags with properties, event objects (input values), two-way binding
- [x] A built-in `Signal<T>` for scalar `T` (general generics still open); callbacks stay inside RuneScript, so no JavaScript callbacks were needed

### Live reload (owner request, 2026-10-16)
- [x] `runescript dev`: serves the entry file's folder on 127.0.0.1 (port 5173, or the next free one), compiles in memory, rebuilds when a program file changes, reloads the browser over Server-Sent Events, and shows compile errors in an overlay; the user's `index.html` or a generated page
- [x] `runescript run --watch`: runs a command-line program again whenever one of its files changes
- [x] Polling file watcher (`internal/watch`), standard library only; tests for the server, the watcher, and watch mode, and a check in Edge of reload and the overlay
- [ ] Hot module replacement for `rune:dom` components, keeping their runes' values (needs components)

### Bundled JavaScript runtime (planned, not started)
- [ ] A small ES2015-compatible JavaScript engine shipped with RuneScript, so compiled programs run without Node.js or Bun. It supports the extras above (`BigInt`, `??`, modules)
- [ ] Conformance: its output is compared against Node.js (ES2022), the reference host

## v0.3: `use` for packages and files ✅

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
- [x] Qualified function values (`geom.area` without a call), 2026-10-16

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
- [x] `std:math` (2026-10-17)

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
  working directory, if wanted (optional; moved to Later)

## v0.4: Strings, output, and the system ✅

- [x] String concatenation `+` (strings only) and interpolation `$"…{x}…"` (2026-10-16)
- [x] `string.len`, `string.slice`, `string.runeCount` in bytes; `char` as a Unicode scalar with `'a'` literals (2026-10-16)
- [x] `std:string` functions that take `char` and callbacks: `containsRune`, `indexRune`, `indexFunc`, `trimFunc`, `map`, … (2026-10-16)
- [x] Escape sequences: `\u{...}` Unicode escapes and `\0` (2026-10-17)
- [x] Colored diagnostics: errors red, warnings yellow, `--color=auto|always|never`, `NO_COLOR` (2026-10-17)
- [x] `std:io`: `print` without a newline, `error`/`warn` lines on standard error (red/yellow on a terminal), standard input by line, byte, or all at once (2026-10-17)
- [x] `std:process`: `args`, `env`, `hasEnv`, `cwd`, `platform`; `runescript run prog.rsc args…` and `node prog.js args…` (2026-10-17)
- [x] `std:fs` on both targets: files, directories (`mkdir`, `mkdirAll`, `readDir`, `walk`, `remove`, `removeAll`), `copyFile`, `rename`, `size`, `tempDir`, with the same error messages everywhere (2026-10-17)
- [x] `std:path`: Node.js's path semantics, implemented rather than wrapped, with platform, `posix…`, and `win32…` versions; tested against Node.js (2026-10-17)
- [x] Example: a Brainfuck interpreter with 32,768 cells (`examples/brainfuck.rsc`) (2026-10-17)
- [x] Binary files: `fs.readBytes`, `fs.writeBytes`, `fs.appendBytes` with `Array<uint8>`, and `string.bytes`/`string.fromBytes` (2026-10-20)

## v0.5: Numbers ✅

- [x] Explicit-width integers (`int8`…`int64`, `uint8`…`uint64`), checked in every width; BigInt on the JavaScript target (2026-10-17)
- [x] Bitwise operators and shifts, with checked shift counts and trapping `<<` (2026-10-17)
- [x] Widening, option A: exact intermediate results within int64, stored values checked; boundary, non-contraction, and randomized cross-target tests (2026-10-17)
- [x] `float`, `float32`, `float64` (v0.3); widening is exact, so
  representation only matters once floats are stored in arrays/structs
- [x] Literal forms: hex `0xFF`, binary `0b1010`, octal `0o17`, `_` separators (2026-10-17)
- [x] `std:math`: Go's `math` on float64, int helpers, constants; transcendental functions from one shared implementation (`internal/fmath`), bit-identical on both targets (2026-10-17)

## v0.6: Aggregates (current)

- [x] `Array<T>`: literals, bounds-checked indexing, `length`, `push`, `pop`, reference semantics (2026-10-16, details provisional)
- [x] `for each (T x in a)` loops, nested arrays (`Array<Array<T>>`), and the methods `insert`, `removeAt`, `slice`, `indexOf`, `contains`, `reverse`, `sort()`, `sort(less)` (stable, same comparator calls on both targets) (2026-10-20)
- [x] `std:fs` binary files (`Array<uint8>`), see v0.4 (2026-10-20)
- [x] `std:csv`: parse and format `Array<Array<string>>` (RFC 4180, any separator, errors with line and column) (2026-10-26)
- [x] `std:json`: an opaque `json.Value` (the first opaque std type) read with checked functions, built with `from…` functions, and written with `json.stringify`/`stringifyIndent`; one parser on both targets, tested against random valid and broken texts (2026-10-20)
- [x] `std:matrix`: functions over `Array<Array<float64>>` (owner chose a library, no new type): products, transpose, determinant, inverse, solve, bit-identical on both targets (2026-10-26)
- [x] `struct`: value semantics (copy on assignment), named-field literals, nested field writes, `==` by field (2026-10-17)
- [ ] Struct methods and constructors

### Next, in order (owner, 2026-10-22)

1. `std:csv` (above, done 2026-10-26)
2. `std:matrix` (above, done 2026-10-26)
3. `std:semver` (done 2026-10-26)
   - [x] `compare`, `gt`, `lt`, `eq`, `satisfies` (and `neq`, `gte`, `lte`, `valid`, `validRange`, `major`, `minor`, `patch`, `prerelease`, `build`) on version strings, following SemVer 2.0.0 precedence and npm-style ranges (`^1.2.0`, `~1.2.0`, `>=1.2.0 <2.0.0`), checked against npm's semver package
   - [x] An invalid version string throws a catchable error (`ArgumentError`)
4. Module versions
   - [ ] Every `.rsm` file must declare `public semver version` or `public string version` (a compile error otherwise); `public string author` and `public string license` are optional
   - [ ] The metadata goes at the top of the file, after the `use` lines
   - [ ] The `semver` type: version literals (`1.2.3-rc.1+build.5`) and the builder form `semver(1, 2, 3, "rc", 1)`; usable only for `version`
   - [ ] Programs read a version as a `string`: `geom.version`, or `geom.version['a']` (the file name without `.rsm`) when files of a package declare different versions
   - [ ] Versions for the `std:` modules, once the standard library is finished
5. Performance
   - [ ] Profile the interpreter on the `bench/` suite
   - [ ] Cheap interpreter fixes: numbered variable slots instead of a map per call, operator codes instead of strings, no `defer` per call
   - [ ] Compile the verified IR to typed Go closures as the `run` engine; keep the tree-walking interpreter as a reference for agreement tests
   - Pure Go only: no `unsafe` and no hand-written machine code

### Editor support (owner request, 2026-10-26)

- [x] Language server `runescript lsp` (`internal/lsp`): diagnostics as you type (unsaved buffers included), hover, go to definition, completion, and document symbols
- [x] VS Code extension (`editors/vscode`): TextMate grammar, language configuration, language client
- [x] Vim and Neovim plugin (`editors/vim`): syntax, indentation, file type, Neovim 0.11 LSP configuration
- [x] Emacs major mode (`editors/emacs`): highlighting, indentation, Eglot and lsp-mode registration (not yet tested in Emacs)
- [ ] Rename, find references, signature help, formatting (needs a formatter), code actions
- [ ] Publish the VS Code extension (Marketplace, Open VSX) and the Emacs mode (MELPA)

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

- [x] Simple classes (owner decisions, 2026-11-01): `public class Name { … }` with fields, methods, and one constructor `public Name(…) { … }`; objects made with `new` (`new pkg.Name(…)`), shared by reference, `==` by identity, traced as `[object Name]`; `this` required; no nested classes. Both targets, guide chapter 24 (2026-11-01)
- [x] Visibility enforcement on members: `public` / `protected` (package) / `private` (file), as at the top level (2026-11-01)
- [x] No null: fields without a zero value need an initializer or a first value at the start of the constructor (2026-11-01, provisional)
- [ ] Inheritance / interfaces (design first)
- [ ] `static` members, methods as values, getters and setters
- [ ] Objects across the JavaScript boundary
- [x] `std:string` additions that need arrays: `split`, `splitN`, `splitAfter`, `splitAfterN`, `fields`, `fieldsFunc`, `join` (2026-10-16)
- [ ] `std:string` additions that need multiple returns or classes (`cut`, builder/replacer)

## v0.9: `std:symrs` (symbolic mathematics)

A built-in port of the owner's SympJS library: symbols, expression trees,
simplification, differentiation, integration, equation solving, matrices,
complex numbers, and Taylor and Fourier series.

- [ ] Design the RuneScript API (expression values with chained methods need classes or an opaque `Expr` type)
- [ ] Port in layers: expressions and simplification → `diff`/`int` → solving → trig → complex → matrices → Taylor/Fourier
- [ ] Output as text and LaTeX instead of SympJS's HTML renderer (there's no browser)

Depends on: floats (v0.5), arrays (v0.6), classes or opaque native types (v0.8), function values (v0.2).

## Later

- [ ] A `runescript` flag to resolve uses from a directory other than the working directory, if wanted
- [ ] Better parser error recovery (report more than one syntax error)
- [ ] Further performance work (bytecode VM, a native-binary build mode) only if the closure compiler is measured and found not fast enough
