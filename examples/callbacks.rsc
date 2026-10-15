// Advanced functions: callbacks, lambdas, nested functions, prototypes,
// and functions that return functions.
package examples.callbacks;

// A prototype: the definition is further down.
public int sumWith(int(int) f, int n);

public void forEach(int from, int to, void(int) body) {
    for (int i = from; i <= to; i++) {
        body(i);
    }
}

public int(int) multiplier(int factor) {
    // Functions do not capture variables, so each factor gets its own function.
    private int times2(int x) { return x * 2; }
    private int times3(int x) { return x * 3; }
    if (factor == 2) {
        return times2;
    }
    if (factor == 3) {
        return times3;
    }
    return (int x) => x;
}

public int main() {
    // Pass a lambda as a callback.
    forEach(1, 3, (int i) => trace("item", i));

    // Store functions in variables and pass named functions around.
    int(int) square = (int x) => x * x;
    trace("sum of squares 1..4:", sumWith(square, 4));
    trace("sum of triples 1..5:", sumWith(multiplier(3), 5));
    trace("double 21:", multiplier(2)(21));

    // Nested functions, with a prototype for mutual recursion.
    private bool isEven(int n);
    private bool isOdd(int n) {
        if (n == 0) { return false; }
        return isEven(n - 1);
    }
    private bool isEven(int n) {
        if (n == 0) { return true; }
        return isOdd(n - 1);
    }
    trace("7 is odd:", isOdd(7));
    return 0;
}

public int sumWith(int(int) f, int n) {
    int total = 0;
    for (int i = 1; i <= n; i++) {
        total += f(i);
    }
    return total;
}
