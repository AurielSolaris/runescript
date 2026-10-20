// A05: top-down merge sort of n pseudo-random integers.
package bench;

use "std:process";

public void sortRange(Array<int> a, Array<int> tmp, int lo, int hi) {
    if (hi - lo < 2) {
        return;
    }
    int mid = (lo + hi) / 2;
    sortRange(a, tmp, lo, mid);
    sortRange(a, tmp, mid, hi);
    int i = lo;
    int j = mid;
    int k = lo;
    while (i < mid && j < hi) {
        if (a[i] <= a[j]) {
            tmp[k] = a[i];
            i++;
        } else {
            tmp[k] = a[j];
            j++;
        }
        k++;
    }
    while (i < mid) {
        tmp[k] = a[i];
        i++;
        k++;
    }
    while (j < hi) {
        tmp[k] = a[j];
        j++;
        k++;
    }
    for (int m = lo; m < hi; m++) {
        a[m] = tmp[m];
    }
}

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    Array<int> a = [];
    Array<int> tmp = [];
    int seed = 3;
    for (int i = 0; i < n; i++) {
        seed = seed * 16807 % 2147483647;
        a.push(seed % 1000000);
        tmp.push(0);
    }
    sortRange(a, tmp, 0, n);
    bool sorted = true;
    int check = 0;
    for (int i = 0; i < n; i++) {
        if (i > 0 && a[i - 1] > a[i]) {
            sorted = false;
        }
        check = (check + a[i] % 1000 * (i % 7 + 1)) % 1000003;
    }
    trace(sorted, a[0], a[n - 1], check);
    return 0;
}
