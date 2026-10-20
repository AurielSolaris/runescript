<!-- source: docs/guide/20-files-and-input.md @ 4541765 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/20-files-and-input.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 20. קבצים, קלט ופלט

ארבע חבילות סטנדרטיות מחברות תוכנית לעולם שסביבה:

| חבילה | בשביל |
|---|---|
| `std:io` | הדפסה בלי שורה חדשה, שורות שגיאה ואזהרה, וקריאת הקלט הסטנדרטי |
| `std:process` | ארגומנטים משורת הפקודה, משתני סביבה, תיקיית העבודה והפלטפורמה |
| `std:fs` | קריאה וכתיבה של קבצים, ויצירה, רישום ומחיקה של תיקיות |
| `std:path` | חיבור, פירוק ופתרון של נתיבים, לפי הכללים של המודול `path` ב-Node.js |

הן עובדות אותו דבר במפרש וב-JavaScript ([פרק 11](11-javascript.md)), עד לרמת הודעות
השגיאה שלהן.

## פלט

`trace` מדפיסה שורה, כמו קודם. `std:io` מוסיפה עוד שלוש דרכים לכתוב:

| פונקציה | כותבת |
|---|---|
| `io.print(string s)` | את `s` לפלט הסטנדרטי, בלי שורה חדשה |
| `io.error(string s)` | את `s` ושורה חדשה לשגיאה הסטנדרטית, באדום במסוף |
| `io.warn(string s)` | את `s` ושורה חדשה לשגיאה הסטנדרטית, בצהוב במסוף |

</div>

```actionscript
package main;

use "std:io";

protected int main() {
    io.print("Loading");
    for (int i = 0; i < 3; i++) {
        io.print(".");
    }
    trace(" done");                     // Loading... done
    io.warn("the cache is empty");      // on standard error
    io.error("could not reach the server");
    return 1;
}
```

<div dir="rtl">

`trace` ו-`io.print` חולקות את הפלט הסטנדרטי, ולכן הטקסט שלהן יוצא בסדר שבו נכתב.
צבעים משמשים רק כשהשגיאה הסטנדרטית היא מסוף ומשתנה הסביבה `NO_COLOR` אינו מוגדר;
גם `runescript run --color=never` מבטל אותם.

## קלט

| פונקציה | מחזירה |
|---|---|
| `io.readLine()` | את השורה הבאה בקלט הסטנדרטי, בלי `"\n"` או `"\r\n"`. בסוף הקלט היא מחזירה את מה שנשאר, וזה `""` אחרי שהכול נקרא |
| `io.readByte()` | את הבית הבא (0 עד 255), או `-1` בסוף הקלט |
| `io.readAll()` | את כל הקלט שנשאר |
| `io.atEnd()` | האם הקלט נגמר. היא ממתינה לקלט אם עוד אין |

טקסט נקרא כ-UTF-8; בתים שאינם UTF-8 תקין הופכים ל-U+FFFD. `io.readByte` נותנת את
הבתים הגולמיים.

</div>

```actionscript
package main;

use "std:io";
use "std:string";

// Numbers each line of its input:  runescript run number.rsc < notes.txt
protected int main() {
    int n = 0;
    while (!io.atEnd()) {
        n++;
        string line = io.readLine();
        trace(string.repeat(" ", 4 - string.len((string) n)) + (string) n, line);
    }
    return 0;
}
```

<div dir="rtl">

## ארגומנטים והסביבה

| פונקציה | מחזירה |
|---|---|
| `process.args()` | את הארגומנטים משורת הפקודה, כ-`Array<string>` חדש |
| `process.env(string name)` | משתנה סביבה, או `""` אם אינו מוגדר |
| `process.hasEnv(string name)` | האם משתנה סביבה מוגדר |
| `process.cwd()` | את תיקיית העבודה |
| `process.platform()` | `"windows"`, `"linux"`, `"macos"`, או שם של מערכת אחרת; `"browser"` בדף רשת |

הארגומנטים הם המילים שאחרי קובץ התוכנית: `runescript run greet.rsc Ada Grace` (או
`node greet.js Ada Grace`) מעבירה `["Ada", "Grace"]`. דגלים של `runescript` עצמה
באים לפני הקובץ.

</div>

```actionscript
package main;

use "std:process";

protected int main() {
    Array<string> names = process.args();
    if (names.length == 0) {
        trace("usage: greet <name>...");
        return 2;
    }
    for (int i = 0; i < names.length; i++) {
        trace($"Hello, {names[i]}!");
    }
    return 0;
}
```

<div dir="rtl">

## קבצים ותיקיות

