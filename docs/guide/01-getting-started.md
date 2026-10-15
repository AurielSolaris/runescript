# 1. Getting started

## Build

RuneScript is written in Go (1.25 or newer).

```sh
go build -o runescript ./cmd/runescript     # runescript.exe on Windows
```

Or run it without installing:

```sh
go run ./cmd/runescript run examples/hello.rsc
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

> RuneScript files use the `.rsc` extension. The CLI accepts any file name.

## Commands

| Command | What it does |
|---|---|
| `runescript run <file>` | Check the program, then run it |
| `runescript check <file>` | Report compile-time errors without running |
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
