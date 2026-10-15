// Floats and conversions: a temperature table.
package examples;

public float toFahrenheit(float celsius) {
    return celsius * 9 / 5 + 32;
}

public int main() {
    for (int c = -40; c <= 100; c += 35) {
        float f = toFahrenheit((float) c);
        trace(c, "C is", f, "F");
    }

    // float32 keeps 32-bit precision; float keeps every digit.
    float exact = 1.0 / 3;
    float32 rounded = (float32) exact;
    trace(exact, rounded);

    // Conversions are checked and truncate toward zero.
    trace((int) toFahrenheit(37.5), (int) -2.9, (float) "98.6");
    return 0;
}
