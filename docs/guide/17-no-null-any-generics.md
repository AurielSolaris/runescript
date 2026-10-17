# 17. No `null`, no `any`, no generics

RuneScript leaves out three things most languages have: a `null` value, a
catch-all `any` type, and generics you write yourself. This chapter
explains why that makes the type system stronger, and shows how to say
"there's no value here" without `null`: with `false`.

## What's missing, exactly

| Feature | In RuneScript | What you get instead |
|---|---|---|
| `null`, `undefined`, `nil` | Doesn't exist (`ASX-NAME-001`) | Every value is a real value of its type. "No value" is a `bool` that is `false` |
| `any`, `Object`, `*` | Doesn't exist (`ASX-TYPE-010`) | Every variable, parameter, and result has one exact type |
| Your own generic functions and types | Can't be declared (`ASX-PARSE-001`) | Write the function for each type you need. The built-in `Array<T>` and `Signal<T>` are the only types that take a type argument, and the language defines them |

```actionscript
package main;

protected int main() {
    string name = null;   // error: ASX-NAME-001
    return 0;
}
```

```actionscript
package main;

protected int main() {
    any value = 5;   // error: ASX-TYPE-010
    return 0;
}
```

## Why leaving them out makes the types stronger

**Without `null`, a type means what it says.** In a language with `null`,
a `string` variable really means "a string, or nothing", and every use is a
place the program can crash. Its inventor, Tony Hoare, called it his
"billion-dollar mistake". In RuneScript a `string` is always a string: every
variable must be initialised when it's declared, there are no default
values, and no operation produces an empty reference. When a value really
can be missing, the program says so with a separate `bool` you can see and
check.

**Without `any`, the compiler checks everything.** An `any` is a hole in
the type system: whatever goes in comes out unchecked, and an error only
shows up when the program runs, often far from its cause. TypeScript's
`any` and unchecked casts are how wrong values get into "typed" programs.
RuneScript has no such hole, which is also why it can check every value
that comes back from JavaScript ([chapter 11](11-javascript.md)): there is
no type that would let an unchecked value through.

**Without user generics, every function has one concrete signature.** A
generic function's meaning depends on the types it's used with, so its
errors appear at the call, in terms of type variables. RuneScript's
functions say exactly what they take and return, its error messages name
real types, and the compiled JavaScript needs no type information at run
time. The built-in `Array<T>` covers the most common need for generics, a
list of anything, without opening the type system up.

The cost is some repetition: a `maxInt` and a `maxFloat64` instead of one
generic `max`. RuneScript accepts that cost in exchange for code whose
types can always be trusted.

## Using `false` as `null`

`false` doesn't replace a missing value by itself, since a `string` can't
hold a `bool` (`ASX-TYPE-001`). Instead, a value that might be missing
comes with a `bool` that is `false` while there's no value, and `true` once
there is one. The compiler then makes the absence visible: you can't read
the value without the flag sitting right next to it.

### A flag next to the value

Declare the flag and a placeholder value together, and set both at once:

```actionscript
package main;

protected int main() {
    bool hasWinner = false;   // false plays the role of null
    string winner = "";       // a placeholder; never read while hasWinner is false

    int score = 42;
    if (score > 40) {
        hasWinner = true;
        winner = "Ada";
    }

    if (hasWinner) {
        trace("winner:", winner);   // winner: Ada
    } else {
        trace("no winner yet");
    }
    return 0;
}
```

### A function that may not find anything

A function can't return "a string or `false`", because its result has one
type. Return an array with zero or one element: empty means "not found".
Its `length > 0` is the `true`/`false` you check:

```actionscript
package main;

use "std:string";

public Array<string> findUser(Array<string> users, string prefix) {
    for (int i = 0; i < users.length; i++) {
        if (string.hasPrefix(users[i], prefix)) {
            return [users[i]];
        }
    }
    return [];   // nothing found: the empty array is the "null"
}

protected int main() {
    Array<string> users = ["ada", "grace", "linus"];
    Array<string> found = findUser(users, "gr");
    bool ok = found.length > 0;
    if (ok) {
        trace("found", found[0]);   // found grace
    }
    trace(findUser(users, "zz").length > 0);   // false
    return 0;
}
```

### A question and an answer

When looking a value up is cheap, ask first with a function that returns
`bool`, then get the value. This is how `string.contains` and
`string.index` work together:

```actionscript
package main;

use "std:string";

protected int main() {
    string line = "key=value";
    if (string.contains(line, "=")) {
        int at = string.index(line, "=");
        trace(string.slice(line, 0, at));   // key
    }
    return 0;
}
```

Functions that return a position use `-1` for "not found", as in Go. That
works because a real position is never negative, so `-1` can't be mistaken
for a result.

### An error when absence is a mistake

When a missing value means something went wrong, don't return a flag at
all. Throw, and let the caller decide ([chapter 14](14-errors.md)):

```actionscript
package main;

public int parsePort(string text) {
    int port = (int) text;   // throws TypeError if text isn't a number
    if (port < 1 || port > 65535) {
        throw new RangeError("port out of range: " + text);
    }
    return port;
}

protected int main() {
    try {
        trace(parsePort("8080"));   // 8080
        trace(parsePort("99999"));
    } catch (RangeError e) {
        trace(e.message);   // port out of range: 99999
    }
    return 0;
}
```

## Choosing a pattern

| The value might be missing because… | Use |
|---|---|
| it isn't known yet | a `bool` flag that is `false` until it's set |
| a search found nothing | an array with zero or one element, or `-1` for a position |
| the caller can ask first | a `bool` question, then the value |
| something went wrong | `throw` an error |

A `struct` keeps the flag and the value together in one value, such as
`{ bool present; int value; }`, so they can't drift apart: see
[Using a struct for "no value"](19-structs.md) in chapter 19.

Next: [Numbers in depth](18-numbers.md)
