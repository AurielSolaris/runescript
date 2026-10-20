// A11: a softened three-body simulation (symplectic Euler) in the plane.
package bench;

use "std:math";
use "std:process";

public float64 energy(Array<float64> x, Array<float64> y, Array<float64> vx, Array<float64> vy, Array<float64> m) {
    float64 e = 0.0;
    for (int i = 0; i < 3; i++) {
        e += 0.5 * m[i] * (vx[i] * vx[i] + vy[i] * vy[i]);
        for (int j = i + 1; j < 3; j++) {
            float64 dx = x[i] - x[j];
            float64 dy = y[i] - y[j];
            e -= m[i] * m[j] / math.sqrt(dx * dx + dy * dy + 0.01);
        }
    }
    return e;
}

protected int main() {
    Array<string> args = process.args();
    int steps = (int) args[0];
    Array<float64> x = [-1.0, 1.0, 0.0];
    Array<float64> y = [0.0, 0.0, 0.5];
    Array<float64> vx = [0.3471, 0.3471, -0.6942];
    Array<float64> vy = [0.5327, 0.5327, -1.0654];
    Array<float64> m = [1.0, 1.0, 1.0];
    float64 dt = 0.001;
    trace(energy(x, y, vx, vy, m));
    for (int s = 0; s < steps; s++) {
        for (int i = 0; i < 3; i++) {
            float64 ax = 0.0;
            float64 ay = 0.0;
            for (int j = 0; j < 3; j++) {
                if (i == j) {
                    continue;
                }
                float64 dx = x[j] - x[i];
                float64 dy = y[j] - y[i];
                float64 d2 = dx * dx + dy * dy + 0.01;
                float64 inv = m[j] / (d2 * math.sqrt(d2));
                ax += dx * inv;
                ay += dy * inv;
            }
            vx[i] += ax * dt;
            vy[i] += ay * dt;
        }
        for (int i = 0; i < 3; i++) {
            x[i] += vx[i] * dt;
            y[i] += vy[i] * dt;
        }
    }
    trace(energy(x, y, vx, vy, m));
    trace(x[0], y[0]);
    return 0;
}
