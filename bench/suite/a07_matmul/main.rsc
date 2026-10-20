// A07: n x n float64 matrix multiplication, flat row-major arrays.
package bench;

use "std:process";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    Array<float64> a = [];
    Array<float64> b = [];
    Array<float64> c = [];
    for (int i = 0; i < n; i++) {
        for (int j = 0; j < n; j++) {
            a.push((float64) (i * j % 7) * 0.5);
            b.push((float64) ((i + j) % 5) * 0.25);
            c.push(0.0);
        }
    }
    for (int i = 0; i < n; i++) {
        for (int k = 0; k < n; k++) {
            float64 aik = a[i * n + k];
            for (int j = 0; j < n; j++) {
                c[i * n + j] += aik * b[k * n + j];
            }
        }
    }
    float64 sum = 0.0;
    float64 diag = 0.0;
    for (int i = 0; i < n * n; i++) {
        sum += c[i];
    }
    for (int i = 0; i < n; i++) {
        diag += c[i * n + i];
    }
    trace(sum, diag, c[n * n - 1]);
    return 0;
}
