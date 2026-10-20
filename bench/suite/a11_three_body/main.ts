// A11: a softened three-body simulation (symplectic Euler) in the plane.
declare var process: any;

function energy(x: number[], y: number[], vx: number[], vy: number[], m: number[]): number {
    var e = 0.0;
    for (var i = 0; i < 3; i++) {
        e += 0.5 * m[i] * (vx[i] * vx[i] + vy[i] * vy[i]);
        for (var j = i + 1; j < 3; j++) {
            var dx = x[i] - x[j];
            var dy = y[i] - y[j];
            e -= m[i] * m[j] / Math.sqrt(dx * dx + dy * dy + 0.01);
        }
    }
    return e;
}

var steps: number = parseInt(process.argv[2], 10);
var x = [-1.0, 1.0, 0.0];
var y = [0.0, 0.0, 0.5];
var vx = [0.3471, 0.3471, -0.6942];
var vy = [0.5327, 0.5327, -1.0654];
var m = [1.0, 1.0, 1.0];
var dt = 0.001;
process.stdout.write(energy(x, y, vx, vy, m) + "\n");
for (var s = 0; s < steps; s++) {
    for (var i = 0; i < 3; i++) {
        var ax = 0.0;
        var ay = 0.0;
        for (var j = 0; j < 3; j++) {
            if (i == j) {
                continue;
            }
            var dx = x[j] - x[i];
            var dy = y[j] - y[i];
            var d2 = dx * dx + dy * dy + 0.01;
            var inv = m[j] / (d2 * Math.sqrt(d2));
            ax += dx * inv;
            ay += dy * inv;
        }
        vx[i] += ax * dt;
        vy[i] += ay * dt;
    }
    for (var i = 0; i < 3; i++) {
        x[i] += vx[i] * dt;
        y[i] += vy[i] * dt;
    }
}
process.stdout.write(energy(x, y, vx, vy, m) + "\n");
process.stdout.write(x[0] + " " + y[0] + "\n");
