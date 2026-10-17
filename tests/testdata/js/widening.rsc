// Widening (OQ-NUM-01, option A): intermediate results of the generic int
// and uint are exact within int64; only the outermost operation's result,
// the value that is stored, passed, or traced, must fit the type.
package tests.js;

public int same(int x) {
    return x;
}

protected int main() {
    int max = 2147483647;
    int min = -2147483648;
    int big = 100000;

    // The products overflow int, but the quotients fit.
    trace(big * big / big);                 // 100000
    trace(max * 2 / 2, min * 3 / 3);        // 2147483647 -2147483648
    trace((max + 1) - 1, (min - 1) + 1);    // through 2^31 and back
    trace(-(min + 1) - 1 + 1);              // 2147483647
    trace(max * max / max / max);           // 1
    trace(same(big * big / 1000000));       // 10000

    // Comparisons see exact intermediates: the usual overflow check works.
    trace(big * big > max, max + 1 > max, min - 1 < min);   // true true true
    trace(max * 2 - max == max);            // true

    // Explicit conversions see them too, so nothing is lost on the way to
    // a wider type, and nothing contracts on the way back.
    int64 wide = (int64) (max * max);
    trace(wide);                            // 4611686014132420609
    trace((int64) (max + max + max) - 3);   // 6442450938
    trace((float64) (big * big));           // 10000000000
    trace((string) (big * big));            // 10000000000

    // uint intermediates may go below zero.
    uint a = 3;
    uint b = 5;
    trace(a - b + 4, a * 2 - b - 1);        // 2 0
    trace(a - b < 0, (int) (a - b + b));    // true 3

    // A wide BigInt-sized intermediate on the JavaScript target.
    int64 huge = (int64) (max * max * 2);
    trace(huge, (int64) (max * max * 2 / max / 2));     // 9223372028264841218 2147483647

    // Stored results must still fit.
    try {
        int x = big * big;
        trace("not reached", x);
    } catch (RangeError e) {
        trace(e.message);
    }
    try {
        uint u = a - b;
        trace("not reached", u);
    } catch (RangeError e) {
        trace(e.message);
    }
    try {
        trace(max * max * max / max / max);
    } catch (RangeError e) {
        trace(e.message);
    }
    try {
        int z = 0;
        trace(big * big / z);
    } catch (RangeError e) {
        trace(e.message);
    }
    return 0;
}
