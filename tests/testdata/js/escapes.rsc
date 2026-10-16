// Escape sequences (OQ-STR-02): \u{...}, \0, \', and the classic ones.
package tests.js;

use "std:string";

public void main() {
    string s = "caf\u{E9} \u{1F600} \u{2603}";
    trace(s, string.len(s), string.runeCount(s));
    char snow = '\u{2603}';
    trace(snow, (int) snow, (int) '\0', (int) '\u{10FFFF}');
    string nul = "a\0b";
    trace(string.len(nul), string.indexRune(nul, '\0'));
    trace("it\'s", '\'', "tab[\t] quote[\"] backslash[\\]");
    trace($"interp \u{E9}{s}\u{E8}");
}
