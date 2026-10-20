# 20. Files, input, and output

Four standard packages connect a program to the world around it:

| Package | For |
|---|---|
| `std:io` | Printing without a newline, error and warning lines, and reading standard input |
| `std:process` | Command-line arguments, environment variables, the working directory, and the platform |
| `std:fs` | Reading and writing files, and making, listing, and removing directories |
| `std:path` | Joining, splitting, and resolving paths, with the rules of Node.js's `path` module |

They work the same in the interpreter and in JavaScript
([chapter 11](11-javascript.md#using-the-module-in-a-web-page)), down to
their error messages.

## Output

`trace` prints a line, as before. `std:io` adds three more ways to
write:

| Function | Writes |
|---|---|
| `io.print(string s)` | `s` to standard output, without a newline |
| `io.error(string s)` | `s` and a newline to standard error, in red on a terminal |
| `io.warn(string s)` | `s` and a newline to standard error, in yellow on a terminal |

```actionscript
package main;

use "std:io";

protected int main() {
    io.print("Loading");
    for (int i = 0; i < 3; i++) {
        io.print(".");
    }
    trace(" done");                     // Loading... done
    io.warn("the cache is empty");      // on standard error
    io.error("could not reach the server");
    return 1;
}
```

`trace` and `io.print` share standard output, so their text comes out in
the order it was written. Colors are used only when standard error is a
terminal and the `NO_COLOR` environment variable isn't set; `runescript
run --color=never` turns them off as well.

## Input

| Function | Returns |
|---|---|
| `io.readLine()` | The next line of standard input, without its `"\n"` or `"\r\n"`. At the end of the input it returns what is left, which is `""` once everything has been read |
| `io.readByte()` | The next byte (0 to 255), or `-1` at the end of the input |
| `io.readAll()` | All the input that is left |
| `io.atEnd()` | Whether the input is used up. It waits for input if there is none yet |

Text is read as UTF-8; bytes that aren't valid UTF-8 become U+FFFD.
`io.readByte` gives the raw bytes.

```actionscript
package main;

use "std:io";
use "std:string";

// Numbers each line of its input:  runescript run number.rsc < notes.txt
protected int main() {
    int n = 0;
    while (!io.atEnd()) {
        n++;
        string line = io.readLine();
        trace(string.repeat(" ", 4 - string.len((string) n)) + (string) n, line);
    }
    return 0;
}
```

## Arguments and the environment

| Function | Returns |
|---|---|
| `process.args()` | The command-line arguments, as a new `Array<string>` |
| `process.env(string name)` | An environment variable, or `""` if it isn't set |
| `process.hasEnv(string name)` | Whether an environment variable is set |
| `process.cwd()` | The working directory |
| `process.platform()` | `"windows"`, `"linux"`, `"macos"`, or another system's name; `"browser"` in a web page |

The arguments are the words after the program's file:
`runescript run greet.rsc Ada Grace` (or `node greet.js Ada Grace`) passes
`["Ada", "Grace"]`. Flags for `runescript` itself go before the file.

```actionscript
package main;

use "std:process";

protected int main() {
    Array<string> names = process.args();
    if (names.length == 0) {
        trace("usage: greet <name>...");
        return 2;
    }
    for (int i = 0; i < names.length; i++) {
        trace($"Hello, {names[i]}!");
    }
    return 0;
}
```

## Files and directories

Paths are relative to the working directory. `readFile`, `writeFile`, and
`appendFile` work with UTF-8 text; for other data, see [binary
files](#binary-files) below.

| Function | Does |
|---|---|
| `fs.readFile(string path)` | Returns the file's text |
| `fs.writeFile(string path, string text)` | Creates or replaces a file |
| `fs.appendFile(string path, string text)` | Adds to the end of a file, creating it if needed |
| `fs.exists(string path)`, `fs.isFile(…)`, `fs.isDir(…)` | Whether something, a file, or a directory is there |
| `fs.size(string path)` | A file's size in bytes, as an `int64` |
| `fs.copyFile(string from, string to)` | Copies a file, replacing `to` |
| `fs.rename(string from, string to)` | Moves or renames a file or directory |
| `fs.remove(string path)` | Removes a file or an empty directory |
| `fs.removeAll(string path)` | Removes a directory and everything in it; nothing if it isn't there |
| `fs.mkdir(string path)` | Makes a directory, whose parent must exist |
| `fs.mkdirAll(string path)` | Makes a directory and any missing parents; nothing if it's already there |
| `fs.readDir(string path)` | The names in a directory, sorted |
| `fs.walk(string path)` | Every file and directory below `path`, as paths relative to it, each directory followed by its contents |
| `fs.tempDir()` | The system's directory for temporary files |

A failure throws an `Error` whose message says what went wrong, in the same
words everywhere: `fs.readFile: cannot read notes.txt: no such file or
directory`. The reasons are "no such file or directory", "already exists",
"is a directory", "not a directory", "directory not empty", and "permission
denied".

```actionscript
package main;

use "std:fs";
use "std:path";

protected int main() {
    string dir = path.join(fs.tempDir(), "runescript-guide");
    fs.removeAll(dir);
    fs.mkdirAll(path.join(dir, "notes"));
    fs.writeFile(path.join(dir, "notes/todo.txt"), "buy milk\n");
    fs.appendFile(path.join(dir, "notes/todo.txt"), "write docs\n");
    trace(fs.readFile(path.join(dir, "notes/todo.txt")));
    trace(fs.size(path.join(dir, "notes/todo.txt")));       // 20
    Array<string> all = fs.walk(dir);
    trace(all.length);                                     // 2: notes, and the file in it
    try {
        trace(fs.readFile(path.join(dir, "missing.txt")));
    } catch (Error e) {
        trace(e.name);                                     // Error
    }
    fs.removeAll(dir);
    trace(fs.exists(dir));                                 // false
    return 0;
}
```

### Binary files

Images, archives, and other data that isn't text are read and written as
bytes, an `Array<uint8>`. The bytes go to and from the file unchanged.

| Function | Does |
|---|---|
| `fs.readBytes(string path)` | Returns the file's bytes, an `Array<uint8>` |
| `fs.writeBytes(string path, Array<uint8> bytes)` | Creates or replaces a file |
| `fs.appendBytes(string path, Array<uint8> bytes)` | Adds to the end of a file, creating it if needed |

They fail like their text versions, with the same messages.
`string.bytes(s)` and `string.fromBytes(b)` ([chapter
10](10-std-string.md#bytes)) convert between text and its UTF-8 bytes.

```actionscript
package main;

use "std:fs";
use "std:path";
use "std:string";

protected int main() {
    string file = path.join(fs.tempDir(), "runescript-guide.bin");
    Array<uint8> header = [0x89, 0x50, 0x4E, 0x47];   // the start of a PNG file
    fs.writeBytes(file, header);
    fs.appendBytes(file, string.bytes("hi"));
    Array<uint8> data = fs.readBytes(file);
    trace(data.length, data[0], data[1]);             // 6 137 80
    trace(fs.size(file));                             // 6
    bool png = data.length >= 4 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47;
    trace(png);                                       // true
    trace(string.fromBytes(data.slice(4, 6)));        // hi
    fs.remove(file);
    return 0;
}
```

## Paths

`std:path` has the functions of Node.js's `path` module, written for
RuneScript (they don't call Node.js, so they work in the interpreter too):

| Function | Returns |
|---|---|
| `path.join(a, b)`, `path.joinAll(parts)` | The parts joined with the separator, then normalized |
| `path.resolve(p)`, `path.resolveAll(parts)` | An absolute path, starting from the working directory |
| `path.normalize(p)` | `p` without `.`, `..`, or repeated separators |
| `path.relative(from, to)` | The path from `from` to `to` |
| `path.dirname(p)`, `path.basename(p)`, `path.basename(p, suffix)`, `path.extname(p)` | The directory, the last part (optionally without `suffix`), and the extension |
| `path.isAbsolute(p)`, `path.root(p)` | Whether `p` is absolute, and its root (`"/"`, `"C:\"`, …) |
| `path.sep()`, `path.delimiter()` | `"/"` and `":"`, or `"\"` and `";"` on Windows |

These follow the rules of the system the program runs on: Windows paths
on Windows (`C:\`, both `\` and `/` as separators, `\\server\share`), POSIX
paths elsewhere. For paths that must look the same everywhere, each
function also comes as `path.posixJoin`, `path.posixResolve`, … and
`path.win32Join`, `path.win32Resolve`, …:

```actionscript
package main;

use "std:path";

protected int main() {
    trace(path.posixJoin("/home/ada", "../grace/notes.txt"));   // /home/grace/notes.txt
    trace(path.win32Join("C:\\Users", "ada\\..\\grace"));        // C:\Users\grace
    trace(path.posixBasename("/srv/www/index.html", ".html"));  // index
    trace(path.posixExtname("archive.tar.gz"));                 // .gz
    trace(path.posixRelative("/data/a/b", "/data/c"));          // ../../c
    trace(path.win32IsAbsolute("C:docs"));                      // false: relative to C:'s current directory
    return 0;
}
```

The results match Node.js's `path` module (the tests compare thousands of
random paths), except that Windows paths are compared ignoring ASCII case
only, and that resolving a path on another drive starts from that drive's
root.

## Example: a Brainfuck interpreter

[`examples/brainfuck.rsc`](../../examples/brainfuck.rsc) runs
[Brainfuck](https://en.wikipedia.org/wiki/Brainfuck) programs on a tape of
32,768 cells, using all four packages:

```text
$ runescript run examples/brainfuck.rsc examples/bf/hello.bf
Hello World!
$ echo "Hello, World!" | runescript run examples/brainfuck.rsc examples/bf/rot13.bf
Uryyb, Jbeyq!
$ runescript run examples/brainfuck.rsc -e "+[>+]"
-e:1:3: the data pointer moved off the tape (to cell 32768; the cells are 0 to 32767)
```

It reads the program with `fs.readFile`, or takes it from `-e`, using
`process.args()`. It compiles it first into a list of structs, folding
runs of `+` and `-` into one step and matching every bracket, so an
unmatched one is reported with its line and column before anything runs.
The tape is an `Array<uint8>` whose cells wrap from 255 to 0 through
`% 256`. `,` reads with `io.readByte()` and leaves the cell unchanged at
the end of the input. `.` collects bytes and decodes them as UTF-8 in
RuneScript before printing with `io.print`, so a Brainfuck program can
print any character. Errors go to `io.error`, and `main` returns the exit
status.

Next: back to the [guide index](README.md)
