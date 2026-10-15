<!-- source: docs/guide/README.md @ 52427d9 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/README.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# RuneScript গাইড (v0.3)

RuneScript-এর যেসব বৈশিষ্ট্য **আজ** আছে, তার একটি ব্যবহারিক পরিচয়। এই পাতাগুলোর
প্রতিটি সম্পূর্ণ প্রোগ্রাম টেস্ট-সংগ্রহ (`tests/docs_test.go`) কম্পাইল করে দেখে।
কোনোটি কম্পাইল হওয়া বন্ধ হলে বিল্ড ব্যর্থ হয়।

1. [শুরু করা](01-getting-started.md): CLI তৈরি করা, একটি প্রোগ্রাম চালানো
2. [প্রোগ্রামের গঠন](02-program-structure.md): প্যাকেজ, কমেন্ট, `main`
3. [টাইপ ও ভেরিয়েবল](03-types-and-variables.md): `int`, `uint`, `bool`, `string`, অ্যাসাইনমেন্ট
4. [অপারেটর](04-operators.md): গাণিতিক, তুলনা, লজিক
5. [কন্ট্রোল ফ্লো](05-control-flow.md): `if`, `while`, `for`, `break`, `continue`
6. [ফাংশন](06-functions.md): প্যারামিটার, রিটার্ন মান, রিকার্শন
7. [`trace` দিয়ে আউটপুট](../../../../docs/guide/07-trace.md)
8. [ডায়াগনস্টিক তালিকা](../../../../docs/guide/08-diagnostics.md): প্রতিটি এরর কোড
9. [ActionScript 3 থেকে আসছেন?](../../../../docs/guide/09-coming-from-actionscript.md)
10. [স্ট্যান্ডার্ড লাইব্রেরি: `std:string`](../../../../docs/guide/10-std-string.md): RuneScript-এ Go-র `strings`
11. [JavaScript-এ কম্পাইল](../../../../docs/guide/11-javascript.md): `runescript build`, এবং কঠোর টাইপিং কীভাবে বজায় থাকে
12. [প্যাকেজ ও ফাইল](../../../../docs/guide/12-packages-and-files.md): `use`, একাধিক ফাইলের প্রোগ্রাম, দৃশ্যমানতা
13. [ফ্লোট ও রূপান্তর](../../../../docs/guide/13-floats-and-conversions.md): `float`, `float32`, `float64`, `(int) x`
14. [এরর](../../../../docs/guide/14-errors.md): `throw`, `try`, `catch`, `finally`, এবং রানটাইম এরর ধরা

সূক্ষ্ম নিয়মের জন্য দেখুন [ভাষার স্পেসিফিকেশন](../../../../docs/LANGUAGE_SPEC.md) (ইংরেজিতে)।
সামনে কী আসছে তা জানতে দেখুন [রোডম্যাপ](../../../../docs/ROADMAP.md) (ইংরেজিতে)।

> **এখনো নেই:** ক্লাস, স্ট্রাক্ট, অ্যারে, `char`, এবং স্ট্রিং জোড়া লাগানো
> (concatenation)। এদের কিওয়ার্ড সংরক্ষিত, আর ব্যবহার করলে স্পষ্ট "এখনো তৈরি হয়নি"
> এরর দেখায়।
