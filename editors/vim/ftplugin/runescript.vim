" Vim filetype plugin
" Language: RuneScript

if exists("b:did_ftplugin")
  finish
endif
let b:did_ftplugin = 1

setlocal commentstring=//\ %s
setlocal comments=s1:/*,mb:*,ex:*/,://
setlocal formatoptions-=t formatoptions+=croql
setlocal suffixesadd=.rsc,.rsm
setlocal expandtab shiftwidth=4 softtabstop=4

let b:undo_ftplugin = "setlocal commentstring< comments< formatoptions< suffixesadd< expandtab< shiftwidth< softtabstop<"
