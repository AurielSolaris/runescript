<!-- source: docs/guide/02-program-structure.md @ 6704ceb -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/02-program-structure.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 2. מבנה התוכנית

## חבילות

כל קובץ מתחיל בהצהרת חבילה. זוהי הוראה אחת שמסתיימת ב-`;`, ולא בלוק:

</div>

```actionscript
package games.puzzle.solver;

public void main() {
}
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

public void main() {
    trace(string.toUpper("hello"));   // HELLO
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

public void main() {
    trace("comments are ignored"); // trailing comments too
}
```

<div dir="rtl">

אי אפשר לקנן הערות בלוק זו בתוך זו.

## הוראות ובלוקים

- כל הוראה מסתיימת ב-`;`.
- סוגריים מסולסלים `{ }` מסמנים את הגוף של פונקציה, `if`, `else`, `while` או `for`.
  בגרסה 0.1 הסוגריים האלה **חובה**, גם סביב הוראה אחת.

## נקודת הכניסה: `main`

הרצת תוכנית קוראת לפונקציה `main` שברמה העליונה. `main` אינה מקבלת פרמטרים
ומחזירה `int` או `void`:

</div>

```actionscript
package main;

protected int main() {
    // do stuff
    return 0;
}
```

<div dir="rtl">

- אם `main` מחזירה `int`, הערך הופך לקוד היציאה של התהליך.
- אם `main` היא `void`, קוד היציאה הוא 0.
- כמו כל פונקציה, `main` מתחילה במשנה גישה: `public`, `protected` או `private`
  (ראו [פונקציות](06-functions.md)).

</div>

```actionscript
package main;

public void main() {
    trace("no return value needed");
}
```

<div dir="rtl">

אין צורך במחלקה. תוכניות בנויות מפונקציות ברמה העליונה, שיכולות להופיע בכל סדר.
אפשר גם לקנן פונקציות בתוך פונקציות אחרות.

## שמות

מזהים מורכבים מאותיות ASCII, ספרות ו-`_`, ואינם יכולים להתחיל בספרה. יש הבחנה בין
אותיות גדולות לקטנות. אי אפשר להשתמש כשמות במילות מפתח כמו `if`, `int`, `use`
ו-`return`. גם מילים השמורות לתכונות עתידיות (`class`, `array` ו-`matrix`) אינן מותרות.

הבא: [טיפוסים ומשתנים](03-types-and-variables.md)

</div>
