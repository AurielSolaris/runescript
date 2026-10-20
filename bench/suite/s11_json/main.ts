// S11 / B08: build a JSON document of n records, parse it, read every
// record, and write it back out (the engine's native JSON).
declare var process: any;

interface Rec {
    id: number;
    name: string;
    price: number;
    tags: string[];
    ok: boolean;
}

var n: number = parseInt(process.argv[2], 10);
var records: string[] = [];
for (var i = 0; i < n; i++) {
    var price = i * 0.25;
    records.push("{\"id\":" + i + ",\"name\":\"item" + i + "\",\"price\":" + price +
        ",\"tags\":[\"a\",\"b" + (i % 3) + "\"],\"ok\":" + (i % 2 == 0) + "}");
}
var text = "[" + records.join(",") + "]";

var doc: Rec[] = JSON.parse(text);
var ids = 0;
var ok = 0;
var total = 0.0;
for (var i = 0; i < doc.length; i++) {
    var r = doc[i];
    ids = (ids + r.id) % 1000003;
    total += r.price;
    if (r.ok) {
        ok++;
    }
}
var out = JSON.stringify(doc);
process.stdout.write(doc.length + " " + ids + " " + ok + " " + total + " " + text.length + " " + out.length + " " + (out == text) + "\n");
