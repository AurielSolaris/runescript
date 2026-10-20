// C02/C03/C06/C08: boundaries, overflow, and division by zero.
package bench;

public int store(int v) {
    return v;
}

protected int main() {
    int max = 2147483647;
    int min = -2147483648;
    int zero = 0;

    // Intermediates widen: the final result fits, so this is fine.
    trace("intermediate", (max + max) / 2, max * 4 / 4 - max);

    try { trace("max+1", store(max + 1)); } catch (RangeError e) { trace("max+1", e.message); }
    try { trace("min-1", store(min - 1)); } catch (RangeError e) { trace("min-1", e.message); }
    try { trace("max*2", store(max * 2)); } catch (RangeError e) { trace("max*2", e.message); }
    try { trace("-min", store(-min)); } catch (RangeError e) { trace("-min", e.message); }
    try { trace("max+max", store(max + max)); } catch (RangeError e) { trace("max+max", e.message); }
    try { trace("1/0", 1 / zero); } catch (RangeError e) { trace("1/0", e.message); }
    try { trace("1%0", 1 % zero); } catch (RangeError e) { trace("1%0", e.message); }

    uint8 b = 255;
    try { b += 1; trace("uint8", b); } catch (RangeError e) { trace("uint8", e.message); }
    int64 big = 9223372036854775807;
    try { big += 1; trace("int64", big); } catch (RangeError e) { trace("int64", e.message); }
    trace("2^53+1", (int64) 9007199254740993);
    return 0;
}
