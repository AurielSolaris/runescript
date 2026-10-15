<!-- source: README.md @ 655da49 -->

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

## অবস্থা: v0.3 তৈরি হচ্ছে (প্রাথমিক পর্যায়)

**এখন যা কাজ করে**, এবং টেস্ট দিয়ে যাচাই করা:

- প্যাকেজ; `public|private|protected <type> <name>(…)` আকারে ঘোষিত ফাংশন; রিকার্শন
- নেস্টেড ফাংশন, প্রোটোটাইপ, ফাংশন টাইপ (`int(int, int)`), কলব্যাক এবং ল্যাম্বডা (`(int a) => a * 2`)
- `int`/`sint`, `uint`, `bool`, `string`, `void`
- ভেরিয়েবল, `=`, `+= -= *= /= %=`, `++`/`--`
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
- ActionScript-ধাঁচের এরর: `throw new RangeError("…")`,
  `try`/`catch (Error e)`/`catch (e)`/`finally`; রানটাইম এররও ধরা (catch) যায়
  ([গাইড](docs/guide/14-errors.md))
- অবস্থান-সহ ডায়াগনস্টিক দিয়ে কম্পাইল-টাইম টাইপ-চেকিং, যার মধ্যে সাইনড → আনসাইনড
  অ্যাসাইনমেন্ট প্রত্যাখ্যানও আছে
- পূর্ণসংখ্যার ওভারফ্লো ও শূন্য দিয়ে ভাগ রানটাইম এরর হিসেবে জানানো হয়, কখনো
  নিঃশব্দে ঘুরে (wrap) যায় না
- JavaScript-এ কম্পাইল (`runescript build`): একটি ES2022 মডিউল, যা ব্রাউজার ও
  Node.js-এ চলে এবং ইন্টারপ্রেটারের মতো হুবহু আচরণ করে
  ([গাইড](docs/guide/11-javascript.md))

**এখনো তৈরি হয়নি:** ক্লাস, স্ট্রাক্ট, অ্যারে, `char`, স্ট্রিং জোড়া লাগানো
(concatenation), `go:` প্যাকেজ, `std:math`, `std:symrs`, এবং পরিকল্পিত নেটিভ
গ্রাফিক্স প্যাকেজ। দেখুন [রোডম্যাপ](../../docs/ROADMAP.md) (ইংরেজিতে)।

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
- ডকুমেন্টেশনের প্রতিটি প্রোগ্রাম কম্পাইল হয় কিনা, তার যাচাই।

## উদাহরণ

[`examples/`](../../examples/README.md)-এ প্রতিটি বৈশিষ্ট্যের প্রোগ্রাম আছে: মৌলিক
বিষয়, `std:string`, কলব্যাক ও ল্যাম্বডা, একটি মৌলিক সংখ্যা গণক, এবং দুটি প্রোগ্রাম
যা কম্পাইলার ইচ্ছাকৃতভাবে প্রত্যাখ্যান করে। প্রতিটি উদাহরণের আউটপুট টেস্ট দিয়ে
যাচাই করা হয়।

## ডকুমেন্টেশন

- [গাইড](docs/guide/README.md): উদাহরণসহ প্রতিটি বৈশিষ্ট্যের পরিচয় (সবগুলো টেস্টে কম্পাইল হয়)
- [ভাষার স্পেসিফিকেশন](../../docs/LANGUAGE_SPEC.md) (ইংরেজিতে): ব্যাকরণ ও নিয়ম, প্রতিটি নিয়ম কতটা বাস্তবায়িত তা সহ
- [স্ট্যান্ডার্ড লাইব্রেরি: std:string](docs/guide/10-std-string.md)
- [JavaScript-এ কম্পাইল](docs/guide/11-javascript.md)
- [ডায়াগনস্টিক তালিকা](docs/guide/08-diagnostics.md)
- [ActionScript 3 থেকে আসছেন?](docs/guide/09-coming-from-actionscript.md)
- [রোডম্যাপ](../../docs/ROADMAP.md) (ইংরেজিতে)

## রিপোজিটরির গঠন

```text
cmd/runescript/        CLI (run, check, build, version)
internal/lexer/        tokens and source positions
internal/parser/       recursive-descent parser → AST
internal/ast/          syntax tree
internal/types/        semantic types
internal/checker/      name resolution and type checking
internal/runtime/      tree-walking interpreter
internal/ir/           typed intermediate representation and its verifier
internal/jsgen/        JavaScript code generator and its inlined runtime
internal/diagnostics/  error codes and formatting
internal/driver/       wires the stages together
stdlib/                standard library packages (std:string), implemented in Go
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
