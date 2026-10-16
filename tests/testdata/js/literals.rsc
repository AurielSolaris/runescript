// Number literal forms (OQ-NUM-06): hex, binary, octal, and separators.
package tests.js;

public void main() {
    int million = 1_000_000;
    uint mask = 0xFFFF_FFFF;
    int flags = 0b1010_0101;
    int perms = 0o755;
    float big = 1_234.567_8;
    float64 avogadro = 6.022_140_76e2_3;
    trace(million, mask, flags, perms, big, avogadro);
    trace(0xff + 0b1 + 0o10, -0x80, (uint) 0x7FFF_FFFF);
}
