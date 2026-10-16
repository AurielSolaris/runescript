<!-- source: docs/guide/17-no-null-any-generics.md @ 3cd2c1f -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/17-no-null-any-generics.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 17. בלי `null`, בלי `any`, בלי גנריות

RuneScript משמיטה שלושה דברים שיש ברוב השפות: ערך `null`, טיפוס `any` שמקבל הכול, וגנריות
שכותבים בעצמכם. הפרק הזה מסביר למה זה מחזק את מערכת הטיפוסים, ומראה איך לומר "אין כאן ערך"
בלי `null`: עם `false`.

## מה בדיוק חסר

| תכונה | ב-RuneScript | מה מקבלים במקום |
|---|---|---|
| `null`, `undefined`, `nil` | לא קיים (`ASX-NAME-001`) | כל ערך הוא ערך אמיתי מהטיפוס שלו. "אין ערך" הוא `bool` שערכו `false` |
| `any`, `Object`, `*` | לא קיים (`ASX-TYPE-010`) | לכל משתנה, פרמטר ותוצאה יש טיפוס מדויק אחד |
| פונקציות וטיפוסים גנריים משלכם | אי אפשר להצהיר עליהם (`ASX-PARSE-001`) | כתבו את הפונקציה לכל טיפוס שאתם צריכים. `Array<T>` ו-`Signal<T>` המובנים הם הטיפוסים היחידים שמקבלים ארגומנט טיפוס, והשפה היא שמגדירה אותם |

</div>

```actionscript
package main;

public void main() {
    string name = null;   // error: ASX-NAME-001
}
```

```actionscript
package main;

public void main() {
    any value = 5;   // error: ASX-TYPE-010
}
```

<div dir="rtl">

## למה ההשמטה מחזקת את הטיפוסים

**בלי `null`, טיפוס אומר את מה שהוא אומר.** בשפה עם `null`, משתנה `string` פירושו בעצם
"מחרוזת, או כלום", וכל שימוש בו הוא מקום שבו התוכנית יכולה לקרוס. הממציא שלו, טוני הואר,
כינה אותו "הטעות של מיליארד הדולר". ב-RuneScript, `string` הוא תמיד מחרוזת: כל משתנה חייב
לקבל ערך התחלתי כשמצהירים עליו, אין ערכי ברירת מחדל, ושום פעולה לא מייצרת הפניה ריקה. כשערך
באמת יכול לחסור, התוכנית אומרת זאת עם `bool` נפרד שאפשר לראות ולבדוק.

**בלי `any`, המהדר בודק הכול.** `any` הוא חור במערכת הטיפוסים: מה שנכנס יוצא בלי בדיקה,
ושגיאה מתגלה רק כשהתוכנית רצה, לעיתים קרובות רחוק מהסיבה שלה. דרך `any` והמרות לא בדוקות של
TypeScript נכנסים ערכים שגויים לתוכניות "עם טיפוסים". ב-RuneScript אין חור כזה, ובגלל זה היא
יכולה לבדוק כל ערך שחוזר מ-JavaScript ([פרק 11](11-javascript.md)): אין טיפוס שיאפשר לערך לא
בדוק לעבור.

**בלי גנריות של המשתמש, לכל פונקציה יש חתימה מוחשית אחת.** המשמעות של פונקציה גנרית תלויה
בטיפוסים שמשתמשים בה איתם, ולכן השגיאות שלה מופיעות בקריאה, במונחים של משתני טיפוס. הפונקציות
של RuneScript אומרות בדיוק מה הן מקבלות ומחזירות, הודעות השגיאה שלה נוקבות בטיפוסים אמיתיים,
וה-JavaScript המהודר לא צריך מידע על טיפוסים בזמן ריצה. `Array<T>` המובנה מכסה את הצורך הנפוץ
ביותר בגנריות, רשימה של כל דבר, בלי לפתוח את מערכת הטיפוסים.

המחיר הוא קצת חזרתיות: `maxInt` ו-`maxFloat64` במקום `max` גנרית אחת. RuneScript מקבלת את
המחיר הזה תמורת קוד שתמיד אפשר לסמוך על הטיפוסים שלו.

## שימוש ב-`false` במקום `null`

