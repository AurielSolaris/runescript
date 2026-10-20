// A14: the Mandelbrot set on an n x n grid, at most 100 iterations.
declare var process: any;

var n: number = parseInt(process.argv[2], 10);
var inside = 0;
var iterations = 0;
for (var py = 0; py < n; py++) {
    var ci = py / n * 3.0 - 1.5;
    for (var px = 0; px < n; px++) {
        var cr = px / n * 3.0 - 2.0;
        var zr = 0.0;
        var zi = 0.0;
        var k = 0;
        while (k < 100 && zr * zr + zi * zi <= 4.0) {
            var t = zr * zr - zi * zi + cr;
            zi = 2.0 * zr * zi + ci;
            zr = t;
            k++;
        }
        if (k == 100) {
            inside++;
        }
        iterations += k;
    }
}
process.stdout.write(inside + " " + iterations + "\n");
