# RuneScript Language Specification

**Version 0.3 (draft, in progress).** This is the formal companion to `CLAUDE.md`. If
the two disagree, the disagreement is flagged in §11 rather than silently
resolved.

Every rule carries one of these labels:

| Label | Meaning |
|---|---|
| **[v0.1]**, **[v0.2]**, **[v0.3]** | Implemented, and covered by tests, in that version. |
| **[v0.1, provisional]** | Implemented, but the owner has not decided the rule. It is the most conservative choice and may change. |
| **[planned]** | Required by the design; not implemented. |
| **[undecided]** | Not yet designed. Programs using it are rejected. |

RuneScript is inspired by ActionScript 3 but is **not** source-compatible
with it, and no compatibility is promised. Where RuneScript's design differs
from AS3, RuneScript's design wins. §10 summarises the differences.

---

## 1. Source text

- **[v0.1]** Source files are UTF-8. **[v0.3, decided: "try to convert,
  else reject"]** A file in another Unicode encoding that can be recognised
  with certainty is converted when it is read: UTF-8 with a byte-order mark
  (dropped), UTF-16 or UTF-32 with a byte-order mark, and UTF-16 without
  one (NUL bytes beside ASCII characters). The CLI notes each conversion.
  Any other file that is not valid UTF-8 is rejected with `ASX-LEX-007` at
  its first invalid byte, on every target. Legacy 8-bit encodings are not
  guessed. So every RuneScript string is valid UTF-8.
- **[v0.3, decided]** Programs conventionally use `.rsc`, and the files of a
  package directory use `.rsm` (§4.2). As in Perl, the extension is a
  convention, not a rule: a file named on the command line or in a file
  `use` may have any extension, as long as it contains valid RuneScript.
- **[v0.1]** Positions in diagnostics are `line:column`, both 1-based.
  Columns count Unicode code points. **[v0.3]** In a program of several
  files, diagnostics also name the file: `geom/area.rsm:4:12`. A file is
  named by its path relative to the working directory.

## 2. Lexical structure

### 2.1 Whitespace and comments **[v0.1]**
Spaces, tabs, CR, and LF separate tokens. `// …` runs to the end of the
line. `/* … */` is a block comment and does not nest. An unterminated block
comment is an error.

### 2.2 Identifiers **[v0.1, provisional]**
`[A-Za-z_][A-Za-z0-9_]*`. Only ASCII is accepted for now.

### 2.3 Keywords
Implemented keywords **[v0.1]**:

```
package use return if else while for break continue true false
public protected private
int sint uint bool string void
float float32 float64                 [v0.3]
char                                  [2026-10-16]
int8 int16 int32 int64                [2026-10-17]
uint8 uint16 uint32 uint64            [2026-10-17]
try catch finally throw new           [v0.3]
extern                                [2026-10-16]
struct                                [2026-10-17]
```

`as` is a keyword only after a `use` path (§4.2). `Error`, `RangeError`,
`TypeError`, and `ArgumentError` are predeclared type names, not keywords.
**[2026-10-16]** `template` is a keyword only when `{` follows it (§13), and
`Element`, `Signal`, and `signal` are names that `use "rune:dom";` makes
available.

Reserved for planned features **[v0.1, provisional]**. They are rejected
with `ASX-PARSE-002`:

```
class array matrix
```

`trace` is not a keyword. It is a built-in function name and can't be
redeclared.

### 2.4 Literals
- **Integer [v0.1]**: decimal digits. `0` is valid, but other numbers with
  leading zeros (`007`) are rejected **[provisional]**. **[2026-10-17]**
  Also hex `0xFF`, binary `0b1010`, and octal `0o17` (lower-case prefixes
  only), and `_` between two digits as a separator: `1_000_000`, `0xFFFF_0000`.
  The largest literal is 2^64 - 1. Suffixes are **[undecided]**. Literals are
  non-negative; `-5` is unary minus applied to `5`. Any malformed literal is
  `ASX-LEX-005`.
- **String [v0.1]**: `"…"` on one line. Escapes: `\n \t \r \\ \" \'`, `\0`
  (U+0000), and **[2026-10-17]** `\u{X…}`: 1 to 6 hex digits naming a
  Unicode scalar value, e.g. `\u{E9}` or `\u{1F600}`. A surrogate
  (`\u{D800}`–`\u{DFFF}`) or a value above `\u{10FFFF}` is `ASX-LEX-003`, so
  every string is valid UTF-8. There are no byte escapes (`\x41`). Any other
  escape is an error **[provisional]**.
- **Interpolated string [decided 2026-10-16]**: `$"text {expr} text"` on one
  line. Each `{expr}` hole holds one expression, which may contain strings
  and other interpolated strings; braces inside a hole nest. In the text,
  `{{` and `}}` stand for `{` and `}`, a single `}` is `ASX-LEX-003`, and
  escapes are those of strings. A hole's value is converted to text as
  `(string) x` converts it (§5.6), so it must be a number, `bool`, or
  `string` (`ASX-TYPE-017`). Without holes, `$"…"` is an ordinary string.
  Format specifiers (`{x:5}`) are **[undecided]**.
- **Boolean [v0.1]**: `true`, `false`.
- **Float [v0.3]**: digits, then optionally `.` and digits, then
  optionally an exponent: `64.3`, `1e10`, `2.5E-3`. Both sides of the `.`
  need digits (`.5` and `5.` are not literals). `_` may separate digits:
  `1_000.5`. There are no hex floats. A literal stands for the
  nearest float64 to its decimal value; one too large for float64, or so
  small it would round to zero, is `ASX-LEX-005`.
- **Char [decided 2026-10-16]**: exactly one character between single
  quotes: `'a'`, `'é'`, `'😀'`. Escapes are those of strings.
  Anything other than one character (`''`, `'ab'`) is `ASX-LEX-006`.

### 2.5 Operators and punctuation **[v0.1]**
```
( ) { } , ; .
=  +=  -=  *=  /=  %=  &=  |=  ^=  <<=  >>=  ++  --
+  -  *  /  %  &  |  ^  ~  <<  >>
== != <  <= >  >=
&& || !
```

## 3. Grammar (v0.2)

EBNF. `{ x }` means zero or more, `[ x ]` means optional.

```ebnf
File         = PackageDecl { UseDecl } { FuncDecl | ExternDecl } EOF ;
ExternDecl   = "extern" StringLit Type Ident "(" [ Params ] ")" ";" ;   (* §12 *)
PackageDecl  = "package" Ident { "." Ident } ";" ;
UseDecl      = "use" StringLit [ "as" Ident ] ";" ;   (* "as" is contextual *)

FuncDecl     = Visibility Type Ident "(" [ Params ] ")" ( Block | ";" ) ;
Visibility   = "public" | "protected" | "private" ;
Params       = Param { "," Param } ;
Param        = Type [ Ident ] ;          (* names are optional only in prototypes *)
Type         = BaseType { "(" [ Type { "," Type } ] ")" } ;   (* function types *)
BaseType     = "int" | "sint" | "uint" | "bool" | "string" | "void"
             | Ident [ "<" Type { "," Type } ">" ] ;           (* Signal<int>, §13 *)

Block        = "{" { Stmt } "}" ;
Stmt         = FuncDecl                 (* nested function or prototype *)
             | SimpleStmt ";"
             | "return" [ Expr ] ";"
             | "break" ";" | "continue" ";"
             | "throw" Expr ";"
             | IfStmt | WhileStmt | ForStmt | TryStmt
             | TemplateStmt ;                               (* §13 *)
SimpleStmt   = VarDecl | Assignment | IncDec | Call ;
VarDecl      = Type Ident "=" Expr ;
Assignment   = ( Ident | Postfix "[" Expr "]" ) AssignOp Expr ;
AssignOp     = "=" | "+=" | "-=" | "*=" | "/=" | "%=" | "&=" | "|=" | "^=" | "<<=" | ">>=" ;
IncDec       = Ident ( "++" | "--" ) ;

IfStmt       = "if" "(" Expr ")" Block [ "else" ( IfStmt | Block ) ] ;
WhileStmt    = "while" "(" Expr ")" Block ;
ForStmt      = "for" "(" [ SimpleStmt ] ";" [ Expr ] ";" [ ForPost ] ")" Block ;
ForPost      = Assignment | IncDec | Call ;
TryStmt      = "try" Block { "catch" "(" Type Ident ")" Block } [ "finally" Block ] ;
               (* at least one catch clause or a finally block *)

Expr         = OrExpr ;
OrExpr       = AndExpr { "||" AndExpr } ;
AndExpr      = BitOr   { "&&" BitOr } ;
BitOr        = BitXor  { "|" BitXor } ;
BitXor       = BitAnd  { "^" BitAnd } ;
BitAnd       = EqExpr  { "&" EqExpr } ;
EqExpr       = RelExpr { ( "==" | "!=" ) RelExpr } ;
RelExpr      = Shift   { ( "<" | "<=" | ">" | ">=" ) Shift } ;
Shift        = AddExpr { ( "<<" | ">>" ) AddExpr } ;
AddExpr      = MulExpr { ( "+" | "-" ) MulExpr } ;
MulExpr      = Unary   { ( "*" | "/" | "%" ) Unary } ;
Unary        = ( "-" | "!" | "~" ) Unary | Cast | Postfix ;
Cast         = "(" CastType ")" Unary ;
CastType     = "int" | "sint" | "uint" | "int8" | "int16" | "int32" | "int64"
             | "uint8" | "uint16" | "uint32" | "uint64"
             | "float" | "float32" | "float64" | "string" | "bool" | "char" ;
Postfix      = Primary { "(" [ Expr { "," Expr } ] ")" | "." Ident | "[" Expr "]" } ;   (* calls, fields, elements *)
Primary      = IntLit | FloatLit | StringLit | InterpString | CharLit | "true" | "false"
             | Ident | Call | Lambda | New | "(" Expr ")"
             | "[" [ Expr { "," Expr } ] "]" ;   (* array literal, §5.1c *)
InterpString = '$"' { Text | "{" Expr "}" } '"' ;   (* §2.4 *)
New          = "new" Ident "(" [ Expr { "," Expr } ] ")" ;
Lambda       = "(" [ Type Ident { "," Type Ident } ] ")" "=>" ( Expr | Block ) ;
Call         = [ Qualifier "." ] Ident "(" [ Expr { "," Expr } ] ")" ;
Qualifier    = Ident | "string" ;   (* the package qualifier; see §4.2 *)
```

Notes:
- Every binary operator is left-associative. Comparison operators chain
  syntactically (`a < b < c` parses), but the type checker rejects
  comparing a `bool` with an integer.
- **[v0.1, provisional]** Bodies of `if`, `else`, `while`, and `for` must be
  braced blocks. Standalone `{ }` blocks are not statements.
- **[v0.1, provisional]** Only calls can be used as expression statements.
- **[v0.2]** `(` begins a lambda when it is followed by `)` or by a type and
  a name. Otherwise it begins a parenthesized expression.
- **[v0.1, provisional]** `++`/`--` are postfix statements only; they can't
  appear inside expressions.
- **[v0.2]** A qualifier can only be followed by a call: `string.toUpper(s)`.
  Packages export nothing but functions so far.
- **[v0.3]** `(` followed by a type keyword and `)` is a conversion, which
  binds like a unary operator: `(int) x + 1` is `((int) x) + 1`. A type
  keyword can't begin an expression, so this is unambiguous.
- **[v0.1]** A `Type` that is an identifier parses, but it is reported as
  `ASX-TYPE-010 unknown type` until user-defined types exist.

## 4. Program structure

### 4.1 Packages
- **[v0.1]** Every file begins with exactly one `package name;`. Names are
  dotted identifiers. A package is a semicolon-terminated declaration, not a
  brace-delimited block (unlike AS3).
- **[v0.3]** A package is identified by its name. Every file of the program
  that declares `package geom;` belongs to package `geom`, however it was
  loaded, so a package can span several files (§4.2).
- **[undecided]** Whether the package name must match the entry file's
  directory, and whether one file may declare several packages (no, so far).

### 4.2 Using packages and files: `use` **[v0.3]**
Syntax (decided by the owner):

```text
use "std:<name>";             // standard library package
use "<package.name>";         // RuneScript package: the .rsm files in package/name/
use "<path/to/file>";         // one file, any extension
use "<...>" as <name>;        // any of the above, with another qualifier
use "go:<package>";           // Go package                 [planned]
```

- `use` declarations come after `package` and before any function
  (`ASX-PARSE-001` otherwise). A `use` applies to the file it is in.
- **Package or file (decided).** A string containing `/` is a file path
  (`"./geom.rsc"`, `"lib/geom"`, `"../shared/x.txt"`). `std:` is the
  standard library and other `scheme:` forms are `ASX-USE-003`. Anything
  else is a package name: dot-separated identifiers that are not keywords
  (`ASX-USE-006` otherwise). Backslashes are rejected (`ASX-USE-006`);
  paths use `/` on every platform.
- **Resolution (decided).** Paths and package directories are relative to
  the *current working directory*, not to the file containing the `use`. The
  same program can therefore resolve differently depending on where it is
  launched from. The owner chose this deliberately.
- **Package directories (decided).** `use "org.x.geom";` loads every regular
  file with the extension `.rsm` directly in `org/x/geom/` (not in its
  subdirectories), in name order. Other files there are ignored. Each must
  declare `package org.x.geom;` (`ASX-USE-007`). A directory with no `.rsm`
  file is `ASX-USE-004`, as is a missing file.
- **Each file is loaded once,** however many files use it and however its
  path is spelled.
- **Qualifiers (decided).** A package's functions are called through a
  qualifier: the last segment of its name (`use "org.x.geom";` →
  `geom.area(2, 3)`), or the name after `as`. A file use binds the package
  the file declares (`use "./lib/text.txt";` in which the file says
  `package text;` → `text.shout(s)`). A type keyword can be a qualifier, so
  `string.` names package `std:string` while `string` alone is still the
  type. Qualifiers belong to the file: other files must `use` the package
  themselves.
- A qualifier that already refers to another package in the file is
  `ASX-USE-008`; using the same package twice under the same qualifier is
  `ASX-USE-002`. Using a package under two different qualifiers is allowed.
- A qualifier without a matching `use` is `ASX-NAME-007`; a function the
  package doesn't have is `ASX-NAME-008`. An unknown `std:` package is
  `ASX-USE-001`.
- **Own package.** A `use` that resolves to the package of the file it is
  in (typically a file of the same package) loads that file but binds no
  qualifier: the package's functions are already in scope.
- **Cycles (decided).** Packages may not use each other in a cycle, directly
  or indirectly (`ASX-USE-005`, reported at the `use` that closes it, with
  the cycle spelled out). Files of one package don't depend on each other.
- **[2026-10-16]** A qualified function of a RuneScript package is also a
  value of its function type: `geom.area` alone, e.g. passed to a function
  or to `dom.renderApp(app.App, "#app")`. Visibility applies as for calls.
  Standard library functions can still only be called (`ASX-NAME-004`);
  wrap them in a lambda.

### 4.3 Functions
- **[v0.2, decided]** Every function declaration, nested ones included,
  has the form `<access> <return type> <name>(<params>) { … }`. `<access>` is
  required and is one of `public`, `private`, `protected`. There is no
  `function` keyword. A missing modifier is `ASX-PARSE-001`.
- **[v0.3, decided]** `public` means usable everywhere, `private` means this
  file only, and `protected` means this package only. Using a function where
  its modifier doesn't allow it is `ASX-NAME-011`. On nested functions the
  modifier is required but has no effect.
- **[v0.3]** A package's public and protected top-level functions are
  visible, unqualified, in every file of the package, and can be called
  before their definition. Recursion is supported. Other packages call
  only the public ones, through a qualifier.
- **[v0.3]** A private top-level function is visible only in its own file.
  Two files may each have a private function of the same name; any other
  repeated top-level name in a package is `ASX-NAME-002`.
- **[v0.1]** Arguments are passed by value. Arity and argument types are
  checked statically. **[provisional]** Parameters are read-only.
- **[v0.1, provisional]** No overloading: a name declares one function.
- **[v0.2] Prototypes.** `<access> <type> <name>(<types>);` declares a
  function without a body. Parameter names are optional. The definition
  must follow in the same scope with an identical type (`ASX-TYPE-015`). A
  prototype may appear at most once, before the definition (`ASX-NAME-002`),
  and must be defined (`ASX-NAME-010`). **[v0.3]** A top-level prototype must
  be defined in the same file (`ASX-NAME-002`); other files of the package
  don't need prototypes.
- **[v0.2] Nested functions.** A function declared in a block is visible from
  its declaration (or prototype) to the end of that block, and within its own
  body. Mutual recursion between nested functions uses a prototype.
- **[v0.2, decided] Function boundaries: no capture.** A function or lambda
  body can use its own parameters and locals, plus every *function* visible
  where it is declared. It cannot use the variables or parameters of
  enclosing functions (`ASX-NAME-009`). Names of enclosing variables may be
  reused inside the boundary. `break`/`continue` don't cross boundaries.
  **[2026-10-16, decided]** One exception: runes (`Signal<T>` variables and
  parameters, §13) of enclosing functions are visible, so event handlers can
  change a component's state. They cannot be assigned across the boundary
  (only `set`), and their names cannot be reused inside it.
- **[v0.2] Function values.** A function name used without a call is a value
  of its function type (§5.6). Function values are called with ordinary call
  syntax, including chained calls `f(1)(2)`. The callee is evaluated before
  the arguments.
- **[v0.2] Lambdas.** `(T a, …) => expr` or `(T a, …) => { … }`. Parameter
  types are always written. Where a function type is expected (assignment,
  argument, return) and the parameter types match, the lambda takes that
  type's result type. Otherwise its result type is the expression's type, or
  that of the first `return` (or `void`). Lambdas take no access modifier.
- **[decided 2026-10-17]** No generic functions (see §5.1).
- **[undecided]** Default arguments, variadic parameters,
  overloading, capturing closures.

### 4.4 Entry point **[v0.1, provisional]**
A program runs by calling a top-level function `main` that takes no
parameters and returns `int`/`sint` or `void`. An `int` result becomes the
process exit code; `void` means exit code 0. **[v0.3, decided]** `main` may
be in any file of the program, but there must be exactly one
(`ASX-ENTRY-003` for the others; `ASX-ENTRY-001` if there is none). The
canonical form is:

```actionscript
package main;

protected int main() {
    return 0;
}
```

Every function needs an access modifier, so `main` has one too; any of the
three is accepted.

## 5. Types

### 5.1 Implemented types

| Type | Status | Values |
|---|---|---|
| `int` | [v0.1] | Signed integer. v0.1 representation: 32-bit (−2147483648 … 2147483647). |
| `sint` | [v0.1] | **The same type as `int`**, written differently. It has no separate representation. |
| `uint` | [v0.1, provisional width] | Unsigned integer. v0.1 representation: 32-bit (0 … 4294967295). |
| `int8`, `int16`, `int32`, `int64` | [2026-10-17] | Signed integers of exactly that many bits, e.g. `int8` is −128 … 127. They never widen. |
| `uint8`, `uint16`, `uint32`, `uint64` | [2026-10-17] | Unsigned integers of exactly that many bits, e.g. `uint64` is 0 … 18446744073709551615. |
| `bool` | [v0.1] | `true`, `false`. |
| `char` | [2026-10-16] | One Unicode scalar value: U+0000 … U+10FFFF except the surrogates U+D800 … U+DFFF. Not a number: `==`, `!=`, `<`, `<=`, `>`, `>=` compare code points; there is no arithmetic. **[provisional]** chars cannot cross the JavaScript boundary or be held in a rune. |
| `string` | [v0.1] | Immutable UTF-8 text. **[v0.2, decided]** Positions and lengths are byte offsets, as in Go. |
| `float` | [v0.3, provisional] | Generic floating point (§5.1b). Default representation float32, widening to float64. |
| `float32`, `float64` | [v0.3, provisional] | IEEE 754 binary32 and binary64. Explicit widths: they never widen. |
| `Error`, `RangeError`, `TypeError`, `ArgumentError` | [v0.3] | Error values (§7.6). The last three are subtypes of `Error`. |
| `void` | [v0.1] | Means "no value". Valid only as a return type; you can't store it, pass it, or trace it. |

There is no `any` type and never will be (CLAUDE.md §4). **[decided
2026-10-17]** There is no `null` either: every value is a real value of its
type, and absence is expressed with a `bool` (guide, chapter 17). And there
are no user-defined generics: functions and types cannot take type
parameters (`ASX-PARSE-001`); `Array<T>` and `Signal<T>` are built-in
type constructors, not a generics feature.

### 5.1a Function types **[v0.2]**
`R(P1, P2, …)` is the type of functions taking `P1, P2, …` and returning
`R` (C-style, owner's choice). Suffixes nest left to right: `int(int)(string)`
takes a `string` and returns an `int(int)`. Two function types are identical
when their parameter and result types are identical. Function values can be
stored, passed, returned, and called, but not traced, compared, or used with
operators (`ASX-TYPE-012`). `void` can be a result but not a parameter.

### 5.1b Floating-point types **[v0.3, provisional]**
Owner decision: `float` follows CLAUDE.md §4.4. It starts with a float32
representation and widens to float64 whenever a value needs it, in
magnitude **or precision**; it never contracts. Rules that follow from
this:

- Widening never changes a value, so the value of every `float` expression
  is exactly what float64 arithmetic gives. `0.1 + 0.2` is
  `0.30000000000000004`. The width a value is stored in is not observable
  in v0.3; it becomes a question of memory layout once floats can be stored
  in arrays and structs (OQ-NUM-01).
- `float32` and `float64` are exact IEEE types and never widen. Every
  `float32` result is rounded to float32.
- **IEEE 754, as in ActionScript 3 (owner decision).** A result too large
  for its type is `Infinity` (or `-Infinity`), and `0 / 0` is `NaN`. Float
  operations never stop the program. `NaN` is unequal to everything,
  itself included. The predeclared constants `NaN` and `Infinity` are
  untyped float constants (a variable or function of the same name hides
  them), and the built-ins `isNaN(x)` and `isFinite(x)` test a float of any
  float type. `-0` exists but prints as `0` and is equal to `0`.
- Operators: `+ - * /`, unary `-`, and all six comparisons, on two operands
  of the same float type. `%` is integers only. `++`/`--` are integers
  only; write `x += 1;`.
- **No implicit conversions** between integers and floats, or between float
  types: `n * 1.5` with an `int n` is `ASX-TYPE-012`, and assigning a
  `float` to a `float64` is `ASX-TYPE-001`. Use a conversion (§5.6).
- Float constants: a float literal (or `-` applied to one) takes the float
  type its context needs, or `float`. An integer constant can also become a
  float. Either must be exactly representable: `float32 x = 0.1;` is
  `ASX-TYPE-018` (write `(float32) 0.1`). Float constant expressions are
  not folded: `0.1 + 0.2` is computed at run time, in float arithmetic.
- `trace` and `(string)` print floats as ActionScript 3 prints Numbers
  (`NaN`, `Infinity`, `-Infinity` for the non-finite values):
  the shortest decimal that reads back as the same value (at float32
  precision for `float32`), plain from 1e-6 to below 1e21 (`64`, `0.5`,
  `0.000001`) and in exponent form otherwise (`1e+21`, `1.5e-7`).

### 5.1c Arrays **[decided 2026-10-16, details provisional]**
`Array<T>` is a growable list of `T`, where `T` is any value type: a
number, `bool`, `string`, `char`, a function type, an error type, or
another array (not `Element` or `Signal`).

- **Reference semantics [provisional]:** assigning or passing an array
  shares it; changes through one name are seen through every other.
- **Literals:** `[a, b, c]`. The element type is the declared type of the
  destination (`Array<uint> u = [1, 2];`); without one, it is the first
  element's type, and every other element must be assignable to it. An
  empty `[]` needs a destination type.
- `a[i]` reads element `i` (an `int` or `uint`); `a[i] = v` and
  `a[i] op= v` write it. `a`, then `i`, are evaluated once. An index
  outside `0 … a.length - 1` throws a `RangeError` (`ASX-RUN-011`).
  For `a[i] op= v`, the element is read before `v` is evaluated, and the
  index is checked again when storing.
- `a.length` is the number of elements (`int`). `a.push(v)` appends `v`;
  `a.pop()` removes and returns the last element (`ASX-RUN-011` when
  empty).
- Arrays cannot be compared, traced, converted, interpolated, held in a
  rune, or passed across the JavaScript boundary **[provisional]**. Use
  `string.join` to show them.
- **[undecided]** for-each loops, slicing, sorting, `a[i]++`, `Array<T>`
  for user generic types, and whether `array` (reserved) means the same.

### 5.1d Explicit-width integers **[decided 2026-10-17, details provisional]**
- Every integer type is a distinct type. `int32` is **not** `int`: the
  generic `int` may widen one day (CLAUDE.md §4.4), an `int32` never does.
  `sint` remains a spelling of `int`.
- Nothing converts implicitly between integer types, not even to a wider
  one: `int16 b = a;` with `a` an `int8` is `ASX-TYPE-001`, and `a + b` with
  different integer types is `ASX-TYPE-012` (`ASX-TYPE-011` if the
  signedness differs). Write `(int16) a`. **[provisional]** Implicit
  lossless widening could be allowed later without breaking programs.
- Integer constants take the destination's type when they fit
  (`uint64 h = 18446744073709551615;`), as in §5.3.
- Arithmetic is exact and checked in every width: a result outside the
  type's range is `ASX-RUN-002`, never a wrapped value.
- On the JavaScript target, `int64` and `uint64` are `BigInt`s; the other
  widths are numbers. **[provisional]** Explicit widths cannot cross the
  `extern` boundary yet (`ASX-TYPE-023`).

### 5.1e Structs **[decided 2026-10-17, details provisional]**
```ebnf
StructDecl = Visibility "struct" Ident "{" { Type Ident ";" } "}" ;
StructLit  = [ Ident "." ] Ident "{" [ Ident ":" Expr { "," Ident ":" Expr } ] "}" ;
```
- A struct is a **value type**: assigning, passing, returning, or storing
  one copies it, so a change through one name is never seen through
  another. (Implementations share unchanged structs and copy on write;
  that is not observable.)
- It is declared at the top level with an access modifier, like a
  function: `private` (this file), `protected` (this package), `public`.
  Another package's struct is named with its qualifier: `geom.Point`.
- Each struct declaration is its own type, even if another has the same
  fields. Field types are any value type; a struct may not contain itself
  by value, directly or through other structs (`ASX-TYPE-027`), but may
  through an `Array`.
- A literal gives **every field exactly once**, in any order; there are
  no default values (`ASX-TYPE-026`). Fields are evaluated in the order
  written.
- `p.x` reads a field. `p.x = v` and `p.x op= v` write one where `p` is a
  variable, an array element, or a field of either (`shapes[i].origin.x`).
  A field of a parameter is read-only, like the parameter
  (`ASX-NAME-006`), and so is a field of a temporary value such as `f().x`.
- `==` and `!=` compare every field; a struct with an array, function, or
  error field cannot be compared. `trace` shows `Point{x: 1, name: "a"}`,
  with strings and chars quoted.
- Arrays held in a struct stay references: copying the struct shares the
  array.
- **[undecided]** Methods, constructors, field visibility, and passing
  structs across the JavaScript boundary.

### 5.2 Planned types **[planned/undecided]**
`matrix` and `class`.
None of these are implemented. Their names are reserved.

### 5.3 Integer constants **[v0.1, provisional]**
Integer literals, and expressions built only from literals using
`+ - * /  %` and unary `-`, are **untyped constants**. They are evaluated
exactly at compile time and then take the type their context requires:

- the declared type of the variable, parameter, or return value they're
  assigned to;
- the type of the other operand of a binary operator;
- `int`, when there's no other context (e.g. `trace(5)`, `1 < 2`).

The constant must fit the target type, otherwise it's an error:
`ASX-TYPE-006` for out of range, `ASX-TYPE-007` for a negative value going
to an unsigned type. Dividing by a constant zero is `ASX-TYPE-014`.

### 5.4 Widening **[undecided]**
CLAUDE.md §4.4 says generic numeric types widen when needed and never
contract. The widening model has not been chosen yet, so v0.1 implements
**no widening**:

- **[v0.1, provisional]** Integer arithmetic is exact. A result outside the
  operand type's v0.1 range stops the program with `ASX-RUN-002`. It never
  wraps. When widening is specified, programs that run today will produce
  the same results.

### 5.5 Assignability **[v0.1]**
A value of type `S` can be stored in, passed as, or returned as a
destination of type `D` only if:

| S → D | Result |
|---|---|
| same type (`int`/`sint` count as the same; function types structurally) | allowed |
| an error type → a supertype (`RangeError` → `Error`) [v0.3] | allowed |
| untyped constant → integer type | allowed if the value fits (§5.3) |
| untyped constant → float type [v0.3] | allowed if the value is exactly representable (§5.1b) |
| signed (`int`/`sint`) → `uint` | **error `ASX-TYPE-004`**, whatever the runtime value |
| `uint` → `int`/`sint` | **error `ASX-TYPE-005`** [provisional: rules undecided] |
| `void` → anything | error `ASX-TYPE-002` |
| anything else | error `ASX-TYPE-001` |

There are no implicit conversions between numbers, `bool`, and `string`.

### 5.6 Conversions **[v0.3, decided syntax]**
`(T) x` converts `x` to `T`, C style (owner decision; this settles
OQ-CONV-01). `T` is any number type (`int`, `sint`, `uint`, the explicit
widths `int8` … `uint64`, `float`, `float32`, `float64`), `string`, `bool`,
or `char`. Conversions are **checked**: they never wrap, and a
value that can't be converted throws instead (`ASX-RUN-006` if nothing
catches it).

| From → to | Rule | Failure |
|---|---|---|
| integer → integer | the same value | out of range: `RangeError` |
| float → integer | truncates toward zero: `(int) -64.7` is `-64` | out of range, `NaN`, or infinite: `RangeError` |
| integer → float | the nearest value of the float type | never |
| float → float | the nearest value; exact into `float64`/`float`; beyond float32's range is `Infinity` | never |
| number, `bool` → `string` | as `trace` prints it | never |
| `string` → integer | an optional sign and decimal digits, nothing else: `"-42"` | not of that form: `TypeError`; out of range: `RangeError` |
| `string` → float | sign, digits, optional fraction and exponent (`"64.7"`, `"1e5"`), or `"NaN"`, `"Infinity"`, `"-Infinity"`; parsed to the nearest float64 (`Infinity` if too large), then rounded for `float32` | not of that form: `TypeError` |
| `string` → `bool` | exactly `"true"` or `"false"` | otherwise: `TypeError` |
| `char` → integer | its code point: `(int) 'a'` is `97` | never |
| integer → `char` | the character with that code point | not a Unicode scalar value: `RangeError` |
| `char` → `string` | a one-character string | never |
| `string` → `char` | its one character | not exactly one character: `TypeError` |
| `char` ↔ float, `bool` | not allowed: `ASX-TYPE-017` | |
| `bool` ↔ number, errors, functions | not allowed: `ASX-TYPE-017` | |

A conversion of a constant happens at compile time, and its failure is a
compile-time error: `(uint) -2.5` is `ASX-TYPE-007`, `(int) 3e10` is
`ASX-TYPE-006`. Unlike an implicit conversion, an explicit one may round:
`(float32) 0.1` is allowed.

## 6. Expressions

### 6.1 Operators **[v0.1]**

| Operator | Operands | Result |
|---|---|---|
| `+ - * / %` | two integers of the same signedness | the same integer type |
| `+ - * /` [v0.3] | two floats of the same type | that type (§5.1b) |
| `+` [2026-10-16] | two `string`s | `string`: the two joined |
| `& \| ^` [2026-10-17] | two integers of the same type | that type: bitwise and, or, xor |
| `~` [2026-10-17] | an integer | that type: the bitwise complement in the type's width |
| `<< >>` [2026-10-17] | an integer and an integer count of any type | the left operand's type |
| unary `-` | a signed integer or a float | the same type (`uint` is rejected) |
| `< <= > >=` | two integers of the same signedness, two floats of the same type, or two `char`s | `bool` |
| `== !=` | two numbers as above, two `bool`s, two `char`s, or two `string`s | `bool` |
| `&& \|\|` | `bool` | `bool`; short-circuit evaluation |
| `!` | `bool` | `bool` |

- **[v0.1, provisional]** Mixing signed and unsigned operands is rejected
  (`ASX-TYPE-011`).
- **[v0.1, provisional]** `/` truncates toward zero. `%` takes the sign of
  the dividend. Division by zero is a runtime error (`ASX-RUN-003`).
- **[decided 2026-10-16]** `+` joins two strings and converts nothing:
  `"n=" + 5` is `ASX-TYPE-012`; write `"n=" + (string) 5` or `$"n={5}"`.
  `s += t` works on a `string` variable.
- **[decided 2026-10-17, details provisional]** Shifts: the count must be
  from 0 to the left operand's width minus 1 (`ASX-TYPE-006` for a
  constant, otherwise a `RangeError`, `ASX-RUN-012`). `>>` is arithmetic on
  signed types and logical on unsigned ones. **[provisional]** `<<` is
  checked like `*`: a result that doesn't fit the type is an overflow
  (`ASX-RUN-002`), never silently truncated; mask first to drop bits on
  purpose. A constant shifted by a variable count is an `int`. Constants
  are folded exactly: `(int64) (1 << 40)`.