נתיבים הם יחסיים לתיקיית העבודה. `readFile`, `writeFile` ו-`appendFile` עובדות עם
טקסט UTF-8; לנתונים אחרים, ראו [קבצים בינאריים](#קבצים-בינאריים) בהמשך.

| פונקציה | מה היא עושה |
|---|---|
| `fs.readFile(string path)` | מחזירה את הטקסט של הקובץ |
| `fs.writeFile(string path, string text)` | יוצרת קובץ או מחליפה אותו |
| `fs.appendFile(string path, string text)` | מוסיפה לסוף קובץ, ויוצרת אותו אם צריך |
| `fs.exists(string path)`, `fs.isFile(…)`, `fs.isDir(…)` | האם יש שם משהו, קובץ או תיקייה |
| `fs.size(string path)` | את גודל הקובץ בבתים, כ-`int64` |
| `fs.copyFile(string from, string to)` | מעתיקה קובץ ומחליפה את `to` |
| `fs.rename(string from, string to)` | מעבירה קובץ או תיקייה, או משנה את שמם |
| `fs.remove(string path)` | מוחקת קובץ או תיקייה ריקה |
| `fs.removeAll(string path)` | מוחקת תיקייה וכל מה שבתוכה; לא עושה דבר אם אינה קיימת |
| `fs.mkdir(string path)` | יוצרת תיקייה, שתיקיית האב שלה חייבת להתקיים |
| `fs.mkdirAll(string path)` | יוצרת תיקייה ואת כל תיקיות האב החסרות; לא עושה דבר אם כבר קיימת |
| `fs.readDir(string path)` | את השמות שבתיקייה, ממוינים |
| `fs.walk(string path)` | כל קובץ ותיקייה מתחת ל-`path`, כנתיבים יחסיים אליו, וכל תיקייה ואחריה התוכן שלה |
| `fs.tempDir()` | את התיקייה של המערכת לקבצים זמניים |

כישלון זורק `Error` שההודעה שלו אומרת מה השתבש, באותן מילים בכל מקום:
`fs.readFile: cannot read notes.txt: no such file or directory`. הסיבות הן
"no such file or directory", "already exists", "is a directory", "not a directory",
"directory not empty" ו-"permission denied".

</div>

```actionscript
package main;

use "std:fs";
use "std:path";

protected int main() {
    string dir = path.join(fs.tempDir(), "runescript-guide");
    fs.removeAll(dir);
    fs.mkdirAll(path.join(dir, "notes"));
    fs.writeFile(path.join(dir, "notes/todo.txt"), "buy milk\n");
    fs.appendFile(path.join(dir, "notes/todo.txt"), "write docs\n");
    trace(fs.readFile(path.join(dir, "notes/todo.txt")));
    trace(fs.size(path.join(dir, "notes/todo.txt")));       // 20
    Array<string> all = fs.walk(dir);
    trace(all.length);                                     // 2: notes, and the file in it
    try {
        trace(fs.readFile(path.join(dir, "missing.txt")));
    } catch (Error e) {
        trace(e.name);                                     // Error
    }
    fs.removeAll(dir);
    trace(fs.exists(dir));                                 // false
    return 0;
}
```

<div dir="rtl">

### קבצים בינאריים

תמונות, ארכיונים ונתונים אחרים שאינם טקסט נקראים ונכתבים כבתים, `Array<uint8>`.
הבתים עוברים אל הקובץ וממנו בלי שינוי.

| פונקציה | מה היא עושה |
|---|---|
| `fs.readBytes(string path)` | מחזירה את הבתים של הקובץ, `Array<uint8>` |
| `fs.writeBytes(string path, Array<uint8> bytes)` | יוצרת קובץ או מחליפה אותו |
| `fs.appendBytes(string path, Array<uint8> bytes)` | מוסיפה לסוף קובץ, ויוצרת אותו אם צריך |

הן נכשלות כמו גרסאות הטקסט שלהן, עם אותן הודעות. `string.bytes(s)` ו-`string.fromBytes(b)`
([פרק 10](10-std-string.md#בתים)) ממירות בין טקסט לבין הבתים שלו ב-UTF-8.

</div>

```actionscript
package main;

use "std:fs";
use "std:path";
use "std:string";

protected int main() {
    string file = path.join(fs.tempDir(), "runescript-guide.bin");
    Array<uint8> header = [0x89, 0x50, 0x4E, 0x47];   // the start of a PNG file
    fs.writeBytes(file, header);
    fs.appendBytes(file, string.bytes("hi"));
    Array<uint8> data = fs.readBytes(file);
    trace(data.length, data[0], data[1]);             // 6 137 80
    trace(fs.size(file));                             // 6
    bool png = data.length >= 4 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47;
    trace(png);                                       // true
    trace(string.fromBytes(data.slice(4, 6)));        // hi
    fs.remove(file);
    return 0;
}
```

<div dir="rtl">

## נתיבים

ב-`std:path` יש את הפונקציות של המודול `path` ב-Node.js, כתובות עבור RuneScript (הן
אינן קוראות ל-Node.js, ולכן עובדות גם במפרש):

| פונקציה | מחזירה |
|---|---|
| `path.join(a, b)`, `path.joinAll(parts)` | את החלקים מחוברים במפריד, ואחר כך מנורמלים |
| `path.resolve(p)`, `path.resolveAll(parts)` | נתיב מוחלט, החל מתיקיית העבודה |
| `path.normalize(p)` | את `p` בלי `.`, `..` או מפרידים כפולים |
| `path.relative(from, to)` | את הנתיב מ-`from` אל `to` |
| `path.dirname(p)`, `path.basename(p)`, `path.basename(p, suffix)`, `path.extname(p)` | את התיקייה, את החלק האחרון (אם רוצים, בלי `suffix`) ואת הסיומת |
| `path.isAbsolute(p)`, `path.root(p)` | האם `p` מוחלט, ואת השורש שלו (`"/"`, `"C:\"`, …) |
| `path.sep()`, `path.delimiter()` | `"/"` ו-`":"`, או `"\"` ו-`";"` ב-Windows |

הן פועלות לפי הכללים של המערכת שהתוכנית רצה בה: נתיבי Windows ב-Windows (`C:\`, גם
`\` וגם `/` כמפרידים, `\\server\share`), ונתיבי POSIX בכל מקום אחר. לנתיבים שצריכים
להיראות אותו דבר בכל מקום, כל פונקציה קיימת גם כ-`path.posixJoin`, `path.posixResolve`,
… וכ-`path.win32Join`, `path.win32Resolve`, …:

</div>

```actionscript
package main;

use "std:path";

protected int main() {
    trace(path.posixJoin("/home/ada", "../grace/notes.txt"));   // /home/grace/notes.txt
    trace(path.win32Join("C:\\Users", "ada\\..\\grace"));        // C:\Users\grace
    trace(path.posixBasename("/srv/www/index.html", ".html"));  // index
    trace(path.posixExtname("archive.tar.gz"));                 // .gz
    trace(path.posixRelative("/data/a/b", "/data/c"));          // ../../c
    trace(path.win32IsAbsolute("C:docs"));                      // false: relative to C:'s current directory
    return 0;
}
```

<div dir="rtl">

התוצאות זהות לאלה של המודול `path` ב-Node.js (הבדיקות משוות אלפי נתיבים אקראיים),
פרט לשני דברים: נתיבי Windows מושווים תוך התעלמות מאותיות גדולות וקטנות של ASCII בלבד,
ופתרון של נתיב בכונן אחר מתחיל מהשורש של אותו כונן.

## דוגמה: מפרש Brainfuck

[`examples/brainfuck.rsc`](../../../../examples/brainfuck.rsc) מריץ תוכניות
[Brainfuck](https://en.wikipedia.org/wiki/Brainfuck) על סרט של 32,768 תאים, ומשתמש בכל
ארבע החבילות:

</div>

```text
$ runescript run examples/brainfuck.rsc examples/bf/hello.bf
Hello World!
$ echo "Hello, World!" | runescript run examples/brainfuck.rsc examples/bf/rot13.bf
Uryyb, Jbeyq!
$ runescript run examples/brainfuck.rsc -e "+[>+]"
-e:1:3: the data pointer moved off the tape (to cell 32768; the cells are 0 to 32767)
```

<div dir="rtl">

הוא קורא את התוכנית עם `fs.readFile`, או לוקח אותה מ-`-e`, בעזרת `process.args()`. קודם
הוא מהדר אותה לרשימה של מבנים: רצפים של `+` ו-`-` מתקפלים לצעד אחד, וכל סוגריים מוצאים
את בן הזוג שלהם, כך שסוגריים בלי זוג מדווחים עם השורה והעמודה שלהם לפני שמשהו רץ. הסרט
הוא `Array<uint8>` שהתאים שלו חוזרים מ-255 ל-0 בעזרת `% 256`. ‏`,` קוראת עם
`io.readByte()` ומשאירה את התא ללא שינוי בסוף הקלט. ‏`.` אוספת בתים ומפענחת אותם כ-UTF-8
בתוך RuneScript לפני ההדפסה עם `io.print`, כך שתוכנית Brainfuck יכולה להדפיס כל תו.
שגיאות הולכות ל-`io.error`, ו-`main` מחזירה את קוד היציאה.

הבא: חזרה ל[תוכן העניינים של המדריך](README.md)

</div>
