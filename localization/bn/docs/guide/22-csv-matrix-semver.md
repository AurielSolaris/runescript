<!-- source: docs/guide/22-csv-matrix-semver.md @ 298a2e7 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/22-csv-matrix-semver.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ২২. CSV, ম্যাট্রিক্স ও ভার্সন

প্রতিদিনের ডেটার জন্য তিনটি ছোট প্যাকেজ:

- `std:csv` কমা দিয়ে আলাদা করা মান (CSV) পড়ে ও লেখে, `Array<Array<string>>` হিসেবে।
- `std:matrix` `Array<Array<float64>>`-এর ওপর লিনিয়ার অ্যালজেব্রা করে।
- `std:semver` `1.4.0-rc.1`-এর মতো ভার্সন স্ট্রিং যাচাই করে, তুলনা করে ও মেলায়।

তিনটিই অ্যারে আর স্ট্রিংয়ের ওপর সাধারণ ফাংশন: নতুন কোনো টাইপ শিখতে হয় না। স্ট্যান্ডার্ড
লাইব্রেরির বাকি অংশের মতো এগুলো ইন্টারপ্রেটার আর JavaScript টার্গেট দুটোতেই একই ফল, আর
একই এরর মেসেজ দেয়।

## CSV: `std:csv`

`csv.parse` একটি লেখাকে রেকর্ডে ভাগ করে, আর প্রতিটি রেকর্ডকে ফিল্ডে। ডাবল কোটের ভেতরের
ফিল্ডে কমা, লাইন ব্রেক আর কোট (দুবার লেখা) থাকতে পারে:

```actionscript
package main;

use "std:csv";

protected int main() {
    string text = "name,city\nAda,London\n\"Turing, Alan\",\"Wilmslow\"\n";
    Array<Array<string>> rows = csv.parse(text);
    trace(rows.length);                      // 3
    for each (Array<string> row in rows) {
        trace(row[0], "|", row[1]);          // name | city, then Ada | London, ...
    }
    return 0;
}
```

একটি ফাইল পড়তে `std:fs`-এর সঙ্গে মেলান: `csv.parse(fs.readFile("data.csv"))`।

`csv.format` উল্টো দিকে কাজ করে। এটি কোনো ফিল্ডে কোট দেয় কেবল যখন দিতেই হয়, আর প্রতিটি
রেকর্ড একটি লাইন ব্রেক দিয়ে শেষ করে, তাই `csv.parse(csv.format(rows))` আবার `rows`
ফিরিয়ে দেয়:

```actionscript
package main;

use "std:csv";

protected int main() {
    Array<Array<string>> rows = [["item", "note"], ["lamp", "red, tall"], ["rug", "says \"hi\""]];
    trace(csv.format(rows));
    // item,note
    // lamp,"red, tall"
    // rug,"says ""hi"""
    return 0;
}
```

`csv.parseWith` আর `csv.formatWith` অন্য একটি বিভাজক নেয়, একটি `char` হিসেবে: ট্যাব দিয়ে
আলাদা করা ফাইলের জন্য `'\t'`, আর ইউরোপে অনেক স্প্রেডশিট যে ফাইল লেখে তার জন্য `';'`:

```actionscript
package main;

use "std:csv";

protected int main() {
    Array<Array<string>> rows = csv.parseWith("a\tb\n1\t2", '\t');
    trace(rows[1][1]);                       // 2
    trace(csv.formatWith(rows, ';'));        // a;b, then 1;2
    return 0;
}
```

নিয়মগুলো:

- রেকর্ড শেষ হয় `\n` বা `\r\n`-এ। খালি লাইন বাদ যায়, তাই খালি লেখায় কোনো রেকর্ড নেই।
- রেকর্ডে ফিল্ডের সংখ্যা আলাদা হতে পারে। আপনার ডেটায় নির্দিষ্ট সংখ্যা লাগলে `row.length`
  দেখে নিন।
- স্পেস ফিল্ডেরই অংশ: `" a "` থাকে `" a "`। কোটের ভেতরের ফিল্ড তাদের লাইন ব্রেক হুবহু রাখে।
- শুরুতে একটি UTF-8 বাইট অর্ডার মার্ক, যা কিছু স্প্রেডশিট লেখে, বাদ দেওয়া হয়।
- কোট ছাড়া ফিল্ডের ভেতরে একটি কোট (`ab"c`), বন্ধ কোটের পরে লেখা (`"a"b`), আর কখনো বন্ধ
  না হওয়া কোট হলো এরর। মেসেজ জানায় কোথায়: `csv.parse: line 2, column 4: …`। কলাম বাইট
  গোনে, স্ট্রিং অফসেটের মতো ([অধ্যায় ১০](10-std-string.md))।
