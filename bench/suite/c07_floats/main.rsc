// C07: floating-point behavior.
package bench;

use "std:math";

protected int main() {
    float64 a = 0.1;
    float64 b = 0.2;
    trace("sum", a + b, a + b == 0.3);
    trace("magnitudes", 1e16 + 1.0, 1e300 * 1e10, 5e-324 / 2.0);

    float64 acc = 0.0;
    for (int i = 0; i < 1000; i++) {
        acc += 0.001;
    }
    trace("accumulate", acc);

    float32 f = (float32) 0.1;
    float32 g = f * (float32) 3;
    trace("float32", f, g, (float64) f, (float32) 16777217, (float64) ((float32) 16777217));

    float64 zero = 0.0;
    float64 negZero = -zero;
    trace("zero", negZero, 1.0 / negZero, negZero == zero, math.signbit(negZero));
    float64 nan = zero / zero;
    trace("special", 1.0 / zero, -1.0 / zero, nan == nan, nan != nan, nan < 1.0, isNaN(nan));
    trace("format", 1e21, 1e-7, 123456789012345680000.0, 0.000001, -1.5e-10);
    return 0;
}
