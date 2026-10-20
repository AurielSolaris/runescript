# RuneScript benchmark and conformance suite

This suite compares RuneScript with **TypeScript 1.0 RC**, its contemporary
baseline. Correctness, runtime speed, generated-code size, compiler speed,
and diagnostics are measured separately. Every benchmark is a small,
self-contained directory, so you can add, run, or delete one without
touching the others.

Each program runs on three implementations, so you can tell where a
difference comes from:

| Implementation | Pipeline | Tells you about |
|---|---|---|
| `rs-vm` | RuneScript → the Go interpreter (`runescript run`) | the interpreter |
| `rs-js` | RuneScript → JavaScript (`runescript build`) → Node.js | the code generator and its runtime helpers |
| `ts` | TypeScript 1.0 RC → JavaScript (`tsc`) → Node.js | the baseline |
| `ts-<variant>` | another TypeScript version of the same program, e.g. `ts-checked` | an equivalent-semantics comparison |

The results are published only in the repository README, not on the
website.

## Setup

```sh
cd bench && npm install && cd ..   # pins TypeScript 0.9.7 and esbuild 0.25.9
```

You also need Go and Node.js on `PATH`. Without the npm install, the
TypeScript columns show as unavailable and minified sizes are left empty.

**Which TypeScript?** npm has no package called "1.0 RC".
`typescript@0.9.7` was published on 2014-02-25, the day the 1.0 RC was
announced. It is the RC build, and `tsc -v` reports `0.9.7.0`. Comparing
against a modern TypeScript 5 would be unfair. TypeScript 1.0 RC has
these limits, which shape the TypeScript sources and flags:

- No `let`/`const`, template strings, union types, or ES2015 library
  (`Math.fround`, `String.prototype.startsWith`, …). Write ES5-era code,
  and reach ES2015 functions through `any`.
- No `@types/node`: each program starts with `declare var process: any;`.
- No `--strict` and no `--noEmit`. Every program is compiled with
  `--target ES5 --noImplicitAny --module commonjs`, and `tsc` always
  checks *and* emits.

## Commands

All commands run from the repository root:

```sh
go run ./bench/rsbench list                     # what exists
go run ./bench/rsbench verify                   # correctness: every program, every implementation
go run ./bench/rsbench run -size small          # runtime (default: 2 warm-up + 10 measured runs)
go run ./bench/rsbench run -size medium a07_matmul a11_three_body -runs 20
go run ./bench/rsbench run -impl rs-js,ts       # only some implementations
go run ./bench/rsbench size                     # generated JS: raw, minified, gzip, brotli
go run ./bench/rsbench compile [-large]         # runescript check/build vs tsc on generated projects
go run ./bench/rsbench diag                     # the same mistakes in both languages
go run ./bench/rsbench bless [benchmark...]     # record expected/test.out from the interpreter
```

Benchmarks are named by directory (`a03_sieve`), ID (`A03`), or name
(`sieve`). With no names, every benchmark in the command's category is
used. Each command writes a CSV and a Markdown summary to
`bench/results/` (git-ignored) and prints the summary.

`go test ./...` also checks the RuneScript side of the suite
(`tests/bench_test.go`): every benchmark is well formed, every program
prints its expected output on both back ends at the `test` size, and
every diagnostics case is rejected on the right line.

## Layout

```text
bench/
├── rsbench/              the runner (Go, part of the main module)
├── suite/<id>_<name>/    one directory per benchmark
│   ├── bench.json        id, name, title, categories, what it measures, sizes, notes
│   ├── main.rsc          the RuneScript program; process.args()[0] is the size
│   ├── main.ts           the TypeScript 1.0 program; process.argv[2] is the size
│   ├── checked.ts        optional variant → implementation "ts-checked"
│   └── expected/
│       ├── test.out      expected output at the "test" size (all implementations)
│       └── test.ts.out   optional: an intended TypeScript divergence
├── diagnostics/<id>_<name>/   bad.rsc + bad.ts with the same mistake, bench.json
├── tools/sizes.mjs       raw / esbuild-minified / gzip / brotli byte counts
├── package.json          the pinned TypeScript-side toolchain
└── results/              CSV and Markdown output (git-ignored)
```

### Categories

- **runtime** is measured by `run`. These benchmarks need `test` and
  `small` sizes, and usually also `medium` and `large`.
- **size** is measured by `size`.
- **conformance** is checked by `verify` only. These programs test
  semantics, not speed.

### Adding a benchmark

1. Create `suite/<id>_<name>/` with `bench.json`, `main.rsc`, and
   `main.ts`. Use the same algorithm, the same inputs, and the same output
   format in both. Print a checksum, so that no engine can skip the work
   and every implementation can be checked.
2. Keep integers inside the range RuneScript's `int` can store (int32),
   and keep TypeScript intermediates below 2^53. The Park–Miller generator
   `seed = seed * 16807 % 2147483647` meets both limits.
3. Run `go run ./bench/rsbench bless <name>`, read
   `expected/test.out`, then run `go run ./bench/rsbench verify <name>`.
4. If TypeScript cannot produce the same output by design (overflow,
   string units, float32 printing), explain why in `notes`. Then record
   TypeScript's output with `bless -from ts <name>`. `go test` rejects a
   `.ts.out` file that has no notes.

## Measurement protocol

- **Correctness first.** Before a runtime cell counts, its output is
  compared with `expected/<size>.out`, or with the interpreter's output
  at sizes without an expected file. `run` reports this in the *Correct*
  column.
- **Separate processes.** Every run is a new process. The timer covers
  start to exit, so startup is included. The empty program (`S01`) is
  always measured too, and *net* time is the median minus that
  implementation's empty-program median. `ts-*` variants use `ts`'s
  baseline.
- **Warm-up runs, then rounds in random order.** Each round runs every
  (benchmark, implementation) pair once, in a newly shuffled order, so
  thermal drift and background load spread across all cells. The seed is
  printed, so you can repeat an order.
- **Statistics.** The results include median, min, max, mean, standard
  deviation, and a 95% interval of the mean. Ratios use net medians:
  *R* = TS ÷ RS-JS (above 1 means RuneScript's JavaScript was faster) and
  *S* = RS-VM ÷ RS-JS (the interpreter slowdown). A ratio is left out
  when a net time is under 1 ms.
- **Sizes** are measured on standalone compiler output, which is already
  a single-file bundle for these programs. RuneScript modules inline only
  the runtime sections they use. TypeScript 1.0 output has no runtime.
  Minification uses esbuild with the same settings for both, gzip uses
  level 9, and Brotli uses quality 11. For tiny programs, read the
  absolute Δ, not the ratio.
- **Compiler** runs are clean builds into an empty output directory.
  `rs-check` (check only) has no TypeScript 1.0 equivalent. Compare `tsc`
  with `rs-build`.

### Known limits

- **Peak memory on Windows** is not recorded. Go's standard library does
  not expose it. The Windows API that does (`GetProcessMemoryInfo`) needs
  `unsafe` or a new dependency, and neither is approved for this project.
  On Linux and macOS, `peak_memory_mb` comes from `getrusage`.
- **No in-process timer.** RuneScript has no clock API yet, so steady-state
  (post-JIT) time can't be measured inside a program. Use a larger size,
  where startup becomes negligible, as the steady-state proxy.
- Not implemented yet from the original plan: incremental and cached
  builds (neither compiler has them), editor/LSP measurements, fuzzed
  differential programs (the main test suite already does random
  interpreter-vs-Node differentials for arithmetic, widths, and
  `std:path`), and benchmarks that need features RuneScript doesn't have
  yet (classes, maps, a clock).
