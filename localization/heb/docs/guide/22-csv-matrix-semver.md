<!-- source: docs/guide/22-csv-matrix-semver.md @ be7087a -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/22-csv-matrix-semver.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 22. CSV, מטריצות וגרסאות

שלוש חבילות קטנות לנתונים של יום-יום:

- ‏`std:csv` קוראת וכותבת ערכים מופרדים בפסיקים (CSV) בתור `Array<Array<string>>`.
- ‏`std:matrix` עושה אלגברה לינארית על `Array<Array<float64>>`.
- ‏`std:semver` בודקת, משווה ומתאימה מחרוזות גרסה כמו `1.4.0-rc.1`.

שלושתן הן פונקציות פשוטות על מערכים ומחרוזות: אין טיפוסים חדשים ללמוד. כמו שאר הספרייה
הסטנדרטית, הן נותנות את אותן תוצאות, ואת אותן הודעות שגיאה, במפרש וביעד ה-JavaScript.

## ‏CSV: ‏`std:csv`

‏`csv.parse` מפצלת טקסט לרשומות, וכל רשומה לשדות. שדה בתוך מירכאות כפולות יכול להכיל
פסיקים, שורות חדשות ומירכאות (שנכתבות פעמיים):

</div>

```actionscript
package main;

use "std:csv";

protected int main() {
    string text = "name,city\nAda,London\n\"Turing, Alan\",\"Wilmslow\"\n";
    Array<Array<string>> rows = csv.parse(text);
    trace(rows.length);                      // 3
    for each (Array<string> row in rows) {
        trace(row[0], "|", row[1]);          // name | city, then Ada | London, ...
    }
    return 0;
}
```

<div dir="rtl">


כדי לקרוא קובץ, שלבו עם `std:fs`: ‏`csv.parse(fs.readFile("data.csv"))`.

‏`csv.format` עושה את הדרך ההפוכה. היא מקיפה שדה במירכאות רק כשחייבים, ומסיימת כל רשומה
בשורה חדשה, כך ש-`csv.parse(csv.format(rows))` מחזירה את `rows`:

</div>

```actionscript
package main;

use "std:csv";

protected int main() {
    Array<Array<string>> rows = [["item", "note"], ["lamp", "red, tall"], ["rug", "says \"hi\""]];
    trace(csv.format(rows));
    // item,note
    // lamp,"red, tall"
    // rug,"says ""hi"""
    return 0;
}
```

<div dir="rtl">


‏`csv.parseWith` ו-`csv.formatWith` מקבלות מפריד אחר, בתור `char`: ‏`'\t'` לקבצים
מופרדים בטאבים, ו-`';'` לקבצים שגיליונות אלקטרוניים רבים כותבים באירופה:

</div>

```actionscript
package main;

use "std:csv";

protected int main() {
    Array<Array<string>> rows = csv.parseWith("a\tb\n1\t2", '\t');
    trace(rows[1][1]);                       // 2
    trace(csv.formatWith(rows, ';'));        // a;b, then 1;2
    return 0;
}
```

<div dir="rtl">


הכללים:

- רשומות מסתיימות ב-`\n` או ב-`\r\n`. שורות ריקות מדולגות, ולכן בטקסט ריק אין רשומות.
- לרשומות יכול להיות מספר שונה של שדות. בדקו את `row.length` אם הנתונים שלכם צריכים מספר
  קבוע.
- רווחים הם חלק מהשדה: ‏`" a "` נשאר `" a "`. שדות במירכאות שומרים את השורות החדשות שלהם
  בדיוק.
- סימן סדר בתים (BOM) של UTF-8 בהתחלה, כמו שגיליונות אלקטרוניים מסוימים כותבים, מדולג.
- מירכאה בתוך שדה בלי מירכאות (`ab"c`), טקסט אחרי מירכאה סוגרת (`"a"b`), ומירכאה שאף פעם
  לא נסגרת הן שגיאות. ההודעה אומרת היכן: ‏`csv.parse: line 2, column 4: …`. העמודות סופרות
  בתים, כמו היסטים במחרוזות ([פרק 10](10-std-string.md)).
