// std:math, phase A: int helpers, exact float functions, constants.
package tests.js;
use "std:math";
public void main() {
    int m = math.max(3, 7);
    float64 r = math.sqrt(2.0);
    trace(m, math.min(-4, 9), math.abs(-12), math.clamp(15, 0, 10), r);
    trace(math.max(1.5, 2), math.round(-2.5), math.roundToEven(2.5), math.mod(7.5, 2.0), math.remainder(7.5, 2.0));
    trace(math.pi, math.e, (float32) math.pi, (int64) math.maxInt64, math.maxUint8, math.signbit(-0.0), math.copysign(3.0, -1.0));
    trace(math.nextafter(1.0, 2.0), math.isInf(math.inf(1), 0), math.floor(-1.5), math.ceil(-1.5), math.trunc(-1.5));
    try { trace(math.abs(math.minInt)); } catch (RangeError e) { trace(e.message); }
    try { trace(math.clamp(1, 5, 2)); } catch (ArgumentError e) { trace(e.name, e.message); }
}
