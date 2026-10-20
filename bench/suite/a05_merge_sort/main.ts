// A05: top-down merge sort of n pseudo-random integers.
declare var process: any;

function sortRange(a: number[], tmp: number[], lo: number, hi: number): void {
    if (hi - lo < 2) {
        return;
    }
    var mid = Math.floor((lo + hi) / 2);
    sortRange(a, tmp, lo, mid);
    sortRange(a, tmp, mid, hi);
    var i = lo;
    var j = mid;
    var k = lo;
    while (i < mid && j < hi) {
        if (a[i] <= a[j]) {
            tmp[k] = a[i];
            i++;
        } else {
            tmp[k] = a[j];
            j++;
        }
        k++;
    }
    while (i < mid) {
        tmp[k] = a[i];
        i++;
        k++;
    }
    while (j < hi) {
        tmp[k] = a[j];
        j++;
        k++;
    }
    for (var m = lo; m < hi; m++) {
        a[m] = tmp[m];
    }
}

var n: number = parseInt(process.argv[2], 10);
var a: number[] = [];
var tmp: number[] = [];
var seed = 3;
for (var i = 0; i < n; i++) {
    seed = seed * 16807 % 2147483647;
    a.push(seed % 1000000);
    tmp.push(0);
}
sortRange(a, tmp, 0, n);
var sorted = true;
var check = 0;
for (var i = 0; i < n; i++) {
    if (i > 0 && a[i - 1] > a[i]) {
        sorted = false;
    }
    check = (check + a[i] % 1000 * (i % 7 + 1)) % 1000003;
}
process.stdout.write(sorted + " " + a[0] + " " + a[n - 1] + " " + check + "\n");
