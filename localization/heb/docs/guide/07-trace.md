<!-- source: docs/guide/07-trace.md @ 34d09ef -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/07-trace.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 7. פלט עם `trace`

`trace` היא הדרך המובנית להדפיס, שם שנשמר מ-ActionScript. היא מקבלת כל מספר של
ארגומנטים מכל טיפוס חוץ מ-`void`. היא מדפיסה אותם לפלט הסטנדרטי, מופרדים ברווח
יחיד, ואחריהם שורה חדשה.

</div>

```actionscript
package main;

public void main() {
    trace("Hello");                 // Hello
    trace("x =", 42);               // x = 42
    trace(1, -2, true, "three");    // 1 -2 true three
    trace();                        // (empty line)
}
```

<div dir="rtl">

| ערך | מודפס כ- |
|---|---|
| `int`, `sint`, `uint` | מספר עשרוני, למשל `-42` |
| `float`, `float32`, `float64` | כמו ש-ActionScript מדפיסה מספרים: `64`, `0.5`, `1e+21`, `NaN` ([פרק 13](13-floats-and-conversions.md)) |
| `bool` | `true` / `false` |
| `string` | הטקסט עצמו, בלי מירכאות |
| `char` | התו עצמו, למשל `é` |
| שגיאה | השם וההודעה שלה, למשל `RangeError: too big` ([פרק 14](14-errors.md)) |

כדי להדפיס מחרוזת אחת שבנויה מכמה ערכים, חברו אותם עם `+` או השתמשו במחרוזת משובצת
כמו `$"{name} is {age}"` ([פרק 16](16-text-and-arrays.md)). אי אפשר להעביר מערכים
ל-trace; חברו אותם קודם עם `string.join`.

`trace` היא פונקציה מובנית שמורה. אי אפשר להצהיר על פונקציה או על משתנה בשם
`trace`.

הבא: [מדריך הודעות האבחון](08-diagnostics.md)

</div>
