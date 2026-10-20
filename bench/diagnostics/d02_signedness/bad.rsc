// D02: a signed value assigned to an unsigned variable.
package bench;

protected int main() {
    sint signedValue = -10;
    uint unsignedValue = signedValue;
    trace(unsignedValue);
    return 0;
}
