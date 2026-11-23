package main;
use "geom";

protected int main() {
    geom.Point a = new geom.Point(4);
    geom.Point b = a;
    b.x += 10;
    geom.Point literal = geom.Point{x: 7, y: 8};
    trace(a.sum(), b.sum(), literal.sum(), new geom.Point(2).sum());
    geom.Callback scale = new geom.Callback((int n) => n * 3);
    trace(scale.apply(4));
    return 0;
}
