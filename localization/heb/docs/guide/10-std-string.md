<!-- source: docs/guide/10-std-string.md @ 34d09ef -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/10-std-string.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 10. הספרייה הסטנדרטית: `std:string`

`std:string` היא הגרסה של RuneScript לחבילה [`strings`](https://pkg.go.dev/strings)
של Go. הכניסו אותה עם `use`, ואז קראו לפונקציות שלה דרך המזהה `string`:

</div>

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

<div dir="rtl">

## מוסכמות

- **קריאה.** `use "std:string";` נכתב אחרי שורת ה-`package` ולפני כל פונקציה. אחר
  כך קוראים ל-`string.<name>(...)`. בלי ה-`use`, המזהה אינו בטווח ההכרה
  (`ASX-NAME-007`).
- **שמות.** השמות של Go מומרים ל-camelCase: `ToUpper` ← `toUpper`, `TrimSpace` ←
  `trimSpace`, `ToValidUTF8` ← `toValidUTF8`. אם כותבים את האיות של Go, הודעת
  השגיאה מציעה את האיות הנכון.
- **מיקומים הם היסטים בבתים.** מחרוזות מכילות טקסט UTF-8, ומיקומים, אורכים וספירות
  נמדדים **בבתים**, בדיוק כמו ב-Go. ב-`"héllo"` התו `é` תופס שני בתים, ולכן
  `string.index("héllo", "l")` הוא `3`.
- **לתוצאות יש טיפוס.** פונקציות שמחזירות מיקומים או ספירות מחזירות `int`. השמה של
  תוצאה כזו ל-`uint` היא שגיאת `ASX-TYPE-004` הרגילה.
- **ההתנהגות תואמת ל-Go.** מקרי קצה, כמו מפריד ריק או `n` שלילי, מתנהגים כפי ש-Go
  מתעדת.
- **חיבור מחרוזות** נעשה עם `+` או עם מחרוזת משובצת `$"…"`, לא עם פונקציית ספרייה
  ([פרק 16](16-text-and-arrays.md)).

## מדריך הפונקציות

### השוואה וחיפוש

| פונקציה | מחזירה | תיאור |
|---|---|---|
| `compare(string a, string b)` | `int` | `-1`, `0` או `1` לפי סדר הבתים |
| `equalFold(string s, string t)` | `bool` | שוות בהתעלם מאותיות גדולות וקטנות ב-Unicode |
| `contains(string s, string substr)` | `bool` | `substr` מופיעה ב-`s` |
| `containsAny(string s, string chars)` | `bool` | תו כלשהו מ-`chars` מופיע ב-`s` |
| `hasPrefix(string s, string prefix)` | `bool` | `s` מתחילה ב-`prefix` |
| `hasSuffix(string s, string suffix)` | `bool` | `s` מסתיימת ב-`suffix` |
| `count(string s, string substr)` | `int` | מספר המופעים שאינם חופפים. עם `substr` ריקה: 1 + מספר התווים |
| `index(string s, string substr)` | `int` | ההיסט בבתים של המופע הראשון, או `-1` |
| `indexAny(string s, string chars)` | `int` | ההיסט בבתים של התו הראשון מתוך `chars`, או `-1` |
| `lastIndex(string s, string substr)` | `int` | ההיסט בבתים של המופע האחרון, או `-1` |
| `lastIndexAny(string s, string chars)` | `int` | ההיסט בבתים של התו האחרון מתוך `chars`, או `-1` |

</div>

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

<div dir="rtl">

### שינוי

| פונקציה | מחזירה | תיאור |
|---|---|---|
| `toUpper(string s)` | `string` | אותיות גדולות |
| `toLower(string s)` | `string` | אותיות קטנות |
| `toTitle(string s)` | `string` | אות כותרת לכל אות (title case של Unicode; ברוב האותיות, כמו אות גדולה) |
| `repeat(string s, int count)` | `string` | `s` חוזרת `count` פעמים. `count` שלילי הוא שגיאת זמן ריצה (`ASX-RUN-004`) |
| `replace(string s, string old, string new, int n)` | `string` | מחליפה את `n` המופעים הראשונים; `n < 0` מחליף את כולם |
| `replaceAll(string s, string old, string new)` | `string` | מחליפה כל מופע |
| `toValidUTF8(string s, string replacement)` | `string` | מחליפה רצפי UTF-8 לא תקינים |
| `clone(string s)` | `string` | עותק חדש (קיימת לשם השלמות) |

</div>

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

<div dir="rtl">

### קיצוץ

| פונקציה | מחזירה | תיאור |
|---|---|---|
| `trimSpace(string s)` | `string` | מסירה רווחים לבנים מההתחלה ומהסוף |
| `trim(string s, string cutset)` | `string` | מסירה מההתחלה ומהסוף תווים שנמצאים ב-`cutset` |
| `trimLeft(string s, string cutset)` | `string` | רק תווים מההתחלה |
| `trimRight(string s, string cutset)` | `string` | רק תווים מהסוף |
| `trimPrefix(string s, string prefix)` | `string` | מסירה את `prefix` פעם אחת, אם הוא קיים |
| `trimSuffix(string s, string suffix)` | `string` | מסירה את `suffix` פעם אחת, אם הוא קיים |

</div>

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

<div dir="rtl">

`trimPrefix`/`trimSuffix` מסירות מחרוזת שלמה פעם אחת. `trimLeft`/`trimRight`
מסירות כל רצף של התווים שברשימה.

### אורך וחיתוך

Go מודדת וחותכת מחרוזות עם `len(s)` ו-`s[i:j]` המובנים שלה. ב-RuneScript יש להן
פונקציות, שגם הן סופרות בבתים:

| פונקציה | מחזירה | תיאור |
|---|---|---|
| `len(string s)` | `int` | האורך בבתים |
| `runeCount(string s)` | `int` | מספר התווים (נקודות הקוד) |
| `slice(string s, int start, int end)` | `string` | הבתים מ-`start` ועד `end`, לא כולל |

`slice` אף פעם לא חוצה תו לשניים. גבולות מחוץ ל-`0 … len(s)`, `start > end`, או גבול
באמצע תו, זורקים `RangeError` (`ASX-RUN-004`).

</div>

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

<div dir="rtl">

שלבו את `index` ו-`slice` כדי לפרק מחרוזת, כמו ב-Go.

### תווים ו-callbacks

הפונקציות האלה מקבלות [`char`](16-text-and-arrays.md), או פונקציה
שבודקת תו אחד בכל פעם. כל callback נקרא באותו סדר כמו ב-Go: משמאל לימין, מלבד
`lastIndexFunc` והצד הימני של החיתוכים, שעובדים מהסוף.

| פונקציה | מחזירה | תיאור |
|---|---|---|
| `containsRune(string s, char c)` | `bool` | `c` מופיע ב-`s` |
| `indexRune(string s, char c)` | `int` | היסט הבתים של ה-`c` הראשון, או `-1` |
| `containsFunc(string s, bool(char) f)` | `bool` | תו כלשהו מקיים את `f` |
| `indexFunc(string s, bool(char) f)` | `int` | היסט הבתים של התו הראשון שמקיים את `f`, או `-1` |
| `lastIndexFunc(string s, bool(char) f)` | `int` | היסט הבתים של האחרון, או `-1` |
| `trimFunc(string s, bool(char) f)` | `string` | מסירה תווים בהתחלה ובסוף שמקיימים את `f` |
| `trimLeftFunc(string s, bool(char) f)` | `string` | רק אלה שבהתחלה |
| `trimRightFunc(string s, bool(char) f)` | `string` | רק אלה שבסוף |
| `map(char(char) f, string s)` | `string` | כל תו מוחלף ב-`f` שלו (שימו לב לסדר הארגומנטים של Go) |

</div>

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

<div dir="rtl">

בניגוד ל-`Map` של Go, `map` לא יכולה להשמיט תווים, כי `char` אף פעם אינו שלילי.

### פיצול וחיבור

הפונקציות האלה מחזירות או מקבלות [`Array<string>`](16-text-and-arrays.md):

| פונקציה | מחזירה | תיאור |
|---|---|---|
| `split(string s, string sep)` | `Array<string>` | החלקים שבין המפרידים. `sep` ריק מפצל לתווים |
| `splitN(string s, string sep, int n)` | `Array<string>` | לכל היותר `n` חלקים, והאחרון מכיל את השאר; `n < 0` פירושו הכול, `n == 0` אף אחד |
| `splitAfter(string s, string sep)` | `Array<string>` | כמו `split`, אבל כל חלק שומר את המפריד שלו |
| `splitAfterN(string s, string sep, int n)` | `Array<string>` | כמו `splitN`, ושומרת את המפרידים |
| `fields(string s)` | `Array<string>` | המילים: רצפי התווים שבין רווחים לבנים |
| `fieldsFunc(string s, bool(char) f)` | `Array<string>` | רצפי התווים שאינם מקיימים את `f` |
| `join(Array<string> elems, string sep)` | `string` | האיברים, עם `sep` ביניהם |

</div>

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

<div dir="rtl">

## עדיין לא זמין

הפונקציות האלה מ-`strings` של Go דורשות תכונות שפה שעדיין אין ב-RuneScript:

| פונקציות Go | דורשות |
|---|---|
| `Cut`, `CutPrefix`, `CutSuffix` | כמה ערכי החזרה |
| `IndexByte`, `LastIndexByte` | טיפוס בית |
| `Lines`, `SplitSeq`, `SplitAfterSeq`, `FieldsSeq`, `FieldsFuncSeq` | איטרטורים |
| `ToUpperSpecial`, `ToLowerSpecial`, `ToTitleSpecial` | ערך `unicode.SpecialCase` |
| `Builder`, `Reader`, `Replacer`, `NewReader`, `NewReplacer` | מחלקות |
| `Title` | לא מתוכנן: הוצא משימוש ב-Go כי הוא מטפל בגבולות מילים של Unicode באופן שגוי |

הבא: [הידור ל-JavaScript](11-javascript.md)

</div>
