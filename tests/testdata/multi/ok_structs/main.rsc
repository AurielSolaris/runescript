package main;

use "shapes";

// Structs of another package are named through its qualifier.
public shapes.Point mid(shapes.Point a, shapes.Point b) {
    return shapes.Point{x: (a.x + b.x) / 2, y: (a.y + b.y) / 2};
}

protected int main() {
    shapes.Circle c = shapes.Circle{center: shapes.Point{x: 0, y: 0}, radius: 2};
    shapes.Circle d = shapes.grow(c, 3);
    d.center.x = 4;
    trace(c, d);
    trace(mid(c.center, d.center), shapes.gridStep());
    Array<shapes.Point> ps = [c.center];
    trace(ps[0] == c.center);
    return 0;
}
