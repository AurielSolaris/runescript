// string.len, string.runeCount, string.slice (UTF-8 bytes, OQ-STR-03).
package tests.js;
use "std:string";
public void main() {
    string s = "héllo 🎉!";
    trace(string.len(s), string.runeCount(s), string.len(""));
    trace(string.slice(s, 0, 1), string.slice(s, 1, 3), string.slice(s, 7, 11), string.slice(s, 11, 12), string.slice(s, 3, 3) == "");
    try { trace(string.slice(s, 0, 2)); } catch (RangeError e) { trace(e.name, e.message, e.errorID); }
    try { trace(string.slice(s, 2, 1)); } catch (RangeError e) { trace(e.message); }
    try { trace(string.slice(s, -1, 2)); } catch (Error e) { trace(e.message); }
    try { trace(string.slice(s, 0, 99)); } catch (Error e) { trace(e.message); }
    try { trace(string.slice(s, 8, 12)); } catch (Error e) { trace(e.message); }
}
