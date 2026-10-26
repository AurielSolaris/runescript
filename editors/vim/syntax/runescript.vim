" Vim syntax file
" Language:    RuneScript (.rsc programs, .rsm modules)
" Maintainer:  RuneScript contributors
" URL:         https://github.com/AurielSolaris/runescript

if exists("b:current_syntax")
  finish
endif

let s:cpo_save = &cpo
set cpo&vim

syn keyword runescriptConditional if else
syn keyword runescriptRepeat      while for each in
syn keyword runescriptStatement   return break continue throw new
syn keyword runescriptException   try catch finally
syn keyword runescriptInclude     package use as
syn keyword runescriptStorage     public protected private extern
syn keyword runescriptStructure   struct
syn keyword runescriptType        int sint uint int8 int16 int32 int64 uint8 uint16 uint32 uint64
syn keyword runescriptType        float float32 float64 bool string char void
syn keyword runescriptType        Array Signal Element Error RangeError TypeError ArgumentError
syn keyword runescriptBoolean     true false
syn keyword runescriptConstant    NaN Infinity
syn keyword runescriptBuiltin     trace isNaN isFinite
syn keyword runescriptReserved    class array
syn match   runescriptTemplate    "\<template\ze\s*{"

syn match   runescriptFunction    "\<\h\w*\ze\s*("
syn match   runescriptStructName  "\(\<struct\s\+\)\@<=\h\w*"

syn match   runescriptNumber      "\<0x[0-9A-Fa-f_]\+\>"
syn match   runescriptNumber      "\<0b[01_]\+\>"
syn match   runescriptNumber      "\<0o[0-7_]\+\>"
syn match   runescriptNumber      "\<\d[0-9_]*\(\.\d[0-9_]*\)\=\([eE][+-]\=\d[0-9_]*\)\=\>"

syn match   runescriptEscape      contained "\\\(u{\x\+}\|.\)"
syn region  runescriptString      start=+"+ skip=+\\\\\|\\"+ end=+"+ contains=runescriptEscape
" The hole comes first: an item defined later wins at the same position,
" so {{ is an escape, not a hole.
syn region  runescriptHole        contained matchgroup=runescriptHoleDelim start=+{+ end=+}+ contains=TOP
syn match   runescriptBraceEscape contained "{{\|}}"
syn region  runescriptInterp      start=+\$"+ skip=+\\\\\|\\"+ end=+"+ contains=runescriptEscape,runescriptBraceEscape,runescriptHole
syn match   runescriptChar        "'\(\\\(u{\x\+}\|.\)\|[^'\\]\)'" contains=runescriptEscape

syn keyword runescriptTodo        contained TODO FIXME XXX NOTE
syn region  runescriptComment     start="/\*" end="\*/" contains=runescriptTodo,@Spell
syn match   runescriptComment     "//.*$" contains=runescriptTodo,@Spell

hi def link runescriptConditional Conditional
hi def link runescriptRepeat      Repeat
hi def link runescriptStatement   Statement
hi def link runescriptException   Exception
hi def link runescriptInclude     Include
hi def link runescriptStorage     StorageClass
hi def link runescriptStructure   Structure
hi def link runescriptType        Type
hi def link runescriptBoolean     Boolean
hi def link runescriptConstant    Constant
hi def link runescriptBuiltin     Function
hi def link runescriptReserved    Error
hi def link runescriptTemplate    Keyword
hi def link runescriptFunction    Function
hi def link runescriptStructName  Typedef
hi def link runescriptNumber      Number
hi def link runescriptEscape      SpecialChar
hi def link runescriptBraceEscape SpecialChar
hi def link runescriptHoleDelim   Delimiter
hi def link runescriptString      String
hi def link runescriptInterp      String
hi def link runescriptChar        Character
hi def link runescriptTodo        Todo
hi def link runescriptComment     Comment

let b:current_syntax = "runescript"

let &cpo = s:cpo_save
unlet s:cpo_save
