# 3. Types and variables

## Declaring variables

Write the type first, then the name, then an initial value:

```actionscript
package main;

public void main() {
    int count = 10;
    uint total = 250;
    bool ready = true;
    string name = "Ada";
    trace(name, count, total, ready);   // Ada 10 250 true
}
```

Every variable **must** be initialised when it's declared. There are no
implicit default values in v0.1.

## The basic types

| Type | Holds | Range |
|---|---|---|
| `int` | Signed whole numbers | −2147483648 … 2147483647 |
| `sint` | **Exactly the same type as `int`.** Use it when you want to say "signed" explicitly. | same as `int` |
| `uint` | Unsigned whole numbers | 0 … 4294967295 |
| `float`, `float32`, `float64` | Floating-point numbers ([chapter 13](13-floats-and-conversions.md)) | |
| `bool` | `true` or `false` | |
| `string` | Text, e.g. `"hello\n"` | |
| `Error` and its kinds | Errors ([chapter 14](14-errors.md)) | |
| `void` | No value; used only as a function's return type | |

`int` and `sint` can be mixed freely:

```actionscript
package main;

public void main() {
    sint a = -5;
    int b = a;      // fine: same type
    sint c = b;     // fine
    trace(a, b, c);
}
```

String escapes: `\n` (newline), `\t` (tab), `\r`, `\\`, and `\"`.

## Assignment

```actionscript
package main;

public void main() {
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

public void main() {
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
}
```

v0.1 doesn't allow **shadowing**. A new variable can't reuse the name of a
visible variable, parameter, or function:

```actionscript
package main;

public void main() {
    int x = 1;
    if (true) {
        int x = 2; // error: ASX-NAME-002
    }
}
```

## Signed and unsigned never mix silently

RuneScript doesn't silently turn a signed integer into an unsigned one. The
compiler checks the *type* of the value, not the value itself. A signed
variable could hold a negative number, so storing it in a `uint` is an
error even when it happens to be positive:

```actionscript
package main;

public int main() {
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

public void main() {
    uint a = 42;            // fine: 42 fits in uint
    uint b = 6 * 7;         // fine: constants are computed at compile time
    int c = -2147483648;    // fine: the smallest int
    trace(a, b, c);
}
```

```actionscript
package main;

public void main() {
    uint u = -1;            // error: ASX-TYPE-007
}
```

```actionscript
package main;

public void main() {
    int big = 3000000000;   // error: ASX-TYPE-006
}
```

The reverse direction (`uint` into `int`) is also rejected, as is mixing
`int` and `uint` in one operation. To convert, write the conversion out:
`(uint) signedValue` checks the value when it runs and throws an error if it
is negative. See [Floats and conversions](13-floats-and-conversions.md),
which also covers `float`, `float32`, and `float64`.

Next: [Operators](04-operators.md)
