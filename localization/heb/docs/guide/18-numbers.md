<!-- source: docs/guide/18-numbers.md @ 362a461 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/18-numbers.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 18. מספרים לעומק: רוחב, ליטרלים, סיביות ו-`std:math`

[פרק 3](03-types-and-variables.md) הציג את `int`, `uint` וטיפוסי הנקודה הצפה. הפרק הזה
מכסה את השאר: מספרים שלמים ברוחב מדויק, דרכים נוספות לכתוב מספר, פעולות על סיביות והספרייה
`std:math`.

## מספרים שלמים ברוחב מדויק

מלבד `int` ו-`uint`, יש טיפוסי מספרים שלמים של בדיוק 8, 16, 32 ו-64 סיביות:

| טיפוס | טווח |
|---|---|
| `int8` | −128 … 127 |
| `int16` | −32768 … 32767 |
| `int32` | −2147483648 … 2147483647 |
| `int64` | −9223372036854775808 … 9223372036854775807 |
| `uint8` | 0 … 255 |
| `uint16` | 0 … 65535 |
| `uint32` | 0 … 4294967295 |
| `uint64` | 0 … 18446744073709551615 |

</div>

```actionscript
package main;

protected int main() {
    uint8 red = 255;
    int64 population = 8_100_000_000;
    uint64 hash = 14695981039346656037;
    trace(red, population, hash);
    return 0;
}
```

<div dir="rtl">

כל טיפוס של מספר שלם הוא טיפוס בפני עצמו, ולכן שום דבר לא מומר ביניהם מעצמו, אפילו לא
לטיפוס רחב יותר. המירו במפורש:

</div>

```actionscript
package main;

protected int main() {
    int8 small = 100;
    int16 wider = (int16) small;   // fine: an explicit conversion
    int64 total = (int64) wider * 1_000_000;
    trace(wider, total);
    int16 oops = small;            // error: ASX-TYPE-001
    return 0;
}
```

<div dir="rtl">

גם `int32` אינו `int`: ‏`int` הוא המספר השלם הכללי, שתוצאות הביניים שלו מתרחבות כשחישוב
צריך יותר מקום ([פרק 4](04-operators.md)), ו-`int32` לעולם לא.

החשבון בכל רוחב נבדק, כמו ב-`int`: תוצאה שאינה נכנסת זורקת `RangeError` במקום להתגלגל.

</div>

```actionscript
package main;

protected int main() {
    uint8 level = 250;
    try {
        level += 10;
        trace(level);
    } catch (RangeError e) {
        trace(e.message);   // integer overflow: 250 + 10 = 260 does not fit in 'uint8'
    }
    return 0;
}
```

<div dir="rtl">

## כתיבת מספרים

אפשר לכתוב מספרים שלמים בהקסדצימלי, בבינארי או באוקטלי, ו-`_` יכול להפריד ספרות בכל מספר:

| צורה | דוגמה | ערך |
|---|---|---|
| עשרוני | `1_000_000` | 1000000 |
| הקסדצימלי | `0xFF`, `0xDEAD_BEEF` | 255, 3735928559 |
| בינארי | `0b1010_0101` | 165 |
| אוקטלי | `0o755` | 493 |
| נקודה צפה | `1_234.5`, `6.022_140_76e23` | |

הקידומות נכתבות באותיות קטנות, ו-`_` חייב לעמוד בין שתי ספרות. אפס מוביל, כמו `0755`, הוא
שגיאה (`ASX-LEX-005`), כדי שלא יתבלבל עם אוקטלי.

## פעולות על סיביות

| אופרטור | משמעות |
|---|---|
| `a & b` | הסיביות שדולקות בשניהם |
| `a \| b` | הסיביות שדולקות באחד מהם |
| `a ^ b` | הסיביות שדולקות בדיוק באחד |
| `~a` | כל סיבית הפוכה, בתוך רוחב הטיפוס |
| `a << n` | הזזה שמאלה ב-`n` סיביות |
| `a >> n` | הזזה ימינה ב-`n` סיביות |

לכל אחד יש גם צורת השמה: `&=`, `|=`, `^=`, `<<=`, `>>=`.

</div>

```actionscript
package main;

protected int main() {
    uint8 flags = 0b0000_0101;
    flags |= 0b1000_0000;           // set a bit
    flags &= ~((uint8) 0b0100);     // clear a bit
    trace(flags, (flags & 0b1000_0000) != 0);   // 129 true

    uint32 color = 0xFF8800;
    uint32 red = (color >> 16) & 0xFF;
    uint32 green = (color >> 8) & 0xFF;
    trace(red, green, color & 0xFF);            // 255 136 0
    return 0;
}
```

