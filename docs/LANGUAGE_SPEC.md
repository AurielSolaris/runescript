# RuneScript Language Specification

**Version 0.1 (draft).** This is the formal companion to `CLAUDE.md`. If
the two disagree, the disagreement is flagged in §11 rather than silently
resolved.

Every rule carries one of these labels:

| Label | Meaning |
|---|---|
| **[v0.1]** | Implemented and covered by tests. |
| **[v0.1, provisional]** | Implemented, but the owner has not decided the rule. It is the most conservative choice and may change. |
| **[planned]** | Required by the design; not implemented. |
| **[undecided]** | Not yet designed. Programs using it are rejected. |

RuneScript is inspired by ActionScript 3 but is **not** source-compatible
with it, and no compatibility is promised. Where RuneScript's design differs
from AS3, RuneScript's design wins. §10 summarises the differences.

---

## 1. Source text

- **[v0.1]** Source files are UTF-8.
- **[v0.1, provisional]** The extension is undecided. `.rs` is the candidate,
  but it is also Rust's extension. The CLI accepts any file name.
- **[v0.1]** Positions in diagnostics are `line:column`, both 1-based.
  Columns count Unicode code points.

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
package return if else while for break continue true false
public protected private
int sint uint bool string void
```

Reserved for planned features **[v0.1, provisional]**. They are rejected
with `ASX-PARSE-002`:

```
class struct use float char array matrix
int8 int16 int32 int64 uint8 uint16 uint32 uint64 float32 float64
```

`trace` is not a keyword. It is a built-in function name and can't be
redeclared.

### 2.4 Literals
- **Integer [v0.1]**: decimal digits. `0` is valid, but other numbers with
  leading zeros (`007`) are rejected **[provisional]**. Hex, binary, digit
  separators, and suffixes are **[undecided]**. Literals are non-negative;
  `-5` is unary minus applied to `5`.
- **String [v0.1]**: `"…"` on one line. Escapes: `\n \t \r \\ \"`. Any other
  escape is an error **[provisional]**.
- **Boolean [v0.1]**: `true`, `false`.
- **Float, char [undecided]**: `1.5` and `'a'` are rejected with
  `ASX-LEX-006`.

### 2.5 Operators and punctuation **[v0.1]**
```
( ) { } , ; .
=  +=  -=  *=  /=  %=  ++  --
+  -  *  /  %
== != <  <= >  >=
&& || !
```

## 3. Grammar (v0.1)

EBNF. `{ x }` means zero or more, `[ x ]` means optional.

```ebnf
File         = PackageDecl { FuncDecl } EOF ;
PackageDecl  = "package" Ident { "." Ident } ";" ;

FuncDecl     = [ Visibility ] Type Ident "(" [ Params ] ")" Block ;
Visibility   = "public" | "protected" | "private" ;
Params       = Param { "," Param } ;
Param        = Type Ident ;
Type         = "int" | "sint" | "uint" | "bool" | "string" | "void" | Ident ;

Block        = "{" { Stmt } "}" ;
Stmt         = SimpleStmt ";"
             | "return" [ Expr ] ";"
             | "break" ";" | "continue" ";"
             | IfStmt | WhileStmt | ForStmt ;
SimpleStmt   = VarDecl | Assignment | IncDec | Call ;
VarDecl      = Type Ident "=" Expr ;
Assignment   = Ident ( "=" | "+=" | "-=" | "*=" | "/=" | "%=" ) Expr ;
IncDec       = Ident ( "++" | "--" ) ;

IfStmt       = "if" "(" Expr ")" Block [ "else" ( IfStmt | Block ) ] ;
WhileStmt    = "while" "(" Expr ")" Block ;
ForStmt      = "for" "(" [ SimpleStmt ] ";" [ Expr ] ";" [ ForPost ] ")" Block ;
ForPost      = Assignment | IncDec | Call ;

Expr         = OrExpr ;
OrExpr       = AndExpr { "||" AndExpr } ;
AndExpr      = EqExpr  { "&&" EqExpr } ;
EqExpr       = RelExpr { ( "==" | "!=" ) RelExpr } ;
RelExpr      = AddExpr { ( "<" | "<=" | ">" | ">=" ) AddExpr } ;
AddExpr      = MulExpr { ( "+" | "-" ) MulExpr } ;
MulExpr      = Unary   { ( "*" | "/" | "%" ) Unary } ;
Unary        = ( "-" | "!" ) Unary | Primary ;
Primary      = IntLit | StringLit | "true" | "false"
             | Ident | Call | "(" Expr ")" ;
Call         = Ident "(" [ Expr { "," Expr } ] ")" ;
```

