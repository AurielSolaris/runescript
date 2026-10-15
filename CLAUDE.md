# RuneScript --- Claude Code Project Instructions

> **Project status:** Language design and implementation specification,
> version 0.1 (draft).
>
> RuneScript is a standalone, statically typed, general-purpose
> programming language inspired by ActionScript, with selected syntax
> and ergonomics from C++ and Kotlin. The compiler and runtime are
> implemented in Go. It is **not** a web-only language and must not
> require a browser, HTML, or Flash Player.

## 1. Instructions for Claude Code

Treat this file as the source of truth for the project. Before
implementing a feature:

1.  Read this file and inspect the existing repository.
2.  Implement incrementally, in small, independently testable steps.
3.  Add or update tests for every language feature and bug fix.
4.  Keep the lexer, parser, AST, semantic analysis/type checker,
    runtime, standard library, CLI, and tests in clearly separated
    packages.
5.  Do not silently invent language semantics where this specification
    is undecided. Record the question in `docs/OPEN_QUESTIONS.md`,
    propose a small number of alternatives, and ask the project owner
    when the choice affects compatibility, typing, memory layout, or
    runtime behavior.
6.  Do not claim that RuneScript is AS3-source-compatible. Compatibility
    has not been promised.
7.  Do not implement unrelated features just because they are easy to
    generate. Prefer a coherent, working language over a large
    collection of incomplete features.
8.  Keep generated code readable and idiomatic Go. Do not use `unsafe`
    unless explicitly justified and approved.
9.  Avoid premature optimization. Begin with a correct, testable
    implementation; measure before optimizing.
10. Keep the README, examples, and tests consistent with the actual
    implementation. Clearly label unimplemented syntax in documentation.

## 2. Project identity

-   **Language name:** RuneScript
-   **Implementation language:** Go
-   **Language family:** ActionScript-inspired, C-family syntax,
    statically typed, general-purpose
-   **Design influences:** ActionScript's identity and object-oriented
    heritage; C++-style declarations and blocks; Kotlin-inspired
    reduction of unnecessary boilerplate and selected modern type-system
    conveniences.
-   **Executable file extension:** `.rs` is a candidate, but must be
    checked for naming conflicts and finalized before the CLI depends on
    it. Do not assume the extension is settled.
-   **CLI name:** `runescript` is the preferred command name, subject to
    repository/package naming constraints.

RuneScript should support small scripts, command-line tools, and larger
object-oriented programs. It must not require a class merely to express
every entry point or small utility, unless the project owner explicitly
chooses that restriction later.

## 3. Syntax principles

RuneScript uses a familiar C-family declaration style:

-   Types precede names: `int count`, `string message`.
-   Functions do not require an ActionScript-style `function` keyword.
-   Function return types appear before function names: `int main()`.
-   Semicolons terminate declarations and statements.
-   Braces delimit blocks.
-   Package declarations are terminated by semicolons; they are not
    enclosing brace blocks.
-   Keep syntax explicit and readable, while reducing historical
    boilerplate.
-   Do not replace this design with Kotlin syntax such as `fun`, `val`,
    or `init` unless the owner changes the specification.

### 3.1 Canonical syntax example

``` actionscript
package main;

protected int main() {
    // do stuff
    return 0;
}
```

This is the currently preferred illustrative entry-point form. The
precise rules for entry-point visibility, whether it must be a free
function or may be a class member, and how it is invoked by the runtime
remain open questions.

### 3.2 Class-oriented example

The following illustrates the intended style, not a claim that every
detail is already implemented:

``` actionscript
package org.apache.royale;

public class WelcomeToActionScript {
    public void greet() {
        string message = "Hello world";
        sayHi(message, 3);
    }

    private void sayHi(string message, int times) {
        for (int i = 0; i < times; i++) {
            trace(message);
        }
    }
}
```

RuneScript should preserve the ability to use classes, methods,
visibility modifiers, and typed parameters without requiring
ActionScript's `function name(...):ReturnType` syntax.

## 4. Type-system specification

The type system must be explicit and checked by the compiler. Do not
introduce an implicit `any` type. Do not silently reinterpret signed and
unsigned integers.

The following lists reflect the current design. Where behavior is not
fully specified, implement only the unambiguous subset and document the
unresolved decision.

### 4.1 Generic / convenience types

-   `int`
-   `float`
-   `string`
-   `char`
-   `bool`
-   `void`

The owner has also referred to generic types and generics. **Generic
type parameters** (for example, `Array<T>`) are a separate language
feature from the convenience type names above; their syntax and
constraints are not yet finalized.

### 4.2 Numeric representation types

Currently named explicit-width types:

-   Signed: `int8`, `int16`, `int32`
-   Floating point: `float32`, `float64`

