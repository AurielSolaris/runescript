<!-- source: docs/guide/21-json.md @ b50b7fd -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/21-json.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ২১. JSON: `std:json`

`use "std:json";` [JSON](https://www.json.org/) লেখা পড়ে ও লেখে। একটি JSON ডকুমেন্ট
হয়ে যায় একটি `json.Value`: একটি null, একটি bool, একটি সংখ্যা, একটি স্ট্রিং, একটি
অ্যারে, অথবা একটি অবজেক্ট। RuneScript-এ সব-ধরা কোনো টাইপ নেই
([অধ্যায় ১৭](17-no-null-any-generics.md)), তাই একটি ভ্যালুর ভেতরে দেখতে হয় যাচাই করা
ফাংশন দিয়ে, `json.asString(v)`, `json.get(v, "name")` ইত্যাদি, আর আপনি যা চেয়েছেন
ভ্যালুটি তা না হলে প্রতিটি ফাংশন একটি এরর থ্রো করে।

## JSON পড়া

`json.parse` লেখাকে একটি ভ্যালুতে পরিণত করে। `json.get` একটি অবজেক্টের সদস্য পড়ে,
`json.at` একটি অ্যারের এলিমেন্ট, আর `as…` ফাংশনগুলো একটি ভ্যালুকে RuneScript-এর
`string`, `bool`, `int`, `int64` বা `float64`-এ পরিণত করে:

```actionscript
package main;

use "std:json";

protected int main() {
    string text = "{\"name\": \"Ada\", \"born\": 1815, \"langs\": [\"English\", \"French\"]}";
    json.Value person = json.parse(text);
    string name = json.asString(json.get(person, "name"));
    int born = json.asInt(json.get(person, "born"));
    trace(name, born);                                   // Ada 1815
    for each (json.Value lang in json.items(json.get(person, "langs"))) {
        trace(json.asString(lang));                      // English, then French
    }
    trace(json.has(person, "died"));                     // false
    return 0;
}
```

নিজের টাইপ নিয়ে কাজ করতে, দরকারি ফিল্ডগুলো একটি struct-এ কপি করুন:

```actionscript
package main;

use "std:json";

public struct Item {
    string name;
    int count;
    float64 price;
}

public Item readItem(json.Value v) {
    return Item{
        name: json.asString(json.get(v, "name")),
        count: json.asInt(json.get(v, "count")),
        price: json.asFloat(json.get(v, "price"))
    };
}

protected int main() {
    json.Value order = json.parse("[{\"name\": \"lamp\", \"count\": 2, \"price\": 19.5}, {\"name\": \"rug\", \"count\": 1, \"price\": 120}]");
    float64 total = 0.0;
    for each (json.Value v in json.items(order)) {
        Item it = readItem(v);
        total += (float64) it.count * it.price;
    }
    trace(total);   // 159
    return 0;
}
```

## পড়ার ফাংশন

| ফাংশন | ফেরত দেয় | বিবরণ |
|---|---|---|
| `kind(json.Value v)` | `string` | `"null"`, `"bool"`, `"number"`, `"string"`, `"array"`, অথবা `"object"` |
| `isNull(json.Value v)` | `bool` | `v` null কি না |
| `asBool(json.Value v)` | `bool` | একটি bool-এর মান |
| `asString(json.Value v)` | `string` | একটি স্ট্রিংয়ের মান |
| `asInt(json.Value v)` | `int` | একটি পূর্ণসংখ্যা, যা `int`-এ আঁটতে হবে |
| `asInt64(json.Value v)` | `int64` | একটি পূর্ণসংখ্যা, যা `int64`-এ আঁটতে হবে |
| `asFloat(json.Value v)` | `float64` | একটি সংখ্যা, নিকটতম `float64`-এ রাউন্ড করা |
| `length(json.Value v)` | `int` | একটি অ্যারের এলিমেন্টের, বা একটি অবজেক্টের সদস্যের সংখ্যা |
| `at(json.Value v, int i)` | `json.Value` | একটি অ্যারের এলিমেন্ট `i` |
| `items(json.Value v)` | `Array<json.Value>` | একটি অ্যারের এলিমেন্টগুলো, একটি নতুন RuneScript অ্যারে হিসেবে |
| `has(json.Value v, string key)` | `bool` | একটি অবজেক্টে `key` সদস্যটি আছে কি না |
| `get(json.Value v, string key)` | `json.Value` | একটি অবজেক্টের `key` সদস্য |
| `keys(json.Value v)` | `Array<string>` | একটি অবজেক্টের কী-গুলো, ডকুমেন্টের ক্রমে |

`asInt` ও `asInt64` কেবল পূর্ণসংখ্যা হিসেবে লেখা সংখ্যা নেয়: `2` নেয়, `2.0` ও `2e0`
নেয় না। সেগুলোর জন্য `asFloat` ব্যবহার করুন।

## JSON তৈরি ও লেখা

`from…` ফাংশনগুলো ভ্যালু তৈরি করে, আর `fromArray` ও `fromObject` সেগুলোকে একসঙ্গে
জোড়ে। `json.stringify` একটি ভ্যালুকে সংক্ষিপ্ত লেখা হিসেবে লেখে;
`json.stringifyIndent` প্রতিটি এলিমেন্ট ও সদস্যকে আলাদা লাইনে বসায়, আপনার দেওয়া
স্ট্রিং দিয়ে ইনডেন্ট করে:

```actionscript
package main;

use "std:json";

protected int main() {
    Array<json.Value> tags = [json.fromString("math"), json.fromString("poetry")];
    json.Value person = json.fromObject(
        ["name", "born", "tags", "spouse"],
        [json.fromString("Ada"), json.fromInt(1815), json.fromArray(tags), json.nullValue()]);
    trace(json.stringify(person));
    // {"name":"Ada","born":1815,"tags":["math","poetry"],"spouse":null}
    trace(json.stringifyIndent(person, "  "));
    return 0;
}
```

দ্বিতীয় `trace` প্রিন্ট করে:

```text
{
  "name": "Ada",
  "born": 1815,
  "tags": [
    "math",
    "poetry"
  ],
  "spouse": null
}
```

| ফাংশন | ফেরত দেয় | বিবরণ |
|---|---|---|
| `nullValue()` | `json.Value` | JSON-এর `null` |
| `fromBool(bool b)` | `json.Value` | `true` অথবা `false` |
| `fromString(string s)` | `json.Value` | একটি স্ট্রিং |
| `fromInt(int n)`, `fromInt64(int64 n)` | `json.Value` | একটি পূর্ণসংখ্যা |
| `fromFloat(float64 f)` | `json.Value` | একটি সংখ্যা, `trace` যেভাবে লেখে সেভাবে লেখা (`0.1`, `1e+21`) |
| `fromArray(Array<json.Value> items)` | `json.Value` | একটি অ্যারে |
| `fromObject(Array<string> keys, Array<json.Value> values)` | `json.Value` | একটি অবজেক্ট; দুটি অ্যারের দৈর্ঘ্য সমান হতে হবে |
| `stringify(json.Value v)` | `string` | সংক্ষিপ্ত JSON লেখা |
| `stringifyIndent(json.Value v, string indent)` | `string` | JSON লেখা, প্রতি লাইনে একটি এলিমেন্ট বা সদস্য |

স্ট্রিং লেখা হয় JavaScript-এর `JSON.stringify` যেভাবে লেখে সেভাবে: `"`, ব্যাকস্ল্যাশ ও
কন্ট্রোল ক্যারেক্টার এস্কেপ করা হয়, বাকি সব যেমন আছে তেমন লেখা হয়।

## ফাইল

একটি JSON ফাইল `fs.readFile` দিয়ে পড়ুন ([অধ্যায় ২০](20-files-and-input.md)) আর
`fs.writeFile` দিয়ে লিখুন:

```actionscript
package main;

use "std:fs";
use "std:json";
use "std:path";

protected int main() {
    string file = path.join(fs.tempDir(), "runescript-settings.json");
    fs.writeFile(file, "{\"volume\": 7, \"theme\": \"ice\"}");

    json.Value settings = json.parse(fs.readFile(file));
    int volume = json.asInt(json.get(settings, "volume"));
    json.Value changed = json.fromObject(["volume", "theme"], [json.fromInt(volume + 1), json.get(settings, "theme")]);
    fs.writeFile(file, json.stringifyIndent(changed, "    ") + "\n");
    trace(json.asInt(json.get(json.parse(fs.readFile(file)), "volume")));   // 8
    fs.remove(file);
    return 0;
}
```

## এরর

| সমস্যা | এরর |
|---|---|
| যে লেখা JSON নয় | `Error`: `json.parse: unexpected character ']' at byte 6` (অফসেট `0` থেকে UTF-8 বাইট গোনে) |
| ভুল ধরনের একটি ভ্যালু | `TypeError`: `json.asString: expected a string, found a number` |
| `1.5`-এর `asInt` | `TypeError`: `json.asInt: 1.5 is not an integer` |
| `int`, `int64` বা `float64`-এর জন্য খুব বড় একটি সংখ্যা | `RangeError`: `json.asInt: 3000000000 does not fit in 'int'` |
| শেষের পরে `at` | `RangeError`: `json.at: index 2 is out of range for an array of length 2` |
| নেই এমন কী-এর `get` | `Error`: `json.get: no key "phone"` (আগে `has` দিয়ে যাচাই করুন) |
| NaN বা অসীমের `fromFloat` | `ArgumentError`: `json.fromFloat: NaN cannot be written in JSON` |
| ভিন্ন দৈর্ঘ্যের অ্যারে দিয়ে `fromObject` | `ArgumentError`: `json.fromObject: 2 keys but 1 value` |

```actionscript
package main;

use "std:json";

protected int main() {
    try {
        json.Value v = json.parse("[1, 2,]");
        trace(json.length(v));
    } catch (Error e) {
        trace(e.message);   // json.parse: unexpected character ']' at byte 6
    }
    json.Value n = json.parse("12.5");
    try {
        trace(json.asString(n));
    } catch (TypeError e) {
        trace(e.message);   // json.asString: expected a string, found a number
    }
    return 0;
}
```

## খুঁটিনাটি

- **একটি ভ্যালু কখনো বদলায় না।** `items` ও `keys` নতুন অ্যারে ফেরত দেয়, আর কোনো
  ফাংশন একটি ভ্যালুকে জায়গায় বদলায় না: `fromArray` বা `fromObject` দিয়ে নতুন একটি
  তৈরি করুন। ভ্যালু নির্দ্বিধায় ভাগাভাগি করা যায়, আর সেগুলো নিজেকে ধারণ করতে পারে না।
- **সংখ্যা তাদের লেখা ধরে রাখে।** `12345678901234567890` হুবহু পড়ে `stringify` ও
  `asInt64` (যা জানায় যে এটি আঁটে না), আর কেবল `asFloat` রাউন্ড করে। পার্স করা একটি
  ডকুমেন্ট লিখলে সংখ্যাগুলো যেভাবে লেখা ছিল সেভাবেই ফেরত আসে: `1.50` থাকে `1.50`।
- **অবজেক্ট তাদের ক্রম ধরে রাখে।** `keys` ও `stringify` ডকুমেন্টের ক্রম মানে। কোনো কী
  দুবার এলে, অবজেক্ট প্রথম অবস্থান আর শেষ মান রাখে, JavaScript-এর মতো।
- **স্ট্রিং বৈধ ইউনিকোড।** `\u` এস্কেপ ডিকোড হয়, একটি সারোগেট জোড়া হয় একটি অক্ষর; একা
  একটি সারোগেট (`"\ud83d"`) হয়ে যায় `�` (U+FFFD), কারণ RuneScript-এর স্ট্রিং সবসময়
  বৈধ UTF-8।
- **কেবল কঠোর JSON।** কোনো মন্তব্য, শেষে অতিরিক্ত কমা, একক উদ্ধৃতি বা শুরুর শূন্য নেই।
  অ্যারে ও অবজেক্ট ১,০০০ স্তর পর্যন্ত নেস্ট করা যায়।
- **`json.Value`-এর কোনো শূন্য মান নেই**, তাই `json.Value v;` একটি এরর; এটিকে একটি মান
  দিন। ভ্যালু `==` দিয়ে তুলনা বা সরাসরি `trace`-এ পাঠানো যায় না; `json.stringify(v)`
  trace করুন।
- **দুই টার্গেট একমত।** ইন্টারপ্রেটার ও JavaScript টার্গেট একই পার্সার ব্যবহার করে, তাই
  একটি লেখা দুটিতেই একইভাবে গৃহীত বা বাতিল হয়, আর প্রতিটি এরর একইভাবে পড়া যায়।

পরবর্তী: আবার [গাইডের সূচিপত্রে](README.md)
