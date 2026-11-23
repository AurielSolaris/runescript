<!-- source: README.md @ 298a2e7 -->

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

RuneScript רצה מכלי שורת פקודה עצמאי, מפיקה מודולי JavaScript ובונה אתרים
סטטיים מרובי דפים. הגרפיקה של raylib רצה בשולחן העבודה ובדפדפן באמצעות
WebAssembly.

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

## מצב: גרסה 0.7 בפיתוח (שלב מוקדם)

**מה עובד כבר עכשיו**, ומכוסה בבדיקות:

- חבילות; פונקציות המוצהרות בצורה `public|private|protected <type> <name>(…)`; רקורסיה;
  נקודת הכניסה `protected int main()`, שהתוצאה שלה היא קוד היציאה
- פונקציות מקוננות, אבות-טיפוס (prototypes), טיפוסי פונקציה (`int(int, int)`), פונקציות callback וביטויי למבדה (`(int a) => a * 2`)
- `int`/`sint`, `uint`, `bool`, `string`, `void`
- משתנים, `=`, `+= -= *= /= %=`, `++`/`--`; `int n;` מתחיל בערך האפס (אין `null`);
  הסתרה בבלוקים מקוננים, עם אזהרה
- אופרטורים אריתמטיים, אופרטורי השוואה ואופרטורים לוגיים
- `if`/`else if`/`else`, `while`, `for`, `break`, `continue`, `return`
- `trace(...)`
- ספרייה סטנדרטית: `use "std:string";` נותן את הפונקציות של `strings` מ-Go, למשל `string.toUpper(s)`
- תוכניות מרובות קבצים: `use "geom";` טוען את החבילה שב-`geom/*.rsm`, ו-
  `use "./lib/x.rsc";` טוען קובץ אחד; `public`/`protected`/`private` נאכפים
  ([מדריך](docs/guide/12-packages-and-files.md))
- `float` (כברירת מחדל float32, מתרחב בלי לאבד ערך), `float32`, `float64`;
  IEEE 754 עם `NaN` ו-`Infinity`, כמו ב-ActionScript
- המרות בדוקות בסגנון C: `(int) 64.3`, `(float) "64.7"`, `(uint) n`
  ([מדריך](docs/guide/13-floats-and-conversions.md))
- מספרים שלמים ברוחב מדויק `int8` … `uint64` (ערכים של 64 סיביות הם BigInt
  ב-JavaScript), ליטרלים הקסדצימליים/בינאריים/אוקטליים עם מפרידי `_`, אופרטורים על
  סיביות, ו-`std:math`, שתוצאותיה זהות סיבית בסיבית בשני היעדים ([מדריך](docs/guide/18-numbers.md))
- מבנים: טיפוסי ערך שנבנים עם `Point{x: 1, y: 2}`, מועתקים בהשמה ומושווים שדה אחר שדה
  ([מדריך](docs/guide/19-structs.md))
- טקסט ורשימות לסקריפטים: `+` ו-`$"{count} files"` למחרוזות, `char`, ו-`Array<T>`
  עם גישה לפי אינדקס שנבדקת, מערכים מקוננים, לולאות `for each (T x in a)`, חיתוך,
  חיפוש ומיון יציב ([מדריך](docs/guide/16-text-and-arrays.md)); `std:string`
  מכסה את חבילת `strings` של Go, כולל `split`, `fields`, `join` ו-callbacks
  ([מדריך](docs/guide/10-std-string.md))
- שגיאות בסגנון ActionScript: `throw new RangeError("…")`,
  `try`/`catch (Error e)`/`catch (e)`/`finally`; אפשר לתפוס גם שגיאות זמן ריצה
  ([מדריך](docs/guide/14-errors.md))
- בדיקת טיפוסים בזמן הידור עם הודעות שמצביעות על המיקום המדויק, כולל דחיית השמה
  של ערך עם סימן למשתנה בלי סימן
- גלישה של מספרים שלמים וחלוקה באפס מדווחות כשגיאות זמן ריצה, ולעולם אינן
  "מתגלגלות" בשקט; תוצאות ביניים של `int`/`uint` מתרחבות במדויק, כך ש-`big * big / big`
  עובד ([מדריך](docs/guide/04-operators.md))
- קבצים, קלט ופלט בשני היעדים: `std:io` (הדפסה בלי שורה חדשה, שורות שגיאה באדום ואזהרה
  בצהוב, קלט סטנדרטי), `std:process` (ארגומנטים, סביבה), `std:fs` (קבצי טקסט וקבצים בינאריים, תיקיות)
  ו-`std:path` (הכללים של path ב-Node.js, ממומשים בזמני הריצה של RuneScript)
  ([מדריך](docs/guide/20-files-and-input.md))