<div dir="rtl">

- `&`, `|` ו-`^` מקבלים שני מספרים שלמים מאותו טיפוס.
- `>>` שומר על הסימן של מספר עם סימן: ‏`-16 >> 2` הוא `-4`.
- מספר ההזזות חייב להיות בין 0 לרוחב פחות 1, כלומר 0 עד 31 עבור `int`. כל דבר אחר הוא
  שגיאה (`RangeError` בזמן ריצה, `ASX-RUN-012`).
- `<<` אף פעם לא מאבד סיביות בשקט. אם התוצאה לא נכנסת לטיפוס, זו שגיאת גלישה, בדיוק כמו
  ב-`*`. כשרוצים להשמיט סיביות עליונות, מסכו את הערך קודם.
- `&` נקשר חלש יותר מ-`==`, כמו ב-C וב-ActionScript, כך ש-`x & mask == 0` פירושו
  `x & (mask == 0)`. זה מערבב מספר שלם עם `bool`, ולכן המהדר עוצר אותו. כתבו
  `(x & mask) == 0`.

## `std:math`

‏`use "std:math";` מביא את החבילה [`math`](https://pkg.go.dev/math) של Go, עם שמות
ב-camelCase, שעובדת על `float64`:

</div>

```actionscript
package main;

use "std:math";

protected int main() {
    float64 pi = math.pi;
    float64 angle = pi / 6;
    trace(math.sin(angle), math.cos(angle));    // 0.49999999999999994 0.8660254037844387
    trace(math.sqrt(2.0), math.pow(2.0, 10.0)); // 1.4142135623730951 1024
    trace(math.floor(-2.5), math.round(-2.5), math.roundToEven(2.5));   // -3 -3 2
    trace(math.max(3, 7), math.abs(-12), math.clamp(15, 0, 10));        // 7 12 10
    return 0;
}
```

<div dir="rtl">

| קבוצה | פונקציות |
|---|---|
| עיגול | `abs`, `ceil`, `floor`, `trunc`, `round` (חצאים הרחק מאפס), `roundToEven` |
| חזקות ושורשים | `sqrt`, `cbrt`, `pow`, `hypot`, `exp`, `exp2`, `expm1`, `ldexp` |
| לוגריתמים | `log`, `log10`, `log2`, `log1p` |
| טריגונומטריה | `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `atan2` |
| היפרבוליות | `sinh`, `cosh`, `tanh`, `asinh`, `acosh`, `atanh` |
| השוואה וחלקים | `min`, `max`, `mod`, `remainder`, `dim`, `copysign`, `signbit`, `nextafter` |
| ערכים מיוחדים | `isInf(x, sign)`, `inf(sign)`, `nan()` (והמובנות `isNaN`, `isFinite`) |
| מספרים שלמים | `abs`, `min`, `max` ו-`clamp(x, lo, hi)` על `int` |

**קבועים** כמו `math.pi`, `math.e`, `math.sqrt2`, `math.maxFloat64`, ו-`math.maxInt64` או
`math.maxUint8` (אחד לכל טיפוס של מספר שלם) הם בלי טיפוס, כמו ליטרל: הם מקבלים את הטיפוס של
המקום שבו משתמשים בהם, אם הם נכנסים. אי אפשר לשמור את `math.pi` ב-`float32` בלי עיגול, ולכן
שם צריך `(float32) math.pi`.

**‏`min`, `max` ו-`abs` עובדות על `int` או על `float64`.** המהדר בוחר בגרסת ה-`int` כשכל
ארגומנט הוא `int` או ליטרל של מספר שלם, ובגרסת ה-`float64` אחרת. ערבוב משתנה `int` עם
`float64` הוא שגיאה: המירו אחד מהם.

**טיפוסי מספר אחרים מומרים קודם:** ‏`math.sqrt((float64) f)` עבור `float` או `float32`,
ו-`(float64) n` עבור מספר שלם.

**אותן סיביות בכל מקום.** כל פונקציה של `std:math` מחזירה בדיוק את אותה תוצאה במפרש
וב-JavaScript, עד הסיבית האחרונה, בכל מחשב. ‏`Math.sin` או `Math.exp` של JavaScript עצמה
יכולות להיות שונות בספרה האחרונה מדפדפן לדפדפן, ולכן RuneScript לא משתמשת בהן: שני היעדים
מריצים את אותו אלגוריתם, זה של Go.

הבא: [מבנים](19-structs.md)

</div>
