# 16. Text and lists: joining strings, `char`, and arrays

This chapter covers the everyday tools of scripting: building strings,
working with single characters, and keeping values in a list. The functions
that search, split, and change strings are in
[chapter 10](10-std-string.md).

## Joining strings with `+`

`+` joins two strings, and `+=` adds to a string variable:

```actionscript
package main;

protected int main() {
    string name = "Ada";
    string greeting = "Hello, " + name + "!";
    greeting += " Welcome.";
    trace(greeting);   // Hello, Ada! Welcome.
    return 0;
}
```

`+` never converts a number to text for you, because RuneScript doesn't
convert values behind your back. Convert it yourself with `(string) x`, or
use an interpolated string (next section):

```actionscript
package main;

protected int main() {
    int count = 3;
    trace("count: " + (string) count);   // count: 3
    trace("count: " + count);            // error: ASX-TYPE-012
    return 0;
}
```

This differs from ActionScript 3 and JavaScript, where `"count: " + 3`
quietly becomes `"count: 3"`.

## Interpolated strings: `$"…"`

Put `$` before a string, and every `{expression}` in it is replaced by the
expression's value:

```actionscript
package main;

protected int main() {
    string item = "apple";
    int count = 3;
    float price = 0.25;
    trace($"{count} x {item} at {price} = {(float) count * price}");   // 3 x apple at 0.25 = 0.75
    trace($"ready: {count > 0}");                                     // ready: true
    trace($"{{braces}} are written twice");                           // {braces} are written twice
    return 0;
}
```

- A hole can hold any expression: arithmetic, calls, other strings, even
  another interpolated string.
- Its value is shown exactly as `(string) x` would show it, so it must be a
  number, a `bool`, a `char`, or a string. Anything else, such as a
  function or an array, is `ASX-TYPE-017`.
- `{{` and `}}` stand for a literal `{` and `}`. A single `}` is an error
  (`ASX-LEX-003`).
- Escapes work as in other strings: `\n`, `\t`, `\"`, and so on.

## Characters: `char`

A `char` is one character, a Unicode code point. Write it between single
quotes:

```actionscript
package main;

protected int main() {
    char letter = 'a';
    char accent = 'é';
    char emoji = '😀';
    char quote = '\'';
    trace(letter, accent, emoji, quote, '\\');   // a é 😀 ' \
    trace(letter < 'b', accent == 'é');          // true true
    return 0;
}
```

- A character literal holds exactly one character. `''` and `'ab'` are
  `ASX-LEX-006`. Escapes are those of strings, plus `\'`.
- Chars can be compared with `==`, `!=`, `<`, `<=`, `>`, and `>=`, by code
  point.
- A `char` is not a number, so there's no `'a' + 1`. Convert it first.

Conversions are checked, like every conversion
([chapter 13](13-floats-and-conversions.md)):

| Conversion | Result | Fails with |
|---|---|---|
| `(int) c`, `(uint) c` | the code point: `(int) 'a'` is `97` | never |
| `(char) n` | the character with code point `n`: `(char) 98` is `b` | `RangeError` if `n` isn't a valid code point (negative, above `0x10FFFF`, or a surrogate `0xD800`–`0xDFFF`) |
| `(string) c` | a one-character string | never |
| `(char) s` | the one character of `s` | `TypeError` unless `s` has exactly one character |

```actionscript
package main;

public char next(char c) {
    return (char) ((int) c + 1);
}

protected int main() {
    trace(next('a'), (int) 'A', (char) "é");   // b 65 é
    trace((string) 'x' + "yz");               // xyz
    return 0;
}
```

