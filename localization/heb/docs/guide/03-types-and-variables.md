<!-- source: docs/guide/03-types-and-variables.md @ 52427d9 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/03-types-and-variables.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 3. טיפוסים ומשתנים

## הצהרה על משתנים

כתבו קודם את הטיפוס, אחריו את השם, ואחריו ערך התחלתי:

</div>

```actionscript
package main;

public void main() {
    int count = 10;
    uint total = 250;
    bool ready = true;
    string name = "Ada";
    trace(name, count, total, ready);   // Ada 10 250 true
}
```

<div dir="rtl">

כל משתנה **חייב** לקבל ערך התחלתי כשהוא מוצהר. בגרסה 0.1 אין ערכי ברירת מחדל
מובלעים.

## הטיפוסים הבסיסיים

| טיפוס | מה הוא מכיל | טווח |
|---|---|---|
| `int` | מספרים שלמים עם סימן | −2147483648 … 2147483647 |
| `sint` | **בדיוק אותו טיפוס כמו `int`.** השתמשו בו כשרוצים לומר במפורש "עם סימן". | כמו `int` |
| `uint` | מספרים שלמים בלי סימן | 0 … 4294967295 |
| `float`, `float32`, `float64` | מספרים עשרוניים (נקודה צפה) ([פרק 13](13-floats-and-conversions.md)) | |
| `bool` | `true` או `false` | |
| `string` | טקסט, למשל `"hello\n"` | |
| `Error` וסוגיו | שגיאות ([פרק 14](14-errors.md)) | |
| `void` | אין ערך; משמש רק כטיפוס ההחזרה של פונקציה | |

אפשר לערבב בחופשיות בין `int` ל-`sint`:

</div>

```actionscript
package main;

public void main() {
    sint a = -5;
    int b = a;      // fine: same type
    sint c = b;     // fine
    trace(a, b, c);
}
```

<div dir="rtl">

תווי בריחה במחרוזות: `\n` (שורה חדשה), `\t` (טאב), `\r`, `\\` ו-`\"`.

## השמה

</div>

```actionscript
package main;

public void main() {
    int x = 1;
    x = 10;     // 10
    x += 5;     // 15
    x -= 3;     // 12
    x *= 2;     // 24
    x /= 5;     // 4
    x %= 3;     // 1
    x++;        // 2
    x--;        // 1
    trace(x);
}
```

<div dir="rtl">

- השמה היא הוראה. אי אפשר לכתוב `a = b = 0;` או `if (x = 1)`.
- `++` ו-`--` הן הוראות שבאות אחרי משתנה (`i++;`). אי אפשר להשתמש בהן בתוך ביטוי
  גדול יותר.
- פרמטרים של פונקציה הם לקריאה בלבד. כדי לשנות אחד, העתיקו אותו קודם למשתנה מקומי.

## טווח הכרה (scope)

משתנה קיים מרגע ההצהרה עליו ועד סוף הבלוק שמכיל אותו. משתנה שהוצהר בכותרת של
`for` קיים רק בתוך הלולאה הזו.

</div>

```actionscript
package main;

public void main() {
    if (true) {
        int inner = 1;
        trace(inner);
    }
    // 'inner' is not visible here

    for (int i = 0; i < 2; i++) {
        trace(i);
    }
    for (int i = 0; i < 2; i++) {   // fine: the first 'i' is gone
        trace(i);
    }
}
```

<div dir="rtl">

גרסה 0.1 אינה מתירה **הסתרה (shadowing)**. משתנה חדש אינו יכול להשתמש שוב בשם של
משתנה, פרמטר או פונקציה הנראים במקום:

</div>

```actionscript
package main;

public void main() {
    int x = 1;
    if (true) {
        int x = 2; // error: ASX-NAME-002
    }
}
```

<div dir="rtl">

## עם סימן ובלי סימן לעולם אינם מתערבבים בשקט

RuneScript אינה הופכת בשקט מספר שלם עם סימן למספר בלי סימן. המהדר בודק את
*הטיפוס* של הערך, ולא את הערך עצמו. משתנה עם סימן יכול להכיל מספר שלילי, ולכן
שמירתו ב-`uint` היא שגיאה גם אם במקרה הערך חיובי:

</div>

```actionscript
package main;

public int main() {
    sint signedValue = -10;
    uint unsignedValue = signedValue; // error: ASX-TYPE-004
    return 0;
}
```

```text
main.rsc:5:26: error ASX-TYPE-004: Cannot assign signed integer 'sint' to unsigned integer 'uint'.
        uint unsignedValue = signedValue;
                             ^^^^^^^^^^^
  help: Use an explicit checked conversion, (uint) x, or change the destination type.
```

<div dir="rtl">

ליטרלים מספריים גמישים יותר. ליטרל מקבל את הטיפוס של מה שמקבל אותו, כל עוד הערך
נכנס בו:

</div>

```actionscript
package main;

public void main() {
    uint a = 42;            // fine: 42 fits in uint
    uint b = 6 * 7;         // fine: constants are computed at compile time
    int c = -2147483648;    // fine: the smallest int
    trace(a, b, c);
}
```

```actionscript
package main;

public void main() {
    uint u = -1;            // error: ASX-TYPE-007
}
```

```actionscript
package main;

public void main() {
    int big = 3000000000;   // error: ASX-TYPE-006
}
```

<div dir="rtl">

גם הכיוון ההפוך (`uint` אל `int`) נדחה, וכך גם ערבוב של `int` ו-`uint` בפעולה אחת.
כדי להמיר, כתבו את ההמרה במפורש: `(uint) signedValue` בודקת את הערך בזמן הריצה
וזורקת שגיאה אם הוא שלילי. ראו [מספרים עשרוניים והמרות](13-floats-and-conversions.md),
שם מוסברים גם `float`, `float32` ו-`float64`.

הבא: [אופרטורים](04-operators.md)

</div>
