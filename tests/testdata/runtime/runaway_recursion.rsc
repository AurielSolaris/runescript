// Mutual recursion between nested functions with no base case.
package tests.runtime;

public void main() {
    private int ping(int n);
    private int pong(int n) {
        return ping(n + 1);
    }
    private int ping(int n) {
        return pong(n); // expect runtime error: ASX-RUN-001 10:16
    }
    trace("start");
    trace(ping(0));
}