Notes:
- Every binary operator is left-associative. Comparison operators chain
  syntactically (`a < b < c` parses), but the type checker rejects
  comparing a `bool` with an integer.
- **[v0.1, provisional]** Bodies of `if`, `else`, `while`, and `for` must be
  braced blocks. Standalone `{ }` blocks are not statements.
- **[v0.1, provisional]** Only calls can be used as expression statements.
- **[v0.1, provisional]** `++`/`--` are postfix statements only; they can't
  appear inside expressions.
- **[v0.1]** A `Type` that is an identifier parses, but it is reported as
  `ASX-TYPE-010 unknown type` until user-defined types exist.

## 4. Program structure

### 4.1 Packages
- **[v0.1]** Every file begins with exactly one `package name;`. Names are
  dotted identifiers. A package is a semicolon-terminated declaration, not a
  brace-delimited block (unlike AS3).
- **[undecided]** How packages map to directories, and multi-file packages.

### 4.2 Using other packages and files **[planned; syntax decided]**
The owner has chosen this syntax for bringing in another package or source
file:

```actionscript
use "<package_name>";        // e.g. use "rune.gfx";
use "<file_path>";           // path relative to the current working directory
```

`use` is reserved in v0.1, and writing it gives `ASX-PARSE-002`. Still
**[undecided]**:
- how the compiler tells a package name from a file path (e.g. a path
  contains `/` or ends in the source extension);
- where `use` may appear (presumably after `package`, before any function);
- how a used package's names are referred to (qualified as `gfx.drawText`,
  or brought directly into scope);
- what visibility `use` exposes, and whether cycles are allowed.

Resolving file paths against the *current working directory* means the
same program can resolve differently depending on where it is launched
from. The owner chose this deliberately, and it is recorded here so that
tools (CLI, tests, editor support) apply it consistently.

### 4.3 Functions
- **[v0.1]** A function declaration is written return type, name, then typed
  parameters: `int add(int a, int b) { … }`. There is no `function`
  keyword.
- **[v0.1]** Functions are top-level. A file may declare any number of them,
  and a function can be called before its declaration. Recursion is
  supported.
- **[v0.1]** Arguments are passed by value. Arity and argument types are
  checked statically.
- **[v0.1, provisional]** Parameters are read-only.
- **[v0.1, provisional]** Function names must be unique: there is no
  overloading.
- **[v0.1, provisional]** Visibility modifiers are parsed and recorded, but
  have no effect yet. At most one modifier is allowed. Their meaning on
  top-level functions and the default visibility are **[undecided]**.
- **[undecided]** Default arguments, return-type inference, nested
  functions, and function values.

### 4.4 Entry point **[v0.1, provisional]**
A program runs by calling a top-level function `main` that takes no
parameters and returns `int`/`sint` or `void`. An `int` result becomes the
process exit code; `void` means exit code 0. Any visibility modifier is
accepted on `main`; the canonical form is:

```actionscript
package main;

protected int main() {
    return 0;
}
```

## 5. Types

### 5.1 Implemented types

| Type | Status | Values |
|---|---|---|
| `int` | [v0.1] | Signed integer. v0.1 representation: 32-bit (−2147483648 … 2147483647). |
| `sint` | [v0.1] | **The same type as `int`**, written differently. It has no separate representation. |
| `uint` | [v0.1, provisional width] | Unsigned integer. v0.1 representation: 32-bit (0 … 4294967295). |
| `bool` | [v0.1] | `true`, `false`. |
| `string` | [v0.1] | Immutable text. |
| `void` | [v0.1] | Means "no value". Valid only as a return type; you can't store it, pass it, or trace it. |

There is no `any` type and never will be (CLAUDE.md §4).

### 5.2 Planned types **[planned/undecided]**
`float` (default float32, widening), `char`, `int8 int16 int32`,
`float32 float64`, plus probably `int64` and `uint8 … uint64`. Also `array`,
`matrix`, `struct`, `class`, and generic type parameters. None of these are
implemented. Their names are reserved.

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
| same type (`int`/`sint` count as the same) | allowed |
| untyped constant → integer type | allowed if the value fits (§5.3) |
| signed (`int`/`sint`) → `uint` | **error `ASX-TYPE-004`**, whatever the runtime value |
| `uint` → `int`/`sint` | **error `ASX-TYPE-005`** [provisional: rules undecided] |
| `void` → anything | error `ASX-TYPE-002` |
| anything else | error `ASX-TYPE-001` |

There are no implicit conversions between integers, `bool`, and `string`.
Explicit checked-conversion syntax is **[undecided]**.

## 6. Expressions

### 6.1 Operators **[v0.1]**

