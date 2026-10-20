// A04: n pseudo-random lookups in a sorted array of n elements.
declare var process: any;

function search(a: number[], key: number): number {
    var lo = 0;
    var hi = a.length - 1;
    while (lo <= hi) {
        var mid = Math.floor((lo + hi) / 2);
        if (a[mid] < key) {
            lo = mid + 1;
        } else if (a[mid] > key) {
            hi = mid - 1;
        } else {
            return mid;
        }
    }
    return -1;
}

var n: number = parseInt(process.argv[2], 10);
var a: number[] = [];
for (var i = 0; i < n; i++) {
    a.push(i * 3);
}
var seed = 11;
var found = 0;
var check = 0;
for (var i = 0; i < n; i++) {
    seed = seed * 16807 % 2147483647;
    var at = search(a, seed % (n * 3));
    if (at >= 0) {
        found++;
        check = (check + at) % 1000003;
    }
}
process.stdout.write(found + " " + check + "\n");
