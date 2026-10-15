# 8. Diagnostics reference

Every error points at the exact source location:

```text
prog.rs:9:26: error ASX-TYPE-004: Cannot assign signed integer 'sint' to unsigned integer 'uint'.
        uint unsignedValue = signedValue;
                             ^^^^^^^^^^^
  help: Use an explicit checked conversion (syntax not yet specified), or change the destination type.
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
| `ASX-LEX-005` | An invalid integer: leading zero (`007`), letters attached (`12abc`), or too large |
| `ASX-LEX-006` | A literal kind that isn't supported yet: floats (`1.5`), chars (`'a'`) |

## Syntax errors

| Code | Meaning |
|---|---|
| `ASX-PARSE-001` | Something unexpected, e.g. a missing `;`, a missing `package`, an AS3-style `function` keyword, a missing initializer |
| `ASX-PARSE-002` | A reserved word for a feature that isn't implemented yet (`class`, `float`, …) |

The parser stops at the first syntax error.

## Name errors

| Code | Meaning |
|---|---|
| `ASX-NAME-001` | Undefined variable or function |
| `ASX-NAME-002` | Name already declared: redeclaration, shadowing, or a duplicate function |
| `ASX-NAME-003` | Calling something that is a variable, not a function |
| `ASX-NAME-004` | Using a function's name as a value without calling it |
| `ASX-NAME-005` | Declaring a function named `trace` |
| `ASX-NAME-006` | Assigning to a parameter |

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
| `ASX-TYPE-012` | Operator not defined for these operand types |
| `ASX-TYPE-013` | Condition isn't `bool` |
| `ASX-TYPE-014` | Division by a constant zero |

## Other compile-time errors

| Code | Meaning |
|---|---|
| `ASX-FLOW-001` | `break` or `continue` outside a loop |
| `ASX-ENTRY-001` | No `main` function |
| `ASX-ENTRY-002` | `main` has parameters, or returns something other than `int`/`void` |

## Runtime errors

These stop the program with exit status 1. Anything traced before the
error has already been printed.

| Code | Meaning |
|---|---|
| `ASX-RUN-001` | Calls nested more than 10,000 deep (usually runaway recursion) |
| `ASX-RUN-002` | Integer overflow: the result doesn't fit its type |
| `ASX-RUN-003` | Integer division or remainder by zero |

Next: [Coming from ActionScript 3](09-coming-from-actionscript.md)
