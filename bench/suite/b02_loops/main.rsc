// B02: loop overhead. n is the side of an n x n nested loop.
package bench;

use "std:process";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];

    int counted = 0;
    for (int i = 0; i < n * n; i++) {
        counted++;
    }

    int nested = 0;
    for (int i = 0; i < n; i++) {
        for (int j = 0; j < n; j++) {
            nested++;
        }
    }

    int whiles = 0;
    int k = n * n;
    while (k > 0) {
        k--;
        whiles++;
    }

    int skipped = 0;
    for (int i = 0; i < n * n; i++) {
        if (i % 3 == 0) {
            continue;
        }
        if (i > n * n - 2) {
            break;
        }
        skipped++;
    }
    trace(counted, nested, whiles, skipped);
    return 0;
}
