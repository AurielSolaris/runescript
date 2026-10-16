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
| `ASX-LEX-006` | A character literal that doesn't hold exactly one character: `''`, `'ab'` |
| `ASX-LEX-007` | The file isn't valid UTF-8 (e.g. it was saved as Windows-1252). UTF-16 and UTF-32 files are converted instead |
| `ASX-LEX-008` | A template, tag, interpolation, or `$if` condition that is never closed |

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
| `ASX-NAME-009` | Using an enclosing function's variable inside a nested function or lambda (functions don't capture, except runes) |
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
| `ASX-USE-009` | An extern names an unknown `rune:` module, or a function that module doesn't have |

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
| `ASX-TYPE-023` | An extern function's parameter or result has a type that can't cross the JavaScript boundary, e.g. a function type |
| `ASX-TYPE-024` | An extern declaration of a `rune:` function doesn't match that function's type |
| `ASX-TYPE-025` | Wrong type arguments: `Array` takes one element type (`Array<string>`), `Signal` takes one number, `bool`, or `string` type (`Signal<int>`); other types take none |

## Template errors

Templates belong to `rune:dom` ([chapter 15](15-web-pages.md)).

| Code | Meaning |
|---|---|
| `ASX-TMPL-001` | A `template` block that isn't the last statement of a function returning `Element` |
| `ASX-TMPL-002` | An element that isn't in the HTML4, HTML5, or SVG list (the help says when it's in the other namespace or differs in case) |
| `ASX-TMPL-003` | An attribute the element doesn't have |
| `ASX-TMPL-004` | A static attribute value the attribute can't take: a value on a boolean attribute, a keyword not in its list, a bad number or length, a URL that would run script, or text for an event handler |
| `ASX-TMPL-005` | A `{{ }}` value of the wrong type, such as a rune that isn't read (`{{ count }}`), or an `int` for a text attribute |
| `ASX-TMPL-006` | The same attribute twice on one element |
| `ASX-TMPL-007` | A template in a file without `use "rune:dom";` |

## Other compile-time errors

| Code | Meaning |
|---|---|
| `ASX-FLOW-001` | `break` or `continue` outside a loop |
| `ASX-ENTRY-001` | No `main` function |
| `ASX-ENTRY-002` | `main` has parameters, or returns something other than `int`/`void` |
| `ASX-ENTRY-003` | More than one file defines `main`; a program has exactly one |
| `ASX-TARGET-002` | `runescript run` was given a program that only the JavaScript target can run: it has extern functions, or uses `rune:dom` |

## Warnings

A warning points at code that is legal but probably a mistake. The program
still runs:

```text
prog.rsc:3:9: warning ASX-WARN-001: variable 'unused' is never used
        int unused = 1;
            ^
  help: remove the declaration
```

Warnings are reported only when a program has no errors, since errors often
cause follow-on warnings that disappear once the errors are fixed.
`runescript run`, `check`, and `build` all take two flags:

- `-Werror` makes every warning an error: the program doesn't run or build,
  and each warning is shown as `error ASX-WARN-…: … [-Werror]`.
- `--allow <codes>` stops reporting the listed warnings, e.g.
  `--allow ASX-WARN-002` or `--allow WARN-002,WARN-005`. The flag may be
  repeated.

| Code | Meaning |
|---|---|
| `ASX-WARN-001` | A local variable is never read. Assigning to it, including `x += 1` and `x++`, doesn't count as reading |
| `ASX-WARN-002` | A parameter of a function or lambda is never read. The variable of a `catch` clause is never reported |
| `ASX-WARN-003` | A `use` declaration whose qualifier is never referenced |
| `ASX-WARN-004` | Statements that can never run, because they come after a `return`, `throw`, `break`, or `continue` (or an `if`/`try` whose every path ends in one). Loops are not followed, so code after `while (true) { }` isn't reported |
| `ASX-WARN-005` | A call used as a statement whose result isn't `void` is thrown away |
| `ASX-WARN-006` | An extern names one host's own module, e.g. `"node:fs"`; prefer a host-neutral `rune:` module |
| `ASX-WARN-007` | An element or attribute that HTML5 or SVG 2 made obsolete, e.g. `<center>`, `bgcolor`, `xlink:href`; the help names the replacement |
| `ASX-WARN-008` | `dom.renderApp` with a selector that isn't a single id, e.g. `".app"`: it may match several elements, or none |

## Runtime errors

These stop the program with exit status 1, unless a `try` statement
catches them (see [chapter 14](14-errors.md)). Anything traced before the
error has already been printed.

| Code | Meaning | Caught as |
|---|---|---|
| `ASX-RUN-001` | Calls nested more than 10,000 deep (usually runaway recursion) | can't be caught |
| `ASX-RUN-002` | Integer overflow: the result doesn't fit its type (floats become `Infinity` instead) | `RangeError` |
| `ASX-RUN-003` | Integer division or remainder by zero (for floats, the result is `Infinity` or `NaN`) | `RangeError` |
| `ASX-RUN-004` | A standard library function failed, e.g. `string.repeat` with a negative count, or `string.slice` with bounds out of range or inside a character | `Error`; `RangeError` for `string.slice` |
| `ASX-RUN-005` | An error created with `new` was thrown and nothing caught it | |
| `ASX-RUN-006` | A conversion failed, e.g. `(int) "12x"` or `(uint) -1` at run time | `TypeError` or `RangeError` |
| `ASX-RUN-007` | A JavaScript function returned a value that doesn't match its extern declaration, e.g. `1.5` for an `int` | `TypeError` |
| `ASX-RUN-008` | A JavaScript function threw an exception, or its module has no such function | `Error` |
| `ASX-RUN-009` | `dom.renderApp` found no element matching its selector | `Error` |
| `ASX-RUN-010` | A bound attribute got a value it can't take: a keyword not in its list, or a URL that would run script | `TypeError` |
| `ASX-RUN-011` | An array index out of range (`a[i]` with `i < 0` or `i >= a.length`), or `pop()` on an empty array | `RangeError` |

On the JavaScript target, `ASX-RUN-001` is also reported when the host's own
stack runs out before 10,000 calls (see [chapter 11](11-javascript.md)).

Next: [Coming from ActionScript 3](09-coming-from-actionscript.md)
