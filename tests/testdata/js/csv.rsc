// std:csv: parsing, quoting, separators, round trips, and the errors
// each function throws, the same on both targets.
package tests.js;

use "std:csv";
use "std:string";

public void attempt(string what, void() f) {
    try {
        f();
        trace(what, "ok");
    } catch (Error e) {
        trace(what, "->", e.name, e.message);
    }
}

// show prints each record's fields between brackets.
public void show(Array<Array<string>> rows) {
    trace("records:", rows.length);
    for each (Array<string> record in rows) {
        Array<string> parts = [];
        for each (string field in record) {
            parts.push("[" + string.replaceAll(string.replaceAll(field, "\n", "\\n"), "\r", "\\r") + "]");
        }
        trace(string.join(parts, " "));
    }
}

protected int main() {
    show(csv.parse("name,age,city\nAda,36,London\r\nAlan,41,\"Wilmslow, Cheshire\"\n"));
    show(csv.parse("\"say \"\"hi\"\"\",\"two\nlines\",,end,"));
    show(csv.parse(""));
    show(csv.parse("\n\nonly\n\n"));
    show(csv.parse(" spaced , kept \ra"));
    show(csv.parseWith("a;b,c;d\n", ';'));
    show(csv.parseWith("x\ty\n1\t2", '\t'));
    show(csv.parseWith("é→ü→\"→\"", '→'));

    Array<Array<string>> rows = [["id", "note"], ["1", "plain"], ["2", "has, comma"], ["3", "has \"quotes\""], ["4", "two\nlines"], [""]];
    string text = csv.format(rows);
    trace(text);
    Array<Array<string>> back = csv.parse(text);
    bool same = back.length == rows.length;
    for (int i = 0; i < back.length && same; i++) {
        same = back[i].length == rows[i].length;
        for (int j = 0; j < back[i].length && same; j++) {
            same = back[i][j] == rows[i][j];
        }
    }
    trace("round trip:", same);
    trace(csv.formatWith([["a;b", "c,d"]], ';'));

    attempt("bare quote", () => { trace(csv.parse("ab\"c").length); });
    attempt("text after quote", () => { trace(csv.parse("ok\n\"a\"b").length); });
    attempt("unclosed quote", () => { trace(csv.parse("a,\"never\nclosed").length); });
    attempt("column in bytes", () => { trace(csv.parse("éé\"").length); });
    attempt("quote separator", () => { trace(csv.parseWith("a", '"').length); });
    attempt("newline separator", () => { trace(csv.formatWith([["a"]], '\n')); });
    attempt("empty record", () => { trace(csv.format([["a"], []])); });
    return 0;
}
