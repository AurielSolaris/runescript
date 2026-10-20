<!-- source: docs/guide/21-json.md @ 6372b35 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/21-json.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 21. JSON: ‏`std:json`

`use "std:json";` קוראת וכותבת טקסט [JSON](https://www.json.org/). מסמך JSON הופך
ל-`json.Value`: ‏null, ‏bool, מספר, מחרוזת, מערך או אובייקט. ב-RuneScript אין טיפוס
שתופס הכול ([פרק 17](17-no-null-any-generics.md)), ולכן מסתכלים לתוך ערך בעזרת פונקציות
בדוקות, `json.asString(v)`, `json.get(v, "name")` וכן הלאה, וכל אחת מהן זורקת שגיאה אם
הערך אינו מה שביקשתם.

## קריאת JSON

`json.parse` הופכת טקסט לערך. `json.get` קוראת איבר של אובייקט, `json.at` איבר של
מערך, ופונקציות ה-`as…` הופכות ערך ל-`string`, `bool`, `int`, `int64` או `float64` של
RuneScript:

</div>

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

<div dir="rtl">

כדי לעבוד עם טיפוסים משלכם, העתיקו את השדות שאתם צריכים למבנה (struct):

</div>

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

<div dir="rtl">

## פונקציות קריאה

| פונקציה | מחזירה | תיאור |
|---|---|---|
| `kind(json.Value v)` | `string` | `"null"`, `"bool"`, `"number"`, `"string"`, `"array"` או `"object"` |
| `isNull(json.Value v)` | `bool` | האם `v` הוא null |
| `asBool(json.Value v)` | `bool` | הערך של bool |
| `asString(json.Value v)` | `string` | הערך של מחרוזת |
| `asInt(json.Value v)` | `int` | מספר שלם, שחייב להיכנס ב-`int` |
| `asInt64(json.Value v)` | `int64` | מספר שלם, שחייב להיכנס ב-`int64` |
| `asFloat(json.Value v)` | `float64` | מספר, מעוגל ל-`float64` הקרוב ביותר |
| `length(json.Value v)` | `int` | מספר האיברים של מערך, או של אובייקט |
| `at(json.Value v, int i)` | `json.Value` | האיבר `i` של מערך |
| `items(json.Value v)` | `Array<json.Value>` | האיברים של מערך, כמערך RuneScript חדש |
| `has(json.Value v, string key)` | `bool` | האם לאובייקט יש את האיבר `key` |
| `get(json.Value v, string key)` | `json.Value` | האיבר `key` של אובייקט |
| `keys(json.Value v)` | `Array<string>` | המפתחות של אובייקט, בסדר שבמסמך |

`asInt` ו-`asInt64` מקבלות רק מספרים שנכתבו כמספרים שלמים: `2` כן, `2.0` ו-`2e0` לא.
להם השתמשו ב-`asFloat`.

## בנייה וכתיבה של JSON

פונקציות ה-`from…` יוצרות ערכים, ו-`fromArray` ו-`fromObject` מחברות אותם יחד.
`json.stringify` כותבת ערך כטקסט דחוס; `json.stringifyIndent` שמה כל איבר בשורה משלו,
מוזח במחרוזת שאתם נותנים:

</div>

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

<div dir="rtl">

ה-`trace` השני מדפיס:

</div>

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

<div dir="rtl">

| פונקציה | מחזירה | תיאור |
|---|---|---|
| `nullValue()` | `json.Value` | ה-`null` של JSON |
| `fromBool(bool b)` | `json.Value` | `true` או `false` |
| `fromString(string s)` | `json.Value` | מחרוזת |
| `fromInt(int n)`, `fromInt64(int64 n)` | `json.Value` | מספר שלם |
| `fromFloat(float64 f)` | `json.Value` | מספר, נכתב כמו ש-`trace` כותב אותו (`0.1`, `1e+21`) |
| `fromArray(Array<json.Value> items)` | `json.Value` | מערך |
| `fromObject(Array<string> keys, Array<json.Value> values)` | `json.Value` | אובייקט; לשני המערכים חייב להיות אותו אורך |
| `stringify(json.Value v)` | `string` | טקסט JSON דחוס |
| `stringifyIndent(json.Value v, string indent)` | `string` | טקסט JSON, איבר אחד בכל שורה |

מחרוזות נכתבות כמו ש-`JSON.stringify` של JavaScript כותבת אותן: `"`, לוכסן הפוך ותווי
בקרה מוברחים (escaped), וכל השאר נכתב כמו שהוא.

## קבצים

קראו קובץ JSON עם `fs.readFile` ([פרק 20](20-files-and-input.md)) וכתבו קובץ עם
`fs.writeFile`:

</div>

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

<div dir="rtl">

## שגיאות

| בעיה | שגיאה |
|---|---|
| טקסט שאינו JSON | `Error`: `json.parse: unexpected character ']' at byte 6` (ההיסטים סופרים בתים של UTF-8 מ-`0`) |
| ערך מהסוג הלא נכון | `TypeError`: `json.asString: expected a string, found a number` |
| `asInt` של `1.5` | `TypeError`: `json.asInt: 1.5 is not an integer` |
| מספר גדול מדי עבור `int`, `int64` או `float64` | `RangeError`: `json.asInt: 3000000000 does not fit in 'int'` |
| `at` אחרי הסוף | `RangeError`: `json.at: index 2 is out of range for an array of length 2` |
| `get` של מפתח חסר | `Error`: `json.get: no key "phone"` (בדקו קודם עם `has`) |
| `fromFloat` של NaN או אינסוף | `ArgumentError`: `json.fromFloat: NaN cannot be written in JSON` |
| `fromObject` עם מערכים באורכים שונים | `ArgumentError`: `json.fromObject: 2 keys but 1 value` |

</div>

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

<div dir="rtl">

## פרטים

- **ערך אף פעם לא משתנה.** `items` ו-`keys` מחזירות מערכים חדשים, ואין פונקציה שמשנה ערך
  במקום: בנו ערך חדש עם `fromArray` או `fromObject`. אפשר לשתף ערכים בחופשיות, והם לא
  יכולים להכיל את עצמם.
- **מספרים שומרים על הטקסט שלהם.** `12345678901234567890` נקרא בדיוק על ידי `stringify` ועל
  ידי `asInt64` (שמדווחת שהוא לא נכנס), ומעוגל רק על ידי `asFloat`. כתיבה של מסמך שפוענח
  מחזירה את המספרים שלו כפי שנכתבו: `1.50` נשאר `1.50`.
- **אובייקטים שומרים על הסדר שלהם.** `keys` ו-`stringify` הולכות לפי המסמך. אם מפתח מופיע
  פעמיים, האובייקט שומר את המקום הראשון ואת הערך האחרון, כמו ב-JavaScript.
- **מחרוזות הן Unicode תקין.** בריחות `\u` מפוענחות, וזוג surrogate הופך לתו אחד; surrogate
  בודד (`"\ud83d"`) הופך ל-`�` (U+FFFD), כי מחרוזות RuneScript הן תמיד UTF-8 תקין.
- **רק JSON מחמיר.** בלי הערות, פסיקים בסוף, מירכאות בודדות או אפסים מובילים. מערכים
  ואובייקטים יכולים להיות מקוננים עד עומק 1,000.
- **ל-`json.Value` אין ערך אפס**, ולכן `json.Value v;` היא שגיאה; תנו לו ערך. אי אפשר
  להשוות ערכים עם `==` או להעביר אותם ישירות ל-`trace`; העבירו ל-trace את
  `json.stringify(v)`.
- **שני היעדים מסכימים.** המפרש ויעד ה-JavaScript משתמשים באותו מפענח, כך שטקסט מתקבל או
  נדחה, וכל שגיאה נקראת, באותו אופן בשניהם.

הבא: חזרה ל[תוכן העניינים של המדריך](README.md)

</div>
