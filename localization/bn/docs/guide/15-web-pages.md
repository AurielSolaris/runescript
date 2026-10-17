<!-- source: docs/guide/15-web-pages.md @ 362a461 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/15-web-pages.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ১৫. `rune:dom` দিয়ে ওয়েব পেজ

`rune:dom` **কম্পোনেন্ট** থেকে ওয়েব পেজ তৈরি করে: এমন ফাংশন যা একটি `Element` ফেরত
দেয়, এবং যা HTML বা SVG-র একটি **টেমপ্লেট** দিয়ে বর্ণিত। অবস্থা (state) থাকে
**রুনে** (সিগন্যাল, SolidJS-এর মতো)। একটি রুন বদলালে ঠিক সেই লেখা, অ্যাট্রিবিউট বা
`$if` আপডেট হয়, যা সেটি পড়েছিল। কোনো ভার্চুয়াল DOM নেই, আর কিছুই নতুন করে রেন্ডার
হয় না।

প্রোগ্রাম কম্পাইলের সময় টেমপ্লেট যাচাই হয়। প্রতিটি এলিমেন্ট ও অ্যাট্রিবিউটকে HTML4,
HTML5 বা SVG এলিমেন্ট-তালিকায় থাকতে হবে, স্থির অ্যাট্রিবিউট মানকে অ্যাট্রিবিউটের
উপযুক্ত হতে হবে, আর প্রতিটি `{{ }}` মানের এমন টাইপ থাকতে হবে যা অ্যাট্রিবিউট বা
লেখা নিতে পারে। লেখা সবসময় এস্কেপ করা হয়।

`rune:dom`-এর একটি ব্রাউজার দরকার, তাই এই প্রোগ্রামগুলো কেবল JavaScript টার্গেটে
চলে। `runescript dev` লাইভ রিলোডসহ একটি পরিবেশন করে ([অধ্যায় ১১](11-javascript.md)),
আর `runescript build` এটিকে প্রোডাকশনের জন্য কম্পাইল করে। `runescript run` জানায়
`ASX-TARGET-002`। [ওয়েব উদাহরণগুলো](../../../../examples/web/) চালানোর জন্য তৈরি:

```sh
runescript dev examples/web/counter/app.rsc
```

## প্রথম কম্পোনেন্ট

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

`use "rune:dom";` `Element` ও `Signal<T>` টাইপ এবং `signal` ফাংশনকে যেমন আছে তেমনই
পাওয়া যায় এমন করে, আর `renderApp`-কে কোয়ালিফায়ার `dom`-এর মাধ্যমে।

## কম্পোনেন্ট ও টেমপ্লেট

- একটি কম্পোনেন্ট এমন একটি ফাংশন, যা `Element` ফেরত দেয়। এর কোড আগে আসে, আর এর শেষ
  স্টেটমেন্ট হলো `template { … }`: মার্কআপ যে এলিমেন্ট বর্ণনা করে, ফাংশনটি সেটিই
  ফেরত দেয় (নইলে `ASX-TMPL-001`)।
- একটি টেমপ্লেটের ঠিক একটি রুট এলিমেন্ট থাকে।
- এলিমেন্ট লেখা হয় HTML-এর মতো: `<p>…</p>`, `<br>` ও `<img …>`-এর মতো ভয়েড
  এলিমেন্ট বন্ধ করার ট্যাগ ছাড়া, আর যেকোনো এলিমেন্ট `<div/>` হিসেবে। `<svg>`-এর
  ভেতরে প্রতিটি এলিমেন্ট নিজেই বন্ধ হতে পারে।
- লাইন-ব্রেক থাকা হোয়াইট স্পেস লেখার শুরু ও শেষে বাদ যায় এবং ভেতরে একটি স্পেসে
  সংকুচিত হয়, তাই টেমপ্লেট ইচ্ছেমতো ইন্ডেন্ট করা যায়। অন্য হোয়াইট স্পেস থেকে যায়।
- টেমপ্লেটের লেখায় `{` ও `}` সংরক্ষিত: `{{` একটি ইন্টারপোলেশন শুরু করে, আর একা একটি
  `}` টেমপ্লেট শেষ করে। আক্ষরিক বন্ধনী লিখুন `{{ "{" }}` হিসেবে।
- `<!-- মন্তব্য -->` লেখা যায়, আর পেজে সেগুলো থাকে না।

## এলিমেন্ট ও অ্যাট্রিবিউট

HTML-এ এলিমেন্টের নাম ছোট হাতের অক্ষরে। `<svg>`-এর ভেতরে নামগুলো আসে SVG থেকে, যেখানে
বড়-ছোট হাতের অক্ষর আলাদা (`linearGradient`, `viewBox`), আর `<foreignObject>`-এর ভেতরে
সেগুলো আবার HTML। অজানা এলিমেন্ট (`ASX-TMPL-002`) বা অ্যাট্রিবিউট (`ASX-TMPL-003`) একটি
এরর, আর নামটি অন্য নেমস্পেসে থাকলে বা কেবল অক্ষরের ছোট-বড়তে আলাদা হলে help তা জানায়।

