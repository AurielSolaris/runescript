// Errors, throw, and try/catch/finally, as in ActionScript 3.
package tests.js;

use "std:string";

public int checkedDivide(int a, int b) {
    if (b == 0) {
        throw new ArgumentError("b must not be zero", 1001);
    }
    return a / b;
}

// finally runs after a return in try; the returned value is kept.
public int withFinally(int n) {
    try {
        return n * 2;
    } finally {
        trace("finally after return", n);
    }
}

// A return in finally replaces the try block's ending, error included.
public int finallyReturns() {
    try {
        throw new Error("discarded");
    } finally {
        return 7;
    }
}

public void rethrow() {
    try {
        int big = 2147483647;
        trace(big + 1);
    } catch (RangeError e) {
        trace("caught once:", e.name, e.errorID);
        throw e;
    }
}

protected int main() {
    // A thrown error is caught by the first matching clause.
    try {
        trace(checkedDivide(10, 2));
        trace(checkedDivide(1, 0));
        trace("not reached");
    } catch (RangeError e) {
        trace("wrong clause");
    } catch (Error e) {
        trace("caught", e.name, e.message, e.errorID);
        trace(e);
    }

    // Runtime errors are catchable: overflow, division by zero,
    // conversions, and failed standard library calls.
    try {
        int z = 0;
        trace(1 / z);
    } catch (RangeError e) {
        trace(e);
    }
    try {
        trace((int) "12x");
    } catch (TypeError e) {
        trace(e.message, e.errorID);
    }
    try {
        float negative = -1.5;
        trace((uint) negative);
    } catch (RangeError e) {
        trace(e.message);
    }
    try {
        trace(string.repeat("ab", -1));
    } catch (Error e) {
        trace(e.name, e.errorID);
    }
    try {
        float huge = 1e308;
        trace(huge * 10); // Infinity: floats follow IEEE 754
        trace((int) (huge * 10));
    } catch (RangeError e) {
        trace(e.message);
    }
    try {
        float zero = 0;
        trace((uint) (zero / zero));
    } catch (e) {
        trace("catch (e) catches", e.name);
    }

    // finally runs on every path: normal, return, break, continue, error.
    trace(withFinally(21));
    trace(finallyReturns());
    for (int i = 0; i < 3; i++) {
        try {
            if (i == 0) {
                continue;
            }
            if (i == 2) {
                break;
            }
            trace("body", i);
        } finally {
            trace("finally", i);
        }
    }
    try {
        try {
            throw new Error();
        } finally {
            trace("inner finally");
        }
    } catch (Error e) {
        trace("outer caught", e, e.message == "", e.errorID);
    }

    // A rethrown runtime error keeps its type, message, and site.
    try {
        rethrow();
    } catch (Error e) {
        trace("caught again:", e.message);
    }

    // break and continue in finally also discard the error in flight.
    for (int k = 0; k < 5; k++) {
        try {
            if (k < 2) {
                throw new Error("dropped");
            }
        } finally {
            trace("finally", k);
            if (k < 1) {
                continue;
            }
            break;
        }
    }

    // Errors are values.
    Error saved = new TypeError("kept");
    RangeError r = new RangeError();
    Error(string) make = (string m) => new Error(m, 7);
    trace(saved, r, make("made"), make("x").errorID);
    return 0;
}
