# Editor support

RuneScript has a language server, `runescript lsp`, built into the
`runescript` command, and editor packages that highlight RuneScript and
connect to it.

| Editor | Package | Highlighting | Language server |
|---|---|---|---|
| Visual Studio Code | [`vscode/`](vscode/README.md) | TextMate grammar | started by the extension |
| Vim, Neovim | [`vim/`](vim/README.md) | Vim syntax file | Neovim 0.11+ built in, vim-lsp, coc.nvim, ALE |
| Emacs | [`emacs/`](emacs/README.md) | `runescript-mode` | Eglot, lsp-mode |

Any other editor with an LSP client works too: run `runescript lsp`
(it talks on standard input and output), for files of the `.rsc` and
`.rsm` extensions.

## What the language server does

- **Diagnostics as you type**: every compiler error and warning, with its
  code and help, before you save. Files open in the editor are read from
  the editor, so a change in one file updates the others.
- **Hover**: the type of a name, a function's signature (including the
  standard library's), a struct's fields, a package's functions.
- **Go to definition** for variables, parameters, functions, structs,
  fields, and the files and packages of `use` declarations.
- **Completion**: keywords, the names in scope, the functions and structs
  of the package, members after `pkg.` (standard library and RuneScript
  packages), array methods, struct and error fields, and `std:` packages
  inside a `use "…"`.
- **Document symbols**: structs with their fields, extern functions, and
  functions with their nested functions.

It checks programs the way `runescript run` does: `use` paths are resolved
against the workspace folder that contains the file (the working
directory, for the compiler). A module (`.rsm`) in its package's
directory is checked together with the rest of its package, and a file
without `main` is not reported for it, since another file may use it.
Warnings appear only when the program has no errors, as on the command
line.

Not yet: rename, find references, formatting, signature help, and code
actions. The packages' highlighting does not cover template markup
inside `template { … }` beyond its RuneScript expressions.