- **[provisional]** Precedence follows C and ActionScript 3 (below). Since
  `x & mask == 0` is then `x & (mask == 0)`, it is a type error, with a
  hint to write `(x & mask) == 0`.
- `&= |= ^= <<= >>=` work like the other compound assignments.
- **[undecided]** `?:`, `?.`, `??`.
- Operands are evaluated left to right.

### 6.2 Calls **[v0.1]**
`f(a, b)` evaluates arguments left to right, then calls `f`. A call to a
`void` function can only be a statement.

## 7. Statements

### 7.1 Variables **[v0.1]**
`Type name = expr;` declares a variable in the current block.
- **[provisional]** An initializer is required.
- **[provisional]** Redeclaring or shadowing any visible variable,
  parameter, or function name is rejected (`ASX-NAME-002`).
- Scope is the block. A variable declared in a `for` initializer is scoped
  to the loop.

### 7.2 Assignment **[v0.1]**
`x = e;` requires `e` to be assignable to `x`'s declared type (§5.5).
`x op= e` is checked and executed as `x = x op e`. `x++` and `x--` require
an integer variable. Assignment is a statement, not an expression.

### 7.3 Conditionals **[v0.1]**
`if (c) { … } else if (c2) { … } else { … }`. Every condition must have type
`bool` (`ASX-TYPE-013`); integers are not truthy **[provisional]**.