- המפריד יכול להיות כל תו חוץ מ-`"`, ‏`\r` ו-`\n` ‏(`ArgumentError`). ‏`csv.format` זורקת
  `ArgumentError` עבור רשומה בלי שדות, ש-CSV לא יכול לכתוב.

## מטריצות: ‏`std:matrix`

מטריצה היא מערך של שורות: ‏`[[1, 2], [3, 4]]` הוא `Array<Array<float64>>`. כל פונקציה
בודקת שלשורות יש אותו אורך ושהגדלים מתאימים לפעולה, ומחזירה מערך חדש: הארגומנטים אף פעם
לא משתנים.

</div>

```actionscript
package main;

use "std:matrix";

protected int main() {
    Array<Array<float64>> a = [[1, 2], [3, 4]];
    Array<Array<float64>> b = [[0, 1], [1, 0]];
    trace(matrix.format(matrix.mul(a, b)));  // [2, 1], then [4, 3]
    trace(matrix.format(matrix.add(a, b)));  // [1, 3], then [4, 4]
    trace(matrix.det(a));                    // -2
    trace(matrix.rows(a), matrix.cols(a));   // 2 2
    return 0;
}
```

<div dir="rtl">


| פונקציה | תוצאה |
|---|---|
| `zeros(rows, cols)`, `identity(n)` | מטריצה חדשה של אפסים, או מטריצת היחידה |
| `rows(m)`, `cols(m)` | הגודל (למטריצה בלי שורות יש 0 עמודות) |
| `add(a, b)`, `sub(a, b)`, `scale(m, k)` | איבר אחר איבר |
| `mul(a, b)`, `mulVec(m, v)` | מכפלת מטריצות, ומטריצה כפול וקטור (`Array<float64>`) |
| `transpose(m)` | השורות הופכות לעמודות |
| `det(m)`, `inverse(m)` | הדטרמיננטה וההופכית של מטריצה ריבועית |
| `solve(a, b)` | ה-`x` שעבורו `a·x = b`, עבור `a` ריבועית |
| `format(m)` | שורה אחת לכל שורת מטריצה: ‏`[1, 2.5]` |

‏`solve` עדיפה על כפל בהופכית: היא עושה פחות עבודה ומאבדת פחות דיוק. כאן היא מוצאת את
המחירים של שני פריטים משתי קבלות:

</div>

```actionscript
package main;

use "std:matrix";

protected int main() {
    // 2 apples + 1 pear cost 4; 1 apple + 3 pears cost 7.
    Array<Array<float64>> counts = [[2, 1], [1, 3]];
    Array<float64> prices = matrix.solve(counts, [4.0, 7.0]);
    trace(prices[0], prices[1]);             // 1 2
    return 0;
}
```

<div dir="rtl">


‏`det`, ‏`inverse` ו-`solve` משתמשות באלימינציית גאוס עם עירור חלקי (partial pivoting).
מטריצה סינגולרית (שאין לה הופכית) נותנת דטרמיננטה `0`, ו-`inverse` ו-`solve` זורקות
`Error`. כמו בכל עבודה בנקודה צפה, התוצאות יכולות לסטות בספרות האחרונות: ‏`0.1 + 0.2`
הוא `0.30000000000000004`. הן סוטות בדיוק באותו אופן בשני היעדים, כי שניהם מריצים את אותו
רצף פעולות.

טעויות בגודל זורקות `ArgumentError` שמציין את הגדלים:

</div>

```text
matrix.mul: cannot multiply a 1x2 matrix by a 1x2 matrix: the first has 2 columns, the second 1 row
```

<div dir="rtl">


‏`matrix` היא עדיין מילה שמורה: אי אפשר לקרוא למשתנה `matrix`. היא יכולה רק להקדים קריאה,
כמו ב-`matrix.mul(a, b)`.

## גרסאות: ‏`std:semver`

