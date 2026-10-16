<!-- source: docs/guide/01-getting-started.md @ 920cb24 -->

> এটি [ইংরেজি মূল লেখার](../../../../docs/guide/01-getting-started.md) অনুবাদ। পার্থক্য থাকলে ইংরেজি সংস্করণটিই সঠিক। অনুবাদটি AI-এর সাহায্যে করা এবং লেখক যাচাই করেছেন।

# ১. শুরু করা

## ইনস্টল

RuneScript সোর্স থেকে তৈরি করা হয়, তাই আপনার **Go 1.25 বা নতুনতর** লাগবে
([go.dev/dl](https://go.dev/dl/))। ইনস্টলার প্রথমে সেটিই যাচাই করে: Go না থাকলে,
খুব পুরনো হলে বা ভাঙা থাকলে, এটি একটি স্পষ্ট বার্তা দিয়ে থেমে যায়।

**Linux ও macOS:**

```sh
curl -fsSL https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.sh | sh
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.ps1 | iex
```

ইনস্টলার যা করে:

1. Go খোঁজে ও যাচাই করে (PATH-এ, Go-র সাধারণ ইনস্টল-স্থানে, অথবা `RUNESCRIPT_GO`-তে);
2. সেই Go দিয়ে `runescript` CLI তৈরি করে, যেটি একসঙ্গে কম্পাইলার ও রানটাইম
   (শুধু Go দিয়ে তৈরি: কোনো C কম্পাইলার লাগে না);
3. এটি `~/.runescript/bin`-এ ইনস্টল করে (Windows-এ:
   `%LOCALAPPDATA%\Programs\RuneScript\bin`), অ্যাডমিনিস্ট্রেটর অধিকার ছাড়াই;
4. ওই ডিরেক্টরিটি আপনার PATH-এ যোগ করে এবং `RUNESCRIPT_HOME` সেট করে: আপনার শেলের
   শুরুর ফাইলগুলোতে (`~/.profile`, `~/.bashrc`, `~/.zshrc`, fish), অথবা Windows-এ
   আপনার ব্যবহারকারী-পরিবেশে;
5. একটি ছোট প্রোগ্রাম চালিয়ে ফলাফল যাচাই করে।

এরপর একটি নতুন টার্মিনাল খুলুন, তারপর:

```text
$ runescript version
runescript 0.3.0-dev (commit 9609a35, go1.25.11 linux/amd64)
```

রিপোজিটরির একটি ক্লোন থেকে, সর্বশেষ সংস্করণের বদলে সেই সোর্স তৈরি করতে সেখানেই
ইনস্টলার চালান: `sh install.sh`, অথবা
`powershell -ExecutionPolicy Bypass -File install.ps1`।

| অপশন (`install.sh` / `install.ps1`) | কাজ |
|---|---|
| `--prefix DIR` / `-InstallDir DIR` | `DIR/bin`-এ ইনস্টল করে (`RUNESCRIPT_HOME` দিয়েও দেওয়া যায়) |
| `--version REF` / `-Version REF` | GitHub থেকে একটি ট্যাগ, ব্রাঞ্চ বা কমিট তৈরি করে (`RUNESCRIPT_VERSION` দিয়েও) |
| `--remote` / `-Remote` | ক্লোনের ভেতরে থেকেও GitHub থেকে তৈরি করে |
| `--no-modify-path` / `-NoModifyPath` | PATH ও শুরুর ফাইলগুলো অপরিবর্তিত রাখে |
| `--uninstall` / `-Uninstall` | RuneScript এবং এর PATH-এন্ট্রি সরিয়ে দেয় |

Node.js ঐচ্ছিক। শুধু `runescript build` যে JavaScript লেখে তা চালাতে এটি লাগে।

### ইনস্টলার ছাড়া

```sh
go build -o runescript ./cmd/runescript     # runescript.exe on Windows
go run ./cmd/runescript run examples/hello.rsc   # or run without installing
```

## আপনার প্রথম প্রোগ্রাম

এটি `hello.rsc` নামে সংরক্ষণ করুন:

```actionscript
package main;

protected int main() {
    trace("Hello, RuneScript!");
    return 0;
}
```

চালান:

```text
$ runescript run hello.rsc
Hello, RuneScript!
```

> RuneScript ফাইল `.rsc` এক্সটেনশন ব্যবহার করে, তবে CLI যেকোনো ফাইলের নাম গ্রহণ
> করে। ফাইলগুলো UTF-8-এ লেখা হয়। UTF-16 বা UTF-32 ফাইল (যেমন কিছু Windows এডিটর
> সংরক্ষণ করে) একটি নোট-সহ রূপান্তর করা হয়; UTF-8 নয় এমন অন্য যেকোনো ফাইল
> `ASX-LEX-007` দিয়ে প্রত্যাখ্যাত হয়।

## কমান্ড

| কমান্ড | কাজ |
|---|---|
| `runescript run <file>` | প্রোগ্রামটি যাচাই করে, তারপর চালায় |
| `runescript check <file>` | না চালিয়ে কম্পাইল-টাইম এরর জানায় |
| `runescript run --watch <file>` | প্রোগ্রামটি চালায়, তারপর এর কোনো ফাইল বদলালেই আবার চালায় (Ctrl+C থামায়) |
| `runescript build [-o out.js] <file>` | একটি JavaScript মডিউলে কম্পাইল করে ([অধ্যায় ১১](11-javascript.md)) |
| `runescript dev [--port 5173] [--open] <file>` | localhost-এ পরিবেশন করে এবং প্রতিটি পরিবর্তনে ব্রাউজার রিলোড করে ([অধ্যায় ১১](11-javascript.md)) |
| `runescript version` | সংস্করণ দেখায় |

`run`, `check`, `build` ও `dev` সতর্কবার্তার জন্য `-Werror` ও `--allow <codes>`-ও নেয়
(দেখুন [অধ্যায় ৮](08-diagnostics.md))।

## এক্সিট স্ট্যাটাস

| অবস্থা | এক্সিট স্ট্যাটাস |
|---|---|
| প্রোগ্রাম চলেছে | `main` যে `int` ফেরত দেয় (`void main` হলে 0) |
| কম্পাইল-টাইম এরর | 1 |
| রানটাইম এরর (ওভারফ্লো, শূন্য দিয়ে ভাগ, …) | 1 |
| ভুল কমান্ড লাইন বা পড়া যায় না এমন ফাইল | 2 |

## টেস্ট চালানো

```sh
go test ./...
```

পরবর্তী: [প্রোগ্রামের গঠন](02-program-structure.md)
