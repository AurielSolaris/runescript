// C07: floating-point behavior; float32 is emulated with Math.fround.
declare var process: any;

var fround: (x: number) => number = (<any>Math).fround;

function line(parts: any[]): void {
    process.stdout.write(parts.join(" ") + "\n");
}

function signbit(x: number): boolean {
    return x < 0 || (x == 0 && 1 / x < 0);
}

var a = 0.1;
var b = 0.2;
line(["sum", a + b, a + b == 0.3]);
line(["magnitudes", 1e16 + 1.0, 1e300 * 1e10, 5e-324 / 2.0]);

var acc = 0.0;
for (var i = 0; i < 1000; i++) {
    acc += 0.001;
}
line(["accumulate", acc]);

var f = fround(0.1);
var g = fround(f * 3);
line(["float32", f, g, f, fround(16777217), fround(16777217)]);

var zero = 0.0;
var negZero = -zero;
// String(-0) is "0", as RuneScript prints it.
line(["zero", String(negZero), 1.0 / negZero, negZero == zero, signbit(negZero)]);
var nan = zero / zero;
line(["special", 1.0 / zero, -1.0 / zero, nan == nan, nan != nan, nan < 1.0, isNaN(nan)]);
line(["format", 1e21, 1e-7, 123456789012345680000.0, 0.000001, -1.5e-10]);
