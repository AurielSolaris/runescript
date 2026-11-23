---
title: Getting started
description: Build the RuneScript CLI and run a script.
summary: Install the CLI and run your first program.
section: Guide
order: 1
source: docs/guide/01-getting-started.md
---

## Install

With Go 1.25 or newer, build the CLI from a checkout of RuneScript.

```sh
go build -o runescript ./cmd/runescript
```

## Your first program

Save this as `hello.rsc`.

```runescript
package main;

protected int main() {
    trace("Hello, RuneScript!");
    return 0;
}
```

Run it with `runescript run hello.rsc`.
