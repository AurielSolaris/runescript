// B07: concatenation, searching, splitting, joining, and replacing.
package bench;

use "std:process";
use "std:string";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    Array<string> words = ["alpha", "beta", "gamma", "delta", "rune"];

    string text = "";
    for (int i = 0; i < n; i++) {
        text = text + words[i % 5] + " ";
    }

    int runes = string.count(text, "rune");
    int firstGamma = string.index(text, "gamma");
    int lastBeta = string.lastIndex(text, "beta");

    Array<string> parts = string.fields(text);
    int prefixed = 0;
    for each (string w in parts) {
        if (string.hasPrefix(w, "d") || string.hasSuffix(w, "a")) {
            prefixed++;
        }
    }
    string joined = string.join(parts, ",");
    string replaced = string.replaceAll(joined, "beta", "BETA");
    trace(string.len(text), runes, firstGamma, lastBeta, parts.length, prefixed,
        string.len(joined), string.count(replaced, "BETA"));
    return 0;
}
