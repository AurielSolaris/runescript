// Functions: parameters, return values, recursion, and calls before declaration.
package examples.functions;

protected int main() {
    trace("5! =", factorial(5));
    trace("fib(20) =", fib(20));
    trace("gcd(84, 36) =", gcd(84, 36));
    return 0;
}

public int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

public int fib(int n) {
    int a = 0;
    int b = 1;
    for (int i = 0; i < n; i++) {
        int next = a + b;
        a = b;
        b = next;
    }
    return a;
}

public uint gcd(uint a, uint b) {
    uint x = a;
    uint y = b;
    while (y != 0) {
        uint t = x % y;
        x = y;
        y = t;
    }
    return x;
}
