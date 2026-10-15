// The classic.
package examples.fizzbuzz;

void fizzbuzz(int limit) {
    for (int i = 1; i <= limit; i++) {
        if (i % 15 == 0) {
            trace("FizzBuzz");
        } else if (i % 3 == 0) {
            trace("Fizz");
        } else if (i % 5 == 0) {
            trace("Buzz");
        } else {
            trace(i);
        }
    }
}

int main() {
    fizzbuzz(15);
    return 0;
}
