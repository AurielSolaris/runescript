" Vim indent file
" Language: RuneScript
" RuneScript's blocks are C's: braces, semicolons, // and /* */ comments.

if exists("b:did_indent")
  finish
endif
let b:did_indent = 1

setlocal cindent
" No extra indent for case labels, function types, or continuation lines
" after a declaration's type.
setlocal cinoptions=j1,J1,(0,u0,W4,m1,+4
setlocal cinkeys-=0#

let b:undo_indent = "setlocal cindent< cinoptions< cinkeys<"
