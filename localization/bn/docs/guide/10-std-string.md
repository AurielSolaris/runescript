<!-- source: docs/guide/10-std-string.md @ c31fd14 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/10-std-string.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ১০. স্ট্যান্ডার্ড লাইব্রেরি: `std:string`

`std:string` হলো Go-র [`strings`](https://pkg.go.dev/strings) প্যাকেজের
RuneScript সংস্করণ। `use` দিয়ে এটি নিয়ে আসুন, তারপর `string` কোয়ালিফায়ারের মাধ্যমে
এর ফাংশনগুলো কল করুন:

```actionscript
package main;

use "std:string";

public void main() {
    string name = string.trimSpace("   rune script   ");
    trace(string.toUpper(name));                       // RUNE SCRIPT
    trace(string.hasPrefix(name, "rune"));             // true
    trace(string.replaceAll(name, " ", "-"));          // rune-script
}
```

## রীতিনীতি

- **কল করা।** `use "std:string";` বসে `package` লাইনের পরে এবং যেকোনো ফাংশনের
  আগে। তারপর কল করুন `string.<name>(...)`। `use` ছাড়া কোয়ালিফায়ারটি স্কোপে থাকে না
  (`ASX-NAME-007`)।
- **নাম।** Go-র নামগুলো camelCase-এ রূপান্তরিত: `ToUpper` → `toUpper`,
  `TrimSpace` → `trimSpace`, `ToValidUTF8` → `toValidUTF8`। আপনি Go-র বানান লিখলে
  এরর বার্তা সঠিকটি সুপারিশ করে।
- **অবস্থান মানে বাইট অফসেট।** স্ট্রিংয়ে থাকে UTF-8 লেখা, আর অবস্থান, দৈর্ঘ্য ও
  গণনা সবই **বাইটে**, ঠিক Go-র মতো। `"héllo"`-তে `é` দুই বাইট নেয়, তাই
  `string.index("héllo", "l")` হলো `3`।
- **ফলের টাইপ নির্দিষ্ট।** যেসব ফাংশন অবস্থান বা গণনা ফেরত দেয়, সেগুলো `int` ফেরত
  দেয়। এমন একটি মান `uint`-এ অ্যাসাইন করলে সেই পরিচিত `ASX-TYPE-004` এরর হয়।
- **আচরণ Go-র সঙ্গে মেলে।** প্রান্তিক ক্ষেত্রগুলো, যেমন খালি বিভাজক বা ঋণাত্মক
  `n`, Go-র ডকুমেন্টেশনে যেমন বলা আছে তেমনই আচরণ করে।
- **স্ট্রিং জোড়া লাগানো** হয় `+` বা ইন্টারপোলেটেড স্ট্রিং `$"…"` দিয়ে, কোনো
  লাইব্রেরি ফাংশন দিয়ে নয় ([অধ্যায় ১৬](16-text-and-arrays.md))।

## ফাংশন তালিকা

### তুলনা ও খোঁজা

| ফাংশন | ফেরত দেয় | বিবরণ |
|---|---|---|
| `compare(string a, string b)` | `int` | বাইট-অনুযায়ী ক্রমে `-1`, `0` বা `1` |
| `equalFold(string s, string t)` | `bool` | Unicode বড়-ছোট হাতের অক্ষর উপেক্ষা করে সমান |
| `contains(string s, string substr)` | `bool` | `s`-এ `substr` আছে |
| `containsAny(string s, string chars)` | `bool` | `chars`-এর যেকোনো অক্ষর `s`-এ আছে |
| `hasPrefix(string s, string prefix)` | `bool` | `s` শুরু হয় `prefix` দিয়ে |
| `hasSuffix(string s, string suffix)` | `bool` | `s` শেষ হয় `suffix` দিয়ে |
| `count(string s, string substr)` | `int` | পরস্পর না-ছোঁয়া উপস্থিতির সংখ্যা। খালি `substr` হলে: ১ + অক্ষরের সংখ্যা |
| `index(string s, string substr)` | `int` | প্রথম উপস্থিতির বাইট অফসেট, অথবা `-1` |
| `indexAny(string s, string chars)` | `int` | `chars`-এর প্রথম অক্ষরটির বাইট অফসেট, অথবা `-1` |
| `lastIndex(string s, string substr)` | `int` | শেষ উপস্থিতির বাইট অফসেট, অথবা `-1` |
| `lastIndexAny(string s, string chars)` | `int` | `chars`-এর শেষ অক্ষরটির বাইট অফসেট, অথবা `-1` |

```actionscript
package main;

use "std:string";

public void main() {
    string s = "go gopher";
    trace(string.contains(s, "ph"), string.index(s, "go"), string.lastIndex(s, "go"));   // true 0 3
    trace(string.count("cheese", "e"));                                                  // 3
    trace(string.equalFold("RuneScript", "runescript"));                                 // true
    if (string.index(s, "x") == -1) {
        trace("no x");
    }
}
```

### রূপান্তর

| ফাংশন | ফেরত দেয় | বিবরণ |
|---|---|---|
| `toUpper(string s)` | `string` | বড় হাতের অক্ষর |
| `toLower(string s)` | `string` | ছোট হাতের অক্ষর |
| `toTitle(string s)` | `string` | প্রতিটি অক্ষর টাইটেল কেসে (Unicode টাইটেল কেস; বেশিরভাগ অক্ষরে বড় হাতের অক্ষরের মতোই) |
| `repeat(string s, int count)` | `string` | `s`-কে `count` বার পুনরাবৃত্তি। ঋণাত্মক `count` একটি রানটাইম এরর (`ASX-RUN-004`) |
| `replace(string s, string old, string new, int n)` | `string` | প্রথম `n`টি উপস্থিতি বদলায়; `n < 0` হলে সবগুলো |
| `replaceAll(string s, string old, string new)` | `string` | প্রতিটি উপস্থিতি বদলায় |
| `toValidUTF8(string s, string replacement)` | `string` | অবৈধ UTF-8 অংশগুলো বদলে দেয় |
| `clone(string s)` | `string` | একটি নতুন কপি (সম্পূর্ণতার জন্য রাখা) |

```actionscript
package main;

use "std:string";

public void main() {
    trace(string.repeat("ab", 3));                          // ababab
    trace(string.replace("oink oink oink", "k", "ky", 2));  // oinky oinky oink
    trace(string.toLower("LOUD"), string.toUpper("quiet"));  // loud QUIET
}
```

### ছাঁটাই

| ফাংশন | ফেরত দেয় | বিবরণ |
|---|---|---|
| `trimSpace(string s)` | `string` | শুরু ও শেষের হোয়াইটস্পেস সরায় |
| `trim(string s, string cutset)` | `string` | শুরু ও শেষ থেকে `cutset`-এ থাকা অক্ষরগুলো সরায় |
| `trimLeft(string s, string cutset)` | `string` | শুধু শুরুর অক্ষরগুলো |
| `trimRight(string s, string cutset)` | `string` | শুধু শেষের অক্ষরগুলো |
| `trimPrefix(string s, string prefix)` | `string` | `prefix` থাকলে একবার সরায় |
| `trimSuffix(string s, string suffix)` | `string` | `suffix` থাকলে একবার সরায় |

```actionscript
package main;

use "std:string";

public void main() {
    trace(string.trim("!!hello!!", "!"));                 // hello
    trace(string.trimLeft("xxhixx", "x"));                // hixx
    trace(string.trimSuffix("game.rsc", ".rsc"));           // game
}
```

`trimPrefix`/`trimSuffix` একটি পুরো স্ট্রিং একবার সরায়। `trimLeft`/`trimRight`
তালিকাভুক্ত অক্ষরগুলোর যেকোনো টানা সারি সরায়।

### দৈর্ঘ্য ও স্লাইস

Go তার বিল্ট-ইন `len(s)` ও `s[i:j]` দিয়ে স্ট্রিং মাপে ও কাটে। RuneScript-এ এদের
জন্য ফাংশন আছে, যেগুলোও বাইটে গোনে:

| ফাংশন | ফেরত দেয় | বিবরণ |
|---|---|---|
| `len(string s)` | `int` | বাইটে দৈর্ঘ্য |
| `runeCount(string s)` | `int` | অক্ষরের (কোড পয়েন্টের) সংখ্যা |
| `slice(string s, int start, int end)` | `string` | `start` থেকে `end`-এর আগ পর্যন্ত বাইটগুলো |

`slice` কখনো একটি অক্ষরকে দুই টুকরো করে না। `0 … len(s)`-এর বাইরের সীমানা,
`start > end`, অথবা কোনো অক্ষরের মাঝখানের সীমানা একটি `RangeError`
(`ASX-RUN-004`) ছোড়ে।

```actionscript
package main;

use "std:string";

public void main() {
    string s = "héllo";
    trace(string.len(s), string.runeCount(s));   // 6 5
    trace(string.slice(s, 0, 3));                // hé
    int at = string.index(s, "llo");
    trace(string.slice(s, at, string.len(s)));   // llo
    try {
        trace(string.slice(s, 0, 2));            // 2 is inside 'é'
    } catch (RangeError e) {
        trace(e.message);   // string.slice: byte offset 2 is inside a character
    }
}
```

Go-তে যেমন, তেমনি `index` ও `slice` মিলিয়ে একটি স্ট্রিংকে টুকরো করুন।

### অক্ষর ও কলব্যাক

এই ফাংশনগুলো একটি [`char`](16-text-and-arrays.md#characters-char) নেয়, অথবা এমন
একটি ফাংশন, যা একবারে একটি অক্ষর দেখে। প্রতিটি কলব্যাক Go-র মতো একই ক্রমে ডাকা হয়:
বাঁ থেকে ডানে, কেবল `lastIndexFunc` ও ট্রিমের ডান দিকটি শেষ থেকে কাজ করে।

| ফাংশন | ফেরত দেয় | বিবরণ |
|---|---|---|
| `containsRune(string s, char c)` | `bool` | `s`-এ `c` আছে |
| `indexRune(string s, char c)` | `int` | প্রথম `c`-এর বাইট অফসেট, অথবা `-1` |
| `containsFunc(string s, bool(char) f)` | `bool` | কোনো একটি অক্ষর `f` পূরণ করে |
| `indexFunc(string s, bool(char) f)` | `int` | `f` পূরণ করে এমন প্রথম অক্ষরের বাইট অফসেট, অথবা `-1` |
| `lastIndexFunc(string s, bool(char) f)` | `int` | শেষটির বাইট অফসেট, অথবা `-1` |
| `trimFunc(string s, bool(char) f)` | `string` | শুরু ও শেষের যেসব অক্ষর `f` পূরণ করে, সেগুলো সরায় |
| `trimLeftFunc(string s, bool(char) f)` | `string` | কেবল শুরুরগুলো |
| `trimRightFunc(string s, bool(char) f)` | `string` | কেবল শেষেরগুলো |
| `map(char(char) f, string s)` | `string` | প্রতিটি অক্ষরের জায়গায় তার `f` (Go-র আর্গুমেন্ট-ক্রম লক্ষ করুন) |

```actionscript
package main;

use "std:string";

public bool isDigit(char c) {
    return c >= '0' && c <= '9';
}

public char hide(char c) {
    if (isDigit(c)) {
        return '#';
    }
    return c;
}

public void main() {
    string order = "order 66 shipped";
    trace(string.indexFunc(order, isDigit));            // 6
    trace(string.trimFunc("2024-report-v2", isDigit));  // -report-v
    trace(string.map(hide, order));                     // order ## shipped
    trace(string.containsRune(order, 'z'));             // false
}
```

Go-র `Map`-এর মতো নয়, `map` কোনো অক্ষর বাদ দিতে পারে না, কারণ `char` কখনো ঋণাত্মক
হয় না।

### ভাগ করা ও জোড়া লাগানো

এগুলো একটি [`Array<string>`](16-text-and-arrays.md#arrays-arrayt) ফেরত দেয় বা নেয়:

| ফাংশন | ফেরত দেয় | বিবরণ |
|---|---|---|
| `split(string s, string sep)` | `Array<string>` | বিভাজকগুলোর মাঝের টুকরো। খালি `sep` অক্ষরে অক্ষরে ভাগ করে |
| `splitN(string s, string sep, int n)` | `Array<string>` | সর্বোচ্চ `n`টি টুকরো, শেষটিতে বাকি সব; `n < 0` মানে সবগুলো, `n == 0` মানে একটিও না |
| `splitAfter(string s, string sep)` | `Array<string>` | `split`-এর মতো, তবে প্রতিটি টুকরো তার বিভাজক রাখে |
| `splitAfterN(string s, string sep, int n)` | `Array<string>` | `splitN`-এর মতো, বিভাজক রেখে |
| `fields(string s)` | `Array<string>` | শব্দগুলো: হোয়াইটস্পেসের মাঝের অক্ষরের সারি |
| `fieldsFunc(string s, bool(char) f)` | `Array<string>` | যেসব অক্ষর `f` পূরণ করে না, তাদের সারি |
| `join(Array<string> elems, string sep)` | `string` | এলিমেন্টগুলো, মাঝে `sep` দিয়ে |

```actionscript
package main;

use "std:string";

public void main() {
    Array<string> cells = string.split("name,age,,city", ",");
    trace(cells.length, cells[1], cells[2] == "");          // 4 age true

    Array<string> words = string.fields("  to be   or not ");
    trace(string.join(words, " "));                          // to be or not

    Array<string> kv = string.splitN("key=value=more", "=", 2);
    trace(kv[0], kv[1]);                                     // key value=more
}
```

## এখনো নেই

Go-র `strings`-এর এই ফাংশনগুলোর জন্য এমন ভাষা-বৈশিষ্ট্য লাগে যা RuneScript-এ এখনো
নেই:

| Go ফাংশন | যা লাগবে |
|---|---|
| `Cut`, `CutPrefix`, `CutSuffix` | একাধিক রিটার্ন মান |
| `IndexByte`, `LastIndexByte` | একটি বাইট টাইপ |
| `Lines`, `SplitSeq`, `SplitAfterSeq`, `FieldsSeq`, `FieldsFuncSeq` | ইটারেটর |
| `ToUpperSpecial`, `ToLowerSpecial`, `ToTitleSpecial` | একটি `unicode.SpecialCase` মান |
| `Builder`, `Reader`, `Replacer`, `NewReader`, `NewReplacer` | ক্লাস |
| `Title` | পরিকল্পনায় নেই: Go-তে এটি অপ্রচলিত, কারণ এটি Unicode শব্দের সীমানা ভুলভাবে সামলায় |

পরবর্তী: [JavaScript-এ কম্পাইল](11-javascript.md)