`false` לבדו לא מחליף ערך חסר, כי `string` לא יכול להכיל `bool` (`ASX-TYPE-001`). במקום זאת,
לערך שעלול לחסור מצורף `bool` שערכו `false` כל עוד אין ערך, ו-`true` כשיש. כך המהדר הופך את
ההיעדר לגלוי: אי אפשר לקרוא את הערך בלי שהדגל יושב ממש לידו.

### דגל ליד הערך

הצהירו על הדגל ועל ערך ממלא מקום יחד, והציבו את שניהם בבת אחת:

</div>

```actionscript
package main;

public void main() {
    bool hasWinner = false;   // false plays the role of null
    string winner = "";       // a placeholder; never read while hasWinner is false

    int score = 42;
    if (score > 40) {
        hasWinner = true;
        winner = "Ada";
    }

    if (hasWinner) {
        trace("winner:", winner);   // winner: Ada
    } else {
        trace("no winner yet");
    }
}
```

<div dir="rtl">

### פונקציה שאולי לא תמצא כלום

פונקציה לא יכולה להחזיר "מחרוזת או `false`", כי לתוצאה שלה יש טיפוס אחד. החזירו מערך עם אפס
איברים או איבר אחד: ריק פירושו "לא נמצא". ה-`length > 0` שלו הוא ה-`true`/`false` שבודקים:

</div>

```actionscript
package main;

use "std:string";

public Array<string> findUser(Array<string> users, string prefix) {
    for (int i = 0; i < users.length; i++) {
        if (string.hasPrefix(users[i], prefix)) {
            return [users[i]];
        }
    }
    return [];   // nothing found: the empty array is the "null"
}

public void main() {
    Array<string> users = ["ada", "grace", "linus"];
    Array<string> found = findUser(users, "gr");
    bool ok = found.length > 0;
    if (ok) {
        trace("found", found[0]);   // found grace
    }
    trace(findUser(users, "zz").length > 0);   // false
}
```

<div dir="rtl">

### שאלה ותשובה

כשחיפוש הערך זול, שאלו קודם בעזרת פונקציה שמחזירה `bool`, ואז קבלו את הערך. כך `string.contains`
ו-`string.index` עובדות יחד:

</div>

```actionscript
package main;

use "std:string";

public void main() {
    string line = "key=value";
    if (string.contains(line, "=")) {
        int at = string.index(line, "=");
        trace(string.slice(line, 0, at));   // key
    }
}
```

<div dir="rtl">

פונקציות שמחזירות מיקום משתמשות ב-`-1` עבור "לא נמצא", כמו ב-Go. זה עובד כי מיקום אמיתי אף
פעם אינו שלילי, ולכן אי אפשר לטעות ב-`-1` כאילו היה תוצאה.

### שגיאה כשההיעדר הוא טעות

כשערך חסר פירושו שמשהו השתבש, אל תחזירו דגל בכלל. זרקו שגיאה, ותנו לקורא להחליט
([פרק 14](14-errors.md)):

</div>

```actionscript
package main;

public int parsePort(string text) {
    int port = (int) text;   // throws TypeError if text isn't a number
    if (port < 1 || port > 65535) {
        throw new RangeError("port out of range: " + text);
    }
    return port;
}

public void main() {
    try {
        trace(parsePort("8080"));   // 8080
        trace(parsePort("99999"));
    } catch (RangeError e) {
        trace(e.message);   // port out of range: 99999
    }
}
```

<div dir="rtl">

## בחירת דפוס

| הערך עלול לחסור כי… | השתמשו ב- |
|---|---|
| הוא עדיין לא ידוע | דגל `bool` שערכו `false` עד שמציבים ערך |
| חיפוש לא מצא כלום | מערך עם אפס איברים או איבר אחד, או `-1` עבור מיקום |
| הקורא יכול לשאול קודם | שאלת `bool`, ואז הערך |
| משהו השתבש | `throw` של שגיאה |

`struct` מחזיק את הדגל ואת הערך יחד בערך אחד, למשל `{ bool present; int value; }`, כך שלא
יוכלו להיפרד: ראו [מבנה עבור "אין ערך"](19-structs.md) בפרק 19.

הבא: [מספרים לעומק](18-numbers.md)

</div>
