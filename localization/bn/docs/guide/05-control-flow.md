<!-- source: docs/guide/05-control-flow.md @ 52427d9 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/05-control-flow.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ৫. কন্ট্রোল ফ্লো

শর্তগুলো অবশ্যই `bool` হতে হবে। `if (n)` নয়, লিখুন `if (n != 0)`। বডিতে সবসময়
বন্ধনী ব্যবহার করা হয়।

## if / else if / else

```actionscript
package main;

public string classify(int n) {
    if (n < 0) {
        return "negative";
    } else if (n == 0) {
        return "zero";
    } else {
        return "positive";
    }
}

protected int main() {
    trace(classify(-5), classify(0), classify(9));   // negative zero positive
    return 0;
}
```

## while

```actionscript
package main;

protected int main() {
    int n = 3;
    while (n > 0) {
        trace(n);
        n--;
    }
    trace("liftoff");
    return 0;
}
```

## for

`for (init; condition; step) { … }`

```actionscript
package main;

public void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}

protected int main() {
    sayHi("Hello world", 3);
    return 0;
}
```

তিনটি অংশের প্রতিটিই ঐচ্ছিক। কোনো শর্ত না থাকলে লুপটি `break` বা `return` পর্যন্ত
চলতে থাকে:

```actionscript
package main;

protected int main() {
    int tries = 0;
    for (;;) {
        tries++;
        if (tries == 3) {
            break;
        }
    }
    trace("stopped after", tries);
    return 0;
}
```

ইনিশিয়ালাইজার একটি ভেরিয়েবল ঘোষণা করতে পারে (যা কেবল লুপের ভেতরে থাকে), বিদ্যমান
একটিতে মান বসাতে পারে, অথবা একটি ফাংশন ডাকতে পারে। স্টেপ হতে পারে একটি অ্যাসাইনমেন্ট,
`++`/`--`, অথবা একটি কল।

## break ও continue

`break` সবচেয়ে ভেতরের লুপ থেকে বেরিয়ে যায়। `continue` পরের পুনরাবৃত্তিতে চলে যায়।
`for` লুপে `continue`-এর পরেও স্টেপটি চলে।

```actionscript
package main;

protected int main() {
    for (int i = 0; i < 10; i++) {
        if (i % 2 == 0) {
            continue;       // skip even numbers
        }
        if (i > 7) {
            break;          // stop at 9
        }
        trace(i);           // 1 3 5 7
    }
    return 0;
}
```

লুপের বাইরে `break` বা `continue` একটি এরর (`ASX-FLOW-001`)।

## প্রতিটি পথে রিটার্ন

যে ফাংশন একটি মান ফেরত দেয়, তাকে প্রতিটি পথেই ফেরত দিতে হবে। কম্পাইলার এমন বডি গ্রহণ
করে যা `return` দিয়ে শেষ হয় (অথবা `throw` দিয়ে, দেখুন [এরর](14-errors.md)), অথবা
এমন একটি `if`/`else` শৃঙ্খল দিয়ে, যার প্রতিটি শাখা রিটার্ন করে:

```actionscript
package main;

public int sign(int n) {
    if (n < 0) {
        return -1;
    } else if (n == 0) {
        return 0;
    } else {
        return 1;
    }
}

protected int main() {
    trace(sign(-8), sign(0), sign(8));
    return 0;
}
```

কম্পাইলার কোনো লুপকে রিটার্নের নিশ্চয়তা হিসেবে ধরে না। লুপের পরে একটি শেষ `return`
যোগ করুন:

```actionscript
package main;

public int firstSquareAbove(int limit) {
    for (int i = 0; i < 1000; i++) {
        if (i * i > limit) {
            return i;
        }
    }
    return -1;
}

protected int main() {
    trace(firstSquareAbove(50));   // 8
    return 0;
}
```

পরবর্তী: [ফাংশন](06-functions.md)
