// An error that nothing catches stops the program at the throw.
// expect runtime error: ASX-RUN-005 9:5
package tests.runtime;

public void check(int n) {
    if (n < 10) {
        return;
    }
    throw new RangeError("n is too big", 3);
}

protected int main() {
    check(1);
    trace("checked 1");
    check(11);
    trace("not reached");
    return 0;
}
