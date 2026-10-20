// S10 / B10: throwing and catching recoverable errors.
package bench;

use "std:process";

public int checkedAdd(int a, int b) {
    return a + b;
}

public int parse(string s) {
    return (int) s;
}

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    int ok = 0;
    int overflows = 0;
    int badInput = 0;
    Array<string> inputs = ["12", "x", "-7", "3.5e", "99"];
    for (int i = 0; i < n; i++) {
        try {
            ok = (ok + checkedAdd(2147483600, i % 100)) % 1000003;
        } catch (RangeError e) {
            overflows++;
        }
        try {
            ok = (ok + parse(inputs[i % 5])) % 1000003;
        } catch (e) {
            badInput++;
        }
    }
    trace(ok, overflows, badInput);
    return 0;
}
