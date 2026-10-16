<!-- source: docs/guide/06-functions.md @ 52427d9 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/06-functions.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ৬. ফাংশন

## ফাংশন ঘোষণা

প্রতিটি ফাংশন ঘোষণার আকার

```text
<access> <return type> <name>(<parameters>) { <body> }
```

যেখানে `<access>` হলো `public`, `private` অথবা `protected`। এটি **বাধ্যতামূলক**,
রিটার্ন টাইপও তাই (কোনো ফল না থাকলে `void`)। কোনো `function` কিওয়ার্ড নেই।

```actionscript
package main;

public int add(int a, int b) {
    return a + b;
}

private void greet(string name) {
    trace("hello", name);
}

public void main() {
    greet("rune");
    trace(add(2, 3));
}
```

| অ্যাক্সেস | অর্থ |
|---|---|
| `public` | সব জায়গা থেকে ব্যবহারযোগ্য |
| `private` | শুধু একই ফাইলের ভেতরে ব্যবহারযোগ্য |
| `protected` | শুধু একই প্যাকেজের ভেতরে ব্যবহারযোগ্য |

> একটি প্রোগ্রামে একাধিক ফাইল থাকলে এই নিয়মগুলো প্রয়োগ করা হয়; দেখুন
> [প্যাকেজ ও ফাইল](12-packages-and-files.md)।

`return;` একটি `void` ফাংশন থেকে আগেভাগে বেরিয়ে যায়।

## কল করা

আর্গুমেন্টগুলো প্যারামিটারের টাইপের সঙ্গে অ্যাসাইনমেন্টের একই নিয়মে যাচাই করা হয়।
যেমন, একটি `int` ভেরিয়েবল `uint` প্যারামিটারে পাঠানো যায় না। টপ-লেভেল ফাংশনগুলো তাদের
সংজ্ঞার আগেই ডাকা যায়:

```actionscript
package main;

public void main() {
    trace(square(12));     // defined below: fine
}

public int square(int n) {
    return n * n;
}
```

যে কলের ফল আপনার দরকার নেই, তা একটি স্টেটমেন্ট হিসেবে ব্যবহার করা যায়:

```actionscript
package main;

public int logAndCount(string message) {
    trace(message);
    return 1;
}

public void main() {
    logAndCount("result ignored");
}
```

## প্যারামিটার হলো কপি, এবং শুধু-পাঠ্য

আর্গুমেন্ট মান হিসেবে (by value) পাঠানো হয়। ফাংশনের ভেতরে প্যারামিটারে নতুন মান বসানো
যায় না (`ASX-NAME-006`)। বদলাতে হলে মানটি একটি লোকাল ভেরিয়েবলে কপি করুন:

```actionscript
package main;

public uint gcd(uint a, uint b) {
    uint x = a;
    uint y = b;
    while (y != 0) {
        uint t = x % y;
        x = y;
        y = t;
    }
    return x;
}

public void main() {
    trace(gcd(84, 36));   // 12
}
```

## রিকার্শন

```actionscript
package main;

public int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

public void main() {
    trace(factorial(10));   // 3628800
}
```

কল ১০,০০০ স্তর পর্যন্ত গভীরে যেতে পারে। এর চেয়ে গভীরে গেলে প্রোগ্রাম ক্র্যাশ না করে
`ASX-RUN-001` দিয়ে থেমে যায়।

## প্রোটোটাইপ (আগাম ঘোষণা)

একটি **প্রোটোটাইপ** বডি ছাড়াই একটি ফাংশনের স্বাক্ষর (signature) ঘোষণা করে। এটি `;`
দিয়ে শেষ হয়, আর প্যারামিটারের নাম ঐচ্ছিক। সংজ্ঞাটি পরে একই স্কোপে, হুবহু একই টাইপসহ
আসতে হবে:

```actionscript
package main;

public int twice(int);          // prototype

public void main() {
    trace(twice(21));           // 42
}

public int twice(int x) {       // definition
    return x * 2;
}
```

