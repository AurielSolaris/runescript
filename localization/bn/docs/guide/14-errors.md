<!-- source: docs/guide/14-errors.md @ 920cb24 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/14-errors.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ১৪. এরর: `throw`, `try`, `catch`, `finally`

এরর ActionScript 3-এর মতোই কাজ করে, তবে RuneScript-এর আগে-টাইপ সিনট্যাক্সে:
`catch (e:Error)`-এর বদলে `catch (Error e)`।

```actionscript
package main;

public int parsePort(string text) {
    int port = (int) text;   // throws a TypeError if text isn't a number
    if (port < 1 || port > 65535) {
        throw new RangeError("port out of range", 1);
    }
    return port;
}

public void main() {
    try {
        trace(parsePort("8080"));
        trace(parsePort("99999"));
    } catch (RangeError e) {
        trace("bad port:", e.message);
    } catch (Error e) {
        trace("not a port:", e);
    } finally {
        trace("done");
    }
}
```

এটি দেখায় `8080`, `bad port: port out of range`, এবং `done`।

## এরর-মান

এররের টাইপ চারটি। শেষ তিনটি `Error`-এরই প্রকারভেদ:

| টাইপ | যার জন্য ব্যবহৃত |
|---|---|
| `Error` | যেকোনো কিছু; মূল টাইপ |
| `RangeError` | সীমার বাইরের মান: ওভারফ্লো, শূন্য দিয়ে ভাগ, যে সংখ্যা আঁটে না |
| `TypeError` | ভুল আকারের মান, যেমন `(int) "abc"` |
| `ArgumentError` | ভুল আর্গুমেন্ট; আপনার নিজের ফাংশনের জন্য |

`new Error(message, errorID)` একটি এরর তৈরি করে। দুটি আর্গুমেন্টই ঐচ্ছিক। প্রতিটি
এররের তিনটি ফিল্ড আছে:

| ফিল্ড | টাইপ | মান |
|---|---|---|
| `e.message` | `string` | বার্তা |
| `e.name` | `string` | টাইপের নাম, যেমন `"RangeError"` |
| `e.errorID` | `int` | `new`-কে দেওয়া ID, অথবা রানটাইম এররের নম্বর |

এরর সাধারণ মান। আপনি এগুলো রাখতে, পাঠাতে, রিটার্ন করতে ও `trace` করতে পারেন:
`trace(e)` দেখায় `RangeError: message`। একটি `RangeError` একটি `Error` ভেরিয়েবলে
রাখা যায়। আপনি শুধু একটি এররই `throw` করতে পারেন, স্ট্রিং বা সংখ্যা নয়।

## `try`, `catch` ও `finally`

- catch ক্লজগুলো ক্রমানুসারে চেষ্টা করা হয়। প্রথম যে ক্লজের টাইপ এররটির আছে (সেই
  টাইপ, বা তার কোনো প্রকারভেদ), সেটি চলে। নির্দিষ্ট টাইপগুলো `Error`-এর আগে রাখুন,
  কারণ যে ক্লজ কখনো চলতে পারে না সেটি একটি এরর (`ASX-TYPE-022`)।
- কোনো ক্লজ না মিললে এররটি এগিয়ে যায় বাইরের `try`-তে এবং কলকারী ফাংশনগুলোতে।
- `finally` ব্লক সবসময় চলে: ব্লক শেষ হওয়ার পরে, `return`, `break` বা `continue`-এর
  পরে, এবং একটি এররের পরে, যা তারপর এগিয়ে যায়।
- `finally` ব্লক `return`, `break` বা `continue` করতে পারে, ActionScript-এর মতো।
  তখন `try` স্টেটমেন্টটি যেভাবেই শেষ হচ্ছিল তা বাতিল হয়ে এটি কার্যকর হয়, আর চলমান
  কোনো এরর ফেলে দেওয়া হয়।
- টাইপ ছাড়া `catch (e)` প্রতিটি এরর ধরে, `catch (Error e)`-এর মতো।
- ধরা এররের ভেরিয়েবলটি শুধু-পাঠ্য।

একটি `throw` একটি পথ শেষ করে `return`-এর মতোই, তাই এই ফাংশনের শেষে কোনো `return`
লাগে না:

```actionscript
package main;

public int sign(int n) {
    if (n > 0) {
        return 1;
    }
    if (n < 0) {
        return -1;
    }
    throw new ArgumentError("zero has no sign");
}

public void main() {
    trace(sign(-5));
}
```

## রানটাইম এরর ধরা যায়

[অধ্যায় ৮](08-diagnostics.md)-এর রানটাইম এররগুলো এমন এরর হিসেবে throw করা হয় যা
আপনি ধরতে পারেন:

| রানটাইম এরর | যে হিসেবে throw হয় |
|---|---|
| `ASX-RUN-002`: পূর্ণসংখ্যা ওভারফ্লো | `RangeError` |
| `ASX-RUN-003`: পূর্ণসংখ্যাকে শূন্য দিয়ে ভাগ | `RangeError` |
| `ASX-RUN-004`: স্ট্যান্ডার্ড লাইব্রেরির কোনো ফাংশন ব্যর্থ | `Error` |
| `ASX-RUN-006`: একটি রূপান্তর ব্যর্থ | `RangeError` বা `TypeError` |

```actionscript
package main;

public void main() {
    int big = 2147483647;
    try {
        trace(big + 1);
    } catch (RangeError e) {
        trace(e.errorID, e.message);
        // 2 integer overflow: 2147483647 + 1 = 2147483648 does not fit in 'int'
    }
}
```

একমাত্র ব্যতিক্রম `ASX-RUN-001`, কল-গভীরতার সীমা (সাধারণত নিয়ন্ত্রণহীন রিকার্শন)।
এটি ধরা যায় না, আর এর জন্য প্রোগ্রাম থামার সময় `finally` ব্লকগুলো চলে না।

## যে এরর কেউ ধরে না

যে এরর কেউ ধরে না, তা যেকোনো রানটাইম এররের মতো প্রোগ্রাম থামিয়ে দেয়:

- একটি রানটাইম এরর জানানো হয় তার নিজের কোড, বার্তা ও অবস্থানসহ, এমনকি সেটি ধরে
  আবার throw করা হলেও।
- `new` দিয়ে আপনার তৈরি করা একটি এরর `throw`-এর জায়গায় `ASX-RUN-005` হিসেবে
  জানানো হয়:

```actionscript
package main;

public void main() {
    throw new ArgumentError("missing input");   // runtime error: ASX-RUN-005
}
```

```text
main.rsc:4:5: error ASX-RUN-005: uncaught ArgumentError: missing input
```

এখনো নেই: স্ট্যাক ট্রেস, `toString()`-এর মতো এরর মেথড, এবং আপনার নিজের এরর টাইপ
(এগুলোর জন্য ক্লাস লাগে)।

পরবর্তী: [`rune:dom` দিয়ে ওয়েব পেজ](15-web-pages.md)