- ‏JSON עם `std:json`: ‏`json.parse`, קריאה בדוקה (`json.get`, `json.asInt`, …), בנייה
  ו-`json.stringify`, עם אותו מפענח ואותן הודעות שגיאה בשני היעדים ([מדריך](docs/guide/21-json.md))
- ‏CSV, מטריצות וגרסאות: ‏`std:csv` (קריאה וכתיבה לפי RFC 4180), ‏`std:matrix` (מכפלות,
  דטרמיננטות, הופכיות ומערכות לינאריות על `Array<Array<float64>>`, זהות עד לביט בשני היעדים)
  ו-`std:semver` (גרסאות SemVer 2.0.0, השוואה והתאמה כמו ב-npm) ([מדריך](docs/guide/22-csv-matrix-semver.md))
- הידור ל-JavaScript (`runescript build`): מודול ES2015 (בתוספת BigInt ו-`??`) שרץ בדפדפנים וב-Node.js
  ומתנהג בדיוק כמו המפרש בכל תוכנית שהמפרש יכול להריץ ([מדריך](docs/guide/11-javascript.md))
- דפי רשת עם `rune:dom`: רכיבים עם תבניות HTML ו-SVG שכל אלמנט ומאפיין בהן נבדק בזמן
  ההידור, טקסט `{{ }}` שעובר escape, `$if`/`$elseif`/`$else`, אירועים, ו-runes (אותות)
  שמעדכנים בדיוק את מה שקרא אותם, בלי DOM וירטואלי ([מדריך](docs/guide/15-web-pages.md))
- קריאה ל-JavaScript דרך הצהרות `extern`, כולל `rune:fs` שאינו תלוי בסביבת ההרצה.
  כל ערך שחוזר מ-JavaScript נבדק מול הטיפוס המוצהר שלו
- אזהרות על משתנים, פרמטרים והצהרות use שאינם בשימוש, על קוד שלא ירוץ לעולם, על
  תוצאות שנזרקות ועל הסתרה, עם `-Werror` ו-`--allow <code>`; במסוף שגיאות באדום ואזהרות
  בצהוב (`--color`)
- תמיכה בעורכים: שרת שפה, `runescript lsp`, עם שגיאות תוך כדי הקלדה, ריחוף, מעבר להגדרה,
  השלמה ומתאר, וחבילות ל-VS Code, ל-Vim ו-Neovim ול-Emacs ([מדריך](docs/guide/23-editors.md),
  [editors/](editors/README.md))
- טעינה מחדש חיה בזמן העבודה: `runescript dev app.rsc` מגישה תוכנית רשת וטוענת מחדש
  את הדפדפן בכל שמירה; `runescript run --watch` מריצה שוב תוכנית שורת פקודה

**על תאימות ל-JavaScript:** RuneScript עדיין אינה מבטיחה תאימות מלאה ל-JavaScript.
מה שהיא מבטיחה הוא קוד קפדני: TypeScript סומכת על ההצהרות שלה ומוחקת אותן, אבל
RuneScript בודקת כל ערך שנכנס מ-JavaScript, ואין בה `any`, המרות מרומזות או
המרות לא בדוקות.

**עדיין לא מומש:** מחלקות, חבילות `go:`,
`std:symrs`, אובייקטים, מערכים, מבנים ו-callbacks שחוצים את הגבול עם JavaScript,
`$foreach` ותגי רכיבים בתבניות, וחבילת הגרפיקה המקורית המתוכננת. ראו את
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
./runescript build --target web --extract-css -o site-dist examples/multi-page
./runescript build --target wasm -o game-web examples/native/keep-up.rsc
./runescript dev examples/fizzbuzz.rsc              # serves it, reloading on every save
./runescript run --watch examples/hello.rsc         # runs it again on every save
./runescript run examples/brainfuck.rsc examples/bf/hello.bf   # arguments follow the file
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
- הרצות ב-Node.js של תוכניות שקוראות ל-JavaScript דרך `extern`, כולל JavaScript
  שמחזירה ערכים שגויים בכוונה;
- בדיקה שה-JavaScript שנוצר לא משתמש בשום דבר חדש מ-ES2015 בתוספת `BigInt`, `??`
  ומודולים;
- בדיקות לשרת הפיתוח, למעקב אחרי קבצים ול-`run --watch`;
- בדיקה שכל תוכנית בתיעוד מתקמפלת, ושהתרגומים שומרים על כל בלוק קוד של הדפים
  באנגלית.

