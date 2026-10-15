<p align="center">
  <img src="assets/runescript.png" alt="RuneScript logo: a rune-styled snowflake with lightning accents" width="160">
</p>

<h1 align="center">RuneScript</h1>

<p align="center">A statically typed, general-purpose language inspired by ActionScript, implemented in Go.</p>

---

RuneScript keeps ActionScript's identity (`package`, `trace`, `int`/`uint`,
`public`/`protected`/`private`) and changes the syntax to a C-family style:
types come before names, every function starts with an access modifier
(`public int add(int a, int b)`) instead of a `function` keyword, and
semicolons are required. Its type checking is strict: nothing is implicitly `any`, and
signed and unsigned integers never mix silently.

It runs as a standalone program: no browser, no HTML, no Flash Player.
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

## Status: v0.2 in progress (early development)

**Works now**, and is covered by tests:

- Packages; functions declared `public|private|protected <type> <name>(…)`; recursion
- Nested functions, prototypes, function types (`int(int, int)`), callbacks, and lambdas (`(int a) => a * 2`)
- `int`/`sint`, `uint`, `bool`, `string`, `void`
- Variables, `=`, `+= -= *= /= %=`, `++`/`--`
- Arithmetic, comparison, and logical operators
- `if`/`else if`/`else`, `while`, `for`, `break`, `continue`, `return`
- `trace(...)`
- Standard library: `use "std:string";` gives Go's `strings` functions as `string.toUpper(s)` and so on
- Compile-time type checking with located diagnostics, including rejection
  of signed → unsigned assignment
- Integer overflow and division by zero are reported as runtime errors and
  never wrap silently

**Not yet implemented:** classes, structs, arrays, floats, `char`, string
concatenation, `use` for files and non-standard packages, type conversions, `std:math`, `std:symrs`, and the planned native
graphics package. See the [roadmap](docs/ROADMAP.md).

## Build and run

Requires Go 1.25+.

```sh
go build -o runescript ./cmd/runescript
./runescript run examples/hello.rs
./runescript check examples/signedness_error.rs   # shows a compile-time error
go test ./...
```

> The `.rs` file extension is provisional, since it clashes with Rust. The
> CLI accepts any file name.

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
- a check that every program in the documentation compiles.

## Examples

[`examples/`](examples/README.md) has programs for every feature: the
basics, `std:string`, callbacks and lambdas, a prime counter, and two
programs the compiler rejects on purpose. Every example's output is
checked by the test suite.

## Documentation

- [Guide](docs/guide/README.md): tour of every feature, with examples (all compiled by the tests)
- [Language specification](docs/LANGUAGE_SPEC.md): grammar and rules, with each rule's implementation status
- [Standard library: std:string](docs/guide/10-std-string.md)
- [Diagnostics reference](docs/guide/08-diagnostics.md)
- [Coming from ActionScript 3](docs/guide/09-coming-from-actionscript.md)
- [Roadmap](docs/ROADMAP.md)

## Repository layout

```text
cmd/runescript/        CLI (run, check, version)
internal/lexer/        tokens and source positions
internal/parser/       recursive-descent parser → AST
internal/ast/          syntax tree
internal/types/        semantic types
internal/checker/      name resolution and type checking
internal/runtime/      tree-walking interpreter
internal/diagnostics/  error codes and formatting
internal/driver/       wires the stages together
stdlib/                standard library packages (std:string), implemented in Go
examples/              example programs
tests/                 end-to-end, golden, and documentation tests
assets/                logo: SVG source and PNGs generated from it
scripts/               render-logo.sh: SVG → PNG via headless Edge/Chrome
```

## Logo

`assets/runescript.svg` (README) and `assets/runescript-icon.svg` (small
icons) are the hand-written source files. The PNGs are generated from
them:

```sh
scripts/render-logo.sh     # needs Microsoft Edge or Chrome; override with BROWSER=...
```
