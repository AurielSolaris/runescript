// A06: in-place quicksort (Hoare partition, middle pivot).
declare var process: any;

function quicksort(a: number[], lo: number, hi: number): void {
    if (lo >= hi) {
        return;
    }
    var pivot = a[Math.floor((lo + hi) / 2)];
    var i = lo;
    var j = hi;
    while (i <= j) {
        while (a[i] < pivot) {
            i++;
        }
        while (a[j] > pivot) {
            j--;
        }
        if (i <= j) {
            var t = a[i];
            a[i] = a[j];
            a[j] = t;
            i++;
            j--;
        }
    }
    quicksort(a, lo, j);
    quicksort(a, i, hi);
}

var n: number = parseInt(process.argv[2], 10);
var a: number[] = [];
var seed = 5;
for (var i = 0; i < n; i++) {
    seed = seed * 16807 % 2147483647;
    a.push(seed % 1000000);
}
quicksort(a, 0, n - 1);
var sorted = true;
var check = 0;
for (var i = 0; i < n; i++) {
    if (i > 0 && a[i - 1] > a[i]) {
        sorted = false;
    }
    check = (check + a[i] % 1000 * (i % 7 + 1)) % 1000003;
}
process.stdout.write(sorted + " " + a[0] + " " + a[n - 1] + " " + check + "\n");
