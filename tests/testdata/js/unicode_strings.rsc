// std:string on non-ASCII text: positions are UTF-8 byte offsets, and case
// mapping, folding, and white space follow Go's unicode package on every
// target.
package tests.js;

use "std:string";

protected int main() {
    string price = "price: 5€, or 😀";
    trace(string.index(price, "€"), string.index(price, "or"), string.lastIndex(price, "😀"));
    trace(string.index(price, ""), string.lastIndex(price, ""), string.index(price, "$"));
    trace(string.indexAny(price, "😀€"), string.lastIndexAny(price, "€p"), string.containsAny(price, "xyz€"));
    trace(string.count("😀a😀", ""), string.count("aaaa", "aa"), string.count("ééé", "é"));

    // Case mapping is per code point (no "ß" -> "SS" expansion).
    trace(string.toUpper("straße ǆ ﬁ ı i"), string.toLower("İSTANBUL Σ ǅ"), string.toTitle("ǆemal ß"));
    trace(string.equalFold("Kelvin", "Kelvin"), string.equalFold("ſ", "S"), string.equalFold("ß", "SS"));
    trace(string.equalFold("Σ", "ς"), string.equalFold("abc", "ab"));

    // White space is Go's unicode.IsSpace: U+0085 and U+00A0 are spaces,
    // U+FEFF is not.
    trace("[", string.trimSpace("  x  "), string.trimSpace(" \tvalue "), "]");
    trace(string.hasPrefix(string.trimSpace("﻿x"), "﻿"));
    trace(string.trim("😀😀mid😀", "😀"), string.trimLeft("ééa", "é"), string.trimRight("aéé", "é"));

    // Code point order, which is UTF-8 byte order (not UTF-16 order).
    trace(string.compare("😀", "￿"), string.compare("a", "b"), string.compare("ab", "a"), string.compare("é", "é"));

    trace(string.replace("héllo", "", "-", 3), string.replaceAll("a😀b😀", "😀", "<>"));
    trace(string.replace("aaa", "a", "b", 2), string.replace("aaa", "a", "a", -1), string.repeat("é", 3));
    trace(string.toValidUTF8("ok ✓", "?"), string.clone("✓"));
    return 0;
}
