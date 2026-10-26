# 21. JSON: `std:json`

`use "std:json";` reads and writes [JSON](https://www.json.org/) text. A
JSON document becomes a `json.Value`: a null, a bool, a number, a string,
an array, or an object. RuneScript has no catch-all type
([chapter 17](17-no-null-any-generics.md)), so you look inside a value with
checked functions, `json.asString(v)`, `json.get(v, "name")`, and so on,
and each one throws an error if the value isn't what you asked for.

## Reading JSON

`json.parse` turns text into a value. `json.get` reads an object's member,
`json.at` an array's element, and the `as…` functions turn a value into a
RuneScript `string`, `bool`, `int`, `int64`, or `float64`:

```actionscript
package main;

use "std:json";

protected int main() {
    string text = "{\"name\": \"Ada\", \"born\": 1815, \"langs\": [\"English\", \"French\"]}";
    json.Value person = json.parse(text);
    string name = json.asString(json.get(person, "name"));
    int born = json.asInt(json.get(person, "born"));
    trace(name, born);                                   // Ada 1815
    for each (json.Value lang in json.items(json.get(person, "langs"))) {
        trace(json.asString(lang));                      // English, then French
    }
    trace(json.has(person, "died"));                     // false
    return 0;
}
```

To work with your own types, copy the fields you need into a struct:

```actionscript
package main;

use "std:json";

public struct Item {
    string name;
    int count;
    float64 price;
}

public Item readItem(json.Value v) {
    return Item{
        name: json.asString(json.get(v, "name")),
        count: json.asInt(json.get(v, "count")),
        price: json.asFloat(json.get(v, "price"))
    };
}

protected int main() {
    json.Value order = json.parse("[{\"name\": \"lamp\", \"count\": 2, \"price\": 19.5}, {\"name\": \"rug\", \"count\": 1, \"price\": 120}]");
    float64 total = 0.0;
    for each (json.Value v in json.items(order)) {
        Item it = readItem(v);
        total += (float64) it.count * it.price;
    }
    trace(total);   // 159
    return 0;
}
```

## Reading functions

| Function | Returns | Description |
|---|---|---|
| `kind(json.Value v)` | `string` | `"null"`, `"bool"`, `"number"`, `"string"`, `"array"`, or `"object"` |
| `isNull(json.Value v)` | `bool` | Whether `v` is null |
| `asBool(json.Value v)` | `bool` | A bool's value |
| `asString(json.Value v)` | `string` | A string's value |
| `asInt(json.Value v)` | `int` | An integer number, which must fit in `int` |
| `asInt64(json.Value v)` | `int64` | An integer number, which must fit in `int64` |
| `asFloat(json.Value v)` | `float64` | A number, rounded to the nearest `float64` |
| `length(json.Value v)` | `int` | The number of elements of an array, or of members of an object |
| `at(json.Value v, int i)` | `json.Value` | Element `i` of an array |
| `items(json.Value v)` | `Array<json.Value>` | An array's elements, as a new RuneScript array |
| `has(json.Value v, string key)` | `bool` | Whether an object has the member `key` |
| `get(json.Value v, string key)` | `json.Value` | The member `key` of an object |
| `keys(json.Value v)` | `Array<string>` | An object's keys, in the order of the document |

`asInt` and `asInt64` take only numbers written as integers: `2` is one,
`2.0` and `2e0` are not. Use `asFloat` for those.

## Building and writing JSON

The `from…` functions make values, and `fromArray` and `fromObject` put them
together. `json.stringify` writes a value as compact text;
`json.stringifyIndent` puts each element and member on its own line,
indented by the string you give:

```actionscript
package main;

use "std:json";

protected int main() {
    Array<json.Value> tags = [json.fromString("math"), json.fromString("poetry")];
    json.Value person = json.fromObject(
        ["name", "born", "tags", "spouse"],
        [json.fromString("Ada"), json.fromInt(1815), json.fromArray(tags), json.nullValue()]);
    trace(json.stringify(person));
    // {"name":"Ada","born":1815,"tags":["math","poetry"],"spouse":null}
    trace(json.stringifyIndent(person, "  "));
    return 0;
}
```

The second `trace` prints:

```text
{
  "name": "Ada",
  "born": 1815,
  "tags": [
    "math",
    "poetry"
  ],
  "spouse": null
}
```

| Function | Returns | Description |
|---|---|---|
| `nullValue()` | `json.Value` | JSON's `null` |
| `fromBool(bool b)` | `json.Value` | `true` or `false` |
| `fromString(string s)` | `json.Value` | A string |
| `fromInt(int n)`, `fromInt64(int64 n)` | `json.Value` | An integer number |
| `fromFloat(float64 f)` | `json.Value` | A number, written as `trace` writes it (`0.1`, `1e+21`) |
| `fromArray(Array<json.Value> items)` | `json.Value` | An array |
| `fromObject(Array<string> keys, Array<json.Value> values)` | `json.Value` | An object; the two arrays must have the same length |
| `stringify(json.Value v)` | `string` | Compact JSON text |
| `stringifyIndent(json.Value v, string indent)` | `string` | JSON text, one element or member per line |

Strings are written as JavaScript's `JSON.stringify` writes them: `"`, `\`,
and control characters are escaped, everything else is written as it is.

## Files

Read a JSON file with `fs.readFile` ([chapter 20](20-files-and-input.md))
and write one with `fs.writeFile`:

```actionscript
package main;

use "std:fs";
use "std:json";
use "std:path";

protected int main() {
    string file = path.join(fs.tempDir(), "runescript-settings.json");
    fs.writeFile(file, "{\"volume\": 7, \"theme\": \"ice\"}");

    json.Value settings = json.parse(fs.readFile(file));
    int volume = json.asInt(json.get(settings, "volume"));
    json.Value changed = json.fromObject(["volume", "theme"], [json.fromInt(volume + 1), json.get(settings, "theme")]);
    fs.writeFile(file, json.stringifyIndent(changed, "    ") + "\n");
    trace(json.asInt(json.get(json.parse(fs.readFile(file)), "volume")));   // 8
    fs.remove(file);
    return 0;
}
```

## Errors

| Problem | Error |
|---|---|
| Text that isn't JSON | `Error`: `json.parse: unexpected character ']' at byte 6` (offsets count UTF-8 bytes from `0`) |
| A value of the wrong kind | `TypeError`: `json.asString: expected a string, found a number` |
| `asInt` of `1.5` | `TypeError`: `json.asInt: 1.5 is not an integer` |
| A number too big for `int`, `int64`, or `float64` | `RangeError`: `json.asInt: 3000000000 does not fit in 'int'` |
| `at` past the end | `RangeError`: `json.at: index 2 is out of range for an array of length 2` |
| `get` of a missing key | `Error`: `json.get: no key "phone"` (check with `has` first) |
| `fromFloat` of NaN or an infinity | `ArgumentError`: `json.fromFloat: NaN cannot be written in JSON` |
| `fromObject` with arrays of different lengths | `ArgumentError`: `json.fromObject: 2 keys but 1 value` |

