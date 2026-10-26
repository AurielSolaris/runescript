# RuneScript for Vim and Neovim

A Vim plugin for RuneScript: `.rsc` and `.rsm` files get syntax
highlighting (including interpolated strings), C-style indentation,
`//` comments for `gc`-style commenting plugins, and four-space indents.
The language server, `runescript lsp`, adds diagnostics as you type, hover,
go to definition, completion, and document symbols through any LSP client.

Vim and Neovim detect `*.rsc` as RouterOS scripts by default; this plugin
makes them RuneScript.

## Installing the plugin

The plugin is the `editors/vim` directory of the RuneScript repository.
With a plugin manager that takes a subdirectory, point it there; for
example, with [vim-plug](https://github.com/junegunn/vim-plug):

```vim
Plug 'AurielSolaris/runescript', { 'rtp': 'editors/vim' }
```

With [lazy.nvim](https://github.com/folke/lazy.nvim):

```lua
{
  'AurielSolaris/runescript',
  config = function(plugin)
    vim.opt.rtp:append(plugin.dir .. '/editors/vim')
  end,
}
```

Or copy it by hand: the `ftdetect`, `ftplugin`, `indent`, `syntax`, and
`lsp` directories go into `~/.vim` (Vim) or `~/.config/nvim` (Neovim).
On Windows, they are `~/vimfiles` and `~/AppData/Local/nvim`.

## The language server

Install RuneScript and check that `runescript version` works in a
terminal. Then connect your LSP client to the command `runescript lsp`.

**Neovim 0.11 or later** (built-in client): the plugin ships the
configuration in `lsp/runescript.lua`, so one line enables it:

```lua
vim.lsp.enable('runescript')
```

**[vim-lsp](https://github.com/prabirshrestha/vim-lsp):**

```vim
if executable('runescript')
  au User lsp_setup call lsp#register_server({
    \ 'name': 'runescript',
    \ 'cmd': {server_info -> ['runescript', 'lsp']},
    \ 'allowlist': ['runescript'],
    \ })
endif
```

**[coc.nvim](https://github.com/neoclide/coc.nvim)** (`:CocConfig`):

```json
{
  "languageserver": {
    "runescript": {
      "command": "runescript",
      "args": ["lsp"],
      "filetypes": ["runescript"],
      "rootPatterns": [".git"]
    }
  }
}
```

**[ALE](https://github.com/dense-analysis/ale):**

```vim
call ale#linter#Define('runescript', {
\   'name': 'runescript',
\   'lsp': 'stdio',
\   'executable': 'runescript',
\   'command': '%e lsp',
\   'project_root': {buffer -> fnamemodify(bufname(buffer), ':p:h')},
\})
```

The server resolves `use` declarations against the project root, as
`runescript run` resolves them against the working directory: start Vim
in, or point the client's root at, the directory you run your programs
from.
