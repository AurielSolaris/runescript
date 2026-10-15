<!-- source: docs/guide/11-javascript.md @ c31fd14 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/11-javascript.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 11. הידור ל-JavaScript

`runescript build` מהדר תוכנית למודול JavaScript שרץ בדפדפן או ב-Node.js. התוכנית
שומרת על הכללים של RuneScript: כל מה שבודק הטיפוסים דוחה עדיין נדחה, וכל מה שהיה
שגיאת זמן ריצה במפרש הוא אותה שגיאת זמן ריצה ב-JavaScript, עם אותו קוד, אותה הודעה
ואותו מיקום בקוד המקור.

</div>

```sh
runescript build examples/fizzbuzz.rsc            # writes examples/fizzbuzz.js
runescript build -o web/app.js examples/hello.rsc # choose the output file
node examples/fizzbuzz.js                         # runs it; main's result is the exit code
```

<div dir="rtl">

`--target js` הוא ברירת המחדל, וכרגע היעד היחיד. תוכנית של כמה קבצים (ראו
[חבילות וקבצים](12-packages-and-files.md)) הופכת למודול אחד, ושגיאות זמן ריצה
מציינות את הקובץ שבו קרו.

## שימוש במודול בדף אינטרנט

הפלט הוא מודול ES2022 אחד, עצמאי, בלי תלויות. הוא מייצא את `run`, שקוראת ל-`main`
ומחזירה את קוד היציאה שלה, ואת `RuneScriptError`, ש-`run` זורקת כשהתוכנית נעצרת עם
שגיאת זמן ריצה.

</div>

```html
<pre id="out"></pre>
<script type="module">
  import { run, RuneScriptError } from "./fizzbuzz.js";
  const out = document.getElementById("out");
  try {
    run({ trace: (line) => { out.textContent += line + "\n"; } });
  } catch (e) {
    if (!(e instanceof RuneScriptError)) throw e;
    out.textContent += e.message; // e.g. "fizzbuzz.rsc:6:15: error ASX-RUN-002: …"
  }
</script>
```

<div dir="rtl">

`options.trace` מקבלת כל שורה של פלט `trace`. בלעדיה, השורות הולכות ל-`console.log`.
ל-`RuneScriptError` יש גם `code`, `file`, `line`, `col` ו-`detail` (ההודעה בלי
המיקום).

## איך הקפדת הטיפוסים שורדת את התרגום

ב-JavaScript יש טיפוס מספר אחד ואין טיפוסים סטטיים, ולכן RuneScript בודקת כל מה
שהיא יכולה לפני יצירת הקוד, ומשאירה לקוד שנוצר לשמור על השאר:

| הכלל ב-RuneScript | JavaScript רגיל | הקוד שנוצר |
|---|---|---|
| לכל ערך יש הטיפוס שהוצהר | משתנים מחזיקים כל דבר | נבדק בזמן ההידור, ואז מאומת שוב על ה-IR המוקלד לפני שנכתב קוד כלשהו. לעולם לא נוצרת המרה לא בדוקה. |
| `int` ו-`uint` הם בני 32 סיביות; גלישה היא `ASX-RUN-002` | `2147483647 + 1` הוא `2147483648` | כל `+ - * / %` על מספרים שלמים עובר בדיקת טווח שעוצרת את התוכנית |
| `/` קוטע את השבר | `7 / 2` הוא `3.5` | חילוק קוטע |
| חלוקה באפס היא `ASX-RUN-003` | `Infinity` או `NaN` | נבדקת, למספרים שלמים ולמספרים עשרוניים |
| `float32` מעוגל ל-32 סיביות | טיפוס מספר אחד בן 64 סיביות | תוצאות `float32` עוברות דרך `Math.fround`, ומודפסות בספרות הקצרות ביותר של float32; `float` ו-`float64` הם מספרי JavaScript (IEEE, עם `NaN` ו-`Infinity`, כמו ב-AS3) |
| `(int) "12x"` זורקת `TypeError` | `parseInt("12x")` הוא `12` | אותם כללי ניתוח כמו במפרש |
| שגיאות זמן ריצה הן ערכי שגיאה שאפשר לתפוס | חריגה יכולה להיות כל ערך | השגיאות הן אובייקטי שגיאה של RuneScript; פסוקיות catch בודקות את הטיפוס שלהן; שגיאת עומק הקריאות עוברת את כל פסוקיות ה-catch |
| קריאות מקוננות עד עומק של 10,000 לכל היותר (`ASX-RUN-001`) | מחסנית שתלויה בסביבת ההרצה | נספרות. אם המחסנית של סביבת ההרצה נגמרת קודם, גם זו `ASX-RUN-001` |
| מיקומים ב-`std:string` הם היסטים בבתים של UTF-8 | מחרוזות הן UTF-16 | ההיסטים מומרים, ולכן `string.index("5€x", "x")` הוא `4` בכל יעד |
| המרת אותיות ורווחים לבנים ב-`std:string` פועלות כמו ב-Go | `"ß".toUpperCase()` הוא `"SS"` | טבלאות שנוצרו מהחבילה `unicode` של Go, ולכן התוצאות זהות |

למשל, אותה תוכנית נעצרת באותו מקום בשני היעדים:

</div>

```actionscript
package demo;

public int main() {
    int big = 2147483647;
    trace(big * big); // runtime error: ASX-RUN-002
    return 0;
}
```

<div dir="rtl">

ההודעה היא `integer overflow: 2147483647 * 2147483647 =
4611686014132420609 does not fit in 'int'`, והיא מחושבת במדויק אף שמספר JavaScript
אינו יכול להחזיק את המכפלה הזו.

חבילת הבדיקות מריצה כל דוגמה וכל תוכנית בדיקה גם במפרש וגם ב-Node.js, ודורשת פלט,
קודי יציאה ושגיאות זהים. היא גם משווה כל פונקציה של `std:string` לזו של Go על טקסט
אקראי שאינו ASCII.

## המגבלות של יעד JavaScript

- **עומק המחסנית.** לדפדפנים ול-Node.js נגמרת המחסנית לעיתים קרובות לפני 10,000
  קריאות מקוננות. כשזה קורה, התוכנית נעצרת עם `ASX-RUN-001` ("the JavaScript host
  ran out of stack") באחד ממקומות הקריאה ברקורסיה. זה לא תמיד המקום שהמפרש היה
  מדווח עליו.
- **מחרוזות ארוכות מאוד.** מנועי JavaScript מגבילים את אורך המחרוזת (בערך
  2<sup>29</sup>–2<sup>30</sup> יחידות UTF-16). חריגה מכך בתוך `string.repeat` או
  `string.replace` מדווחת כ-`ASX-RUN-004`.
- **עדיין אין עבודה משותפת עם JavaScript.** תוכנית אינה יכולה לקרוא לפונקציות
  JavaScript או ל-DOM, ו-JavaScript יכולה לקרוא רק ל-`run`. קריאה ל-JavaScript
  מתוכננת בצורת `use "js:<module>";` עם הצהרות מוקלדות, שהערכים שלהן נבדקים כשהם
  נכנסים ל-RuneScript.

הבא: [חבילות וקבצים](12-packages-and-files.md)

</div>
