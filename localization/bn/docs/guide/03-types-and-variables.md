<!-- source: docs/guide/03-types-and-variables.md @ 6704ceb -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/03-types-and-variables.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ৩. টাইপ ও ভেরিয়েবল

## ভেরিয়েবল ঘোষণা

আগে টাইপ, তারপর নাম, তারপর প্রাথমিক মান লিখুন:

```actionscript
package main;

protected int main() {
    int count = 10;
    uint total = 250;
    bool ready = true;
    string name = "Ada";
    trace(name, count, total, ready);   // Ada 10 250 true
    return 0;
}
```

প্রাথমিক মান ছাড়া ঘোষিত ভেরিয়েবল তার টাইপের **শূন্য মান (zero value)** দিয়ে শুরু হয়:

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

protected int main() {
    int count;            // 0
    float ratio;          // 0
    string name;          // ""
    bool ready;           // false
    char initial;         // the character U+0000
    Array<string> names;  // a new, empty array
    Point origin;         // Point{x: 0, y: 0}: every field at its zero value
    trace(count, ratio, "[" + name + "]", ready, names.length, origin);
    return 0;                            // 0 0 [] false 0 Point{x: 0, y: 0}
}
```

ফাংশন টাইপ, এরর এবং রুনের কোনো শূন্য মান নেই: "এখনো সেট হয়নি" বোঝানোর জন্য
RuneScript-এ কোনো `null` নেই ([অধ্যায় ১৭](17-no-null-any-generics.md))। এগুলোকে মান ছাড়া
ঘোষণা করা একটি এরর (`ASX-TYPE-028`):

```actionscript
package main;

protected int main() {
    int(int) twice; // error: ASX-TYPE-028
    return 0;
}
```

## মৌলিক টাইপগুলো

| টাইপ | যা ধারণ করে | পরিসর |
|---|---|---|
| `int` | সাইনড (চিহ্নযুক্ত) পূর্ণসংখ্যা | −2147483648 … 2147483647 |
| `sint` | **হুবহু `int`-এর মতোই একই টাইপ।** "সাইনড" কথাটি স্পষ্ট করে বলতে চাইলে এটি ব্যবহার করুন। | `int`-এর মতোই |
| `uint` | আনসাইনড (চিহ্নহীন) পূর্ণসংখ্যা | 0 … 4294967295 |
| `int8` … `int64`, `uint8` … `uint64` | নির্দিষ্ট প্রস্থের পূর্ণসংখ্যা ([অধ্যায় ১৮](18-numbers.md)) | যেমন `uint8` হলো 0 … 255 |
| `float`, `float32`, `float64` | দশমিক (ফ্লোটিং-পয়েন্ট) সংখ্যা ([অধ্যায় ১৩](13-floats-and-conversions.md)) | |
| `bool` | `true` অথবা `false` | |
| `string` | লেখা, যেমন `"hello\n"` | |
| `char` | একটি অক্ষর, যেমন `'a'` ([অধ্যায় ১৬](16-text-and-arrays.md)) | যেকোনো Unicode কোড পয়েন্ট |
| `Array<T>` | `T` মানের একটি তালিকা, যেমন `["a", "b"]` ([অধ্যায় ১৬](16-text-and-arrays.md)) | |
| একটি `struct` | একটি মানে নামযুক্ত ফিল্ড, যেমন `Point{x: 1, y: 2}` ([অধ্যায় ১৯](19-structs.md)) | |
| `Error` ও এর প্রকারভেদ | এরর ([অধ্যায় ১৪](14-errors.md)) | |
| `void` | কোনো মান নয়; শুধু ফাংশনের রিটার্ন টাইপ হিসেবে ব্যবহৃত | |

`int` ও `sint` অবাধে মেশানো যায়:

```actionscript
package main;

protected int main() {
    sint a = -5;
    int b = a;      // fine: same type
    sint c = b;     // fine
    trace(a, b, c);
    return 0;
}
```

স্ট্রিংয়ের এস্কেপ: `\n` (নতুন লাইন), `\t` (ট্যাব), `\r`, `\0`, `\\`, `\"`, `\'`, এবং যেকোনো
Unicode অক্ষরের জন্য `\u{...}`, যেমন `\u{E9}` (é) বা `\u{1F600}`। স্ট্রিং
জোড়া লাগে `+` দিয়ে, আর `$"total: {n}"` মানগুলো থেকে একটি স্ট্রিং তৈরি করে
([অধ্যায় ১৬](16-text-and-arrays.md))।

## অ্যাসাইনমেন্ট

```actionscript
package main;

protected int main() {
    int x = 1;
    x = 10;     // 10
    x += 5;     // 15
    x -= 3;     // 12
    x *= 2;     // 24
    x /= 5;     // 4
    x %= 3;     // 1
    x++;        // 2
    x--;        // 1
    trace(x);
    return 0;
}
```