### 7.4 Loops **[v0.1]**
- `while (c) { … }`
- `for (init; cond; post) { … }`. Each clause is optional, and an omitted
  `cond` means "forever". `init` may be a declaration, assignment,
  increment, or call; `post` is the same minus declarations.
- `break;` and `continue;` apply to the innermost loop. Using them outside a
  loop is an error (`ASX-FLOW-001`). In a `for` loop, `continue` runs `post`
  next.

### 7.5 Return **[v0.1]**
`return e;` in a non-void function; `return;` in a `void` function. A
non-void function must return on every path (`ASX-TYPE-008`). A body
returns on every path when it ends in `return`, or in an
`if`/`else if`/`else` chain whose every branch does. **[v0.3]** `throw`
counts like `return`, and so does a `try` statement whose block and every
catch clause end that way, or whose `finally` block does. Loops are never
treated as guaranteeing a return **[provisional]**.

### 7.6 Errors: `throw` and `try` **[v0.3]**
Modelled on ActionScript 3 (owner decision).

- **Error values.** `new Error(message, errorID)` creates an error; both
  arguments are optional (`""` and `0`). `RangeError`, `TypeError`, and
  `ArgumentError` are built-in subtypes of `Error`, created the same way.
  An error has the read-only fields `message` (`string`), `name`
  (`string`, the type's name), and `errorID` (`int`). Errors are values:
  they can be stored, passed, returned, and traced (`trace(e)` prints
  `RangeError: message`, or just the name if the message is empty), but not
  compared. `new` can create only error types until classes exist
  (`ASX-TYPE-019`).
- **`throw e;`** throws an error value (anything else is `ASX-TYPE-021`). A
  throw ends a path for the definite-return rule (§7.5).
- **`try { } catch (T e) { } … finally { }`** runs the block. If it throws,
  the first catch clause whose type the error has (the type or a subtype)
  runs, with the error in its read-only variable. A clause that an earlier
  one makes unreachable is `ASX-TYPE-022`. If no clause matches, the error
  goes on to the enclosing try statements and calling functions. The
  `finally` block runs however the statement ends: normally, by `return`,
  `break`, or `continue`, or by an error, which then goes on.
- **A `finally` block may `return`, `break`, or `continue`** (owner
  decision, as in ActionScript 3 and Java). That replaces however the try
  statement was ending: a `return` value from the block, or an error in
  flight, which is discarded. A `finally` block that throws replaces the
  error in flight.
- **`catch (e)`** (owner decision) has no type and catches every error;
  it means `catch (Error e)`. `catch (Error)` without a name is
  `ASX-PARSE-001`.
- **Runtime errors are catchable** (owner decision), except the call-depth
  limit: integer overflow and integer division by zero throw a
  `RangeError`, a failed conversion a `RangeError` or `TypeError`, and a
  failed standard library call an `Error`. Their `errorID` is their code's
  number (2 for `ASX-RUN-002`), and their message is the runtime error's.
- **Uncaught errors.** A runtime error that nothing catches is reported
  with its own code, message, and location, even if it was caught and
  thrown again. An error created with `new` that nothing catches is
  `ASX-RUN-005` ("uncaught RangeError: message") at the `throw` that first
  threw it.
- **The call-depth limit (`ASX-RUN-001`) can't be caught.** While the
  program stops for it, catch clauses don't run and neither do `finally`
  blocks.
- **[undecided]** Stack traces (`getStackTrace()`), methods such as
  `toString()`, and user-defined error types (with classes).

## 8. Built-ins

### 8.1 `trace(…)` **[v0.1, provisional formatting]**
Takes any number of arguments of any non-void type. It writes their text
forms separated by one space, followed by a newline, to standard output.
Integers print in decimal, booleans as `true`/`false`, and strings as-is.
An integer constant argument is typed `int`.

```actionscript
trace("x =", 42, true);   // x = 42 true
```

### 8.2 Standard library **[v0.2]**
Standard library packages are implemented natively and brought in with
`use "std:<name>";`. Their function names are Go's names converted to
camelCase **[decided]**. Integer results are range-checked against their
declared RuneScript type. A failure inside a library function is a runtime
error (`ASX-RUN-004`).

**[2026-10-17]** A standard library name may be **overloaded**: `math.max`
takes two `int`s or two `float64`s. A call uses the first version whose
parameters accept every argument: an integer constant fits an integer or a
float parameter, a float constant only a float one, and a typed value only
its exact type. So `math.max(3, 7)` is an `int`, `math.max(1.5, 2)` a
`float64`, and `math.max(x, n)` with a `float64` and an `int` is
`ASX-TYPE-001`. User functions cannot be overloaded.

Library functions on floats give bit-identical results on both targets
and on every processor: the JavaScript runtime computes exactly what Go
computes, never an approximation from the host. The transcendental
functions run one shared algorithm, Go's pure-Go code (`internal/fmath`),
with every multiply-add kept unfused, and a line-by-line JavaScript port
of it. A NaN's sign is not observable:
`math.signbit(NaN)` is `false`.

| Package | Contents |
|---|---|
| `std:string` | 43 functions: Go's `strings` functions expressible with `string`, `int`, `bool`, `char`, callbacks, and `Array<string>`, plus `len`, `runeCount`, and `slice` (see `docs/guide/10-std-string.md`) |
| `std:math` | [2026-10-17] Go's `math` on `float64`: `abs ceil floor trunc round roundToEven sqrt min max mod remainder dim nextafter copysign signbit isInf inf nan ldexp`, and `sin cos tan asin acos atan atan2 sinh cosh tanh asinh acosh atanh exp exp2 expm1 log log10 log2 log1p cbrt pow hypot`; the int helpers `abs min max clamp`; untyped constants `pi e phi sqrt2 sqrtE sqrtPi sqrtPhi ln2 log2E ln10 log10E maxFloat64 smallestNonzeroFloat64 maxFloat32 smallestNonzeroFloat32` and `minInt/maxInt` … `maxUint64` for every integer type. |
| `std:symrs` | [planned] Symbolic mathematics, a port of the owner's SympJS library; needs objects, floats, and arrays |

## 9. Diagnostics and runtime errors

Format:

```text
file.rsc:LINE:COL: error CODE: message
    <source line>
    <caret underline>
  help: <suggestion>
```

The `ASX` code prefix is **provisional**. Each code is
`ASX-<STAGE>-<NNN>`, where the stages are `LEX`, `PARSE`, `NAME`, `TYPE`,
`FLOW`, `USE`, `ENTRY`, `TARGET`, `RUN`, and `WARN`. The full list is in `docs/guide/08-diagnostics.md`.

**[decided 2026-10-16]** Warnings (`ASX-WARN-…`) report code that is legal
but probably wrong: unused locals, parameters, and uses; unreachable code;
discarded results; externs from one host's own module. They are reported
only for programs without errors, never stop a program, and are controlled
with `-Werror` (every warning becomes an error) and `--allow <codes>`.

The checker reports every semantic error it finds. The parser stops at the
first syntax error. A runtime error that nothing catches (§7.6) ends the
program with exit status 1 and reports the source location where it
happened. **[v0.3]** The runtime errors `ASX-RUN-002` to `ASX-RUN-004` and
`ASX-RUN-006` to `ASX-RUN-008` are thrown as error values and can be caught; exceeding the
call depth (`ASX-RUN-001`) cannot be caught.

## 10. Comparison with ActionScript 3

RuneScript keeps AS3's names and feel (`package`, `trace`, `int`/`uint`,
`public`/`protected`/`private`), but its syntax and typing differ on
purpose:

