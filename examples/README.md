# Examples

Run any example with:

```sh
go run ./cmd/runescript run examples/hello.rsc
go run ./cmd/runescript run --watch examples/hello.rsc   # runs it again on every save
go run ./cmd/runescript dev examples/fizzbuzz.rsc        # in the browser, with live reload
go run ./cmd/runescript run examples/brainfuck.rsc examples/bf/hello.bf   # with arguments
echo "Hello, World!" | go run ./cmd/runescript run examples/brainfuck.rsc examples/bf/rot13.bf
```

Programs that call JavaScript through `extern` run only on the JavaScript
target, so they live in [`tests/testdata/extern/`](../tests/testdata/extern/)
with their JavaScript helper modules instead of here.

| Example | Shows |
|---|---|
| [`hello.rsc`](hello.rsc) | The minimal program |
| [`variables.rsc`](variables.rsc) | Declarations, assignment, compound assignment, `sint`/`uint`/`bool`/`string` |
| [`conditionals.rsc`](conditionals.rsc) | `if` / `else if` / `else`, comparisons, boolean logic |
| [`loops.rsc`](loops.rsc) | `for`, `while`, `break`, `continue` |
| [`functions.rsc`](functions.rsc) | Parameters, return values, recursion, calling before declaration |
| [`fizzbuzz.rsc`](fizzbuzz.rsc) | Putting the basics together |
| [`strings.rsc`](strings.rsc) | The `std:string` standard library |
| [`callbacks.rsc`](callbacks.rsc) | Function types, callbacks, lambdas, nested functions, prototypes, returning functions |
| [`primes.rsc`](primes.rsc) | Predicate callbacks, a lambda, a nested helper, and `string.repeat` |
| [`temperatures.rsc`](temperatures.rsc) | `float`, `float32`, and checked conversions such as `(int) f` and `(float) "98.6"` |
| [`errors.rsc`](errors.rsc) | `throw`, `try`/`catch`/`finally`, error types, and catching a runtime error |
| [`wordcount.rsc`](wordcount.rsc) | A text report: `Array<T>`, `char` callbacks, `+`, `$"…"`, and `string.split`/`fields`/`map`/`join` |
| [`brainfuck.rsc`](brainfuck.rsc) | A Brainfuck interpreter with 32,768 cells: `std:fs`, `std:io`, `std:process`, structs, `Array<uint8>`. Programs to run are in [`bf/`](bf/): `hello.bf`, `alphabet.bf`, `cat.bf`, `rot13.bf` |

These two are **rejected on purpose**. Run them with `check` to see the
diagnostics:

| Example | Error |
|---|---|
| [`signedness_error.rsc`](signedness_error.rsc) | `ASX-TYPE-004`: a signed value can't go into a `uint` |
| [`boundary_error.rsc`](boundary_error.rsc) | `ASX-NAME-009`: nested functions don't capture outer variables |

## Web pages

[`web/`](web/) has `rune:dom` programs ([guide chapter 15](../docs/guide/15-web-pages.md)).
They need a browser, so run them with `runescript dev`, which reloads the
page on every save:

```sh
go run ./cmd/runescript dev examples/web/counter/app.rsc
```

| Example | Shows |
|---|---|
| [`web/hello`](web/hello/app.rsc) | A static page: checked elements and attributes, escaped text, character references |
| [`web/counter`](web/counter/app.rsc) | A rune, events, a bound `disabled`, and `$if` |
| [`web/toggle`](web/toggle/app.rsc) | `$if` / `$else`, `aria-expanded` bound to a `bool` rune |
| [`web/dial`](web/dial/app.rsc) | SVG: a circle whose radius is a rune |
| [`web/multi-file`](web/multi-file/) | `dom.renderApp(app.App, "#app")`, components in other files, and an `index.html` for `runescript build`. Run it from its folder: `cd examples/web/multi-file && runescript dev main.rsc` |

`tests/dom_test.go` runs every web example on a small fake DOM under
Node.js, clicks what its `// click:` comments name, and compares the HTML
with `tests/testdata/web/<name>.out`.

## Tests

Every example is tested by `tests/golden_test.go`. Programs are run and
their output compared with `tests/testdata/examples/<name>.out`. Programs
ending in `_error` must produce exactly the diagnostics in their
`// expect:` comments.

