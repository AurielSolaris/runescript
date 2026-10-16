# 1. Getting started

## Install

RuneScript is built from source, so you need **Go 1.25 or newer**
([go.dev/dl](https://go.dev/dl/)). The installers check that first: they
stop with a clear message if Go is missing, too old, or broken.

**Linux and macOS:**

```sh
curl -fsSL https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.sh | sh
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.ps1 | iex
```

The installer:

1. checks Go (on PATH, in its usual install locations, or at `RUNESCRIPT_GO`);
2. builds the `runescript` CLI, which is the compiler and the runtime, with
   that Go (a pure-Go build: no C compiler is needed);
3. installs it in `~/.runescript/bin` (Windows:
   `%LOCALAPPDATA%\Programs\RuneScript\bin`), with no administrator rights;
4. adds that directory to your PATH and sets `RUNESCRIPT_HOME`: in your shell
   startup files (`~/.profile`, `~/.bashrc`, `~/.zshrc`, fish), or in your
   Windows user environment;
5. checks the result by running a small program.

Open a new terminal afterwards, then:

```text
$ runescript version
runescript 0.3.0-dev (commit 9609a35, go1.25.11 linux/amd64)
```

From a clone of the repository, run the installer there to build that
source instead of the latest version: `sh install.sh`, or
`powershell -ExecutionPolicy Bypass -File install.ps1`.

| Option (`install.sh` / `install.ps1`) | What it does |
|---|---|
| `--prefix DIR` / `-InstallDir DIR` | Install into `DIR/bin` (also `RUNESCRIPT_HOME`) |
| `--version REF` / `-Version REF` | Build a tag, branch, or commit from GitHub (also `RUNESCRIPT_VERSION`) |
| `--remote` / `-Remote` | Build from GitHub even inside a clone |
| `--no-modify-path` / `-NoModifyPath` | Leave PATH and the startup files alone |
| `--uninstall` / `-Uninstall` | Remove RuneScript and its PATH entry |

Node.js is optional. It's needed only to run the JavaScript that
`runescript build` writes.

### Without the installer

```sh
go build -o runescript ./cmd/runescript     # runescript.exe on Windows
go run ./cmd/runescript run examples/hello.rsc   # or run without installing
```

## Your first program

Save this as `hello.rsc`:

```actionscript
package main;

protected int main() {
    trace("Hello, RuneScript!");
    return 0;
}
```

Run it:

```text
$ runescript run hello.rsc
Hello, RuneScript!
```

> RuneScript files use the `.rsc` extension, but the CLI accepts any file
> name. Files are UTF-8. A UTF-16 or UTF-32 file (as some Windows editors
> save) is converted, with a note; anything else that isn't UTF-8 is
> rejected with `ASX-LEX-007`.

## Commands

| Command | What it does |
|---|---|
| `runescript run <file>` | Check the program, then run it |
| `runescript check <file>` | Report compile-time errors without running |
| `runescript run --watch <file>` | Run it, then run it again whenever one of its files changes (Ctrl+C stops) |
| `runescript build [-o out.js] <file>` | Compile to a JavaScript module ([chapter 11](11-javascript.md)) |
| `runescript dev [--port 5173] [--open] <file>` | Serve it on localhost and reload the browser on every change ([chapter 11](11-javascript.md#live-reload-runescript-dev)) |

`run`, `check`, `build`, and `dev` also take `-Werror` and `--allow <codes>`
for warnings (see [chapter 8](08-diagnostics.md#warnings)).
| `runescript version` | Print the version |

## Exit status

| Situation | Exit status |
|---|---|
| Program ran | The `int` returned by `main` (0 for `void main`) |
| Compile-time error | 1 |
| Runtime error (overflow, division by zero, …) | 1 |
| Bad command line or unreadable file | 2 |

## Running the tests

```sh
go test ./...
```

Next: [Program structure](02-program-structure.md)
