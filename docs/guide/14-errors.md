# 14. Errors: `throw`, `try`, `catch`, `finally`

Errors work as in ActionScript 3, with RuneScript's type-first syntax:
`catch (Error e)` instead of `catch (e:Error)`.

```actionscript
package main;

public int parsePort(string text) {
    int port = (int) text;   // throws a TypeError if text isn't a number
    if (port < 1 || port > 65535) {
        throw new RangeError("port out of range", 1);
    }
    return port;
}

public void main() {
    try {
        trace(parsePort("8080"));
        trace(parsePort("99999"));
    } catch (RangeError e) {
        trace("bad port:", e.message);
    } catch (Error e) {
        trace("not a port:", e);
    } finally {
        trace("done");
    }
}
```

This prints `8080`, `bad port: port out of range`, and `done`.

## Error values

There are four error types. The last three are kinds of `Error`:

| Type | Used for |
|---|---|
| `Error` | Anything; the base type |
| `RangeError` | A value out of range: overflow, division by zero, a number that doesn't fit |
| `TypeError` | A value of the wrong form, e.g. `(int) "abc"` |
| `ArgumentError` | A bad argument; for your own functions |

`new Error(message, errorID)` creates an error. Both arguments are optional.
Every error has three fields:

| Field | Type | Value |
|---|---|---|
| `e.message` | `string` | The message |
| `e.name` | `string` | The type's name, e.g. `"RangeError"` |
| `e.errorID` | `int` | The ID passed to `new`, or the runtime error's number |

Errors are ordinary values. You can store them, pass them, return them,
and trace them: `trace(e)` prints `RangeError: message`. A `RangeError` can
be stored in an `Error` variable. You can only `throw` an error, not a
string or a number.

## `try`, `catch`, and `finally`

- The catch clauses are tried in order. The first one whose type the error
  has (that type, or a kind of it) runs. Put specific types before `Error`,
  because a clause that can never run is an error (`ASX-TYPE-022`).
- If no clause matches, the error goes on to the enclosing `try` and to the
  calling functions.
- The `finally` block always runs: after the block finishes, after a
  `return`, `break`, or `continue`, and after an error, which then goes on.
- A `finally` block may `return`, `break`, or `continue`, as in
  ActionScript. That replaces however the `try` statement was ending, and
  an error in flight is discarded.
- `catch (e)`, with no type, catches every error, like `catch (Error e)`.
- The caught error's variable is read-only.

A `throw` ends a path like `return` does, so this function needs no final
`return`:

```actionscript
package main;

public int sign(int n) {
    if (n > 0) {
        return 1;
    }
    if (n < 0) {
        return -1;
    }
    throw new ArgumentError("zero has no sign");
}

public void main() {
    trace(sign(-5));
}
```

## Runtime errors can be caught

The runtime errors in [chapter 8](08-diagnostics.md) are thrown as errors
you can catch:

| Runtime error | Thrown as |
|---|---|
| `ASX-RUN-002`: integer overflow | `RangeError` |
| `ASX-RUN-003`: integer division by zero | `RangeError` |
| `ASX-RUN-004`: a standard library function failed | `Error` |
| `ASX-RUN-006`: a conversion failed | `RangeError` or `TypeError` |

```actionscript
package main;

public void main() {
    int big = 2147483647;
    try {
        trace(big + 1);
    } catch (RangeError e) {
        trace(e.errorID, e.message);
        // 2 integer overflow: 2147483647 + 1 = 2147483648 does not fit in 'int'
    }
}
```

The one exception is `ASX-RUN-001`, the call-depth limit (usually runaway
recursion). It can't be caught, and `finally` blocks don't run while the
program stops for it.

## Errors nothing catches

An error that nothing catches stops the program, like any runtime error:

- A runtime error is reported with its own code, message, and location,
  even if it was caught and thrown again.
- An error you created with `new` is reported as `ASX-RUN-005` at the
  `throw`:

```actionscript
package main;

public void main() {
    throw new ArgumentError("missing input");   // runtime error: ASX-RUN-005
}
```

```text
main.rsc:4:5: error ASX-RUN-005: uncaught ArgumentError: missing input
```

Not yet available: stack traces, error methods such as `toString()`, your
and your own error types (they need classes).

Next: [Web pages with `rune:dom`](15-web-pages.md)
