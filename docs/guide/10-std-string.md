# 10. Standard library: `std:string`

`std:string` is RuneScript's version of Go's
[`strings`](https://pkg.go.dev/strings) package. Bring it in with `use`,
then call its functions through the `string` qualifier:

```actionscript
package main;

use "std:string";

public void main() {
    string name = string.trimSpace("   rune script   ");
    trace(string.toUpper(name));                       // RUNE SCRIPT
    trace(string.hasPrefix(name, "rune"));             // true
    trace(string.replaceAll(name, " ", "-"));          // rune-script
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

public void main() {
    string s = "go gopher";
    trace(string.contains(s, "ph"), string.index(s, "go"), string.lastIndex(s, "go"));   // true 0 3
    trace(string.count("cheese", "e"));                                                  // 3
    trace(string.equalFold("RuneScript", "runescript"));                                 // true
    if (string.index(s, "x") == -1) {
        trace("no x");
    }
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

public void main() {
    trace(string.repeat("ab", 3));                          // ababab
    trace(string.replace("oink oink oink", "k", "ky", 2));  // oinky oinky oink
    trace(string.toLower("LOUD"), string.toUpper("quiet"));  // loud QUIET
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

public void main() {
    trace(string.trim("!!hello!!", "!"));                 // hello
    trace(string.trimLeft("xxhixx", "x"));                // hixx
    trace(string.trimSuffix("game.rs", ".rs"));           // game
}
```

`trimPrefix`/`trimSuffix` remove a whole string once. `trimLeft`/`trimRight`
remove any run of the listed characters.

## Not yet available

These Go `strings` functions need language features RuneScript doesn't
have yet:

| Go function(s) | Needs |
|---|---|
| `Split`, `SplitN`, `SplitAfter`, `SplitAfterN`, `Fields`, `Join` | arrays (`[]string`) |
| `Cut`, `CutPrefix`, `CutSuffix` | multiple return values |
| `ContainsRune`, `IndexRune`, `IndexByte`, `LastIndexByte` | `char` / byte type |
| `ContainsFunc`, `IndexFunc`, `LastIndexFunc`, `FieldsFunc`, `TrimFunc`, `TrimLeftFunc`, `TrimRightFunc`, `Map` | function values |
| `Lines`, `SplitSeq`, `SplitAfterSeq`, `FieldsSeq`, `FieldsFuncSeq` | iterators |
| `ToUpperSpecial`, `ToLowerSpecial`, `ToTitleSpecial` | a `unicode.SpecialCase` value |
| `Builder`, `Reader`, `Replacer`, `NewReader`, `NewReplacer` | classes |
| `Title` | not planned: deprecated in Go because it handles Unicode word boundaries incorrectly |

There's also no length function yet, because Go uses its built-in `len`.
A length operation is planned alongside arrays.

Next: back to the [guide index](README.md)