যেসব এলিমেন্ট ও অ্যাট্রিবিউট HTML5 বা SVG 2 অপ্রচলিত করেছে, সেগুলো এখনো কাজ করে,
কিন্তু বিকল্পসহ `ASX-WARN-007` সতর্কবার্তা দেয়: `<center>` CSS-এর পরামর্শ দেয়, আর
`bgcolor` ও `xlink:href` CSS ও `href`-এর।

একটি স্থির মানকে অ্যাট্রিবিউটের উপযুক্ত হতে হবে (`ASX-TMPL-004`):

| অ্যাট্রিবিউট | স্থির মান |
|---|---|
| বুলিয়ান, যেমন `disabled`, `required` | কোনো মান নয়: শুধু `disabled` (অথবা `disabled="disabled"`) |
| কিওয়ার্ড, যেমন `type`, `method`, `dir` | এর কিওয়ার্ডগুলোর একটি: `<input type="email">` |
| সংখ্যা, যেমন `colspan`, `maxlength`, `tabindex` | অঙ্ক (চিহ্ন কেবল যেখানে ঋণাত্মক মান অর্থবহ) |
| দৈর্ঘ্য (SVG), যেমন `r`, `width` | ঐচ্ছিক একক বা `%` সহ একটি সংখ্যা: `4`, `10px`, `50%` |
| URL, যেমন `href`, `src` | যেকোনো URL, স্ক্রিপ্ট চালায় এমনটি ছাড়া (`javascript:`, `vbscript:`, `data:`) |
| ইভেন্ট হ্যান্ডলার, যেমন `onclick` | কখনো লেখা নয়: একটি RuneScript ফাংশন, নিচে দেখুন |
| অন্য সবকিছু, `data-*` ও `aria-*` সহ | লেখা |

একটি মান ইন্টারপোলেশনও হতে পারে, `name={{ expression }}`। এটি একটি বাইন্ডিং: যখনই
এটি যে রুন পড়েছে তা বদলায়, এটি আবার সেট হয়। এর টাইপকে অ্যাট্রিবিউটের উপযুক্ত হতে হবে
(`ASX-TMPL-005`):

| অ্যাট্রিবিউট | `{{ }}`-এর টাইপ |
|---|---|
| বুলিয়ান | `bool`: `disabled={{ count() == 0 }}` অ্যাট্রিবিউটটি যোগ করে বা সরায় |
| পূর্ণসংখ্যা | `int` (অথবা `colspan`-এর মতো অঋণাত্মকগুলোর জন্য `uint`) |
| সংখ্যা | একটি ফ্লোট টাইপ |
| দৈর্ঘ্য (SVG) | একটি সংখ্যা, অথবা `"10px"`-এর মতো একটি স্ট্রিং |
| ইভেন্ট হ্যান্ডলার | এমন একটি ফাংশন, যা কিছু নেয় না এবং `void` ফেরত দেয় |
| অন্য সবকিছু | `string`। অন্য মান স্পষ্টভাবে রূপান্তর করুন: `title={{ (string) n }}` |

একটি বাঁধা কিওয়ার্ড অ্যাট্রিবিউট সেট হওয়ার সময় যাচাই হয়: এটি যে মান অনুমোদন করে না,
তা একটি `TypeError` (`ASX-RUN-010`) থ্রো করে। স্ক্রিপ্ট চালাবে এমন একটি বাঁধা URL-ও
তাই করে, ফলে লেখা বা লিঙ্ক কোনোটিই কোড ঢোকাতে পারে না।

## রুন

একটি রুন একটি সংখ্যা, একটি `bool` বা একটি `string` রাখে:

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

- `signal(0)` একটি রুন তৈরি করে। এর মানের টাইপ আসে ঘোষণা থেকে (`Signal<int>`), অথবা
  প্রাথমিক মান থেকে।
- `count()` ও `count.get()` এটি পড়ে। `count.set(value)` এটি বদলায়, আর মানটির টাইপ হতে
  হবে রুনটির টাইপ।
- একটি রুন লেখা নয়: `{{ count }}` একটি এরর, আর help বলে `{{ count() }}` লিখতে।
- একটি `{{ }}`, একটি বাঁধা অ্যাট্রিবিউট বা একটি `$if` শর্তের ভেতরে রুন পড়লে পেজের সেই
  অংশ রুনটির সঙ্গে বাঁধা পড়ে। অন্য কোথাও পড়লে, যেমন একটি ইভেন্ট হ্যান্ডলারে, তা হয় না।
- অন্য টাইপের রুন, যেমন এরর বা ফাংশন, সমর্থিত নয় (`ASX-TYPE-025`)।

