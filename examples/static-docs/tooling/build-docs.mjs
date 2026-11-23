import {
  copyFileSync, cpSync, existsSync, mkdirSync, mkdtempSync, readFileSync, realpathSync,
  renameSync, rmSync, writeFileSync,
} from "node:fs";
import { basename, dirname, join, relative, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { parseArgs } from "node:util";
import { defaultTreeAdapter as tree, parse, parseFragment, serialize } from "parse5";
import { createHighlighter } from "shiki";

const here = fileURLToPath(new URL(".", import.meta.url));
const example = resolve(here, "..");
const repo = resolve(example, "../..");
const { positionals, values } = parseArgs({
  allowPositionals: true,
  options: { "source-map": { type: "boolean" } },
});
if (positionals.length > 1) throw new Error("usage: bun run build [--source-map] [output-directory]");
const output = positionals[0] ? resolve(positionals[0]) : join(example, "dist-enhanced");
if (existsSync(output)) throw new Error(`Output already exists: ${output}`);

const parent = realpathSync(dirname(output));
const work = mkdtempSync(join(parent, ".runescript-docs-"));
const stage = join(work, "site");

function elements(node, tag) {
  const found = node.tagName === tag ? [node] : [];
  for (const child of node.childNodes ?? []) found.push(...elements(child, tag));
  return found;
}

function attribute(node, name) {
  return node.attrs.find(attr => attr.name === name);
}

function codeText(node) {
  return node.nodeName === "#text" ? node.value : (node.childNodes ?? []).map(codeText).join("");
}

try {
  const source = join(work, "source");
  const sheets = join(source, "styles");
  mkdirSync(sheets, { recursive: true });
  mkdirSync(join(source, "public"));
  cpSync(join(example, "content"), join(source, "content"), { recursive: true });
  copyFileSync(join(here, "styles.scss"), join(sheets, "docs.scss"));
  copyFileSync(join(example, "public/docs.css"), join(sheets, "base.css"));
  const manifest = JSON.parse(readFileSync(join(example, "runescript.web.json"), "utf8"));
  writeFileSync(join(source, "runescript.web.json"), JSON.stringify({ ...manifest, styles: ["styles/docs.scss"] }));
  const layout = readFileSync(join(example, "layout.html"), "utf8");
  writeFileSync(join(source, "layout.html"), layout.replace('<link rel="stylesheet" href="/docs.css">', "<!-- runescript -->"));
  const flags = values["source-map"] ? ["--source-map"] : [];
  const built = Bun.spawnSync(["go", "run", "./cmd/runescript", "build", "--target", "web", ...flags, "-o", stage, source], {
    cwd: repo,
    stdout: "inherit",
    stderr: "inherit",
  });
  if (built.exitCode !== 0) throw new Error(`RuneScript build failed with exit code ${built.exitCode}`);

  const indexPath = join(stage, "docs.html");
  const chapterPath = join(stage, "docs/getting-started.html");
  const index = parse(readFileSync(indexPath, "utf8"));
  const chapter = parse(readFileSync(chapterPath, "utf8"));
  const assets = join(stage, "assets");
  mkdirSync(assets);

  const site = resolve(process.env.RUNESCRIPT_SITE ?? join(repo, "../runescript.site"));
  const grammarFile = join(site, "src/lib/runescript.grammar.ts");
  const runeLang = existsSync(grammarFile)
    ? (await import(pathToFileURL(grammarFile).href)).runescript
    : "actionscript-3";
  const highlighter = await createHighlighter({ themes: ["github-light", "github-dark"], langs: ["bash", runeLang] });
  try {
    for (const pre of elements(chapter, "pre")) {
      const code = pre.childNodes.find(node => node.tagName === "code");
      if (!code) continue;
      const language = attribute(code, "class")?.value;
      const lang = language === "language-runescript" ? (typeof runeLang === "string" ? runeLang : runeLang.name)
        : language === "language-sh" ? "bash" : "text";
      const highlighted = parseFragment(highlighter.codeToHtml(codeText(code).replace(/\n$/, ""), {
        lang,
        themes: { light: "github-light", dark: "github-dark" },
      })).childNodes[0];
      const wrapper = parseFragment(
        '<div class="code-example" x-data="codeCopy"><button type="button" x-cloak x-on:click="copy" x-text="copied ? \'Copied\' : \'Copy code\'">Copy code</button></div>',
      ).childNodes[0];
      tree.appendChild(wrapper, highlighted);
      tree.insertBefore(pre.parentNode, wrapper, pre);
      tree.detachNode(pre);
    }
  } finally {
    highlighter.dispose();
  }

  const bundle = await Bun.build({
    entrypoints: [join(here, "chapter.ts")],
    outdir: assets,
    naming: "chapter-[hash].[ext]",
    target: "browser",
    format: "esm",
    minify: true,
    sourcemap: values["source-map"] ? "linked" : "none",
  });
  const entry = bundle.outputs.find(file => file.kind === "entry-point");
  if (!bundle.success || !entry) throw new Error("Bun did not emit a chapter entry");
  const script = parseFragment(`<script type="module" src="/assets/${basename(entry.path)}"></script>`).childNodes[0];
  tree.appendChild(elements(chapter, "body")[0], script);

  writeFileSync(indexPath, serialize(index));
  writeFileSync(chapterPath, serialize(chapter));
  renameSync(stage, output);
  console.log(`Built enhanced static docs in ${output}`);
} finally {
  const resolved = realpathSync(work);
  if (dirname(resolved) !== parent || !basename(relative(parent, resolved)).startsWith(".runescript-docs-")) {
    throw new Error("Unexpected temporary build directory");
  }
  rmSync(resolved, { recursive: true, force: true });
}
