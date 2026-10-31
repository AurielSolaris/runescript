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

RuneScript, as a top-level function (no class is needed; a class version
is in [chapter 24](24-classes.md)):

```actionscript
package main;

public void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

protected int main() {
    sayHi("hi", 2);
    return 0;
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
| `Array`, `Vector.<String>` | `Array<string>`: every element has the declared type |
| `"n=" + 5` | `"n=" + (string) 5` or `$"n={5}"`: `+` never converts numbers |
| `null`, `*` | Neither exists: no value is ever missing or untyped ([chapter 17](17-no-null-any-generics.md)) |

## Stricter integers

| Code | AS3 | RuneScript |
|---|---|---|
| `var u:uint = -1;` | Becomes 4294967295 | Compile error `ASX-TYPE-007` |
| `uint` ← `int` variable | Converted silently | Compile error `ASX-TYPE-004` |
| `int + uint` | Converted silently | Compile error `ASX-TYPE-011` |
| `int.MAX_VALUE + 1` | Wraps to a negative number | Never wraps silently (v0.1: runtime error `ASX-RUN-002`) |
| `7 / 2` | `3.5` (a `Number`) | `3` (integer division) |

## Stricter conditions

AS3 lets you write `if (count)` or `if (name)`. RuneScript requires a
`bool`, so write `if (count != 0)`.

## Closures

AS3 closures capture the enclosing function's variables. RuneScript
functions don't: nested functions and lambdas can only use their own
parameters and locals, plus other functions. Pass values in as
parameters instead.

| AS3 | RuneScript |
|---|---|
| `function(a:int):int { return a * 2; }` | `(int a) => a * 2` |
| `var f:Function = add;` | `int(int, int) f = add;` (typed) |
| `public function f():void` | `public void f()` (the access modifier is required) |

## Block scope

In AS3, a `var` declared inside a loop is visible in the whole function.
In RuneScript, a variable lives only in its block. A variable in a nested
block may hide one of an enclosing block, with a warning, but not one of
the same block. As in AS3, `int n;` starts at `0`; unlike AS3, `string s;`
starts at `""`, since there is no `null`.

## Entry point

AS3 programs start from a document class. RuneScript programs start from a
top-level `protected int main()` function, and no class is needed.

## The same as AS3

- `trace(a, b, c)` prints its arguments separated by spaces.
- `public`, `protected`, and `private` are visibility modifiers (their exact
  rules are still being designed).
- `int` and `uint` are the familiar integer names, and `int` is 32-bit by
  default.
- C-style `if`, `while`, `for`, `break`, `continue`, `&&`, `||`, and `!`.

## Not yet available

Inheritance, interfaces, `static`, metadata, E4X XML, and graphics (a
native graphics package is planned). Classes exist, without inheritance
([chapter 24](24-classes.md)); unlike AS3, `this` is required to reach a
field or method, and the constructor is written without `function`:
`public Point(int x, int y) { ... }`. See the [roadmap](../ROADMAP.md).

Next: [Standard library: `std:string`](10-std-string.md)