- অ্যাসাইনমেন্ট একটি স্টেটমেন্ট। `a = b = 0;` বা `if (x = 1)` লেখা যায় না।
- `++` ও `--` এমন স্টেটমেন্ট যা ভেরিয়েবলের পরে বসে (`i++;`)। বড় কোনো এক্সপ্রেশনের
  ভেতরে এগুলো ব্যবহার করা যায় না।
- ফাংশনের প্যারামিটার শুধু পড়া যায় (read-only)। বদলাতে চাইলে আগে সেটি একটি লোকাল
  ভেরিয়েবলে কপি করুন।

## স্কোপ

একটি ভেরিয়েবল তার ঘোষণা থেকে শুরু করে যে ব্লকের ভেতরে আছে তার শেষ পর্যন্ত থাকে।
`for`-এর হেডারে ঘোষিত ভেরিয়েবল শুধু সেই লুপের ভেতরেই থাকে।

```actionscript
package main;

protected int main() {
    if (true) {
        int inner = 1;
        trace(inner);
    }
    // 'inner' is not visible here

    for (int i = 0; i < 2; i++) {
        trace(i);
    }
    for (int i = 0; i < 2; i++) {   // fine: the first 'i' is gone
        trace(i);
    }
    return 0;
}
```

**নেস্টেড ব্লকের** একটি ভেরিয়েবল বাইরের কোনো ব্লকের ভেরিয়েবল বা প্যারামিটারের নাম আবার
ব্যবহার করতে পারে। একে **শ্যাডোইং** বলে: ব্লকের ভেতরে নামটি নতুন ভেরিয়েবলকে বোঝায়, আর
ব্লক শেষ হলে বাইরেরটি ফিরে আসে। কম্পাইলার এ নিয়ে সতর্ক করে (`ASX-WARN-009`), কারণ
ভুল করে এটা করে ফেলা সহজ:

```actionscript
package main;

protected int main() {
    int x = 1;
    if (true) {
        int x = x + 10;   // warning ASX-WARN-009; the initializer still reads the outer x
        trace(x);         // 11
    }
    trace(x);             // 1
    return 0;
}
```

**একই ব্লকে** কোনো নাম আবার ব্যবহার করা এখনো একটি এরর, এবং কোনো ফাংশন, স্ট্রাক্ট,
অথবা ফাংশনের নিজস্ব সবচেয়ে বাইরের ব্লকে কোনো প্যারামিটারের নাম আবার ব্যবহার করাও:

```actionscript
package main;

protected int main() {
    int x = 1;
    int x = 2; // error: ASX-NAME-002
    return 0;
}
```

## সাইনড ও আনসাইনড কখনো নিঃশব্দে মেশে না

RuneScript কোনো সাইনড পূর্ণসংখ্যাকে নিঃশব্দে আনসাইনডে পরিণত করে না। কম্পাইলার মানটির
*টাইপ* যাচাই করে, মানটি নয়। একটি সাইনড ভেরিয়েবলে ঋণাত্মক সংখ্যা থাকতে পারে, তাই সেটি
`uint`-এ রাখা এরর, এমনকি সেই মুহূর্তে মানটি ধনাত্মক হলেও:

```actionscript
package main;

protected int main() {
    sint signedValue = -10;
    uint unsignedValue = signedValue; // error: ASX-TYPE-004
    return 0;
}
```

```text
main.rsc:5:26: error ASX-TYPE-004: Cannot assign signed integer 'sint' to unsigned integer 'uint'.
        uint unsignedValue = signedValue;
                             ^^^^^^^^^^^
  help: Use an explicit checked conversion, (uint) x, or change the destination type.
```

সংখ্যার লিটারাল বেশি নমনীয়। একটি লিটারাল যে তাকে গ্রহণ করে তার টাইপ নেয়, যদি মানটি
সেখানে আঁটে:

```actionscript
package main;

protected int main() {
    uint a = 42;            // fine: 42 fits in uint
    uint b = 6 * 7;         // fine: constants are computed at compile time
    int c = -2147483648;    // fine: the smallest int
    trace(a, b, c);
    return 0;
}
```

```actionscript
package main;

protected int main() {
    uint u = -1;            // error: ASX-TYPE-007
    return 0;
}
```

```actionscript
package main;

protected int main() {
    int big = 3000000000;   // error: ASX-TYPE-006
    return 0;
}
```

উল্টো দিকটিও (`uint` থেকে `int`) প্রত্যাখ্যাত, যেমন একটি অপারেশনে `int` ও `uint`
মেশানোও। রূপান্তর করতে হলে তা স্পষ্ট করে লিখুন: `(uint) signedValue` চলার সময় মানটি
যাচাই করে, আর ঋণাত্মক হলে একটি এরর ছোড়ে (throw করে)। দেখুন
[ফ্লোট ও রূপান্তর](13-floats-and-conversions.md), যেখানে `float`, `float32` ও
`float64`-ও আলোচিত।

পরবর্তী: [অপারেটর](04-operators.md)