## מדדי ביצועים

ב-[`bench/`](../../bench/README.md) משווים את מנוע ה-Go של RuneScript ואת ה-JavaScript
שהוא מייצר ל-TypeScript 1.0 RC. לבדיקת נכונות מריצים `go run ./bench/rsbench verify`.
`go run ./bench/rsbench web` משווה את גודל הקבצים ואת זמן יצירת ה-DOM של אתר בן
שני עמודים ל-JavaScript שנכתב ידנית. התוצאות העדכניות מופיעות [ב-README הראשי](../../README.md#benchmarks).

## דוגמאות

ב-[`examples/`](../../examples/README.md) יש תוכניות לכל תכונה: היסודות,
`std:string`, פונקציות callback וביטויי למבדה, מונה מספרים ראשוניים,
[מפרש Brainfuck](../../examples/brainfuck.rsc) עם 32,768 תאים, ושתי
תוכניות שהמהדר דוחה בכוונה. הפלט של כל דוגמה נבדק בחבילת הבדיקות.

## תיעוד

- [מדריך](docs/guide/README.md): סיור בכל התכונות, עם דוגמאות (כולן מתקמפלות בבדיקות)
- [מפרט השפה](../../docs/LANGUAGE_SPEC.md) (באנגלית): הדקדוק והכללים, ומצב המימוש של כל כלל
- [הספרייה הסטנדרטית: std:string](docs/guide/10-std-string.md)
- [טקסט ורשימות: +, $"…", char, Array<T>](docs/guide/16-text-and-arrays.md)
- [בלי null, בלי any, בלי גנריות](docs/guide/17-no-null-any-generics.md)
- [מספרים לעומק: רוחב, ליטרלים, סיביות, std:math](docs/guide/18-numbers.md)
- [מבנים](docs/guide/19-structs.md)
- [קבצים, קלט ופלט: std:io, std:process, std:fs, std:path](docs/guide/20-files-and-input.md)
- [JSON: ‏std:json](docs/guide/21-json.md)
- [CSV, מטריצות וגרסאות: ‏std:csv, ‏std:matrix, ‏std:semver](docs/guide/22-csv-matrix-semver.md)
- [עורכים: ‏VS Code, ‏Vim, ‏Neovim, ‏Emacs ו-runescript lsp](docs/guide/23-editors.md)
- [הידור ל-JavaScript](docs/guide/11-javascript.md)
- [מדריך הודעות האבחון](docs/guide/08-diagnostics.md)
- [למי שמגיע מ-ActionScript 3](docs/guide/09-coming-from-actionscript.md)
- [מפת הדרכים](../../docs/ROADMAP.md) (באנגלית)

## מבנה המאגר

</div>

```text
cmd/runescript/        CLI: run, check, build, dev, lsp, version
internal/
  srcenc/              Source encoding detection and UTF-8 conversion
  lexer/               Tokens and source positions
  parser/              RuneScript and template parser
  ast/                 Syntax tree
  loader/              Package and file imports
  checker/             Name resolution, static types, module metadata, warnings
  types/               Semantic types
  ir/                  Typed intermediate representation, lowering, verification
  runtime/             Go closure engine and reference tree walker
  jsgen/               JavaScript emission, runtime modules, source maps, CSS
  jshost/              Host-neutral JavaScript adapters
  webbuild/            Multi-page HTML and asset assembly
  wasmbuild/           Go and raylib WebAssembly bundle assembly
  dom/                 HTML and SVG metadata for template checks
  devserver/           Local HTTP server and browser live-reload client
  watch/               File polling for dev and run --watch
  lsp/                 Language server
  conv/                Checked conversions and number formatting
  fmath/               Floating-point math shared by the Go and JS targets
  diagnostics/         Error and warning codes, positions, and formatting
  term/                Terminal detection and color support
  driver/              Shared compiler and execution entry points
stdlib/                Go-native standard library, graphics, and widgets
examples/              CLI programs, web templates, multi-page site, and games
editors/               VS Code, Vim and Neovim, and Emacs integrations
tests/                 End-to-end and documentation tests with testdata fixtures
bench/                 rsbench runner, TypeScript suite, profiling, and web benchmark
docs/                  Language specification, guide, and roadmap
localization/          Bengali and Hebrew README and guide translations
assets/                RuneScript SVGs and generated PNGs
scripts/               Logo rendering and translation status
install.sh, install.ps1  Per-user installers
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
