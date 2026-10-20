// A14: the Mandelbrot set on an n x n grid, at most 100 iterations.
package bench;

use "std:process";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    int inside = 0;
    int iterations = 0;
    for (int py = 0; py < n; py++) {
        float64 ci = (float64) py / (float64) n * 3.0 - 1.5;
        for (int px = 0; px < n; px++) {
            float64 cr = (float64) px / (float64) n * 3.0 - 2.0;
            float64 zr = 0.0;
            float64 zi = 0.0;
            int k = 0;
            while (k < 100 && zr * zr + zi * zi <= 4.0) {
                float64 t = zr * zr - zi * zi + cr;
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
    trace(inside, iterations);
    return 0;
}
