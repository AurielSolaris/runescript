<!-- source: docs/guide/12-packages-and-files.md @ 920cb24 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/12-packages-and-files.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ১২. প্যাকেজ ও ফাইল

`use` দিয়ে একটি প্রোগ্রামকে কয়েকটি ফাইলে ভাগ করা যায়। `use` তিন রকমের:

```text
use "std:string";          // a standard library package
use "geom";                // a RuneScript package: every .rsm file in geom/
use "./util/text.txt";     // one file, whatever its extension
```

যে স্ট্রিংয়ে `/` আছে, সেটি একটি ফাইলের পাথ। বাকি সবকিছু প্যাকেজের নাম, যেমন
`geom` বা `org.games.engine`। পাথ ও প্যাকেজের ডিরেক্টরি ধরা হয় **যে ডিরেক্টরি থেকে
আপনি `runescript` চালান** তার সাপেক্ষে, `use` লেখা ফাইলটির সাপেক্ষে নয়।

## একটি উদাহরণ

এই প্রোগ্রামটি টেস্ট-সংগ্রহে আছে `tests/testdata/multi/ok_packages`-এ। ওই
ডিরেক্টরি থেকে `runescript run main.rsc` দিয়ে চালান।

```text
ok_packages/
├── main.rsc            package app;   the entry file
├── helpers.rsc         package app;   another file of the same package
├── geom/
│   ├── area.rsm        package geom;
│   └── perimeter.rsm   package geom;
└── util/
    └── text.txt        package text;  any extension works
```

`main.rsc`:

```text
package app;

use "geom";
use "./util/text.txt" as t;
use "./helpers.rsc";

protected int main() {
    trace(geom.area(3, 4), geom.perimeter(3, 4));
    trace(t.shout("hi"));
    trace(twice(21));
    return 0;
}
```

`geom/perimeter.rsm`:

```text
package geom;

use "./util/text.txt";

public int perimeter(int w, int h) {
    trace(text.shout("perimeter"));
    return scale(2 * (w + h));
}

// Protected: shared by both files of package geom.
protected int scale(int n) {
    return n * one();
}

// Private to this file.
private int one() {
    return 1;
}
```

এটি দেখায়:

```text
PERIMETER
12 14
HI
42
```

## প্যাকেজ

একটি প্যাকেজ হলো সেই সব ফাইল, যারা একই `package` নাম ঘোষণা করে।

- `use "geom";` সরাসরি `geom/` ডিরেক্টরিতে থাকা `.rsm` ফাইলগুলো লোড করে।
  `use "org.games.engine";` সেগুলো লোড করে `org/games/engine/` থেকে। ডিরেক্টরির
  অন্য ফাইলগুলো উপেক্ষা করা হয়, সাব-ডিরেক্টরিগুলোও। প্রতিটি `.rsm` ফাইলকে ওই
  প্যাকেজটিই ঘোষণা করতে হবে (`ASX-USE-007`)।
- একই প্যাকেজের ফাইলগুলো একে অপরের `public` ও `protected` ফাংশন কোয়ালিফায়ার বা
  প্রোটোটাইপ ছাড়াই দেখতে পায়। উদাহরণে `area.rsm` কল করে `scale`, যা সংজ্ঞায়িত
  `perimeter.rsm`-এ।
- যে ফাইল আপনি পাথ দিয়ে `use` করেন, সেটি তার ঘোষিত প্যাকেজেই যোগ দেয়। উদাহরণে
  `helpers.rsc` ঘোষণা করে `package app;`, তাই তার `twice` সরাসরি `main.rsc` থেকে কল
  করা হয়।

যতগুলো ফাইলই একটি ফাইল ব্যবহার করুক, প্রতিটি ফাইল একবারই লোড হয়। `main.rsc` ও
`perimeter.rsm` দুটিই `util/text.txt` ব্যবহার করে, আর সেটি একবার লোড হয়।

## কোয়ালিফায়ার

অন্য প্যাকেজের ফাংশন কল করা হয় একটি **কোয়ালিফায়ার**-এর মাধ্যমে। ডিফল্টভাবে এটি
প্যাকেজের নামের শেষ অংশ: `geom`-এর জন্য `geom`, আর `org.games.engine`-এর জন্য
`engine`। ফাইলের ক্ষেত্রে এটি ফাইলটির ঘোষিত প্যাকেজ: `util/text.txt`-এর জন্য
`text`। অন্য নাম বেছে নিতে `as` ব্যবহার করুন:

```text
use "./util/text.txt" as t;     // t.shout("hi")
use "org.games.engine" as eng;  // eng.start()
```

কোয়ালিফায়ার সেই ফাইলের, যেখানে এটি ঘোষিত। `geom.area` কল করে এমন প্রতিটি ফাইলের
নিজের `use "geom";` লাগে।

