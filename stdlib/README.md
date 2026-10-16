# stdlib

Standard library packages, implemented natively in Go and available to
RuneScript programs through `use "std:<name>";`.

| Package | Status | Source |
|---|---|---|
| `std:string` | v0.1 | [`string.go`](string.go) — Go's `strings`, camelCase names, byte offsets |

User documentation: [`docs/guide/10-std-string.md`](../docs/guide/10-std-string.md).

The host-neutral JavaScript modules for extern functions (`rune:fs`) are
not part of the standard library: they run only on the JavaScript target,
and live in [`internal/jshost`](../internal/jshost/jshost.go) (guide
[chapter 11](../docs/guide/11-javascript.md#host-neutral-modules)).
