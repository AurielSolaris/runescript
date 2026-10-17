# stdlib

Standard library packages, implemented natively in Go and available to
RuneScript programs through `use "std:<name>";`.

| Package | Status | Source |
|---|---|---|
| `std:string` | v0.1 | [`string.go`](string.go): Go's `strings`, camelCase names, byte offsets |
| `std:math` | 2026-10-17 | [`math.go`](math.go): Go's `math`, bit-identical on both targets (`internal/fmath`) |
| `std:io` | 2026-10-17 | [`io.go`](io.go): `print`, `error`, `warn`, standard input; [`utf8.go`](utf8.go) decodes input the WHATWG way |
| `std:process` | 2026-10-17 | [`process.go`](process.go): arguments, environment, working directory, platform |
| `std:fs` | 2026-10-17 | [`fs.go`](fs.go): files and directories, with the same error messages on both targets |
| `std:path` | 2026-10-17 | [`path.go`](path.go): Node.js's `path` semantics, tested against Node.js |

Functions that reach outside the program (`std:io`, `std:process`,
`std:fs`, `std:path`) take an [`Env`](env.go): the program's streams,
arguments, and platform. Every package has a JavaScript twin in
[`internal/jsgen/runtime.js`](../internal/jsgen/runtime.js), and the tests
run both on the same programs.

User documentation: [`std:string`](../docs/guide/10-std-string.md),
[`std:math`](../docs/guide/18-numbers.md), and
[files, input, and output](../docs/guide/20-files-and-input.md).

The host-neutral JavaScript modules for extern functions (`rune:fs`) are
not part of the standard library: they run only on the JavaScript target,
and live in [`internal/jshost`](../internal/jshost/jshost.go) (guide
[chapter 11](../docs/guide/11-javascript.md#host-neutral-modules)).
