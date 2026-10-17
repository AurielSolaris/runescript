// The std:string standard library package (Go's strings, camelCase names).
package examples.strings;

use "std:string";

public string slug(string title) {
    string s = string.toLower(string.trimSpace(title));
    return string.replaceAll(s, " ", "-");
}

protected int main() {
    string title = "   Hello RuneScript World   ";
    trace("slug:", slug(title));

    string path = "assets/runescript.svg";
    if (string.hasSuffix(path, ".svg")) {
        trace("image:", string.trimSuffix(path, ".svg"));
    }

    int slash = string.index(path, "/");
    trace("first slash at byte", slash);
    trace("vowels in title:", string.count(title, "o") + string.count(title, "e"));
    trace(string.repeat("=", 20));
    return 0;
}
