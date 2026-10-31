package main;

use "shapes";

// Classes of another package are named through its qualifier; only their
// public members are visible here.
public int areaOf(shapes.Circle c) {
    return c.area();
}

protected int main() {
    shapes.Circle c = new shapes.Circle(2);
    shapes.Circle same = c;
    same.radius = 3;
    trace(areaOf(c), c.radius, c);
    shapes.Circle big = shapes.doubled(c);
    trace(big.radius, big == c, shapes.count(big));
    return 0;
}
