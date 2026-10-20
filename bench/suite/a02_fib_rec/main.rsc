// A02: naive recursive Fibonacci.
package bench;

use "std:process";

public int fib(int n) {
    if (n < 2) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

protected int main() {
    Array<string> args = process.args();
    trace(fib((int) args[0]));
    return 0;
}
