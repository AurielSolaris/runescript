// B05: appends, sequential reads and writes, and random reads.
package bench;

use "std:process";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    Array<int> a = [];
    for (int i = 0; i < n; i++) {
        a.push(i % 1000);
    }

    int sum = 0;
    for (int i = 0; i < n; i++) {
        sum = (sum + a[i]) % 1000003;
    }

    for (int i = 0; i < n; i++) {
        a[i] = (a[i] * 3 + 1) % 1000;
    }

    int seed = 7;
    int picked = 0;
    for (int i = 0; i < n; i++) {
        seed = seed * 16807 % 2147483647;
        picked = (picked + a[seed % n]) % 1000003;
    }
    trace(a.length, sum, picked);
    return 0;
}
