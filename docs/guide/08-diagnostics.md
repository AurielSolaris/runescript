# 8. Diagnostics reference

Every error points at the exact source location:

```text
prog.rsc:9:26: error ASX-TYPE-004: Cannot assign signed integer 'sint' to unsigned integer 'uint'.
        uint unsignedValue = signedValue;
                             ^^^^^^^^^^^
  help: Use an explicit checked conversion, (uint) x, or change the destination type.
```

The `ASX` prefix is provisional and may be renamed. The numbers stay the
same.

## Lexical errors

| Code | Meaning |
|---|---|
| `ASX-LEX-001` | A character that isn't part of the language, e.g. `#`, or a lone `&` |
| `ASX-LEX-002` | A string literal with no closing `"` before the end of the line |
| `ASX-LEX-003` | An unknown escape sequence such as `\q` |
| `ASX-LEX-004` | A `/*` comment that is never closed |
| `ASX-LEX-005` | An invalid number: leading zero (`007`), letters attached (`12abc`), too many dots (`1.2.3`), or too large or too small (`1e999`, `1e-999`) |
| `ASX-LEX-006` | A literal kind that isn't supported yet: chars (`'a'`) |
| `ASX-LEX-007` | The file isn't valid UTF-8 (e.g. it was saved as Windows-1252). UTF-16 and UTF-32 files are converted instead |

## Syntax errors

| Code | Meaning |
|---|---|
| `ASX-PARSE-001` | Something unexpected, e.g. a missing `;`, a missing `package`, a function without an access modifier, an AS3-style `function` keyword, a missing initializer |
| `ASX-PARSE-002` | A reserved word for a feature that isn't implemented yet (`class`, `char`, …) |

The parser stops at the first syntax error.

## Name errors

| Code | Meaning |
|---|---|
| `ASX-NAME-001` | Undefined variable or function |
| `ASX-NAME-002` | Name already declared: redeclaration, shadowing, or a duplicate function |
| `ASX-NAME-003` | Calling something that is a variable, not a function |
| `ASX-NAME-004` | Using the built-in `trace` as a value |
| `ASX-NAME-005` | Declaring a function named `trace` |
| `ASX-NAME-006` | Assigning to a parameter or a function |
| `ASX-NAME-007` | Package qualifier not in scope, e.g. `string.toUpper` without `use "std:string";` |
| `ASX-NAME-008` | The package has no function with that name (the help text suggests the camelCase spelling if you used Go's) |
| `ASX-NAME-009` | Using an enclosing function's variable inside a nested function or lambda (functions don't capture) |
| `ASX-NAME-010` | A prototype whose function is never defined |
| `ASX-NAME-011` | A function used where its access modifier forbids it: another file's `private` function, or another package's `protected` one |

## `use` errors

| Code | Meaning |
|---|---|
| `ASX-USE-001` | Unknown standard library package, e.g. `use "std:nope";` |
| `ASX-USE-002` | The same package is used twice under the same name |
| `ASX-USE-003` | A `use` form that isn't supported yet, e.g. `go:` packages |
| `ASX-USE-004` | The used file doesn't exist, or the package's directory has no `.rsm` files (both are relative to the working directory) |
| `ASX-USE-005` | Packages use each other in a cycle; the message shows the cycle |
| `ASX-USE-006` | Not a valid package name, e.g. `"1abc"` or `"if.then"`, or a path written with `\` |
| `ASX-USE-007` | A `.rsm` file in a package's directory declares a different package |
| `ASX-USE-008` | Two uses in one file would have the same qualifier; give one an alias with `as` |

## Type errors

| Code | Meaning |
|---|---|
| `ASX-TYPE-001` | Incompatible types, e.g. `int x = "a";` |
| `ASX-TYPE-002` | `void` used as a value, or as a variable/parameter type |
| `ASX-TYPE-003` | Wrong number of arguments |
| `ASX-TYPE-004` | **Signed integer assigned, passed, or returned as `uint`** |
| `ASX-TYPE-005` | `uint` assigned, passed, or returned as a signed integer |
| `ASX-TYPE-006` | Constant too large or too small for its type |
| `ASX-TYPE-007` | Negative constant used as `uint` |
| `ASX-TYPE-008` | Function may finish without returning a value |
| `ASX-TYPE-009` | `return value;` in a `void` function, or bare `return;` in a non-void one |
| `ASX-TYPE-010` | Unknown type name |
| `ASX-TYPE-011` | Signed and unsigned operands mixed in one operator |
| `ASX-TYPE-012` | Operator not defined for these operand types, or tracing a function value |
| `ASX-TYPE-013` | Condition isn't `bool` |
| `ASX-TYPE-014` | Division by a constant zero |
| `ASX-TYPE-015` | A function definition doesn't match its prototype |
| `ASX-TYPE-016` | Calling a value that isn't a function, e.g. `(1)(2)` |
| `ASX-TYPE-017` | A conversion that isn't allowed, e.g. `(int) true` or `(string) e` for an error `e` |
| `ASX-TYPE-018` | A constant that its float type can't represent exactly, e.g. `float32 f = 0.1;` (write `(float32) 0.1`) |
| `ASX-TYPE-019` | `new` of something other than an error type |
| `ASX-TYPE-020` | A field that doesn't exist, e.g. `e.what`, a field of a non-error value, or a method call such as `e.toString()` |
| `ASX-TYPE-021` | `throw` of a value that isn't an error, or a catch clause whose type isn't an error type |
| `ASX-TYPE-022` | A catch clause that can never run, because an earlier clause already catches its type |

## Other compile-time errors

| Code | Meaning |
|---|---|
| `ASX-FLOW-001` | `break` or `continue` outside a loop |
| `ASX-ENTRY-001` | No `main` function |
| `ASX-ENTRY-002` | `main` has parameters, or returns something other than `int`/`void` |
| `ASX-ENTRY-003` | More than one file defines `main`; a program has exactly one |

## Runtime errors

These stop the program with exit status 1, unless a `try` statement
catches them (see [chapter 14](14-errors.md)). Anything traced before the
error has already been printed.

| Code | Meaning | Caught as |
|---|---|---|
| `ASX-RUN-001` | Calls nested more than 10,000 deep (usually runaway recursion) | can't be caught |
| `ASX-RUN-002` | Integer overflow: the result doesn't fit its type (floats become `Infinity` instead) | `RangeError` |
| `ASX-RUN-003` | Integer division or remainder by zero (for floats, the result is `Infinity` or `NaN`) | `RangeError` |
| `ASX-RUN-004` | A standard library function failed, e.g. `string.repeat` with a negative count | `Error` |
| `ASX-RUN-005` | An error created with `new` was thrown and nothing caught it | |
| `ASX-RUN-006` | A conversion failed, e.g. `(int) "12x"` or `(uint) -1` at run time | `TypeError` or `RangeError` |

On the JavaScript target, `ASX-RUN-001` is also reported when the host's own
stack runs out before 10,000 calls (see [chapter 11](11-javascript.md)).

Next: [Coming from ActionScript 3](09-coming-from-actionscript.md)
