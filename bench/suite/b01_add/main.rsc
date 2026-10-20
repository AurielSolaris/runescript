// B01a: integer addition throughput.
package bench;

use "std:process";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    int acc = 0;
    for (int i = 0; i < n; i++) {
        acc = acc + i;
        if (acc > 1000000000) {
            acc = acc - 1000000000;
        }
    }
    trace(acc);
    return 0;
}
