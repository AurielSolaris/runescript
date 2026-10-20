// S11 / B08: build a JSON document of n records, parse it, read every
// record, and write it back out.
package bench;

use "std:json";
use "std:process";
use "std:string";

protected int main() {
    Array<string> args = process.args();
    int n = (int) args[0];
    Array<string> records = [];
    for (int i = 0; i < n; i++) {
        float64 price = (float64) i * 0.25;
        records.push($"{{\"id\":{i},\"name\":\"item{i}\",\"price\":{price},\"tags\":[\"a\",\"b{i % 3}\"],\"ok\":{i % 2 == 0}}}");
    }
    string text = "[" + string.join(records, ",") + "]";

    json.Value doc = json.parse(text);
    int ids = 0;
    int ok = 0;
    float64 total = 0.0;
    for each (json.Value r in json.items(doc)) {
        ids = (ids + json.asInt(json.get(r, "id"))) % 1000003;
        total += json.asFloat(json.get(r, "price"));
        if (json.asBool(json.get(r, "ok"))) {
            ok++;
        }
    }
    string out = json.stringify(doc);
    trace(json.length(doc), ids, ok, total, string.len(text), string.len(out), out == text);
    return 0;
}
