// The call-depth limit cannot be caught, and finally blocks do not run
// while the program stops for it.
// expect runtime error: ASX-RUN-001 9:20
package tests.runtime;

public int down(int n) {
    try {
        try {
            return down(n + 1);
        } finally {
            trace("finally");
        }
    } catch (Error e) {
        trace("caught");
        return 0;
    }
}

protected int main() {
    trace("start");
    return down(0);
}
