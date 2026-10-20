// D04: a call with too few arguments, then one with a wrong type.
package bench;

public int area(int width, int height) {
    return width * height;
}

protected int main() {
    trace(area(3));
    trace(area(3, "4"));
    return 0;
}
