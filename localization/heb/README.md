<!-- source: README.md @ 655da49 -->

<div dir="rtl">

> זהו תרגום של [קובץ ה-README באנגלית](../../README.md). אם יש הבדל בין השניים, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר; תיקונים יתקבלו בברכה.

</div>

<p align="center">
  <img src="../../assets/runescript.png" alt="הלוגו של RuneScript: פתית שלג בסגנון רוני עם קווי ברק" width="160">
</p>

<h1 align="center">RuneScript</h1>

<p align="center" dir="rtl">שפת תכנות כללית עם טיפוסים סטטיים, בהשראת ActionScript, הממומשת ב-Go.</p>

<p align="center"><a href="../../README.md" lang="en">English</a> · <a href="../bn/README.md" lang="bn">বাংলা</a> (בנגלית) · <strong>עברית</strong></p>

---

<div dir="rtl">

RuneScript שומרת על הזהות של ActionScript (`package`, `trace`, `int`/`uint`,
`public`/`protected`/`private`) ומשנה את התחביר לסגנון של משפחת C: הטיפוס
נכתב לפני השם, כל פונקציה מתחילה במשנה גישה (`public int add(int a, int b)`)
במקום במילת המפתח `function`, ונקודה-פסיק היא חובה. בדיקת הטיפוסים קפדנית:
שום דבר אינו `any` באופן מובלע, ומספרים שלמים עם סימן ובלי סימן לעולם אינם
מתערבבים בשקט.

השפה רצה כתוכנית עצמאית: בלי דפדפן, בלי HTML ובלי Flash Player. היא יכולה גם
להתקמפל למודול JavaScript לשימוש באינטרנט, עם אותה הקפדה על טיפוסים ואותן
שגיאות זמן ריצה. היא **אינה** תואמת ברמת קוד המקור ל-ActionScript 3.

</div>

```actionscript
package main;

public void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

protected int main() {
    sayHi("Hello world", 3);
    return 0;
}
```

<div dir="rtl">

## מצב: גרסה 0.3 בפיתוח (שלב מוקדם)

**מה עובד כבר עכשיו**, ומכוסה בבדיקות:

- חבילות; פונקציות המוצהרות בצורה `public|private|protected <type> <name>(…)`; רקורסיה
- פונקציות מקוננות, אבות-טיפוס (prototypes), טיפוסי פונקציה (`int(int, int)`), פונקציות callback וביטויי למבדה (`(int a) => a * 2`)
- `int`/`sint`, `uint`, `bool`, `string`, `void`
- משתנים, `=`, `+= -= *= /= %=`, `++`/`--`
- אופרטורים אריתמטיים, אופרטורי השוואה ואופרטורים לוגיים
- `if`/`else if`/`else`, `while`, `for`, `break`, `continue`, `return`
- `trace(...)`
- ספרייה סטנדרטית: `use "std:string";` נותן את הפונקציות של `strings` מ-Go, למשל `string.toUpper(s)`
- תוכניות מרובות קבצים: `use "geom";` טוען את החבילה שב-`geom/*.rsm`, ו-
  `use "./lib/x.rsc";` טוען קובץ אחד; `public`/`protected`/`private` נאכפים
  ([מדריך](../../docs/guide/12-packages-and-files.md))
- `float` (כברירת מחדל float32, מתרחב בלי לאבד ערך), `float32`, `float64`;
  IEEE 754 עם `NaN` ו-`Infinity`, כמו ב-ActionScript
- המרות בדוקות בסגנון C: `(int) 64.3`, `(float) "64.7"`, `(uint) n`
  ([מדריך](../../docs/guide/13-floats-and-conversions.md))
- שגיאות בסגנון ActionScript: `throw new RangeError("…")`,
  `try`/`catch (Error e)`/`catch (e)`/`finally`; אפשר לתפוס גם שגיאות זמן ריצה
  ([מדריך](../../docs/guide/14-errors.md))
- בדיקת טיפוסים בזמן הידור עם הודעות שמצביעות על המיקום המדויק, כולל דחיית השמה
  של ערך עם סימן למשתנה בלי סימן
- גלישה של מספרים שלמים וחלוקה באפס מדווחות כשגיאות זמן ריצה, ולעולם אינן
  "מתגלגלות" בשקט
- הידור ל-JavaScript (`runescript build`): מודול ES2022 שרץ בדפדפנים וב-Node.js
  ומתנהג בדיוק כמו המפרש ([מדריך](../../docs/guide/11-javascript.md))

**עדיין לא מומש:** מחלקות, מבנים (structs), מערכים, `char`, שרשור מחרוזות,
חבילות `go:`, `std:math`, `std:symrs`, וחבילת הגרפיקה המקורית המתוכננת. ראו את
[מפת הדרכים](../../docs/ROADMAP.md) (באנגלית).

## התקנה