| Operator | Operands | Result |
|---|---|---|
| `+ - * / %` | two integers of the same signedness | the same integer type |
| unary `-` | a signed integer | the same type (`uint` is rejected) |
| `< <= > >=` | two integers of the same signedness | `bool` |
| `== !=` | two integers of the same signedness, two `bool`s, or two `string`s | `bool` |
| `&& \|\|` | `bool` | `bool`; short-circuit evaluation |
| `!` | `bool` | `bool` |

- **[v0.1, provisional]** Mixing signed and unsigned operands is rejected
  (`ASX-TYPE-011`).
- **[v0.1, provisional]** `/` truncates toward zero. `%` takes the sign of
  the dividend. Division by zero is a runtime error (`ASX-RUN-003`).
- **[undecided]** String concatenation (`"a" + "b"` is rejected), bitwise
  operators, shifts, `?:`, `?.`, `??`.
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
`if`/`else if`/`else` chain whose every branch does. Loops are never treated
as guaranteeing a return **[provisional]**.

## 8. Built-ins

### 8.1 `trace(…)` **[v0.1, provisional formatting]**
Takes any number of arguments of any non-void type. It writes their text
forms separated by one space, followed by a newline, to standard output.
Integers print in decimal, booleans as `true`/`false`, and strings as-is.
An integer constant argument is typed `int`.

```actionscript
trace("x =", 42, true);   // x = 42 true
```

## 9. Diagnostics and runtime errors

Format:

```text
file.rs:LINE:COL: error CODE: message
    <source line>
    <caret underline>
  help: <suggestion>
```

The `ASX` code prefix is **provisional**. Each code is
`ASX-<STAGE>-<NNN>`, where the stages are `LEX`, `PARSE`, `NAME`, `TYPE`,
`FLOW`, `ENTRY`, and `RUN`. The full list is in `docs/guide/08-diagnostics.md`.

The checker reports every semantic error it finds. The parser stops at the
first syntax error. A runtime error ends the program with exit status 1 and
reports the source location where it happened.

## 10. Comparison with ActionScript 3

RuneScript keeps AS3's names and feel (`package`, `trace`, `int`/`uint`,
`public`/`protected`/`private`), but its syntax and typing differ on
purpose:

| Topic | ActionScript 3 | RuneScript |
|---|---|---|
| Package | `package a.b { … }` block | `package a.b;` statement |
| Function | `function add(a:int, b:int):int { }` | `int add(int a, int b) { }` |
| Variable | `var n:int = 0;` | `int n = 0;` |
| Semicolons | Optional | Required |
| Untyped values | `*` and missing annotations allowed | No `any`; every declaration is explicitly typed |
| Top-level code | One externally visible definition per file; no `main` | Many functions per file; `main()` entry point |
| `int`/`uint` mixing | Implicit coercion; `var u:uint = -1` wraps to 4294967295 | Compile errors (`ASX-TYPE-004/005/007/011`) |
| Integer overflow | Wraps | Never wraps silently (v0.1: runtime error; widening planned) |
| Integer division | Produces `Number` (`7/2 == 3.5`) | Truncating integer division |
| Floating point | `Number` (64-bit) | `float` (default 32-bit, widening) [planned] |
| Type names | `String`, `Boolean` | `string`, `bool`; `sint` alias added |
| Truthiness | `if (count)` allowed | Conditions must be `bool` |
| Variable scope | `var` is function-scoped and hoisted | Block-scoped; no shadowing [provisional] |
| Visibility | `public protected private internal`; default `internal` | `public protected private`; default and meaning undecided |
| Casts | `T(x)` and `x as T` | Explicit checked conversions [undecided] |
| Imports | `import a.b.C;` | `use "a.b";` or `use "path/file";` [planned] |
| `trace` | `trace(...args)`, space-separated | Same, written to stdout |

**Games and graphics.** AS3 was used mostly for games and interactive
media in Flash Player. RuneScript plans native (non-browser) graphics, input,
and UI as a standard package brought in with `use` **[planned]**. The
recommended backend is raylib; this is not yet confirmed.

AS3 and Apache Royale features that are **[undecided]** for RuneScript:
classes and interfaces, `static`, abstract classes, private constructors,
metadata, E4X XML, `Vector.<T>` (vs. `Array<T>` generics), type inference,
arrow functions, `?.`, `??`, import aliases, and verbatim strings.

## 11. Conflicts with CLAUDE.md

1. **First-milestone scope.** CLAUDE.md §10 lists literals, declarations,
   calls, and `return`. The owner later widened v0.1 to include
   conditionals, loops, assignment, and basic functions, and this document
   follows the owner. CLAUDE.md §10 should be updated.
2. **Open-questions file.** CLAUDE.md §6 and §9 list
   `docs/OPEN_QUESTIONS.md` as a repository file. The owner has asked that
   it stay local and uncommitted.
