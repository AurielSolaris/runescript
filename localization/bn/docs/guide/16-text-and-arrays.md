<!-- source: docs/guide/16-text-and-arrays.md @ c142fe7 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/16-text-and-arrays.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ১৬. লেখা ও তালিকা: স্ট্রিং জোড়া লাগানো, `char`, এবং অ্যারে

এই অধ্যায়ে স্ক্রিপ্ট লেখার রোজকার হাতিয়ারগুলো আছে: স্ট্রিং তৈরি করা, একটি করে
অক্ষর নিয়ে কাজ করা, আর মানগুলো একটি তালিকায় রাখা। স্ট্রিং খোঁজা, ভাগ করা ও বদলানোর
ফাংশনগুলো আছে [অধ্যায় ১০](10-std-string.md)-এ।

## `+` দিয়ে স্ট্রিং জোড়া লাগানো

`+` দুটি স্ট্রিং জোড়া লাগায়, আর `+=` একটি স্ট্রিং ভেরিয়েবলের শেষে যোগ করে:

```actionscript
package main;

protected int main() {
    string name = "Ada";
    string greeting = "Hello, " + name + "!";
    greeting += " Welcome.";
    trace(greeting);   // Hello, Ada! Welcome.
    return 0;
}
```

`+` কখনো নিজে থেকে কোনো সংখ্যাকে লেখায় রূপান্তর করে না, কারণ RuneScript আপনার
অজান্তে কোনো মান রূপান্তর করে না। নিজে `(string) x` দিয়ে রূপান্তর করুন, অথবা
ইন্টারপোলেটেড স্ট্রিং ব্যবহার করুন (পরের অংশ):

```actionscript
package main;

protected int main() {
    int count = 3;
    trace("count: " + (string) count);   // count: 3
    trace("count: " + count);            // error: ASX-TYPE-012
    return 0;
}
```

এটি ActionScript 3 ও JavaScript থেকে আলাদা, যেখানে `"count: " + 3` চুপচাপ
`"count: 3"` হয়ে যায়।

## ইন্টারপোলেটেড স্ট্রিং: `$"…"`

একটি স্ট্রিংয়ের আগে `$` বসালে তার ভেতরের প্রতিটি `{expression}`-এর জায়গায় সেই
এক্সপ্রেশনের মান বসে:

```actionscript
package main;

protected int main() {
    string item = "apple";
    int count = 3;
    float price = 0.25;
    trace($"{count} x {item} at {price} = {(float) count * price}");   // 3 x apple at 0.25 = 0.75
    trace($"ready: {count > 0}");                                     // ready: true
    trace($"{{braces}} are written twice");                           // {braces} are written twice
    return 0;
}
```

- একটি ফাঁকে যেকোনো এক্সপ্রেশন থাকতে পারে: গণনা, কল, অন্য স্ট্রিং, এমনকি আরেকটি
  ইন্টারপোলেটেড স্ট্রিংও।
- এর মান ঠিক সেভাবেই দেখানো হয় যেভাবে `(string) x` দেখায়, তাই সেটি একটি সংখ্যা,
  `bool`, `char` বা স্ট্রিং হতে হবে। অন্য কিছু, যেমন ফাংশন বা অ্যারে, হলো
  `ASX-TYPE-017`।
- `{{` ও `}}` মানে আক্ষরিক `{` ও `}`। একটি একক `}` এরর (`ASX-LEX-003`)।
- এস্কেপ অন্য স্ট্রিংয়ের মতোই কাজ করে: `\n`, `\t`, `\"`, ইত্যাদি।

## অক্ষর: `char`

একটি `char` হলো একটি অক্ষর, একটি Unicode কোড পয়েন্ট। এটি একক উদ্ধৃতিচিহ্নের মধ্যে
লিখুন:

```actionscript
package main;

protected int main() {
    char letter = 'a';
    char accent = 'é';
    char emoji = '😀';
    char quote = '\'';
    trace(letter, accent, emoji, quote, '\\');   // a é 😀 ' \
    trace(letter < 'b', accent == 'é');          // true true
    return 0;
}
```

- একটি char লিটারালে ঠিক একটি অক্ষর থাকে। `''` ও `'ab'` হলো `ASX-LEX-006`।
  এস্কেপগুলো স্ট্রিংয়ের মতোই, সঙ্গে `\'`।
