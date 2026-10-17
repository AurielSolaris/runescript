<!-- source: docs/guide/15-web-pages.md @ 3cd2c1f -->

<div dir="rtl">

> זהו תרגום של [המקור באנגלית](../../../../docs/guide/15-web-pages.md). אם יש הבדל, הגרסה האנגלית היא הקובעת. התרגום נעשה בעזרת בינה מלאכותית ונבדק בידי המחבר.

# 15. דפי רשת עם `rune:dom`

`rune:dom` בונה דפי רשת מ**רכיבים**: פונקציות שמחזירות `Element`, ומתוארות ב**תבנית**
של HTML או SVG. המצב נשמר ב-**runes** (אותות, כמו ב-SolidJS). כש-rune משתנה, מתעדכנים
בדיוק הטקסט, המאפיין או ה-`$if` שקראו אותו. אין DOM וירטואלי ושום דבר לא מרונדר מחדש.

התבניות נבדקות כשהתוכנית מתהדרת. כל אלמנט ומאפיין חייבים להופיע ברשימת האלמנטים של HTML4,
HTML5 או SVG, ערכים קבועים של מאפיינים חייבים להתאים למאפיין, ולכל ערך `{{ }}` חייב
להיות טיפוס שהמאפיין או הטקסט יכולים לקבל. טקסט תמיד עובר escape.

`rune:dom` צריך דפדפן, ולכן התוכניות האלה רצות רק ביעד JavaScript. `runescript dev`
מגישה אותן עם טעינה מחדש חיה ([פרק 11](11-javascript.md)), ו-`runescript build` מהדרת
אותן לסביבת הייצור. `runescript run` מדווחת `ASX-TARGET-002`.
[דוגמאות הרשת](../../../../examples/web/) מוכנות להרצה:

</div>

```sh
runescript dev examples/web/counter/app.rsc
```

<div dir="rtl">

## רכיב ראשון

</div>

```actionscript
package main;

use "rune:dom";

public Element Hello() {
    template {
        <article class="card">
            <h1>Hello from RuneScript</h1>
            <p>Text is escaped: {{ "<b>not bold</b>" }} &amp; character references work.</p>
            <a href="https://github.com/AurielSolaris/runescript" target="_blank">The repository</a>
        </article>
    }
}

protected int main() {
    dom.renderApp(Hello, "#app");
    return 0;
}
```

<div dir="rtl">

`use "rune:dom";` הופך את הטיפוסים `Element` ו-`Signal<T>` ואת הפונקציה `signal` לזמינים
כמו שהם, ואת `renderApp` דרך המזהה `dom`.

## רכיבים ותבניות

- רכיב הוא פונקציה שמחזירה `Element`. הקוד שלו בא קודם, וההוראה האחרונה שלו היא
  `template { … }`: האלמנט שהתבנית מתארת הוא מה שהפונקציה מחזירה (אחרת `ASX-TMPL-001`).
- לתבנית יש בדיוק אלמנט שורש אחד.
- אלמנטים נכתבים כמו ב-HTML: `<p>…</p>`, אלמנטים ריקים כמו `<br>` ו-`<img …>` בלי תג
  סוגר, וכל אלמנט כ-`<div/>`. בתוך `<svg>`, כל אלמנט יכול להיסגר בעצמו.
- רווח לבן שכולל שבירת שורה מוסר בתחילת הטקסט ובסופו ומכווץ לרווח אחד בתוכו, כך שאפשר
  להזיח תבניות בחופשיות. רווח לבן אחר נשמר.
- `{` ו-`}` שמורים בטקסט של תבנית: `{{` פותח אינטרפולציה, ו-`}` בודד מסיים את התבנית.
  סוגר מסולסל ממשי נכתב `{{ "{" }}`.
- מותר לכתוב `<!-- הערות -->`, והן לא מופיעות בדף.

## אלמנטים ומאפיינים

ב-HTML, שמות האלמנטים כתובים באותיות קטנות. בתוך `<svg>` השמות באים מ-SVG, שבה יש
חשיבות לרישיות (`linearGradient`, `viewBox`), ובתוך `<foreignObject>` הם שוב HTML. אלמנט
לא מוכר (`ASX-TMPL-002`) או מאפיין לא מוכר (`ASX-TMPL-003`) הם שגיאה, והעזרה אומרת מתי
השם נמצא במרחב השמות האחר או שונה רק ברישיות.

אלמנטים ומאפיינים ש-HTML5 או SVG 2 הפכו למיושנים עדיין עובדים, אבל הם נותנים את האזהרה
`ASX-WARN-007` עם החלופה: `<center>` מציע CSS, ו-`bgcolor` ו-`xlink:href` מציעים CSS
ו-`href`.

