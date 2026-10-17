<!-- source: README.md @ 362a461 -->

> এটি [ইংরেজি README](../../README.md)-এর অনুবাদ। দুটির মধ্যে পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক বলে ধরা হবে। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক নিজে যাচাই করেছেন; সংশোধন সাদরে গ্রহণ করা হয়।

<p align="center">
  <img src="../../assets/runescript.png" alt="RuneScript লোগো: বিদ্যুতের রেখা-সহ রুন-শৈলীর একটি তুষারকণা" width="160">
</p>

<h1 align="center">RuneScript</h1>

<p align="center">ActionScript থেকে অনুপ্রাণিত, Go-তে তৈরি একটি স্ট্যাটিক্যালি টাইপড, সাধারণ-উদ্দেশ্যের প্রোগ্রামিং ভাষা।</p>

<p align="center"><a href="../../README.md" lang="en">English</a> · <strong>বাংলা</strong> · <a href="../heb/README.md" lang="he">עברית</a> (হিব্রু)</p>

---

RuneScript ActionScript-এর পরিচয় ধরে রাখে (`package`, `trace`, `int`/`uint`,
`public`/`protected`/`private`), তবে সিনট্যাক্স বদলে C-পরিবারের ধাঁচে নিয়ে
এসেছে: টাইপ লেখা হয় নামের আগে, প্রতিটি ফাংশন `function` কিওয়ার্ডের বদলে
একটি অ্যাক্সেস মডিফায়ার দিয়ে শুরু হয় (`public int add(int a, int b)`), আর
সেমিকোলন বাধ্যতামূলক। এর টাইপ-চেকিং কঠোর: কোনো কিছুই নিজে থেকে `any` হয় না,
আর সাইনড ও আনসাইনড পূর্ণসংখ্যা কখনো নিঃশব্দে মিশে যায় না।

এটি একটি স্বতন্ত্র প্রোগ্রাম হিসেবে চলে: ব্রাউজার, HTML বা Flash Player লাগে না।
ওয়েবের জন্য এটি একটি JavaScript মডিউলেও কম্পাইল হতে পারে, একই কঠোর টাইপিং
এবং একই রানটাইম এরর সহ। এটি ActionScript 3-এর সঙ্গে সোর্স-সামঞ্জস্যপূর্ণ
**নয়**।

```actionscript
package main;

public void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

protected int main() {
    sayHi("Hello world", 3);
    return 0;
}
```

## অবস্থা: v0.6 তৈরি হচ্ছে (প্রাথমিক পর্যায়)

**এখন যা কাজ করে**, এবং টেস্ট দিয়ে যাচাই করা:

- প্যাকেজ; `public|private|protected <type> <name>(…)` আকারে ঘোষিত ফাংশন; রিকার্শন;
  প্রবেশবিন্দু `protected int main()`, যার ফল এক্সিট স্ট্যাটাস
- নেস্টেড ফাংশন, প্রোটোটাইপ, ফাংশন টাইপ (`int(int, int)`), কলব্যাক এবং ল্যাম্বডা (`(int a) => a * 2`)
- `int`/`sint`, `uint`, `bool`, `string`, `void`
- ভেরিয়েবল, `=`, `+= -= *= /= %=`, `++`/`--`; `int n;` শূন্য মান দিয়ে শুরু হয়
  (কোনো `null` নেই); নেস্টেড ব্লকে শ্যাডোইং, একটি সতর্কবার্তাসহ
- গাণিতিক, তুলনা ও লজিক্যাল অপারেটর
- `if`/`else if`/`else`, `while`, `for`, `break`, `continue`, `return`
- `trace(...)`
- স্ট্যান্ডার্ড লাইব্রেরি: `use "std:string";` দিয়ে Go-র `strings` ফাংশনগুলো পাওয়া যায়, যেমন `string.toUpper(s)`
- একাধিক ফাইলের প্রোগ্রাম: `use "geom";` লোড করে `geom/*.rsm`-এর প্যাকেজ, আর
  `use "./lib/x.rsc";` লোড করে একটি ফাইল; `public`/`protected`/`private` প্রয়োগ
  করা হয় ([গাইড](docs/guide/12-packages-and-files.md))
