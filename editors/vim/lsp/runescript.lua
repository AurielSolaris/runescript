-- Neovim 0.11+ language server configuration for RuneScript. With this
-- directory on 'runtimepath', enable it with:
--
--   vim.lsp.enable('runescript')
--
-- The server resolves `use` declarations against the root directory, as
-- `runescript run` resolves them against the working directory.
return {
  cmd = { 'runescript', 'lsp' },
  filetypes = { 'runescript' },
  root_markers = { '.git' },
}
