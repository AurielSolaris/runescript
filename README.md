<p align="center">
  <img src="assets/runescript.png" alt="RuneScript logo: a rune-styled snowflake with lightning accents" width="160">
</p>

<h1 align="center">RuneScript</h1>

<p align="center">A statically typed, general-purpose language inspired by ActionScript, implemented in Go.</p>

---

RuneScript keeps ActionScript's identity (`package`, `trace`, `int`/`uint`,
`public`/`protected`/`private`) and changes the syntax to a C-family style:
types come before names, there's no `function` keyword, and semicolons are
required. Its type checking is strict: nothing is implicitly `any`, and
signed and unsigned integers never mix silently.

It runs as a standalone program: no browser, no HTML, no Flash Player.
It is **not** source-compatible with ActionScript 3.

```actionscript
package main;

void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

protected int main() {
    sayHi("Hello world", 3);
    return 0;
}
```

## Status: v0.1 (early development)

**Works now**, and is covered by tests:

- Packages, top-level functions with typed parameters, recursion
- `int`/`sint`, `uint`, `bool`, `string`, `void`
- Variables, `=`, `+= -= *= /= %=`, `++`/`--`
- Arithmetic, comparison, and logical operators
- `if`/`else if`/`else`, `while`, `for`, `break`, `continue`, `return`
- `trace(...)`
- Compile-time type checking with located diagnostics, including rejection
  of signed → unsigned assignment
- Integer overflow and division by zero are reported as runtime errors and
  never wrap silently

**Not yet implemented:** classes, structs, arrays, floats, `char`, string
concatenation, `use` (packages and files), type conversions, and the
planned native graphics package. See the [roadmap](docs/ROADMAP.md).

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

## Examples

| File | Shows |
|---|---|
| [`hello.rs`](examples/hello.rs) | The minimal program |
| [`variables.rs`](examples/variables.rs) | Declarations, assignment, compound assignment |
| [`conditionals.rs`](examples/conditionals.rs) | `if` / `else if` / `else`, boolean logic |
| [`loops.rs`](examples/loops.rs) | `for`, `while`, `break`, `continue` |
| [`functions.rs`](examples/functions.rs) | Parameters, return values, recursion |
| [`fizzbuzz.rs`](examples/fizzbuzz.rs) | Putting it together |
| [`signedness_error.rs`](examples/signedness_error.rs) | A program the compiler rejects, on purpose |

Each example's output is checked by `tests/golden_test.go`.

## Documentation

- [Guide](docs/guide/README.md): tour of every v0.1 feature, with examples (all compiled by the tests)
- [Language specification](docs/LANGUAGE_SPEC.md): grammar and rules, with each rule's implementation status
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
