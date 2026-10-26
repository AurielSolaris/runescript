# 22. CSV, matrices, and versions

Three small packages for everyday data:

- `std:csv` reads and writes comma-separated values as
  `Array<Array<string>>`.
- `std:matrix` does linear algebra on `Array<Array<float64>>`.
- `std:semver` checks, compares, and matches version strings such as
  `1.4.0-rc.1`.

All three are plain functions over arrays and strings: there are no new
types to learn. Like the rest of the standard library, they give the same
results, and the same error messages, on the interpreter and on the
JavaScript target.

## CSV: `std:csv`

`csv.parse` splits a text into records, and each record into fields. A
field in double quotes may contain commas, line breaks, and quotes (written
twice):

```actionscript
package main;

use "std:csv";

protected int main() {
    string text = "name,city\nAda,London\n\"Turing, Alan\",\"Wilmslow\"\n";
    Array<Array<string>> rows = csv.parse(text);
    trace(rows.length);                      // 3
    for each (Array<string> row in rows) {
        trace(row[0], "|", row[1]);          // name | city, then Ada | London, ...
    }
    return 0;
}
```

To read a file, combine it with `std:fs`: `csv.parse(fs.readFile("data.csv"))`.

`csv.format` goes the other way. It quotes a field only when it has to, and
ends every record with a line break, so `csv.parse(csv.format(rows))` gives
back `rows`:

```actionscript
package main;

use "std:csv";

protected int main() {
    Array<Array<string>> rows = [["item", "note"], ["lamp", "red, tall"], ["rug", "says \"hi\""]];
    trace(csv.format(rows));
    // item,note
    // lamp,"red, tall"
    // rug,"says ""hi"""
    return 0;
}
```

`csv.parseWith` and `csv.formatWith` take another separator, as a `char`:
`'\t'` for tab-separated files, `';'` for the files many spreadsheets
write in Europe:

```actionscript
package main;

use "std:csv";

protected int main() {
    Array<Array<string>> rows = csv.parseWith("a\tb\n1\t2", '\t');
    trace(rows[1][1]);                       // 2
    trace(csv.formatWith(rows, ';'));        // a;b, then 1;2
    return 0;
}
```

The rules:

- Records end at `\n` or `\r\n`. Empty lines are skipped, so an empty
  text has no records.
- Records may have different numbers of fields. Check `row.length` if your
  data needs a fixed number.
- Spaces are part of a field: `" a "` stays `" a "`. Quoted fields keep
  their line breaks exactly.
- A UTF-8 byte order mark at the start, as some spreadsheets write, is
  skipped.
- A quote inside an unquoted field (`ab"c`), text after a closing quote
  (`"a"b`), and a quote that is never closed are errors. The message says
  where: `csv.parse: line 2, column 4: …`. Columns count bytes, like
  string offsets ([chapter 10](10-std-string.md)).
- The separator can be any character except `"`, `\r`, and `\n`
  (`ArgumentError`). `csv.format` throws an `ArgumentError` for a record
  with no fields, which CSV can't write.

## Matrices: `std:matrix`

A matrix is an array of rows: `[[1, 2], [3, 4]]` is
`Array<Array<float64>>`. Every function checks that the rows have the same
length and that the sizes fit the operation, and returns a new array: the
arguments are never changed.

```actionscript
package main;

use "std:matrix";

protected int main() {
    Array<Array<float64>> a = [[1, 2], [3, 4]];
    Array<Array<float64>> b = [[0, 1], [1, 0]];
    trace(matrix.format(matrix.mul(a, b)));  // [2, 1], then [4, 3]
    trace(matrix.format(matrix.add(a, b)));  // [1, 3], then [4, 4]
    trace(matrix.det(a));                    // -2
    trace(matrix.rows(a), matrix.cols(a));   // 2 2
    return 0;
}
```

| Function | Result |
|---|---|
| `zeros(rows, cols)`, `identity(n)` | a new matrix of zeros, or the identity |
| `rows(m)`, `cols(m)` | the size (a matrix with no rows has 0 columns) |
| `add(a, b)`, `sub(a, b)`, `scale(m, k)` | element by element |
| `mul(a, b)`, `mulVec(m, v)` | the matrix product, and a matrix times a vector (`Array<float64>`) |
| `transpose(m)` | rows become columns |
| `det(m)`, `inverse(m)` | the determinant and the inverse of a square matrix |
| `solve(a, b)` | the `x` with `a·x = b`, for a square `a` |
| `format(m)` | one line per row: `[1, 2.5]` |

`solve` is better than multiplying by the inverse: it does less work and
loses less precision. Here it finds the prices of two items from two
receipts:

