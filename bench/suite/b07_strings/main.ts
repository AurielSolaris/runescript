// B07: concatenation, searching, splitting, joining, and replacing.
declare var process: any;

function count(s: string, sub: string): number {
    return s.split(sub).length - 1;
}

function hasPrefix(s: string, p: string): boolean {
    return s.lastIndexOf(p, 0) == 0;
}

function hasSuffix(s: string, p: string): boolean {
    return s.length >= p.length && s.indexOf(p, s.length - p.length) != -1;
}

function fields(s: string): string[] {
    var out: string[] = [];
    var parts = s.split(" ");
    for (var i = 0; i < parts.length; i++) {
        if (parts[i].length > 0) {
            out.push(parts[i]);
        }
    }
    return out;
}

var n: number = parseInt(process.argv[2], 10);
var words = ["alpha", "beta", "gamma", "delta", "rune"];

var text = "";
for (var i = 0; i < n; i++) {
    text = text + words[i % 5] + " ";
}

var runes = count(text, "rune");
var firstGamma = text.indexOf("gamma");
var lastBeta = text.lastIndexOf("beta");

var parts = fields(text);
var prefixed = 0;
for (var i = 0; i < parts.length; i++) {
    var w = parts[i];
    if (hasPrefix(w, "d") || hasSuffix(w, "a")) {
        prefixed++;
    }
}
var joined = parts.join(",");
var replaced = joined.split("beta").join("BETA");
process.stdout.write(text.length + " " + runes + " " + firstGamma + " " + lastBeta + " " + parts.length + " " +
    prefixed + " " + joined.length + " " + count(replaced, "BETA") + "\n");