- char-গুলোকে `==`, `!=`, `<`, `<=`, `>` ও `>=` দিয়ে কোড পয়েন্ট অনুযায়ী তুলনা
  করা যায়।
- `char` কোনো সংখ্যা নয়, তাই `'a' + 1` হয় না। আগে রূপান্তর করুন।

প্রতিটি রূপান্তরের মতো এগুলোও যাচাই করা হয়
([অধ্যায় ১৩](13-floats-and-conversions.md)):

| রূপান্তর | ফল | ব্যর্থ হয় |
|---|---|---|
| `(int) c`, `(uint) c` | কোড পয়েন্ট: `(int) 'a'` হলো `97` | কখনো না |
| `(char) n` | কোড পয়েন্ট `n`-এর অক্ষর: `(char) 98` হলো `b` | `RangeError`, যদি `n` বৈধ কোড পয়েন্ট না হয় (ঋণাত্মক, `0x10FFFF`-এর বেশি, অথবা একটি সারোগেট `0xD800`–`0xDFFF`) |
| `(string) c` | এক অক্ষরের একটি স্ট্রিং | কখনো না |
| `(char) s` | `s`-এর একমাত্র অক্ষর | `TypeError`, যদি না `s`-এ ঠিক একটি অক্ষর থাকে |

```actionscript
package main;

public char next(char c) {
    return (char) ((int) c + 1);
}

protected int main() {
    trace(next('a'), (int) 'A', (char) "é");   // b 65 é
    trace((string) 'x' + "yz");               // xyz
    return 0;
}
```

`std:string`-এর অক্ষর-ফাংশনগুলো char নেয়: দেখুন
[অক্ষর ও কলব্যাক](10-std-string.md)।

## অ্যারে: `Array<T>`

একটি `Array<T>` হলো `T` টাইপের মানের একটি তালিকা, যা বাড়তে ও কমতে পারে। বর্গবন্ধনী
দিয়ে একটি তৈরি করুন:

```actionscript
package main;

protected int main() {
    Array<string> names = ["Ada", "Grace"];
    names.push("Linus");                      // add to the end
    trace(names.length, names[0], names[2]);  // 3 Ada Linus

    names[1] = "Hopper";                      // replace an element
    names[0] += "!";
    string last = names.pop();                // remove the last one
    trace(last, names.length, names[1]);      // Linus 2 Hopper
    return 0;
}
```

| অপারেশন | অর্থ |
|---|---|
| `[a, b, c]` | একটি নতুন অ্যারে; `[]` একটি খালি অ্যারে |
| `a[i]` | ইনডেক্স `i`-এর (একটি `int` বা `uint`) এলিমেন্ট, `0` থেকে গোনা |
| `a[i] = v`, `a[i] += v`, … | একটি এলিমেন্ট বদলানো |
| `a.length` | এলিমেন্টের সংখ্যা |
| `a.push(v)` | শেষে `v` যোগ করা |
| `a.pop()` | শেষ এলিমেন্টটি সরিয়ে ফেরত দেওয়া |

এলিমেন্ট টাইপ `T` যেকোনো মান-টাইপ হতে পারে: সংখ্যা, `bool`, `string`, `char`,
ফাংশন, এরর, এবং অন্য অ্যারে (`Array<Array<int>>`)।

**টাইপ আসে অ্যারেটি যেখানে যায় সেখান থেকে।** `Array<uint> u = [1, 2];` লিখলে `1`
ও `2` হয় `uint`। আশেপাশে কোনো ঘোষিত টাইপ না থাকলে লিটারালটি তার প্রথম
এলিমেন্টের টাইপ নেয়, তাই `[1, 2]` একটি `Array<int>`। খালি `[]`-এর জন্য একটি
ঘোষিত টাইপ লাগে: `Array<string> names = [];`।

**ইনডেক্স যাচাই করা হয়।** `0`-এর কম, বা `length`-এর কম নয় এমন ইনডেক্স, এবং খালি
অ্যারেতে `pop()`, একটি `RangeError` (`ASX-RUN-011`) ছোড়ে, যা ধরা যায়:

```actionscript
package main;

protected int main() {
    Array<int> scores = [90, 75];
    try {
        trace(scores[2]);
    } catch (RangeError e) {
        trace(e.message);   // index 2 is out of range for an array of length 2
    }
    return 0;
}
```

