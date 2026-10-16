// Bitwise operators and shifts (OQ-OP-02).
package tests.js;

public uint32 rotl(uint32 x, int n) {
    // Mask before shifting left, so no bits are lost (<< traps otherwise).
    uint32 low = x & ((((uint32) 1) << (32 - n)) - 1);
    return (low << n) | (x >> (32 - n));
}

public void main() {
    int flags = 0b0101;
    flags |= 0b1000;
    flags &= ~0b0001;
    flags ^= 0b0010;
    trace(flags, (flags & 0b1000) != 0);

    uint8 b = 0xF0;
    trace(~b, b >> 4, b & 0x3C, b | 0x0F, b ^ 0xFF);
    int8 neg = -128;
    trace(neg >> 1, neg >> 7, ~neg, -1 >> 3);

    uint64 big = 0xFFFF_FFFF_FFFF_FFFF;
    trace(big >> 60, big & 0xFF, ~big, (big >> 1) << 1);
    int64 sixtyThird = ((int64) 1) << 62;
    trace(sixtyThird, -sixtyThird >> 62);

    trace(rotl(0x8000_0001, 1), rotl(0x1234_5678, 8));
    trace((int64) (1 << 40), 0xFF >> 4);  // constants are exact
    uint u = 1;
    u <<= 31;
    trace(u, u >> 31);

    try {
        trace(u << 1);
    } catch (RangeError e) {
        trace(e.name, e.message);
    }
    int count = 33;
    try {
        trace(flags << count);
    } catch (RangeError e) {
        trace(e.message);
    }
    count = -1;
    try {
        trace(flags >> count);
    } catch (RangeError e) {
        trace(e.message);
    }
}
