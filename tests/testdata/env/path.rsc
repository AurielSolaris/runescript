// std:path: POSIX and Windows rules explicitly, and the platform's rules
// checked against the working directory.
package tests.env;

use "std:path";
use "std:process";

protected int main() {
    trace(path.posixJoin("a/b", "../c"), path.win32Join("a\\b", "..\\c"), path.win32Join("a/b", "/c"));
    trace(path.posixNormalize("/x/./y/../z/"), path.win32Normalize("C:/x/../y"), path.posixNormalize(""));
    trace(path.posixDirname("/a/b/c.txt"), path.posixBasename("/a/b/c.txt"), path.posixBasename("/a/b/c.txt", ".txt"));
    trace(path.posixExtname("archive.tar.gz"), "[" + path.posixExtname(".bashrc") + "]", path.posixExtname("a."));
    trace(path.win32Dirname("C:\\a\\b.txt"), path.win32Basename("C:\\a\\b.txt", ".txt"), path.win32Root("\\\\srv\\share\\x"));
    trace("[" + path.posixRoot("x") + "]", path.posixRoot("/x"), path.win32Root("C:x"), path.win32Root("C:\\x"));
    trace(path.posixIsAbsolute("/x"), path.win32IsAbsolute("C:x"), path.win32IsAbsolute("C:\\x"), path.win32IsAbsolute("/x"));
    trace(path.posixRelative("/data/a/b", "/data/c"), path.win32Relative("C:\\data\\a", "c:\\DATA\\b"), path.win32Relative("C:\\x", "D:\\y"));
    trace(path.posixJoinAll(["a", "", "b", "../../.."]), path.posixResolveAll(["/a", "b", "../c"]), path.win32ResolveAll(["C:\\a", "D:\\b", "c"]));
    trace(path.posixSep(), path.win32Sep(), path.posixDelimiter(), path.win32Delimiter());

    // The platform's rules agree with the working directory.
    string here = process.cwd();
    trace(path.isAbsolute(here), path.resolve("x") == path.join(here, "x"));
    trace(path.relative(here, path.resolve("a/b")) == path.join("a", "b"), path.dirname(path.join(here, "f")) == here);
    trace(path.sep() == path.posixSep() || path.sep() == path.win32Sep());
    trace(path.basename(path.join(here, "file.txt"), ".txt"), path.extname(path.join("dir", "file.txt")));
    return 0;
}