Chars are what the character functions of `std:string` take: see
[Characters and callbacks](10-std-string.md#characters-and-callbacks).

## Arrays: `Array<T>`

An `Array<T>` is a list of values of type `T` that can grow and shrink.
Create one with square brackets:

```actionscript
package main;

protected int main() {
    Array<string> names = ["Ada", "Grace"];
    names.push("Linus");                      // add to the end
    trace(names.length, names[0], names[2]);  // 3 Ada Linus

    names[1] = "Hopper";                      // replace an element
    names[0] += "!";
    string last = names.pop();                // remove the last one
    trace(last, names.length, names[1]);      // Linus 2 Hopper
    return 0;
}
```

| Operation | Meaning |
|---|---|
| `[a, b, c]` | a new array; `[]` is an empty one |
| `a[i]` | the element at index `i` (an `int` or `uint`), counting from `0` |
| `a[i] = v`, `a[i] += v`, … | replace an element |
| `a.length` | the number of elements |
| `a.push(v)` | add `v` at the end |
| `a.pop()` | remove the last element and return it |

The element type `T` can be any value type: numbers, `bool`, `string`,
`char`, functions, errors, and other arrays (`Array<Array<int>>`).

**The type comes from where the array goes.** `Array<uint> u = [1, 2];`
makes the `1` and `2` `uint`s. With no declared type around it, a literal
takes the type of its first element, so `[1, 2]` is an `Array<int>`. An
empty `[]` needs a declared type: `Array<string> names = [];`.

**Indexes are checked.** An index below `0` or not below `length`, and
`pop()` on an empty array, throw a `RangeError` (`ASX-RUN-011`) that you
can catch:

```actionscript
package main;

protected int main() {
    Array<int> scores = [90, 75];
    try {
        trace(scores[2]);
    } catch (RangeError e) {
        trace(e.message);   // index 2 is out of range for an array of length 2
    }
    return 0;
}
```

**Loop over an array** with an index:

```actionscript
package main;

public int sum(Array<int> values) {
    int total = 0;
    for (int i = 0; i < values.length; i++) {
        total += values[i];
    }
    return total;
}

protected int main() {
    trace(sum([3, 4, 5]));   // 12
    return 0;
}
```

**Or loop with `for each`**, which visits every element in order. The
variable before `in` is a new variable that holds one element at a time:

```actionscript
package main;

protected int main() {
    Array<string> names = ["Ada", "Grace", "Linus"];
    for each (string name in names) {
        trace("hello", name);
    }
    return 0;
}
```

The loop variable's type must match the elements' type: `for each (uint x in
ints)` is an error, as `uint x = ints[0];` would be. `break` and `continue`
work as in other loops. The array expression is evaluated once; the
length is checked before every step, so elements pushed inside the loop are
visited too. `each` and `in` are not reserved words: they only mean
something at this place, so `in` is still a valid variable name.

**Arrays are shared, not copied.** Assigning an array to another variable,
or passing it to a function, gives a second name for the *same* array, as
in ActionScript 3 and JavaScript. A change through one name shows through
all of them:

```actionscript
package main;

public void addDefaults(Array<string> list) {
    list.push("default");
}

protected int main() {
    Array<string> mine = ["first"];
    Array<string> same = mine;
    addDefaults(same);
    trace(mine.length, mine[1]);   // 2 default
    return 0;
}
```

Arrays can't be traced directly. Join them into one string with
`string.join` ([chapter 10](10-std-string.md#splitting-and-joining)):

```actionscript
package main;

use "std:string";

protected int main() {
    Array<string> words = string.fields("  the quick  brown fox ");
    trace(words.length, string.join(words, "-"));   // 4 the-quick-brown-fox
    return 0;
}
```

## Nested arrays

An array can hold arrays: `Array<Array<int>>` is a list of rows, and each
row is an `Array<int>` of its own. Index twice to reach an element. The
rows can have different lengths:

```actionscript
package main;

protected int main() {
    Array<Array<int>> grid = [[1, 2, 3], [4, 5], []];
    grid[1][0] = 40;
    grid[2].push(6);
    int total = 0;
    for each (Array<int> row in grid) {
        for each (int v in row) {
            total += v;
        }
    }
    trace(grid.length, grid[1][0], total);   // 3 40 57
    return 0;
}
```

The rows are arrays, so they are shared too: `Array<int> row = grid[0];`
names the same row as `grid[0]`. To build a table of a given size, push new
rows in a loop:

```actionscript
package main;

protected int main() {
    Array<Array<string>> table = [];
    for (int r = 0; r < 2; r++) {
        Array<string> row = [];
        for (int c = 0; c < 3; c++) {
            row.push($"{r},{c}");
        }
        table.push(row);
    }
    trace(table.length, table[0].length, table[1][2]);   // 2 3 1,2
    return 0;
}
```

## Array methods

| Method | Result | Meaning |
|---|---|---|
| `a.insert(i, v)` | `void` | put `v` before index `i` (`0` to `length`) |
| `a.removeAt(i)` | `T` | remove the element at `i` and return it |
| `a.slice(start, end)` | `Array<T>` | a new array with the elements from `start` up to, not including, `end` |
| `a.indexOf(v)` | `int` | the index of the first element equal to `v`, or `-1` |
| `a.contains(v)` | `bool` | whether an element equals `v` |
| `a.reverse()` | `void` | reverse the order, in place |
| `a.sort()` | `void` | sort in place, smallest first |
| `a.sort(less)` | `void` | sort in place, by a function `bool less(T x, T y)` |

```actionscript
package main;

use "std:string";

protected int main() {
    Array<string> fruit = ["pear", "fig", "apple"];
    fruit.insert(1, "kiwi");
    trace(string.join(fruit, " "));             // pear kiwi fig apple
    trace(fruit.indexOf("fig"), fruit.contains("plum"));   // 2 false
    string gone = fruit.removeAt(0);
    Array<string> firstTwo = fruit.slice(0, 2);
    trace(gone, string.join(firstTwo, " "));    // pear kiwi fig
    fruit.sort();
    trace(string.join(fruit, " "));             // apple fig kiwi
    fruit.sort((string x, string y) => string.len(x) > string.len(y));
    trace(string.join(fruit, " "));             // apple kiwi fig
    return 0;
}
```

- **Bad indexes throw.** `insert` accepts `0` to `length`, `removeAt` an
  existing index, and `slice` needs `0 <= start <= end <= length`;
  anything else throws a `RangeError`.
- **`slice` copies.** Changing the new array doesn't change the old one.
- **`indexOf` and `contains` compare with `==`**, so they work on the
  types `==` works on: numbers, `bool`, `string`, `char`, and structs of
  those. They are not available on arrays of arrays, functions, or errors.
- **`sort()` without a function** works on numbers, `char`s, and strings.
  Strings are ordered by their UTF-8 bytes (`string.compare`), so `"Z"`
  comes before `"a"`.
- **`sort(less)`** works on any element type. `less(x, y)` returns `true`
  when `x` belongs before `y`. The sort is *stable*: elements that are
  neither less nor greater keep their order. If `less` throws, the array
  is left as it was.

> **Not yet available:** `a[i]++`, comparing arrays with `==`, tracing an
> array directly, and passing arrays to JavaScript `extern` functions.
> `main` can't take command-line arguments; use `process.args()`
> ([chapter 20](20-files-and-input.md)).

Next: [No `null`, no `any`, no generics](17-no-null-any-generics.md)
