<!-- source: docs/guide/01-getting-started.md @ 920cb24 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/01-getting-started.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 1. מתחילים

## התקנה

RuneScript נבנית מקוד המקור, ולכן צריך **Go 1.25 ומעלה**
([go.dev/dl](https://go.dev/dl/)). תוכניות ההתקנה בודקות זאת קודם כול: הן נעצרות
עם הודעה ברורה אם Go חסר, ישן מדי או פגום.

**Linux ו-macOS:**

</div>

```sh
curl -fsSL https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.sh | sh
```

<div dir="rtl">

**Windows (PowerShell):**

</div>

```powershell
irm https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.ps1 | iex
```

<div dir="rtl">

תוכנית ההתקנה:

1. בודקת את Go (ב-PATH, במקומות ההתקנה הרגילים שלו, או ב-`RUNESCRIPT_GO`);
2. בונה בעזרת ה-Go הזה את ה-CLI של `runescript`, שהוא גם המהדר וגם סביבת הריצה
   (בנייה ב-Go טהור: לא צריך מהדר C);
3. מתקינה אותו ב-`~/.runescript/bin` (ב-Windows:
   `%LOCALAPPDATA%\Programs\RuneScript\bin`), בלי הרשאות מנהל;
4. מוסיפה את התיקייה הזו ל-PATH ומגדירה את `RUNESCRIPT_HOME`: בקובצי האתחול של
   ה-shell (`~/.profile`, `~/.bashrc`, `~/.zshrc`, fish), או בסביבת המשתמש
   ב-Windows;
5. בודקת את התוצאה על ידי הרצת תוכנית קטנה.

אחר כך פתחו מסוף חדש, ואז:

</div>

```text
$ runescript version
runescript 0.6.0-dev (commit 0fdbc61, go1.25.11 linux/amd64)
```

<div dir="rtl">

מתוך עותק מקומי (clone) של המאגר, הריצו שם את תוכנית ההתקנה כדי לבנות את קוד
המקור הזה במקום את הגרסה האחרונה: `sh install.sh`, או
`powershell -ExecutionPolicy Bypass -File install.ps1`.

| אפשרות (`install.sh` / `install.ps1`) | מה היא עושה |
|---|---|
| `--prefix DIR` / `-InstallDir DIR` | מתקינה ב-`DIR/bin` (אפשר גם דרך `RUNESCRIPT_HOME`) |
| `--version REF` / `-Version REF` | בונה תגית, ענף או קומיט מ-GitHub (אפשר גם דרך `RUNESCRIPT_VERSION`) |
| `--remote` / `-Remote` | בונה מ-GitHub גם מתוך עותק מקומי |
| `--no-modify-path` / `-NoModifyPath` | לא נוגעת ב-PATH ובקובצי האתחול |
| `--uninstall` / `-Uninstall` | מסירה את RuneScript ואת הרשומה שלה ב-PATH |

Node.js אינו חובה. הוא נחוץ רק כדי להריץ את ה-JavaScript ש-`runescript build`
כותב.

### בלי תוכנית ההתקנה

</div>

```sh
go build -o runescript ./cmd/runescript     # runescript.exe on Windows
go run ./cmd/runescript run examples/hello.rsc   # or run without installing
```

<div dir="rtl">

## התוכנית הראשונה שלכם

שמרו את זה בשם `hello.rsc`:

</div>

```actionscript
package main;

protected int main() {
    trace("Hello, RuneScript!");
    return 0;
}
```

<div dir="rtl">

הריצו:

</div>

```text
$ runescript run hello.rsc
Hello, RuneScript!
```

<div dir="rtl">

> קובצי RuneScript משתמשים בסיומת `.rsc`, אבל ה-CLI מקבל כל שם קובץ. הקבצים
> מקודדים ב-UTF-8. קובץ UTF-16 או UTF-32 (כפי שחלק מעורכי הטקסט ב-Windows שומרים)
> מומר, עם הודעה על כך; כל קובץ אחר שאינו UTF-8 נדחה עם `ASX-LEX-007`.

## פקודות

| פקודה | מה היא עושה |
|---|---|
| `runescript run <file> [args...]` | בודקת את התוכנית ואז מריצה אותה; המילים שאחרי הקובץ הן הארגומנטים שלה (`process.args()`, [פרק 20](20-files-and-input.md)) |
| `runescript check <file>` | מדווחת על שגיאות הידור בלי להריץ |
| `runescript run --watch <file>` | מריצה את התוכנית, ושוב בכל פעם שאחד הקבצים שלה משתנה (Ctrl+C עוצר) |
| `runescript build [-o out.js] <file>` | מהדרת למודול JavaScript ([פרק 11](11-javascript.md)) |
| `runescript dev [--port 5173] [--open] <file>` | מגישה את התוכנית ב-localhost וטוענת מחדש את הדפדפן בכל שינוי ([פרק 11](11-javascript.md)) |
| `runescript version` | מדפיסה את הגרסה |

`run`, `check`, `build` ו-`dev` מקבלות גם `-Werror` ו-`--allow <codes>` עבור אזהרות
(ראו [פרק 8](08-diagnostics.md)), וגם `--color=auto|always|never`: במסוף, שגיאות
מוצגות באדום ואזהרות בצהוב, אלא אם משתנה הסביבה `NO_COLOR` מוגדר. הדגלים באים לפני
הקובץ; כל מה שאחריו עובר לתוכנית.

## קוד היציאה

| מצב | קוד היציאה |
|---|---|
| התוכנית רצה | ה-`int` ש-`main` מחזירה |
| שגיאת הידור | 1 |
| שגיאת זמן ריצה (גלישה, חלוקה באפס, …) | 1 |
| שורת פקודה שגויה או קובץ שאי אפשר לקרוא | 2 |

## הרצת הבדיקות

</div>

```sh
go test ./...
```

<div dir="rtl">

הבא: [מבנה התוכנית](02-program-structure.md)

</div>
