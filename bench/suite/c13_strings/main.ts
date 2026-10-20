// C13: string units. length/substring/indexOf count UTF-16 code units.
declare var process: any;

function codepoints(s: string): number {
    var n = 0;
    for (var i = 0; i < s.length; i++) {
        var c = s.charCodeAt(i);
        if (c >= 0xD800 && c <= 0xDBFF && i + 1 < s.length) {
            i++;
        }
        n++;
    }
    return n;
}

// substring never fails: cutting a surrogate pair returns half of it,
// which is printed here as its code unit.
function head(s: string): string {
    if (s.length == 0) {
        return "";
    }
    var h = s.substring(0, 1);
    var c = h.charCodeAt(0);
    return c >= 0xD800 && c <= 0xDFFF ? "<lone surrogate " + c.toString(16) + ">" : h;
}

var samples = ["", "hello", "café", "é", "😀", "こんにちは", "e\u0301"];
for (var i = 0; i < samples.length; i++) {
    var s = samples[i];
    var n = s.length;
    var first = head(s);
    process.stdout.write("[" + s + "] native=" + n + " codepoints=" + codepoints(s) +
        " head=[" + first + "] e@" + s.indexOf("e") + "\n");
}
process.stdout.write("precomposed == combining: " + ("é" == "e\u0301") + "\n");
