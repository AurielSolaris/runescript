// B02: loop overhead. n is the side of an n x n nested loop.
declare var process: any;

var n: number = parseInt(process.argv[2], 10);

var counted = 0;
for (var i = 0; i < n * n; i++) {
    counted++;
}

var nested = 0;
for (var i = 0; i < n; i++) {
    for (var j = 0; j < n; j++) {
        nested++;
    }
}

var whiles = 0;
var k = n * n;
while (k > 0) {
    k--;
    whiles++;
}

var skipped = 0;
for (var i = 0; i < n * n; i++) {
    if (i % 3 == 0) {
        continue;
    }
    if (i > n * n - 2) {
        break;
    }
    skipped++;
}
process.stdout.write(counted + " " + nested + " " + whiles + " " + skipped + "\n");
