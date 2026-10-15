// if / else if / else, comparisons, and boolean operators.
package examples.conditionals;

string classify(int n) {
    if (n < 0) {
        return "negative";
    } else if (n == 0) {
        return "zero";
    } else if (n % 2 == 0 && n > 100) {
        return "large even";
    } else {
        return "positive";
    }
}

void main() {
    trace(-5, classify(-5));
    trace(0, classify(0));
    trace(7, classify(7));
    trace(200, classify(200));

    bool a = true;
    bool b = !a || 3 >= 2;
    trace("b is", b);
}
