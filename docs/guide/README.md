# RuneScript Guide (v0.6)

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
15. [Web pages with `rune:dom`](15-web-pages.md): components, checked HTML and SVG templates, runes, events, `$if`, `renderApp`
16. [Text and lists](16-text-and-arrays.md): joining strings with `+`, interpolated strings `$"…"`, `char`, and `Array<T>`
17. [No `null`, no `any`, no generics](17-no-null-any-generics.md): why, and how to use `false` where other languages use `null`
18. [Numbers in depth](18-numbers.md): `int8` … `uint64`, hex and binary literals, bit operations, `std:math`
19. [Structs](19-structs.md): declaring, building, and copying value types
20. [Files, input, and output](20-files-and-input.md): `std:io`, `std:process`, `std:fs`, `std:path`, and a Brainfuck interpreter
21. [JSON](21-json.md): `std:json`: parsing, reading with checked functions, building, and writing JSON

For the precise rules, see the [language specification](../LANGUAGE_SPEC.md).
For what's coming next, see the [roadmap](../ROADMAP.md).

> **Not yet available:** classes. Their keyword is reserved, and using it
> gives a clear "not implemented yet" error. Objects, arrays, structs, and
> callbacks can't cross the JavaScript boundary yet either.
