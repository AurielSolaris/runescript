// char (OQ-CHAR-01): literals, comparison, conversions, and the
// std:string functions that take a char or a char callback.
package tests.js;

use "std:string";

public bool isDigit(char c) {
    return c >= '0' && c <= '9';
}

public char shout(char c) {
    if (c >= 'a' && c <= 'z') {
        return (char) ((int) c - 32);
    }
    return c;
}

public void main() {
    char a = 'a';
    char e = 'é';
    char smile = '😀';
    char quote = '\'';
    trace(a, e, smile, quote, '\\', '"', (int) '\n');
    trace(a == 'a', a != e, a < e, e < smile, 'Z' < 'a');
    trace((int) a, (uint) smile, (char) 98, (char) "é", (string) smile + "!");
    trace($"char {a} code {(int) a}");

    string s = "  42 apples, 7 pears  ";
    trace(string.containsRune(s, 'p'), string.containsRune(s, 'z'), string.indexRune("héllo", 'l'));
    trace(string.indexFunc(s, isDigit), string.lastIndexFunc(s, isDigit), string.containsFunc("abc", isDigit));
    trace(string.trimFunc("123abc456", isDigit), string.trimLeftFunc("123abc456", isDigit), string.trimRightFunc("123abc456", isDigit));
    trace(string.map(shout, "hello, wörld"));
    trace(string.map((char c) => {
        if (c == ' ') {
            return '_';
        }
        return c;
    }, "a b c"));

    bool(char) c = isDigit; // a callback in a variable named c
    trace(string.indexFunc("ab3", c));

    // Callbacks run in order; a callback's error goes through.
    trace(string.trimFunc("xxaxx", (char c) => {
        trace("saw", c);
        return c == 'x';
    }));
    try {
        trace(string.indexFunc("abc", (char c) => {
            if (c == 'b') {
                throw new ArgumentError("no b");
            }
            return false;
        }));
    } catch (ArgumentError err) {
        trace("caught", err.message);
    }

    try {
        trace((char) 55296);
    } catch (RangeError err) {
        trace(err.message);
    }
    try {
        trace((char) "ab");
    } catch (TypeError err) {
        trace(err.message);
    }
    int big = 1114112;
    try {
        trace((char) big);
    } catch (RangeError err) {
        trace(err.message);
    }
}
