<!-- source: docs/guide/02-program-structure.md @ 362a461 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/02-program-structure.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 2. מבנה התוכנית

## חבילות

כל קובץ מתחיל בהצהרת חבילה. זוהי הוראה אחת שמסתיימת ב-`;`, ולא בלוק:

</div>

```actionscript
package games.puzzle.solver;

protected int main() { return 0; }
```

<div dir="rtl">

שמות חבילות הם מזהים המופרדים בנקודות. תוכנית יכולה להתפרש על פני כמה קבצים
(ראו [חבילות וקבצים](12-packages-and-files.md)).

## שימוש בחבילות: `use`

הצהרות `use` באות מיד אחרי שורת ה-`package`, לפני כל פונקציה. חבילות של הספרייה
הסטנדרטית נכתבות בצורה `"std:<name>"`:

</div>

```actionscript
package main;

use "std:string";

protected int main() {
    trace(string.toUpper("hello"));   // HELLO
    return 0;
}
```

<div dir="rtl">

את הפונקציות קוראים אחר כך דרך המזהה של החבילה (החלק האחרון בשמה). ראו
[הספרייה הסטנדרטית: std:string](10-std-string.md).

`use` טוען גם חבילות RuneScript אחרות (`use "games.engine";`) וקבצים
(`use "./lib/helpers.rsc";`), כך שתוכנית יכולה להתפרש על פני כמה קבצים. ראו
[חבילות וקבצים](12-packages-and-files.md). חבילות Go (`use "go:<package>";`)
מתוכננות, וכיום הן מניבות `ASX-USE-003`.

## הערות

</div>

```actionscript
package main;

// A line comment runs to the end of the line.

/* A block comment
   can span several lines. */

protected int main() {
    trace("comments are ignored"); // trailing comments too
    return 0;
}
```

<div dir="rtl">

אי אפשר לקנן הערות בלוק זו בתוך זו.

## הוראות ובלוקים

- כל הוראה מסתיימת ב-`;`.
- סוגריים מסולסלים `{ }` מסמנים את הגוף של פונקציה, `if`, `else`, `while` או `for`.
  בגרסה 0.1 הסוגריים האלה **חובה**, גם סביב הוראה אחת.

## נקודת הכניסה: `main`

הרצת תוכנית קוראת לפונקציה `main` שברמה העליונה, שתמיד מוצהרת בדיוק כך:

</div>

```actionscript
package main;

protected int main() {
    // do stuff
    return 0;
}
```

<div dir="rtl">

- ה-`int` שהיא מחזירה הופך לקוד היציאה של התהליך: `0` פירושו הצלחה.
- היא `protected`, אינה מקבלת פרמטרים ומחזירה `int`. כל צורה אחרת, כמו
  `public int main()` או `protected void main()`, היא השגיאה `ASX-ENTRY-002`, והעזרה
  שלה אומרת מה לשנות.
- לתוכנית יש בדיוק `main` אחת, באחד מהקבצים שלה.
- ארגומנטים משורת הפקודה מגיעים מ-`process.args()` שב-`std:process`
  ([פרק 20](20-files-and-input.md)).

אין צורך במחלקה. תוכניות בנויות מפונקציות ברמה העליונה, שיכולות להופיע בכל סדר.
אפשר גם לקנן פונקציות בתוך פונקציות אחרות.

## שמות

מזהים מורכבים מאותיות ASCII, ספרות ו-`_`, ואינם יכולים להתחיל בספרה. יש הבחנה בין
אותיות גדולות לקטנות. אי אפשר להשתמש כשמות במילות מפתח כמו `if`, `int`, `use`
ו-`return`. גם מילים השמורות לתכונות עתידיות (`class`, `array` ו-`matrix`) אינן מותרות.

הבא: [טיפוסים ומשתנים](03-types-and-variables.md)

</div>
