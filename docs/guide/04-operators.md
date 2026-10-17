# 4. Operators

## Arithmetic

| Operator | Meaning | Example | Result |
|---|---|---|---|
| `+` | add | `7 + 2` | `9` |
| `-` | subtract | `7 - 9` | `-2` |
| `*` | multiply | `7 * 3` | `21` |
| `/` | divide (whole numbers, rounds toward zero) | `-7 / 2` | `-3` |
| `%` | remainder (same sign as the left side) | `-7 % 2` | `-1` |
| unary `-` | negate (signed integers only) | `-x` | |

```actionscript
package main;

protected int main() {
    int x = 7;
    trace(x + 2, x - 9, x * 3, x / 2, x % 4, -x);   // 9 -2 21 3 3 -7
    return 0;
}
```

Both operands must have the same signedness: `int`/`sint` with
`int`/`sint`, or `uint` with `uint`. A number literal adapts to the other
operand, so `u + 1` works when `u` is a `uint`. The same operators (except
`%`) work on floats; see [Floats and conversions](13-floats-and-conversions.md).

### Overflow never wraps silently

If an integer value doesn't fit its type, the program stops with a runtime
error that points at the operation. It never wraps around:

```actionscript
package main;

protected int main() {
    int x = 2147483647;
    x = x + 1;          // runtime error: ASX-RUN-002 integer overflow
    return 0;
}
```

Inside a longer expression, `int` and `uint` **widen**: every result that
feeds another operation, a comparison, or a conversion such as `(int64)` is
computed exactly, up to 64 bits. Only the final value has to fit its type.
So the middle of a calculation can go past the limits, as long as the
answer comes back:

```actionscript
package main;

protected int main() {
    int big = 100000;
    trace(big * big / big);         // 100000: the product only exists in between
    trace(big * big > 2147483647);  // true: comparisons see the exact value
    int64 exact = (int64) (big * big);
    trace(exact);                   // 10000000000
    return 0;
}
```

Variables, parameters, array elements, and struct fields keep their
declared type: `int x = big * big;` still stops with `ASX-RUN-002`, because
`x` is an `int`. A value never shrinks to a narrower type behind your back.
The explicit widths (`int8` … `uint64`, [chapter 18](18-numbers.md)) never
widen. Dividing an integer by zero is also a runtime error (`ASX-RUN-003`).
Both are `RangeError`s, which a `try` statement can catch (see
[Errors](14-errors.md)).

## Comparison

`==  !=  <  <=  >  >=` compare two integers of the same signedness, or two
floats of the same type, and produce a `bool`. `==` and `!=` also work on two `bool`s or two `string`s.

```actionscript
package main;

protected int main() {
    int a = 1;
    int b = 2;
    string s = "rune";
    trace(a < b, a == b, s == "rune", true != false);   // true false true true
    return 0;
}
```

## Logic

| Operator | Meaning |
|---|---|
| `a && b` | and: `b` is evaluated only if `a` is `true` |
| `a \|\| b` | or: `b` is evaluated only if `a` is `false` |
| `!a` | not |

Both operands must be `bool`. Numbers are never treated as true or false.

```actionscript
package main;

protected int main() {
    int age = 30;
    bool adult = age >= 18 && !(age > 120);
    trace(adult);   // true
    return 0;
}
```

## Precedence

From tightest to loosest binding:

1. `( … )`, calls
2. unary `-`, `!`, conversions `(T) x`
3. `*`, `/`, `%`
4. `+`, `-`
5. `<<`, `>>`
6. `<`, `<=`, `>`, `>=`
7. `==`, `!=`
8. `&`
9. `^`
10. `|`
11. `&&`
12. `||`

Unary `~` sits with `-` and `!`. The bit operators are covered in
[chapter 18](18-numbers.md).

All binary operators are left-associative: `10 - 4 - 3` is `(10 - 4) - 3`.

`+` also joins two strings, and `==`, `!=`, `<`, `<=`, `>`, `>=` compare
two `char`s. Neither converts anything: `"n=" + 5` is an error
([chapter 16](16-text-and-arrays.md)).

## Not available yet

`?:`, `?.`, and `??`.

Next: [Control flow](05-control-flow.md)
