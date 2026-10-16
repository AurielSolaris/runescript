# 16. Text and lists: joining strings, `char`, and arrays

This chapter covers the everyday tools of scripting: building strings,
working with single characters, and keeping values in a list. The functions
that search, split, and change strings are in
[chapter 10](10-std-string.md).

## Joining strings with `+`

`+` joins two strings, and `+=` adds to a string variable:

```actionscript
package main;

public void main() {
    string name = "Ada";
    string greeting = "Hello, " + name + "!";
    greeting += " Welcome.";
    trace(greeting);   // Hello, Ada! Welcome.
}
```

`+` never converts a number to text for you, because RuneScript doesn't
convert values behind your back. Convert it yourself with `(string) x`, or
use an interpolated string (next section):

```actionscript
package main;

public void main() {
    int count = 3;
    trace("count: " + (string) count);   // count: 3
    trace("count: " + count);            // error: ASX-TYPE-012
}
```

This differs from ActionScript 3 and JavaScript, where `"count: " + 3`
quietly becomes `"count: 3"`.

## Interpolated strings: `$"…"`

Put `$` before a string, and every `{expression}` in it is replaced by the
expression's value:

```actionscript
package main;

public void main() {
    string item = "apple";
    int count = 3;
    float price = 0.25;
    trace($"{count} x {item} at {price} = {(float) count * price}");   // 3 x apple at 0.25 = 0.75
    trace($"ready: {count > 0}");                                     // ready: true
    trace($"{{braces}} are written twice");                           // {braces} are written twice
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

public void main() {
    char letter = 'a';
    char accent = 'é';
    char emoji = '😀';
    char quote = '\'';
    trace(letter, accent, emoji, quote, '\\');   // a é 😀 ' \
    trace(letter < 'b', accent == 'é');          // true true
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

public void main() {
    trace(next('a'), (int) 'A', (char) "é");   // b 65 é
    trace((string) 'x' + "yz");               // xyz
}
```

Chars are what the character functions of `std:string` take: see
[Characters and callbacks](10-std-string.md#characters-and-callbacks).

## Arrays: `Array<T>`

An `Array<T>` is a list of values of type `T` that can grow and shrink.
Create one with square brackets:

```actionscript
package main;

public void main() {
    Array<string> names = ["Ada", "Grace"];
    names.push("Linus");                      // add to the end
    trace(names.length, names[0], names[2]);  // 3 Ada Linus

    names[1] = "Hopper";                      // replace an element
    names[0] += "!";
    string last = names.pop();                // remove the last one
    trace(last, names.length, names[1]);      // Linus 2 Hopper
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

public void main() {
    Array<int> scores = [90, 75];
    try {
        trace(scores[2]);
    } catch (RangeError e) {
        trace(e.message);   // index 2 is out of range for an array of length 2
    }
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

public void main() {
    trace(sum([3, 4, 5]));   // 12
}
```

**Arrays are shared, not copied.** Assigning an array to another variable,
or passing it to a function, gives a second name for the *same* array, as
in ActionScript 3 and JavaScript. A change through one name shows through
all of them:

```actionscript
package main;

public void addDefaults(Array<string> list) {
    list.push("default");
}

public void main() {
    Array<string> mine = ["first"];
    Array<string> same = mine;
    addDefaults(same);
    trace(mine.length, mine[1]);   // 2 default
}
```

Arrays can't be traced directly. Join them into one string with
`string.join` ([chapter 10](10-std-string.md#splitting-and-joining)):

```actionscript
package main;

use "std:string";

public void main() {
    Array<string> words = string.fields("  the quick  brown fox ");
    trace(words.length, string.join(words, "-"));   // 4 the-quick-brown-fox
}
```

> **Not yet available:** a for-each loop, slicing, sorting, `a[i]++`,
> comparing arrays, and passing arrays to JavaScript `extern` functions.
> `main` can't take command-line arguments yet.

Next: [No `null`, no `any`, no generics](17-no-null-any-generics.md)
