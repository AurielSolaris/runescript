<!-- source: docs/guide/20-files-and-input.md @ 4541765 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/20-files-and-input.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ২০. ফাইল, ইনপুট ও আউটপুট

চারটি স্ট্যান্ডার্ড প্যাকেজ একটি প্রোগ্রামকে তার চারপাশের জগতের সঙ্গে যুক্ত করে:

| প্যাকেজ | কাজ |
|---|---|
| `std:io` | নতুন লাইন ছাড়া প্রিন্ট, এরর ও সতর্কবার্তার লাইন, এবং স্ট্যান্ডার্ড ইনপুট পড়া |
| `std:process` | কমান্ড-লাইন আর্গুমেন্ট, এনভায়রনমেন্ট ভেরিয়েবল, ওয়ার্কিং ডিরেক্টরি ও প্ল্যাটফর্ম |
| `std:fs` | ফাইল পড়া ও লেখা, এবং ডিরেক্টরি তৈরি, তালিকা করা ও মুছে ফেলা |
| `std:path` | পাথ জোড়া, ভাগ করা ও পূর্ণ (resolve) করা, Node.js-এর `path` মডিউলের নিয়মে |

এগুলো ইন্টারপ্রেটারে ও JavaScript-এ একইভাবে কাজ করে
([অধ্যায় ১১](11-javascript.md)), এমনকি এদের এরর বার্তাও এক।

## আউটপুট

`trace` আগের মতোই একটি লাইন প্রিন্ট করে। `std:io` লেখার আরও তিনটি উপায় যোগ করে:

| ফাংশন | যা লেখে |
|---|---|
| `io.print(string s)` | `s` স্ট্যান্ডার্ড আউটপুটে, নতুন লাইন ছাড়া |
| `io.error(string s)` | `s` ও একটি নতুন লাইন স্ট্যান্ডার্ড এররে, টার্মিনালে লাল রঙে |
| `io.warn(string s)` | `s` ও একটি নতুন লাইন স্ট্যান্ডার্ড এররে, টার্মিনালে হলুদ রঙে |

```actionscript
package main;

use "std:io";

protected int main() {
    io.print("Loading");
    for (int i = 0; i < 3; i++) {
        io.print(".");
    }
    trace(" done");                     // Loading... done
    io.warn("the cache is empty");      // on standard error
    io.error("could not reach the server");
    return 1;
}
```

`trace` ও `io.print` একই স্ট্যান্ডার্ড আউটপুট ব্যবহার করে, তাই তাদের লেখা যে ক্রমে লেখা
হয়েছে সেই ক্রমেই বের হয়। রং কেবল তখনই ব্যবহার হয় যখন স্ট্যান্ডার্ড এরর একটি টার্মিনাল
এবং `NO_COLOR` এনভায়রনমেন্ট ভেরিয়েবল সেট করা নেই; `runescript run --color=never`-ও রং
বন্ধ করে।

## ইনপুট

| ফাংশন | যা ফেরত দেয় |
|---|---|
| `io.readLine()` | স্ট্যান্ডার্ড ইনপুটের পরের লাইন, এর `"\n"` বা `"\r\n"` ছাড়া। ইনপুটের শেষে যা বাকি থাকে তা ফেরত দেয়, আর সব পড়া হয়ে গেলে সেটি `""` |
| `io.readByte()` | পরের বাইট (0 থেকে 255), অথবা ইনপুটের শেষে `-1` |
| `io.readAll()` | ইনপুটের বাকি সবটুকু |
| `io.atEnd()` | ইনপুট শেষ হয়েছে কি না। এখনো কোনো ইনপুট না থাকলে এটি অপেক্ষা করে |

লেখা UTF-8 হিসেবে পড়া হয়; যে বাইটগুলো বৈধ UTF-8 নয় সেগুলো U+FFFD হয়ে যায়।
`io.readByte` কাঁচা বাইটগুলো দেয়।

```actionscript
package main;

use "std:io";
use "std:string";

// Numbers each line of its input:  runescript run number.rsc < notes.txt
protected int main() {
    int n = 0;
    while (!io.atEnd()) {
        n++;
        string line = io.readLine();
        trace(string.repeat(" ", 4 - string.len((string) n)) + (string) n, line);
    }
    return 0;
}
```

