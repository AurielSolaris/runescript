# 9. Coming from ActionScript 3

RuneScript is *inspired by* ActionScript 3, but it is not compatible with
it. AS3 code won't compile unchanged. The main differences:

## Declarations: type first, no `function`, no `var`

ActionScript 3:

```as3
package {
    public class Greeter {
        public function sayHi(message:String, times:int):void {
            for (var i:int = 0; i < times; i++) {
                trace(message);
            }
        }
    }
}
```

RuneScript (v0.1 has no classes yet, so this is a top-level function):

```actionscript
package main;

public void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

void main() {
    sayHi("hi", 2);
}
```

| AS3 | RuneScript |
|---|---|
| `package x { … }` | `package x;` |
| `function f(a:int):int` | `int f(int a)` |
| `var n:int = 0;` | `int n = 0;` |
| `String`, `Boolean` | `string`, `bool` |
| Semicolons optional | Semicolons required |
| `*` / missing types allowed | Every variable, parameter, and function is explicitly typed; there is no `any` |

## Stricter integers

| | AS3 | RuneScript |
|---|---|---|
| `var u:uint = -1;` | Becomes 4294967295 | Compile error `ASX-TYPE-007` |
| `uint` ← `int` variable | Converted silently | Compile error `ASX-TYPE-004` |
| `int + uint` | Converted silently | Compile error `ASX-TYPE-011` |
| `int.MAX_VALUE + 1` | Wraps to a negative number | Never wraps silently (v0.1: runtime error `ASX-RUN-002`) |
| `7 / 2` | `3.5` (a `Number`) | `3` (integer division) |

## Stricter conditions

AS3 lets you write `if (count)` or `if (name)`. RuneScript requires a
`bool`, so write `if (count != 0)`.

## Block scope

In AS3, a `var` declared inside a loop is visible in the whole function.
In RuneScript, a variable lives only in its block, and a new variable can't
reuse the name of one that's still visible.

## Entry point

AS3 programs start from a document class. RuneScript programs start from a
top-level `main()` function, and no class is needed.

## The same as AS3

- `trace(a, b, c)` prints its arguments separated by spaces.
- `public`, `protected`, and `private` are visibility modifiers (their exact
  rules are still being designed).
- `int` and `uint` are the familiar integer names, and `int` is 32-bit by
  default.
- C-style `if`, `while`, `for`, `break`, `continue`, `&&`, `||`, and `!`.

## Not yet available

Classes, interfaces, `static`, `Vector.<T>`/arrays, `Number`/floats, string
concatenation, casts, `null`, `use` (RuneScript's version of `import`),
metadata, E4X XML, and graphics (a native graphics package is planned). See the
[roadmap](../ROADMAP.md).
