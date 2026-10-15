// A failed conversion that nothing catches.
// expect runtime error: ASX-RUN-006 8:11
package tests.runtime;

public int main() {
    string s = "12abc";
    trace("converting");
    trace((int) s);
    return 0;
}
