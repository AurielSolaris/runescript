# 2. Program structure

## Packages

Every file starts with a package declaration. It's a single statement ending
in `;`, not a block:

```actionscript
package games.puzzle.solver;

void main() {
}
```

Package names are dotted identifiers. In v0.1, each program is a single
file.

> **Planned, not yet implemented:** other packages and files will be brought
> in with `use`:
>
> ```text
> use "rune.gfx";          // a package
> use "lib/helpers.rs";    // a file, relative to the current directory
> ```
>
> `use` is already reserved. Writing it today gives `ASX-PARSE-002`.

## Comments

```actionscript
package main;

// A line comment runs to the end of the line.

/* A block comment
   can span several lines. */

void main() {
    trace("comments are ignored"); // trailing comments too
}
```

Block comments don't nest.

## Statements and blocks

- Every statement ends with `;`.
- Braces `{ }` mark the body of a function, `if`, `else`, `while`, or `for`.
  In v0.1 these braces are **required**, even around a single statement.

## The entry point: `main`

Running a program calls its top-level `main` function. `main` takes no
parameters and returns either `int` or `void`:

```actionscript
package main;

protected int main() {
    // do stuff
    return 0;
}
```

- If `main` returns `int`, the value becomes the process exit status.
- If `main` is `void`, the exit status is 0.
- A visibility modifier (`public`, `protected`, `private`) is optional. In
  v0.1 the modifier is recorded but has no effect yet.

```actionscript
package main;

void main() {
    trace("no return value needed");
}
```

No class is needed. Programs are made of top-level functions, which can
appear in any order.

## Names

Identifiers are made of ASCII letters, digits, and `_`, and can't start
with a digit. They are case-sensitive. Keywords such as `if`, `int`, and
`return` can't be used as names. Words reserved for future features
(`class`, `struct`, `use`, `float`, `char`, `array`, `matrix`, and the sized types
like `int64`) can't be used either.

Next: [Types and variables](03-types-and-variables.md)
