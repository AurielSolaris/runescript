# 11. Compiling to JavaScript

`runescript build` compiles a program to a JavaScript module that runs in
a browser or in Node.js. The program keeps RuneScript's rules: anything
the type checker rejects is still rejected, and anything that would be a
runtime error in the interpreter is the same runtime error in JavaScript,
with the same code, message, and source position. To build web pages from
components and templates, see [chapter 15](15-web-pages.md).

```sh
runescript build examples/fizzbuzz.rsc            # writes examples/fizzbuzz.js
runescript build -o web/app.js examples/hello.rsc # choose the output file
node examples/fizzbuzz.js a b                     # runs it with arguments; main's result is the exit code
```

`--target js` is the default and currently the only target. A program of
several files (see [Packages and files](12-packages-and-files.md)) becomes
one module, and runtime errors name the file they happened in.

**What is promised, and what isn't.** RuneScript does **not** yet promise
full JavaScript compatibility. The module needs an engine with ES2015,
`BigInt`, `??`, and ES modules. It exchanges only numbers, `bool`,
`string`, and errors with JavaScript, through
[extern functions](#calling-javascript-extern-functions). The standard
packages `std:io`, `std:process`, `std:fs`, and `std:path`
([chapter 20](20-files-and-input.md)) work on both targets; in JavaScript
they use what `run` is given, which Node.js provides when it runs the
module. What it **does** promise is strict code:
nothing gets into a RuneScript program without being checked. TypeScript
trusts its declarations and erases them, so a `.d.ts` that says `number`
can hand you a string at run time. RuneScript checks every value that comes
back from JavaScript against its declared type, and stops with a
`TypeError` when it doesn't match. There is no `any`, no implicit
conversion, and no unchecked cast.

## Using the module in a web page

The output is one self-contained ES2015 module (plus BigInt and `??`) with no dependencies. It
exports `run`, which calls `main` and returns its exit code, and
`RuneScriptError`, which `run` throws when the program stops with a
runtime error.

```html
<pre id="out"></pre>
<script type="module">
  import { run, RuneScriptError } from "./fizzbuzz.js";
  const out = document.getElementById("out");
  try {
    run({ trace: (line) => { out.textContent += line + "\n"; } });
  } catch (e) {
    if (!(e instanceof RuneScriptError)) throw e;
    out.textContent += e.message; // e.g. "fizzbuzz.rsc:6:15: error ASX-RUN-002: …"
  }
</script>
```

`options.trace` receives each line of output (`trace` and `io.print`).
Without it, lines go to `console.log`. A `RuneScriptError` also has `code`,
`file`, `line`, `col`, and `detail` (the message without the location).

The other options are the program's environment, all optional:

| Option | Used by | Default |
|---|---|---|
| `stdout(text)`, `stderr(text)` | `trace`, `io.print`; `io.error`, `io.warn` | whole lines to `options.trace` (or `console.log`) and `console.error` |
| `stdin()` | `io.readLine`, `io.readByte`, … | no input. Return the next chunk as a `Uint8Array`, or `null` at the end |
| `args`, `env` | `process.args()`, `process.env(name)` | `[]`, `{}` |
| `cwd()`, `platform` | `process.cwd()`, `std:path` | `"/"`, `"browser"` (Node.js's `process.platform` names) |
| `color` | `io.error`, `io.warn` | `false` |
| `fs`, `tempDir()` | `std:fs` | none: `std:fs` throws an `Error` |

When Node.js runs the module directly (`node app.js a b`), the module
passes Node's own standard streams, `process.argv`, `process.env`, working
directory, platform, and `node:fs`.

A program with extern functions imports their modules, so it runs only
where those modules exist: a module using `rune:fs` runs in Node.js, not in
a browser.

## Source maps

With `--source-map`, `build` also writes a source map beside the module
and ends the module with a `//# sourceMappingURL=` comment that points at
it. Debuggers that read source maps (browser developer tools, Node.js's
inspector, VS Code) then show the RuneScript source: each statement and
function of the module maps to the line it came from, and breakpoints can
be set in the `.rsc` file. The map names the source files relative to
itself and also holds their text, so it works even where the sources
aren't served. `runescript dev` always includes a map, inside the module.

```sh
runescript build --source-map -o dist/app.js app.rsc   # also writes dist/app.js.map
```

## How strict typing survives the translation

JavaScript has one number type and no static types, so RuneScript checks
everything it can before generating code, and has the generated code
guard the rest:

| RuneScript rule | Plain JavaScript | The generated code |
|---|---|---|
| Every value has its declared type | Variables hold anything | Checked at compile time, then re-verified on the typed IR before any code is written. No unchecked conversion is ever emitted. |
| `int` and `uint` are 32-bit; intermediate results widen exactly; overflow is `ASX-RUN-002` | `2147483647 + 1` is `2147483648`, and large products are rounded | Every integer `+ - * / %` goes through a range check that stops the program; wide intermediates switch to `BigInt` beyond 2^53 |
| `/` truncates | `7 / 2` is `3.5` | Truncating division |
| Division by zero is `ASX-RUN-003` | `Infinity` or `NaN` | Checked, for integers and floats |
| `float32` rounds to 32 bits | One 64-bit number type | `float32` results go through `Math.fround`, and print with float32's shortest digits; `float` and `float64` are JavaScript numbers (IEEE, with `NaN` and `Infinity`, as in AS3) |
| `(int) "12x"` throws a `TypeError` | `parseInt("12x")` is `12` | The same parsing rules as the interpreter |
| Runtime errors are catchable error values | Exceptions are any value | Errors are RuneScript error objects; catch clauses test their type; the call-depth error passes every catch clause |
| Calls nest at most 10,000 deep (`ASX-RUN-001`) | Host-dependent stack | Counted. If the host's stack runs out first, that's also `ASX-RUN-001` |
| `std:string` positions are UTF-8 byte offsets | Strings are UTF-16 | Offsets are converted, so `string.index("5€x", "x")` is `4` on every target |
| `std:string` case mapping and white space follow Go | `"ß".toUpperCase()` is `"SS"` | Tables generated from Go's `unicode` package, so the results are identical |

For example, the same program stops at the same place on both targets:

```actionscript
package demo;

protected int main() {
    int big = 2147483647;
    trace(big * big); // runtime error: ASX-RUN-002
    return 0;
}
```

The message is `integer overflow: 2147483647 * 2147483647 =
4611686014132420609 does not fit in 'int'`, computed exactly even though a
JavaScript number can't hold that product.

The test suite runs every example and test program on both the
interpreter and Node.js and requires identical output, exit codes, and
errors. It also compares every `std:string` function against Go's on
random non-ASCII text.

## Calling JavaScript: extern functions

An `extern` declaration names a function that a JavaScript module
implements:

```actionscript
package notes;

extern "rune:fs" void writeFile(string path, string data);
extern "rune:fs" string readFile(string path);
extern "./clock.mjs" float64 now();

protected int main() {
    writeFile("notes.txt", "ᚱᚢᚾᛖ");
    trace(readFile("notes.txt"), now() > 0);
    return 0;
}
```

```js
// clock.mjs, next to the compiled module
export const now = () => Date.now();
```

- An extern declaration goes at the top level of a file. It has no access
  modifier and is visible only in that file. Parameter names are
  optional.
- An extern function can only be called. To pass it around, wrap it in a
  lambda: `(string p) => readFile(p)`.
- The module string is either a **host-neutral module** (`rune:…`, below),
  or anything JavaScript's `import` accepts: a file such as
  `"./clock.mjs"` (relative to the compiled module), a package name, or a
  host's own module such as `"node:fs"`. A host's own module works, but it
  ties the program to that host, so it's reported as warning
  `ASX-WARN-006`.
- Only the JavaScript target runs extern functions. `runescript check`
  accepts the program, but `runescript run` reports `ASX-TARGET-002`. A
  JavaScript runtime bundled with RuneScript is planned, so that such
  programs will run without Node.js.

### Host-neutral modules

A `rune:` module is part of RuneScript, not of one host. Its functions
are fixed, and a declaration must match one exactly (`ASX-TYPE-024`
otherwise). The compiler maps each function to the host's own API; today
that host is Node.js.

| `rune:fs` function | Does |
|---|---|
| `string readFile(string path)` | Returns the contents of a file, read as UTF-8 (invalid bytes become U+FFFD) |
| `void writeFile(string path, string data)` | Creates or replaces a file |
| `void appendFile(string path, string data)` | Adds to the end of a file, creating it if needed |
| `bool exists(string path)` | Whether a file or directory exists |
| `void remove(string path)` | Deletes a file |

A failing function, e.g. `readFile` on a missing file, throws an `Error`
(`ASX-RUN-008`).

`rune:fs` came first and runs only on the JavaScript target. For new
programs, prefer the standard package `std:fs` ([chapter 20](20-files-and-input.md)):
it does more (directories, sizes, copying, walking a tree), runs in the
interpreter too, and fails with the same messages on both targets.

### What crosses the boundary

Arguments go to JavaScript as they are: numbers, booleans, strings, and
errors as `RSError` objects (below). Results are **checked** on the way
back:

| Declared result | JavaScript must return |
|---|---|
| `int`, `uint` | An integral number in the type's range (`-0` becomes `0`) |
| `float`, `float64` | Any number, including `NaN` and `Infinity` |
| `float32` | Any number. It is rounded to float32 with `Math.fround` |
| `bool` | `true` or `false` |
| `string` | A string without unpaired surrogates, so that it is valid UTF-8 |
| `void` | `undefined` |
| `Error`, `RangeError`, `TypeError`, `ArgumentError` | An `RSError` of that type (or a subtype), or a JavaScript `Error`, which is converted: `RangeError` and `TypeError` keep their names, any other becomes `Error`, and the `errorID` is `0` |

Anything else throws a `TypeError` with code `ASX-RUN-007` at the call:

```actionscript
package demo;

extern "./lib.mjs" int half(int n);

protected int main() {
    try {
        trace(half(3)); // JavaScript returns 1.5
    } catch (TypeError e) {
        trace(e.message); // half returned 1.5, but its result type is 'int'
    }
    return 0;
}
```

An exception thrown by JavaScript becomes an `Error` with code
`ASX-RUN-008` and a message such as `readFile threw Error: ENOENT: …`.
Nothing catches it unless your program does.

### Errors in both directions

RuneScript errors and JavaScript share one class. The module exports
`RSError`, the class of RuneScript error values. JavaScript receives errors
as `RSError` objects with `name`, `message`, and `errorID`, and it can
create, return, and throw them. An `RSError` that JavaScript throws reaches
RuneScript unchanged, so it's caught by its own type. With
`extern "./even.mjs" int halveEven(int n);`, `halveEven(3)` throws an
`ArgumentError` that `catch (ArgumentError e)` catches:

```js
// even.mjs: the compiled program is prog.js
import { RSError } from "./prog.js";

export const halveEven = (n) => {
  if (n % 2 !== 0) throw new RSError("ArgumentError", "n must be even", 1);
  return n / 2;
};
```

## Live reload: `runescript dev`

`runescript dev` serves a program on your machine while you work on it.
Every time you save one of its files, it compiles the program again and
reloads the browser. When the program doesn't compile, the page keeps
running the last good build and shows the errors in an overlay until you
fix them. Warnings go to the browser console and the terminal.

```sh
runescript dev app.rsc             # http://127.0.0.1:5173/, reloads on every save
runescript dev --open --port 8080 app.rsc
runescript run --watch tool.rsc    # for command-line programs: runs again on every save
```

- It serves the folder of the entry file. An `index.html` there is used as
  the page, with the dev client added. It should load the module as
  `./app.js`, the same name `runescript build` writes, so the page works
  unchanged in production. Without an `index.html`, a page with
  `<div id="app">` is generated that runs the program and shows its `trace`
  output.
- The compiled module is kept in memory, so nothing is written to disk.
  Other files in the folder are served as they are, such as helper modules
  for extern functions, style sheets, and images.
- It listens only on `127.0.0.1`. The default port is 5173, and the next
  free port is used if it's taken. An explicit `--port` must be free.
- `-Werror` and `--allow` work as for `build`.

This is live reload: the page starts over after each change. Hot module
replacement, which swaps changed components and keeps their runes'
values, is planned for `rune:dom` components.

For programs that don't use a browser, `runescript run --watch` runs the
program again whenever one of its files changes. A run that is still going
when you save finishes first.

## Limits of the JavaScript target

- **Stack depth.** Browsers and Node.js often run out of stack before
  10,000 nested calls. When that happens the program stops with
  `ASX-RUN-001` ("the JavaScript host ran out of stack") at a call site in
  the recursion. That site isn't always the one the interpreter would
  report.
- **Very long strings.** JavaScript engines cap string length (around
  2<sup>29</sup>–2<sup>30</sup> UTF-16 units). Exceeding it inside
  `string.repeat` or `string.replace` is reported as `ASX-RUN-004`.
- **Limited JavaScript interop.** Extern functions exchange only numbers,
  `bool`, `string`, and errors. Objects, arrays, callbacks, and the DOM
  can't cross yet, and JavaScript can call into RuneScript only through
  `run`.
- **Not full JavaScript compatibility.** The output targets ES2015 plus
  `BigInt`, `??`, and ES modules, and it is tested on Node.js. Other hosts
  are expected to work but aren't tested.

Next: [Packages and files](12-packages-and-files.md)
