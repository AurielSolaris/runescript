// B01c: integer division throughput.
package bench;

use "std:process";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    int acc = 0;
    for (int i = 0; i < n; i++) {
        acc = (acc + (i * 7 + 3) / (i % 13 + 1)) % 1000000007;
    }
    trace(acc);
    return 0;
}
