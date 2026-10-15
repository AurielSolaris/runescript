<!-- source: docs/guide/05-control-flow.md @ 52427d9 -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/05-control-flow.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 5. בקרת זרימה

תנאים חייבים להיות `bool`. כתבו `if (n != 0)`, ולא `if (n)`. גופים תמיד נכתבים
בסוגריים מסולסלים.

## if / else if / else

</div>

```actionscript
package main;

public string classify(int n) {
    if (n < 0) {
        return "negative";
    } else if (n == 0) {
        return "zero";
    } else {
        return "positive";
    }
}

public void main() {
    trace(classify(-5), classify(0), classify(9));   // negative zero positive
}
```

<div dir="rtl">

## while

</div>

```actionscript
package main;

public void main() {
    int n = 3;
    while (n > 0) {
        trace(n);
        n--;
    }
    trace("liftoff");
}
```

<div dir="rtl">

## for

`for (init; condition; step) { … }`

</div>

```actionscript
package main;

public void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

public void main() {
    sayHi("Hello world", 3);
}
```

<div dir="rtl">

כל אחד משלושת החלקים הוא רשות. בלי תנאי, הלולאה רצה עד `break` או `return`:

</div>

```actionscript
package main;

public void main() {
    int tries = 0;
    for (;;) {
        tries++;
        if (tries == 3) {
            break;
        }
    }
    trace("stopped after", tries);
}
```

<div dir="rtl">

חלק האתחול יכול להצהיר על משתנה (שקיים רק בתוך הלולאה), להשים ערך במשתנה קיים, או
לקרוא לפונקציה. הצעד יכול להיות השמה, `++`/`--`, או קריאה.

## break ו-continue

`break` יוצאת מהלולאה הפנימית ביותר. `continue` מדלגת לאיטרציה הבאה. בלולאת `for`,
גם אחרי `continue` הצעד מתבצע.

</div>

```actionscript
package main;

public void main() {
    for (int i = 0; i < 10; i++) {
        if (i % 2 == 0) {
            continue;       // skip even numbers
        }
        if (i > 7) {
            break;          // stop at 9
        }
        trace(i);           // 1 3 5 7
    }
}
```

<div dir="rtl">

`break` או `continue` מחוץ ללולאה הם שגיאה (`ASX-FLOW-001`).

## החזרת ערך בכל מסלול

פונקציה שמחזירה ערך חייבת להחזיר בכל מסלול. המהדר מקבל גוף שמסתיים ב-`return` (או
ב-`throw`, ראו [שגיאות](14-errors.md)), או בשרשרת `if`/`else` שכל ענף בה מחזיר:

</div>

```actionscript
package main;

public int sign(int n) {
    if (n < 0) {
        return -1;
    } else if (n == 0) {
        return 0;
    } else {
        return 1;
    }
}

public void main() {
    trace(sign(-8), sign(0), sign(8));
}
```

<div dir="rtl">

המהדר אינו מתייחס ללולאה כמבטיחה החזרה. הוסיפו `return` אחרון אחרי הלולאה:

</div>

```actionscript
package main;

public int firstSquareAbove(int limit) {
    for (int i = 0; i < 1000; i++) {
        if (i * i > limit) {
            return i;
        }
    }
    return -1;
}

public void main() {
    trace(firstSquareAbove(50));   // 8
}
```

<div dir="rtl">

הבא: [פונקציות](06-functions.md)

</div>
