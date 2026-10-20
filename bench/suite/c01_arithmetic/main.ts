// C01/C06: integer arithmetic and division edge cases, with RuneScript's
// truncating integer division written out.
declare var process: any;

function trunc(v: number): number {
    return v < 0 ? Math.ceil(v) : Math.floor(v);
}

var xs = [7, -7, 0, 1, -1, 1000000, -1000000];
var ds = [2, -2, 3, 7];
for (var a = 0; a < xs.length; a++) {
    for (var b = 0; b < ds.length; b++) {
        var x = xs[a];
        var d = ds[b];
        // + 0 turns -0 into 0, as an integer has no negative zero.
        process.stdout.write(x + " " + d + " " + (x + d) + " " + (x - d) + " " + (x * d) + " " +
            (trunc(x / d) + 0) + " " + (x % d + 0) + "\n");
    }
}
