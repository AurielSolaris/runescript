// A runtime error in a used file is reported in that file.
package app;

use "calc";

public int main() {
    trace("before");
    trace(calc.square(65536));
    return 0;
}
