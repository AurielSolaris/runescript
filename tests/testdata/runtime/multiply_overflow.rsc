// The exact product appears in the message, even where a double cannot hold it.
package tests.runtime;

protected int main() {
    int big = 2147483647;
    trace("before");
    trace(big * big); // expect runtime error: ASX-RUN-002 7:15
    return 0;
}
