<!-- source: docs/guide/13-floats-and-conversions.md @ c31fd14 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/13-floats-and-conversions.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 13. מספרים עשרוניים והמרות

## מספרים בנקודה צפה

ב-RuneScript יש שלושה טיפוסים של נקודה צפה:

| טיפוס | מה הוא |
|---|---|
| `float` | המספר העשרוני היומיומי. הוא מתחיל כמספר בן 32 סיביות ומתרחב ל-64 סיביות בכל פעם שערך צריך זאת, ולכן לעולם אינו מאבד ערך. |
| `float32` | בדיוק מספר IEEE בן 32 סיביות. כל תוצאה מעוגלת ל-32 סיביות. |
| `float64` | בדיוק מספר IEEE בן 64 סיביות, כמו `Number` של ActionScript. |

</div>

```actionscript
package main;

protected int main() {
    float a = 0.1;
    float b = 0.2;
    trace(a + b);          // 0.30000000000000004
    trace(7.0 / 2, 1e21);  // 3.5 1e+21

    float32 f = (float32) 0.1;
    trace(f * 3);          // 0.3, rounded to float32
    return 0;
}
```

<div dir="rtl">

מכיוון ש-`float` מתרחב במקום להתעגל, חישוב ב-`float` נותן תמיד את אותה תוצאה כמו
חשבון של 64 סיביות. השתמשו ב-`float32` כשאתם רוצים עיגול ל-32 סיביות.

ליטרלים נכתבים `64.3`, `1e10` או `2.5e-3`. מספרים מודפסים כמו ש-ActionScript
מדפיסה אותם: `64.0` מודפס כ-`64`, ומספרים גדולים מאוד או קטנים מאוד מוצגים בכתיב
מעריכי.

### `NaN` ו-`Infinity`

מספרים עשרוניים פועלים לפי IEEE 754, כמו `Number` של ActionScript. תוצאה גדולה מדי
לטיפוס שלה היא `Infinity`, ו-`0 / 0` הוא `NaN` ("not a number", כלומר "לא מספר").
בניגוד לגלישה במספרים שלמים, אלה לעולם אינם עוצרים את התוכנית.

</div>

```actionscript
package main;

protected int main() {
    float zero = 0;
    float nan = zero / zero;
    trace(1 / zero, -1 / zero, nan);           // Infinity -Infinity NaN
    trace(nan == nan, isNaN(nan));             // false true
    trace(isFinite(1e308 * 10), Infinity > 1); // false true
    return 0;
}
```

<div dir="rtl">

`NaN` אינו שווה לשום דבר, אפילו לא לעצמו, ולכן בודקים אותו עם `isNaN(x)`.
`isFinite(x)` הוא false עבור `NaN` ועבור שני האינסופים. `NaN` ו-`Infinity` הם
קבועים מובנים; משתנה באותו שם מסתיר אותם.

### אין המרות מרומזות

מספרים שלמים ומספרים עשרוניים אינם מתערבבים, וגם לא טיפוסים עשרוניים שונים. המירו
את אחד הצדדים במפורש:

</div>

```actionscript
package main;

protected int main() {
    int n = 3;
    float x = 1.5;
    trace(x * n);   // error: ASX-TYPE-012
    return 0;
}
```

```actionscript
package main;

protected int main() {
    int n = 3;
    float x = 1.5;
    trace(x * (float) n);   // 4.5
    return 0;
}
```

<div dir="rtl">

ליטרלים מספריים מתאימים את עצמם: ב-`x * 2`, ה-`2` הופך למספר עשרוני. עם זאת,
ליטרל חייב להיכנס בטיפוס שלו במדויק. `float32 f = 0.1;` היא שגיאה (`ASX-TYPE-018`),
כי ל-0.1 אין ערך float32 מדויק. כתבו `(float32) 0.1` כדי לעגל אותו בכוונה.

`%`, `++` ו-`--` עובדים רק על מספרים שלמים. עבור מספר עשרוני כתבו `x += 1;`.

## המרות

`(T) value` ממיר ערך ל-`T`, כמו cast ב-C. `T` יכול להיות `int`, `sint`, `uint`,
`float`, `float32`, `float64`, `string` או `bool`.

</div>

```actionscript
package main;

protected int main() {
    trace((int) 64.3, (int) -64.7);       // 64 -64: truncates toward zero
    trace((float) "64.7", (int) "-42");   // 64.7 -42: parses strings
    trace((string) 1.5, (bool) "true");   // 1.5 true
    int count = 7;
    trace((float) count / 2);             // 3.5: the cast binds tighter than /
    return 0;
}
```

<div dir="rtl">

כל המרה **נבדקת**. היא לעולם אינה מתגלגלת ואינה מנחשת. ערך שאינו נכנס, או מחרוזת
שאינה מספר, זורקים שגיאה:

| המרה | מתי היא נכשלת | שגיאה |
|---|---|---|
| למספר שלם | הערך מחוץ לטווח (למשל `(uint) -1.5`, `(int) 3e10`), `NaN`, או אינסופי | `RangeError` |
| ממחרוזת למספר | המחרוזת אינה מספר עשרוני פשוט (או `"NaN"`, `"Infinity"`, `"-Infinity"` עבור מספרים עשרוניים), למשל `"12x"`, `" 5"`, `"0x10"` | `TypeError` |
| ממחרוזת ל-`bool` | המחרוזת אינה בדיוק `"true"` או `"false"` | `TypeError` |

</div>

```actionscript
package main;

protected int main() {
    string input = "12x";
    trace((int) input);   // runtime error: ASX-RUN-006
    return 0;
}
```

<div dir="rtl">

כשהערך הוא קבוע, ההמרה מתבצעת בזמן ההידור, וכך גם השגיאה: `(uint) -2.5` היא
`ASX-TYPE-007`.

`bool` ומספרים אינם מומרים זה לזה. השוו במקום זאת (`n != 0`).

לכלל של עם סימן ובלי סימן מ[פרק 3](03-types-and-variables.md) יש עכשיו צורה
מפורשת:

</div>

```actionscript
package main;

protected int main() {
    int signedValue = 10;
    uint u = (uint) signedValue;   // checked: a negative value would throw
    trace(u);
    return 0;
}
```

<div dir="rtl">

הבא: [שגיאות](14-errors.md)

</div>
