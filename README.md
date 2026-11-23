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

RuneScript runs from a standalone CLI and can emit JavaScript modules
or build static multi-page websites. Its raylib graphics packages work
on desktop and target browsers through WebAssembly. The language uses
its own syntax and type rules.

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

## Status: v0.7 in progress (early development)

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
- Multi-file programs use packages such as `use "geom";` or direct file
  imports such as `use "./lib/x.rsc";`. Each `.rsm` module declares its
  SemVer version and may declare its author and license.
  [Packages and files](docs/guide/12-packages-and-files.md)
- `float` (float32 by default, widening without loss), `float32`, `float64`;
  IEEE 754 with `NaN` and `Infinity`, as in ActionScript
- Checked C-style conversions: `(int) 64.3`, `(float) "64.7"`, `(uint) n`
  ([guide](docs/guide/13-floats-and-conversions.md))
- Exact-width integers `int8` … `uint64` (64-bit values are BigInts in
  JavaScript), hex/binary/octal literals with `_` separators, bit
  operators, and `std:math`, whose results are bit-identical on both
  targets ([guide](docs/guide/18-numbers.md))
- Structs with value semantics, field-by-field equality, named-field
  literals, methods, and constructors. [Structs](docs/guide/19-structs.md)
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
- `build --target js` emits a standalone ES module with optional source
  maps. `build --target web` emits static HTML pages with separate entry
  modules, shared runtime assets, and optional CSS extraction.
  [JavaScript and web builds](docs/guide/11-javascript.md)
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
- Editor support: a language server, `runescript lsp`, with diagnostics,
  navigation, signature help, local rename, formatting, and quick fixes, plus packages
  for VS Code, Vim and Neovim, and Emacs ([guide](docs/guide/23-editors.md),
  [editors/](editors/README.md))
- Graphics through `rune.gfx` and `rune.gui` on desktop and in the browser
  with `build --target wasm`. [Graphics guide](docs/guide/25-native-graphics.md)
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
callbacks across the JavaScript boundary, and component tags in templates.
See the [roadmap](docs/ROADMAP.md).

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
./runescript build --target web --extract-css -o site-dist examples/multi-page
./runescript build --target wasm -o game-web examples/native/keep-up.rsc
./runescript dev examples/fizzbuzz.rsc              # serves it, reloading on every save
./runescript run --watch examples/hello.rsc         # runs it again on every save
./runescript run examples/brainfuck.rsc examples/bf/hello.bf   # arguments follow the file
go test ./...
```

> RuneScript programs use the `.rsc` extension, and package directories hold
> `.rsm` files. Like Perl, the extension is a convention: the CLI and file
> `use`s accept other file names. For the web target, the manifest uses
> `.rsc` entry files and `.html` pages.

## Tests

```sh
go test ./...                                  # unit, golden, docs, and CLI tests
go test ./tests -run '^$' -fuzz=FuzzCheck      # fuzz the parser and checker
go test ./tests -run '^$' -fuzz=FuzzRun        # fuzz the interpreter
```

The suite includes:
- unit tests for every compiler stage;
- golden tests for command-line examples, plus build and binding tests
  for web and graphics examples;
- error programs whose expected diagnostics are written inline as
  `// expect: CODE line:col`;
- runtime-error programs;
- differential tests that compare the Go execution engine and generated
  JavaScript for programs they both support (skipped without Node.js);
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
The suite compares the Go CLI (RS-VM), generated JavaScript (RS-JS), and
equivalent TypeScript (TS). The JavaScript and TypeScript cases run on
Node.js. Every benchmark is a self-contained
directory. Run a category with `go run ./bench/rsbench run`, `size`,
`compile`, or `diag`, and check correctness with `go run ./bench/rsbench
verify`.

Measured on 2026-11-24 with RuneScript v0.7.0-dev, Go 1.25.11, and Node.js
24.16.0 on Windows x64. Timings are medians of five runs. Runtime rows
use small inputs and include process startup. The empty program shows
the startup cost for each runner.

