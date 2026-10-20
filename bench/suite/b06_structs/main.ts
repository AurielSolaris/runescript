// B06: struct copies, field updates, and structs passed and returned by value.
// TypeScript objects are references: every RuneScript value copy is an
// explicit copy here.
declare var process: any;

interface Vec {
    x: number;
    y: number;
}

interface Particle {
    pos: Vec;
    vel: Vec;
    id: number;
}

function copyVec(v: Vec): Vec {
    return { x: v.x, y: v.y };
}

function copyParticle(p: Particle): Particle {
    return { pos: copyVec(p.pos), vel: copyVec(p.vel), id: p.id };
}

function add(a: Vec, b: Vec): Vec {
    return { x: (a.x + b.x) % 10007, y: (a.y + b.y) % 10007 };
}

function step(p: Particle): Particle {
    var q = copyParticle(p);
    q.pos = add(q.pos, q.vel);
    q.vel.x = (q.vel.x * 3 + 1) % 101;
    return q;
}

var n: number = parseInt(process.argv[2], 10);
var p: Particle = { pos: { x: 0, y: 0 }, vel: { x: 1, y: 2 }, id: 1 };
var start = copyParticle(p);
var check = 0;
for (var i = 0; i < n; i++) {
    p = step(p);
    var copy = copyVec(p.pos);
    copy.x = 0;
    check = (check + p.pos.x + copy.y) % 1000003;
}
process.stdout.write(p.pos.x + " " + p.pos.y + " " + p.vel.x + " " + start.pos.x + " " + check + "\n");
