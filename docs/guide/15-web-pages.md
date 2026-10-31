# 15. Web pages with `rune:dom`

`rune:dom` builds web pages from **components**: functions that return an
`Element`, described by a **template** of HTML or SVG. State lives in
**runes** (signals, as in SolidJS). When a rune changes, exactly the text,
attribute, or `$if` that read it is updated. There is no virtual DOM and
nothing is re-rendered.

Templates are checked when the program compiles. Every element and
attribute must exist in the HTML4, HTML5, or SVG element list, static
attribute values must suit the attribute, and every `{{ }}` value must have
a type the attribute or text can take. Text is always escaped.

`rune:dom` needs a browser, so these programs run only on the JavaScript
target. `runescript dev` serves one with live reload
([chapter 11](11-javascript.md#live-reload-runescript-dev)), and
`runescript build` compiles it for production. `runescript run` reports
`ASX-TARGET-002`. The [web examples](../../examples/web/) are ready to run:

```sh
runescript dev examples/web/counter/app.rsc
```

## A first component

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

`use "rune:dom";` makes the types `Element` and `Signal<T>` and the
function `signal` available as they are, and `renderApp` through the
qualifier `dom`.

## Components and templates

- A component is a function that returns `Element`. Its code comes first,
  and its last statement is `template { … }`: the element the markup
  describes is what the function returns (`ASX-TMPL-001` otherwise).
- A template has exactly one root element.
- Elements are written as in HTML: `<p>…</p>`, void elements such as
  `<br>` and `<img …>` without a closing tag, and any element as `<div/>`.
  Inside `<svg>`, every element may be self-closed.
- White space that contains a line break is removed at the start and end
  of text and collapsed to one space inside it, so templates can be
  indented freely. Other white space is kept.
- `{` and `}` are reserved in template text: `{{` starts an
  interpolation, and a lone `}` ends the template. Write a literal brace
  as `{{ "{" }}`.
- `<!-- comments -->` are allowed and left out of the page.

## Elements and attributes

Element names are lowercase in HTML. Inside `<svg>`, the names come from
SVG, where they are case-sensitive (`linearGradient`, `viewBox`), and
inside `<foreignObject>` they are HTML again. An unknown element
(`ASX-TMPL-002`) or attribute (`ASX-TMPL-003`) is an error, and the help
says when the name is in the other namespace or differs only in case.

Elements and attributes that HTML5 or SVG 2 made obsolete still work, but
they give warning `ASX-WARN-007` with the replacement: `<center>` suggests
CSS, and `bgcolor` and `xlink:href` suggest CSS and `href`.

A static value must suit the attribute (`ASX-TMPL-004`):

| Attribute | Static value |
|---|---|
| Boolean, e.g. `disabled`, `required` | No value: `disabled` alone (or `disabled="disabled"`) |
| Keyword, e.g. `type`, `method`, `dir` | One of its keywords: `<input type="email">` |
| Number, e.g. `colspan`, `maxlength`, `tabindex` | Digits (a sign only where negatives make sense) |
| Length (SVG), e.g. `r`, `width` | A number with an optional unit or `%`: `4`, `10px`, `50%` |
| URL, e.g. `href`, `src` | Any URL except one that runs script (`javascript:`, `vbscript:`, `data:`) |
| Event handler, e.g. `onclick` | Never text: a RuneScript function, below |
| Anything else, including `data-*` and `aria-*` | Text |

A value can also be an interpolation, `name={{ expression }}`. It is a
binding: it is set again whenever a rune it reads changes. Its type must
suit the attribute (`ASX-TMPL-005`):

| Attribute | Type of `{{ }}` |
|---|---|
| Boolean | `bool`: `disabled={{ count() == 0 }}` adds or removes the attribute |
| Integer | `int` (or `uint` for non-negative ones such as `colspan`) |
| Number | a float type |
| Length (SVG) | a number, or a string such as `"10px"` |
| Event handler | a function taking nothing and returning `void` |
| Anything else | `string`. Convert other values explicitly: `title={{ (string) n }}` |

A bound keyword attribute is checked when it is set: a value it doesn't
allow throws a `TypeError` (`ASX-RUN-010`). So does a bound URL that would
run script, so neither text nor links can inject code.

## Runes

A rune holds a number, a `bool`, or a `string`:

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

- `signal(0)` creates a rune. Its value type comes from the declaration
  (`Signal<int>`), or from the initial value.
- `count()` and `count.get()` read it. `count.set(value)` changes it, and
  the value must have the rune's type.
- A rune is not text: `{{ count }}` is an error, and the help says to
  write `{{ count() }}`.
- Reading a rune inside a `{{ }}`, a bound attribute, or a `$if`
  condition ties that piece of the page to the rune. Reading it anywhere
  else, such as in an event handler, doesn't.
- Runes of other types, such as errors or functions, aren't supported
  (`ASX-TYPE-025`).

## Events

An event attribute (`onclick`, `oninput`, `onsubmit`, …) takes a
RuneScript function that takes nothing and returns `void`: a nested
function, `onclick={{ increment }}`, or a lambda,
`onclick={{ () => count.set(0) }}`.

Nested functions and lambdas may use the **runes** of the component they
are in. Other variables stay out of reach, as everywhere in RuneScript
([chapter 6](06-functions.md#function-boundaries)), and give
`ASX-NAME-009`. A rune is a shared cell, so handlers change the
component's state through `set`.

A runtime error in a handler that nothing catches is reported like any
other uncaught RuneScript error, and `runescript dev` shows it in its
overlay.

## `$if`

`$if (condition) … $elseif (condition) … $else … $endif` shows the first
branch whose condition holds. `$elseif` and `$else` are optional. When the
chosen branch changes, the old branch is removed with all its bindings and
the new one is built. While it stays the same, nothing is rebuilt.

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

## `$foreach`

`$foreach (Type name in array) … $endforeach` repeats its content for every
element of an array, with `name` holding the element; the header is
written as in a `for each` loop. The whole list is built again when a rune
that the array expression reads changes. Changes to the array itself
(`push`, `pop`, setting an element) are not tracked, so keep the list's
state in a rune and compute the array from it, as here:

```actionscript
package main;

use "rune:dom";

public Array<string> first(Array<string> all, int n) {
    Array<string> out = [];
    for (int i = 0; i < n && i < all.length; i++) {
        out.push(all[i]);
    }
    return out;
}

public Element Tasks() {
    Array<string> tasks = ["Write the docs", "Test the list", "Ship it"];
    Signal<int> shown = signal(1);

    template {
        <div>
            <ul>
                $foreach (string task in first(tasks, shown()))
                    <li>{{ task }}</li>
                $endforeach
            </ul>
            <button type="button" onclick={{ () => shown.set(shown() + 1) }}>Show more</button>
        </div>
    }
}

protected int main() {
    dom.renderApp(Tasks, "#app");
    return 0;
}
```

## Child components

A component is a function, so another component embeds it with an
interpolation whose value is an `Element`: `{{ Badge("new") }}`. Each
call is a separate instance with its own runes. Components in another
file are reached through that file's qualifier, like any function
([chapter 12](12-packages-and-files.md)):

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

`examples/web/multi-file` is this program, with a header in a third file.

## Mounting the app

There is no router. `dom.renderApp(component, selector)` calls the
component and puts its element into the first element the CSS selector
matches, replacing what was there. Any selector works, but one other than
a single id, such as `".app"`, gives warning `ASX-WARN-008`, since it may
match several elements or none. When nothing matches, `renderApp` throws
an `Error` (`ASX-RUN-009`).

The page needs that element and needs to load the compiled module. This
`index.html` works with both `runescript dev` and `runescript build`,
which writes `main.js` next to it:

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

Without an `index.html`, `runescript dev` generates a page with
`<div id="app">`.

## A checked mistake

Templates fail at compile time, not in the browser:

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

## Not yet available

- Component tags with properties (`<Badge label="new"/>`): embed
  components with `{{ Badge("new") }}`.
- Event objects, so a handler can't read an input's value yet, and
  two-way binding.
- Raw HTML insertion, MathML, custom elements, and server-side rendering.
- Hot module replacement: `runescript dev` reloads the page.

Next: [Text and lists](16-text-and-arrays.md)