‏`std:semver` עובדת עם מחרוזות של [Semantic Versioning 2.0.0](https://semver.org/):
‏`MAJOR.MINOR.PATCH`, ואחריהן, אם רוצים, גרסת טרום-הפצה (`-rc.1`) ומטא-נתוני בנייה
(`+build.5`). היא משווה גרסאות ומתאימה טווחים כמו ש-npm עושה.

השוואה של גרסאות כמחרוזות נותנת תשובות שגויות, כי אות אחר אות `"1.10.0"` הייתה באה לפני
`"1.9.0"`. הפונקציות של `semver` משוות מספרים כמספרים, ומציבות טרום-הפצה לפני ההפצה שלה:

</div>

```actionscript
package main;

use "std:semver";

protected int main() {
    trace(semver.lt("1.9.0", "1.10.0"));             // true
    trace(semver.lt("1.0.0-rc.1", "1.0.0"));         // true
    trace(semver.compare("2.0.0", "1.99.0"));        // 1
    trace(semver.eq("1.0.0+build.1", "1.0.0"));      // true: build metadata is ignored
    trace(semver.valid("1.2"), semver.valid("1.2.3-rc.1")); // false true
    return 0;
}
```

<div dir="rtl">


| פונקציה | תוצאה |
|---|---|
| `valid(v)`, `validRange(r)` | האם `v` היא גרסה, או `r` טווח (אף פעם לא זורקות) |
| `compare(a, b)` | ‏`-1`, ‏`0` או `1` |
| `eq`, `neq`, `gt`, `gte`, `lt`, `lte` | השוואות `bool` |
| `satisfies(v, range)` | האם `v` נמצאת בטווח |
| `major(v)`, `minor(v)`, `patch(v)` | המספרים, בתור `int` |
| `prerelease(v)`, `build(v)` | ‏`"rc.1"`, ‏`"build.5"` או `""` |

‏`satisfies` מבינה את הטווחים של npm:

</div>

```actionscript
package main;

use "std:semver";

protected int main() {
    trace(semver.satisfies("1.4.0", "^1.2.0"));            // true: 1.x from 1.2.0 on
    trace(semver.satisfies("2.0.0", "^1.2.0"));            // false
    trace(semver.satisfies("1.2.9", "~1.2.3"));            // true: 1.2.x from 1.2.3 on
    trace(semver.satisfies("1.5.0", ">=1.2.0 <2.0.0"));    // true
    trace(semver.satisfies("3.1.0", "1.x || >=3.0.0"));    // true
    trace(semver.satisfies("2.3.9", "1.2.3 - 2.3"));       // true
    trace(semver.satisfies("1.3.0-beta", "^1.2.0"));       // false: see below
    return 0;
}
```

<div dir="rtl">

| טווח | משמעות |
|---|---|
| `1.2.3`, `=1.2.3` | בדיוק הגרסה הזאת |
| `>1.2.3`, `>=`, `<`, `<=` | השוואות; כמה מהן מופרדות ברווחים חייבות כולן להתקיים |
| `a \|\| b` | אחד משני הטווחים |
| `1.2.x`, `1.x`, `*`, `1.2` | כל גרסה שמתאימה לחלקים שניתנו |
| `~1.2.3` | ‏`>=1.2.3 <1.3.0`: עדכוני תיקון (patch) |
| `^1.2.3` | ‏`>=1.2.3 <2.0.0`: בלי שינויים שוברים (`^0.2.3` הוא `<0.3.0`) |
| `1.2.3 - 2.3.4` | ‏`>=1.2.3 <=2.3.4` |

כמו ב-npm, גרסת טרום-הפצה מתאימה לטווח רק אם הטווח מציין טרום-הפצה של אותו
`MAJOR.MINOR.PATCH`: ‏`1.3.0-beta` אינה בטווח `^1.2.0`, אבל `1.2.3-beta.4` נמצאת בטווח
`^1.2.3-beta.2`. כך גרסאות ניסיון אף פעם לא נבחרות בטעות.

הכללים:

- גרסאות הן SemVer 2.0.0 מחמיר: בלי אפסים מובילים (`01.2.3`), בלי קידומת `v` ‏(`v1.2.3`),
  בלי רווחים. במספרים יכולות להיות כמה ספרות שרוצים; ‏`major`, ‏`minor` ו-`patch` זורקות
  `RangeError` אם מספר לא נכנס ב-`int`.
- כל פונקציה חוץ מ-`valid` ו-`validRange` זורקת `ArgumentError` עבור גרסה או טווח לא
  תקינים: ‏`semver.compare: "1.2" is not a valid version: …`.

הבא: [עורכים](23-editors.md)

</div>
