// std:math: int helpers, exact float functions, constants, and the
// transcendental functions (one shared implementation, internal/fmath).
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
    trace(math.sin(1.0), math.cos(1.0), math.tan(1.0), math.sin(1e22), math.cos(-3e9));
    trace(math.asin(0.5), math.acos(0.5), math.atan(1.0), math.atan2(-1.0, -1.0));
    trace(math.sinh(1.0), math.cosh(1.0), math.tanh(0.5), math.asinh(2.0), math.acosh(2.0), math.atanh(0.5));
    trace(math.exp(1.0), math.exp2(10.0), math.expm1(1e-10), math.log(10.0), math.log10(1000.0), math.log2(8.0), math.log1p(1e-15));
    trace(math.cbrt(27.0), math.pow(2.0, 0.5), math.pow(-8.0, 3.0), math.pow(10.0, -2.0), math.hypot(3.0, 4.0), math.ldexp(0.75, 4));
    float32 angle = 0.5;
    trace((float32) math.sin((float64) angle));
}
