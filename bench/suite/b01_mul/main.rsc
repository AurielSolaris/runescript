// B01b: integer multiply and remainder throughput.
package bench;

use "std:process";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    int acc = 1;
    for (int i = 0; i < n; i++) {
        acc = (acc * 31 + i) % 1000003;
    }
    trace(acc);
    return 0;
}
