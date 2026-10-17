<!-- source: docs/guide/13-floats-and-conversions.md @ 362a461 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/13-floats-and-conversions.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ১৩. ফ্লোট ও রূপান্তর

## ফ্লোটিং-পয়েন্ট সংখ্যা

RuneScript-এ ফ্লোটিং-পয়েন্ট টাইপ তিনটি:

| টাইপ | এটি কী |
|---|---|
| `float` | রোজকার ফ্লোট। এটি শুরু হয় 32-বিট ফ্লোট হিসেবে, এবং যখনই কোনো মানের প্রয়োজন হয় তখনই 64 বিটে প্রশস্ত হয়, তাই কখনো কোনো মান হারায় না। |
| `float32` | হুবহু একটি IEEE 32-বিট ফ্লোট। প্রতিটি ফল 32 বিটে রাউন্ড করা হয়। |
| `float64` | হুবহু একটি IEEE 64-বিট ফ্লোট, ActionScript-এর `Number`-এর মতো। |

```actionscript
package main;

protected int main() {
    float a = 0.1;
    float b = 0.2;
    trace(a + b);          // 0.30000000000000004
    trace(7.0 / 2, 1e21);  // 3.5 1e+21

    float32 f = (float32) 0.1;
    trace(f * 3);          // 0.3, rounded to float32
    return 0;
}
```

`float` রাউন্ড না করে প্রশস্ত হয় বলে, `float`-এর হিসাব সবসময় 64-বিট গাণিতিকের
সমান ফল দেয়। 32-বিট রাউন্ডিং চাইলে `float32` ব্যবহার করুন।

লিটারাল লেখা হয় `64.3`, `1e10`, বা `2.5e-3` হিসেবে। সংখ্যাগুলো সেভাবেই দেখায়
যেভাবে ActionScript দেখায়: `64.0` দেখায় `64`, আর খুব বড় বা খুব ছোট সংখ্যা
ঘাতের (exponent) রূপ ব্যবহার করে।

### `NaN` ও `Infinity`

ফ্লোট IEEE 754 মেনে চলে, ActionScript-এর `Number`-এর মতোই। টাইপের তুলনায় খুব বড়
ফল হলো `Infinity`, আর `0 / 0` হলো `NaN` ("not a number", অর্থাৎ সংখ্যা নয়)।
পূর্ণসংখ্যার ওভারফ্লোর মতো নয়, এগুলো কখনো প্রোগ্রাম থামায় না।

```actionscript
package main;

protected int main() {
    float zero = 0;
    float nan = zero / zero;
    trace(1 / zero, -1 / zero, nan);           // Infinity -Infinity NaN
    trace(nan == nan, isNaN(nan));             // false true
    trace(isFinite(1e308 * 10), Infinity > 1); // false true
    return 0;
}
```

`NaN` কোনো কিছুর সমান নয়, এমনকি নিজেরও নয়, তাই এটি যাচাই করুন `isNaN(x)` দিয়ে।
`NaN` ও দুটি অসীমের জন্য `isFinite(x)` মিথ্যা। `NaN` ও `Infinity` বিল্ট-ইন
কনস্ট্যান্ট; একই নামের একটি ভেরিয়েবল এদের আড়াল করে।

### কোনো অন্তর্নিহিত রূপান্তর নেই

পূর্ণসংখ্যা ও ফ্লোট মেশে না, ভিন্ন ভিন্ন ফ্লোট টাইপও মেশে না। একটি দিককে স্পষ্টভাবে
রূপান্তর করুন:

```actionscript
package main;

protected int main() {
    int n = 3;
    float x = 1.5;
    trace(x * n);   // error: ASX-TYPE-012
    return 0;
}
```

```actionscript
package main;

protected int main() {
    int n = 3;
    float x = 1.5;
    trace(x * (float) n);   // 4.5
    return 0;
}
```

সংখ্যার লিটারাল মানিয়ে নেয়: `x * 2`-তে `2` একটি ফ্লোট হয়ে যায়। তবে একটি লিটারালকে
তার টাইপে হুবহু আঁটতে হবে। `float32 f = 0.1;` একটি এরর (`ASX-TYPE-018`), কারণ
0.1-এর কোনো হুবহু float32 মান নেই। ইচ্ছে করে রাউন্ড করতে লিখুন `(float32) 0.1`।

`%`, `++` ও `--` শুধু পূর্ণসংখ্যায় কাজ করে। ফ্লোটের জন্য লিখুন `x += 1;`।

## রূপান্তর

`(T) value` একটি মানকে `T`-তে রূপান্তর করে, C-র কাস্টের মতো। `T` হতে পারে `int`,
`sint`, `uint`, `float`, `float32`, `float64`, `string` বা `bool`।

```actionscript
package main;

protected int main() {
    trace((int) 64.3, (int) -64.7);       // 64 -64: truncates toward zero
    trace((float) "64.7", (int) "-42");   // 64.7 -42: parses strings
    trace((string) 1.5, (bool) "true");   // 1.5 true
    int count = 7;
    trace((float) count / 2);             // 3.5: the cast binds tighter than /
    return 0;
}
```

প্রতিটি রূপান্তর **যাচাই করা হয়**। এটি কখনো ঘুরে যায় না বা অনুমান করে না। যে মান
আঁটে না, অথবা যে স্ট্রিং সংখ্যা নয়, তা একটি এরর throw করে:

| রূপান্তর | কখন ব্যর্থ হয় | এরর |
|---|---|---|
| পূর্ণসংখ্যায় | মানটি সীমার বাইরে (যেমন `(uint) -1.5`, `(int) 3e10`), `NaN`, অথবা অসীম | `RangeError` |
| স্ট্রিং থেকে সংখ্যায় | স্ট্রিংটি সাধারণ দশমিক সংখ্যা নয় (অথবা ফ্লোটের জন্য `"NaN"`, `"Infinity"`, `"-Infinity"` নয়), যেমন `"12x"`, `" 5"`, `"0x10"` | `TypeError` |
| স্ট্রিং থেকে `bool`-এ | স্ট্রিংটি হুবহু `"true"` বা `"false"` নয় | `TypeError` |

```actionscript
package main;

protected int main() {
    string input = "12x";
    trace((int) input);   // runtime error: ASX-RUN-006
    return 0;
}
```

মানটি কনস্ট্যান্ট হলে রূপান্তরটি কম্পাইলের সময়েই ঘটে, এররটিও: `(uint) -2.5` হলো
`ASX-TYPE-007`।

`bool` ও সংখ্যা একে অপরে রূপান্তরিত হয় না। বরং তুলনা করুন (`n != 0`)।

[অধ্যায় ৩](03-types-and-variables.md)-এর সাইনড/আনসাইনড নিয়মটি এবার তার স্পষ্ট রূপ
পেল:

```actionscript
package main;

protected int main() {
    int signedValue = 10;
    uint u = (uint) signedValue;   // checked: a negative value would throw
    trace(u);
    return 0;
}
```

পরবর্তী: [এরর](14-errors.md)
