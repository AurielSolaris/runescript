// B01b, checked: every stored result is range-checked like RuneScript's int.
declare var process: any;

function int32(v: number): number {
    if (v > 2147483647 || v < -2147483648) {
        throw new RangeError("integer overflow: " + v + " does not fit in 'int'");
    }
    return v;
}

var n: number = parseInt(process.argv[2], 10);
var acc = 1;
for (var i = 0; i < n; i = int32(i + 1)) {
    acc = int32((acc * 31 + i) % 1000003);
}
process.stdout.write(acc + "\n");
