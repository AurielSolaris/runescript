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

public void main() {
    int x = 7;
    trace(x + 2, x - 9, x * 3, x / 2, x % 4, -x);   // 9 -2 21 3 3 -7
}
```

Both operands must have the same signedness: `int`/`sint` with
`int`/`sint`, or `uint` with `uint`. A number literal adapts to the other
operand, so `u + 1` works when `u` is a `uint`.

### Overflow never wraps silently

In v0.1, if a result doesn't fit its type, the program stops with a runtime
error that points at the operation. It never wraps around:

```actionscript
package main;

public void main() {
    int x = 2147483647;
    x = x + 1;          // runtime error: ASX-RUN-002 integer overflow
}
```

The language design calls for numbers to *widen* automatically. Widening
isn't designed yet, so v0.1 stops instead of guessing. Dividing by zero is
also a runtime error (`ASX-RUN-003`).

## Comparison

`==  !=  <  <=  >  >=` compare two integers of the same signedness and
produce a `bool`. `==` and `!=` also work on two `bool`s or two `string`s.

```actionscript
package main;

public void main() {
    int a = 1;
    int b = 2;
    string s = "rune";
    trace(a < b, a == b, s == "rune", true != false);   // true false true true
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

public void main() {
    int age = 30;
    bool adult = age >= 18 && !(age > 120);
    trace(adult);   // true
}
```

## Precedence

From tightest to loosest binding:

1. `( … )`, calls
2. unary `-`, `!`
3. `*`, `/`, `%`
4. `+`, `-`
5. `<`, `<=`, `>`, `>=`
6. `==`, `!=`
7. `&&`
8. `||`

All binary operators are left-associative: `10 - 4 - 3` is `(10 - 4) - 3`.

## Not available in v0.1

String concatenation with `+`, bitwise operators, shifts, `?:`, `?.`, and
`??`. To print several values together, pass them all to `trace` (see
[Output with `trace`](07-trace.md)).

Next: [Control flow](05-control-flow.md)
