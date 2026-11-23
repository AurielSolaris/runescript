# RuneScript performance and tree-shaking plan

Audit baseline: commit `a988284`, Go 1.25.11, Windows amd64. This plan covers the native execution engine, the shared compiler front end, JavaScript and multi-page web output, and the Go/raylib WASM bundle.

## Priority

1. Remove avoidable allocations in checked integer arithmetic and function calls. These are measured costs in the current execution engine.
2. Skip source-map work when maps are off, then measure each compiler stage. Use those measurements to decide where caching or front-end changes pay off.
3. Add typed IR reachability before JavaScript emission. Use the reachable code to select runtime helpers and standard-library functions.
4. Build WASM with a run-only entry point. Evaluate precompiled program input and a bytecode VM after the earlier work has a new baseline.

## Semantic requirements

- Preserve the exact operand, destination, and result types recorded by the checker. Keep explicit integer widths, signedness, nominal aggregate types, and function signatures. Existing contextual literal rules and verified generic `int`/`uint` intermediates continue to apply. Generic intermediates are exact within signed 64-bit range, including negative intermediates in a `uint` expression where the current rules permit them. At each existing range boundary, check the declared destination width. Explicit-width arithmetic keeps its current overflow points.
- Preserve the checker’s requirements for explicit conversions, the exactness rules for contextual literals, and the runtime rules for truncation, rounding, invalid code points, and out-of-range results. Preserve float32 rounding at the existing points, and test NaN, infinities, signed zero, and division separately. Avoid arithmetic reassociation or fused multiply-add that changes rounding or exception order.
- The language has no `any` or `null` type. Runtime slots, temporaries, and serialized operands must have a statically known type. Internal Go `any` at host adapters is an implementation detail with a concrete declared RuneScript signature and a checked boundary. Clear internal references only as frame bookkeeping. Preserve valid zero values and required initializers for types that have no zero value.
- Preserve left-to-right evaluation, short-circuiting, catchable fault type/code/message/site, call-depth failures, and `finally` control flow. Keep struct value semantics, array and class reference semantics, stable array sorting, and callback reentrancy. Keep current diagnostics and warning policy for every loaded source file, including code that is later removed from output.

