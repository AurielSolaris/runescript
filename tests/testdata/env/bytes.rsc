// std:fs binary files: readBytes, writeBytes, and appendBytes move bytes
// unchanged, including ones that are not valid UTF-8; string.bytes and
// string.fromBytes convert between text and its UTF-8 bytes.
package tests.env;

use "std:fs";
use "std:string";

public string hex(Array<uint8> bytes) {
    string digits = "0123456789abcdef";
    string out = "";
    for each (uint8 b in bytes) {
        int n = (int) b;
        out += string.slice(digits, n / 16, n / 16 + 1) + string.slice(digits, n % 16, n % 16 + 1) + " ";
    }
    return out;
}

public void attempt(string what, void() f) {
    try {
        f();
        trace(what, "ok");
    } catch (Error e) {
        trace(what, "->", e.message);
    }
}

protected int main() {
    // Every byte value survives a round trip.
    Array<uint8> all = [];
    for (int i = 0; i < 256; i++) {
        all.push((uint8) i);
    }
    fs.writeBytes("all.bin", all);
    Array<uint8> back = fs.readBytes("all.bin");
    bool same = back.length == 256;
    for (int i = 0; i < back.length; i++) {
        if ((int) back[i] != i) {
            same = false;
        }
    }
    trace("round trip", back.length, same, fs.size("all.bin"));

    // A small binary header, appended to.
    Array<uint8> header = [0x89, 0x50, 0x4E, 0x47];
    fs.writeBytes("img.bin", header);
    fs.appendBytes("img.bin", [0x0D, 0x0A, 0x1A, 0x0A]);
    fs.appendBytes("new.bin", [0xFF]);
    trace(hex(fs.readBytes("img.bin")), hex(fs.readBytes("new.bin")));
    fs.writeBytes("empty.bin", []);
    trace("empty", fs.readBytes("empty.bin").length, fs.size("empty.bin"));

    // Text and bytes.
    Array<uint8> utf8 = string.bytes("héllo €");
    trace(utf8.length, hex(utf8));
    trace(string.fromBytes(utf8));
    fs.writeBytes("text.txt", utf8);
    trace(fs.readFile("text.txt"));
    fs.writeFile("text2.txt", "añ");
    trace(hex(fs.readBytes("text2.txt")));
    // Invalid UTF-8 becomes U+FFFD, as readFile does.
    Array<uint8> bad = [0x61, 0xFF, 0x62, 0xC3];
    fs.writeBytes("bad.txt", bad);
    trace(string.fromBytes(bad), fs.readFile("bad.txt"), hex(fs.readBytes("bad.txt")));
    trace(string.bytes("").length, string.fromBytes([]) == "");

    // Failures read the same as for text files.
    fs.mkdir("dir");
    attempt("read missing", () => trace(fs.readBytes("missing.bin").length));
    attempt("read dir", () => trace(fs.readBytes("dir").length));
    attempt("write dir", () => fs.writeBytes("dir", [1]));
    attempt("append dir", () => fs.appendBytes("dir", [1]));
    attempt("write in missing dir", () => fs.writeBytes("nope/x.bin", [1]));
    return 0;
}
