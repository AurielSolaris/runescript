# 13. Floats and conversions

## Floating-point numbers

RuneScript has three floating-point types:

| Type | What it is |
|---|---|
| `float` | The everyday float. It starts as a 32-bit float and widens to 64 bits whenever a value needs it, so it never loses a value. |
| `float32` | Exactly an IEEE 32-bit float. Every result is rounded to 32 bits. |
| `float64` | Exactly an IEEE 64-bit float, like ActionScript's `Number`. |

```actionscript
package main;

public void main() {
    float a = 0.1;
    float b = 0.2;
    trace(a + b);          // 0.30000000000000004
    trace(7.0 / 2, 1e21);  // 3.5 1e+21

    float32 f = (float32) 0.1;
    trace(f * 3);          // 0.3, rounded to float32
}
```

Because `float` widens instead of rounding, a `float` computation always
gives the same result as 64-bit arithmetic. Use `float32` when you want
32-bit rounding.

Literals are written `64.3`, `1e10`, or `2.5e-3`. Numbers print the way
ActionScript prints them: `64.0` prints as `64`, and very large or very
small numbers use exponent form.

### `NaN` and `Infinity`

Floats follow IEEE 754, as ActionScript's `Number` does. A result too large
for its type is `Infinity`, and `0 / 0` is `NaN` ("not a number"). Unlike
integer overflow, these never stop the program.

```actionscript
package main;

public void main() {
    float zero = 0;
    float nan = zero / zero;
    trace(1 / zero, -1 / zero, nan);           // Infinity -Infinity NaN
    trace(nan == nan, isNaN(nan));             // false true
    trace(isFinite(1e308 * 10), Infinity > 1); // false true
}
```

`NaN` is not equal to anything, not even itself, so test for it with
`isNaN(x)`. `isFinite(x)` is false for `NaN` and both infinities. `NaN` and
`Infinity` are built-in constants; a variable of the same name hides them.

### No implicit conversions

Integers and floats don't mix, and neither do different float types.
Convert one side explicitly:

```actionscript
package main;

public void main() {
    int n = 3;
    float x = 1.5;
    trace(x * n);   // error: ASX-TYPE-012
}
```

```actionscript
package main;

public void main() {
    int n = 3;
    float x = 1.5;
    trace(x * (float) n);   // 4.5
}
```

Number literals adapt: in `x * 2`, the `2` becomes a float. A literal must
fit its type exactly, though. `float32 f = 0.1;` is an error (`ASX-TYPE-018`)
because 0.1 has no exact float32 value. Write `(float32) 0.1` to round it
on purpose.

`%`, `++`, and `--` work on integers only. Write `x += 1;` for a float.

## Conversions

`(T) value` converts a value to `T`, like a cast in C. `T` can be `int`,
`sint`, `uint`, `float`, `float32`, `float64`, `string`, or `bool`.

```actionscript
package main;

public void main() {
    trace((int) 64.3, (int) -64.7);       // 64 -64: truncates toward zero
    trace((float) "64.7", (int) "-42");   // 64.7 -42: parses strings
    trace((string) 1.5, (bool) "true");   // 1.5 true
    int count = 7;
    trace((float) count / 2);             // 3.5: the cast binds tighter than /
}
```

Every conversion is **checked**. It never wraps around or guesses. A value
that doesn't fit, or a string that isn't a number, throws an error:

| Conversion | When it fails | Error |
|---|---|---|
| to an integer | the value is out of range (e.g. `(uint) -1.5`, `(int) 3e10`), `NaN`, or infinite | `RangeError` |
| a string to a number | the string isn't a plain decimal number (or `"NaN"`, `"Infinity"`, `"-Infinity"` for floats), e.g. `"12x"`, `" 5"`, `"0x10"` | `TypeError` |
| a string to `bool` | the string isn't exactly `"true"` or `"false"` | `TypeError` |

```actionscript
package main;

public void main() {
    string input = "12x";
    trace((int) input);   // runtime error: ASX-RUN-006
}
```

When the value is a constant, the conversion happens while compiling, and
so does the error: `(uint) -2.5` is `ASX-TYPE-007`.

`bool` and numbers don't convert to each other. Compare instead
(`n != 0`).

The signed/unsigned rule from [chapter 3](03-types-and-variables.md) now
has its explicit form:

```actionscript
package main;

public void main() {
    int signedValue = 10;
    uint u = (uint) signedValue;   // checked: a negative value would throw
    trace(u);
}
```

Next: [Errors](14-errors.md)
