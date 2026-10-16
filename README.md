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

## Status: v0.3 in progress (early development)

**Works now**, and is covered by tests:

- Packages; functions declared `public|private|protected <type> <name>(…)`; recursion
- Nested functions, prototypes, function types (`int(int, int)`), callbacks, and lambdas (`(int a) => a * 2`)
- `int`/`sint`, `uint`, `bool`, `string`, `void`
- Variables, `=`, `+= -= *= /= %=`, `++`/`--`
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
- Text and lists for scripting: `+` and `$"{count} files"` for strings,
  `char`, and `Array<T>` with bounds-checked indexing
  ([guide](docs/guide/16-text-and-arrays.md)); `std:string` covers Go's
  `strings` package, including `split`, `fields`, `join`, and callbacks
  ([guide](docs/guide/10-std-string.md))
- ActionScript-style errors: `throw new RangeError("…")`,
  `try`/`catch (Error e)`/`catch (e)`/`finally`; runtime errors are catchable
  ([guide](docs/guide/14-errors.md))
- Compile-time type checking with located diagnostics, including rejection
  of signed → unsigned assignment
- Integer overflow and division by zero are reported as runtime errors and
  never wrap silently
- Compilation to JavaScript (`runescript build`): an ES2015 module (plus BigInt and `??`) that runs
  in browsers and Node.js and behaves exactly like the interpreter for every
  program the interpreter can run ([guide](docs/guide/11-javascript.md))
- Web pages with `rune:dom`: components with HTML and SVG templates whose
  every element and attribute is checked at compile time, escaped `{{ }}`
  text, `$if`/`$elseif`/`$else`, events, and runes (signals) that update
  exactly what read them, with no virtual DOM
  ([guide](docs/guide/15-web-pages.md), [examples](examples/web/))
- Calling JavaScript through `extern` declarations, including the
  host-neutral `rune:fs`. Every value that comes back from JavaScript is
  checked against its declared type
- Warnings for unused variables, parameters, and uses, unreachable code, and
  discarded results, with `-Werror` and `--allow <code>`
- Live reload while you work: `runescript dev app.rsc` serves a web program
  and reloads the browser on every save; `runescript run --watch` re-runs a
  command-line program

**On JavaScript compatibility:** RuneScript does not yet promise full
JavaScript compatibility. What it promises is strict code: unlike
TypeScript, which trusts its declarations and erases them, RuneScript
checks every value that crosses in from JavaScript, and it has no `any`, no
implicit conversions, and no unchecked casts.

**Not yet implemented:** classes, structs, for-each loops, command-line
arguments, `go:` packages, `std:math`, `std:symrs`, objects, arrays, and
callbacks across the JavaScript boundary, `rune:process`, `$foreach` and
component tags in templates, and the planned native graphics package. See
the [roadmap](docs/ROADMAP.md).

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

## Examples

[`examples/`](examples/README.md) has programs for every feature: the
basics, `std:string`, callbacks and lambdas, a prime counter, two
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
internal/driver/       wires the stages together
internal/watch/        file watcher for run --watch and dev
internal/devserver/    the runescript dev server and its browser client
stdlib/                standard library packages (std:string), implemented in Go
examples/              example programs
tests/                 end-to-end, golden, and documentation tests
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