The language rules are in [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md#5-types). The back-end invariant check is in [ir.Verify](../internal/ir/verify.go#L16).

## What the current code does

| Path | Current behavior | Consequence |
| --- | --- | --- |
| Native `run` | Load, check, lower and verify. The execution engine lazily compiles each called function to Go closures. Unsupported IR cases use the tree evaluator. Frames are already reused. | There is a per-run closure cache and a per-call lookup, and many operations still pass the whole `Value` carrier. |
| `check`, dev, LSP | The loader builds a fresh file and package graph per analysis, and the checker builds full-program semantic maps. Dev rebuilds the JS output. LSP supplies unsaved buffers through the loader overlay. | Large shared dependencies can be read and parsed repeatedly. LSP-specific indexes and diagnostic work need to be measured separately from batch compilation. |
| JS and web | Lowering visits all function bodies and aggregate members. JS verifies IR, names and emits every top-level function, and records helper usage during emission. Web repeats the pipeline per entry and deduplicates byte-identical generated runtime and CSS assets afterward. | Dead functions can pull in code, site entries, CSS, stdlib sections, and extern imports. Independent pages repeat work even where they share source. |
| WASM | The build checks the project on the host, archives loaded source and assets, and builds the full CLI for `js/wasm`. The browser starts its `run` command with the archived entry source. | The host invokes the Go toolchain and the browser runs the RuneScript front end again. JS tree shaking alone cannot change this path. |

Sources: [driver](../internal/driver/driver.go#L70), [closure compiler](../internal/runtime/compile.go#L15), [LSP analysis](../internal/lsp/analyze.go#L31), [IR lowering](../internal/ir/lower.go#L17), [JS generator](../internal/jsgen/jsgen.go#L91), [web page builder](../internal/webbuild/build.go#L142), [WASM bundler](../internal/wasmbuild/bundle.go#L25).

## Baseline measurements

`go test ./...` passed. These are local measurements on an Intel Core Ultra 7 155H, with short samples and some overlapping audit activity. Use them to select work and remeasure each change in isolation.

| Existing small workload | Tree evaluator, ms | Compiled closures, ms | Compiled allocations per run |
| --- | ---: | ---: | ---: |
| Integer addition | 265 | 266 | 64 |
| Function calls | 721 | 743 | 5.60 million |
| Arrays | 357 | 363 | 5.00 million |
| Structs | 505 | 417 | 4.00 million |
| Matrix multiply | 221 | 181 | 1.61 million |
| Recursive Fibonacci | 44 | 49 | 150 thousand |
| Strings | 77 | 54 | 103 thousand |
| Errors | 80 | 81 | 657 thousand |

The table shows medians of three samples. A separate single small run of the division case took 1.11 s and allocated 12.70 million objects. A pooled allocation profile of division, calls, arrays and structs placed 84% of sampled allocations on stacks under `genericArith`. Argument-slice allocation was the largest single flat site by allocated space at 26%. The pooled CPU profile attributed 18% of samples directly to Go's copy routine. After reducing allocations, collect a fresh per-case profile before attributing copy cost to specific callers or changing value layout.

| Compiler stage on a synthetic, mostly-dead, 2,000-function file | Median ms | Allocated MiB |
| --- | ---: | ---: |
| In-memory load and parse | 73.8 | 77.8 |
| Check loaded program | 28.5 | 13.0 |
| Lower checked program | 20.6 | 10.0 |
| Verify existing IR | 4.4 | 1.2 |
| Verify and emit JS, maps off | 68.2 | 35.2 |
| Verify and emit JS, maps on | 69.9 | 42.1 |

The stage measurements use a temporary benchmark and medians of three samples. They exclude disk I/O and process startup. Both JS emission cases currently do marker collection and stripping, so their difference does not measure the proposed map-free fast path. Large-source parsing and mostly-dead emission are the largest measured stages for this fixture. Measure an all-live and a many-file fixture before making a general front-end claim.

The separate five-sample clean CLI benchmark recorded check/build medians of 0.65/0.71 s for the empty program, 0.91/0.98 s for 1,001 files, and 2.19/2.79 s for 20,000 functions. These measurements include process startup and writes for builds. Use the in-process numbers to evaluate individual stages.

| Generated JS | Raw bytes | Minified + gzip bytes | Minified + Brotli bytes |
| --- | ---: | ---: | ---: |
| Empty program | 27,748 | 4,862 | 4,397 |
| One function | 28,023 | 4,920 | 4,456 |
| String workload | 40,628 | 7,117 | 6,449 |
| JSON workload | 52,552 | 9,314 | 8,434 |

The JS size baseline uses the existing `rsbench size` workflow. The WASM and browser-startup findings come from source inspection. Record those baselines before changing the runner.

## 1. Speed up native execution

### 1a. Checked integer operations

- Start in [genericArith](../internal/runtime/runtime.go#L773). Generic division and remainder always enter the `big.Int` path today. Use exact signed 64-bit division and remainder when applicable. Guard zero divisors and handle `MinInt64 / -1` and `MinInt64 % -1` explicitly. Preserve truncation toward zero, the sign of the remainder, and the correct range check at the original operation or destination.
- Extend the fast path to generic addition, subtraction, and multiplication when their exact result fits int64. For overflow, keep the exact path where it is needed to produce the existing diagnostic. Treat generic `uint` intermediate arithmetic as signed exact arithmetic as specified. Handle explicit `uint64` separately because the runtime carrier stores its bits in an `int64` field.
- Independently benchmark [types.Fits](../internal/types/types.go#L377). It reconstructs big integer bounds on each call. Compare an allocation-free fit check using sign, bit length, and the signed minimum edge against the current implementation. Keep the present exact bound checks as the oracle in tests. Follow with native-width implementations for explicit 64-bit operations only where profiles show value.
- Gate the patch with signed/unsigned minima and maxima for every width, generic positive and negative intermediates, division by zero, quotient and remainder sign, and fault diagnostic goldens. Compare native, reference, and JS outcomes.

### 1b. Calls, frames, and values

- [compileArgs](../internal/runtime/compile.go#L172) and the tree evaluator allocate a result slice for argument values. Give small common arities per-activation storage, or an argument region owned by the calling frame. Evaluate arguments before entering the callee, in the current order. Verify with Go escape analysis and `allocs/op`. Keep scratch distinct across nested argument calls, recursion, reentrant stdlib callbacks, and exceptions during argument evaluation. Keep the native bridge’s lifetime requirements explicit.
- After the allocation work, replace the per-run function pointer map with an indexed closure table only if call profiles warrant it. Use a stable function ID for top-level functions, methods, constructors, nested functions, and lambdas. Keep lazy compilation for one-shot scripts and fresh execution state for each run. Measure cold execution separately from repeated execution of a prepared program.
- Inspect remaining full-width `Value` copies and slot clearing. Specialize hot scalar expressions and compare/branch operations using verified types. If that still leaves a measurable cost, test smaller scalar storage and reference storage on a narrow workload before changing the representation of host values, arrays, structs, or objects. Frame clearing must still release references on normal return and unwind.
- Add compiled implementations for frequently sampled fallback nodes after re-profiling. Examine array indexing/updates, struct field operations, and common loop shapes first. Retain shared checked operations and exercise index/receiver evaluation exactly once for compound assignments. Consider eliminating a specific runtime check only when a verified range or control-flow fact proves it cannot fault and no side effect or diagnostic site changes.

### 1c. Bytecode decision

Compare the improved closure engine against a pure-Go register bytecode prototype that reuses existing slots. Start with typed loads/stores, a small checked arithmetic subset, compare-and-branch, and direct calls. Keep an instruction-to-source table and verify opcode types, widths, register initialization, bounds, and branch targets. Use the tree evaluator as the correctness oracle and keep unsupported functions on a whole-function fallback during the experiment. If the prototype is competitive after the allocation fixes, add indirect calls, aggregates, native adapters, and exception regions before considering a default switch. Require a material throughput gain on multiple workloads and a separately measured cold-start budget.

## 2. Reduce compile latency

1. **Separate source maps from ordinary emit.** [GenerateMap](../internal/jsgen/jsgen.go#L91) currently creates source markers and runs [unmark](../internal/jsgen/sourcemap.go#L48) even when `SourceMap` is false. Carry the option into marker emission, guard the footer marker, and skip the strip/mapping pass entirely when disabled. Keep the runtime site table, which serves a separate purpose. Verify that ordinary JS bytes match the existing output, and that enabled source maps still map UTF-16 columns, lambdas, templates, and relocated web paths.
2. **Add stage benchmarks.** Extend the existing generated projects in [rsbench compile](../bench/rsbench/cmd_compile.go#L26), with in-process measurements for read/normalize/lex/parse, load/resolve, check, lower, verify, emit/runtime assembly, map generation, and filesystem writes. Measure both all-live and mostly-dead projects. Capture cold CLI wall time, warm rebuild latency, per-stage B/op, and peak memory. Profile large files and many-file projects separately. Keep target-specific work visible for web and WASM.
3. **Reduce repeated front-end work where profiling confirms it.** Candidate changes include scanning numeric literal values once in lexer/parser, making optional editor indexes conditional on their consumers, and indexing package-level name lookup where repeated scans dominate. Preserve raw token text, exact diagnostics and ordering, checker decisions, and LSP navigation. Make IR statement span collection optional only after auditing its use. Keep operation and declaration spans used for runtime faults.
4. **Cache parse results within one build or editor session.** Start with a shared parse cache for multi-page builds and dev/LSP. Use file identity plus content or overlay revision, preserve each load's source naming and diagnostic context, and confirm ASTs are immutable before sharing them. Build each entry’s package graph independently over cached parses. Invalidate by reverse dependencies and directory membership, including added/deleted files, newly resolved imports, modified metadata, and unsaved overlays. Keep entry, working directory, target and options in the keys for any later checked or emitted products. Bound cache lifetime and memory.
5. **Optimize JS emission after measuring it.** Test appending large/deep expressions into a writer to avoid repeatedly copying growing expression strings. Preserve precedence and names. Memoize runtime assembly per build by the section/helper/table signature, and preserve deterministic hashes. Only then test bounded parallel page work. Keep deterministic diagnostic order and atomic site publication. Check staged asset handling and warning sinks before sharing them across workers.

## 3. Shake generated code safely

### 3a. Remove unreachable RuneScript functions first

- Place reachability between verified IR and JS/web emission. Run full-project checking, target validation, and IR verification before pruning. This keeps errors in unreachable source visible. Start with an emission view over verified IR so dev/LSP and the native reference still have the complete graph.
- Root each standalone entry at `main`. Walk the reachable IR across direct calls, function references, lambdas, method/constructor bodies, and callback expressions in arrays, stdlib calls, and DOM templates or bindings. `CallValue` must traverse its callee and argument expressions. Conservatively retain every function whose value is materialized in reachable IR, even if an indirect call cannot be resolved statically. Treat nested `FuncStmt` as a declaration and follow its function body when referenced. Retain class/default initializer dependencies and traverse `try`, every catch, and `finally`.
- Track live externs separately. A generated ESM import evaluates its module on load, even when the originating function is dead. Preserve the current module import and evaluation order in the initial pass, while pruning unused declarations and adapters within that contract. Decide separately whether changing dead-only import side effects is an acceptable language or build-option change.
- Emit functions in existing order for reproducibility. Preserve usable source maps, relocated imports, and diagnostics for retained functions. For web, compute liveness from each entry. Share reachable application code across pages only after measuring the resulting requests and cached versus first-visit bytes. Keep per-page runtime state, CSS, and the existing content-addressed assets.

### 3b. Trim the JavaScript runtime by dependency

- [JS runtime assembly](../internal/jsgen/jsgen.go#L380) already removes entire optional sections and selects Unicode tables. The core runtime still constructs a broad helper object, and one stdlib call can retain a whole package. Once dead user functions are gone, derive roots from retained IR and the generated module API. Track stdlib overloads by package and emitted JS name.
- Start with a dependency manifest for the largest runtime groups. Record helper-to-helper, stdlib-to-helper, Unicode table and host adapter dependencies. Emit only transitively required definitions and entries in the returned runtime object. Keep launch, error reporting, boundary validation, and exported `run`, `RSError`, and `RuneScriptError` dependencies rooted. Keep the full runtime fixture for direct runtime conformance tests.
- Then consider intra-function dead code. Start with statements after unconditional terminators and provably constant branches. Delete an unused local initializer only when it has no effects and cannot fault, or preserve the initializer evaluation as a standalone statement. Treat conversions, checked arithmetic, indexing, object construction, native calls and callbacks as potentially observable. Preserve catch/finally behavior and diagnostics before considering stronger control-flow analyses.
- Test output structure and end-to-end behavior together. Cover dead recursive groups, callback-only functions, methods and constructors, extern import side effects, Unicode operations, templates, CSS extraction, source maps and runtime errors. Report raw, gzip and Brotli bytes per page, cold site totals, and additional bytes during navigation.

### 3c. WASM and repository cleanup

- Replace the full CLI used for `app.wasm` with a dedicated run-only entry and package boundary. Ensure its import graph excludes JS generation, web build, CLI/dev server and LSP. Split the run path from the combined driver where needed. Compare the linked symbol inventory and compressed bundle sizes, then run a browser smoke test with raylib, raygui and staged assets.
- Measure each WASM cost: RuneScript validation, Go compile/link, asset staging, transfer, and time to the first frame. Reuse the built runner when toolchain, runner sources, dependencies and tags match, with the application archive and assets keyed separately. If browser front-end time remains material, prototype a versioned typed program artifact with checked loading, explicit width tags, source maps or source sites, and atomic output. A smaller package-specific runner is a later option if registry roots still dominate measured size.
- In the Go repository, run a reachability audit against production entry points and separately against tests, graphics tags, and the JS/WASM target. Inspect registration through package initialization, adapters, and exported API before deleting candidates. A concrete cleanup candidate is [gen.diags](../internal/jsgen/jsgen.go#L209), which has one read path and no writers in this checkout. Audit its wrapper return contract before simplifying it. Keep the tree evaluator as a differential test oracle and a current fallback.

## 4. Verification and merge order

| Patch | Initial success criterion |
| --- | --- |
| Exact arithmetic | Edge-case parity and at least 50% fewer allocations on the generic division benchmark. |
| Argument storage | Zero steady-state argument-container allocations for the targeted small-arity call paths. Confirm lower B/op on the calls workload and keep error/callback behavior. |
| Map-free JS emit | Same ordinary JS bytes and correct map output when enabled. Show less allocation on large no-map builds. |
| Reachability | Unreachable bodies and their unused helper roots disappear. Retained programs, diagnostics, imports, and source mapping agree. |
| Runtime selection | Show compressed gains on empty, one-function, string, and DOM cases with full runtime conformance. |
| Parse reuse | Unchanged shared files parse once per build or revision. Show warm rebuild gains on shared modules and clean invalidation after edits. |
| Bytecode or WASM artifact | Require a measured gain large enough to justify another representation, plus differential coverage and, for WASM, browser coverage. |

Use at least 10 paired or interleaved samples on the same machine and settings for patch comparisons. Establish variance before treating a small change as a regression or improvement. Keep separate measurements for fresh CLI invocation, execution on prechecked IR including first-use closure preparation, and execution using prepared code. Run `go test ./...`, the existing tests for integer widths and widening, both native engines against golden fixtures, Node JS and extern fixtures, and web/graphics coverage. Add independent boundary/property tests where the reference and optimized engines share an arithmetic helper. Verify floating-point results by bit pattern where relevant and compare failure codes and positions as well as normal output.

Useful commands are in [bench/README.md](../bench/README.md). The current execution benchmarks in [tests/bench_test.go](../tests/bench_test.go#L65) exclude source compilation from their timed section, but include fresh run setup and lazy closure preparation on each iteration.
