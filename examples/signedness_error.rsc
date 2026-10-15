// This program is intentionally rejected: RuneScript never silently
// converts a signed integer to an unsigned one.
//
//   runescript check examples/signedness_error.rsc
package examples.errors;

public int main() {
    sint signedValue = -10;
    uint unsignedValue = signedValue; // expect: ASX-TYPE-004 9:26
    return 0;
}
