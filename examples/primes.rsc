// Counting and charting primes with predicate callbacks, a lambda,
// a nested helper function, and std:string.
package examples.primes;

use "std:string";

public bool isPrime(int n) {
    if (n < 2) {
        return false;
    }
    for (int d = 2; d * d <= n; d++) {
        if (n % d == 0) {
            return false;
        }
    }
    return true;
}

// Counts the numbers in [from, to] for which pred returns true.
public int countWhere(int from, int to, bool(int) pred) {
    int count = 0;
    for (int n = from; n <= to; n++) {
        if (pred(n)) {
            count++;
        }
    }
    return count;
}

protected int main() {
    trace("primes up to 100:", countWhere(1, 100, isPrime));
    trace("twin primes up to 100:", countWhere(1, 100, (int n) => isPrime(n) && isPrime(n + 2)));

    private void bar(int from, int count) {
        trace(from, string.repeat("#", count), count);
    }
    trace("primes per hundred:");
    for (int from = 0; from < 500; from += 100) {
        bar(from, countWhere(from, from + 99, isPrime));
    }
    return 0;
}
