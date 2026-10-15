// while, for, break, and continue.
package examples.loops;

void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

int main() {
    sayHi("Hello world", 3);

    int n = 0;
    while (n < 10) {
        n++;
        if (n % 2 == 0) {
            continue;
        }
        if (n > 7) {
            break;
        }
        trace("odd", n);
    }
    return 0;
}
