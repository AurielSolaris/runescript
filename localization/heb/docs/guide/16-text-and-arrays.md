<!-- source: docs/guide/16-text-and-arrays.md @ 34d09ef -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/16-text-and-arrays.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 16. טקסט ורשימות: חיבור מחרוזות, `char` ומערכים

הפרק הזה עוסק בכלים היומיומיים של כתיבת סקריפטים: בניית מחרוזות, עבודה עם תווים
בודדים ושמירת ערכים ברשימה. הפונקציות שמחפשות, מפצלות ומשנות מחרוזות נמצאות
ב[פרק 10](10-std-string.md).

## חיבור מחרוזות עם `+`

`+` מחבר שתי מחרוזות, ו-`+=` מוסיף למשתנה מחרוזת:

</div>

```actionscript
package main;

public void main() {
    string name = "Ada";
    string greeting = "Hello, " + name + "!";
    greeting += " Welcome.";
    trace(greeting);   // Hello, Ada! Welcome.
}
```

<div dir="rtl">

`+` אף פעם לא ממיר מספר לטקסט בשבילכם, כי RuneScript לא ממירה ערכים מאחורי הגב.
המירו בעצמכם עם `(string) x`, או השתמשו במחרוזת משובצת (בסעיף הבא):

</div>

```actionscript
package main;

public void main() {
    int count = 3;
    trace("count: " + (string) count);   // count: 3
    trace("count: " + count);            // error: ASX-TYPE-012
}
```

<div dir="rtl">

זה שונה מ-ActionScript 3 ומ-JavaScript, שבהן `"count: " + 3` הופך בשקט ל-`"count: 3"`.

## מחרוזות משובצות: `$"…"`

כתבו `$` לפני מחרוזת, וכל `{expression}` בתוכה יוחלף בערך של הביטוי:

</div>

```actionscript
package main;

public void main() {
    string item = "apple";
    int count = 3;
    float price = 0.25;
    trace($"{count} x {item} at {price} = {(float) count * price}");   // 3 x apple at 0.25 = 0.75
    trace($"ready: {count > 0}");                                     // ready: true
    trace($"{{braces}} are written twice");                           // {braces} are written twice
}
```

<div dir="rtl">

- חור יכול להכיל כל ביטוי: חשבון, קריאות, מחרוזות אחרות, ואפילו מחרוזת משובצת נוספת.
- הערך שלו מוצג בדיוק כפי ש-`(string) x` היה מציג אותו, ולכן הוא חייב להיות מספר,
  `bool`, `char` או מחרוזת. כל דבר אחר, כמו פונקציה או מערך, הוא `ASX-TYPE-017`.
- `{{` ו-`}}` מייצגים `{` ו-`}` מילוליים. `}` בודד הוא שגיאה (`ASX-LEX-003`).
- תווי בריחה עובדים כמו במחרוזות אחרות: `\n`, `\t`, `\"` וכן הלאה.

## תווים: `char`

`char` הוא תו אחד, נקודת קוד של Unicode. כתבו אותו בין גרשיים בודדים:

</div>

```actionscript
package main;

public void main() {
    char letter = 'a';
    char accent = 'é';
    char emoji = '😀';
    char quote = '\'';
    trace(letter, accent, emoji, quote, '\\');   // a é 😀 ' \
    trace(letter < 'b', accent == 'é');          // true true
}
```

<div dir="rtl">

- ליטרל תו מכיל בדיוק תו אחד. `''` ו-`'ab'` הם `ASX-LEX-006`. תווי הבריחה הם אלה של
  מחרוזות, ובנוסף `\'`.
- אפשר להשוות תווים עם `==`, `!=`, `<`, `<=`, `>` ו-`>=`, לפי נקודת הקוד.
- `char` אינו מספר, ולכן אין `'a' + 1`. המירו אותו קודם.

ההמרות נבדקות, כמו כל המרה ([פרק 13](13-floats-and-conversions.md)):

| המרה | תוצאה | נכשלת עם |
|---|---|---|
| `(int) c`, `(uint) c` | נקודת הקוד: `(int) 'a'` הוא `97` | אף פעם |
| `(char) n` | התו שנקודת הקוד שלו היא `n`: `(char) 98` הוא `b` | `RangeError` אם `n` אינה נקודת קוד תקינה (שלילית, מעל `0x10FFFF`, או surrogate בטווח `0xD800`–`0xDFFF`) |
| `(string) c` | מחרוזת של תו אחד | אף פעם |
| `(char) s` | התו היחיד של `s` | `TypeError` אלא אם ב-`s` יש בדיוק תו אחד |

