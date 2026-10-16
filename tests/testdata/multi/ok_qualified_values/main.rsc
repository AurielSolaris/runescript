package main;

use "geom";

// A function of another package used as a value, without a call.
public int apply(int(int, int) f, int a, int b) {
    return f(a, b);
}

public int main() {
    int(int, int) area = geom.area;
    trace(area(3, 4), apply(geom.area, 5, 6));
    return 0;
}
