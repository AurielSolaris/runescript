// B04: direct calls, nested calls, and calls through function values.
declare var process: any;

function inc(x: number): number {
    return x + 1;
}

function mix(a: number, b: number): number {
    return (a * 3 + b) % 1000003;
}

function apply(f: (x: number) => number, x: number): number {
    return f(x);
}

var n: number = parseInt(process.argv[2], 10);

var direct = 0;
for (var i = 0; i < n; i++) {
    direct = inc(direct);
}

var nested = 0;
for (var i = 0; i < n; i++) {
    nested = mix(mix(nested, i), inc(i));
}

var f: (x: number) => number = inc;
var viaValue = 0;
for (var i = 0; i < n; i++) {
    viaValue = apply(f, viaValue);
}

var viaLambda = 0;
var g = (x: number) => (x * 7 + 1) % 1000003;
for (var i = 0; i < n; i++) {
    viaLambda = g(viaLambda);
}
process.stdout.write(direct + " " + nested + " " + viaValue + " " + viaLambda + "\n");
