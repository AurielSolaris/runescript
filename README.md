<p align="center">
  <img src="assets/runescript.png" alt="RuneScript logo: a rune-styled snowflake with lightning accents" width="160">
</p>

<h1 align="center">RuneScript</h1>

<p align="center">A statically typed, general-purpose language inspired by ActionScript, implemented in Go.</p>

<p align="center"><strong>English</strong> · <a href="localization/bn/README.md" lang="bn">বাংলা</a> (Bengali) · <a href="localization/heb/README.md" lang="he">עברית</a> (Hebrew)</p>

---

RuneScript keeps ActionScript's identity (`package`, `trace`, `int`/`uint`,
`public`/`protected`/`private`) and changes the syntax to a C-family style:
types come before names, every function starts with an access modifier
(`public int add(int a, int b)`) instead of a `function` keyword, and
semicolons are required. Its type checking is strict: nothing is implicitly `any`, and
signed and unsigned integers never mix silently.

It runs as a standalone program: no browser, no HTML, no Flash Player.
It can also compile to a JavaScript module for the web, with the same
strict typing and the same runtime errors.
It is **not** source-compatible with ActionScript 3.

```actionscript
package main;

public void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

protected int main() {
    sayHi("Hello world", 3);
    return 0;
}
```

## Status: v0.6 in progress (early development)

**Works now**, and is covered by tests:

- Packages; functions declared `public|private|protected <type> <name>(…)`; recursion;
  the entry point `protected int main()`, whose result is the exit status
- Nested functions, prototypes, function types (`int(int, int)`), callbacks, and lambdas (`(int a) => a * 2`)
- `int`/`sint`, `uint`, `bool`, `string`, `void`
- Variables, `=`, `+= -= *= /= %=`, `++`/`--`; `int n;` starts at the zero
  value (there is no `null`); shadowing in nested blocks, with a warning
- Arithmetic, comparison, and logical operators
- `if`/`else if`/`else`, `while`, `for`, `break`, `continue`, `return`
- `trace(...)`
- Standard library: `use "std:string";` gives Go's `strings` functions as `string.toUpper(s)` and so on
- Multi-file programs: `use "geom";` loads the package in `geom/*.rsm`, and
  `use "./lib/x.rsc";` loads one file; `public`/`protected`/`private` are
  enforced ([guide](docs/guide/12-packages-and-files.md))
- `float` (float32 by default, widening without loss), `float32`, `float64`;
  IEEE 754 with `NaN` and `Infinity`, as in ActionScript
- Checked C-style conversions: `(int) 64.3`, `(float) "64.7"`, `(uint) n`
  ([guide](docs/guide/13-floats-and-conversions.md))
- Exact-width integers `int8` … `uint64` (64-bit values are BigInts in
  JavaScript), hex/binary/octal literals with `_` separators, bit
  operators, and `std:math`, whose results are bit-identical on both
  targets ([guide](docs/guide/18-numbers.md))
- Structs: value types built with `Point{x: 1, y: 2}`, copied on
  assignment, compared field by field ([guide](docs/guide/19-structs.md))
- Classes: fields, methods, one constructor, and `this`; objects are
  created with `new`, shared by reference, and compared by identity. No
  inheritance yet ([guide](docs/guide/24-classes.md))
- Text and lists for scripting: `+` and `$"{count} files"` for strings,
  `char`, and `Array<T>` with bounds-checked indexing, nested arrays,
  `for each (T x in a)` loops, slicing, searching, and stable sorting
  ([guide](docs/guide/16-text-and-arrays.md)); `std:string` covers Go's
  `strings` package, including `split`, `fields`, `join`, and callbacks
  ([guide](docs/guide/10-std-string.md))
- ActionScript-style errors: `throw new RangeError("…")`,
  `try`/`catch (Error e)`/`catch (e)`/`finally`; runtime errors are catchable
  ([guide](docs/guide/14-errors.md))
- Compile-time type checking with located diagnostics, including rejection
  of signed → unsigned assignment