| Topic | ActionScript 3 | RuneScript |
|---|---|---|
| Package | `package a.b { … }` block | `package a.b;` statement |
| Function | `function add(a:int, b:int):int { }` | `public int add(int a, int b) { }` (access modifier required) |
| Function values | `var f:Function = add;` (untyped) | `int(int, int) f = add;` (typed) |
| Anonymous functions | `function(a:int):int { … }`, Royale `=>` | `(int a) => a * 2` |
| Closures | Capture enclosing variables | No capture: nested functions and lambdas see only their own variables, other functions, and (rune:dom) the runes of enclosing functions |
| Variable | `var n:int = 0;` | `int n = 0;` |
| Semicolons | Optional | Required |
| Untyped values | `*` and missing annotations allowed | No `any`; every declaration is explicitly typed |
| Top-level code | One externally visible definition per file; no `main` | Many functions per file; `main()` entry point |
| `int`/`uint` mixing | Implicit coercion; `var u:uint = -1` wraps to 4294967295 | Compile errors (`ASX-TYPE-004/005/007/011`) |
| Integer overflow | Wraps | Never wraps silently (v0.1: runtime error; widening planned) |
| Integer division | Produces `Number` (`7/2 == 3.5`) | Truncating integer division |
| Floating point | `Number` (64-bit), with `NaN` and `Infinity` | `float` (float32 by default, widening exactly), `float32`, `float64`; IEEE, with `NaN` and `Infinity`, as in AS3 [v0.3] |
| Type names | `String`, `Boolean` | `string`, `bool`; `sint` alias added |
| Truthiness | `if (count)` allowed | Conditions must be `bool` |
| Variable scope | `var` is function-scoped and hoisted | Block-scoped; no shadowing [provisional] |
| Visibility | `public protected private internal`; default `internal` | `public` (everywhere), `protected` (package), `private` (file); always written |
| Casts | `T(x)` and `x as T`; `int("x")` is `0` | `(T) x`, checked: `(int) "x"` throws a `TypeError` [v0.3] |
| Exceptions | `throw` any value; `catch (e:Error)`, `catch (e)` | `throw` error values only; `catch (Error e)`, `catch (e)`; runtime errors are catchable [v0.3] |
| Imports | `import a.b.C;` imports one definition | `use "a.b";` and `use "./file";` make a whole package callable as `b.f()` [v0.3]; `use "go:pkg";` [planned] |
| `trace` | `trace(...args)`, space-separated | Same, written to stdout |

