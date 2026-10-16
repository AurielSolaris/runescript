# Examples

Run any example with:

```sh
go run ./cmd/runescript run examples/hello.rsc
go run ./cmd/runescript run --watch examples/hello.rsc   # runs it again on every save
go run ./cmd/runescript dev examples/fizzbuzz.rsc        # in the browser, with live reload
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

These two are **rejected on purpose**. Run them with `check` to see the
diagnostics:

| Example | Error |
|---|---|
| [`signedness_error.rsc`](signedness_error.rsc) | `ASX-TYPE-004`: a signed value can't go into a `uint` |
| [`boundary_error.rsc`](boundary_error.rsc) | `ASX-NAME-009`: nested functions don't capture outer variables |

Every example is tested by `tests/golden_test.go`. Programs are run and
their output compared with `tests/testdata/examples/<name>.out`. Programs
ending in `_error` must produce exactly the diagnostics in their
`// expect:` comments.

