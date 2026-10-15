// A failing standard library call made through a function value.
package tests.runtime;

use "std:string";

public void main() {
    string(string, int) rep = (string s, int n) => string.repeat(s, n); // expect runtime error: ASX-RUN-004 7:52
    trace(rep("ab", 2));
    trace(rep("ab", -1));
}
