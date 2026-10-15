# 5. Control flow

Conditions must be `bool`. Write `if (n != 0)`, not `if (n)`. Bodies always
use braces.

## if / else if / else

```actionscript
package main;

string classify(int n) {
    if (n < 0) {
        return "negative";
    } else if (n == 0) {
        return "zero";
    } else {
        return "positive";
    }
}

void main() {
    trace(classify(-5), classify(0), classify(9));   // negative zero positive
}
```

## while

```actionscript
package main;

void main() {
    int n = 3;
    while (n > 0) {
        trace(n);
        n--;
    }
    trace("liftoff");
}
```

## for

`for (init; condition; step) { … }`

```actionscript
package main;

void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

void main() {
    sayHi("Hello world", 3);
}
```

Each of the three parts is optional. With no condition, the loop runs until
`break` or `return`:

```actionscript
package main;

void main() {
    int tries = 0;
    for (;;) {
        tries++;
        if (tries == 3) {
            break;
        }
    }
    trace("stopped after", tries);
}
```

The initializer can declare a variable (scoped to the loop), assign to an
existing one, or call a function. The step can be an assignment, `++`/`--`,
or a call.

## break and continue

`break` leaves the innermost loop. `continue` skips to the next iteration.
In a `for` loop, `continue` still runs the step.

```actionscript
package main;

void main() {
    for (int i = 0; i < 10; i++) {
        if (i % 2 == 0) {
            continue;       // skip even numbers
        }
        if (i > 7) {
            break;          // stop at 9
        }
        trace(i);           // 1 3 5 7
    }
}
```

`break` or `continue` outside a loop is an error (`ASX-FLOW-001`).

## Returning from every path

A function that returns a value must return on every path. The compiler
accepts a body that ends with `return`, or with an `if`/`else` chain whose
every branch returns:

```actionscript
package main;

int sign(int n) {
    if (n < 0) {
        return -1;
    } else if (n == 0) {
        return 0;
    } else {
        return 1;
    }
}

void main() {
    trace(sign(-8), sign(0), sign(8));
}
```

The compiler doesn't treat a loop as guaranteeing a return. Add a final
`return` after the loop:

```actionscript
package main;

int firstSquareAbove(int limit) {
    for (int i = 0; i < 1000; i++) {
        if (i * i > limit) {
            return i;
        }
    }
    return -1;
}

void main() {
    trace(firstSquareAbove(50));   // 8
}
```

Next: [Functions](06-functions.md)
