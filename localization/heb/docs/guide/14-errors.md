<!-- source: docs/guide/14-errors.md @ 362a461 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/14-errors.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 14. שגיאות: `throw`, `try`, `catch`, `finally`

שגיאות עובדות כמו ב-ActionScript 3, עם התחביר של RuneScript שבו הטיפוס בא קודם:
`catch (Error e)` במקום `catch (e:Error)`.

</div>

```actionscript
package main;

public int parsePort(string text) {
    int port = (int) text;   // throws a TypeError if text isn't a number
    if (port < 1 || port > 65535) {
        throw new RangeError("port out of range", 1);
    }
    return port;
}

protected int main() {
    try {
        trace(parsePort("8080"));
        trace(parsePort("99999"));
    } catch (RangeError e) {
        trace("bad port:", e.message);
    } catch (Error e) {
        trace("not a port:", e);
    } finally {
        trace("done");
    }
    return 0;
}
```

<div dir="rtl">

התוכנית מדפיסה `8080`, `bad port: port out of range` ו-`done`.

## ערכי שגיאה

יש ארבעה טיפוסי שגיאה. שלושת האחרונים הם סוגים של `Error`:

| טיפוס | משמש עבור |
|---|---|
| `Error` | כל דבר; טיפוס הבסיס |
| `RangeError` | ערך מחוץ לטווח: גלישה, חלוקה באפס, מספר שאינו נכנס |
| `TypeError` | ערך בצורה שגויה, למשל `(int) "abc"` |
| `ArgumentError` | ארגומנט שגוי; עבור הפונקציות שלכם |

`new Error(message, errorID)` יוצר שגיאה. שני הארגומנטים אופציונליים. לכל שגיאה יש
שלושה שדות:

| שדה | טיפוס | ערך |
|---|---|---|
| `e.message` | `string` | ההודעה |
| `e.name` | `string` | שם הטיפוס, למשל `"RangeError"` |
| `e.errorID` | `int` | המזהה שהועבר ל-`new`, או המספר של שגיאת זמן הריצה |

שגיאות הן ערכים רגילים. אפשר לשמור אותן, להעביר אותן, להחזיר אותן ולהדפיס אותן
ב-`trace`: `trace(e)` מדפיסה `RangeError: message`. אפשר לשמור `RangeError` במשתנה
מטיפוס `Error`. אפשר לזרוק (`throw`) רק שגיאה, לא מחרוזת או מספר.

## `try`, `catch` ו-`finally`

- פסוקיות ה-catch נבדקות לפי הסדר. הראשונה שהשגיאה שייכת לטיפוס שלה (הטיפוס הזה,
  או סוג שלו) רצה. שימו טיפוסים ספציפיים לפני `Error`, כי פסוקית שלעולם לא יכולה
  לרוץ היא שגיאה (`ASX-TYPE-022`).
- אם אף פסוקית לא מתאימה, השגיאה ממשיכה ל-`try` העוטף ולפונקציות הקוראות.
- בלוק ה-`finally` רץ תמיד: אחרי שהבלוק מסתיים, אחרי `return`, `break` או
  `continue`, ואחרי שגיאה, שממשיכה אחר כך הלאה.
- בלוק `finally` יכול לבצע `return`, `break` או `continue`, כמו ב-ActionScript. זה
  מחליף את הדרך שבה הוראת ה-`try` עמדה להסתיים, ושגיאה שבדרך נזנחת.
- `catch (e)`, בלי טיפוס, תופס כל שגיאה, כמו `catch (Error e)`.
- המשתנה של השגיאה שנתפסה הוא לקריאה בלבד.

`throw` מסיים מסלול כמו ש-`return` מסיים, ולכן הפונקציה הזו אינה צריכה `return`
בסופה:

</div>

```actionscript
package main;

public int sign(int n) {
    if (n > 0) {
        return 1;
    }
    if (n < 0) {
        return -1;
    }
    throw new ArgumentError("zero has no sign");
}

protected int main() {
    trace(sign(-5));
    return 0;
}
```

<div dir="rtl">

## אפשר לתפוס שגיאות זמן ריצה

שגיאות זמן הריצה מ[פרק 8](08-diagnostics.md) נזרקות כשגיאות שאפשר לתפוס:

| שגיאת זמן ריצה | נזרקת כ- |
|---|---|
| `ASX-RUN-002`: גלישה במספר שלם | `RangeError` |
| `ASX-RUN-003`: חלוקה של מספר שלם באפס | `RangeError` |
| `ASX-RUN-004`: פונקציה של הספרייה הסטנדרטית נכשלה | `Error` |
| `ASX-RUN-006`: המרה נכשלה | `RangeError` או `TypeError` |

</div>

```actionscript
package main;

protected int main() {
    int big = 2147483647;
    try {
        trace(big + 1);
    } catch (RangeError e) {
        trace(e.errorID, e.message);
        // 2 integer overflow: 2147483647 + 1 = 2147483648 does not fit in 'int'
    }
    return 0;
}
```

<div dir="rtl">

היוצא מן הכלל היחיד הוא `ASX-RUN-001`, מגבלת עומק הקריאות (בדרך כלל רקורסיה
שיצאה משליטה). אי אפשר לתפוס אותה, ובלוקי `finally` אינם רצים כשהתוכנית נעצרת
בגללה.

## שגיאות ששום דבר לא תופס

שגיאה ששום דבר לא תופס עוצרת את התוכנית, כמו כל שגיאת זמן ריצה:

- שגיאת זמן ריצה מדווחת עם הקוד, ההודעה והמיקום שלה עצמה, גם אם נתפסה ונזרקה
  שוב.
- שגיאה שיצרתם עם `new` מדווחת כ-`ASX-RUN-005` במקום ה-`throw`:

</div>

```actionscript
package main;

protected int main() {
    throw new ArgumentError("missing input");   // runtime error: ASX-RUN-005
    return 0;
}
```

```text
main.rsc:4:5: error ASX-RUN-005: uncaught ArgumentError: missing input
```

<div dir="rtl">

עדיין לא זמין: עקבות מחסנית (stack traces), מתודות של שגיאות כמו `toString()`,
וטיפוסי שגיאה משלכם (הם דורשים מחלקות).

הבא: [דפי רשת עם `rune:dom`](15-web-pages.md)

</div>