- `float` (ডিফল্ট float32, মান না হারিয়ে প্রশস্ত হয়), `float32`, `float64`;
  ActionScript-এর মতোই `NaN` ও `Infinity` সহ IEEE 754
- যাচাইকৃত C-ধাঁচের রূপান্তর: `(int) 64.3`, `(float) "64.7"`, `(uint) n`
  ([গাইড](docs/guide/13-floats-and-conversions.md))
- নির্দিষ্ট প্রস্থের পূর্ণসংখ্যা `int8` … `uint64` (JavaScript-এ ৬৪-বিটের মান BigInt),
  `_` বিভাজকসহ হেক্স/বাইনারি/অক্টাল লিটারাল, বিট অপারেটর, এবং `std:math`, যার ফল দুটি
  টার্গেটে বিট-বিট এক ([গাইড](docs/guide/18-numbers.md))
- স্ট্রাক্ট: `Point{x: 1, y: 2}` দিয়ে তৈরি মান-টাইপ, অ্যাসাইনমেন্টে কপি হয়, ফিল্ড ধরে তুলনা
  হয় ([গাইড](docs/guide/19-structs.md))
- স্ক্রিপ্টের জন্য লেখা ও তালিকা: স্ট্রিংয়ের জন্য `+` ও `$"{count} files"`, `char`,
  এবং সীমা-যাচাই করা ইনডেক্সসহ `Array<T>` ([গাইড](docs/guide/16-text-and-arrays.md));
  `std:string` Go-র `strings` প্যাকেজ জুড়ে আছে, `split`, `fields`, `join` ও
  কলব্যাকসহ ([গাইড](docs/guide/10-std-string.md))
- ActionScript-ধাঁচের এরর: `throw new RangeError("…")`,
  `try`/`catch (Error e)`/`catch (e)`/`finally`; রানটাইম এররও ধরা (catch) যায়
  ([গাইড](docs/guide/14-errors.md))
- অবস্থান-সহ ডায়াগনস্টিক দিয়ে কম্পাইল-টাইম টাইপ-চেকিং, যার মধ্যে সাইনড → আনসাইনড
  অ্যাসাইনমেন্ট প্রত্যাখ্যানও আছে
- পূর্ণসংখ্যার ওভারফ্লো ও শূন্য দিয়ে ভাগ রানটাইম এরর হিসেবে জানানো হয়, কখনো
  নিঃশব্দে ঘুরে (wrap) যায় না; `int`/`uint`-এর মধ্যবর্তী ফল হুবহু প্রশস্ত হয়, তাই
  `big * big / big` কাজ করে ([গাইড](docs/guide/04-operators.md))
- দুটি টার্গেটেই ফাইল, ইনপুট ও আউটপুট: `std:io` (নতুন লাইন ছাড়া প্রিন্ট, লাল এরর ও
  হলুদ সতর্কবার্তার লাইন, স্ট্যান্ডার্ড ইনপুট), `std:process` (আর্গুমেন্ট, এনভায়রনমেন্ট),
  `std:fs` (ফাইল ও ডিরেক্টরি), এবং `std:path` (Node.js-এর path-এর নিয়ম, RuneScript-এর
  রানটাইমে বাস্তবায়িত) ([গাইড](docs/guide/20-files-and-input.md))
- JavaScript-এ কম্পাইল (`runescript build`): একটি ES2015 মডিউল (সাথে BigInt ও `??`), যা ব্রাউজার ও
  Node.js-এ চলে এবং ইন্টারপ্রেটার যে প্রোগ্রাম চালাতে পারে তার প্রতিটিতে
  ইন্টারপ্রেটারের মতো হুবহু আচরণ করে ([গাইড](docs/guide/11-javascript.md))
