// C02/C03/C06/C08: the same expressions with TypeScript's number (a
// double): nothing overflows at 2^31, division by zero is Infinity, and
// integers above 2^53 are rounded.
declare var process: any;

function out(label: string, v: any): void {
    process.stdout.write(label + " " + v + "\n");
}

var max = 2147483647;
var min = -2147483648;
var zero = 0;

process.stdout.write("intermediate " + (max + max) / 2 + " " + (max * 4 / 4 - max) + "\n");
out("max+1", max + 1);
out("min-1", min - 1);
out("max*2", max * 2);
out("-min", -min);
out("max+max", max + max);
out("1/0", 1 / zero);
out("1%0", 1 % zero);

var b = 255;
b += 1;
out("uint8", b);
var big = 9223372036854775807;
big += 1;
out("int64", big);
out("2^53+1", 9007199254740993);
