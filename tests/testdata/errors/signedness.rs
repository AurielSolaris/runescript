// Every way a signed value can reach an unsigned destination is rejected,
// as is the reverse direction (whose conversion rules are not yet decided).
package tests.errors;

uint fromReturn(int s) {
    return s; // expect: ASX-TYPE-004 6:12
}

void takesUnsigned(uint u) {
}

int main() {
    sint s = 5;                 // positive at runtime, but its type is signed
    uint a = s;                 // expect: ASX-TYPE-004 14:14
    uint b = 0;
    b = s;                      // expect: ASX-TYPE-004 16:9
    takesUnsigned(s);           // expect: ASX-TYPE-004 17:19
    uint c = -1;                // expect: ASX-TYPE-007 18:14
    int d = b;                  // expect: ASX-TYPE-005 19:13
    trace(s + b);               // expect: ASX-TYPE-011 20:13
    return 0;
}
