# 11. Compiling to JavaScript

`runescript build` compiles a program to a JavaScript module that runs in
a browser or in Node.js. The program keeps RuneScript's rules: anything
the type checker rejects is still rejected, and anything that would be a
runtime error in the interpreter is the same runtime error in JavaScript,
with the same code, message, and source position.

```sh
runescript build examples/fizzbuzz.rsc            # writes examples/fizzbuzz.js
runescript build -o web/app.js examples/hello.rsc # choose the output file
node examples/fizzbuzz.js                         # runs it; main's result is the exit code
```

`--target js` is the default and currently the only target. A program of
several files (see [Packages and files](12-packages-and-files.md)) becomes
one module, and runtime errors name the file they happened in.

## Using the module in a web page

The output is one self-contained ES2022 module with no dependencies. It
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

`options.trace` receives each line of `trace` output. Without it, lines go
to `console.log`. A `RuneScriptError` also has `code`, `file`, `line`,
`col`, and `detail` (the message without the location).

## How strict typing survives the translation

JavaScript has one number type and no static types, so RuneScript checks
everything it can before generating code, and has the generated code
guard the rest:

| RuneScript rule | Plain JavaScript | The generated code |
|---|---|---|
| Every value has its declared type | Variables hold anything | Checked at compile time, then re-verified on the typed IR before any code is written. No unchecked conversion is ever emitted. |
| `int` and `uint` are 32-bit; overflow is `ASX-RUN-002` | `2147483647 + 1` is `2147483648` | Every integer `+ - * / %` goes through a range check that stops the program |
| `/` truncates | `7 / 2` is `3.5` | Truncating division |
| Division by zero is `ASX-RUN-003` | `Infinity` or `NaN` | Checked |
| Calls nest at most 10,000 deep (`ASX-RUN-001`) | Host-dependent stack | Counted. If the host's stack runs out first, that's also `ASX-RUN-001` |
| `std:string` positions are UTF-8 byte offsets | Strings are UTF-16 | Offsets are converted, so `string.index("5€x", "x")` is `4` on every target |
| `std:string` case mapping and white space follow Go | `"ß".toUpperCase()` is `"SS"` | Tables generated from Go's `unicode` package, so the results are identical |

For example, the same program stops at the same place on both targets:

```actionscript
package demo;

public int main() {
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

## Limits of the JavaScript target

- **Stack depth.** Browsers and Node.js often run out of stack before
  10,000 nested calls. When that happens the program stops with
  `ASX-RUN-001` ("the JavaScript host ran out of stack") at a call site in
  the recursion. That site isn't always the one the interpreter would
  report.
- **Very long strings.** JavaScript engines cap string length (around
  2<sup>29</sup>–2<sup>30</sup> UTF-16 units). Exceeding it inside
  `string.repeat` or `string.replace` is reported as `ASX-RUN-004`.
- **String literals must be valid UTF-8.** A literal containing invalid
  UTF-8 bytes can't become a JavaScript string, so it's rejected with
  `ASX-TARGET-001`.
- **No JavaScript interop yet.** A program can't call JavaScript functions
  or the DOM, and JavaScript can call only `run`. Calling into JavaScript
  is planned as `use "js:<module>";` with typed declarations whose values
  are checked when they enter RuneScript.

Next: [Packages and files](12-packages-and-files.md)
