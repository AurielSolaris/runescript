# 2. Program structure

## Packages

Every file starts with a package declaration. It's a single statement ending
in `;`, not a block:

```actionscript
package games.puzzle.solver;

protected int main() { return 0; }
```

Package names are dotted identifiers. A program can span several files
(see [Packages and files](12-packages-and-files.md)).

## Using packages: `use`

`use` declarations come right after the `package` line, before any
function. Standard library packages are written `"std:<name>"`:

```actionscript
package main;

use "std:string";

protected int main() {
    trace(string.toUpper("hello"));   // HELLO
    return 0;
}
```

The functions are then called through the package's qualifier (the last
part of its name). See [Standard library: std:string](10-std-string.md).

`use` also loads other RuneScript packages (`use "games.engine";`) and
files (`use "./lib/helpers.rsc";`), so a program can span several files.
See [Packages and files](12-packages-and-files.md). Go packages
(`use "go:<package>";`) are planned and give `ASX-USE-003` today.

## Comments

```actionscript
package main;

// A line comment runs to the end of the line.

/* A block comment
   can span several lines. */

protected int main() {
    trace("comments are ignored"); // trailing comments too
    return 0;
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
- Like every function, `main` starts with an access modifier: `public`,
  `protected`, or `private` (see [Functions](06-functions.md)).

```actionscript
package main;

protected int main() {
    trace("no return value needed");
    return 0;
}
```

No class is needed. Programs are made of top-level functions, which can
appear in any order. Functions can also be nested inside other functions.

## Names

Identifiers are made of ASCII letters, digits, and `_`, and can't start
with a digit. They are case-sensitive. Keywords such as `if`, `int`, `use`,
and `return` can't be used as names. Words reserved for future features
(`class`, `array`, and `matrix`) can't be used either.

Next: [Types and variables](03-types-and-variables.md)
