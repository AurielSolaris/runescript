// Overflow inside a lambda passed as a callback is reported at the lambda.
package tests.runtime;

public int fold(int(int, int) step, int start, int times) {
    int acc = start;
    for (int i = 0; i < times; i++) {
        acc = step(acc, i);
        trace("acc", acc);
    }
    return acc;
}

public void main() {
    trace(fold((int acc, int i) => acc * 1000, 3, 5)); // expect runtime error: ASX-RUN-002 14:40
}