The type family should also consider `int64`, `uint8`, `uint16`,
`uint32`, and `uint64` so widening and unsigned widths can be
represented consistently. Add these as part of the implementation only
after documenting the decision in the type specification; do not
silently treat the currently listed types as exhaustive.

### 4.3 Default signedness

-   Integer types are **signed by default**.
-   `int` means a signed integer.
-   `sint` is an explicit spelling of the same signed-integer semantics
    as `int`. It exists for readability and developer confidence; it
    must not introduce a distinct runtime representation or incompatible
    type.
-   `uint` means an unsigned integer.
-   Explicit-width signed and unsigned types must retain their
    signedness.

### 4.4 Generic numeric defaults and widening

-   Generic `int` starts with a default representation of `int32`.
-   Generic `float` starts with a default representation of `float32`.
-   Generic numeric types may **widen** when the value or operation
    requires a wider representation.
-   Generic numeric types must not automatically contract to a narrower
    representation after widening.
-   Widening must preserve the numeric value whenever the destination
    representation can represent it.
-   The precise widening ladder, overflow behavior, literal inference,
    variable storage model, and expression-promotion rules are not yet
    finalized. Do not invent incompatible rules independently in
    different compiler stages.
-   In particular, do not assume that Go's architecture-dependent native
    `int` defines RuneScript's `int` semantics. RuneScript behavior
    should be consistent across supported platforms.
-   Floating-point widening must account for precision as well as
    magnitude. A value fitting within `float32`'s magnitude range does
    not necessarily have the precision of `float64`.

**Important open design point:** The owner has stated that generic
numeric types widen but do not contract. It remains to be decided
whether a variable declared `int` can change its internal representation
during its lifetime or whether widening applies to inferred values and
expression results while declared variable types remain stable. Document
this distinction before implementing assignments, arrays, structs,
method parameters, or memory layout.

### 4.5 Signed/unsigned assignment safety

-   Assigning a signed integer value or expression to a `uint`
    destination must produce a **compile-time error** unless the
    programmer explicitly resolves the signedness mismatch.
-   The compiler must not silently wrap, reinterpret, or guess the
    intended conversion.
-   A signed type can hold negative values even when a particular
    runtime value happens to be positive. The source type---not merely
    the current value---matters.
-   Diagnostics should explain the mismatch and suggest an explicit,
    checked conversion or changing the destination type.
-   The exact explicit-conversion syntax and rules for checked
    conversions remain to be finalized.
-   Define corresponding rules for unsigned-to-signed assignments and
    mixed-signedness arithmetic explicitly; do not assume these cases
    have already been settled.
-   Where a value is statically known to be negative, conversion to an
    unsigned type should be rejected at compile time.

Illustrative invalid example:

``` actionscript
sint signedValue = -10;
uint unsignedValue = signedValue; // Compile-time error
```

Illustrative diagnostic style:

``` text
ASX-TYPE-004: Cannot assign signed integer 'sint' to unsigned integer 'uint'.
Help: Use an explicit checked conversion, or change the destination type.
```

`ASX` in this example is only a provisional diagnostic prefix; settle a
RuneScript-specific prefix before standardizing diagnostic codes.

### 4.6 Other types

-   `string`: text type.
-   `char`: character type. Encoding, Unicode model, and literal rules
    need to be specified.
-   `bool`: boolean type.
-   `void`: absence of a return value; it is not a normal storable
    value.
-   `array`: built-in array/collection category; syntax, mutability, and
    bounds behavior remain to be defined.
-   `matrix`: special built-in type/category; dimensionality, element
    types, operators, and storage semantics remain to be defined.
-   `struct`: user-defined value/data-layout type; copy semantics,
    layout guarantees, and method support remain to be defined.
-   `class`: user-defined reference/object type; allocation,
    inheritance, constructors, and runtime behavior remain to be
    defined.

Do not assume that `struct` and `class` have identical memory or copying
semantics. Define their differences in the specification before relying
on them in the runtime.

## 5. Functions, visibility, and control flow

Current syntax direction:

``` actionscript
protected int main() {
    return 0;
}

private void sayHi(string message, int times) {
    for (int i = 0; i < times; i++) {
        trace(message);
    }
}
```

-   Return types precede function names.
-   Parameters have explicit types.
-   `private`, `protected`, and `public` are desired visibility
    modifiers, subject to a formal scope/accessibility specification.
-   `return` returns a value or exits a `void` function.
-   Conventional `if`, `else`, `for`, and `while` blocks are intended.
-   `trace(...)` is a familiar ActionScript-inspired debugging/output
    facility and should be available in the initial language/runtime.
-   Exact rules for default visibility, top-level functions, overloads,
    default arguments, generics, and return-type inference remain open.

