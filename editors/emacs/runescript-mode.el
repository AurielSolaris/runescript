;;; runescript-mode.el --- Major mode for RuneScript -*- lexical-binding: t; -*-

;; Copyright (c) 2026 AurielSolaris and RuneScript contributors
;; SPDX-License-Identifier: MIT

;; Author: RuneScript contributors
;; Version: 0.6.0
;; Package-Requires: ((emacs "27.1"))
;; Keywords: languages
;; URL: https://github.com/AurielSolaris/runescript

;;; Commentary:

;; A major mode for RuneScript programs (.rsc) and modules (.rsm):
;; syntax highlighting, brace-based indentation, and comment commands.
;;
;; The RuneScript language server, `runescript lsp', adds diagnostics as
;; you type, hover, go to definition, completion, and an outline.  This
;; file registers it with Eglot (built into Emacs 29) and with lsp-mode,
;; so `M-x eglot' or `M-x lsp' starts it.
;;
;; Emacs maps no other mode to .rsc or .rsm by default.

;;; Code:

(defgroup runescript nil
  "Major mode for RuneScript."
  :group 'languages
  :prefix "runescript-")

(defcustom runescript-indent-offset 4
  "Number of columns per level of indentation in RuneScript code."
  :type 'integer
  :safe #'integerp)

(defcustom runescript-server-command '("runescript" "lsp")
  "The command that starts the RuneScript language server, as a list."
  :type '(repeat string))

(defconst runescript-keywords
  '("package" "use" "as" "return" "if" "else" "while" "for" "each" "in"
    "break" "continue" "try" "catch" "finally" "throw" "new"
    "public" "protected" "private" "extern" "struct" "template")
  "RuneScript keywords, and the contextual words each, in, as, template.")

(defconst runescript-types
  '("int" "sint" "uint" "int8" "int16" "int32" "int64"
    "uint8" "uint16" "uint32" "uint64" "float" "float32" "float64"
    "bool" "string" "char" "void"
    "Array" "Signal" "Element" "Error" "RangeError" "TypeError" "ArgumentError")
  "RuneScript's built-in types.")

(defconst runescript-constants '("true" "false" "NaN" "Infinity")
  "RuneScript's constants.")

(defconst runescript-builtins '("trace" "isNaN" "isFinite")
  "RuneScript's built-in functions.")

(defconst runescript-reserved '("class" "array")
  "Words reserved for planned features.")

(defvar runescript-font-lock-keywords
  `((,(regexp-opt runescript-keywords 'symbols) . font-lock-keyword-face)
    (,(regexp-opt runescript-types 'symbols) . font-lock-type-face)
    (,(regexp-opt runescript-constants 'symbols) . font-lock-constant-face)
    (,(regexp-opt runescript-builtins 'symbols) . font-lock-builtin-face)
    (,(regexp-opt runescript-reserved 'symbols) . font-lock-warning-face)
    ("\\_<struct\\_>[ \t]+\\([A-Za-z_][A-Za-z0-9_]*\\)" 1 font-lock-type-face)
    ("\\_<package\\_>[ \t]+\\([A-Za-z0-9_.]+\\)" 1 font-lock-constant-face)
    ("\\_<\\([A-Za-z_][A-Za-z0-9_]*\\)[ \t]*(" 1 font-lock-function-name-face)
    ("\\_<\\(?:0x[0-9A-Fa-f_]+\\|0b[01_]+\\|0o[0-7_]+\\|[0-9][0-9_]*\\(?:\\.[0-9][0-9_]*\\)?\\(?:[eE][-+]?[0-9][0-9_]*\\)?\\)\\_>"
     . font-lock-constant-face))
  "Highlighting for RuneScript mode.")

(defvar runescript-mode-syntax-table
  (let ((st (make-syntax-table)))
    ;; Comments: // to the end of the line, and /* ... */.
    (modify-syntax-entry ?/ ". 124b" st)
    (modify-syntax-entry ?* ". 23" st)
    (modify-syntax-entry ?\n "> b" st)
    ;; Strings and characters.
    (modify-syntax-entry ?\" "\"" st)
    (modify-syntax-entry ?' "\"" st)
    (modify-syntax-entry ?\\ "\\" st)
    ;; Identifiers.
    (modify-syntax-entry ?_ "_" st)
    ;; Operators.
    (dolist (c '(?+ ?- ?% ?= ?< ?> ?! ?& ?| ?^ ?~ ?? ?: ?$))
      (modify-syntax-entry c "." st))
    st)
  "Syntax table for RuneScript mode.")

(defun runescript-indent-line ()
  "Indent the current line by the depth of its enclosing brackets."
  (interactive)
  (let* ((ppss (save-excursion (syntax-ppss (line-beginning-position))))
         (depth (car ppss))
         (in-comment-or-string (nth 8 ppss))
         (closing (save-excursion
                    (back-to-indentation)
                    (looking-at "[]})]")))
         (column (* runescript-indent-offset
                    (max 0 (- depth (if closing 1 0)))))
         (offset (- (current-column) (current-indentation))))
    (unless in-comment-or-string
      (indent-line-to column)
      (when (> offset 0)
        (move-to-column (+ column offset))))))

;;;###autoload
(define-derived-mode runescript-mode prog-mode "RuneScript"
  "Major mode for editing RuneScript programs and modules."
  :syntax-table runescript-mode-syntax-table
  (setq-local font-lock-defaults '(runescript-font-lock-keywords))
  (setq-local comment-start "// ")
  (setq-local comment-end "")
  (setq-local comment-start-skip "\\(?://+\\|/\\*+\\)\\s *")
  (setq-local indent-line-function #'runescript-indent-line)
  (setq-local indent-tabs-mode nil)
  (setq-local electric-indent-chars (append '(?\} ?\)) electric-indent-chars)))

;;;###autoload
(add-to-list 'auto-mode-alist '("\\.rs[cm]\\'" . runescript-mode))

;; The language server.

(defvar eglot-server-programs)
(with-eval-after-load 'eglot
  (add-to-list 'eglot-server-programs
               `(runescript-mode . ,runescript-server-command)))

(declare-function lsp-register-client "lsp-mode")
(declare-function make-lsp-client "lsp-mode")
(declare-function lsp-stdio-connection "lsp-mode")
(declare-function lsp-activate-on "lsp-mode")
(defvar lsp-language-id-configuration)
(with-eval-after-load 'lsp-mode
  (add-to-list 'lsp-language-id-configuration '(runescript-mode . "runescript"))
  (lsp-register-client
   (make-lsp-client
    :new-connection (lsp-stdio-connection (lambda () runescript-server-command))
    :activation-fn (lsp-activate-on "runescript")
    :server-id 'runescript)))

(provide 'runescript-mode)

;;; runescript-mode.el ends here