একই ফাইলের দুটি `use`-এর কোয়ালিফায়ার একই হয়ে গেলে, সেটি `ASX-USE-008`। `as` দিয়ে
একটির নাম বদলে দিন।

একটি কোয়ালিফাইড ফাংশন একটি মানও: `int(int, int) f = geom.area;` `geom.area`-কে কল
না করেই পাঠায়, যেমন `dom.renderApp(app.App, "#app")` একটি কম্পোনেন্টের সঙ্গে করে
([অধ্যায় ১৫](15-web-pages.md))। স্ট্যান্ডার্ড লাইব্রেরির ফাংশন আপাতত কেবল কল করা
যায়: সেগুলো একটি ল্যাম্বডায় মুড়ে দিন, যেমন `(string s) => string.toUpper(s)`।

## দৃশ্যমানতা

এখন প্রতিটি ফাংশনের অ্যাক্সেস মডিফায়ার কার্যকর করা হয়:

| মডিফায়ার | কে কল করতে পারে |
|---|---|
| `public` | প্রতিটি প্যাকেজের প্রতিটি ফাইল (অন্য প্যাকেজ কোয়ালিফায়ার ব্যবহার করে) |
| `protected` | নিজের প্যাকেজের প্রতিটি ফাইল |
| `private` | শুধু নিজের ফাইল |

যে ফাংশনের মডিফায়ার অনুমতি দেয় না তাকে কল করলে `ASX-NAME-011` হয়, আর বার্তাটি
বলে দেয় ফাংশনটি কোথায়। দুটি ফাইলের প্রতিটিতে একই নামের নিজস্ব `private` ফাংশন
থাকতে পারে। একটি প্যাকেজে অন্য যেকোনো পুনরাবৃত্ত টপ-লেভেল নাম হলো `ASX-NAME-002`।

## প্রবেশবিন্দু

একটি প্রোগ্রামে ঠিক একটি `main` থাকে, এবং সেটি এর যেকোনো ফাইলে থাকতে পারে।
`runescript run`-কে যে ফাইল দেন, তাতে এটি থাকতেই হবে এমন নয়। একাধিক `main` হলে
`ASX-ENTRY-003`।

## চক্র

প্যাকেজগুলো চক্রাকারে একে অপরকে ব্যবহার করতে পারে না। `a` যদি `b` ব্যবহার করে আর
`b` ব্যবহার করে `a`, সরাসরি বা অন্য প্যাকেজের মাধ্যমে, তবে সেটি `ASX-USE-005`, আর
বার্তাটি চক্রটি দেখায়: `import cycle: a → b → a`। ঠিক করতে, তারা যে ফাংশনগুলো ভাগ
করে নেয় সেগুলো একটি তৃতীয় প্যাকেজে সরিয়ে দিন।

## ফাইলের নাম ও এক্সটেনশন

রীতি অনুযায়ী প্রোগ্রাম হলো `.rsc` ফাইল, আর প্যাকেজের ডিরেক্টরিতে থাকে `.rsm`
ফাইল। Perl-এর মতো, আপনি নিজে কোনো ফাইলের নাম দিলে এক্সটেনশন যাচাই করা হয় না:
`runescript run tool.txt` ও `use "./lib/x.inc";` দুটিই কাজ করে, যতক্ষণ ফাইলে
RuneScript থাকে। শুধু প্যাকেজের ডিরেক্টরিই এক্সটেনশন দেখে, যাতে পাশে থাকা README
ও ডেটা ফাইল উপেক্ষিত হয়।

ডায়াগনস্টিক জানায় কোন ফাইলে সমস্যা, ওয়ার্কিং ডিরেক্টরির সাপেক্ষে:

```text
calc/square.rsm:4:14: error ASX-RUN-002: integer overflow: 65536 * 65536 = 4294967296 does not fit in 'int'
```

## এরর

| কোড | কারণ |
|---|---|
| `ASX-USE-004` | ফাইল নেই, অথবা প্যাকেজের ডিরেক্টরিতে কোনো `.rsm` ফাইল নেই |
| `ASX-USE-005` | প্যাকেজগুলো চক্রাকারে একে অপরকে ব্যবহার করে |
| `ASX-USE-006` | অবৈধ প্যাকেজ নাম, অথবা `\` দিয়ে লেখা পাথ |
| `ASX-USE-007` | এমন `.rsm` ফাইল, যা তার ডিরেক্টরির থেকে ভিন্ন প্যাকেজ ঘোষণা করে |
| `ASX-USE-008` | একই ফাইলে একই কোয়ালিফায়ারের দুটি `use` |
| `ASX-NAME-011` | একটি `private` বা `protected` ফাংশন বাইরে থেকে ব্যবহার |
| `ASX-ENTRY-003` | একাধিক `main` |

পরবর্তী: [ফ্লোট ও রূপান্তর](13-floats-and-conversions.md)
