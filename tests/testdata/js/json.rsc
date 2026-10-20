// std:json: parsing, reading with checked functions, building, writing,
// and the errors each function throws, the same on both targets.
package tests.js;

use "std:json";
use "std:string";

public struct Item {
    string name;
    int count;
    float64 price;
}

public void attempt(string what, void() f) {
    try {
        f();
        trace(what, "ok");
    } catch (Error e) {
        trace(what, "->", e.name, e.message);
    }
}

// attemptOn is attempt with a value to work on (lambdas do not capture).
public void attemptOn(string what, json.Value v, void(json.Value) f) {
    try {
        f(v);
        trace(what, "ok");
    } catch (Error e) {
        trace(what, "->", e.name, e.message);
    }
}

// describe shows a value's shape, depth first.
public string describe(json.Value v) {
    string kind = json.kind(v);
    if (kind == "array") {
        Array<string> parts = [];
        for each (json.Value e in json.items(v)) {
            parts.push(describe(e));
        }
        return "[" + string.join(parts, " ") + "]";
    }
    if (kind == "object") {
        Array<string> parts = [];
        for each (string k in json.keys(v)) {
            parts.push(k + ":" + describe(json.get(v, k)));
        }
        return "{" + string.join(parts, " ") + "}";
    }
    return kind;
}

protected int main() {
    string text = "{\n  \"store\": \"Rune & Co\",\n  \"open\": true,\n  \"items\": [\n    {\"name\": \"lamp\", \"count\": 2, \"price\": 19.5},\n    {\"name\": \"rug\", \"count\": 1, \"price\": 120}\n  ],\n  \"note\": null,\n  \"id\": 12345678901234567890\n}";
    json.Value doc = json.parse(text);
    trace(describe(doc));
    trace(json.asString(json.get(doc, "store")), json.asBool(json.get(doc, "open")), json.isNull(json.get(doc, "note")));
    trace(json.has(doc, "items"), json.has(doc, "missing"), json.length(doc));

    // Into structs, field by field.
    Array<Item> items = [];
    for each (json.Value v in json.items(json.get(doc, "items"))) {
        items.push(Item{name: json.asString(json.get(v, "name")), count: json.asInt(json.get(v, "count")), price: json.asFloat(json.get(v, "price"))});
    }
    float64 total = 0.0;
    for each (Item it in items) {
        total += (float64) it.count * it.price;
    }
    trace(items.length, items[1].name, total);

    // Numbers keep their text: the id is written back exactly.
    json.Value id = json.get(doc, "id");
    trace(json.stringify(id), json.asFloat(id));
    attemptOn("id as int64", id, (json.Value d) => trace(json.asInt64(d)));
    attemptOn("price as int", doc, (json.Value d) => trace(json.asInt(json.get(json.at(json.get(d, "items"), 0), "price"))));
    trace(json.asInt64(json.parse("-9223372036854775808")), json.asInt(json.parse("-0")), json.asFloat(json.parse("1E-7")));

    // Building values and writing them.
    Array<json.Value> outItems = [];
    for each (Item it in items) {
        outItems.push(json.fromObject(["name", "total"], [json.fromString(string.toUpper(it.name)), json.fromFloat((float64) it.count * it.price)]));
    }
    json.Value report = json.fromObject(
        ["store", "items", "count", "big", "ok", "none", "text"],
        [json.get(doc, "store"), json.fromArray(outItems), json.fromInt(items.length), json.fromInt64((int64) 9007199254740993),
         json.fromBool(true), json.nullValue(), json.fromString("tab\tquote\" é 😀 \u{1}")]);
    trace(json.stringify(report));
    trace(json.stringifyIndent(report, "  "));
    trace(json.stringifyIndent(json.parse("[[], {}, [1], {\"a\": []}]"), "\t"));

    // Keys keep document order; a repeated key keeps its first place and
    // its last value.
    json.Value dup = json.parse("{\"b\": 1, \"a\": 2, \"b\": 3, \"10\": 4, \"2\": 5}");
    trace(json.stringify(dup), string.join(json.keys(dup), ","));
    json.Value built = json.fromObject(["x", "y", "x"], [json.fromInt(1), json.fromInt(2), json.fromInt(3)]);
    trace(json.stringify(built));

    // Escapes: a surrogate pair is one character; a lone surrogate is U+FFFD.
    json.Value s = json.parse("[\"\\u00e9\\u20AC\\ud83d\\ude00\", \"\\ud83d!\", \"\\/\\b\\f\\n\\r\\t\"]");
    trace(json.asString(json.at(s, 0)), json.asString(json.at(s, 1)), json.stringify(json.at(s, 2)));

    // Values are never changed: items() and keys() return new arrays.
    Array<json.Value> copy = json.items(json.get(doc, "items"));
    trace(json.kind(copy.pop()));
    trace(copy.length, json.length(json.get(doc, "items")));

    attempt("trailing comma", () => trace(json.kind(json.parse("[1, 2,]"))));
    attempt("bare word", () => trace(json.kind(json.parse("{\"a\": nul}"))));
    attempt("single quotes", () => trace(json.kind(json.parse("{'a': 1}"))));
    attempt("unclosed", () => trace(json.kind(json.parse("{\"a\": [1, 2"))));
    attempt("leading zero", () => trace(json.kind(json.parse("012"))));
    attempt("bad number", () => trace(json.kind(json.parse("[1.]"))));
    attempt("control char", () => trace(json.kind(json.parse("\"a\nb\""))));
    attempt("bad escape", () => trace(json.kind(json.parse("\"\\x\""))));
    attempt("after é", () => trace(json.kind(json.parse("[\"é\" x]"))));
    attempt("empty", () => trace(json.kind(json.parse(""))));
    attemptOn("asString of number", id, (json.Value d) => trace(json.asString(d)));
    attemptOn("get on array", doc, (json.Value d) => trace(json.kind(json.get(json.get(d, "items"), "x"))));
    attemptOn("missing key", doc, (json.Value d) => trace(json.kind(json.get(d, "phone"))));
    attemptOn("at out of range", doc, (json.Value d) => trace(json.kind(json.at(json.get(d, "items"), 2))));
    attemptOn("length of null", doc, (json.Value d) => trace(json.length(json.get(d, "note"))));
    attempt("huge float", () => trace(json.asFloat(json.parse("1e400"))));
    attempt("int range", () => trace(json.asInt(json.parse("2147483648"))));
    attempt("NaN", () => {
        float64 zero = 0.0;
        trace(json.kind(json.fromFloat(zero / zero)));
    });
    attempt("keys and values", () => trace(json.kind(json.fromObject(["a"], []))));
    return 0;
}