## 6. Implementation architecture

Implement RuneScript in Go. Keep the implementation modular; an initial
layout may be adapted to the existing repository:

``` text
runescript/
├── CLAUDE.md
├── README.md
├── go.mod
├── docs/
│   ├── LANGUAGE_SPEC.md
│   ├── OPEN_QUESTIONS.md
│   └── ROADMAP.md
├── cmd/
│   └── runescript/
├── internal/
│   ├── lexer/
│   ├── parser/
│   ├── ast/
│   ├── checker/
│   ├── types/
│   ├── runtime/
│   ├── bytecode/       # only if/when a VM is implemented
│   └── diagnostics/
├── stdlib/
├── tests/
└── assets/
    ├── runescript.svg
    └── runescript.png
```

This is a suggested structure, not permission to restructure an existing
repository blindly. Inspect the repository first and preserve existing
working conventions where reasonable.

### 6.1 Implementation stages

1.  Write down the initial grammar and unresolved decisions.
2.  Implement lexer and source-position-aware diagnostics.
3.  Implement parser and AST for package declarations, functions,
    blocks, literals, variables, calls, return statements, and basic
    arithmetic.
4.  Implement semantic analysis and type checking, including
    signed/unsigned assignment rejection.
5.  Implement a minimal evaluator/interpreter and `trace`.
6.  Add conditionals, loops, and functions.
7.  Add classes, structs, arrays, and the remaining type features in
    small increments.
8.  Add CLI behavior, standard library modules, tests, and
    documentation.
9.  Consider bytecode or other performance work only after the
    language's semantics are stable and measured.

A tree-walking interpreter is an acceptable starting point. Do not build
a bytecode VM merely for architectural appearance.

## 7. Tests and diagnostics

-   Every language feature must have positive tests and, where relevant,
    negative compile-time tests.
-   Add regression tests for every fixed bug.
-   Test numeric boundaries, widening, non-contraction, signedness
    mismatches, invalid conversions, and mixed-type expressions.
-   Include tests for source locations in diagnostics.
-   Tests must run with standard Go tooling where possible
    (`go test ./...`).
-   Never make a test pass by weakening a type-safety rule without an
    explicit design decision.
-   Do not report a feature as supported until a test demonstrates it.

## 8. RuneScript logo

Create a RuneScript logo featuring an **electrically charged snowflake
rendered in a rune-like visual style**.

Requirements:

-   Draw the logo as original, hand-authored **SVG vector markup**. Do
    not use an AI image-generation tool.
-   The snowflake should have a strong, recognizable central structure
    and angular, rune-inspired branches.
-   Suggest electrical charge with a few deliberate lightning-like
    accents or controlled arcs. Keep the silhouette legible at small
    sizes.
-   Use a restrained, high-contrast palette; the mark should also remain
    recognizable in monochrome.
-   Prefer clean paths, simple geometry, and a transparent background.
-   Include a version suitable for a README and a small application
    icon.
-   Store the source as `assets/runescript.svg`.
-   Rasterize that SVG into `assets/runescript.png` using an SVG
    rasterizer available in the development environment. The PNG must be
    generated from the SVG source, not drawn independently.
-   If a rasterizer is unavailable, document the missing dependency and
    the exact command needed; do not pretend rasterization succeeded.
-   Keep the SVG source as the canonical asset so it can be edited
    later.
-   Do not add unrelated text or intricate detail that disappears when
    the icon is scaled down.

## 9. Documentation and delivery

-   `README.md` should explain what RuneScript is, its current
    implementation status, how to build/run it, and include the logo.
-   `docs/LANGUAGE_SPEC.md` should contain the evolving formal language
    specification. Keep it synchronized with this file; when the two
    conflict, flag the conflict rather than silently choosing one.
-   `docs/OPEN_QUESTIONS.md` should track unresolved design decisions
    with status and rationale.
-   `docs/ROADMAP.md` should track small, testable milestones.
-   Avoid claiming AS3 compatibility, native compilation, concurrency,
    graphics support, or a complete standard library until those
    features exist and are tested.

## 10. Definition of done for the first milestone

The first milestone is complete when RuneScript can:

1.  Read a `.rs`-style source file (final extension still to be
    decided).
2.  Parse a package declaration and a function using the chosen
    declaration syntax.
3.  Parse basic integer/string literals, local declarations, function
    calls, and `return`.
4.  Execute a minimal entry point.
5.  Provide a working `trace` operation.
6.  Reject at least the specified signed-to-unsigned assignment with a
    useful compile-time diagnostic.
7.  Run its automated tests successfully.
8.  Include the SVG logo source and, if the required rasterizer is
    available, the rasterized PNG.

Do not implement the entire language in one pass. Deliver each milestone
as a small, tested change.
