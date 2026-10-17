// Floats and checked conversions.
package tests.js;

use "std:string";

public float average(float a, float b) {
    return (a + b) / 2;
}

protected int main() {
    // float: float32 by default, widening exactly; values are float64's.
    float a = 0.1;
    float b = 0.2;
    trace(a + b, a * 3, average(1, 2), -a, 1.5e300 * 1e8 > 1e300);
    trace(64.0, 1e21, 1e-7, 0.000001, 123.456, -0.0);

    // float32 rounds every result to float32; float64 is IEEE double.
    float32 f = (float32) 0.1;
    float32 g = f * 3;
    float64 d = 0.1;
    trace(f, g, (float64) f, d * 3, (float32) 16777217);

    // C-style casts: truncation toward zero, exact range checks.
    trace((int) 64.3, (int) -64.7, (uint) 3.99, (int) 2147483647.9);
    trace((float) 7 / 2, 7 / 2, (int) ((float) 7 / 2));

    // Strings and bools.
    trace((float) "64.7", (int) "-42", (uint) "+7", (float32) "0.1", (bool) "true");
    trace((string) 1.5, (string) 42, (string) true, string.count((string) 1e21, "0"));

    // Compound assignment and comparison.
    float x = 1;
    x *= 2.5;
    x -= 0.5;
    trace(x, x == 2, x != 2.5, x <= 2);

    // IEEE 754, as in ActionScript 3: Infinity and NaN.
    float zero = 0;
    float nan = zero / zero;
    trace(1 / zero, -1 / zero, nan, nan == nan, nan != nan, nan < 1, isNaN(nan));
    trace(1e308 * 10, isFinite(1e308 * 10), isFinite(1.5), Infinity > 1e308, -Infinity);
    float32 small = (float32) 3e38;
    trace(small * 2, (float32) 1e39, isNaN((float32) nan), (float64) Infinity);
    trace((string) nan, (float) "-Infinity", (float) "NaN", (float32) "1e400", (float) "1e400");
    int NaN = 3; // a variable may hide the constant
    trace(NaN);
    return 0;
}