- Integer overflow and division by zero are reported as runtime errors and
  never wrap silently; intermediate `int`/`uint` results widen exactly, so
  `big * big / big` works ([guide](docs/guide/04-operators.md#overflow-never-wraps-silently))
- Files, input, and output on both targets: `std:io` (print without a
  newline, red error and yellow warning lines, standard input), `std:process`
  (arguments, environment), `std:fs` (text and binary files, directories), and `std:path`
  (Node.js's path rules, implemented in RuneScript's runtimes)
  ([guide](docs/guide/20-files-and-input.md))
- JSON with `std:json`: `json.parse`, checked reading (`json.get`,
  `json.asInt`, …), building, and `json.stringify`, with the same parser and
  error messages on both targets ([guide](docs/guide/21-json.md))
- CSV, matrices, and versions: `std:csv` (RFC 4180 reading and writing),
  `std:matrix` (products, determinants, inverses, and linear systems on
  `Array<Array<float64>>`, bit-identical on both targets), and `std:semver`
  (SemVer 2.0.0 versions compared and matched like npm)
  ([guide](docs/guide/22-csv-matrix-semver.md))
- Compilation to JavaScript (`runescript build`): an ES2015 module (plus BigInt and `??`) that runs
  in browsers and Node.js and behaves exactly like the interpreter for every
  program the interpreter can run, with optional source maps
  (`--source-map`) ([guide](docs/guide/11-javascript.md))
- Web pages with `rune:dom`: components with HTML and SVG templates whose
  every element and attribute is checked at compile time, escaped `{{ }}`
  text, `$if`/`$elseif`/`$else`, `$foreach`, events, and runes (signals) that update
  exactly what read them, with no virtual DOM
  ([guide](docs/guide/15-web-pages.md), [examples](examples/web/))
- Calling JavaScript through `extern` declarations, including the
  host-neutral `rune:fs`. Every value that comes back from JavaScript is
  checked against its declared type
- Warnings for unused variables, parameters, and uses, unreachable code,
  discarded results, and shadowing, with `-Werror` and `--allow <code>`;
  errors in red and warnings in yellow on a terminal (`--color`)
- Editor support: a language server, `runescript lsp`, with errors as you
  type, hover, go to definition, completion, and an outline, plus packages
  for VS Code, Vim and Neovim, and Emacs ([guide](docs/guide/23-editors.md),
  [editors/](editors/README.md))
- Live reload while you work: `runescript dev app.rsc` serves a web program
  and reloads the browser on every save; `runescript run --watch` re-runs a
  command-line program

**On JavaScript compatibility:** RuneScript does not yet promise full
JavaScript compatibility. What it promises is strict code: unlike
TypeScript, which trusts its declarations and erases them, RuneScript
checks every value that crosses in from JavaScript, and it has no `any`, no
implicit conversions, and no unchecked casts.

**Not yet implemented:** class inheritance, interfaces, and `static`
members, `go:` packages, `std:symrs`, objects, arrays, structs, and
callbacks across the JavaScript boundary, component tags in templates, and
the planned native graphics package. See the [roadmap](docs/ROADMAP.md).

## Install

Requires Go 1.25+ ([go.dev/dl](https://go.dev/dl/)); the installer checks it,
builds the `runescript` CLI, and adds it to your PATH.

```sh
# Linux, macOS
curl -fsSL https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.sh | sh
```

```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.ps1 | iex
```

Options, uninstalling, and installing from a clone are in the
[getting-started guide](docs/guide/01-getting-started.md).

## Build and run from source

```sh
go build -o runescript ./cmd/runescript
./runescript run examples/hello.rsc
./runescript check examples/signedness_error.rsc   # shows a compile-time error
./runescript build examples/fizzbuzz.rsc            # writes examples/fizzbuzz.js
node examples/fizzbuzz.js                         # runs it (Node.js 24+)
./runescript dev examples/fizzbuzz.rsc              # serves it, reloading on every save
./runescript run --watch examples/hello.rsc         # runs it again on every save
./runescript run examples/brainfuck.rsc examples/bf/hello.bf   # arguments follow the file
go test ./...
```

> RuneScript programs use the `.rsc` extension, and package directories hold
> `.rsm` files. Like Perl, the extension is a convention: the CLI and file
> `use`s accept any file name, as long as the file contains RuneScript.

## Tests

```sh
go test ./...                                  # unit, golden, docs, and CLI tests
go test ./tests -run '^$' -fuzz=FuzzCheck      # fuzz the parser and checker
go test ./tests -run '^$' -fuzz=FuzzRun        # fuzz the interpreter
```

The suite includes:
- unit tests for every compiler stage;
- golden tests that run every example and compare its output;
- error programs whose expected diagnostics are written inline as
  `// expect: CODE line:col`;
- runtime-error programs;
- differential tests that run every program on both the interpreter and
  Node.js and require identical results (skipped if `node` is missing);
- Node.js runs of programs that call JavaScript through `extern`, including
  JavaScript that returns wrong values on purpose;
- a check that generated JavaScript uses nothing newer than ES2015 plus
  `BigInt`, `??`, and modules;
- tests for the dev server, the file watcher, and `run --watch`;
- a check that every program in the documentation compiles, and that the
  translations keep every code block of the English pages.

## Benchmarks

[`bench/`](bench/README.md) compares RuneScript with **TypeScript 1.0 RC**
(npm `typescript@0.9.7`, installed locally by `npm install` in `bench/`).
Each program runs three ways: on the interpreter (RS-VM), compiled to
JavaScript (RS-JS), and as the same algorithm in TypeScript (TS). All
three run on Node.js except RS-VM. Every benchmark is a self-contained
directory. Run a category with `go run ./bench/rsbench run`, `size`,
`compile`, or `diag`, and check correctness with `go run ./bench/rsbench
verify`.

First results, 2026-10-20 (v0.6.0-dev, Windows 10 x64, Node 24.16, small
inputs, median of 5 runs). These are early numbers. Small inputs keep the
interpreter runs short, so several TypeScript times are only a few
milliseconds above Node's startup time.

| | RS-VM | RS-JS | TS |
|---|---:|---:|---:|
| Three-body simulation, 100k steps (net ms) | 1276 | 65 | 8 |
| Merge sort, 50k integers (net ms) | 1040 | 66 | 23 |
| Matrix multiply, 64×64 float64 (net ms) | 172 | 28 | 8 |
| Recursive Fibonacci, n = 24 (net ms) | 115 | 10 | 5 |
| Throw and catch 20k errors (net ms) | 97 | 39 | 182 |
| Empty program, minified + gzip (bytes) | – | 4,823 | 20 |
| Compile a 2,000-function file (ms) | – | 171 | 1,464 |

- **Correctness.** Every benchmark gives the same output on all three
  implementations. Three conformance programs differ from TypeScript on
  purpose: overflow, float32 printing, and string units (bytes vs UTF-16).
  Each difference is recorded and explained in that program's notes.
- **Speed.** RuneScript's generated JavaScript is slower than plain
  TypeScript, by 1.2× to 25× across the suite. Most of that is the price
  of checked integer arithmetic. A TypeScript version with the same int32
  checks (`ts-checked`) closes most of the gap on the addition benchmark.
  RuneScript was faster only where TypeScript has to check by hand:
  throwing and catching errors (4.6×). The interpreter is 2.5× to 33×
  slower than RS-JS, except on JSON, where it is faster: its `std:json`
  is native Go code.
- **Size.** Each RuneScript module carries about 4.8 KB (minified +
  gzip) of runtime. Using `std:math` adds about 6 KB more, even for just
  `math.sqrt`. That makes the runtime the main size target.
- **Compiler.** `runescript build` was 9× to 51× faster than TypeScript
  1.0's `tsc` on the generated projects. `tsc` spends almost a second
  starting up on modern Node.
- **Diagnostics.** On six paired mistakes, both compilers put the first
  error on the right line. TypeScript has no equivalent of RuneScript's
  signed-to-unsigned check.

The method, the full tables, and how to add a benchmark are in
[`bench/README.md`](bench/README.md). Benchmark results are published
only here, not on the website.

## Examples

[`examples/`](examples/README.md) has programs for every feature: the
basics, `std:string`, callbacks and lambdas, a prime counter, a
[Brainfuck interpreter](examples/brainfuck.rsc) with 32,768 cells, two
programs the compiler rejects on purpose, and web pages in
[`examples/web/`](examples/web/) (a static page, a counter, a toggle, an
SVG dial, and a multi-file app). Every example's output is checked by the
test suite; the web examples run on a fake DOM under Node.js.

## Documentation

- [Guide](docs/guide/README.md): tour of every feature, with examples (all compiled by the tests)
- [Language specification](docs/LANGUAGE_SPEC.md): grammar and rules, with each rule's implementation status
- [Standard library: std:string](docs/guide/10-std-string.md)
- [Text and lists: +, $"…", char, Array<T>](docs/guide/16-text-and-arrays.md)
- [No null, no any, no generics](docs/guide/17-no-null-any-generics.md)
- [Numbers in depth: widths, literals, bits, std:math](docs/guide/18-numbers.md)
- [Structs](docs/guide/19-structs.md)
- [Classes](docs/guide/24-classes.md)
- [Files, input, and output: std:io, std:process, std:fs, std:path](docs/guide/20-files-and-input.md)
- [JSON: std:json](docs/guide/21-json.md)
- [CSV, matrices, and versions: std:csv, std:matrix, std:semver](docs/guide/22-csv-matrix-semver.md)
- [Editors: VS Code, Vim, Neovim, Emacs, and runescript lsp](docs/guide/23-editors.md)
- [Compiling to JavaScript](docs/guide/11-javascript.md)
- [Web pages with rune:dom](docs/guide/15-web-pages.md)
- [Diagnostics reference](docs/guide/08-diagnostics.md)
- [Coming from ActionScript 3](docs/guide/09-coming-from-actionscript.md)
- [Roadmap](docs/ROADMAP.md)

## Repository layout

```text
cmd/runescript/        CLI (run, check, build, dev, version)
internal/lexer/        tokens and source positions
internal/parser/       recursive-descent parser → AST
internal/ast/          syntax tree
internal/types/        semantic types
internal/srcenc/       reads source files (UTF-8; UTF-16 and UTF-32 are converted)
internal/loader/       loads a program's files and resolves uses
internal/checker/      name resolution, type checking, and warnings
internal/runtime/      tree-walking interpreter
internal/conv/         number and string conversions shared by both targets
internal/ir/           typed intermediate representation and its verifier
internal/jsgen/        JavaScript code generator and its inlined runtime
internal/jshost/       host-neutral JavaScript modules (rune:fs)
internal/dom/          HTML4, HTML5, and SVG elements and attributes (for templates)
internal/diagnostics/  error codes and formatting
internal/term/         terminal detection for colored output
internal/driver/       wires the stages together
internal/watch/        file watcher for run --watch and dev
internal/devserver/    the runescript dev server and its browser client
internal/lsp/          the language server (runescript lsp)
stdlib/                standard library packages (std:string, std:math, std:io, std:process,
                       std:fs, std:path, std:json, std:csv, std:matrix, std:semver), implemented
                       in Go; their JavaScript twins are in internal/jsgen/runtime.js
examples/              example programs
editors/               VS Code extension, Vim plugin, and Emacs mode
tests/                 end-to-end, golden, and documentation tests
bench/                 benchmark and conformance suite vs TypeScript 1.0 RC (go run ./bench/rsbench)
assets/                logo: SVG source and PNGs generated from it
scripts/               render-logo.sh: SVG → PNG via headless Edge/Chrome
```

## Translations

English is the original: this README and [`docs/`](docs/) are the source of
truth. Translations live in [`localization/`](localization/README.md), one
folder per language:

| Language | Folder | Covers |
|---|---|---|
| বাংলা (Bengali) | [`localization/bn/`](localization/bn/README.md) | README and the guide |
| עברית (Hebrew) | [`localization/heb/`](localization/heb/README.md) | README and the guide |

The Bengali and Hebrew translations are maintained with AI and checked by
hand by the author. **Help is appreciated:** corrections from native
speakers, and new languages, are very welcome. See
[localization/README.md](localization/README.md) for how translations are
organised and kept in step with the English docs.

## Logo

`assets/runescript.svg` (README) and `assets/runescript-icon.svg` (small
icons) are the hand-written source files. The PNGs are generated from
them:

```sh
scripts/render-logo.sh     # needs Microsoft Edge or Chrome; override with BROWSER=...
```

## License

RuneScript is released under the [MIT License](LICENSE).
