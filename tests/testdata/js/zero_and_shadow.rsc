// Zero values for declarations without an initializer (OQ-VAR-01), and
// shadowing in nested blocks (OQ-SCOPE-01, ASX-WARN-009).
package tests.js;

public struct Point {
    int x;
    int y;
}

public struct Tagged {
    Point at;
    string tag;
    Array<int> marks;
    char c;
    bool on;
    float64 weight;
}

public int count() {
    Array<int> fresh;     // a new empty array on every call
    fresh.push(1);
    return fresh.length;
}

protected int main() {
    int i;
    uint u;
    int64 big;
    uint8 small;
    float f;
    float32 f32;
    string s;
    bool b;
    char c;
    trace(i, u, big, small, f, f32, "[" + s + "]", b, (int) c);

    Point p;
    trace(p);
    Tagged t;
    trace(t.at, "[" + t.tag + "]", t.marks.length, (int) t.c, t.on, t.weight);
    t.marks.push(3);
    Tagged t2;
    trace(t.marks.length, t2.marks.length);
    trace(count(), count());

    for (int n; n < 3; n++) {
        trace("n", n);
    }

    int x = 1;
    if (true) {
        int x = x + 10;   // the initializer reads the outer x
        trace("inner", x);
        if (x > 0) {
            string x = "deepest";
            trace(x);
        }
        x += 1;
        trace("inner again", x);
    }
    trace("outer", x);
    int(int) add = (int x) => x + 100;
    trace(add(x));
    return 0;
}
