# 6. Functions

## Declaring a function

Every function declaration has the form

```text
<access> <return type> <name>(<parameters>) { <body> }
```

where `<access>` is `public`, `private`, or `protected`. It's **required**,
as is the return type (`void` when there is no result). There is no
`function` keyword.

```actionscript
package main;

public int add(int a, int b) {
    return a + b;
}

private void greet(string name) {
    trace("hello", name);
}

public void main() {
    greet("rune");
    trace(add(2, 3));
}
```

| Access | Meaning |
|---|---|
| `public` | Usable from everywhere |
| `private` | Usable only in the same file |
| `protected` | Usable only in the same package |

> These rules are enforced when a program has several files; see
> [Packages and files](12-packages-and-files.md).

`return;` exits a `void` function early.

## Calling

Arguments are checked against the parameter types using the same rules as
assignment. For example, you can't pass an `int` variable to a `uint`
parameter. Top-level functions can be called before their definition:

```actionscript
package main;

public void main() {
    trace(square(12));     // defined below: fine
}

public int square(int n) {
    return n * n;
}
```

A call whose result you don't need can be used as a statement:

```actionscript
package main;

public int logAndCount(string message) {
    trace(message);
    return 1;
}

public void main() {
    logAndCount("result ignored");
}
```

## Parameters are copies, and read-only

Arguments are passed by value. Inside the function, parameters can't be
reassigned (`ASX-NAME-006`). Copy the value into a local variable if you
need to change it:

```actionscript
package main;

public uint gcd(uint a, uint b) {
    uint x = a;
    uint y = b;
    while (y != 0) {
        uint t = x % y;
        x = y;
        y = t;
    }
    return x;
}

public void main() {
    trace(gcd(84, 36));   // 12
}
```

## Recursion

```actionscript
package main;

public int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

public void main() {
    trace(factorial(10));   // 3628800
}
```

Calls can nest up to 10,000 deep. Going deeper stops the program with
`ASX-RUN-001` instead of crashing.

## Prototypes (forward declarations)

A **prototype** declares a function's signature without its body. It ends
in `;`, and parameter names are optional. The definition must come later
in the same scope, with exactly the same types:

```actionscript
package main;

public int twice(int);          // prototype

public void main() {
    trace(twice(21));           // 42
}

public int twice(int x) {       // definition
    return x * 2;
}
```

Top-level functions are visible everywhere anyway, so prototypes are
optional there. Inside a function they're how two nested functions call
each other (see below). Declaring a prototype twice, or after the
definition, is an error (`ASX-NAME-002`). A definition with different types
is `ASX-TYPE-015`. A prototype that is never defined is `ASX-NAME-010`.

## Nested functions

Functions can be declared inside other functions, with an access modifier
like any other function. A nested function is visible from its declaration
to the end of the enclosing block, and inside its own body (so it can be
recursive):

```actionscript
package main;

public void main() {
    private int fib(int n) {
        if (n < 2) {
            return n;
        }
        return fib(n - 1) + fib(n - 2);
    }
    trace(fib(15));   // 610
}
```

To make two nested functions call each other, declare one with a
prototype first:

```actionscript
package main;

public void main() {
    private bool isEven(int n);

    private bool isOdd(int n) {
        if (n == 0) {
            return false;
        }
        return isEven(n - 1);
    }

    private bool isEven(int n) {
        if (n == 0) {
            return true;
        }
        return isOdd(n - 1);
    }

    trace(isEven(10), isOdd(7));   // true true
}
```

## Function boundaries

Every function body, including nested functions and lambdas, is a
**boundary**. Code inside it can use:

- its own parameters and local variables;
- functions: top-level functions, and nested functions visible where it is
  declared (including itself).

It **cannot** use the local variables or parameters of the function it's
nested in. RuneScript functions don't capture variables (`ASX-NAME-009`):

```actionscript
package main;

public void main() {
    int count = 0;
    private void bump() {
        count++;      // error: ASX-NAME-009
    }
}
```

Pass the value in as a parameter, and return the result:

```actionscript
package main;

public void main() {
    private int bump(int count) {
        return count + 1;
    }
    int count = 0;
    count = bump(count);
    trace(count);   // 1
}
```

Because of the boundary, a nested function can reuse a name from the
enclosing function. Inside the nested function, that name refers only to
the nested function's own variable. `break` and `continue` can't cross a
boundary either.

## Function types

A function type is written C-style: the return type, followed by the
parameter types in parentheses.

| Type | Meaning |
|---|---|
| `int(int, int)` | Takes two `int`s, returns an `int` |
| `void(string)` | Takes a `string`, returns nothing |
| `bool()` | Takes nothing, returns a `bool` |
| `int(int)(string)` | Takes a `string`, returns an `int(int)` function |

Function types can be used for variables, parameters, and return types.
Two function types are the same when their parameter and return types
are the same (`int` and `sint` count as the same type).

## Functions as values (callbacks)

Naming a function without calling it gives a function value. You can store
it, pass it, return it, and call it:

```actionscript
package main;

public int apply(int(int, int) op, int a, int b) {
    return op(a, b);
}

public int multiply(int a, int b) {
    return a * b;
}

public void main() {
    int(int, int) f = multiply;
    trace(f(6, 7));                   // 42
    trace(apply(multiply, 3, 4));     // 12
}
```

Function values can't be traced or compared with `==`. The built-in
`trace` isn't a value; wrap it in a lambda instead.

## Lambdas

A **lambda** is an anonymous function: typed parameters in parentheses,
then `=>`, then either an expression or a block.

```actionscript
package main;

public void repeat(int times, void(int) body) {
    for (int i = 0; i < times; i++) {
        body(i);
    }
}

public void main() {
    int(int, int) add = (int a, int b) => a + b;
    trace(add(2, 3));                                 // 5

    repeat(2, (int i) => trace("tick", i));           // tick 0, tick 1

    int(int) abs = (int x) => {
        if (x < 0) {
            return -x;
        }
        return x;
    };
    trace(abs(-4));                                   // 4
}
```

- **Return type.** When a lambda is assigned, passed, or returned where a
  function type is expected, it takes that type's return type. Otherwise
  the return type comes from the body: the expression's type, or the type
  of the first `return`.
- Lambdas follow the same boundary rule as nested functions: they can't
  use the enclosing function's variables.
- Lambdas don't take an access modifier.

## Returning functions

```actionscript
package main;

public int(int) pick(string op) {
    private int double(int n) {
        return n * 2;
    }
    if (op == "double") {
        return double;
    }
    return (int n) => n * n;
}

public void main() {
    trace(pick("double")(5), pick("square")(5));   // 10 25
}
```

`pick("double")(5)` calls the function returned by `pick`. The function
value is evaluated first, then the arguments.

## Not yet available

Overloading (two functions with the same name), default arguments,
variadic parameters, generic functions, and closures that capture
variables.

Next: [Output with `trace`](07-trace.md)