- `rune:dom` দিয়ে ওয়েব পেজ: HTML ও SVG টেমপ্লেটসহ কম্পোনেন্ট, যাদের প্রতিটি এলিমেন্ট
  ও অ্যাট্রিবিউট কম্পাইলের সময় যাচাই হয়, এস্কেপ করা `{{ }}` লেখা, `$if`/`$elseif`/`$else`,
  ইভেন্ট, এবং রুন (সিগন্যাল), যা কেবল সেগুলোই আপডেট করে যেগুলো সেটি পড়েছে, ভার্চুয়াল
  DOM ছাড়া ([গাইড](docs/guide/15-web-pages.md))
- `extern` ঘোষণার মাধ্যমে JavaScript কল করা, হোস্ট-নিরপেক্ষ `rune:fs` সহ।
  JavaScript থেকে ফেরত আসা প্রতিটি মান তার ঘোষিত টাইপের সঙ্গে যাচাই করা হয়
- অব্যবহৃত ভেরিয়েবল, প্যারামিটার ও use, অগম্য কোড, ফেলে দেওয়া ফলাফল এবং শ্যাডোইংয়ের
  জন্য সতর্কবার্তা, `-Werror` ও `--allow <code>` সহ; টার্মিনালে এরর লাল ও সতর্কবার্তা
  হলুদ (`--color`)
- কাজের সময় লাইভ রিলোড: `runescript dev app.rsc` একটি ওয়েব প্রোগ্রাম পরিবেশন করে
  এবং প্রতিটি সেভে ব্রাউজার রিলোড করে; `runescript run --watch` একটি কমান্ড-লাইন
  প্রোগ্রাম আবার চালায়

**JavaScript সামঞ্জস্য নিয়ে:** RuneScript এখনো JavaScript-এর সঙ্গে পূর্ণ
সামঞ্জস্যের প্রতিশ্রুতি দেয় না। এটি যার প্রতিশ্রুতি দেয় তা হলো কঠোর কোড:
TypeScript তার ঘোষণাগুলোকে বিশ্বাস করে এবং মুছে ফেলে, কিন্তু RuneScript
JavaScript থেকে আসা প্রতিটি মান যাচাই করে, এবং এতে কোনো `any`, কোনো অন্তর্নিহিত
রূপান্তর বা কোনো যাচাইহীন কাস্ট নেই।

**এখনো তৈরি হয়নি:** ক্লাস, for-each লুপ, বাইনারি ফাইল, `go:` প্যাকেজ, `std:symrs`,
JavaScript-এর সীমানা পেরিয়ে অবজেক্ট, অ্যারে, স্ট্রাক্ট ও কলব্যাক, টেমপ্লেটে `$foreach`
ও কম্পোনেন্ট ট্যাগ, এবং পরিকল্পিত নেটিভ গ্রাফিক্স প্যাকেজ। দেখুন [রোডম্যাপ](../../docs/ROADMAP.md) (ইংরেজিতে)।

## ইনস্টল

