<!-- source: docs/guide/09-coming-from-actionscript.md @ 34d09ef -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/09-coming-from-actionscript.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 9. למי שמגיע מ-ActionScript 3

RuneScript *שואבת השראה* מ-ActionScript 3, אבל אינה תואמת לה. קוד AS3 לא יתקמפל
בלי שינויים. ההבדלים העיקריים:

## הצהרות: קודם הטיפוס, בלי `function`, בלי `var`

ActionScript 3:

</div>

```as3
package {
    public class Greeter {
        public function sayHi(message:String, times:int):void {
            for (var i:int = 0; i < times; i++) {
                trace(message);
            }
        }
    }
}
```

<div dir="rtl">

RuneScript (בגרסה 0.1 עדיין אין מחלקות, ולכן זו פונקציה ברמה העליונה):

</div>

```actionscript
package main;

public void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

public void main() {
    sayHi("hi", 2);
}
```

<div dir="rtl">

| AS3 | RuneScript |
|---|---|
| `package x { … }` | `package x;` |
| `function f(a:int):int` | `int f(int a)` |
| `var n:int = 0;` | `int n = 0;` |
| `String`, `Boolean` | `string`, `bool` |
| נקודה-פסיק אופציונלית | נקודה-פסיק חובה |
| מותר `*` או להשמיט טיפוסים | לכל משתנה, פרמטר ופונקציה יש טיפוס מפורש; אין `any` |
| `Array`, `Vector.<String>` | `Array<string>`: לכל איבר יש הטיפוס המוצהר |
| `"n=" + 5` | `"n=" + (string) 5` או `$"n={5}"`: `+` אף פעם לא ממיר מספרים |

## מספרים שלמים בהקפדה רבה יותר

| קוד | AS3 | RuneScript |
|---|---|---|
| `var u:uint = -1;` | הופך ל-4294967295 | שגיאת הידור `ASX-TYPE-007` |
| `uint` → משתנה `int` | מומר בשקט | שגיאת הידור `ASX-TYPE-004` |
| `int + uint` | מומר בשקט | שגיאת הידור `ASX-TYPE-011` |
| `int.MAX_VALUE + 1` | מתגלגל למספר שלילי | לעולם אינו מתגלגל בשקט (גרסה 0.1: שגיאת זמן ריצה `ASX-RUN-002`) |
| `7 / 2` | `3.5` (ערך `Number`) | `3` (חילוק של מספרים שלמים) |

## תנאים בהקפדה רבה יותר

ב-AS3 אפשר לכתוב `if (count)` או `if (name)`. RuneScript דורשת `bool`, ולכן כתבו
`if (count != 0)`.

## סְגוֹרים (closures)

סגורים ב-AS3 לוכדים את המשתנים של הפונקציה העוטפת. פונקציות RuneScript אינן עושות
זאת: פונקציות מקוננות וביטויי למבדה יכולים להשתמש רק בפרמטרים ובמשתנים המקומיים
שלהם, וגם בפונקציות אחרות. במקום זאת, העבירו ערכים כפרמטרים.

| AS3 | RuneScript |
|---|---|
| `function(a:int):int { return a * 2; }` | `(int a) => a * 2` |
| `var f:Function = add;` | `int(int, int) f = add;` (עם טיפוס) |
| `public function f():void` | `public void f()` (משנה הגישה הוא חובה) |

## טווח הכרה של בלוק

ב-AS3, `var` שהוצהר בתוך לולאה נראה בכל הפונקציה. ב-RuneScript משתנה חי רק בבלוק
שלו, ומשתנה חדש אינו יכול להשתמש שוב בשם של משתנה שעדיין נראה.

## נקודת הכניסה

תוכניות AS3 מתחילות ממחלקת מסמך (document class). תוכניות RuneScript מתחילות
מפונקציית `main()` ברמה העליונה, ואין צורך במחלקה.

## כמו ב-AS3

- `trace(a, b, c)` מדפיסה את הארגומנטים שלה מופרדים ברווחים.
- `public`, `protected` ו-`private` הם משני נראות (הכללים המדויקים שלהם עדיין
  בתכנון).
- `int` ו-`uint` הם השמות המוכרים למספרים שלמים, ו-`int` הוא בן 32 סיביות כברירת
  מחדל.
- `if`, `while`, `for`, `break`, `continue`, `&&`, `||` ו-`!` בסגנון C.

## עדיין לא זמין

מחלקות, ממשקים, `static`, `null`, מתודות מערך מעבר ל-`push` ו-`pop`, מטא-נתונים,
E4X XML, וגרפיקה (מתוכננת חבילת גרפיקה מקורית). ראו את [מפת הדרכים](../../../../docs/ROADMAP.md)
(באנגלית).

הבא: [הספרייה הסטנדרטית: `std:string`](10-std-string.md)

</div>
