# Examples

Run any example with:

```sh
go run ./cmd/runescript run examples/hello.rs
```

| Example | Shows |
|---|---|
| [`hello.rs`](hello.rs) | The minimal program |
| [`variables.rs`](variables.rs) | Declarations, assignment, compound assignment, `sint`/`uint`/`bool`/`string` |
| [`conditionals.rs`](conditionals.rs) | `if` / `else if` / `else`, comparisons, boolean logic |
| [`loops.rs`](loops.rs) | `for`, `while`, `break`, `continue` |
| [`functions.rs`](functions.rs) | Parameters, return values, recursion, calling before declaration |
| [`fizzbuzz.rs`](fizzbuzz.rs) | Putting the basics together |
| [`strings.rs`](strings.rs) | The `std:string` standard library |
| [`callbacks.rs`](callbacks.rs) | Function types, callbacks, lambdas, nested functions, prototypes, returning functions |
| [`primes.rs`](primes.rs) | Predicate callbacks, a lambda, a nested helper, and `string.repeat` |

These two are **rejected on purpose**. Run them with `check` to see the
diagnostics:

| Example | Error |
|---|---|
| [`signedness_error.rs`](signedness_error.rs) | `ASX-TYPE-004`: a signed value can't go into a `uint` |
| [`boundary_error.rs`](boundary_error.rs) | `ASX-NAME-009`: nested functions don't capture outer variables |

Every example is tested by `tests/golden_test.go`. Programs are run and
their output compared with `tests/testdata/examples/<name>.out`. Programs
ending in `_error` must produce exactly the diagnostics in their
`// expect:` comments.

> The `.rs` extension is provisional (it is also Rust's).
