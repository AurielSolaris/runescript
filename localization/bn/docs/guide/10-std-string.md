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

## এখনো নেই

Go-র `strings`-এর এই ফাংশনগুলোর জন্য এমন ভাষা-বৈশিষ্ট্য লাগে যা RuneScript-এ এখনো
নেই:

| Go ফাংশন | যা লাগবে |
|---|---|
| `Split`, `SplitN`, `SplitAfter`, `SplitAfterN`, `Fields`, `Join` | অ্যারে (`[]string`) |
| `Cut`, `CutPrefix`, `CutSuffix` | একাধিক রিটার্ন মান |
| `ContainsRune`, `IndexRune`, `IndexByte`, `LastIndexByte` | `char` / বাইট টাইপ |
| `ContainsFunc`, `IndexFunc`, `LastIndexFunc`, `FieldsFunc`, `TrimFunc`, `TrimLeftFunc`, `TrimRightFunc`, `Map` | ফাংশন মান |
| `Lines`, `SplitSeq`, `SplitAfterSeq`, `FieldsSeq`, `FieldsFuncSeq` | ইটারেটর |
| `ToUpperSpecial`, `ToLowerSpecial`, `ToTitleSpecial` | একটি `unicode.SpecialCase` মান |
| `Builder`, `Reader`, `Replacer`, `NewReader`, `NewReplacer` | ক্লাস |
| `Title` | পরিকল্পনায় নেই: Go-তে এটি অপ্রচলিত, কারণ এটি Unicode শব্দের সীমানা ভুলভাবে সামলায় |

দৈর্ঘ্য মাপার ফাংশনও এখনো নেই, কারণ Go তার বিল্ট-ইন `len` ব্যবহার করে। অ্যারের
সঙ্গে একটি দৈর্ঘ্য-অপারেশনের পরিকল্পনা আছে।

পরবর্তী: [JavaScript-এ কম্পাইল](11-javascript.md)
