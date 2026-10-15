<!-- source: docs/guide/10-std-string.md @ c31fd14 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/10-std-string.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 10. הספרייה הסטנדרטית: `std:string`

`std:string` היא הגרסה של RuneScript לחבילה [`strings`](https://pkg.go.dev/strings)
של Go. הכניסו אותה עם `use`, ואז קראו לפונקציות שלה דרך המזהה `string`:

</div>

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

public void main() {
    trace(string.repeat("ab", 3));                          // ababab
    trace(string.replace("oink oink oink", "k", "ky", 2));  // oinky oinky oink
    trace(string.toLower("LOUD"), string.toUpper("quiet"));  // loud QUIET
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

public void main() {
    trace(string.trim("!!hello!!", "!"));                 // hello
    trace(string.trimLeft("xxhixx", "x"));                // hixx
    trace(string.trimSuffix("game.rsc", ".rsc"));           // game
}
```

<div dir="rtl">

`trimPrefix`/`trimSuffix` מסירות מחרוזת שלמה פעם אחת. `trimLeft`/`trimRight`
מסירות כל רצף של התווים שברשימה.

## עדיין לא זמין

הפונקציות האלה מ-`strings` של Go דורשות תכונות שפה שעדיין אין ב-RuneScript:

| פונקציות Go | דורשות |
|---|---|
| `Split`, `SplitN`, `SplitAfter`, `SplitAfterN`, `Fields`, `Join` | מערכים (`[]string`) |
| `Cut`, `CutPrefix`, `CutSuffix` | כמה ערכי החזרה |
| `ContainsRune`, `IndexRune`, `IndexByte`, `LastIndexByte` | טיפוס `char` / בית |
| `ContainsFunc`, `IndexFunc`, `LastIndexFunc`, `FieldsFunc`, `TrimFunc`, `TrimLeftFunc`, `TrimRightFunc`, `Map` | ערכי פונקציה |
| `Lines`, `SplitSeq`, `SplitAfterSeq`, `FieldsSeq`, `FieldsFuncSeq` | איטרטורים |
| `ToUpperSpecial`, `ToLowerSpecial`, `ToTitleSpecial` | ערך `unicode.SpecialCase` |
| `Builder`, `Reader`, `Replacer`, `NewReader`, `NewReplacer` | מחלקות |
| `Title` | לא מתוכנן: הוצא משימוש ב-Go כי הוא מטפל בגבולות מילים של Unicode באופן שגוי |

עדיין אין גם פונקציית אורך, כי Go משתמשת ב-`len` המובנית שלה. פעולת אורך מתוכננת
יחד עם מערכים.

הבא: [הידור ל-JavaScript](11-javascript.md)

</div>