টপ-লেভেল ফাংশন তো এমনিতেই সব জায়গা থেকে দেখা যায়, তাই সেখানে প্রোটোটাইপ ঐচ্ছিক।
ফাংশনের ভেতরে এগুলো দিয়েই দুটি নেস্টেড ফাংশন একে অপরকে ডাকে (নিচে দেখুন)। একটি
প্রোটোটাইপ দুবার, বা সংজ্ঞার পরে ঘোষণা করা এরর (`ASX-NAME-002`)। ভিন্ন টাইপের
সংজ্ঞা হলে `ASX-TYPE-015`। যে প্রোটোটাইপের কখনো সংজ্ঞা দেওয়া হয় না, তা
`ASX-NAME-010`।

## নেস্টেড ফাংশন

ফাংশনের ভেতরে ফাংশন ঘোষণা করা যায়, অন্য যেকোনো ফাংশনের মতোই একটি অ্যাক্সেস মডিফায়ার
দিয়ে। একটি নেস্টেড ফাংশন তার ঘোষণা থেকে বাইরের ব্লকের শেষ পর্যন্ত, এবং নিজের বডির
ভেতরেও দৃশ্যমান (তাই এটি রিকার্সিভ হতে পারে):

```actionscript
package main;

public void main() {
    private int fib(int n) {
        if (n < 2) {
            return n;
        }
        return fib(n - 1) + fib(n - 2);
    }
    trace(fib(15));   // 610
}
```

দুটি নেস্টেড ফাংশনকে একে অপরকে ডাকাতে, আগে একটিকে প্রোটোটাইপ দিয়ে ঘোষণা করুন:

```actionscript
package main;

public void main() {
    private bool isEven(int n);

    private bool isOdd(int n) {
        if (n == 0) {
            return false;
        }
        return isEven(n - 1);
    }

    private bool isEven(int n) {
        if (n == 0) {
            return true;
        }
        return isOdd(n - 1);
    }

    trace(isEven(10), isOdd(7));   // true true
}
```

## ফাংশনের সীমানা

নেস্টেড ফাংশন ও ল্যাম্বডা সহ প্রতিটি ফাংশনের বডি একটি **সীমানা** (boundary)। এর
ভেতরের কোড ব্যবহার করতে পারে:

- নিজের প্যারামিটার ও লোকাল ভেরিয়েবল;
- ফাংশন: টপ-লেভেল ফাংশন, এবং যেখানে এটি ঘোষিত সেখানে দৃশ্যমান নেস্টেড ফাংশন (নিজেকেও)।

এটি যে ফাংশনের ভেতরে নেস্ট করা, তার লোকাল ভেরিয়েবল বা প্যারামিটার ব্যবহার করতে
**পারে না**। RuneScript-এর ফাংশন ভেরিয়েবল ধরে রাখে (capture) না (`ASX-NAME-009`)।
একমাত্র ব্যতিক্রম ওয়েব কম্পোনেন্টের রুন, যা নেস্টেড ফাংশন ও ল্যাম্বডা ব্যবহার করতে
পারে ([অধ্যায় ১৫](15-web-pages.md)):

```actionscript
package main;

public void main() {
    int count = 0;
    private void bump() {
        count++;      // error: ASX-NAME-009
    }
}
```

মানটি প্যারামিটার হিসেবে পাঠান, আর ফলটি ফেরত দিন:

```actionscript
package main;

public void main() {
    private int bump(int count) {
        return count + 1;
    }
    int count = 0;
    count = bump(count);
    trace(count);   // 1
}
```

সীমানার কারণে, একটি নেস্টেড ফাংশন বাইরের ফাংশনের কোনো নাম আবার ব্যবহার করতে পারে।
নেস্টেড ফাংশনের ভেতরে সেই নামটি শুধু তার নিজের ভেরিয়েবলকেই বোঝায়। `break` ও
`continue`-ও কোনো সীমানা পেরোতে পারে না।

## ফাংশন টাইপ

ফাংশন টাইপ C-ধাঁচে লেখা হয়: রিটার্ন টাইপ, তারপর বন্ধনীর ভেতরে প্যারামিটারের টাইপ।

