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

**מה מובטח ומה לא.** RuneScript עדיין **אינה** מבטיחה תאימות מלאה ל-JavaScript.
המודול צריך מנוע שתומך ב-ES2015, ב-`BigInt`, ב-`??` ובמודולי ES. הוא מחליף עם
JavaScript רק מספרים, `bool`, `string` ושגיאות, דרך פונקציות extern (בהמשך). ו-`rune:fs`
הוא בינתיים ה-API היחיד שאינו תלוי בסביבת ההרצה. מה שהיא **כן** מבטיחה הוא קוד קפדני:
שום דבר לא נכנס לתוכנית RuneScript בלי בדיקה. TypeScript סומכת על ההצהרות שלה
ומוחקת אותן, ולכן `.d.ts` שאומר `number` יכול להחזיר מחרוזת בזמן ריצה. RuneScript
בודקת כל ערך שחוזר מ-JavaScript מול הטיפוס המוצהר שלו, ועוצרת עם `TypeError` כשהוא
לא תואם. אין `any`, אין המרה מרומזת ואין המרה לא בדוקה.

## שימוש במודול בדף אינטרנט

הפלט הוא מודול ES2015 אחד (בתוספת BigInt ו-`??`), עצמאי, בלי תלויות. הוא מייצא את `run`, שקוראת ל-`main`
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

תוכנית עם פונקציות extern מייבאת את המודולים שלהן, ולכן היא רצה רק היכן שהמודולים
האלה קיימים: מודול שמשתמש ב-`rune:fs` רץ ב-Node.js, לא בדפדפן.

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

## קריאה ל-JavaScript: פונקציות extern

הצהרת `extern` נותנת שם לפונקציה שמודול JavaScript מממש:

</div>

```actionscript
package notes;

extern "rune:fs" void writeFile(string path, string data);
extern "rune:fs" string readFile(string path);
extern "./clock.mjs" float64 now();

public int main() {
    writeFile("notes.txt", "ᚱᚢᚾᛖ");
    trace(readFile("notes.txt"), now() > 0);
    return 0;
}
```

```js
// clock.mjs, next to the compiled module
export const now = () => Date.now();
```

<div dir="rtl">

- הצהרת extern נכתבת ברמה העליונה של הקובץ. אין לה מגדיר גישה, והיא נראית רק
  באותו קובץ. שמות הפרמטרים אינם חובה.
- אפשר רק לקרוא לפונקציית extern. כדי להעביר אותה כערך, עטפו אותה בלמבדה:
  `(string p) => readFile(p)`.
- מחרוזת המודול היא **מודול שאינו תלוי בסביבת ההרצה** (`rune:…`, בהמשך), או כל
  דבר ש-`import` של JavaScript מקבל: קובץ כמו `"./clock.mjs"` (יחסית למודול
  המהודר), שם של חבילה, או מודול של סביבת הרצה מסוימת כמו `"node:fs"`. מודול כזה
  עובד, אבל הוא קושר את התוכנית לסביבה אחת, ולכן מדווחת עליו האזהרה `ASX-WARN-006`.
- רק יעד JavaScript מריץ פונקציות extern. `runescript check` מקבלת את התוכנית, אבל
  `runescript run` מדווחת `ASX-TARGET-002`. מתוכננת סביבת הרצה של JavaScript שתגיע
  יחד עם RuneScript, כדי שתוכניות כאלה ירוצו בלי Node.js.

### מודולים שאינם תלויים בסביבת ההרצה

מודול `rune:` הוא חלק מ-RuneScript, לא מסביבת הרצה אחת. הפונקציות שלו קבועות,
והצהרה חייבת להתאים לאחת מהן בדיוק (אחרת `ASX-TYPE-024`). המהדר ממפה כל פונקציה
ל-API של סביבת ההרצה עצמה; היום הסביבה הזו היא Node.js.

| פונקציה של `rune:fs` | מה היא עושה |
|---|---|
| `string readFile(string path)` | מחזירה את תוכן הקובץ, שנקרא כ-UTF-8 (בתים לא תקינים הופכים ל-U+FFFD) |
| `void writeFile(string path, string data)` | יוצרת קובץ או מחליפה אותו |
| `void appendFile(string path, string data)` | מוסיפה לסוף הקובץ, ויוצרת אותו אם צריך |
| `bool exists(string path)` | האם קובץ או תיקייה קיימים |
| `void remove(string path)` | מוחקת קובץ |

