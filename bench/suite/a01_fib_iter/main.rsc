// A01: iterative Fibonacci, modulo 1,000,003.
package bench;

use "std:process";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    int a = 0;
    int b = 1;
    for (int i = 0; i < n; i++) {
        int next = (a + b) % 1000003;
        a = b;
        b = next;
    }
    trace(a);
    return 0;
}
