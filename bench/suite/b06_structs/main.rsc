// B06: struct copies, field updates, and structs passed and returned by value.
package bench;

use "std:process";

public struct Vec {
    int x;
    int y;
}

public struct Particle {
    Vec pos;
    Vec vel;
    int id;
}

public Vec add(Vec a, Vec b) {
    return Vec{x: (a.x + b.x) % 10007, y: (a.y + b.y) % 10007};
}

public Particle step(Particle p) {
    Particle q = p;
    q.pos = add(q.pos, q.vel);
    q.vel.x = (q.vel.x * 3 + 1) % 101;
    return q;
}

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    Particle p = Particle{pos: Vec{x: 0, y: 0}, vel: Vec{x: 1, y: 2}, id: 1};
    Particle start = p;
    int check = 0;
    for (int i = 0; i < n; i++) {
        p = step(p);
        Vec copy = p.pos;
        copy.x = 0;
        check = (check + p.pos.x + copy.y) % 1000003;
    }
    trace(p.pos.x, p.pos.y, p.vel.x, start.pos.x, check);
    return 0;
}