</div>

```actionscript
package main;

public char next(char c) {
    return (char) ((int) c + 1);
}

public void main() {
    trace(next('a'), (int) 'A', (char) "é");   // b 65 é
    trace((string) 'x' + "yz");               // xyz
}
```

<div dir="rtl">

פונקציות התווים של `std:string` מקבלות תווים: ראו
[תווים ו-callbacks](10-std-string.md).

## מערכים: `Array<T>`

`Array<T>` הוא רשימה של ערכים מטיפוס `T` שיכולה לגדול ולקטון. צרו אחד עם סוגריים
מרובעים:

</div>

```actionscript
package main;

public void main() {
    Array<string> names = ["Ada", "Grace"];
    names.push("Linus");                      // add to the end
    trace(names.length, names[0], names[2]);  // 3 Ada Linus

    names[1] = "Hopper";                      // replace an element
    names[0] += "!";
    string last = names.pop();                // remove the last one
    trace(last, names.length, names[1]);      // Linus 2 Hopper
}
```

<div dir="rtl">

| פעולה | משמעות |
|---|---|
| `[a, b, c]` | מערך חדש; `[]` הוא מערך ריק |
| `a[i]` | האיבר באינדקס `i` (מטיפוס `int` או `uint`), בספירה מ-`0` |
| `a[i] = v`, `a[i] += v`, … | החלפת איבר |
| `a.length` | מספר האיברים |
| `a.push(v)` | הוספת `v` בסוף |
| `a.pop()` | הסרת האיבר האחרון והחזרתו |

טיפוס האיבר `T` יכול להיות כל טיפוס ערך: מספרים, `bool`, `string`, `char`,
פונקציות, שגיאות ומערכים אחרים (`Array<Array<int>>`).

**הטיפוס נקבע לפי המקום שאליו המערך הולך.** `Array<uint> u = [1, 2];` הופך את `1`
ו-`2` ל-`uint`. כשאין טיפוס מוצהר מסביב, ליטרל מקבל את הטיפוס של האיבר הראשון שלו,
ולכן `[1, 2]` הוא `Array<int>`. `[]` ריק צריך טיפוס מוצהר: `Array<string> names = [];`.

**האינדקסים נבדקים.** אינדקס קטן מ-`0` או שאינו קטן מ-`length`, וכן `pop()` על מערך
ריק, זורקים `RangeError` (`ASX-RUN-011`) שאפשר לתפוס:

</div>

```actionscript
package main;

public void main() {
    Array<int> scores = [90, 75];
    try {
        trace(scores[2]);
    } catch (RangeError e) {
        trace(e.message);   // index 2 is out of range for an array of length 2
    }
}
```

<div dir="rtl">

**עברו על מערך בלולאה** בעזרת אינדקס:

</div>

```actionscript
package main;

public int sum(Array<int> values) {
    int total = 0;
    for (int i = 0; i < values.length; i++) {
        total += values[i];
    }
    return total;
}

public void main() {
    trace(sum([3, 4, 5]));   // 12
}
```

<div dir="rtl">

**מערכים משותפים, לא מועתקים.** השמת מערך למשתנה אחר, או העברתו לפונקציה, נותנת שם
שני לאותו מערך *עצמו*, כמו ב-ActionScript 3 וב-JavaScript. שינוי דרך שם אחד נראה
דרך כל השמות:

</div>

```actionscript
package main;

public void addDefaults(Array<string> list) {
    list.push("default");
}

public void main() {
    Array<string> mine = ["first"];
    Array<string> same = mine;
    addDefaults(same);
    trace(mine.length, mine[1]);   // 2 default
}
```

<div dir="rtl">

אי אפשר להעביר מערכים ישירות ל-trace. חברו אותם למחרוזת אחת עם `string.join`
([פרק 10](10-std-string.md)):

</div>

```actionscript
package main;

use "std:string";

public void main() {
    Array<string> words = string.fields("  the quick  brown fox ");
    trace(words.length, string.join(words, "-"));   // 4 the-quick-brown-fox
}
```

<div dir="rtl">

> **עדיין לא זמין:** לולאת for-each, חיתוך, מיון, `a[i]++`, השוואת מערכים והעברת
> מערכים לפונקציות `extern` של JavaScript. `main` עדיין לא יכולה לקבל ארגומנטים משורת
> הפקודה.

הבא: [בלי `null`, בלי `any`, בלי גנריות](17-no-null-any-generics.md)

</div>