Go 1.25 বা নতুনতর লাগবে ([go.dev/dl](https://go.dev/dl/)); ইনস্টলার সেটি যাচাই
করে, `runescript` CLI তৈরি করে এবং সেটি আপনার PATH-এ যোগ করে।

```sh
# Linux, macOS
curl -fsSL https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.sh | sh
```

```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.ps1 | iex
```

অপশন, আনইনস্টল করা, এবং ক্লোন থেকে ইনস্টল করার নিয়ম আছে
[শুরু করার গাইডে](docs/guide/01-getting-started.md)।

## সোর্স থেকে তৈরি ও চালানো

```sh
go build -o runescript ./cmd/runescript
./runescript run examples/hello.rsc
./runescript check examples/signedness_error.rsc   # shows a compile-time error
./runescript build examples/fizzbuzz.rsc            # writes examples/fizzbuzz.js
node examples/fizzbuzz.js                         # runs it (Node.js 24+)
./runescript dev examples/fizzbuzz.rsc              # serves it, reloading on every save
./runescript run --watch examples/hello.rsc         # runs it again on every save
./runescript run examples/brainfuck.rsc examples/bf/hello.bf   # arguments follow the file
go test ./...
```

> RuneScript প্রোগ্রাম `.rsc` এক্সটেনশন ব্যবহার করে, আর প্যাকেজ ডিরেক্টরিতে থাকে
> `.rsm` ফাইল। Perl-এর মতো, এক্সটেনশন কেবল একটি প্রথা: CLI এবং ফাইল `use` যেকোনো
> নামের ফাইল গ্রহণ করে, যদি ফাইলটিতে RuneScript থাকে।

## টেস্ট

```sh
go test ./...                                  # unit, golden, docs, and CLI tests
go test ./tests -run '^$' -fuzz=FuzzCheck      # fuzz the parser and checker
go test ./tests -run '^$' -fuzz=FuzzRun        # fuzz the interpreter
```

টেস্ট-সংগ্রহে আছে:
- কম্পাইলারের প্রতিটি ধাপের ইউনিট টেস্ট;
- গোল্ডেন টেস্ট, যা প্রতিটি উদাহরণ চালিয়ে তার আউটপুট মিলিয়ে দেখে;
- এরর প্রোগ্রাম, যাদের প্রত্যাশিত ডায়াগনস্টিক লাইনের ভেতরেই
  `// expect: CODE line:col` হিসেবে লেখা থাকে;
- রানটাইম-এরর প্রোগ্রাম;
- ডিফারেনশিয়াল টেস্ট, যা প্রতিটি প্রোগ্রাম ইন্টারপ্রেটার ও Node.js দুটিতেই চালায়
  এবং হুবহু একই ফল চায় (`node` না থাকলে বাদ যায়);
- `extern` দিয়ে JavaScript কল করা প্রোগ্রাম Node.js-এ চালানো, ইচ্ছে করে ভুল মান
  ফেরত দেওয়া JavaScript সহ;
- তৈরি হওয়া JavaScript ES2015, `BigInt`, `??` ও মডিউলের চেয়ে নতুন কিছু ব্যবহার
  করে না, তার যাচাই;
- ডেভ সার্ভার, ফাইল ওয়াচার ও `run --watch`-এর টেস্ট;
- ডকুমেন্টেশনের প্রতিটি প্রোগ্রাম কম্পাইল হয় কিনা, এবং অনুবাদগুলো ইংরেজি পেজের প্রতিটি
  কোড ব্লক রাখে কিনা, তার যাচাই।

## উদাহরণ

[`examples/`](../../examples/README.md)-এ প্রতিটি বৈশিষ্ট্যের প্রোগ্রাম আছে: মৌলিক
বিষয়, `std:string`, কলব্যাক ও ল্যাম্বডা, একটি মৌলিক সংখ্যা গণক, ৩২,৭৬৮ ঘরের একটি
[Brainfuck ইন্টারপ্রেটার](../../examples/brainfuck.rsc), এবং দুটি প্রোগ্রাম
যা কম্পাইলার ইচ্ছাকৃতভাবে প্রত্যাখ্যান করে। প্রতিটি উদাহরণের আউটপুট টেস্ট দিয়ে
যাচাই করা হয়।

## ডকুমেন্টেশন

- [গাইড](docs/guide/README.md): উদাহরণসহ প্রতিটি বৈশিষ্ট্যের পরিচয় (সবগুলো টেস্টে কম্পাইল হয়)
- [ভাষার স্পেসিফিকেশন](../../docs/LANGUAGE_SPEC.md) (ইংরেজিতে): ব্যাকরণ ও নিয়ম, প্রতিটি নিয়ম কতটা বাস্তবায়িত তা সহ
- [স্ট্যান্ডার্ড লাইব্রেরি: std:string](docs/guide/10-std-string.md)
- [লেখা ও তালিকা: +, $"…", char, Array<T>](docs/guide/16-text-and-arrays.md)
- [null নেই, any নেই, জেনেরিক নেই](docs/guide/17-no-null-any-generics.md)
- [সংখ্যার গভীরে: প্রস্থ, লিটারাল, বিট, std:math](docs/guide/18-numbers.md)
- [স্ট্রাক্ট](docs/guide/19-structs.md)
- [ফাইল, ইনপুট ও আউটপুট: std:io, std:process, std:fs, std:path](docs/guide/20-files-and-input.md)
- [JavaScript-এ কম্পাইল](docs/guide/11-javascript.md)
- [ডায়াগনস্টিক তালিকা](docs/guide/08-diagnostics.md)
- [ActionScript 3 থেকে আসছেন?](docs/guide/09-coming-from-actionscript.md)
- [রোডম্যাপ](../../docs/ROADMAP.md) (ইংরেজিতে)

## রিপোজিটরির গঠন

```text
cmd/runescript/        CLI (run, check, build, dev, version)
internal/lexer/        tokens and source positions
internal/parser/       recursive-descent parser → AST
internal/ast/          syntax tree
internal/types/        semantic types
internal/srcenc/       reads source files (UTF-8; UTF-16 and UTF-32 are converted)
internal/loader/       loads a program's files and resolves uses
internal/checker/      name resolution, type checking, and warnings
internal/runtime/      tree-walking interpreter
internal/conv/         number and string conversions shared by both targets
internal/ir/           typed intermediate representation and its verifier
internal/jsgen/        JavaScript code generator and its inlined runtime
internal/jshost/       host-neutral JavaScript modules (rune:fs)
internal/dom/          HTML4, HTML5, and SVG elements and attributes (for templates)
internal/diagnostics/  error codes and formatting
internal/term/         terminal detection for colored output
internal/driver/       wires the stages together
internal/watch/        file watcher for run --watch and dev
internal/devserver/    the runescript dev server and its browser client
stdlib/                standard library packages (std:string, std:math, std:io, std:process,
                       std:fs, std:path), implemented in Go; their JavaScript twins are in
                       internal/jsgen/runtime.js
examples/              example programs
tests/                 end-to-end, golden, and documentation tests
assets/                logo: SVG source and PNGs generated from it
scripts/               render-logo.sh: SVG → PNG via headless Edge/Chrome
```

## অনুবাদ

ইংরেজিই মূল: ইংরেজি README এবং [`docs/`](../../docs/) হলো চূড়ান্ত উৎস। অনুবাদগুলো
থাকে [`localization/`](../README.md)-এ, প্রতিটি ভাষার জন্য একটি ফোল্ডার:

| ভাষা | ফোল্ডার | যা অনূদিত |
|---|---|---|
| বাংলা | [`localization/bn/`](README.md) | README এবং গাইড |
| עברית (হিব্রু) | [`localization/heb/`](../heb/README.md) | README এবং গাইড |

বাংলা ও হিব্রু অনুবাদ AI-এর সাহায্যে রক্ষণাবেক্ষণ করা হয় এবং লেখক নিজে হাতে যাচাই
করেন। **সাহায্য সাদরে গ্রহণীয়:** মাতৃভাষী পাঠকদের সংশোধন এবং নতুন ভাষা, দুটোই
খুব কাম্য। অনুবাদ কীভাবে সাজানো ও ইংরেজি ডকুমেন্টের সঙ্গে হালনাগাদ রাখা হয়, তা
দেখুন [localization/README.md](../README.md)-এ (ইংরেজিতে)।

## লোগো

`assets/runescript.svg` (README-এর জন্য) এবং `assets/runescript-icon.svg` (ছোট
আইকনের জন্য) হাতে লেখা সোর্স ফাইল। PNG-গুলো এগুলো থেকেই তৈরি করা হয়:

```sh
scripts/render-logo.sh     # needs Microsoft Edge or Chrome; override with BROWSER=...
```

## লাইসেন্স

RuneScript [MIT লাইসেন্সের](../../LICENSE) অধীনে প্রকাশিত।
