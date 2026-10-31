// Classes: no null, `this` is required, members are checked.
package tests.errors;

public class Point {
    public int x;
    public Error last;          // expect: ASX-TYPE-028 6:18
    private int y = this.x;     // expect: ASX-NAME-001 7:21

    public Point(int x) {
        this.x = x;
    }

    public int sum() {
        return x + this.y;      // expect: ASX-NAME-001 14:16
    }
}

protected int main() {
    Point p;                    // expect: ASX-TYPE-028 19:11
    Point q = new Point();      // expect: ASX-TYPE-003 20:15
    trace(q.z);                 // expect: ASX-TYPE-020 21:13
    trace(q == 1);              // expect: ASX-TYPE-012 22:13
    trace(this);                // expect: ASX-NAME-001 23:11
    return 0;
}
