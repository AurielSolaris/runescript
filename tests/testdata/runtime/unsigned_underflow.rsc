// Unsigned arithmetic below zero is an error, not a wrap to 4294967295.
package tests.runtime;

protected int main() {
    uint u = 1;
    u -= 1;
    trace(u);
    u -= 1; // expect runtime error: ASX-RUN-002 8:5
    return 0;
}
