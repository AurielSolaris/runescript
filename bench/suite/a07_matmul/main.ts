// A07: n x n float64 matrix multiplication, flat row-major arrays.
declare var process: any;

var n: number = parseInt(process.argv[2], 10);
var a: number[] = [];
var b: number[] = [];
var c: number[] = [];
for (var i = 0; i < n; i++) {
    for (var j = 0; j < n; j++) {
        a.push((i * j % 7) * 0.5);
        b.push(((i + j) % 5) * 0.25);
        c.push(0.0);
    }
}
for (var i = 0; i < n; i++) {
    for (var k = 0; k < n; k++) {
        var aik = a[i * n + k];
        for (var j = 0; j < n; j++) {
            c[i * n + j] += aik * b[k * n + j];
        }
    }
}
var sum = 0.0;
var diag = 0.0;
for (var i = 0; i < n * n; i++) {
    sum += c[i];
}
for (var i = 0; i < n; i++) {
    diag += c[i * n + i];
}
process.stdout.write(sum + " " + diag + " " + c[n * n - 1] + "\n");