- বিভাজক `"`, `\r` আর `\n` ছাড়া যেকোনো ক্যারেক্টার হতে পারে (`ArgumentError`)। কোনো
  ফিল্ড ছাড়া রেকর্ডের জন্য `csv.format` একটি `ArgumentError` থ্রো করে, কারণ CSV তা লিখতে
  পারে না।

## ম্যাট্রিক্স: `std:matrix`

একটি ম্যাট্রিক্স হলো সারির একটি অ্যারে: `[[1, 2], [3, 4]]` হলো `Array<Array<float64>>`।
প্রতিটি ফাংশন যাচাই করে যে সারিগুলোর দৈর্ঘ্য সমান আর মাপগুলো অপারেশনের সঙ্গে মেলে, আর
একটি নতুন অ্যারে ফেরত দেয়: আর্গুমেন্ট কখনো বদলায় না।

```actionscript
package main;

use "std:matrix";

protected int main() {
    Array<Array<float64>> a = [[1, 2], [3, 4]];
    Array<Array<float64>> b = [[0, 1], [1, 0]];
    trace(matrix.format(matrix.mul(a, b)));  // [2, 1], then [4, 3]
    trace(matrix.format(matrix.add(a, b)));  // [1, 3], then [4, 4]
    trace(matrix.det(a));                    // -2
    trace(matrix.rows(a), matrix.cols(a));   // 2 2
    return 0;
}
```

| ফাংশন | ফল |
|---|---|
| `zeros(rows, cols)`, `identity(n)` | শূন্য দিয়ে ভরা একটি নতুন ম্যাট্রিক্স, বা আইডেন্টিটি |
| `rows(m)`, `cols(m)` | মাপ (সারি ছাড়া ম্যাট্রিক্সে 0টি কলাম) |
| `add(a, b)`, `sub(a, b)`, `scale(m, k)` | উপাদান ধরে ধরে |
| `mul(a, b)`, `mulVec(m, v)` | ম্যাট্রিক্স গুণফল, আর একটি ম্যাট্রিক্স গুণ একটি ভেক্টর (`Array<float64>`) |
| `transpose(m)` | সারি হয়ে যায় কলাম |
| `det(m)`, `inverse(m)` | একটি বর্গ ম্যাট্রিক্সের ডিটারমিন্যান্ট আর ইনভার্স |
| `solve(a, b)` | সেই `x` যার জন্য `a·x = b`, একটি বর্গ `a`-এর জন্য |
| `format(m)` | প্রতি সারিতে একটি লাইন: `[1, 2.5]` |

ইনভার্স দিয়ে গুণ করার চেয়ে `solve` ভালো: এটি কম কাজ করে আর কম নির্ভুলতা হারায়। এখানে এটি
দুটি রসিদ থেকে দুটি জিনিসের দাম বের করে:

```actionscript
package main;

use "std:matrix";

protected int main() {
    // 2 apples + 1 pear cost 4; 1 apple + 3 pears cost 7.
    Array<Array<float64>> counts = [[2, 1], [1, 3]];
    Array<float64> prices = matrix.solve(counts, [4.0, 7.0]);
    trace(prices[0], prices[1]);             // 1 2
    return 0;
}
```

`det`, `inverse` আর `solve` আংশিক পিভটিং সহ গাউসিয়ান এলিমিনেশন ব্যবহার করে। একটি
সিঙ্গুলার ম্যাট্রিক্স (যার ইনভার্স নেই) ডিটারমিন্যান্ট দেয় `0`, আর `inverse` ও `solve`
একটি `Error` থ্রো করে। সব ফ্লোটিং-পয়েন্ট কাজের মতো, ফল শেষের অঙ্কে একটু ভুল হতে পারে:
`0.1 + 0.2` হলো `0.30000000000000004`। দুটি টার্গেটেই ভুলটা হুবহু একই রকম, কারণ দুটোই একই
ক্রমে একই অপারেশন চালায়।

মাপের ভুল একটি `ArgumentError` থ্রো করে যা মাপগুলোর নাম বলে:

```text
matrix.mul: cannot multiply a 1x2 matrix by a 1x2 matrix: the first has 2 columns, the second 1 row
```

`matrix` এখনো একটি সংরক্ষিত শব্দ: কোনো ভ্যারিয়েবলের নাম `matrix` দেওয়া যায় না। এটি কেবল
একটি কলকে চিহ্নিত করতে পারে, যেমন `matrix.mul(a, b)`।

## ভার্সন: `std:semver`