## ইভেন্ট

একটি ইভেন্ট অ্যাট্রিবিউট (`onclick`, `oninput`, `onsubmit`, …) এমন একটি RuneScript ফাংশন
নেয়, যা কিছু নেয় না এবং `void` ফেরত দেয়: একটি নেস্টেড ফাংশন, `onclick={{ increment }}`,
অথবা একটি ল্যাম্বডা, `onclick={{ () => count.set(0) }}`।

নেস্টেড ফাংশন ও ল্যাম্বডা যে কম্পোনেন্টের ভেতরে আছে তার **রুন** ব্যবহার করতে পারে। অন্য
ভেরিয়েবল নাগালের বাইরে থাকে, RuneScript-এর সব জায়গার মতো
([অধ্যায় ৬](06-functions.md)), এবং `ASX-NAME-009` দেয়। একটি রুন একটি ভাগ করা ঘর
(cell), তাই হ্যান্ডলার `set` দিয়ে কম্পোনেন্টের অবস্থা বদলায়।

একটি হ্যান্ডলারের রানটাইম এরর কেউ না ধরলে তা অন্য যেকোনো না-ধরা RuneScript এররের মতোই
জানানো হয়, আর `runescript dev` সেটি তার ওভারলেতে দেখায়।

## `$if`

`$if (condition) … $elseif (condition) … $else … $endif` প্রথম সেই শাখা দেখায়, যার শর্ত
সত্য। `$elseif` ও `$else` ঐচ্ছিক। বেছে নেওয়া শাখা বদলালে পুরোনো শাখাটি তার সব বাইন্ডিংসহ
সরানো হয় এবং নতুনটি তৈরি হয়। যতক্ষণ সেটি একই থাকে, কিছুই নতুন করে তৈরি হয় না।

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

## চাইল্ড কম্পোনেন্ট

একটি কম্পোনেন্ট একটি ফাংশন, তাই আরেকটি কম্পোনেন্ট এটিকে এমন একটি ইন্টারপোলেশন দিয়ে
বসায়, যার মান একটি `Element`: `{{ Badge("new") }}`। প্রতিটি কল আলাদা একটি উদাহরণ, তার
নিজের রুনসহ। অন্য ফাইলের কম্পোনেন্টে পৌঁছানো যায় সেই ফাইলের কোয়ালিফায়ার দিয়ে, যেকোনো
ফাংশনের মতো ([অধ্যায় ১২](12-packages-and-files.md)):

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

`examples/web/multi-file` এই প্রোগ্রামটিই, তৃতীয় একটি ফাইলে একটি হেডারসহ।

## অ্যাপ মাউন্ট করা

কোনো রাউটার নেই। `dom.renderApp(component, selector)` কম্পোনেন্টটি কল করে এবং তার
এলিমেন্টটিকে প্রথম সেই এলিমেন্টে রাখে যা CSS সিলেক্টরটি মেলায়, সেখানে যা ছিল তা সরিয়ে।
যেকোনো সিলেক্টর কাজ করে, কিন্তু একটিমাত্র id ছাড়া অন্য কিছু, যেমন `".app"`,
`ASX-WARN-008` সতর্কবার্তা দেয়, কারণ এটি একাধিক এলিমেন্ট মেলাতে পারে, অথবা একটিও নয়।
কিছুই না মিললে `renderApp` একটি `Error` (`ASX-RUN-009`) থ্রো করে।

পেজটির সেই এলিমেন্ট দরকার, এবং কম্পাইল করা মডিউলটি লোড করা দরকার। এই `index.html`
`runescript dev` ও `runescript build` দুটির সঙ্গেই কাজ করে, যা এর পাশে `main.js` লেখে:

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

`index.html` না থাকলে `runescript dev` `<div id="app">` সহ একটি পেজ তৈরি করে।

## একটি যাচাই করা ভুল

টেমপ্লেট ব্যর্থ হয় কম্পাইলের সময়, ব্রাউজারে নয়:

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

## এখনো নেই

- `$foreach`, যা অ্যারের জন্য অপেক্ষা করছে।
- প্রপার্টিসহ কম্পোনেন্ট ট্যাগ (`<Badge label="new"/>`): কম্পোনেন্ট বসান
  `{{ Badge("new") }}` দিয়ে।
- ইভেন্ট অবজেক্ট, তাই একটি হ্যান্ডলার এখনো একটি ইনপুটের মান পড়তে পারে না, এবং দ্বিমুখী
  বাইন্ডিং।
- কাঁচা HTML ঢোকানো, MathML, কাস্টম এলিমেন্ট, এবং সার্ভার-সাইড রেন্ডারিং।
- হট মডিউল রিপ্লেসমেন্ট: `runescript dev` পেজটি রিলোড করে।

পরবর্তী: [লেখা ও তালিকা](16-text-and-arrays.md)
