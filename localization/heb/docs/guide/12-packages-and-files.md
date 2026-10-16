<!-- source: docs/guide/12-packages-and-files.md @ c31fd14 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/12-packages-and-files.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 12. חבילות וקבצים

אפשר לפצל תוכנית לכמה קבצים עם `use`. יש שלושה סוגים של `use`:

</div>

```text
use "std:string";          // a standard library package
use "geom";                // a RuneScript package: every .rsm file in geom/
use "./util/text.txt";     // one file, whatever its extension
```

<div dir="rtl">

מחרוזת שיש בה `/` היא נתיב לקובץ. כל דבר אחר הוא שם של חבילה, כמו `geom` או
`org.games.engine`. נתיבים ותיקיות של חבילות הם יחסיים **לתיקייה שממנה מריצים את
`runescript`**, ולא לקובץ שמכיל את ה-`use`.

## דוגמה

התוכנית הזו נמצאת בחבילת הבדיקות ב-`tests/testdata/multi/ok_packages`. הריצו אותה
מהתיקייה הזו עם `runescript run main.rsc`.

</div>

```text
ok_packages/
├── main.rsc            package app;   the entry file
├── helpers.rsc         package app;   another file of the same package
├── geom/
│   ├── area.rsm        package geom;
│   └── perimeter.rsm   package geom;
└── util/
    └── text.txt        package text;  any extension works
```

<div dir="rtl">

`main.rsc`:

</div>

```text
package app;

use "geom";
use "./util/text.txt" as t;
use "./helpers.rsc";

public int main() {
    trace(geom.area(3, 4), geom.perimeter(3, 4));
    trace(t.shout("hi"));
    trace(twice(21));
    return 0;
}
```

<div dir="rtl">

`geom/perimeter.rsm`:

</div>

```text
package geom;

use "./util/text.txt";

public int perimeter(int w, int h) {
    trace(text.shout("perimeter"));
    return scale(2 * (w + h));
}

// Protected: shared by both files of package geom.
protected int scale(int n) {
    return n * one();
}

// Private to this file.
private int one() {
    return 1;
}
```

<div dir="rtl">

היא מדפיסה:

</div>

```text
PERIMETER
12 14
HI
42
```

<div dir="rtl">

## חבילות

חבילה היא כל הקבצים שמצהירים על אותו שם `package`.

- `use "geom";` טוען את קובצי ה-`.rsm` שנמצאים ישירות בתיקייה `geom/`.
  `use "org.games.engine";` טוען אותם מ-`org/games/engine/`. קבצים אחרים בתיקייה
  מתעלמים מהם, וגם מתיקיות משנה. כל קובץ `.rsm` חייב להצהיר על החבילה הזו
  (`ASX-USE-007`).
- קבצים של אותה חבילה רואים זה את פונקציות ה-`public` וה-`protected` של זה, בלי
  מזהה חבילה ובלי לכתוב אבות-טיפוס. בדוגמה, `area.rsm` קורא ל-`scale`, שמוגדרת
  ב-`perimeter.rsm`.
- קובץ שעושים לו `use` לפי נתיב מצטרף לחבילה שהוא מצהיר עליה. בדוגמה,
  `helpers.rsc` מצהיר על `package app;`, ולכן קוראים ל-`twice` שלו ישירות
  מ-`main.rsc`.

כל קובץ נטען פעם אחת, לא משנה כמה קבצים משתמשים בו. גם `main.rsc` וגם
`perimeter.rsm` משתמשים ב-`util/text.txt`, והוא נטען פעם אחת.

## מזהי חבילה

קוראים לפונקציות של חבילה אחרת דרך **מזהה חבילה** (qualifier). כברירת מחדל זה
החלק האחרון בשם החבילה: `geom` עבור `geom`, ו-`engine` עבור `org.games.engine`.
עבור קובץ, זו החבילה שהקובץ מצהיר עליה: `text` עבור `util/text.txt`. כדי לבחור שם
אחר, השתמשו ב-`as`:

</div>

```text
use "./util/text.txt" as t;     // t.shout("hi")
use "org.games.engine" as eng;  // eng.start()
```

