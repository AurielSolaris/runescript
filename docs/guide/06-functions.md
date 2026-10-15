# 6. Functions

## Declaring a function

The return type comes first, then the name, then typed parameters. There is
no `function` keyword.

```actionscript
package main;

int add(int a, int b) {
    return a + b;
}

void greet(string name) {
    trace("hello", name);
}

void main() {
    greet("rune");
    trace(add(2, 3));
}
```

- Use `void` for functions that don't return a value. `return;` exits one
  early.
- A visibility modifier (`public`, `protected`, `private`) can come before
  the return type. In v0.1 it is recorded but has no effect.

## Calling

Arguments are checked against the parameter types using the same rules as
assignment. For example, you can't pass an `int` variable to a `uint`
parameter. Functions can be called before their declaration:

```actionscript
package main;

void main() {
    trace(square(12));     // defined below: fine
}

int square(int n) {
    return n * n;
}
```

A call whose result you don't need can be used as a statement:

```actionscript
package main;

int logAndCount(string message) {
    trace(message);
    return 1;
}

void main() {
    logAndCount("result ignored");
}
```

## Parameters are copies, and read-only

Arguments are passed by value. Inside the function, parameters can't be
reassigned (`ASX-NAME-006`). Copy the value into a local variable if you
need to change it:

```actionscript
package main;

uint gcd(uint a, uint b) {
    uint x = a;
    uint y = b;
    while (y != 0) {
        uint t = x % y;
        x = y;
        y = t;
    }
    return x;
}

void main() {
    trace(gcd(84, 36));   // 12
}
```

## Recursion

```actionscript
package main;

int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

void main() {
    trace(factorial(10));   // 3628800
}
```

Calls can nest up to 10,000 deep. Going deeper stops the program with
`ASX-RUN-001` instead of crashing.

## Not available in v0.1

Overloading (two functions with the same name), default arguments,
functions as values, lambdas, and nested functions.

Next: [Output with `trace`](07-trace.md)
