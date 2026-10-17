// Function values, lambdas, and nested functions on the JavaScript target.
package tests.js;

public int(int)(int) chooser() {
    return (int a) => (int b) => b * 2;
}

public int apply(int(int) f, int x) {
    return f(x);
}

public int depth(int n) {
    if (n == 0) {
        return 0;
    }
    return depth(n - 1) + 1;
}

protected int main() {
    trace(chooser()(3)(4));
    trace(((int x) => x * 10)(5));
    trace(apply((int x) => { int y = x + 1; return y * y; }, 4));
    void(int) show = (int x) => x + 1;
    show(1);

    private int fact(int n) {
        if (n <= 1) {
            return 1;
        }
        return n * fact(n - 1);
    }
    trace(fact(12));
    int(int) f = fact;
    trace(apply(f, 5));
    trace("depth", depth(5000));

    int count = 0;
    for (;;) {
        count++;
        if (count < 3) {
            continue;
        }
        break;
    }
    while (count > 0) {
        count -= 1;
    }
    trace("count", count);
    return 0;
}
