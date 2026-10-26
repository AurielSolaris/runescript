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
| `std:json` | 2026-10-20 | [`json.go`](json.go): the opaque `json.Value`; one parser on both targets |
| `std:csv` | 2026-10-26 | [`csv.go`](csv.go): RFC 4180, as `Array<Array<string>>` |
| `std:matrix` | 2026-10-26 | [`matrix.go`](matrix.go): linear algebra on `Array<Array<float64>>`, bit-identical on both targets |
| `std:semver` | 2026-10-26 | [`semver.go`](semver.go): SemVer 2.0.0 with npm's ranges, tested against npm's semver package |

Functions that reach outside the program (`std:io`, `std:process`,
`std:fs`, `std:path`) take an [`Env`](env.go): the program's streams,
arguments, and platform. Every package has a JavaScript twin in
[`internal/jsgen/runtime.js`](../internal/jsgen/runtime.js), and the tests
run both on the same programs.

User documentation: [`std:string`](../docs/guide/10-std-string.md),
[`std:math`](../docs/guide/18-numbers.md),
[files, input, and output](../docs/guide/20-files-and-input.md),
[`std:json`](../docs/guide/21-json.md), and
[`std:csv`, `std:matrix`, and `std:semver`](../docs/guide/22-csv-matrix-semver.md).

The host-neutral JavaScript modules for extern functions (`rune:fs`) are
not part of the standard library: they run only on the JavaScript target,
and live in [`internal/jshost`](../internal/jshost/jshost.go) (guide
[chapter 11](../docs/guide/11-javascript.md#host-neutral-modules)).