```actionscript
package main;

use "std:json";

protected int main() {
    try {
        json.Value v = json.parse("[1, 2,]");
        trace(json.length(v));
    } catch (Error e) {
        trace(e.message);   // json.parse: unexpected character ']' at byte 6
    }
    json.Value n = json.parse("12.5");
    try {
        trace(json.asString(n));
    } catch (TypeError e) {
        trace(e.message);   // json.asString: expected a string, found a number
    }
    return 0;
}
```

## Details

- **A value never changes.** `items` and `keys` return new arrays, and
  there is no function that changes a value in place: build a new one with
  `fromArray` or `fromObject`. Values can be shared freely and can't
  contain themselves.
- **Numbers keep their text.** `12345678901234567890` is read exactly by
  `stringify` and by `asInt64` (which reports that it doesn't fit), and
  rounded only by `asFloat`. Writing a parsed document gives back its
  numbers as they were written: `1.50` stays `1.50`.
- **Objects keep their order.** `keys` and `stringify` follow the document.
  If a key appears twice, the object keeps the first position and the last
  value, as JavaScript does.
- **Strings are valid Unicode.** `\u` escapes are decoded, a surrogate pair
  becoming one character; a lone surrogate (`"\ud83d"`) becomes `�`
  (U+FFFD), because RuneScript strings are always valid UTF-8.
- **Strict JSON only.** No comments, trailing commas, single quotes, or
  leading zeros. Arrays and objects can nest 1,000 deep.
- **`json.Value` has no zero value**, so `json.Value v;` is an error; give
  it a value. Values can't be compared with `==` or passed to `trace`
  directly; trace `json.stringify(v)`.
- **Both targets agree.** The interpreter and the JavaScript target use the
  same parser, so a text is accepted or rejected, and every error reads,
  the same way on both.

Next: [CSV, matrices, and versions](22-csv-matrix-semver.md)
