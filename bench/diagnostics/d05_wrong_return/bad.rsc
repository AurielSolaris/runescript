// D05: a function declared to return int returns a string.
package bench;

public int label(int n) {
    return "item " + n;
}

protected int main() {
    trace(label(1));
    return 0;
}
