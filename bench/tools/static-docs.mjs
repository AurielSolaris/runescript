import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { copyFileSync, cpSync, mkdirSync, readFileSync, readdirSync, symlinkSync, unlinkSync, writeFileSync } from "node:fs";
import { basename, join, resolve } from "node:path";
import { gzipSync } from "node:zlib";

const [runescript, scratch, astroProject, runsArg, scssArg] = process.argv.slice(2);
const useSCSS = scssArg === "true";
const runs = Number(runsArg);
const example = resolve("examples/static-docs");
const astro = join(scratch, "astro");
const modules = join(astroProject, "node_modules");
const astroPackage = join(modules, "astro");
let packageInfo;
try {
  packageInfo = JSON.parse(readFileSync(join(astroPackage, "package.json"), "utf8"));
} catch (error) {
  throw new Error(`Install Astro in ${astroProject} or pass -astro to rsbench: ${error.message}`);
}
cpSync("bench/fixtures/astro-docs", astro, { recursive: true });
cpSync(join(example, "content"), join(astro, "content"), { recursive: true });
cpSync(join(example, "public"), join(astro, "public"), { recursive: true });
copyFileSync(join(example, "runescript.web.json"), join(astro, "src/site.json"));

let runeProject = example;
if (useSCSS) {
  runeProject = join(scratch, "rune-project");
  cpSync(join(example, "content"), join(runeProject, "content"), { recursive: true });
  mkdirSync(join(runeProject, "public"));
  unlinkSync(join(astro, "public/docs.css"));
  for (const project of [runeProject, astro]) {
    mkdirSync(join(project, "styles"));
    copyFileSync(join(example, "public/docs.css"), join(project, "styles/base.css"));
    copyFileSync(join(example, "tooling/styles.scss"), join(project, "styles/docs.scss"));
  }
  const manifest = JSON.parse(readFileSync(join(example, "runescript.web.json"), "utf8"));
  writeFileSync(join(runeProject, "runescript.web.json"), JSON.stringify({ ...manifest, styles: ["styles/docs.scss"] }));
  const styleLink = '<link rel="stylesheet" href="/docs.css">';
  const layout = readFileSync(join(example, "layout.html"), "utf8");
  writeFileSync(join(runeProject, "layout.html"), layout.replace(styleLink, "<!-- runescript -->"));
  const astroLayout = join(astro, "src/Layout.astro");
  writeFileSync(astroLayout, readFileSync(astroLayout, "utf8")
    .replace(/^---\r?\n/, '---\nimport "../styles/docs.scss";\n').replace(styleLink, ""));
  const astroConfig = join(astro, "astro.config.mjs");
  writeFileSync(astroConfig, readFileSync(astroConfig, "utf8")
    .replace('format: "file"', 'format: "file", inlineStylesheets: "never"'));
}

const dependencyLink = join(astro, "node_modules");
symlinkSync(modules, dependencyLink, "junction");
try {
  const builders = [
    { name: "RuneScript", root: scratch, times: [], build: out => execFileSync(runescript, ["build", "--target", "web", "-o", out, runeProject]) },
    { name: "Astro", root: astro, times: [], build: out => execFileSync(process.execPath, [join(astroPackage, packageInfo.bin.astro), "build", "--outDir", out], {
      cwd: astro,
      env: { ...process.env, ASTRO_TELEMETRY_DISABLED: "1" },
    }) },
  ];
  for (let i = -2; i < runs; i++) {
    const order = i % 2 === 0 ? builders : [...builders].reverse();
    for (const builder of order) {
      builder.out = join(builder.root, `dist-${i + 2}`);
      const start = performance.now();
      builder.build(builder.out);
      if (i >= 0) builder.times.push(performance.now() - start);
    }
  }

  const expectedCSS = readFileSync(join(example, "public", "docs.css"));
  function payload(dir) {
    const paths = readdirSync(dir, { recursive: true, withFileTypes: true })
      .filter(entry => entry.isFile())
      .map(entry => join(entry.parentPath, entry.name).slice(dir.length + 1).replaceAll("\\", "/"))
      .sort();
    const styles = paths.filter(path => path.endsWith(".css"));
    assert.equal(styles.length, 1);
    assert.deepEqual(paths.filter(path => path !== styles[0]), ["docs.html", "docs/getting-started.html"]);
    const [css, index, chapter] = [styles[0], "docs.html", "docs/getting-started.html"]
      .map(path => readFileSync(join(dir, path)));
    if (useSCSS) {
      assert.match(css.toString(), /\.code-example pre/);
      assert.match(css.toString(), /prefers-color-scheme:\s*dark/);
    } else {
      assert.deepEqual(css, expectedCSS);
    }
    for (const html of [index, chapter]) {
      assert.ok(html.toString().includes(basename(styles[0])));
      assert.doesNotMatch(html.toString(), /<script\b/i);
    }
    assert.match(index.toString(), /href="\/docs\/getting-started"/);
    assert.match(chapter.toString(), /href="#your-first-program"/);
    assert.match(chapter.toString(), /id="your-first-program"/);
    assert.match(chapter.toString(), /Hello, RuneScript!/);
    const files = [css, index, chapter];
    return [
      files.reduce((n, file) => n + file.length, 0),
      files.reduce((n, file) => n + gzipSync(file, { level: 9 }).length, 0),
    ];
  }
  console.log(`Astro ${packageInfo.version}, Node ${process.version}, ${process.platform}/${process.arch}`);
  if (useSCSS) console.log("Both builders compile the same SCSS and link a shared CSS asset.");
  console.log(`Clean output directories, 2 warm-ups and ${runs} measured runs per builder.`);
  console.log("\n| Builder | Median build (ms) | HTML + CSS (B) | Gzip (B) |");
  console.log("| --- | ---: | ---: | ---: |");
  for (const { name, times, out } of builders) {
    times.sort((a, b) => a - b);
    const median = (times[Math.floor((runs - 1) / 2)] + times[Math.floor(runs / 2)]) / 2;
    const [raw, gzip] = payload(out);
    console.log(`| ${name} | ${median.toFixed(1)} | ${raw} | ${gzip} |`);
  }
} finally {
  unlinkSync(dependencyLink);
}
