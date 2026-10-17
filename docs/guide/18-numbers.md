# 18. Numbers in depth: widths, literals, bits, and `std:math`

[Chapter 3](03-types-and-variables.md) introduced `int`, `uint`, and the
float types. This chapter covers the rest: integers of an exact width,
other ways to write a number, bit operations, and the `std:math` library.

## Integers of an exact width

Besides `int` and `uint`, there are integer types of exactly 8, 16, 32,
and 64 bits:

| Type | Range |
|---|---|
| `int8` | −128 … 127 |
| `int16` | −32768 … 32767 |
| `int32` | −2147483648 … 2147483647 |
| `int64` | −9223372036854775808 … 9223372036854775807 |
| `uint8` | 0 … 255 |
| `uint16` | 0 … 65535 |
| `uint32` | 0 … 4294967295 |
| `uint64` | 0 … 18446744073709551615 |

```actionscript
package main;

protected int main() {
    uint8 red = 255;
    int64 population = 8_100_000_000;
    uint64 hash = 14695981039346656037;
    trace(red, population, hash);
    return 0;
}
```

Every integer type is a type of its own, so nothing converts between them
on its own, not even to a wider type. Convert explicitly:

```actionscript
package main;

protected int main() {
    int8 small = 100;
    int16 wider = (int16) small;   // fine: an explicit conversion
    int64 total = (int64) wider * 1_000_000;
    trace(wider, total);
    int16 oops = small;            // error: ASX-TYPE-001
    return 0;
}
```

`int32` is not `int` either: `int` is the general integer, whose
intermediate results widen when a calculation needs more room
([chapter 4](04-operators.md#overflow-never-wraps-silently)), and an
`int32` never does.

Arithmetic in every width is checked, as for `int`: a result that doesn't
fit throws a `RangeError` instead of wrapping around.

```actionscript
package main;

protected int main() {
    uint8 level = 250;
    try {
        level += 10;
        trace(level);
    } catch (RangeError e) {
        trace(e.message);   // integer overflow: 250 + 10 = 260 does not fit in 'uint8'
    }
    return 0;
}
```

## Writing numbers

Integers can be written in hex, binary, or octal, and `_` can separate
digits in any number:

| Form | Example | Value |
|---|---|---|
| decimal | `1_000_000` | 1000000 |
| hex | `0xFF`, `0xDEAD_BEEF` | 255, 3735928559 |
| binary | `0b1010_0101` | 165 |
| octal | `0o755` | 493 |
| float | `1_234.5`, `6.022_140_76e23` | |

The prefixes are lower case, and `_` must stand between two digits. A
leading zero, as in `0755`, is an error (`ASX-LEX-005`) so that it can't be
mistaken for octal.

## Bit operations

| Operator | Meaning |
|---|---|
| `a & b` | bits set in both |
| `a \| b` | bits set in either |
| `a ^ b` | bits set in exactly one |
| `~a` | every bit flipped, within the type's width |
| `a << n` | shift left by `n` bits |
| `a >> n` | shift right by `n` bits |

Each also has an assignment form: `&=`, `|=`, `^=`, `<<=`, `>>=`.

```actionscript
package main;

protected int main() {
    uint8 flags = 0b0000_0101;
    flags |= 0b1000_0000;           // set a bit
    flags &= ~((uint8) 0b0100);     // clear a bit
    trace(flags, (flags & 0b1000_0000) != 0);   // 129 true

    uint32 color = 0xFF8800;
    uint32 red = (color >> 16) & 0xFF;
    uint32 green = (color >> 8) & 0xFF;
    trace(red, green, color & 0xFF);            // 255 136 0
    return 0;
}
```

- `&`, `|`, and `^` take two integers of the same type.
- `>>` keeps the sign of a signed number: `-16 >> 2` is `-4`.
- The shift count must be from 0 to the width minus 1, so 0 to 31 for an
  `int`. Anything else is an error (a `RangeError` at run time,
  `ASX-RUN-012`).
- `<<` never loses bits silently. If the result doesn't fit the type, it's
  an overflow error, just like `*`. Mask the value first when you mean to
  drop high bits.
- `&` binds less tightly than `==`, as in C and ActionScript, so
  `x & mask == 0` means `x & (mask == 0)`. That mixes an integer with a
  `bool`, so the compiler stops it. Write `(x & mask) == 0`.

## `std:math`

`use "std:math";` brings in Go's [`math`](https://pkg.go.dev/math) package,
with camelCase names, working on `float64`:

```actionscript
package main;

use "std:math";

protected int main() {
    float64 pi = math.pi;
    float64 angle = pi / 6;
    trace(math.sin(angle), math.cos(angle));    // 0.49999999999999994 0.8660254037844387
    trace(math.sqrt(2.0), math.pow(2.0, 10.0)); // 1.4142135623730951 1024
    trace(math.floor(-2.5), math.round(-2.5), math.roundToEven(2.5));   // -3 -3 2
    trace(math.max(3, 7), math.abs(-12), math.clamp(15, 0, 10));        // 7 12 10
    return 0;
}
```

| Group | Functions |
|---|---|
| Rounding | `abs`, `ceil`, `floor`, `trunc`, `round` (halves away from zero), `roundToEven` |
| Powers and roots | `sqrt`, `cbrt`, `pow`, `hypot`, `exp`, `exp2`, `expm1`, `ldexp` |
| Logarithms | `log`, `log10`, `log2`, `log1p` |
| Trigonometry | `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `atan2` |
| Hyperbolic | `sinh`, `cosh`, `tanh`, `asinh`, `acosh`, `atanh` |
| Comparing and parts | `min`, `max`, `mod`, `remainder`, `dim`, `copysign`, `signbit`, `nextafter` |
| Special values | `isInf(x, sign)`, `inf(sign)`, `nan()` (and the built-ins `isNaN`, `isFinite`) |
| Integers | `abs`, `min`, `max`, and `clamp(x, lo, hi)` on `int` |

**Constants** such as `math.pi`, `math.e`, `math.sqrt2`, `math.maxFloat64`,
and `math.maxInt64` or `math.maxUint8` (one for every integer type) are
untyped, like a literal: they take the type of where they're used, if they
fit. `math.pi` can't be stored in a `float32` without rounding, so that
needs `(float32) math.pi`.

**`min`, `max`, and `abs` work on `int` or `float64`.** The compiler uses
the `int` version when every argument is an `int` or an integer literal,
and the `float64` one otherwise. Mixing an `int` variable with a `float64`
is an error: convert one of them.

**Other number types are converted first:** `math.sqrt((float64) f)` for a
`float` or `float32`, `(float64) n` for an integer.

**The same bits everywhere.** Every `std:math` function returns exactly the
same result in the interpreter and in JavaScript, down to the last bit, on
every computer. JavaScript's own `Math.sin` or `Math.exp` can differ from
one browser to the next in the last digit, so RuneScript doesn't use them:
both targets run the same algorithm, Go's.

Next: [Structs](19-structs.md)
