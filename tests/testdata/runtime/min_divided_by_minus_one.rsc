// The one integer division whose quotient is out of range.
package tests.runtime;

protected int main() {
    int min = -2147483648;
    int m = -1;
    trace(min % m);
    trace(min / m); // expect runtime error: ASX-RUN-002 8:15
    return 0;
}
