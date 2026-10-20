// C13: string units. string.len/slice/index count UTF-8 bytes.
package bench;

use "std:string";

protected int main() {
    Array<string> samples = ["", "hello", "café", "é", "😀", "こんにちは", "e\u{301}"];
    for each (string s in samples) {
        int n = string.len(s);
        string head = "";
        if (n > 0) {
            try {
                head = string.slice(s, 0, 1);
            } catch (e) {
                head = e.message;
            }
        }
        trace($"[{s}] native={n} codepoints={string.runeCount(s)} head=[{head}] e@{string.index(s, "e")}");
    }
    trace("precomposed == combining:", "é" == "e\u{301}");
    return 0;
}
