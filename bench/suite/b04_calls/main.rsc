// B04: direct calls, nested calls, and calls through function values.
package bench;

use "std:process";

public int inc(int x) {
    return x + 1;
}

public int mix(int a, int b) {
    return (a * 3 + b) % 1000003;
}

public int apply(int(int) f, int x) {
    return f(x);
}

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];

    int direct = 0;
    for (int i = 0; i < n; i++) {
        direct = inc(direct);
    }

    int nested = 0;
    for (int i = 0; i < n; i++) {
        nested = mix(mix(nested, i), inc(i));
    }

    int(int) f = inc;
    int viaValue = 0;
    for (int i = 0; i < n; i++) {
        viaValue = apply(f, viaValue);
    }

    int viaLambda = 0;
    int(int) g = (int x) => (x * 7 + 1) % 1000003;
    for (int i = 0; i < n; i++) {
        viaLambda = g(viaLambda);
    }
    trace(direct, nested, viaValue, viaLambda);
    return 0;
}
