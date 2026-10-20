// C01/C06: integer arithmetic and division edge cases.
package bench;

protected int main() {
    Array<int> xs = [7, -7, 0, 1, -1, 1000000, -1000000];
    Array<int> ds = [2, -2, 3, 7];
    for each (int x in xs) {
        for each (int d in ds) {
            trace(x, d, x + d, x - d, x * d, x / d, x % d);
        }
    }
    return 0;
}