ערך קבוע חייב להתאים למאפיין (`ASX-TMPL-004`):

| מאפיין | ערך קבוע |
|---|---|
| בוליאני, למשל `disabled`, `required` | בלי ערך: `disabled` לבד (או `disabled="disabled"`) |
| מילת מפתח, למשל `type`, `method`, `dir` | אחת ממילות המפתח שלו: `<input type="email">` |
| מספר, למשל `colspan`, `maxlength`, `tabindex` | ספרות (סימן רק היכן שיש משמעות לערכים שליליים) |
| אורך (SVG), למשל `r`, `width` | מספר עם יחידה או `%` אופציונליים: `4`, `10px`, `50%` |
| כתובת URL, למשל `href`, `src` | כל כתובת, מלבד כזו שמריצה סקריפט (`javascript:`, `vbscript:`, `data:`) |
| מטפל אירועים, למשל `onclick` | לעולם לא טקסט: פונקציה של RuneScript, ראו בהמשך |
| כל השאר, כולל `data-*` ו-`aria-*` | טקסט |

ערך יכול להיות גם אינטרפולציה, `name={{ expression }}`. זו קשירה: היא נקבעת מחדש בכל פעם
ש-rune שהיא קראה משתנה. הטיפוס שלה חייב להתאים למאפיין (`ASX-TMPL-005`):

| מאפיין | הטיפוס של `{{ }}` |
|---|---|
| בוליאני | `bool`: `disabled={{ count() == 0 }}` מוסיף את המאפיין או מסיר אותו |
| מספר שלם | `int` (או `uint` למאפיינים אי-שליליים כמו `colspan`) |
| מספר | טיפוס עשרוני |
| אורך (SVG) | מספר, או מחרוזת כמו `"10px"` |
| מטפל אירועים | פונקציה שלא מקבלת דבר ומחזירה `void` |
| כל השאר | `string`. המירו ערכים אחרים במפורש: `title={{ (string) n }}` |

מאפיין מילת מפתח קשור נבדק כשהוא נקבע: ערך שהוא לא מתיר זורק `TypeError`
(`ASX-RUN-010`). כך גם כתובת URL קשורה שתריץ סקריפט, ולכן לא טקסט ולא קישורים יכולים
להחדיר קוד.

## Runes

rune מחזיק מספר, `bool` או `string`:

</div>

```actionscript
package main;

use "rune:dom";

public Element Counter() {
    Signal<int> count = signal(0);

    private void increment() {
        count.set(count() + 1);
    }

    template {
        <section>
            <h2>Clicked {{ count() }} times</h2>
            <button type="button" onclick={{ increment }}>+1</button>
            <button type="button" disabled={{ count() == 0 }} onclick={{ () => count.set(0) }}>Reset</button>
        </section>
    }
}

protected int main() {
    dom.renderApp(Counter, "#app");
    return 0;
}
```

<div dir="rtl">

- `signal(0)` יוצר rune. טיפוס הערך שלו בא מההצהרה (`Signal<int>`), או מהערך ההתחלתי.
- `count()` ו-`count.get()` קוראים אותו. `count.set(value)` משנה אותו, והערך חייב להיות
  מהטיפוס של ה-rune.
- rune אינו טקסט: `{{ count }}` היא שגיאה, והעזרה אומרת לכתוב `{{ count() }}`.
- קריאת rune בתוך `{{ }}`, מאפיין קשור או תנאי של `$if` קושרת את החלק הזה של הדף ל-rune.
  קריאה במקום אחר, למשל במטפל אירועים, לא קושרת.
- runes מטיפוסים אחרים, כמו שגיאות או פונקציות, אינם נתמכים (`ASX-TYPE-025`).

## אירועים

מאפיין אירוע (`onclick`, `oninput`, `onsubmit`, …) מקבל פונקציה של RuneScript שלא מקבלת
דבר ומחזירה `void`: פונקציה מקוננת, `onclick={{ increment }}`, או ביטוי למבדה,
`onclick={{ () => count.set(0) }}`.

פונקציות מקוננות וביטויי למבדה רשאים להשתמש ב-**runes** של הרכיב שבתוכו הם נמצאים.
משתנים אחרים נשארים מחוץ להישג יד, כמו בכל RuneScript ([פרק 6](06-functions.md)),
ונותנים `ASX-NAME-009`. rune הוא תא משותף, ולכן מטפלים משנים את מצב הרכיב דרך `set`.