<div dir="rtl">

מזהה חבילה שייך לקובץ שבו הוא הוצהר. כל קובץ שקורא ל-`geom.area` צריך `use "geom";`
משלו.

אם לשני שימושים באותו קובץ יהיה אותו מזהה, זו `ASX-USE-008`. תנו לאחד מהם שם אחר
עם `as`.

פונקציה עם מזהה חבילה היא גם ערך: `int(int, int) f = geom.area;` מעביר את
`geom.area` בלי לקרוא לה, כמו ש-`dom.renderApp(app.App, "#app")` עושה עם רכיב
([פרק 15](15-web-pages.md)). לפונקציות של הספרייה הסטנדרטית אפשר בינתיים רק לקרוא:
עטפו אותן בביטוי למבדה, למשל `(string s) => string.toUpper(s)`.

## נראות

עכשיו נאכף משנה הגישה של כל פונקציה:

| משנה | מי יכול לקרוא לה |
|---|---|
| `public` | כל קובץ בכל חבילה (חבילות אחרות משתמשות במזהה חבילה) |
| `protected` | כל קובץ בחבילה שלה |
| `private` | רק הקובץ שלה |

קריאה לפונקציה שמשנה הגישה שלה אינו מתיר זאת היא `ASX-NAME-011`, וההודעה אומרת
איפה הפונקציה נמצאת. לשני קבצים יכולה להיות לכל אחד פונקציית `private` משלו באותו
שם. כל שם אחר שחוזר ברמה העליונה בחבילה הוא `ASX-NAME-002`.

## נקודת הכניסה

לתוכנית יש בדיוק `main` אחת, והיא יכולה להיות בכל אחד מהקבצים שלה. הקובץ שמעבירים
ל-`runescript run` אינו חייב להכיל אותה. יותר מ-`main` אחת היא `ASX-ENTRY-003`.

## מעגלים

חבילות אינן יכולות להשתמש זו בזו במעגל. אם `a` משתמשת ב-`b` ו-`b` משתמשת ב-`a`,
ישירות או דרך חבילות אחרות, זו `ASX-USE-005`, וההודעה מציגה את המעגל:
`import cycle: a → b → a`. כדי לתקן, העבירו את הפונקציות המשותפות להן לחבילה
שלישית.

## שמות קבצים וסיומות

לפי המוסכמה, תוכניות הן קובצי `.rsc`, ותיקיות של חבילות מכילות קובצי `.rsm`. כמו
ב-Perl, הסיומת אינה נבדקת כשאתם נותנים שם של קובץ בעצמכם: `runescript run tool.txt`
ו-`use "./lib/x.inc";` עובדים שניהם, כל עוד הקובץ מכיל RuneScript. רק תיקיות של
חבילות מסתכלות על הסיומת, כך שקובצי README וקובצי נתונים שלידן לא נלקחים בחשבון.

הודעות אבחון מציינות את הקובץ שבו הן נמצאות, יחסית לתיקיית העבודה:

</div>

```text
calc/square.rsm:4:14: error ASX-RUN-002: integer overflow: 65536 * 65536 = 4294967296 does not fit in 'int'
```

<div dir="rtl">

## שגיאות

| קוד | סיבה |
|---|---|
| `ASX-USE-004` | קובץ חסר, או תיקיית חבילה בלי קובצי `.rsm` |
| `ASX-USE-005` | חבילות שמשתמשות זו בזו במעגל |
| `ASX-USE-006` | שם חבילה לא תקין, או נתיב שנכתב עם `\` |
| `ASX-USE-007` | קובץ `.rsm` שמצהיר על חבילה שונה מזו של התיקייה שלו |
| `ASX-USE-008` | שני שימושים עם אותו מזהה באותו קובץ |
| `ASX-NAME-011` | פונקציה `private` או `protected` שבשימוש מבחוץ |
| `ASX-ENTRY-003` | יותר מ-`main` אחת |

הבא: [מספרים עשרוניים והמרות](13-floats-and-conversions.md)

</div>
