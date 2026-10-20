// A04: n pseudo-random lookups in a sorted array of n elements.
package bench;

use "std:process";

public int search(Array<int> a, int key) {
    int lo = 0;
    int hi = a.length - 1;
    while (lo <= hi) {
        int mid = (lo + hi) / 2;
        if (a[mid] < key) {
            lo = mid + 1;
        } else if (a[mid] > key) {
            hi = mid - 1;
        } else {
            return mid;
        }
    }
    return -1;
}

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    Array<int> a = [];
    for (int i = 0; i < n; i++) {
        a.push(i * 3);
    }
    int seed = 11;
    int found = 0;
    int check = 0;
    for (int i = 0; i < n; i++) {
        seed = seed * 16807 % 2147483647;
        int at = search(a, seed % (n * 3));
        if (at >= 0) {
            found++;
            check = (check + at) % 1000003;
        }
    }
    trace(found, check);
    return 0;
}
