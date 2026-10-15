# RuneScript Guide (v0.1)

A practical tour of the RuneScript features that exist **today**. Every
complete program in these pages is compiled by the test suite
(`tests/docs_test.go`). If one stops compiling, the build fails.

1. [Getting started](01-getting-started.md): build the CLI, run a program
2. [Program structure](02-program-structure.md): packages, comments, `main`
3. [Types and variables](03-types-and-variables.md): `int`, `uint`, `bool`, `string`, assignment
4. [Operators](04-operators.md): arithmetic, comparison, logic
5. [Control flow](05-control-flow.md): `if`, `while`, `for`, `break`, `continue`
6. [Functions](06-functions.md): parameters, return values, recursion
7. [Output with `trace`](07-trace.md)
8. [Diagnostics reference](08-diagnostics.md): every error code
9. [Coming from ActionScript 3](09-coming-from-actionscript.md)
10. [Standard library: `std:string`](10-std-string.md): Go's `strings`, in RuneScript
11. [Compiling to JavaScript](11-javascript.md): `runescript build`, and how strict typing is kept

For the precise rules, see the [language specification](../LANGUAGE_SPEC.md).
For what's coming next, see the [roadmap](../ROADMAP.md).

> **Not yet available:** classes, structs, arrays, floats, `char`, string
> concatenation, using files and non-standard packages, and type conversions. Their keywords are
> reserved, and using them gives a clear "not implemented yet" error.
