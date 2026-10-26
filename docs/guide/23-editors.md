# 23. Editors

RuneScript works in Visual Studio Code, Vim, Neovim, and Emacs. Each has a
package in [`editors/`](../../editors/README.md) that highlights RuneScript,
and all of them use the same language server, `runescript lsp`, built
into the `runescript` command. The language server gives you:

- **Errors as you type.** The compiler's errors and warnings appear under
  the code before you save, with their codes and help, exactly as
  `runescript check` reports them.
- **Hover.** Point at a name to see its type, at a call to see the
  function's signature (standard library functions included), or at a
  struct to see its fields.
- **Go to definition** for variables, parameters, functions (also in other
  files and packages), structs, fields, and the files a `use` loads.
- **Completion** of keywords, names in scope, functions, package members
  after `csv.` or `geom.`, array methods, struct fields, and `std:`
  packages inside `use "…"`.
- **An outline** of the file's structs and functions.

## Setting up

First install RuneScript ([chapter 1](01-getting-started.md)) and check
that `runescript version` works in a terminal. Then:

**Visual Studio Code.** Build and install the extension from
[`editors/vscode`](../../editors/vscode/README.md):

```sh
cd editors/vscode
npm install
npx vsce package
code --install-extension runescript-0.6.0.vsix
```

If `runescript` is not on your `PATH`, set **RuneScript › Server: Path**
in the settings.

**Vim and Neovim.** Add [`editors/vim`](../../editors/vim/README.md) to
your runtime path. In Neovim 0.11 or later, one line starts the language
server:

```lua
vim.lsp.enable('runescript')
```

In Vim, connect vim-lsp, coc.nvim, or ALE to the command `runescript lsp`;
the plugin's README shows each.

**Emacs.** Load [`editors/emacs/runescript-mode.el`](../../editors/emacs/README.md),
then start the language server with Eglot:

```elisp
(require 'runescript-mode)
(add-hook 'runescript-mode-hook #'eglot-ensure)
```

**Any other editor** with a language server client: run `runescript lsp`,
which talks on standard input and output, for `.rsc` and `.rsm` files.

## How your code is checked

The language server checks your program the way `runescript run` would,
adjusted for code that is still being written:

- **Use paths** are resolved against the workspace folder that contains the
  file, as `runescript run` resolves them against the working directory.
  Open the folder you run your programs from.
- **Unsaved changes count.** Files open in the editor are read from the
  editor, so renaming a function in one file shows the errors it causes in
  the files that use it, before you save either.
- **A file without `main` is fine**, since another file may use it. A
  module (`.rsm`) in its package's directory is checked together with the
  rest of its package.
- **Warnings** appear once the program has no errors, as on the command
  line.

Next: back to the [guide index](README.md)