`std:semver` [Semantic Versioning 2.0.0](https://semver.org/) স্ট্রিং নিয়ে কাজ করে:
`MAJOR.MINOR.PATCH`, ইচ্ছে হলে তারপর একটি প্রিরিলিজ (`-rc.1`) আর বিল্ড মেটাডেটা
(`+build.5`)। এটি ভার্সন তুলনা করে আর রেঞ্জ মেলায় npm যেভাবে করে।

ভার্সনকে স্ট্রিং হিসেবে তুলনা করলে ভুল উত্তর আসে, কারণ অক্ষর ধরে ধরে `"1.10.0"` আসত
`"1.9.0"`-এর আগে। `semver` ফাংশনগুলো সংখ্যাকে সংখ্যা হিসেবে তুলনা করে, আর প্রিরিলিজকে তার
রিলিজের আগে রাখে:

```actionscript
package main;

use "std:semver";

protected int main() {
    trace(semver.lt("1.9.0", "1.10.0"));             // true
    trace(semver.lt("1.0.0-rc.1", "1.0.0"));         // true
    trace(semver.compare("2.0.0", "1.99.0"));        // 1
    trace(semver.eq("1.0.0+build.1", "1.0.0"));      // true: build metadata is ignored
    trace(semver.valid("1.2"), semver.valid("1.2.3-rc.1")); // false true
    return 0;
}
```

| ফাংশন | ফল |
|---|---|
| `valid(v)`, `validRange(r)` | `v` একটি ভার্সন কি না, বা `r` একটি রেঞ্জ কি না (কখনো থ্রো করে না) |
| `compare(a, b)` | `-1`, `0`, অথবা `1` |
| `eq`, `neq`, `gt`, `gte`, `lt`, `lte` | `bool` তুলনা |
| `satisfies(v, range)` | `v` রেঞ্জের ভেতরে কি না |
| `major(v)`, `minor(v)`, `patch(v)` | সংখ্যাগুলো, `int` হিসেবে |
| `prerelease(v)`, `build(v)` | `"rc.1"`, `"build.5"`, অথবা `""` |

`satisfies` npm-এর রেঞ্জ বোঝে:

```actionscript
package main;

use "std:semver";

protected int main() {
    trace(semver.satisfies("1.4.0", "^1.2.0"));            // true: 1.x from 1.2.0 on
    trace(semver.satisfies("2.0.0", "^1.2.0"));            // false
    trace(semver.satisfies("1.2.9", "~1.2.3"));            // true: 1.2.x from 1.2.3 on
    trace(semver.satisfies("1.5.0", ">=1.2.0 <2.0.0"));    // true
    trace(semver.satisfies("3.1.0", "1.x || >=3.0.0"));    // true
    trace(semver.satisfies("2.3.9", "1.2.3 - 2.3"));       // true
    trace(semver.satisfies("1.3.0-beta", "^1.2.0"));       // false: see below
    return 0;
}
```

| রেঞ্জ | অর্থ |
|---|---|
| `1.2.3`, `=1.2.3` | ঠিক সেই ভার্সন |
| `>1.2.3`, `>=`, `<`, `<=` | তুলনা; স্পেস দিয়ে আলাদা করা কয়েকটি থাকলে সবগুলোই সত্য হতে হবে |
| `a \|\| b` | যেকোনো একটি রেঞ্জ |
| `1.2.x`, `1.x`, `*`, `1.2` | দেওয়া অংশগুলোর সঙ্গে মেলে এমন যেকোনো ভার্সন |
| `~1.2.3` | `>=1.2.3 <1.3.0`: প্যাচ আপডেট |
| `^1.2.3` | `>=1.2.3 <2.0.0`: কোনো ভাঙা পরিবর্তন নেই (`^0.2.3` হলো `<0.3.0`) |
| `1.2.3 - 2.3.4` | `>=1.2.3 <=2.3.4` |

npm-এর মতো, একটি প্রিরিলিজ ভার্সন কোনো রেঞ্জে পড়ে কেবল যদি রেঞ্জটি একই
`MAJOR.MINOR.PATCH`-এর একটি প্রিরিলিজের নাম বলে: `1.3.0-beta` `^1.2.0`-এ পড়ে না, কিন্তু
`1.2.3-beta.4` পড়ে `^1.2.3-beta.2`-এ। এভাবে পরীক্ষামূলক ভার্সন কখনো ভুল করে বেছে নেওয়া
হয় না।

নিয়মগুলো:

- ভার্সন কঠোর SemVer 2.0.0: সামনে শূন্য নয় (`01.2.3`), `v` উপসর্গ নয় (`v1.2.3`), স্পেস নয়।
  সংখ্যায় যত খুশি অঙ্ক থাকতে পারে; কোনোটি `int`-এ না ধরলে `major`, `minor` আর `patch` একটি
  `RangeError` থ্রো করে।
- `valid` আর `validRange` ছাড়া প্রতিটি ফাংশন অবৈধ ভার্সন বা রেঞ্জের জন্য একটি
  `ArgumentError` থ্রো করে: `semver.compare: "1.2" is not a valid version: …`।

পরবর্তী: [এডিটর](23-editors.md)