| টাইপ | অর্থ |
|---|---|
| `int(int, int)` | দুটি `int` নেয়, একটি `int` ফেরত দেয় |
| `void(string)` | একটি `string` নেয়, কিছু ফেরত দেয় না |
| `bool()` | কিছু নেয় না, একটি `bool` ফেরত দেয় |
| `int(int)(string)` | একটি `string` নেয়, একটি `int(int)` ফাংশন ফেরত দেয় |

ফাংশন টাইপ ভেরিয়েবল, প্যারামিটার ও রিটার্ন টাইপ হিসেবে ব্যবহার করা যায়। দুটি ফাংশন
টাইপ একই, যখন তাদের প্যারামিটার ও রিটার্ন টাইপ একই (`int` ও `sint` একই টাইপ ধরা হয়)।

## মান হিসেবে ফাংশন (কলব্যাক)

একটি ফাংশনের নাম না ডেকে লিখলে একটি ফাংশন-মান পাওয়া যায়। এটি রাখা, পাঠানো, ফেরত
দেওয়া এবং ডাকা যায়:

```actionscript
package main;

public int apply(int(int, int) op, int a, int b) {
    return op(a, b);
}

public int multiply(int a, int b) {
    return a * b;
}

public void main() {
    int(int, int) f = multiply;
    trace(f(6, 7));                   // 42
    trace(apply(multiply, 3, 4));     // 12
}
```

ফাংশন-মান `trace` করা বা `==` দিয়ে তুলনা করা যায় না। বিল্ট-ইন `trace` কোনো মান নয়;
এর বদলে এটিকে একটি ল্যাম্বডায় মুড়ে নিন।

## ল্যাম্বডা

একটি **ল্যাম্বডা** হলো নামহীন ফাংশন: বন্ধনীর ভেতরে টাইপসহ প্যারামিটার, তারপর `=>`,
তারপর একটি এক্সপ্রেশন অথবা একটি ব্লক।

```actionscript
package main;

public void repeat(int times, void(int) body) {
    for (int i = 0; i < times; i++) {
        body(i);
    }
}

public void main() {
    int(int, int) add = (int a, int b) => a + b;
    trace(add(2, 3));                                 // 5

    repeat(2, (int i) => trace("tick", i));           // tick 0, tick 1

    int(int) abs = (int x) => {
        if (x < 0) {
            return -x;
        }
        return x;
    };
    trace(abs(-4));                                   // 4
}
```

- **রিটার্ন টাইপ।** যেখানে একটি ফাংশন টাইপ প্রত্যাশিত, সেখানে কোনো ল্যাম্বডা অ্যাসাইন,
  পাঠানো বা ফেরত দেওয়া হলে, এটি সেই টাইপের রিটার্ন টাইপ নেয়। অন্যথায় রিটার্ন টাইপ আসে
  বডি থেকে: এক্সপ্রেশনটির টাইপ, অথবা প্রথম `return`-এর টাইপ।
- ল্যাম্বডাও নেস্টেড ফাংশনের মতো একই সীমানার নিয়ম মানে: এগুলো বাইরের ফাংশনের
  ভেরিয়েবল ব্যবহার করতে পারে না।
- ল্যাম্বডায় কোনো অ্যাক্সেস মডিফায়ার লাগে না।

## ফাংশন ফেরত দেওয়া

```actionscript
package main;

public int(int) pick(string op) {
    private int double(int n) {
        return n * 2;
    }
    if (op == "double") {
        return double;
    }
    return (int n) => n * n;
}

public void main() {
    trace(pick("double")(5), pick("square")(5));   // 10 25
}
```

`pick("double")(5)` সেই ফাংশনটিকে ডাকে যা `pick` ফেরত দেয়। আগে ফাংশন-মানটি মূল্যায়ন
হয়, তারপর আর্গুমেন্টগুলো।

## এখনো নেই

ওভারলোডিং (একই নামের দুটি ফাংশন), ডিফল্ট আর্গুমেন্ট, পরিবর্তনশীল-সংখ্যক প্যারামিটার
(variadic), জেনেরিক ফাংশন, এবং ভেরিয়েবল ধরে রাখা ক্লোজার।

পরবর্তী: [`trace` দিয়ে আউটপুট](07-trace.md)