שגיאת זמן ריצה במטפל שאיש לא תופס מדווחת כמו כל שגיאת RuneScript אחרת שלא נתפסה,
ו-`runescript dev` מציגה אותה בשכבת-העל שלה.

## `$if`

`$if (condition) … $elseif (condition) … $else … $endif` מציג את הענף הראשון שהתנאי שלו
מתקיים. `$elseif` ו-`$else` אופציונליים. כשהענף הנבחר משתנה, הענף הישן מוסר עם כל
הקשירות שלו והחדש נבנה. כל עוד הוא נשאר אותו ענף, שום דבר לא נבנה מחדש.

</div>

```actionscript
package main;

use "rune:dom";

public Element Details() {
    Signal<bool> open = signal(false);
    Signal<int> views = signal(0);

    private void toggle() {
        if (!open()) {
            views.set(views() + 1);
        }
        open.set(!open());
    }

    template {
        <div>
            <button type="button" aria-expanded={{ (string) open() }} onclick={{ toggle }}>Details</button>
            $if (open())
                <p>Opened {{ views() }} time(s).</p>
            $elseif (views() > 0)
                <p><em>Closed again.</em></p>
            $else
                <p><em>Not opened yet.</em></p>
            $endif
        </div>
    }
}

protected int main() {
    dom.renderApp(Details, "#app");
    return 0;
}
```

<div dir="rtl">

## רכיבי ילד

רכיב הוא פונקציה, ולכן רכיב אחר משבץ אותו באינטרפולציה שהערך שלה הוא `Element`:
`{{ Badge("new") }}`. כל קריאה היא מופע נפרד עם runes משלו. לרכיבים בקובץ אחר מגיעים דרך
המזהה של הקובץ ההוא, כמו לכל פונקציה ([פרק 12](12-packages-and-files.md)):

</div>

```actionscript
// main.rsc
package main;

use "rune:dom";
use "./app.rsc";

protected int main() {
    dom.renderApp(app.App, "#app");
    return 0;
}
```

```actionscript
// app.rsc
package app;

use "rune:dom";

public Element Badge(string label) {
    template {
        <span class="badge">{{ label }}</span>
    }
}

public Element App() {
    template {
        <main>
            <h1>Shop {{ Badge("new") }}</h1>
        </main>
    }
}
```

<div dir="rtl">

`examples/web/multi-file` היא התוכנית הזו, עם כותרת בקובץ שלישי.

## הרכבת האפליקציה

אין נתב. `dom.renderApp(component, selector)` קוראת לרכיב ומכניסה את האלמנט שלו לאלמנט
הראשון שבורר ה-CSS מתאים לו, במקום מה שהיה שם. כל בורר עובד, אבל בורר שאינו id יחיד,
כמו `".app"`, נותן את האזהרה `ASX-WARN-008`, כי הוא עשוי להתאים לכמה אלמנטים או לאף אחד.
כשאין התאמה, `renderApp` זורקת `Error` (`ASX-RUN-009`).

הדף צריך את האלמנט הזה וצריך לטעון את המודול המהודר. ה-`index.html` הזה עובד גם עם
`runescript dev` וגם עם `runescript build`, שכותבת את `main.js` לידו:

</div>

```html
<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>Shop</title></head>
<body>
  <div id="app"></div>
  <script type="module">
    import { run } from "./main.js";
    run();
  </script>
</body>
</html>
```

<div dir="rtl">

בלי `index.html`, `runescript dev` יוצרת דף עם `<div id="app">`.

## טעות שנבדקת

תבניות נכשלות בזמן ההידור, לא בדפדפן:

</div>

```actionscript
package main;

use "rune:dom";

public Element App() {
    // error: ASX-TMPL-002 (the typo below)
    template {
        <dvi>Typo</dvi>
    }
}

protected int main() {
    dom.renderApp(App, "#app");
    return 0;
}
```

```text
prog.rsc:8:10: error ASX-TMPL-002: unknown element <dvi>
```

<div dir="rtl">

## עדיין לא זמין

- `$foreach`, שמחכה למערכים.
- תגי רכיבים עם מאפיינים (`<Badge label="new"/>`): משבצים רכיבים עם `{{ Badge("new") }}`.
- אובייקטי אירוע, ולכן מטפל עדיין לא יכול לקרוא את הערך של שדה קלט, וקשירה דו-כיוונית.
- הכנסת HTML גולמי, MathML, אלמנטים מותאמים אישית ורינדור בצד השרת.
- החלפת מודולים חמה: `runescript dev` טוענת את הדף מחדש.

הבא: [טקסט ורשימות](16-text-and-arrays.md)

</div>