## আর্গুমেন্ট ও এনভায়রনমেন্ট

| ফাংশন | যা ফেরত দেয় |
|---|---|
| `process.args()` | কমান্ড-লাইন আর্গুমেন্টগুলো, একটি নতুন `Array<string>` হিসেবে |
| `process.env(string name)` | একটি এনভায়রনমেন্ট ভেরিয়েবল, অথবা সেট করা না থাকলে `""` |
| `process.hasEnv(string name)` | একটি এনভায়রনমেন্ট ভেরিয়েবল সেট করা আছে কি না |
| `process.cwd()` | ওয়ার্কিং ডিরেক্টরি |
| `process.platform()` | `"windows"`, `"linux"`, `"macos"`, অথবা অন্য কোনো সিস্টেমের নাম; ওয়েব পেজে `"browser"` |

আর্গুমেন্ট হলো প্রোগ্রামের ফাইলের পরের শব্দগুলো: `runescript run greet.rsc Ada Grace`
(অথবা `node greet.js Ada Grace`) পাঠায় `["Ada", "Grace"]`। `runescript`-এর নিজের
ফ্ল্যাগগুলো ফাইলের আগে বসে।

```actionscript
package main;

use "std:process";

protected int main() {
    Array<string> names = process.args();
    if (names.length == 0) {
        trace("usage: greet <name>...");
        return 2;
    }
    for (int i = 0; i < names.length; i++) {
        trace($"Hello, {names[i]}!");
    }
    return 0;
}
```

## ফাইল ও ডিরেক্টরি

