// A06: in-place quicksort (Hoare partition, middle pivot).
package bench;

use "std:process";

public void quicksort(Array<int> a, int lo, int hi) {
    if (lo >= hi) {
        return;
    }
    int pivot = a[(lo + hi) / 2];
    int i = lo;
    int j = hi;
    while (i <= j) {
        while (a[i] < pivot) {
            i++;
        }
        while (a[j] > pivot) {
            j--;
        }
        if (i <= j) {
            int t = a[i];
            a[i] = a[j];
            a[j] = t;
            i++;
            j--;
        }
    }
    quicksort(a, lo, j);
    quicksort(a, i, hi);
}

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    Array<int> a = [];
    int seed = 5;
    for (int i = 0; i < n; i++) {
        seed = seed * 16807 % 2147483647;
        a.push(seed % 1000000);
    }
    quicksort(a, 0, n - 1);
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
