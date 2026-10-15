// Floats, conversions, and errors.
package tests.errors;

use "std:string";

public int main() {
    int i = 64.3;                      // expect: ASX-TYPE-001 7:13
    float32 a = 0.1;                   // expect: ASX-TYPE-018 8:17
    float32 b = 1e39;                  // expect: ASX-TYPE-006 9:17
    float f = 1.5;
    trace(f + i);                      // expect: ASX-TYPE-012 11:13
    trace(f % 2.0);                    // expect: ASX-TYPE-012 12:13
    float64 d = f;                     // expect: ASX-TYPE-001 13:17
    trace((uint) -2.5);                // expect: ASX-TYPE-007 14:18
    trace((int) 3e10);                 // expect: ASX-TYPE-006 15:17
    trace((int) true);                 // expect: ASX-TYPE-017 16:11
    trace((bool) i);                   // expect: ASX-TYPE-017 17:11
    Error e = new Error("x");
    trace((string) e);                 // expect: ASX-TYPE-017 19:11
    trace(new string("x"));            // expect: ASX-TYPE-019 20:15
    trace(new Error("x", 1, 2));       // expect: ASX-TYPE-003 21:11
    trace(e.what);                     // expect: ASX-TYPE-020 22:13
    trace(i.message);                  // expect: ASX-TYPE-020 23:13
    trace(e.toString());               // expect: ASX-TYPE-020 24:13
    trace(string.toUpper);             // expect: ASX-NAME-004 25:11
    throw "oops";                      // expect: ASX-TYPE-021 26:11
    try {
        trace(1);
    } catch (Error e2) {
        e2 = new Error();              // expect: ASX-NAME-006 30:9
    } catch (RangeError e3) {          // expect: ASX-TYPE-022 31:14
    } catch (int n) {                  // expect: ASX-TYPE-021 32:14
    } finally {
        trace("finally may return");
    }
    while (true) {
        try {
            trace(2);
        } finally {
            break;
        }
    }
    return 0;
}