পাথগুলো ওয়ার্কিং ডিরেক্টরির সাপেক্ষে। `readFile`, `writeFile` ও `appendFile` UTF-8
লেখা নিয়ে কাজ করে; অন্য ডেটার জন্য নিচের [বাইনারি ফাইল](#বাইনারি-ফাইল) দেখুন।

| ফাংশন | কাজ |
|---|---|
| `fs.readFile(string path)` | ফাইলের লেখা ফেরত দেয় |
| `fs.writeFile(string path, string text)` | একটি ফাইল তৈরি করে বা বদলে দেয় |
| `fs.appendFile(string path, string text)` | ফাইলের শেষে যোগ করে, দরকার হলে ফাইলটি তৈরি করে |
| `fs.exists(string path)`, `fs.isFile(…)`, `fs.isDir(…)` | কিছু, একটি ফাইল, বা একটি ডিরেক্টরি সেখানে আছে কি না |
| `fs.size(string path)` | একটি ফাইলের আকার বাইটে, একটি `int64` হিসেবে |
| `fs.copyFile(string from, string to)` | একটি ফাইল কপি করে, `to`-কে বদলে দিয়ে |
| `fs.rename(string from, string to)` | একটি ফাইল বা ডিরেক্টরি সরায় বা নাম বদলায় |
| `fs.remove(string path)` | একটি ফাইল বা খালি ডিরেক্টরি মুছে ফেলে |
| `fs.removeAll(string path)` | একটি ডিরেক্টরি ও এর ভেতরের সবকিছু মুছে ফেলে; না থাকলে কিছুই করে না |
| `fs.mkdir(string path)` | একটি ডিরেক্টরি তৈরি করে, যার প্যারেন্ট অবশ্যই থাকতে হবে |
| `fs.mkdirAll(string path)` | একটি ডিরেক্টরি ও অনুপস্থিত প্যারেন্টগুলো তৈরি করে; আগে থেকে থাকলে কিছুই করে না |
| `fs.readDir(string path)` | একটি ডিরেক্টরির নামগুলো, সাজানো অবস্থায় |
| `fs.walk(string path)` | `path`-এর নিচের প্রতিটি ফাইল ও ডিরেক্টরি, এর সাপেক্ষে পাথ হিসেবে; প্রতিটি ডিরেক্টরির পরে তার ভেতরের জিনিসগুলো |
| `fs.tempDir()` | অস্থায়ী ফাইলের জন্য সিস্টেমের ডিরেক্টরি |

ব্যর্থ হলে একটি `Error` থ্রো হয়, যার বার্তা বলে কী ভুল হয়েছে, সব জায়গায় একই শব্দে:
`fs.readFile: cannot read notes.txt: no such file or directory`। কারণগুলো হলো
"no such file or directory", "already exists", "is a directory", "not a directory",
"directory not empty" ও "permission denied"।

```actionscript
package main;

use "std:fs";
use "std:path";

protected int main() {
    string dir = path.join(fs.tempDir(), "runescript-guide");
    fs.removeAll(dir);
    fs.mkdirAll(path.join(dir, "notes"));
    fs.writeFile(path.join(dir, "notes/todo.txt"), "buy milk\n");
    fs.appendFile(path.join(dir, "notes/todo.txt"), "write docs\n");
    trace(fs.readFile(path.join(dir, "notes/todo.txt")));
    trace(fs.size(path.join(dir, "notes/todo.txt")));       // 20
    Array<string> all = fs.walk(dir);
    trace(all.length);                                     // 2: notes, and the file in it
    try {
        trace(fs.readFile(path.join(dir, "missing.txt")));
    } catch (Error e) {
        trace(e.name);                                     // Error
    }
    fs.removeAll(dir);
    trace(fs.exists(dir));                                 // false
    return 0;
}
```

### বাইনারি ফাইল

ছবি, আর্কাইভ এবং লেখা নয় এমন অন্য ডেটা বাইট হিসেবে, একটি `Array<uint8>` হিসেবে,
পড়া ও লেখা হয়। বাইটগুলো কোনো বদল ছাড়াই ফাইলে যায় ও ফাইল থেকে আসে।

| ফাংশন | কী করে |
|---|---|
| `fs.readBytes(string path)` | ফাইলের বাইটগুলো ফেরত দেয়, একটি `Array<uint8>` |
| `fs.writeBytes(string path, Array<uint8> bytes)` | একটি ফাইল তৈরি করে বা বদলে দেয় |
| `fs.appendBytes(string path, Array<uint8> bytes)` | ফাইলের শেষে যোগ করে, দরকার হলে ফাইলটি তৈরি করে |

এগুলো তাদের লেখার সংস্করণের মতোই, একই বার্তা দিয়ে, ব্যর্থ হয়। `string.bytes(s)` ও
`string.fromBytes(b)` ([অধ্যায় ১০](10-std-string.md#বাইট)) লেখা ও তার UTF-8 বাইটের
মধ্যে রূপান্তর করে।

```actionscript
package main;

use "std:fs";
use "std:path";
use "std:string";

protected int main() {
    string file = path.join(fs.tempDir(), "runescript-guide.bin");
    Array<uint8> header = [0x89, 0x50, 0x4E, 0x47];   // the start of a PNG file
    fs.writeBytes(file, header);
    fs.appendBytes(file, string.bytes("hi"));
    Array<uint8> data = fs.readBytes(file);
    trace(data.length, data[0], data[1]);             // 6 137 80
    trace(fs.size(file));                             // 6
    bool png = data.length >= 4 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47;
    trace(png);                                       // true
    trace(string.fromBytes(data.slice(4, 6)));        // hi
    fs.remove(file);
    return 0;
}
```

## পাথ

`std:path`-এ Node.js-এর `path` মডিউলের ফাংশনগুলো আছে, RuneScript-এর জন্য লেখা (এগুলো
Node.js-কে ডাকে না, তাই ইন্টারপ্রেটারেও কাজ করে):

| ফাংশন | যা ফেরত দেয় |
|---|---|
| `path.join(a, b)`, `path.joinAll(parts)` | বিভাজক দিয়ে জোড়া অংশগুলো, তারপর স্বাভাবিক (normalize) করা |
| `path.resolve(p)`, `path.resolveAll(parts)` | ওয়ার্কিং ডিরেক্টরি থেকে শুরু করে একটি পূর্ণ (absolute) পাথ |
| `path.normalize(p)` | `.`, `..` বা পুনরাবৃত্ত বিভাজক ছাড়া `p` |
| `path.relative(from, to)` | `from` থেকে `to`-তে যাওয়ার পাথ |
| `path.dirname(p)`, `path.basename(p)`, `path.basename(p, suffix)`, `path.extname(p)` | ডিরেক্টরি, শেষ অংশ (চাইলে `suffix` ছাড়া), ও এক্সটেনশন |
| `path.isAbsolute(p)`, `path.root(p)` | `p` পূর্ণ পাথ কি না, এবং এর রুট (`"/"`, `"C:\"`, …) |
| `path.sep()`, `path.delimiter()` | `"/"` ও `":"`, অথবা Windows-এ `"\"` ও `";"` |

এগুলো প্রোগ্রাম যে সিস্টেমে চলে তার নিয়ম মানে: Windows-এ Windows পাথ (`C:\`,
বিভাজক হিসেবে `\` ও `/` দুটোই, `\\server\share`), অন্য সব জায়গায় POSIX পাথ। যে পাথ
সব জায়গায় একরকম দেখাতে হবে, তার জন্য প্রতিটি ফাংশন `path.posixJoin`, `path.posixResolve`,
… এবং `path.win32Join`, `path.win32Resolve`, … হিসেবেও আছে:

```actionscript
package main;

use "std:path";

protected int main() {
    trace(path.posixJoin("/home/ada", "../grace/notes.txt"));   // /home/grace/notes.txt
    trace(path.win32Join("C:\\Users", "ada\\..\\grace"));        // C:\Users\grace
    trace(path.posixBasename("/srv/www/index.html", ".html"));  // index
    trace(path.posixExtname("archive.tar.gz"));                 // .gz
    trace(path.posixRelative("/data/a/b", "/data/c"));          // ../../c
    trace(path.win32IsAbsolute("C:docs"));                      // false: relative to C:'s current directory
    return 0;
}
```

ফলগুলো Node.js-এর `path` মডিউলের সঙ্গে মেলে (টেস্টগুলো হাজার হাজার এলোমেলো পাথ
তুলনা করে), দুটি ব্যতিক্রম ছাড়া: Windows পাথ কেবল ASCII-এর বড়-ছোট হাতের অক্ষর উপেক্ষা
করে তুলনা করা হয়, আর অন্য ড্রাইভের পাথ পূর্ণ করার সময় সেই ড্রাইভের রুট থেকে শুরু হয়।

## উদাহরণ: একটি Brainfuck ইন্টারপ্রেটার

[`examples/brainfuck.rsc`](../../../../examples/brainfuck.rsc)
[Brainfuck](https://en.wikipedia.org/wiki/Brainfuck) প্রোগ্রাম চালায় ৩২,৭৬৮ ঘরের একটি
টেপের ওপর, চারটি প্যাকেজই ব্যবহার করে:

```text
$ runescript run examples/brainfuck.rsc examples/bf/hello.bf
Hello World!
$ echo "Hello, World!" | runescript run examples/brainfuck.rsc examples/bf/rot13.bf
Uryyb, Jbeyq!
$ runescript run examples/brainfuck.rsc -e "+[>+]"
-e:1:3: the data pointer moved off the tape (to cell 32768; the cells are 0 to 32767)
```

এটি `process.args()` দেখে প্রোগ্রামটি `fs.readFile` দিয়ে পড়ে, অথবা `-e` থেকে নেয়।
এটি প্রথমে প্রোগ্রামটিকে স্ট্রাক্টের একটি তালিকায় কম্পাইল করে: `+` ও `-`-এর টানা সারিকে
একটি ধাপে গুটিয়ে আনে এবং প্রতিটি বন্ধনীর জোড়া মেলায়, তাই কিছু চালানোর আগেই একটি
জোড়াহীন বন্ধনী তার লাইন ও কলামসহ জানানো হয়। টেপটি একটি `Array<uint8>`, যার ঘরগুলো
`% 256` দিয়ে 255 থেকে 0-তে ঘুরে আসে। `,` `io.readByte()` দিয়ে পড়ে এবং ইনপুটের শেষে
ঘরটি অপরিবর্তিত রাখে। `.` বাইটগুলো জমা করে এবং `io.print` দিয়ে প্রিন্ট করার আগে
RuneScript-এর ভেতরেই সেগুলোকে UTF-8 হিসেবে ডিকোড করে, তাই একটি Brainfuck প্রোগ্রাম যেকোনো
অক্ষর প্রিন্ট করতে পারে। এররগুলো যায় `io.error`-এ, আর `main` এক্সিট স্ট্যাটাস ফেরত দেয়।

পরবর্তী: [JSON: std:json](21-json.md)
