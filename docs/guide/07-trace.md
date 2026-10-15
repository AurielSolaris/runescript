# 7. Output with `trace`

`trace` is the built-in way to print, a name kept from ActionScript. It
takes any number of arguments of any type except `void`. It prints them
separated by single spaces, followed by a newline, to standard output.

```actionscript
package main;

public void main() {
    trace("Hello");                 // Hello
    trace("x =", 42);               // x = 42
    trace(1, -2, true, "three");    // 1 -2 true three
    trace();                        // (empty line)
}
```

| Value | Printed as |
|---|---|
| `int`, `sint`, `uint` | decimal, e.g. `-42` |
| `float`, `float32`, `float64` | as ActionScript prints numbers: `64`, `0.5`, `1e+21`, `NaN` ([chapter 13](13-floats-and-conversions.md)) |
| `bool` | `true` / `false` |
| `string` | the text itself, with no quotes |
| an error | its name and message, e.g. `RangeError: too big` ([chapter 14](14-errors.md)) |

There's no string concatenation yet, so pass several arguments instead of
building one string.

`trace` is a reserved built-in. You can't declare a function or variable
named `trace`.

Next: [Diagnostics reference](08-diagnostics.md)
