<!-- source: docs/guide/19-structs.md @ 6704ceb -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/19-structs.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 19. מבנים (structs)

מבנה מקבץ ערכים בעלי שם לערך אחד מטיפוס חדש:

</div>

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

protected int main() {
    Point p = Point{x: 3, y: 4};
    trace(p, p.x + p.y);   // Point{x: 3, y: 4} 7
    p.y = 10;
    trace(p);              // Point{x: 3, y: 10}
    return 0;
}
```

<div dir="rtl">

## הצהרה על מבנה

מצהירים על מבנה ברמה העליונה של קובץ, עם משנה גישה כמו של פונקציה, ועם רשימה של שדות
מטיפוסים, שכל אחד מהם מסתיים ב-`;`:

</div>

```actionscript
package main;

public struct Player {
    string name;
    int score;
    Array<string> badges;
}

private struct Settings {
    bool sound;
    float volume;
}

protected int main() {
    Player p = Player{name: "Ada", score: 0, badges: []};
    Settings s = Settings{sound: true, volume: 0.5};
    trace(p.name, s.volume);   // Ada 0.5
    return 0;
}
```

<div dir="rtl">

- במבנים `public` אפשר להשתמש בכל מקום, במבנים `protected` רק בחבילה שלהם, ובמבנים
  `private` רק בקובץ שלהם.
- לשדה יכול להיות כל טיפוס ערך: מספרים, `bool`, `string`, `char`, מערכים, פונקציות, שגיאות
  ומבנים אחרים.
- לשדות אין ערכי ברירת מחדל. זו הסיבה שכל ליטרל חייב לנקוב בכולם.

## בניית מבנה

**ליטרל מבנה** נוקב בשם המבנה, ואז נותן ערך לכל שדה, עם `name: value`:

</div>

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

protected int main() {
    Point a = Point{x: 1, y: 2};
    Point b = Point{y: 2, x: 1};   // any order
    Point c = Point{x: 1};         // error: ASX-TYPE-026
    return 0;
}
```

<div dir="rtl">

השמטת שדה, מתן שדה פעמיים או נקיבה בשדה שאין במבנה הם `ASX-TYPE-026`. הערכים מחושבים בסדר
שבו כותבים אותם.

## מבנים הם ערכים

השמת מבנה, העברתו לפונקציה, החזרתו או הכנסתו למערך **מעתיקה** אותו. שינוי ההעתק אף פעם לא
משנה את המקור:

</div>

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

public Point moved(Point p, int dx) {
    Point q = p;
    q.x += dx;
    return q;
}

protected int main() {
    Point a = Point{x: 1, y: 1};
    Point b = moved(a, 5);
    trace(a, b);   // Point{x: 1, y: 1} Point{x: 6, y: 1}
    return 0;
}
```

<div dir="rtl">

זה ההפך ממערכים, שהם משותפים ([פרק 16](16-text-and-arrays.md)). אבל מערך *בתוך* מבנה הוא
עדיין מערך: העתקת המבנה משתפת את המערך.

הפרמטרים של פונקציה הם לקריאה בלבד, וכך גם השדות של פרמטר שהוא מבנה: ‏`p.x = 1` בתוך `moved`
היה `ASX-NAME-006`. העתיקו אותו קודם למשתנה מקומי, כמו למעלה.

## שדות בתוך שדות, ומערכים של מבנים

השמות מגיעות דרך שדות ואיברי מערך:

</div>

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

public struct Box {
    Point min;
    Point max;
}

protected int main() {
    Box box = Box{min: Point{x: 0, y: 0}, max: Point{x: 2, y: 2}};
    box.max.x = 5;

    Array<Box> boxes = [box, box];
    boxes[1].min.y = -3;
    trace(boxes[0].min, boxes[1].min, box.max.x);   // Point{x: 0, y: 0} Point{x: 0, y: -3} 5
    return 0;
}
```

<div dir="rtl">

## השוואה והדפסה

‏`==` ו-`!=` משווים כל שדה. אי אפשר להשוות מבנה עם שדה של מערך, פונקציה או שגיאה; השוו במקום
זאת את השדות שאתם צריכים.

‏`trace` מציג מבנה עם הטיפוס והשדות שלו, כשמחרוזות ותווים במירכאות:
`Player{name: "Ada", score: 3}`.

## מבנים של חבילות אחרות

נקבו במבנה של חבילה אחרת דרך המזהה שלה ([פרק 12](12-packages-and-files.md)):
`geom.Point p = geom.Point{x: 1, y: 2};`. המבנה חייב להיות `public`.

## מבנה עבור "אין ערך"

מבנה הוא הדרך הנקייה ביותר להחזיק ערך ואת הדגל "האם הוא קיים?" שלו יחד, הדפוס של `false`
במקום `null` מ[פרק 17](17-no-null-any-generics.md):

</div>

```actionscript
package main;

use "std:string";

public struct MaybeInt {
    bool present;   // false means "no value"
    int value;
}

public MaybeInt parseAge(string text) {
    string digits = string.trimSpace(text);
    if (digits == "" || string.indexFunc(digits, (char c) => c < '0' || c > '9') >= 0) {
        return MaybeInt{present: false, value: 0};
    }
    return MaybeInt{present: true, value: (int) digits};
}

protected int main() {
    MaybeInt age = parseAge(" 42 ");
    if (age.present) {
        trace("age", age.value);   // age 42
    }
    trace(parseAge("forty").present);   // false
    return 0;
}
```

<div dir="rtl">

## עדיין לא זמין

מתודות, בנאים, שדות פרטיים, והעברת מבנים לפונקציות `extern` של JavaScript.

הבא: חזרה ל[תוכן העניינים של המדריך](README.md)

</div>