**Games and graphics.** AS3 was used mostly for games and interactive
media in Flash Player. RuneScript plans native (non-browser) graphics, input,
and UI as a standard package brought in with `use` **[planned]**. The
recommended backend is raylib; this is not yet confirmed.

AS3 and Apache Royale features that are **[undecided]** for RuneScript:
classes and interfaces, `static`, abstract classes, private constructors,
metadata, E4X XML, `Vector.<T>` (vs. `Array<T>` generics), type inference,
`?.`, `??`, import aliases, and verbatim strings.

## 11. Conflicts with CLAUDE.md

1. **First-milestone scope.** CLAUDE.md §10 lists literals, declarations,
   calls, and `return`. The owner later widened v0.1 to include
   conditionals, loops, assignment, and basic functions, and this document
   follows the owner. CLAUDE.md §10 should be updated.
2. **Open-questions file.** CLAUDE.md §6 and §9 list
   `docs/OPEN_QUESTIONS.md` as a repository file. The owner has asked that
   it stay local and uncommitted.

## 12. JavaScript target **[v0.2]**

`runescript build [--target js] [-o out.js] file.rsc` compiles a program to
a single ES module (owner decision). **[decided 2026-10-16]** The module
uses only ES2015 plus `BigInt`, `??`, and ES modules (`import`/`export`,
`import.meta`), so that a small bundled JavaScript runtime can run it
(planned; Node.js, ES2022, stays the reference host). `go test ./tests`
checks the generated code for newer syntax and built-ins. **[v0.3]** Every file of the
program is compiled into that one module, and runtime errors name the file
they happened in, as the interpreter's do. The module exports
`run(options)`, which calls `main` and returns the exit code, and
`RuneScriptError`. Run directly with Node.js (`node out.js`), the module
calls `run` itself.