```actionscript
package main;

use "std:matrix";

protected int main() {
    // 2 apples + 1 pear cost 4; 1 apple + 3 pears cost 7.
    Array<Array<float64>> counts = [[2, 1], [1, 3]];
    Array<float64> prices = matrix.solve(counts, [4.0, 7.0]);
    trace(prices[0], prices[1]);             // 1 2
    return 0;
}
```

`det`, `inverse`, and `solve` use Gaussian elimination with partial
pivoting. A singular matrix (one with no inverse) gives a determinant of
`0`, and `inverse` and `solve` throw an `Error`. Like all floating-point
work, results can be off in the last digits: `0.1 + 0.2` is
`0.30000000000000004`. They are off in exactly the same way on both
targets, because both run the same sequence of operations.

Size mistakes throw an `ArgumentError` that names the sizes:

```text
matrix.mul: cannot multiply a 1x2 matrix by a 1x2 matrix: the first has 2 columns, the second 1 row
```

`matrix` is still a reserved word: you can't name a variable `matrix`.
It can only qualify a call, as in `matrix.mul(a, b)`.

## Versions: `std:semver`

`std:semver` works with [Semantic Versioning 2.0.0](https://semver.org/)
strings: `MAJOR.MINOR.PATCH`, optionally followed by a prerelease
(`-rc.1`) and build metadata (`+build.5`). It compares versions and
matches ranges the way npm does.

Comparing versions as strings gives wrong answers, because `"1.10.0"`
would sort before `"1.9.0"` letter by letter. The `semver` functions
compare numbers as numbers, and put a prerelease before its release:

```actionscript
package main;

use "std:semver";

protected int main() {
    trace(semver.lt("1.9.0", "1.10.0"));             // true
    trace(semver.lt("1.0.0-rc.1", "1.0.0"));         // true
    trace(semver.compare("2.0.0", "1.99.0"));        // 1
    trace(semver.eq("1.0.0+build.1", "1.0.0"));      // true: build metadata is ignored
    trace(semver.valid("1.2"), semver.valid("1.2.3-rc.1")); // false true
    return 0;
}
```

| Function | Result |
|---|---|
| `valid(v)`, `validRange(r)` | whether `v` is a version, or `r` a range (never throws) |
| `compare(a, b)` | `-1`, `0`, or `1` |
| `eq`, `neq`, `gt`, `gte`, `lt`, `lte` | `bool` comparisons |
| `satisfies(v, range)` | whether `v` is in the range |
| `major(v)`, `minor(v)`, `patch(v)` | the numbers, as `int` |
| `prerelease(v)`, `build(v)` | `"rc.1"`, `"build.5"`, or `""` |

`satisfies` understands npm's ranges:

```actionscript
package main;

use "std:semver";

protected int main() {
    trace(semver.satisfies("1.4.0", "^1.2.0"));            // true: 1.x from 1.2.0 on
    trace(semver.satisfies("2.0.0", "^1.2.0"));            // false
    trace(semver.satisfies("1.2.9", "~1.2.3"));            // true: 1.2.x from 1.2.3 on
    trace(semver.satisfies("1.5.0", ">=1.2.0 <2.0.0"));    // true
    trace(semver.satisfies("3.1.0", "1.x || >=3.0.0"));    // true
    trace(semver.satisfies("2.3.9", "1.2.3 - 2.3"));       // true
    trace(semver.satisfies("1.3.0-beta", "^1.2.0"));       // false: see below
    return 0;
}
```

| Range | Means |
|---|---|
| `1.2.3`, `=1.2.3` | exactly that version |
| `>1.2.3`, `>=`, `<`, `<=` | comparisons; several separated by spaces must all hold |
| `a \|\| b` | either range |
| `1.2.x`, `1.x`, `*`, `1.2` | any version matching the given parts |
| `~1.2.3` | `>=1.2.3 <1.3.0`: patch updates |
| `^1.2.3` | `>=1.2.3 <2.0.0`: no breaking changes (`^0.2.3` is `<0.3.0`) |
| `1.2.3 - 2.3.4` | `>=1.2.3 <=2.3.4` |

As in npm, a prerelease version satisfies a range only if the range names
a prerelease of the same `MAJOR.MINOR.PATCH`: `1.3.0-beta` is not in
`^1.2.0`, but `1.2.3-beta.4` is in `^1.2.3-beta.2`. That way, testing
versions are never picked by accident.

The rules:

- Versions are strict SemVer 2.0.0: no leading zeros (`01.2.3`), no `v`
  prefix (`v1.2.3`), no spaces. Numbers may have any number of digits;
  `major`, `minor`, and `patch` throw a `RangeError` if one doesn't fit
  in `int`.
- Every function except `valid` and `validRange` throws an
  `ArgumentError` for an invalid version or range:
  `semver.compare: "1.2" is not a valid version: …`.

Next: [Editors](23-editors.md)
