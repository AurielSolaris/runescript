// A caught runtime error that is thrown again, and then not caught, is
// reported where it first happened.
// expect runtime error: ASX-RUN-002 9:19
package tests.runtime;

protected int main() {
    try {
        int big = 2147483647;
        trace(big * 2);
    } catch (RangeError e) {
        trace("rethrowing");
        throw e;
    }
    return 0;
}
