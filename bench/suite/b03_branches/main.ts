// B03: predictable, alternating, and pseudo-random branches.
declare var process: any;

var n: number = parseInt(process.argv[2], 10);

var predictable = 0;
for (var i = 0; i < n; i++) {
    if (i >= 0) {
        predictable++;
    } else {
        predictable--;
    }
}

var alternating = 0;
for (var i = 0; i < n; i++) {
    if (i % 2 == 0) {
        alternating += 3;
    } else {
        alternating -= 1;
    }
}

var random = 0;
var seed = 42;
for (var i = 0; i < n; i++) {
    seed = seed * 16807 % 2147483647;
    if (seed % 2 == 0) {
        random += 3;
    } else if (seed % 3 == 0) {
        random -= 2;
    } else {
        random += 1;
    }
}
process.stdout.write(predictable + " " + alternating + " " + random + " " + seed + "\n");
