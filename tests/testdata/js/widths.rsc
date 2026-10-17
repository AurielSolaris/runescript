// Explicit-width integers (OQ-NUM-07): ranges, checked arithmetic,
// conversions between widths, and 64-bit values.
package tests.js;

public int64 factorial(int64 n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

protected int main() {
    int8 a = 127;
    int16 b = -32768;
    int32 c = 2147483647;
    int64 d = 9223372036854775807;
    uint8 e = 255;
    uint16 f = 65535;
    uint32 g = 4294967295;
    uint64 h = 18446744073709551615;
    trace(a, b, c, d, e, f, g, h);
    trace(-d - 1, h / 3, h % 1000, d / -7, -d % 10);
    trace(factorial(20), (int) (factorial(12) / 1000));
    trace(h > 9223372036854775807, (uint64) d + 1, h - (uint64) d);

    // Conversions between widths are explicit and checked.
    int big = 70000;
    trace((int64) big * 100000, (int8) -128, (uint8) (int16) 200, (float64) h, (float32) d);
    trace((int64) "-9223372036854775808", (uint64) "18446744073709551615", (int16) 3.9);
    trace($"{h} {d} {(string) e}", (char) (uint64) 65, (int64) 'é');

    Array<int64> xs = [1, 2, 3];
    xs[0] += 9223372036854775800;
    trace(xs[0], xs[(uint64) 2]);

    try {
        trace(a + 1);
    } catch (RangeError err) {
        trace(err.message);
    }
    try {
        trace(h + 1);
    } catch (RangeError err) {
        trace(err.message);
    }
    try {
        trace(d * 2);
    } catch (RangeError err) {
        trace(err.message);
    }
    try {
        trace(factorial(21));
    } catch (RangeError err) {
        trace(err.message);
    }
    try {
        trace((uint8) big);
    } catch (RangeError err) {
        trace(err.message);
    }
    try {
        float64 huge = 1e19;
        trace((int64) huge);
    } catch (RangeError err) {
        trace(err.message);
    }
    try {
        trace((uint64) "18446744073709551616");
    } catch (RangeError err) {
        trace(err.message);
    }
    try {
        trace(xs[h]);
    } catch (RangeError err) {
        trace(err.message);
    }
    // Regression: a uint product beyond int64 is reported exactly.
    uint u = 4294967295;
    try {
        trace(u * u);
    } catch (RangeError err) {
        trace(err.message);
    }
    int64 zero = 0;
    try {
        trace(d / zero);
    } catch (RangeError err) {
        trace(err.message);
    }
    return 0;
}