פונקציה שנכשלת, למשל `readFile` על קובץ חסר, זורקת `Error` (`ASX-RUN-008`).
`rune:process` הוא הבא בתור.

### מה חוצה את הגבול

ארגומנטים עוברים ל-JavaScript כמו שהם: מספרים, ערכים בוליאניים, מחרוזות, ושגיאות
כאובייקטי `RSError` (בהמשך). התוצאות **נבדקות** בדרך חזרה:

| התוצאה המוצהרת | מה JavaScript חייבת להחזיר |
|---|---|
| `int`, `uint` | מספר שלם בטווח של הטיפוס (`-0` הופך ל-`0`) |
| `float`, `float64` | כל מספר, כולל `NaN` ו-`Infinity` |
| `float32` | כל מספר. הוא מעוגל ל-float32 עם `Math.fround` |
| `bool` | `true` או `false` |
| `string` | מחרוזת בלי surrogates בודדים, כדי שתהיה UTF-8 תקין |
| `void` | `undefined` |
| `Error`, `RangeError`, `TypeError`, `ArgumentError` | `RSError` מאותו טיפוס (או מתת-טיפוס), או `Error` של JavaScript, שמומר: `RangeError` ו-`TypeError` שומרים על שמם, כל אחר הופך ל-`Error`, וה-`errorID` הוא `0` |

כל דבר אחר זורק `TypeError` עם הקוד `ASX-RUN-007` במקום הקריאה:

</div>

```actionscript
package demo;

extern "./lib.mjs" int half(int n);

public int main() {
    try {
        trace(half(3)); // JavaScript returns 1.5
    } catch (TypeError e) {
        trace(e.message); // half returned 1.5, but its result type is 'int'
    }
    return 0;
}
```

<div dir="rtl">

חריגה ש-JavaScript זורקת הופכת ל-`Error` עם הקוד `ASX-RUN-008` והודעה כמו
`readFile threw Error: ENOENT: …`. איש לא תופס אותה, אלא אם התוכנית שלכם תופסת.

### שגיאות בשני הכיוונים

לשגיאות של RuneScript ול-JavaScript יש מחלקה משותפת. המודול מייצא את `RSError`,
המחלקה של ערכי השגיאה של RuneScript. JavaScript מקבלת שגיאות כאובייקטי `RSError` עם
`name`, `message` ו-`errorID`, ויכולה ליצור אותן, להחזיר אותן ולזרוק אותן. `RSError`
ש-JavaScript זורקת מגיע ל-RuneScript בלי שינוי, ולכן הוא נתפס לפי הטיפוס שלו. עם
`extern "./even.mjs" int halveEven(int n);`, הקריאה `halveEven(3)` זורקת
`ArgumentError` ש-`catch (ArgumentError e)` תופס:

</div>

```js
// even.mjs: the compiled program is prog.js
import { RSError } from "./prog.js";

export const halveEven = (n) => {
  if (n % 2 !== 0) throw new RSError("ArgumentError", "n must be even", 1);
  return n / 2;
};
```

<div dir="rtl">

## המגבלות של יעד JavaScript

- **עומק המחסנית.** לדפדפנים ול-Node.js נגמרת המחסנית לעיתים קרובות לפני 10,000
  קריאות מקוננות. כשזה קורה, התוכנית נעצרת עם `ASX-RUN-001` ("the JavaScript host
  ran out of stack") באחד ממקומות הקריאה ברקורסיה. זה לא תמיד המקום שהמפרש היה
  מדווח עליו.
- **מחרוזות ארוכות מאוד.** מנועי JavaScript מגבילים את אורך המחרוזת (בערך
  2<sup>29</sup>–2<sup>30</sup> יחידות UTF-16). חריגה מכך בתוך `string.repeat` או
  `string.replace` מדווחת כ-`ASX-RUN-004`.
- **עבודה משותפת מוגבלת עם JavaScript.** פונקציות extern מחליפות רק מספרים, `bool`,
  `string` ושגיאות. אובייקטים, מערכים, פונקציות callback וה-DOM עדיין לא יכולים לחצות,
  ו-JavaScript יכולה לקרוא ל-RuneScript רק דרך `run`.
- **לא תאימות מלאה ל-JavaScript.** הפלט מכוון ל-ES2015 בתוספת `BigInt`, `??` ומודולי
  ES, והוא נבדק ב-Node.js. סביבות אחרות אמורות לעבוד, אבל הן לא נבדקות.

הבא: [חבילות וקבצים](12-packages-and-files.md)

</div>