**Semantics are the language's, not JavaScript's.** A program that checks
behaves identically on the interpreter and on the JavaScript target: same
output, same exit code, same runtime errors (code, message, position).
The one exception is listed under host limits below. **[2026-10-16]**
Programs with extern functions (§12.1) run only on the JavaScript target.

- **[v0.2]** Static safety: code is generated from a typed intermediate
  representation (`internal/ir`) in which every expression has a concrete
  type. The IR is re-verified independently of the checker before code is
  generated. Types are then erased: no value of the wrong type can reach a
  variable, parameter, or return, so no per-value tag checks are needed.
- **[v0.2]** Dynamic safety: wherever JavaScript would differ, the
  generated code calls the runtime instead of using a JavaScript operator.
  Integer `+ - * / %` is range-checked (§5.4, OQ-NUM-01), `/` truncates,
  division by zero traps, results are normalized (no `-0`), and call depth
  is counted (§9).
- **[v0.2, decided]** `std:string` keeps UTF-8 byte offsets (§5.1) on the
  JavaScript target. Case mapping, case folding, and white space use
  tables generated from Go's `unicode` package.
- **[v0.2, decided]** Widening (OQ-NUM-01) is still undecided. The
  JavaScript target traps on overflow exactly like the interpreter. If
  64-bit widening is adopted, the JavaScript target will need `BigInt`.
