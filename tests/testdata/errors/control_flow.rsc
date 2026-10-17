// Conditions must be bool, branches need loops, and every path must return.
package tests.errors;

public int pick(bool b) {
    if (b) {
        return 1;
    }
} // expect: ASX-TYPE-008 8:1

protected int main() {
    int n = 3;
    if (n) {             // expect: ASX-TYPE-013 12:9
    }
    while (n - 1) {      // expect: ASX-TYPE-013 14:12
    }
    break;               // expect: ASX-FLOW-001 16:5
    for (int i = 0; i < 3; i++) {
        int i = 4;       // expect: ASX-NAME-002 18:13
    }
    trace(i);            // expect: ASX-NAME-001 20:11
    return 0;
}
