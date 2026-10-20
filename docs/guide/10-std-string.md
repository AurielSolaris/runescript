# 10. Standard library: `std:string`

`std:string` is RuneScript's version of Go's
[`strings`](https://pkg.go.dev/strings) package. Bring it in with `use`,
then call its functions through the `string` qualifier:

```actionscript
package main;

use "std:string";

protected int main() {
    string name = string.trimSpace("   rune script   ");
    trace(string.toUpper(name));                       // RUNE SCRIPT
    trace(string.hasPrefix(name, "rune"));             // true
    trace(string.replaceAll(name, " ", "-"));          // rune-script
    return 0;
}
```

## Conventions

- **Calling.** `use "std:string";` goes after the `package` line and before
  any function. Then call `string.<name>(...)`. Without the `use`, the
  qualifier isn't in scope (`ASX-NAME-007`).
- **Names.** Go's names converted to camelCase: `ToUpper` → `toUpper`,
  `TrimSpace` → `trimSpace`, `ToValidUTF8` → `toValidUTF8`. If you write the
  Go spelling, the error message suggests the right one.
- **Positions are byte offsets.** Strings hold UTF-8 text, and positions,
  lengths, and counts are in **bytes**, exactly as in Go. In `"héllo"`, `é`
  takes two bytes, so `string.index("héllo", "l")` is `3`.
- **Results are typed.** Functions that return positions or counts return
  `int`. Assigning one to a `uint` is the usual `ASX-TYPE-004` error.
- **Behaviour matches Go.** Edge cases, such as an empty separator or a
  negative `n`, behave as Go documents them.
- **Joining strings** uses `+` or an interpolated string `$"…"`, not a
  library function ([chapter 16](16-text-and-arrays.md)).

## Function reference

### Comparing and searching

| Function | Returns | Description |
|---|---|---|
| `compare(string a, string b)` | `int` | `-1`, `0`, or `1` by byte-wise order |
| `equalFold(string s, string t)` | `bool` | Equal ignoring Unicode case |
| `contains(string s, string substr)` | `bool` | `substr` occurs in `s` |
| `containsAny(string s, string chars)` | `bool` | Any character of `chars` occurs in `s` |
| `hasPrefix(string s, string prefix)` | `bool` | `s` starts with `prefix` |
| `hasSuffix(string s, string suffix)` | `bool` | `s` ends with `suffix` |
| `count(string s, string substr)` | `int` | Number of non-overlapping occurrences. With an empty `substr`: 1 + the number of characters |
| `index(string s, string substr)` | `int` | Byte offset of the first occurrence, or `-1` |
| `indexAny(string s, string chars)` | `int` | Byte offset of the first character from `chars`, or `-1` |
| `lastIndex(string s, string substr)` | `int` | Byte offset of the last occurrence, or `-1` |
| `lastIndexAny(string s, string chars)` | `int` | Byte offset of the last character from `chars`, or `-1` |

```actionscript
package main;

use "std:string";

protected int main() {
    string s = "go gopher";
    trace(string.contains(s, "ph"), string.index(s, "go"), string.lastIndex(s, "go"));   // true 0 3
    trace(string.count("cheese", "e"));                                                  // 3
    trace(string.equalFold("RuneScript", "runescript"));                                 // true
    if (string.index(s, "x") == -1) {
        trace("no x");
    }
    return 0;
}
```

### Transforming

| Function | Returns | Description |
|---|---|---|
| `toUpper(string s)` | `string` | Upper case |
| `toLower(string s)` | `string` | Lower case |
| `toTitle(string s)` | `string` | Title case for every letter (Unicode title case; for most letters, same as upper case) |
| `repeat(string s, int count)` | `string` | `s` repeated `count` times. A negative `count` is a runtime error (`ASX-RUN-004`) |
| `replace(string s, string old, string new, int n)` | `string` | Replaces the first `n` occurrences; `n < 0` replaces all |
| `replaceAll(string s, string old, string new)` | `string` | Replaces every occurrence |
| `toValidUTF8(string s, string replacement)` | `string` | Replaces invalid UTF-8 sequences |
| `clone(string s)` | `string` | A fresh copy (provided for completeness) |

```actionscript
package main;

use "std:string";

protected int main() {
    trace(string.repeat("ab", 3));                          // ababab
    trace(string.replace("oink oink oink", "k", "ky", 2));  // oinky oinky oink
    trace(string.toLower("LOUD"), string.toUpper("quiet"));  // loud QUIET
    return 0;
}
```

### Trimming

| Function | Returns | Description |
|---|---|---|
| `trimSpace(string s)` | `string` | Removes leading and trailing whitespace |
| `trim(string s, string cutset)` | `string` | Removes leading and trailing characters found in `cutset` |
| `trimLeft(string s, string cutset)` | `string` | Leading characters only |
| `trimRight(string s, string cutset)` | `string` | Trailing characters only |
| `trimPrefix(string s, string prefix)` | `string` | Removes `prefix` once, if present |
| `trimSuffix(string s, string suffix)` | `string` | Removes `suffix` once, if present |

```actionscript
package main;

use "std:string";

protected int main() {
    trace(string.trim("!!hello!!", "!"));                 // hello
    trace(string.trimLeft("xxhixx", "x"));                // hixx
    trace(string.trimSuffix("game.rsc", ".rsc"));           // game
    return 0;
}
```

`trimPrefix`/`trimSuffix` remove a whole string once. `trimLeft`/`trimRight`
remove any run of the listed characters.

### Length and slicing

Go measures and cuts strings with its built-ins `len(s)` and `s[i:j]`.
RuneScript has functions for them, which also count in bytes:

| Function | Returns | Description |
|---|---|---|
| `len(string s)` | `int` | The length in bytes |
| `runeCount(string s)` | `int` | The number of characters (code points) |
| `slice(string s, int start, int end)` | `string` | The bytes from `start` up to, not including, `end` |

`slice` never cuts a character in two. Bounds outside `0 … len(s)`,
`start > end`, or a bound in the middle of a character throw a
`RangeError` (`ASX-RUN-004`).

```actionscript
package main;

use "std:string";

protected int main() {
    string s = "héllo";
    trace(string.len(s), string.runeCount(s));   // 6 5
    trace(string.slice(s, 0, 3));                // hé
    int at = string.index(s, "llo");
    trace(string.slice(s, at, string.len(s)));   // llo
    try {
        trace(string.slice(s, 0, 2));            // 2 is inside 'é'
    } catch (RangeError e) {
        trace(e.message);   // string.slice: byte offset 2 is inside a character
    }
    return 0;
}
```

Combine `index` and `slice` to take a string apart, as you would in Go.

### Bytes

Go converts with `[]byte(s)` and `string(b)`. RuneScript has two functions
that convert between a string and its UTF-8 bytes, an `Array<uint8>`:

| Function | Returns | Description |
|---|---|---|
| `bytes(string s)` | `Array<uint8>` | The UTF-8 encoding of `s`, `len(s)` bytes |
| `fromBytes(Array<uint8> b)` | `string` | The text in `b`; invalid UTF-8 becomes `�` (U+FFFD), as when `fs.readFile` reads a file |

```actionscript
package main;

use "std:string";

protected int main() {
    Array<uint8> b = string.bytes("hé");
    trace(b.length, b[0], b[1], b[2]);           // 3 104 195 169
    trace(string.fromBytes(b));                  // hé
    Array<uint8> broken = [0x61, 0xFF, 0x62];
    trace(string.fromBytes(broken));             // a�b
    return 0;
}
```

Use them with `fs.readBytes` and `fs.writeBytes` ([chapter
20](20-files-and-input.md#binary-files)) for files that mix text and binary
data.

### Characters and callbacks

These functions take a [`char`](16-text-and-arrays.md#characters-char),
or a function that looks at one character at a time. Each callback is called
in the same order as in Go: left to right, except that `lastIndexFunc` and
the right-hand side of the trims work from the end.

| Function | Returns | Description |
|---|---|---|
| `containsRune(string s, char c)` | `bool` | `c` occurs in `s` |
| `indexRune(string s, char c)` | `int` | Byte offset of the first `c`, or `-1` |
| `containsFunc(string s, bool(char) f)` | `bool` | Some character satisfies `f` |
| `indexFunc(string s, bool(char) f)` | `int` | Byte offset of the first character satisfying `f`, or `-1` |
| `lastIndexFunc(string s, bool(char) f)` | `int` | Byte offset of the last one, or `-1` |
| `trimFunc(string s, bool(char) f)` | `string` | Removes leading and trailing characters satisfying `f` |
| `trimLeftFunc(string s, bool(char) f)` | `string` | Leading ones only |
| `trimRightFunc(string s, bool(char) f)` | `string` | Trailing ones only |
| `map(char(char) f, string s)` | `string` | Every character replaced by `f` of it (note Go's argument order) |

```actionscript
package main;

use "std:string";

public bool isDigit(char c) {
    return c >= '0' && c <= '9';
}

public char hide(char c) {
    if (isDigit(c)) {
        return '#';
    }
    return c;
}

protected int main() {
    string order = "order 66 shipped";
    trace(string.indexFunc(order, isDigit));            // 6
    trace(string.trimFunc("2024-report-v2", isDigit));  // -report-v
    trace(string.map(hide, order));                     // order ## shipped
    trace(string.containsRune(order, 'z'));             // false
    return 0;
}
```

Unlike Go's `Map`, `map` can't drop characters, because a `char` is never
negative.

### Splitting and joining

These return or take an [`Array<string>`](16-text-and-arrays.md#arrays-arrayt):

| Function | Returns | Description |
|---|---|---|
| `split(string s, string sep)` | `Array<string>` | The pieces between the separators. An empty `sep` splits into characters |
| `splitN(string s, string sep, int n)` | `Array<string>` | At most `n` pieces, the last holding the rest; `n < 0` means all, `n == 0` none |
| `splitAfter(string s, string sep)` | `Array<string>` | Like `split`, but each piece keeps its separator |
| `splitAfterN(string s, string sep, int n)` | `Array<string>` | Like `splitN`, keeping the separators |
| `fields(string s)` | `Array<string>` | The words: the runs of characters between whitespace |
| `fieldsFunc(string s, bool(char) f)` | `Array<string>` | The runs of characters that don't satisfy `f` |
| `join(Array<string> elems, string sep)` | `string` | The elements with `sep` between them |

```actionscript
package main;

use "std:string";

protected int main() {
    Array<string> cells = string.split("name,age,,city", ",");
    trace(cells.length, cells[1], cells[2] == "");          // 4 age true

    Array<string> words = string.fields("  to be   or not ");
    trace(string.join(words, " "));                          // to be or not

    Array<string> kv = string.splitN("key=value=more", "=", 2);
    trace(kv[0], kv[1]);                                     // key value=more
    return 0;
}
```

## Not yet available

These Go `strings` functions need language features RuneScript doesn't
have yet:

| Go function(s) | Needs |
|---|---|
| `Cut`, `CutPrefix`, `CutSuffix` | multiple return values |
| `IndexByte`, `LastIndexByte` | a byte type |
| `Lines`, `SplitSeq`, `SplitAfterSeq`, `FieldsSeq`, `FieldsFuncSeq` | iterators |
| `ToUpperSpecial`, `ToLowerSpecial`, `ToTitleSpecial` | a `unicode.SpecialCase` value |
| `Builder`, `Reader`, `Replacer`, `NewReader`, `NewReplacer` | classes |
| `Title` | not planned: deprecated in Go because it handles Unicode word boundaries incorrectly |

Next: [Compiling to JavaScript](11-javascript.md)
