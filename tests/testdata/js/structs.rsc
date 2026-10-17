// Structs (OQ-STRUCT-01): value semantics, literals, fields, nesting,
// arrays of structs, equality, and trace.
package tests.js;

public struct Point {
    int x;
    int y;
}

public struct Rect {
    Point origin;
    Point size;
    string label;
}

private struct Inventory {
    Array<string> items;
    int count;
}

public Point move(Point p, int dx) {
    Point q = p;          // a copy
    q.x += dx;
    return q;
}

public int area(Rect r) {
    return r.size.x * r.size.y;
}

protected int main() {
    Point a = Point{x: 1, y: 2};
    Point b = a;          // a copy: changing b leaves a alone
    b.x = 10;
    trace(a, b, a == b, a == Point{y: 2, x: 1});

    Point c = move(a, 5);
    trace(a.x, c.x);

    Rect r = Rect{origin: a, size: Point{x: 3, y: 4}, label: "box"};
    r.origin.y = 99;      // a nested field
    trace(r, area(r), a.y);

    Array<Point> path = [a, b, Point{x: 7, y: 7}];
    path[0].x = -1;       // the element changes, a does not
    path[2].y *= 3;
    Point first = path[0];
    first.y = 0;
    trace(path[0], path[2], a, first);

    // Arrays inside a struct are references, as everywhere.
    Inventory inv = Inventory{items: ["sword"], count: 1};
    Inventory copy = inv;
    copy.items.push("shield");
    copy.count = 2;
    trace(inv.items.length, inv.count, copy.count);

    // Fields are evaluated in the order the literal is written.
    Point order = Point{y: tick("y"), x: tick("x")};
    trace(order);

    float32 f = (float32) 0.1;
    trace(Measure{value: f, unit: 'm', ok: true});
    return 0;
}

public struct Measure {
    float32 value;
    char unit;
    bool ok;
}

public int tick(string name) {
    trace("evaluating", name);
    return 1;
}
