# 7. Output with `trace`

`trace` is the built-in way to print, a name kept from ActionScript. It
takes any number of arguments of any type except `void`. It prints them
separated by single spaces, followed by a newline, to standard output.

```actionscript
package main;

protected int main() {
    trace("Hello");                 // Hello
    trace("x =", 42);               // x = 42
    trace(1, -2, true, "three");    // 1 -2 true three
    trace();                        // (empty line)
    return 0;
}
```

| Value | Printed as |
|---|---|
| `int`, `sint`, `uint` | decimal, e.g. `-42` |
| `float`, `float32`, `float64` | as ActionScript prints numbers: `64`, `0.5`, `1e+21`, `NaN` ([chapter 13](13-floats-and-conversions.md)) |
| `bool` | `true` / `false` |
| `string` | the text itself, with no quotes |
| `char` | the character itself, e.g. `é` |
| an error | its name and message, e.g. `RangeError: too big` ([chapter 14](14-errors.md)) |

To print one string built from several values, join them with `+` or use
an interpolated string such as `$"{name} is {age}"`
([chapter 16](16-text-and-arrays.md)). Arrays can't be traced; join them
with `string.join` first.

`trace` is a reserved built-in. You can't declare a function or variable
named `trace`.

Next: [Diagnostics reference](08-diagnostics.md)