נדרש Go 1.25 ומעלה ([go.dev/dl](https://go.dev/dl/)); תוכנית ההתקנה בודקת זאת,
בונה את ה-CLI של `runescript` ומוסיפה אותו ל-PATH שלכם.

</div>

```sh
# Linux, macOS
curl -fsSL https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.sh | sh
```

```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.ps1 | iex
```

<div dir="rtl">

אפשרויות, הסרת התקנה והתקנה מעותק מקומי של המאגר מתוארות
[במדריך "מתחילים"](docs/guide/01-getting-started.md).

## בנייה והרצה מקוד המקור

</div>

```sh
go build -o runescript ./cmd/runescript
./runescript run examples/hello.rsc
./runescript check examples/signedness_error.rsc   # shows a compile-time error
./runescript build examples/fizzbuzz.rsc            # writes examples/fizzbuzz.js
node examples/fizzbuzz.js                         # runs it (Node.js 24+)
go test ./...
```

<div dir="rtl">

> תוכניות RuneScript משתמשות בסיומת `.rsc`, ותיקיות של חבילות מכילות קבצי `.rsm`.
> כמו ב-Perl, הסיומת היא מוסכמה בלבד: ה-CLI והוראות `use` של קבצים מקבלים כל שם
> קובץ, כל עוד הקובץ מכיל RuneScript.

## בדיקות

</div>

```sh
go test ./...                                  # unit, golden, docs, and CLI tests
go test ./tests -run '^$' -fuzz=FuzzCheck      # fuzz the parser and checker
go test ./tests -run '^$' -fuzz=FuzzRun        # fuzz the interpreter
```

<div dir="rtl">

חבילת הבדיקות כוללת:
- בדיקות יחידה לכל שלב במהדר;
- בדיקות "זהב" שמריצות כל דוגמה ומשוות את הפלט שלה;
- תוכניות שגיאה שהודעות האבחון הצפויות שלהן כתובות בתוכן הקובץ בצורה
  `// expect: CODE line:col`;
- תוכניות של שגיאות זמן ריצה;
- בדיקות השוואתיות שמריצות כל תוכנית גם במפרש וגם ב-Node.js ודורשות תוצאות זהות
  (מדולגות אם `node` אינו מותקן);
- בדיקה שכל תוכנית בתיעוד מתקמפלת.

## דוגמאות

ב-[`examples/`](../../examples/README.md) יש תוכניות לכל תכונה: היסודות,
`std:string`, פונקציות callback וביטויי למבדה, מונה מספרים ראשוניים, ושתי
תוכניות שהמהדר דוחה בכוונה. הפלט של כל דוגמה נבדק בחבילת הבדיקות.

## תיעוד

- [מדריך](docs/guide/README.md): סיור בכל התכונות, עם דוגמאות (כולן מתקמפלות בבדיקות)
- [מפרט השפה](../../docs/LANGUAGE_SPEC.md) (באנגלית): הדקדוק והכללים, ומצב המימוש של כל כלל
- [הספרייה הסטנדרטית: std:string](../../docs/guide/10-std-string.md)
- [הידור ל-JavaScript](../../docs/guide/11-javascript.md)
- [מדריך הודעות האבחון](../../docs/guide/08-diagnostics.md)
- [למי שמגיע מ-ActionScript 3](../../docs/guide/09-coming-from-actionscript.md)
- [מפת הדרכים](../../docs/ROADMAP.md) (באנגלית)

## מבנה המאגר

</div>

```text
cmd/runescript/        CLI (run, check, build, version)
internal/lexer/        tokens and source positions
internal/parser/       recursive-descent parser → AST
internal/ast/          syntax tree
internal/types/        semantic types
internal/checker/      name resolution and type checking
internal/runtime/      tree-walking interpreter
internal/ir/           typed intermediate representation and its verifier
internal/jsgen/        JavaScript code generator and its inlined runtime
internal/diagnostics/  error codes and formatting
internal/driver/       wires the stages together
stdlib/                standard library packages (std:string), implemented in Go
examples/              example programs
tests/                 end-to-end, golden, and documentation tests
assets/                logo: SVG source and PNGs generated from it
scripts/               render-logo.sh: SVG → PNG via headless Edge/Chrome
```

<div dir="rtl">

## תרגומים

האנגלית היא המקור: קובץ ה-README האנגלי והתיקייה [`docs/`](../../docs/) הם מקור
האמת. התרגומים נמצאים ב-[`localization/`](../README.md), תיקייה אחת לכל שפה:

| שפה | תיקייה | מה מתורגם |
|---|---|---|
| বাংলা (בנגלית) | [`localization/bn/`](../bn/README.md) | ה-README והמדריך |
| עברית | [`localization/heb/`](README.md) | ה-README והמדריך |

התרגומים לבנגלית ולעברית מתוחזקים בעזרת בינה מלאכותית ונבדקים ידנית בידי
המחבר. **עזרה תתקבל בברכה:** תיקונים מדוברי השפה, וכן שפות חדשות, רצויים מאוד.
איך התרגומים מאורגנים ונשמרים מעודכנים מול התיעוד האנגלי, ראו
[localization/README.md](../README.md) (באנגלית).

## לוגו

`assets/runescript.svg` (ל-README) ו-`assets/runescript-icon.svg` (לסמלים קטנים)
הם קובצי המקור שנכתבו ביד. קובצי ה-PNG נוצרים מהם:

</div>

```sh
scripts/render-logo.sh     # needs Microsoft Edge or Chrome; override with BROWSER=...
```

<div dir="rtl">

## רישיון

RuneScript מופצת תחת [רישיון MIT](../../LICENSE).

</div>
