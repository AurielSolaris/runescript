// std:fs: files and directories, and the errors of failed operations,
// which read the same on both targets.
package tests.env;

use "std:fs";
use "std:path";
use "std:string";

// slash shows a path with '/' on every platform.
public string slash(string p) {
    return string.replaceAll(p, path.sep(), "/");
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
    trace(fs.exists("data"), fs.isDir("data"));
    fs.mkdir("data");
    trace(fs.exists("data"), fs.isDir("data"), fs.isFile("data"));

    fs.writeFile("data/a.txt", "héllo\n");
    fs.appendFile("data/a.txt", "wörld\n");
    trace(fs.readFile("data/a.txt"));
    trace(fs.size("data/a.txt"), fs.isFile("data/a.txt"));

    fs.mkdirAll("data/sub/deeper");
    fs.mkdirAll("data/sub/deeper");        // already there: fine
    fs.writeFile("data/sub/b.txt", "b");
    fs.writeFile("data/sub/b.txt", "replaced");
    fs.copyFile("data/a.txt", "data/sub/deeper/c.txt");
    fs.rename("data/sub/b.txt", "data/sub/renamed.txt");
    trace(fs.readFile("data/sub/renamed.txt"), fs.readFile("data/sub/deeper/c.txt") == fs.readFile("data/a.txt"));
    trace(string.join(fs.readDir("data"), ","));
    Array<string> all = fs.walk("data");
    for (int i = 0; i < all.length; i++) {
        trace(slash(all[i]));
    }

    attempt("read a missing file", () => { trace(fs.readFile("data/missing.txt")); });
    attempt("read a directory", () => { trace(fs.readFile("data")); });
    attempt("make an existing directory", () => fs.mkdir("data"));
    attempt("make a directory over a file", () => fs.mkdirAll("data/a.txt"));
    attempt("remove a non-empty directory", () => fs.remove("data/sub"));
    attempt("remove a missing file", () => fs.remove("data/nope"));
    attempt("list a file", () => { trace(fs.readDir("data/a.txt").length); });
    attempt("list a missing directory", () => { trace(fs.walk("nowhere").length); });
    attempt("measure a directory", () => { trace(fs.size("data")); });
    attempt("write over a directory", () => fs.writeFile("data/sub", "x"));
    attempt("rename a missing file", () => fs.rename("nope", "nope2"));
    attempt("copy a missing file", () => fs.copyFile("nope", "copy"));
    attempt("copy onto a directory", () => fs.copyFile("data/a.txt", "data/sub"));

    fs.remove("data/sub/renamed.txt");
    fs.removeAll("data/sub");
    fs.removeAll("data/sub");               // already gone: fine
    trace(string.join(fs.readDir("data"), ","));
    fs.remove("data/a.txt");
    fs.remove("data");
    trace(fs.exists("data"), string.len(fs.tempDir()) > 0);
    return 0;
}