| | RS-VM | RS-JS | TS |
|---|---:|---:|---:|
| Empty program (ms) | 635.7 | 86.9 | 84.5 |
| Three-body simulation, 100k steps (ms) | 1,978.3 | 161.5 | 92.5 |
| Merge sort, 50k integers (ms) | 1,670.4 | 169.6 | 107.4 |
| Matrix multiply, 64×64 float64 (ms) | 918.3 | 131.3 | 92.3 |
| Recursive Fibonacci, n = 24 (ms) | 681.5 | 100.6 | 83.1 |
| Throw and catch 20k errors (ms) | 727.6 | 134.2 | 347.6 |
| Empty program, minified + gzip (bytes) | – | 4,862 | 20 |
| Compile a 2,000-function file (ms) | – | 948.1 | 1,894.9 |

Run `go run ./bench/rsbench web` to build the two-page example and compare
its payload and mount time with hand-written JavaScript. The two-page
payload counts identical runtime and CSS assets once. Mount timings use
a fake DOM and exclude browser layout and paint.

On the two-page sample with Node.js 24.16, these were the gzipped
JavaScript and CSS totals, excluding HTML.

| Build | First page | Both pages |
| --- | ---: | ---: |
| Hand-written JavaScript and CSS | 399 B | 700 B |
| Standalone RuneScript modules | 11,739 B | 23,483 B |
| RuneScript web build with inline CSS | 11,687 B | 12,784 B |
| RuneScript web build with CSS extraction | 11,796 B | 12,864 B |

Sharing the runtime cut the two-page payload by about 45%. Extracting CSS
reduced the first page's own gzipped JavaScript from 1,093 B to 1,064 B.
The separate stylesheet increased the combined payload by 80 B on this
sample.

See [`bench/README.md`](bench/README.md) for the methodology, commands
that generate full reports, and instructions for adding a benchmark.

## Examples

[`examples/`](examples/README.md) has programs for every feature: the
basics, `std:string`, callbacks and lambdas, a prime counter, a
[Brainfuck interpreter](examples/brainfuck.rsc) with 32,768 cells, two
programs the compiler rejects on purpose, and web pages in
[`examples/web/`](examples/web/) (a static page, a counter, a toggle, an
SVG dial, and a multi-file app). There are also a
[multi-page site](examples/multi-page/) and
[desktop and WASM games](examples/native/). The suite checks command-line
output, DOM behavior under Node.js, multi-page build artifacts, and
graphics bindings.

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
- [Desktop and WASM graphics](docs/guide/25-native-graphics.md)
- [Diagnostics reference](docs/guide/08-diagnostics.md)
- [Coming from ActionScript 3](docs/guide/09-coming-from-actionscript.md)
- [Roadmap](docs/ROADMAP.md)

## Repository layout

```text
cmd/runescript/        CLI: run, check, build, dev, lsp, version
internal/
  srcenc/              Source encoding detection and UTF-8 conversion
  lexer/               Tokens and source positions
  parser/              RuneScript and template parser
  ast/                 Syntax tree
  loader/              Package and file imports
  checker/             Name resolution, static types, module metadata, warnings
  types/               Semantic types
  ir/                  Typed intermediate representation, lowering, verification
  runtime/             Go closure engine and reference tree walker
  jsgen/               JavaScript emission, runtime modules, source maps, CSS
  jshost/              Host-neutral JavaScript adapters
  webbuild/            Multi-page HTML and asset assembly
  wasmbuild/           Go and raylib WebAssembly bundle assembly
  dom/                 HTML and SVG metadata for template checks
  devserver/           Local HTTP server and browser live-reload client
  watch/               File polling for dev and run --watch
  lsp/                 Language server
  conv/                Checked conversions and number formatting
  fmath/               Floating-point math shared by the Go and JS targets
  diagnostics/         Error and warning codes, positions, and formatting
  term/                Terminal detection and color support
  driver/              Shared compiler and execution entry points
stdlib/                Go-native standard library, graphics, and widgets
examples/              CLI programs, web templates, multi-page site, and games
editors/               VS Code, Vim and Neovim, and Emacs integrations
tests/                 End-to-end and documentation tests with testdata fixtures
bench/                 rsbench runner, TypeScript suite, profiling, and web benchmark
docs/                  Language specification, guide, and roadmap
localization/          Bengali and Hebrew README and guide translations
assets/                RuneScript SVGs and generated PNGs
scripts/               Logo rendering and translation status
install.sh, install.ps1  Per-user installers
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