**অ্যারের ওপর লুপ চালান** একটি ইনডেক্স দিয়ে:

```actionscript
package main;

public int sum(Array<int> values) {
    int total = 0;
    for (int i = 0; i < values.length; i++) {
        total += values[i];
    }
    return total;
}

protected int main() {
    trace(sum([3, 4, 5]));   // 12
    return 0;
}
```

**অথবা `for each` দিয়ে লুপ চালান**, যা প্রতিটি এলিমেন্টে ক্রমানুসারে যায়। `in`-এর
আগের ভেরিয়েবলটি একটি নতুন ভেরিয়েবল, যা একবারে একটি করে এলিমেন্ট ধরে রাখে:

```actionscript
package main;

protected int main() {
    Array<string> names = ["Ada", "Grace", "Linus"];
    for each (string name in names) {
        trace("hello", name);
    }
    return 0;
}
```

লুপ ভেরিয়েবলের টাইপ এলিমেন্টের টাইপের সঙ্গে মিলতে হবে: `for each (uint x in
ints)` একটি এরর, যেমন `uint x = ints[0];` এরর হতো। `break` ও `continue` অন্য
লুপের মতোই কাজ করে। অ্যারের এক্সপ্রেশনটি একবারই মূল্যায়ন হয়; প্রতিটি ধাপের আগে
দৈর্ঘ্য যাচাই হয়, তাই লুপের ভেতরে push করা এলিমেন্টেও লুপ পৌঁছায়। `each` ও `in`
সংরক্ষিত শব্দ নয়: শুধু এই জায়গাতেই এদের বিশেষ অর্থ আছে, তাই `in` এখনো একটি বৈধ
ভেরিয়েবলের নাম।

**অ্যারে ভাগাভাগি হয়, কপি হয় না।** একটি অ্যারে অন্য ভেরিয়েবলে অ্যাসাইন করলে বা
কোনো ফাংশনে পাঠালে *একই* অ্যারের আরেকটি নাম পাওয়া যায়, ActionScript 3 ও
JavaScript-এর মতো। এক নাম দিয়ে করা পরিবর্তন সব নামেই দেখা যায়:

```actionscript
package main;

public void addDefaults(Array<string> list) {
    list.push("default");
}

protected int main() {
    Array<string> mine = ["first"];
    Array<string> same = mine;
    addDefaults(same);
    trace(mine.length, mine[1]);   // 2 default
    return 0;
}
```

অ্যারে সরাসরি trace করা যায় না। `string.join` দিয়ে সেগুলোকে একটি স্ট্রিংয়ে জোড়া
লাগান ([অধ্যায় ১০](10-std-string.md)):

```actionscript
package main;

use "std:string";

protected int main() {
    Array<string> words = string.fields("  the quick  brown fox ");
    trace(words.length, string.join(words, "-"));   // 4 the-quick-brown-fox
    return 0;
}
```

## নেস্টেড অ্যারে

একটি অ্যারে অন্য অ্যারে ধরে রাখতে পারে: `Array<Array<int>>` হলো সারির একটি তালিকা,
আর প্রতিটি সারি নিজেই একটি `Array<int>`। একটি এলিমেন্টে পৌঁছাতে দুবার ইনডেক্স
করুন। সারিগুলোর দৈর্ঘ্য আলাদা হতে পারে:

```actionscript
package main;

protected int main() {
    Array<Array<int>> grid = [[1, 2, 3], [4, 5], []];
    grid[1][0] = 40;
    grid[2].push(6);
    int total = 0;
    for each (Array<int> row in grid) {
        for each (int v in row) {
            total += v;
        }
    }
    trace(grid.length, grid[1][0], total);   // 3 40 57
    return 0;
}
```

সারিগুলোও অ্যারে, তাই সেগুলোও ভাগাভাগি হয়: `Array<int> row = grid[0];` দিয়ে
`grid[0]`-এর একই সারির আরেকটি নাম পাওয়া যায়। নির্দিষ্ট মাপের একটি টেবিল বানাতে
লুপে নতুন সারি push করুন:

```actionscript
package main;

protected int main() {
    Array<Array<string>> table = [];
    for (int r = 0; r < 2; r++) {
        Array<string> row = [];
        for (int c = 0; c < 3; c++) {
            row.push($"{r},{c}");
        }
        table.push(row);
    }
    trace(table.length, table[0].length, table[1][2]);   // 2 3 1,2
    return 0;
}
```

