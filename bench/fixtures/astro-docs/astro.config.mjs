import { defineConfig } from "astro/config";

export default defineConfig({
  trailingSlash: "never",
  build: { format: "file" },
  cacheDir: "./.astro-cache",
  vite: { cacheDir: "./.vite-cache" },
  markdown: { syntaxHighlight: false },
});
