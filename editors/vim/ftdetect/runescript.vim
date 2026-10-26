" RuneScript programs (.rsc) and modules (.rsm).
" Vim and Neovim detect *.rsc as RouterOS scripts; this plugin overrides
" that, so 'set filetype' rather than 'setfiletype'.
autocmd BufRead,BufNewFile *.rsc,*.rsm set filetype=runescript
