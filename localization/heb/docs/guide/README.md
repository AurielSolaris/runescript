<!-- source: docs/guide/README.md @ b50b7fd -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/README.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# המדריך ל-RuneScript (גרסה 0.6)

סיור מעשי בתכונות של RuneScript שקיימות **היום**. כל תוכנית שלמה בדפים האלה
מתקמפלת בחבילת הבדיקות (`tests/docs_test.go`). אם אחת מהן מפסיקה להתקמפל,
הבנייה נכשלת.

1. [מתחילים](01-getting-started.md): בניית ה-CLI, הרצת תוכנית ו-`run --watch`
2. [מבנה התוכנית](02-program-structure.md): חבילות, הערות, `main`
3. [טיפוסים ומשתנים](03-types-and-variables.md): `int`, `uint`, `bool`, `string`, השמה
4. [אופרטורים](04-operators.md): חשבון, השוואה, לוגיקה
5. [בקרת זרימה](05-control-flow.md): `if`, `while`, `for`, `break`, `continue`
6. [פונקציות](06-functions.md): פרמטרים, ערכי החזרה, רקורסיה
7. [פלט עם `trace`](07-trace.md)
8. [מדריך הודעות האבחון](08-diagnostics.md): כל קודי השגיאה והאזהרה, `-Werror`, `--allow`
9. [למי שמגיע מ-ActionScript 3](09-coming-from-actionscript.md)
10. [הספרייה הסטנדרטית: `std:string`](10-std-string.md): ה-`strings` של Go, ב-RuneScript
11. [הידור ל-JavaScript](11-javascript.md): `runescript build`, איך הקפדת הטיפוסים נשמרת, קריאה ל-JavaScript עם `extern`, וטעינה מחדש חיה עם `runescript dev`
12. [חבילות וקבצים](12-packages-and-files.md): `use`, תוכניות מרובות קבצים, נראות
13. [מספרים עשרוניים והמרות](13-floats-and-conversions.md): `float`, `float32`, `float64`, `(int) x`
14. [שגיאות](14-errors.md): `throw`, `try`, `catch`, `finally`, ותפיסת שגיאות זמן ריצה
15. [דפי רשת עם `rune:dom`](15-web-pages.md): רכיבים, תבניות HTML ו-SVG בדוקות, runes, אירועים, `$if`, `renderApp`
16. [טקסט ורשימות](16-text-and-arrays.md): חיבור מחרוזות עם `+`, מחרוזות משובצות `$"…"`, `char` ו-`Array<T>`
17. [בלי `null`, בלי `any`, בלי גנריות](17-no-null-any-generics.md): למה, ואיך להשתמש ב-`false` במקום שבו שפות אחרות משתמשות ב-`null`
18. [מספרים לעומק](18-numbers.md): `int8` … `uint64`, ליטרלים הקסדצימליים ובינאריים, פעולות על סיביות, `std:math`
19. [מבנים](19-structs.md): הצהרה, בנייה והעתקה של טיפוסי ערך
20. [קבצים, קלט ופלט](20-files-and-input.md): `std:io`, `std:process`, `std:fs`, `std:path`, ומפרש Brainfuck
21. [JSON](21-json.md): ‏`std:json`: פענוח JSON, קריאה בפונקציות בדוקות, בנייה וכתיבה

לכללים המדויקים, ראו את [מפרט השפה](../../../../docs/LANGUAGE_SPEC.md) (באנגלית).
למה שמתוכנן בהמשך, ראו את [מפת הדרכים](../../../../docs/ROADMAP.md) (באנגלית).

> **עדיין לא זמין:** מחלקות. מילת המפתח שלהן שמורה, ושימוש בה מניב שגיאה ברורה של
> "עדיין לא מומש". גם אובייקטים, מערכים, מבנים ו-callbacks עדיין לא יכולים לחצות את
> הגבול עם JavaScript.

</div>
