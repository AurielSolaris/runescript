<!-- source: docs/guide/19-structs.md @ 6704ceb -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/19-structs.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ১৯. স্ট্রাক্ট

একটি স্ট্রাক্ট কয়েকটি নামযুক্ত মানকে একটি নতুন টাইপের একটি মানে একসঙ্গে রাখে:

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

public void main() {
    Point p = Point{x: 3, y: 4};
    trace(p, p.x + p.y);   // Point{x: 3, y: 4} 7
    p.y = 10;
    trace(p);              // Point{x: 3, y: 10}
}
```

## একটি স্ট্রাক্ট ঘোষণা

একটি স্ট্রাক্ট ঘোষণা করা হয় ফাইলের শীর্ষ স্তরে, ফাংশনের মতো একটি অ্যাক্সেস মডিফায়ার
দিয়ে, আর টাইপসহ ফিল্ডের একটি তালিকা দিয়ে, প্রতিটি `;`-এ শেষ:

```actionscript
package main;

public struct Player {
    string name;
    int score;
    Array<string> badges;
}

private struct Settings {
    bool sound;
    float volume;
}

public void main() {
    Player p = Player{name: "Ada", score: 0, badges: []};
    Settings s = Settings{sound: true, volume: 0.5};
    trace(p.name, s.volume);   // Ada 0.5
}
```

- `public` স্ট্রাক্ট সবখানে ব্যবহার করা যায়, `protected` কেবল নিজের প্যাকেজে, আর `private`
  কেবল নিজের ফাইলে।
- একটি ফিল্ড যেকোনো মান-টাইপের হতে পারে: সংখ্যা, `bool`, `string`, `char`, অ্যারে, ফাংশন,
  এরর, এবং অন্য স্ট্রাক্ট।
- ফিল্ডের কোনো ডিফল্ট মান নেই। সে কারণেই প্রতিটি লিটারালকে সবগুলোর নাম দিতে হয়।

## একটি স্ট্রাক্ট তৈরি

একটি **স্ট্রাক্ট লিটারাল** স্ট্রাক্টের নাম দেয়, তারপর প্রতিটি ফিল্ডকে `name: value` দিয়ে
একটি মান দেয়:

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

public void main() {
    Point a = Point{x: 1, y: 2};
    Point b = Point{y: 2, x: 1};   // any order
    Point c = Point{x: 1};         // error: ASX-TYPE-026
}
```

একটি ফিল্ড বাদ দেওয়া, একটি দুবার দেওয়া, বা স্ট্রাক্টে নেই এমন একটির নাম দেওয়া হলো
`ASX-TYPE-026`। মানগুলো গণনা করা হয় আপনি যে ক্রমে লেখেন সেই ক্রমে।

## স্ট্রাক্ট হলো মান

একটি স্ট্রাক্ট অ্যাসাইন করা, ফাংশনে পাঠানো, ফেরত দেওয়া বা অ্যারেতে রাখা হলে সেটি **কপি**
হয়। কপি বদলালে মূলটি কখনো বদলায় না:

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

public Point moved(Point p, int dx) {
    Point q = p;
    q.x += dx;
    return q;
}

public void main() {
    Point a = Point{x: 1, y: 1};
    Point b = moved(a, 5);
    trace(a, b);   // Point{x: 1, y: 1} Point{x: 6, y: 1}
}
```

এটি অ্যারের উল্টো, যা ভাগাভাগি হয় ([অধ্যায় ১৬](16-text-and-arrays.md))। তবে একটি স্ট্রাক্টের
*ভেতরের* অ্যারে তবুও একটি অ্যারে: স্ট্রাক্ট কপি করলে অ্যারেটি ভাগাভাগি হয়।

একটি ফাংশনের প্যারামিটার শুধু-পড়ার জন্য, আর একটি স্ট্রাক্ট প্যারামিটারের ফিল্ডও তাই:
`moved`-এর ভেতরে `p.x = 1` হতো `ASX-NAME-006`। আগে একটি লোকাল ভেরিয়েবলে কপি করুন, যেমন
ওপরে।

## ফিল্ডের ভেতরে ফিল্ড, আর স্ট্রাক্টের অ্যারে

অ্যাসাইনমেন্ট ফিল্ড ও অ্যারের এলিমেন্টের ভেতর দিয়ে পৌঁছায়:

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

public struct Box {
    Point min;
    Point max;
}

public void main() {
    Box box = Box{min: Point{x: 0, y: 0}, max: Point{x: 2, y: 2}};
    box.max.x = 5;

    Array<Box> boxes = [box, box];
    boxes[1].min.y = -3;
    trace(boxes[0].min, boxes[1].min, box.max.x);   // Point{x: 0, y: 0} Point{x: 0, y: -3} 5
}
```

## তুলনা ও প্রিন্ট

`==` ও `!=` প্রতিটি ফিল্ড তুলনা করে। অ্যারে, ফাংশন বা এরর ফিল্ডওয়ালা স্ট্রাক্ট তুলনা করা যায়
না; বদলে যে ফিল্ডগুলো দরকার সেগুলো তুলনা করুন।

`trace` একটি স্ট্রাক্টকে তার টাইপ ও ফিল্ডসহ দেখায়, স্ট্রিং ও char উদ্ধৃতিচিহ্নে:
`Player{name: "Ada", score: 3}`।

## অন্য প্যাকেজের স্ট্রাক্ট

অন্য প্যাকেজের স্ট্রাক্টের নাম দিন তার কোয়ালিফায়ার দিয়ে
([অধ্যায় ১২](12-packages-and-files.md)): `geom.Point p = geom.Point{x: 1, y: 2};`।
স্ট্রাক্টটি `public` হতে হবে।

## "কোনো মান নেই"-এর জন্য একটি স্ট্রাক্ট

একটি মান আর তার "এটি কি আছে?" ফ্ল্যাগকে একসঙ্গে রাখার সবচেয়ে পরিষ্কার উপায় হলো একটি
স্ট্রাক্ট, [অধ্যায় ১৭](17-no-null-any-generics.md)-এর `null`-এর জায়গায় `false` প্যাটার্ন:

```actionscript
package main;

use "std:string";

public struct MaybeInt {
    bool present;   // false means "no value"
    int value;
}

public MaybeInt parseAge(string text) {
    string digits = string.trimSpace(text);
    if (digits == "" || string.indexFunc(digits, (char c) => c < '0' || c > '9') >= 0) {
        return MaybeInt{present: false, value: 0};
    }
    return MaybeInt{present: true, value: (int) digits};
}

public void main() {
    MaybeInt age = parseAge(" 42 ");
    if (age.present) {
        trace("age", age.value);   // age 42
    }
    trace(parseAge("forty").present);   // false
}
```

## এখনো নেই

মেথড, কনস্ট্রাক্টর, প্রাইভেট ফিল্ড, এবং JavaScript-এর `extern` ফাংশনে স্ট্রাক্ট পাঠানো।

পরবর্তী: আবার [গাইডের সূচিপত্রে](README.md)
