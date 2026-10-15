// An error thrown 9000 calls deep unwinds through a finally block in every
// frame. (The interpreter's unwinding is linear in the depth; see
// internal/runtime TestDeepUnwindIsLinear.)
package tests.js;

public int down(int n) {
    try {
        if (n == 9000) {
            throw new Error("bottom");
        }
        return down(n + 1);
    } finally {
        if (n % 3000 == 0) {
            trace("finally", n);
        }
    }
}

public int main() {
    try {
        down(0);
    } catch (Error e) {
        trace(e);
    }
    return 0;
}
