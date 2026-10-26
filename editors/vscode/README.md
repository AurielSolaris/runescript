# RuneScript for Visual Studio Code

Language support for [RuneScript](https://github.com/AurielSolaris/runescript),
a statically typed, ActionScript-inspired language implemented in Go.

- **Syntax highlighting** for `.rsc` programs and `.rsm` modules, including
  interpolated strings (`$"total: {n}"`).
- **Diagnostics as you type**: the compiler's errors and warnings, with
  their codes (`ASX-TYPE-004`) and help, before you save.
- **Hover** shows the type of a name, a function's signature, or a struct's
  fields.
- **Go to definition** for variables, functions (in other files and
  packages too), structs, fields, and `use` declarations.
- **Completion** of keywords, local names, functions, package members after
  `csv.`, array methods, struct fields, and `std:` packages in a `use`.
- **Outline** of a file's structs, functions, and nested functions.
- Comment toggling, bracket matching, and indentation.

## Requirements

Everything except highlighting comes from the RuneScript language server,
`runescript lsp`, which is part of the `runescript` command. Install
RuneScript (see the project's README), and check that `runescript version`
works in a terminal. If `runescript` is not on your `PATH`, set
**RuneScript › Server: Path** (`runescript.server.path`) to the command's
full path.

The server resolves `use` declarations against the workspace folder that
contains the file, as `runescript run` resolves them against the working
directory: open the folder you run your programs from.

## Settings

| Setting | Default | |
|---|---|---|
| `runescript.server.enabled` | `true` | Start the language server. With `false`, you get highlighting only. |
| `runescript.server.path` | `runescript` | The `runescript` command. |
| `runescript.trace.server` | `off` | Log the messages to and from the server (`messages`, `verbose`). |

The command **RuneScript: Restart Language Server** restarts it, for
example after installing a new version of RuneScript.

## Building the extension

```sh
cd editors/vscode
npm install
npx vsce package          # writes runescript-0.6.0.vsix
code --install-extension runescript-0.6.0.vsix
```

To try it without installing, open `editors/vscode` in VS Code and press
F5: a second window opens with the extension loaded.
