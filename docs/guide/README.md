# RuneScript Guide (v0.3)

A practical tour of the RuneScript features that exist **today**. Every
complete program in these pages is compiled by the test suite
(`tests/docs_test.go`). If one stops compiling, the build fails.

1. [Getting started](01-getting-started.md): build the CLI, run a program, `run --watch`
2. [Program structure](02-program-structure.md): packages, comments, `main`
3. [Types and variables](03-types-and-variables.md): `int`, `uint`, `bool`, `string`, assignment
4. [Operators](04-operators.md): arithmetic, comparison, logic
5. [Control flow](05-control-flow.md): `if`, `while`, `for`, `break`, `continue`
6. [Functions](06-functions.md): parameters, return values, recursion
7. [Output with `trace`](07-trace.md)
8. [Diagnostics reference](08-diagnostics.md): every error and warning code, `-Werror`, `--allow`
9. [Coming from ActionScript 3](09-coming-from-actionscript.md)
10. [Standard library: `std:string`](10-std-string.md): Go's `strings`, in RuneScript
11. [Compiling to JavaScript](11-javascript.md): `runescript build`, how strict typing is kept, calling JavaScript with `extern`, and live reload with `runescript dev`
12. [Packages and files](12-packages-and-files.md): `use`, multi-file programs, visibility
13. [Floats and conversions](13-floats-and-conversions.md): `float`, `float32`, `float64`, `(int) x`
14. [Errors](14-errors.md): `throw`, `try`, `catch`, `finally`, and catching runtime errors

For the precise rules, see the [language specification](../LANGUAGE_SPEC.md).
For what's coming next, see the [roadmap](../ROADMAP.md).

> **Not yet available:** classes, structs, arrays, `char`, and string
> concatenation. Their keywords are reserved, and using them gives a clear
> "not implemented yet" error. Objects and callbacks can't cross the
> JavaScript boundary yet either.
