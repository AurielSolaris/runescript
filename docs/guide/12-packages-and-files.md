# 12. Packages and files

A program can be split into several files with `use`. There are three kinds
of use:

```text
use "std:string";          // a standard library package
use "geom";                // a RuneScript package: every .rsm file in geom/
use "./util/text.txt";     // one file, whatever its extension
```

A string with a `/` in it is a file path. Anything else is a package name,
such as `geom` or `org.games.engine`. Paths and package directories are
relative to the **directory you run `runescript` from**, not to the file
that contains the `use`.

## An example

This program is in the test suite at `tests/testdata/multi/ok_packages`.
Run it from that directory with `runescript run main.rsc`.

```text
ok_packages/
├── main.rsc            package app;   the entry file
├── helpers.rsc         package app;   another file of the same package
├── geom/
│   ├── area.rsm        package geom;
│   └── perimeter.rsm   package geom;
└── util/
    └── text.txt        package text;  any extension works
```

`main.rsc`:

```text
package app;

use "geom";
use "./util/text.txt" as t;
use "./helpers.rsc";

public int main() {
    trace(geom.area(3, 4), geom.perimeter(3, 4));
    trace(t.shout("hi"));
    trace(twice(21));
    return 0;
}
```

`geom/perimeter.rsm`:

```text
package geom;

use "./util/text.txt";

public int perimeter(int w, int h) {
    trace(text.shout("perimeter"));
    return scale(2 * (w + h));
}

// Protected: shared by both files of package geom.
protected int scale(int n) {
    return n * one();
}

// Private to this file.
private int one() {
    return 1;
}
```

It prints:

```text
PERIMETER
12 14
HI
42
```

## Packages

A package is every file that declares the same `package` name.

- `use "geom";` loads the `.rsm` files directly in the directory `geom/`.
  `use "org.games.engine";` loads them from `org/games/engine/`. Other files
  in the directory are ignored, and so are subdirectories. Each `.rsm` file
  must declare that package (`ASX-USE-007`).
- Files of the same package see each other's `public` and `protected`
  functions without qualifying them or writing prototypes. In the example,
  `area.rsm` calls `scale`, which is defined in `perimeter.rsm`.
- A file that you `use` by path joins whichever package it declares. In the
  example, `helpers.rsc` declares `package app;`, so its `twice` is called
  directly from `main.rsc`.

Every file is loaded once, however many files use it. Both `main.rsc` and
`perimeter.rsm` use `util/text.txt`, and it's loaded once.

## Qualifiers

Another package's functions are called through a **qualifier**. By default
this is the last part of the package's name: `geom` for `geom`, and `engine`
for `org.games.engine`. For a file, it's the package the file declares:
`text` for `util/text.txt`. To choose another name, use `as`:

```text
use "./util/text.txt" as t;     // t.shout("hi")
use "org.games.engine" as eng;  // eng.start()
```

A qualifier belongs to the file it's declared in. Every file that calls
`geom.area` needs its own `use "geom";`.

If two uses in one file would have the same qualifier, that's `ASX-USE-008`.
Rename one of them with `as`.

A qualified function is also a value: `int(int, int) f = geom.area;`
passes `geom.area` without calling it, as `dom.renderApp(app.App, "#app")`
does with a component ([chapter 15](15-web-pages.md)). Standard library
functions can only be called so far: wrap them in a lambda, e.g.
`(string s) => string.toUpper(s)`.

## Visibility

Every function's access modifier is now enforced:

| Modifier | Who can call it |
|---|---|
| `public` | Every file of every package (other packages use a qualifier) |
| `protected` | Every file of its own package |
| `private` | Its own file only |

Calling a function its modifier doesn't allow is `ASX-NAME-011`, and the
message says where the function is. Two files may each have their own
`private` function with the same name. Any other repeated top-level name in
a package is `ASX-NAME-002`.

## The entry point

A program has exactly one `main`, and it can be in any of its files. The
file you pass to `runescript run` doesn't have to contain it. More than one
`main` is `ASX-ENTRY-003`.

## Cycles

Packages can't use each other in a cycle. If `a` uses `b` and `b` uses `a`,
directly or through other packages, that's `ASX-USE-005`, and the message
shows the cycle: `import cycle: a → b → a`. To fix it, move the functions
they share into a third package.

## File names and extensions

By convention, programs are `.rsc` files and package directories hold `.rsm`
files. As in Perl, the extension isn't checked when you name a file
yourself: `runescript run tool.txt` and `use "./lib/x.inc";` both work, as
long as the file contains RuneScript. Only package directories look at the
extension, so that READMEs and data files next to them are ignored.

Diagnostics name the file they're in, relative to the working directory:

```text
calc/square.rsm:4:14: error ASX-RUN-002: integer overflow: 65536 * 65536 = 4294967296 does not fit in 'int'
```

## Errors

| Code | Cause |
|---|---|
| `ASX-USE-004` | A missing file, or a package directory with no `.rsm` files |
| `ASX-USE-005` | Packages that use each other in a cycle |
| `ASX-USE-006` | An invalid package name, or a path written with `\` |
| `ASX-USE-007` | A `.rsm` file that declares a different package from its directory |
| `ASX-USE-008` | Two uses with the same qualifier in one file |
| `ASX-NAME-011` | A `private` or `protected` function used from outside |
| `ASX-ENTRY-003` | More than one `main` |

Next: [Floats and conversions](13-floats-and-conversions.md)
