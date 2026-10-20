// B03: predictable, alternating, and pseudo-random branches.
package bench;

use "std:process";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];

    int predictable = 0;
    for (int i = 0; i < n; i++) {
        if (i >= 0) {
            predictable++;
        } else {
            predictable--;
        }
    }

    int alternating = 0;
    for (int i = 0; i < n; i++) {
        if (i % 2 == 0) {
            alternating += 3;
        } else {
            alternating -= 1;
        }
    }

    int random = 0;
    int seed = 42;
    for (int i = 0; i < n; i++) {
        seed = seed * 16807 % 2147483647;
        if (seed % 2 == 0) {
            random += 3;
        } else if (seed % 3 == 0) {
            random -= 2;
        } else {
            random += 1;
        }
    }
    trace(predictable, alternating, random, seed);
    return 0;
}
