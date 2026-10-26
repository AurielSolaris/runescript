# RuneScript for Emacs

`runescript-mode.el` is a major mode for RuneScript: `.rsc` and `.rsm`
files get syntax highlighting, indentation by brace depth (four columns,
`runescript-indent-offset`), and `//` comments for `M-;`. It also
registers the language server, `runescript lsp`, with Eglot and lsp-mode,
for diagnostics as you type, hover, go to definition, completion, and
`imenu`-style document symbols.

Emacs 27.1 or later. Eglot is built into Emacs 29.

## Installing

Put `runescript-mode.el` on your `load-path` and require it:

```elisp
(add-to-list 'load-path "~/src/runescript/editors/emacs")
(require 'runescript-mode)
```

With `use-package` (Emacs 30 can fetch it straight from the repository):

```elisp
(use-package runescript-mode
  :vc (:url "https://github.com/AurielSolaris/runescript"
       :lisp-dir "editors/emacs"))
```

## The language server

Install RuneScript and check that `runescript version` works in a
terminal. Then, in a RuneScript buffer, run `M-x eglot` (or `M-x lsp`
with lsp-mode). To start it automatically:

```elisp
(add-hook 'runescript-mode-hook #'eglot-ensure)
```

If `runescript` is not on Emacs's `exec-path`, set the command:

```elisp
(setq runescript-server-command '("/path/to/runescript" "lsp"))
```

The server resolves `use` declarations against the project root (the
version-control root that Eglot finds), as `runescript run` resolves
them against the working directory.