- **[v0.2] Host limits.** If the JavaScript host's stack runs out before
  the 10,000-call limit, the program stops with `ASX-RUN-001` at a call
  site inside the recursion. That site may differ from the interpreter's.
  Engine string-length limits surface as `ASX-RUN-004` inside `std:string`.
- **[decided 2026-10-16] Compatibility.** Full JavaScript compatibility is
  **not** promised yet. What is promised is strictness: no value enters a
  RuneScript program from JavaScript unchecked. TypeScript trusts its
  declarations and erases them; RuneScript checks them at run time.

### 12.1 Extern functions **[decided 2026-10-16]**

`extern "module" Type name(Type p, …);` declares a function implemented by
a JavaScript module. Rules:

- Top level only; no access modifier; visible only in the declaring file
  (provisional). Parameter names are optional. The name may not clash with
  a function of the file or package, and is never shadowed.
- An extern function can only be called (`ASX-NAME-004` otherwise).
- Parameter and result types: integer, float, `bool`, `string`, and error
  types; `void` as the result. Anything else is `ASX-TYPE-023`.
- `rune:<name>` modules are host-neutral and part of RuneScript; their
  functions are fixed and a declaration must match one exactly
  (`ASX-USE-009`, `ASX-TYPE-024`). `rune:fs` provides `readFile`,
  `writeFile`, `appendFile`, `exists`, and `remove` (guide chapter 11).
  Any other module string is passed to JavaScript's `import` unchanged;
  `node:`, `bun:`, and `deno:` modules give warning `ASX-WARN-006`.
