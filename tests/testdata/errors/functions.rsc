// Function rules: boundaries, prototypes, function types, and calls.
package tests.errors;

public int twice(int);                 // never defined
// expect: ASX-NAME-010 4:12

public int half(int);
public int half(string s) {            // expect: ASX-TYPE-015 8:12
    return 0;
}

public int apply(int(int) f, int x) {
    return f(x);
}

protected int main() {
    int base = 10;
    private int addBase(int x) {
        return x + base;               // expect: ASX-NAME-009 19:20
    }
    int(int) g = (int x) => x * base;  // expect: ASX-NAME-009 21:33
    int(int) h = (string s) => 1;      // expect: ASX-TYPE-001 22:18
    trace(apply(twice, 1, 2));         // expect: ASX-TYPE-003 23:11
    trace((base)(1));                  // expect: ASX-TYPE-016 24:11
    trace(base(1));                    // expect: ASX-NAME-003 25:11
    int(int, int) k = twice;           // expect: ASX-TYPE-001 26:23
    trace(g);                          // expect: ASX-TYPE-012 27:11
    while (true) {
        private void stop() {
            break;                     // expect: ASX-FLOW-001 30:13
        }
        break;
    }
    return 0;
}
