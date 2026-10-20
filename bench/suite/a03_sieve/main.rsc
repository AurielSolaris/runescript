// A03: Sieve of Eratosthenes; counts the primes up to n.
package bench;

use "std:process";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    Array<bool> composite = [];
    for (int i = 0; i <= n; i++) {
        composite.push(false);
    }
    int count = 0;
    int last = 0;
    for (int i = 2; i <= n; i++) {
        if (composite[i]) {
            continue;
        }
        count++;
        last = i;
        // i * i would overflow int for i > 46340: only sieve while it fits.
        if (i <= n / i) {
            for (int j = i * i; j <= n; j += i) {
                composite[j] = true;
            }
        }
    }
    trace(count, last);
    return 0;
}
