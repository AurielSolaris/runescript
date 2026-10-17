# 3. Types and variables

## Declaring variables

Write the type first, then the name, then an initial value:

```actionscript
package main;

protected int main() {
    int count = 10;
    uint total = 250;
    bool ready = true;
    string name = "Ada";
    trace(name, count, total, ready);   // Ada 10 250 true
    return 0;
}
```

A declaration without an initial value starts at the type's **zero
value**:

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

protected int main() {
    int count;            // 0
    float ratio;          // 0
    string name;          // ""
    bool ready;           // false
    char initial;         // the character U+0000
    Array<string> names;  // a new, empty array
    Point origin;         // Point{x: 0, y: 0}: every field at its zero value
    trace(count, ratio, "[" + name + "]", ready, names.length, origin);
    return 0;                            // 0 0 [] false 0 Point{x: 0, y: 0}
}
```

Function types, errors, and runes have no zero value: RuneScript has no
`null` to stand for "not set yet" ([chapter 17](17-no-null-any-generics.md)).
Declaring one without a value is an error (`ASX-TYPE-028`):

```actionscript
package main;

protected int main() {
    int(int) twice; // error: ASX-TYPE-028
    return 0;
}
```

## The basic types

| Type | Holds | Range |
|---|---|---|
| `int` | Signed whole numbers | −2147483648 … 2147483647 |
| `sint` | **Exactly the same type as `int`.** Use it when you want to say "signed" explicitly. | same as `int` |
| `uint` | Unsigned whole numbers | 0 … 4294967295 |
| `int8` … `int64`, `uint8` … `uint64` | Whole numbers of an exact width ([chapter 18](18-numbers.md)) | e.g. `uint8` is 0 … 255 |
| `float`, `float32`, `float64` | Floating-point numbers ([chapter 13](13-floats-and-conversions.md)) | |
| `bool` | `true` or `false` | |
| `string` | Text, e.g. `"hello\n"` | |
| `char` | One character, e.g. `'a'` ([chapter 16](16-text-and-arrays.md)) | any Unicode code point |
| `Array<T>` | A list of `T` values, e.g. `["a", "b"]` ([chapter 16](16-text-and-arrays.md)) | |
| a `struct` | Named fields in one value, e.g. `Point{x: 1, y: 2}` ([chapter 19](19-structs.md)) | |
| `Error` and its kinds | Errors ([chapter 14](14-errors.md)) | |
| `void` | No value; used only as a function's return type | |

`int` and `sint` can be mixed freely:

```actionscript
package main;

protected int main() {
    sint a = -5;
    int b = a;      // fine: same type
    sint c = b;     // fine
    trace(a, b, c);
    return 0;
}
```

String escapes: `\n` (newline), `\t` (tab), `\r`, `\0`, `\\`, `\"`, `\'`, and
`\u{...}` for any Unicode character, e.g. `\u{E9}` (é) or `\u{1F600}`. Strings are
joined with `+`, and `$"total: {n}"` builds one from values
([chapter 16](16-text-and-arrays.md)).

## Assignment

```actionscript
package main;

protected int main() {
    int x = 1;
    x = 10;     // 10
    x += 5;     // 15
    x -= 3;     // 12
    x *= 2;     // 24
    x /= 5;     // 4
    x %= 3;     // 1
    x++;        // 2
    x--;        // 1
    trace(x);
    return 0;
}
```

- Assignment is a statement. You can't write `a = b = 0;` or
  `if (x = 1)`.
- `++` and `--` are statements that come after a variable (`i++;`). They
  can't be used inside a larger expression.
- Function parameters are read-only. To modify one, copy it into a local
  variable first.

## Scope

A variable exists from its declaration to the end of the enclosing block.
A variable declared in a `for` header exists only inside that loop.

```actionscript
package main;

protected int main() {
    if (true) {
        int inner = 1;
        trace(inner);
    }
    // 'inner' is not visible here

    for (int i = 0; i < 2; i++) {
        trace(i);
    }
    for (int i = 0; i < 2; i++) {   // fine: the first 'i' is gone
        trace(i);
    }
    return 0;
}
```

A variable in a **nested block** may reuse the name of a variable or
parameter from an enclosing block. This is called **shadowing**: inside the
block, the name means the new variable, and the outer one comes back when
the block ends. The compiler warns about it (`ASX-WARN-009`), because it's
easy to do by accident:

```actionscript
package main;

protected int main() {
    int x = 1;
    if (true) {
        int x = x + 10;   // warning ASX-WARN-009; the initializer still reads the outer x
        trace(x);         // 11
    }
    trace(x);             // 1
    return 0;
}
```

Reusing a name **in the same block** is still an error, and so is reusing
the name of a function, a struct, or a parameter in the function's own
outermost block:

```actionscript
package main;

protected int main() {
    int x = 1;
    int x = 2; // error: ASX-NAME-002
    return 0;
}
```

## Signed and unsigned never mix silently

RuneScript doesn't silently turn a signed integer into an unsigned one. The
compiler checks the *type* of the value, not the value itself. A signed
variable could hold a negative number, so storing it in a `uint` is an
error even when it happens to be positive:

```actionscript
package main;

protected int main() {
    sint signedValue = -10;
    uint unsignedValue = signedValue; // error: ASX-TYPE-004
    return 0;
}
```

```text
main.rsc:5:26: error ASX-TYPE-004: Cannot assign signed integer 'sint' to unsigned integer 'uint'.
        uint unsignedValue = signedValue;
                             ^^^^^^^^^^^
  help: Use an explicit checked conversion, (uint) x, or change the destination type.
```

Number literals are more flexible. A literal takes the type of whatever
receives it, as long as the value fits:

```actionscript
package main;

protected int main() {
    uint a = 42;            // fine: 42 fits in uint
    uint b = 6 * 7;         // fine: constants are computed at compile time
    int c = -2147483648;    // fine: the smallest int
    trace(a, b, c);
    return 0;
}
```

```actionscript
package main;

protected int main() {
    uint u = -1;            // error: ASX-TYPE-007
    return 0;
}
```

```actionscript
package main;

protected int main() {
    int big = 3000000000;   // error: ASX-TYPE-006
    return 0;
}
```

The reverse direction (`uint` into `int`) is also rejected, as is mixing
`int` and `uint` in one operation. To convert, write the conversion out:
`(uint) signedValue` checks the value when it runs and throws an error if it
is negative. See [Floats and conversions](13-floats-and-conversions.md),
which also covers `float`, `float32`, and `float64`.

Next: [Operators](04-operators.md)
