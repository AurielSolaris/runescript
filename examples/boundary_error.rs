// This program is intentionally rejected: nested functions and lambdas do
// not capture the variables of the function around them.
//
//   runescript check examples/boundary_error.rs
package examples.errors;

public int main() {
    int total = 0;
    private void add(int n) {
        total += n; // expect: ASX-NAME-009 10:9
    }
    add(5);
    return total;
}