- Arguments cross unchanged; error values cross as `RSError` objects, a
  class the module exports so JavaScript can create, return, and throw
  RuneScript errors.
- Results are checked: integers must be integral and in range (`-0` becomes
  `0`); `float`/`float64` accept any number; `float32` accepts any number
  and rounds it with `Math.fround` (owner decision); `bool` needs a
  boolean; `string` needs a string without unpaired surrogates; `void`
  needs `undefined`; an error type needs a valid `RSError` of that type or
  a subtype, or a JavaScript `Error` (converted: `RangeError` and
  `TypeError` keep their names, others become `Error`, `errorID` 0). A
  mismatch throws a `TypeError` (`ASX-RUN-007`) at the call.
- A JavaScript exception becomes an `Error` (`ASX-RUN-008`); an `RSError`
  that JavaScript throws crosses unchanged once checked.
- The interpreter does not run extern functions: `runescript run` reports
  `ASX-TARGET-002`, while `check` and `build` accept the program. A bundled
  JavaScript runtime is planned (Node.js, ES2022, stays the reference).
- **[planned]** Source maps.

## 13. Web pages: `rune:dom` **[decided 2026-10-16]**

`use "rune:dom";` provides components, templates, and runes (owner
decisions; see the guide, chapter 15). Programs that use it run only on the
JavaScript target (`ASX-TARGET-002` in the interpreter).

### 13.1 Templates

```ebnf
TemplateStmt = "template" "{" Content "}" ;      (* exactly one root Element *)
Content      = { Element | Text | "{{" Expr "}}" | If } ;
Element      = "<" Name { Attribute } ( "/>" | ">" Content "</" Name ">" ) ;
Attribute    = Name [ "=" ( Quoted | "{{" Expr "}}" ) ] ;
If           = "$if" "(" Expr ")" Content { "$elseif" "(" Expr ")" Content }
               [ "$else" Content ] "$endif" ;
```

- A component is a function returning `Element` whose last statement is
  its one `TemplateStmt` (`ASX-TMPL-001`); the element is its result.
- HTML void elements need no closing tag. Inside `<svg>`, names are SVG's
  (case-sensitive) and every element may be self-closed; inside
  `<foreignObject>`, they are HTML again.
- Text: character references are decoded; white space containing a line
  break is removed at the ends of a text and collapsed to one space inside
  it; `{` and `}` are reserved (a literal brace is `{{ "{" }}`).
  `<!-- comments -->` are dropped. White space around the root is ignored.
- Every element and attribute must be in the internal HTML4/HTML5/SVG list
  (`ASX-TMPL-002`, `ASX-TMPL-003`); obsolete ones warn (`ASX-WARN-007`).
  Static values must suit the attribute's kind (`ASX-TMPL-004`); `{{ }}`
  values must have the kind's type: `bool`, `int`, `uint`, a float type, a
  number or string for SVG lengths, `void()` for event handlers, and
  `string` otherwise (`ASX-TMPL-005`). Repeated attributes are
  `ASX-TMPL-006`.
- `{{ e }}` in content shows a string, number, or `bool` as text (always
  escaped), or inserts an `Element` (a child component).
- Every `{{ }}`, bound attribute, and `$if` condition is a binding: it is
  computed again when a rune it read changes, and only its own part of the
  DOM is updated. `$if` rebuilds a branch only when the chosen branch
  changes; a removed branch's bindings are disposed.
- At run time, a bound keyword attribute with a value outside its list, and
  a bound URL that would run script (`javascript:`, `vbscript:`, `data:`),
  throw a `TypeError` (`ASX-RUN-010`). Static URLs of that kind are
  `ASX-TMPL-004`.

### 13.2 Runes

- `Signal<T>` is a built-in type for a number, `bool`, or `string` `T`
  (`ASX-TYPE-025` otherwise). There are no user generics (§5.1).
- `signal(v)` creates one; with a declared `Signal<T>`, `v` must be
  assignable to `T`, otherwise `T` is `v`'s type.
- `r()` and `r.get()` read a rune; `r.set(v)` stores `v`, which must be
  assignable to `T`, and runs the bindings that read it. Setting the same
  value does nothing.
- Nested functions and lambdas may use the runes of enclosing functions
  (§4.3).

### 13.3 Mounting

`dom.renderApp(component, selector)` takes an `Element()` function and a CSS
selector. It calls the component and replaces the content of the first
element the selector matches with the result. A literal selector that is
not a single id selector gives `ASX-WARN-008`; no match at run time throws
an `Error` (`ASX-RUN-009`). There is no router.

### 13.4 Not yet specified

`$foreach` (needs arrays), component tags with properties, event objects
and two-way binding, raw HTML, MathML, custom elements, server-side
rendering, and hot module replacement.
