<!-- source: docs/guide/03-types-and-variables.md @ 52427d9 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/03-types-and-variables.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ৩. টাইপ ও ভেরিয়েবল

## ভেরিয়েবল ঘোষণা

আগে টাইপ, তারপর নাম, তারপর প্রাথমিক মান লিখুন:

```actionscript
package main;

public void main() {
    int count = 10;
    uint total = 250;
    bool ready = true;
    string name = "Ada";
    trace(name, count, total, ready);   // Ada 10 250 true
}
```

প্রতিটি ভেরিয়েবল ঘোষণার সময়েই **অবশ্যই** একটি প্রাথমিক মান পাবে। v0.1-এ নিজে থেকে
কোনো ডিফল্ট মান বসে না।

## মৌলিক টাইপগুলো

| টাইপ | যা ধারণ করে | পরিসর |
|---|---|---|
| `int` | সাইনড (চিহ্নযুক্ত) পূর্ণসংখ্যা | −2147483648 … 2147483647 |
| `sint` | **হুবহু `int`-এর মতোই একই টাইপ।** "সাইনড" কথাটি স্পষ্ট করে বলতে চাইলে এটি ব্যবহার করুন। | `int`-এর মতোই |
| `uint` | আনসাইনড (চিহ্নহীন) পূর্ণসংখ্যা | 0 … 4294967295 |
| `float`, `float32`, `float64` | দশমিক (ফ্লোটিং-পয়েন্ট) সংখ্যা ([অধ্যায় ১৩](13-floats-and-conversions.md)) | |
| `bool` | `true` অথবা `false` | |
| `string` | লেখা, যেমন `"hello\n"` | |
| `Error` ও এর প্রকারভেদ | এরর ([অধ্যায় ১৪](14-errors.md)) | |
| `void` | কোনো মান নয়; শুধু ফাংশনের রিটার্ন টাইপ হিসেবে ব্যবহৃত | |

`int` ও `sint` অবাধে মেশানো যায়:

```actionscript
package main;

public void main() {
    sint a = -5;
    int b = a;      // fine: same type
    sint c = b;     // fine
    trace(a, b, c);
}
```

স্ট্রিংয়ের এস্কেপ: `\n` (নতুন লাইন), `\t` (ট্যাব), `\r`, `\\`, এবং `\"`।

## অ্যাসাইনমেন্ট

```actionscript
package main;

public void main() {
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

public void main() {
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
}
```

v0.1 **শ্যাডোইং** অনুমোদন করে না। নতুন ভেরিয়েবল কোনো দৃশ্যমান ভেরিয়েবল, প্যারামিটার
বা ফাংশনের নাম আবার ব্যবহার করতে পারে না:

```actionscript
package main;

public void main() {
    int x = 1;
    if (true) {
        int x = 2; // error: ASX-NAME-002
    }
}
```

## সাইনড ও আনসাইনড কখনো নিঃশব্দে মেশে না

RuneScript কোনো সাইনড পূর্ণসংখ্যাকে নিঃশব্দে আনসাইনডে পরিণত করে না। কম্পাইলার মানটির
*টাইপ* যাচাই করে, মানটি নয়। একটি সাইনড ভেরিয়েবলে ঋণাত্মক সংখ্যা থাকতে পারে, তাই সেটি
`uint`-এ রাখা এরর, এমনকি সেই মুহূর্তে মানটি ধনাত্মক হলেও:

```actionscript
package main;

public int main() {
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

public void main() {
    uint a = 42;            // fine: 42 fits in uint
    uint b = 6 * 7;         // fine: constants are computed at compile time
    int c = -2147483648;    // fine: the smallest int
    trace(a, b, c);
}
```

```actionscript
package main;

public void main() {
    uint u = -1;            // error: ASX-TYPE-007
}
```

```actionscript
package main;

public void main() {
    int big = 3000000000;   // error: ASX-TYPE-006
}
```

উল্টো দিকটিও (`uint` থেকে `int`) প্রত্যাখ্যাত, যেমন একটি অপারেশনে `int` ও `uint`
মেশানোও। রূপান্তর করতে হলে তা স্পষ্ট করে লিখুন: `(uint) signedValue` চলার সময় মানটি
যাচাই করে, আর ঋণাত্মক হলে একটি এরর ছোড়ে (throw করে)। দেখুন
[ফ্লোট ও রূপান্তর](13-floats-and-conversions.md), যেখানে `float`, `float32` ও
`float64`-ও আলোচিত।

পরবর্তী: [অপারেটর](04-operators.md)
