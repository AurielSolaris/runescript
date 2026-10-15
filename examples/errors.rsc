// Errors: throw, try/catch/finally, and catching runtime errors.
package examples;

// parseAge turns text into an age, or throws.
public int parseAge(string text) {
    int age = (int) text; // a TypeError if text isn't a whole number
    if (age < 0 || age > 150) {
        throw new RangeError("not a plausible age", 100);
    }
    return age;
}

public void report(string text) {
    try {
        trace(text, "->", parseAge(text));
    } catch (RangeError e) {
        trace(text, "-> out of range:", e.message);
    } catch (TypeError e) {
        trace(text, "-> not a number");
    }
}

public int main() {
    report("42");
    report("-7");
    report("forty");

    // Runtime errors are errors too, and finally always runs.
    int big = 2147483647;
    for (int step = 1; step <= 2; step++) {
        try {
            trace("adding", step);
            big += 2147483647 * (step - 1) + 1;
            trace("never printed");
        } catch (RangeError e) {
            trace("caught", e.name, e.errorID);
        } finally {
            trace("finally", step);
        }
    }
    return 0;
}