## অ্যারের মেথড

| মেথড | ফল | অর্থ |
|---|---|---|
| `a.insert(i, v)` | `void` | ইনডেক্স `i`-এর আগে `v` বসানো (`0` থেকে `length`) |
| `a.removeAt(i)` | `T` | `i`-এর এলিমেন্টটি সরিয়ে ফেরত দেওয়া |
| `a.slice(start, end)` | `Array<T>` | `start` থেকে `end`-এর আগ পর্যন্ত (`end` বাদে) এলিমেন্ট নিয়ে একটি নতুন অ্যারে |
| `a.indexOf(v)` | `int` | `v`-এর সমান প্রথম এলিমেন্টের ইনডেক্স, অথবা `-1` |
| `a.contains(v)` | `bool` | কোনো এলিমেন্ট `v`-এর সমান কি না |
| `a.reverse()` | `void` | ক্রম উল্টে দেওয়া, একই অ্যারেতে |
| `a.sort()` | `void` | একই অ্যারেতে সাজানো, ছোটটি আগে |
| `a.sort(less)` | `void` | একটি ফাংশন `bool less(T x, T y)` অনুযায়ী একই অ্যারেতে সাজানো |

```actionscript
package main;

use "std:string";

protected int main() {
    Array<string> fruit = ["pear", "fig", "apple"];
    fruit.insert(1, "kiwi");
    trace(string.join(fruit, " "));             // pear kiwi fig apple
    trace(fruit.indexOf("fig"), fruit.contains("plum"));   // 2 false
    string gone = fruit.removeAt(0);
    Array<string> firstTwo = fruit.slice(0, 2);
    trace(gone, string.join(firstTwo, " "));    // pear kiwi fig
    fruit.sort();
    trace(string.join(fruit, " "));             // apple fig kiwi
    fruit.sort((string x, string y) => string.len(x) > string.len(y));
    trace(string.join(fruit, " "));             // apple kiwi fig
    return 0;
}
```

- **ভুল ইনডেক্স এরর থ্রো করে।** `insert` নেয় `0` থেকে `length`, `removeAt` একটি
  বিদ্যমান ইনডেক্স, আর `slice`-এর দরকার `0 <= start <= end <= length`; অন্য কিছু
  হলে একটি `RangeError` থ্রো হয়।
- **`slice` কপি করে।** নতুন অ্যারে বদলালে পুরোনোটি বদলায় না।
- **`indexOf` ও `contains` `==` দিয়ে তুলনা করে**, তাই এগুলো সেই টাইপগুলোতে কাজ
  করে যেগুলোতে `==` কাজ করে: সংখ্যা, `bool`, `string`, `char`, এবং এদের দিয়ে গড়া
  struct। অ্যারের অ্যারে, ফাংশন বা এররের অ্যারেতে এগুলো নেই।
- **ফাংশন ছাড়া `sort()`** সংখ্যা, `char` ও স্ট্রিংয়ে কাজ করে। স্ট্রিং সাজানো হয়
  তাদের UTF-8 বাইট অনুযায়ী (`string.compare`), তাই `"Z"` আসে `"a"`-এর আগে।
- **`sort(less)`** যেকোনো এলিমেন্ট টাইপে কাজ করে। `x`-কে `y`-এর আগে রাখতে হলে
  `less(x, y)` `true` ফেরত দেয়। সাজানোটি *স্থিতিশীল* (stable): যে এলিমেন্টগুলো
  একে অপরের চেয়ে ছোটও নয় বড়ও নয়, সেগুলোর ক্রম বজায় থাকে। `less` থ্রো করলে
  অ্যারেটি আগের মতোই থাকে।

> **এখনো নেই:** `a[i]++`, `==` দিয়ে অ্যারে তুলনা, সরাসরি অ্যারে trace করা, এবং
> JavaScript-এর `extern` ফাংশনে অ্যারে পাঠানো। `main` কমান্ড-লাইন আর্গুমেন্ট নিতে
> পারে না; `process.args()` ব্যবহার করুন ([অধ্যায় ২০](20-files-and-input.md))।

পরবর্তী: [`null` নেই, `any` নেই, জেনেরিক নেই](17-no-null-any-generics.md)
