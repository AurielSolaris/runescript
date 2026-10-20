// S10 / B10: throwing and catching recoverable errors, with RuneScript's
// overflow and conversion checks written out.
declare var process: any;

function checkedAdd(a: number, b: number): number {
    var r = a + b;
    if (r > 2147483647 || r < -2147483648) {
        throw new RangeError("integer overflow: " + a + " + " + b + " = " + r + " does not fit in 'int'");
    }
    return r;
}

function parse(s: string): number {
    if (!/^[+-]?[0-9]+$/.test(s)) {
        throw new TypeError("cannot convert '" + s + "' to 'int'");
    }
    return parseInt(s, 10);
}

var n: number = parseInt(process.argv[2], 10);
var ok = 0;
var overflows = 0;
var badInput = 0;
var inputs = ["12", "x", "-7", "3.5e", "99"];
for (var i = 0; i < n; i++) {
    try {
        ok = (ok + checkedAdd(2147483600, i % 100)) % 1000003;
    } catch (e) {
        if (!(e instanceof RangeError)) {
            throw e;
        }
        overflows++;
    }
    try {
        ok = (ok + parse(inputs[i % 5])) % 1000003;
    } catch (e) {
        badInput++;
    }
}
process.stdout.write(ok + " " + overflows + " " + badInput + "\n");
