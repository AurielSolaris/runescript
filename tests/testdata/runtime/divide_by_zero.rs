// Division by zero inside a nested function.
package tests.runtime;

public int main() {
    private int average(int total, int count) {
        return total / count; // expect runtime error: ASX-RUN-003 6:22
    }
    trace(average(10, 2));
    trace(average(10, 0));
    return 0;
}
