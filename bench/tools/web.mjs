import { execFileSync } from "node:child_process";
import { readFileSync, readdirSync } from "node:fs";
import { performance } from "node:perf_hooks";
import { join, resolve } from "node:path";
import { gzipSync } from "node:zlib";
import { pathToFileURL } from "node:url";

// rsbench supplies a freshly built CLI and an empty scratch directory.
const [cli, scratch, runsArg, iterationsArg] = process.argv.slice(2);
const root = resolve(scratch);
const runsCount = Number(runsArg);
const iterations = Number(iterationsArg);
const example = "examples/multi-page";
const build = (...args) => execFileSync(cli, ["build", ...args], { stdio: "inherit" });

const pages = [
  { name: "home", entry: "index.js", title: "Home", message: "Each page starts its own RuneScript program." },
  { name: "about", entry: "about/index.js", title: "About", message: "Both pages use the same cached runtime module." },
];
const singleBundle = page => page.name + "-single.mjs";
for (const page of pages) {
  build("-o", join(root, singleBundle(page)), example + "/" + page.name + ".rsc");
}
build("--target", "web", "-o", join(root, "inline"), example);
build("--target", "web", "--extract-css", "-o", join(root, "extracted"), example);
function rawModule(title, message) {
  return `export function run() {
  const target = document.querySelector("#app");
  const card = document.createElement("section");
  card.setAttribute("class", "card");
  const title = document.createElement("h1");
  title.appendChild(document.createTextNode(${JSON.stringify(title)}));
  const message = document.createElement("p");
  message.appendChild(document.createTextNode(${JSON.stringify(message)}));
  card.appendChild(title);
  card.appendChild(message);
  while (target.firstChild) target.removeChild(target.firstChild);
  target.appendChild(card);
  return 0;
}
if (typeof document !== "undefined") run();
`;
}
const rawSources = pages.map(({title, message}) => rawModule(title, message));
const [rawHome, rawAbout] = rawSources;
const rawCSS = ".card { padding: 24px; background-color: #edf4ff; color: #183454; border-radius: 12px; }\n";
const shared = readdirSync(join(root, "extracted/__runescript"));
const runtimes = shared.filter(name => name.startsWith("runtime-"));
const styles = shared.filter(name => name.endsWith(".css"));
if (runtimes.length !== 1 || styles.length !== 1) {
  throw new Error("the example should produce exactly one shared runtime and stylesheet");
}
const [runtime] = runtimes;
const [css] = styles;
const read = path => readFileSync(join(root, path));
const stats = files => ({
  bytes: files.reduce((sum, file) => sum + Buffer.byteLength(file), 0),
  gzipBytes: files.reduce((sum, file) => sum + gzipSync(file).length, 0)
});
const extractedRuntime = read("extracted/__runescript/" + runtime);
const extractedCSS = read("extracted/__runescript/" + css);
const extractedHome = read("extracted/__runescript/pages/index.js");
const extractedAbout = read("extracted/__runescript/pages/about/index.js");
const inlineHome = read("inline/__runescript/pages/index.js");
const inlineAbout = read("inline/__runescript/pages/about/index.js");
const inlineRuntime = read("inline/__runescript/" + runtime);
console.log(JSON.stringify({
  node: process.version,
  rawJS: {first: stats([rawHome, rawCSS]), twoPages: stats([rawHome, rawAbout, rawCSS])},
  singleBundles: {first: stats([read(singleBundle(pages[0]))]), twoPages: stats(pages.map(page => read(singleBundle(page))))},
  splitRuntime: {first: stats([inlineHome, inlineRuntime]), twoPages: stats([inlineHome, inlineAbout, inlineRuntime])},
  extracted: {first: stats([extractedHome, extractedRuntime, extractedCSS]), twoPages: stats([extractedHome, extractedAbout, extractedRuntime, extractedCSS])},
  homeEntryBeforeCSSExtraction: stats([inlineHome]),
  homeEntryAfterCSSExtraction: stats([extractedHome]),
}, null, 2));

const modules = await Promise.all(pages.map((page, index) => Promise.all([
  import("data:text/javascript," + encodeURIComponent(rawSources[index])),
  import(pathToFileURL(join(root, "inline/__runescript/pages", page.entry))),
  import(pathToFileURL(join(root, "extracted/__runescript/pages", page.entry))),
  import(pathToFileURL(join(root, singleBundle(page)))),
])));
class Element {
  constructor(tag) { this.tag = tag; this.children = []; this.attrs = {}; }
  get firstChild() { return this.children[0]; }
  appendChild(node) { this.children.push(node); return node; }
  removeChild(node) { this.children.splice(this.children.indexOf(node), 1); }
  setAttribute(name, value) { this.attrs[name] = value; }
}
const target = new Element("main");
globalThis.document = {
  querySelector: () => target,
  createElement: tag => new Element(tag),
  createTextNode: text => text,
};
function shape(node) {
  return typeof node === "string" ? node : [node.tag, ...node.children.map(shape)];
}
for (const [index, {title, message}] of pages.entries()) {
  const expected = JSON.stringify(["main", ["section", ["h1", title], ["p", message]]]);
  for (const {run} of modules[index]) {
    if (run() !== 0 || JSON.stringify(shape(target)) !== expected) {
      throw new Error(`${title} rendered a different DOM structure or text`);
    }
  }
}
console.log("DOM checks passed for both pages in all four builds.");
const [hand, runeInline, runeCSS] = modules[0];
const runs = [hand, runeInline, runeCSS];
for (const {run} of runs) {
  for (let i = 0; i < Math.min(20_000, iterations); i++) run();
}
const samples = [[], [], []];
for (let round = 0; round < runsCount; round++) {
  for (let n = 0; n < runs.length; n++) {
    const pos = (round + n) % runs.length;
    const {run} = runs[pos];
    const start = performance.now();
    for (let i = 0; i < iterations; i++) run();
    samples[pos].push((performance.now() - start) * 1_000_000 / iterations);
  }
}
const median = values => {
  values.sort((a, b) => a - b);
  return Math.round((values[Math.floor((values.length - 1) / 2)] + values[Math.floor(values.length / 2)]) / 2);
};
console.log(`Fake DOM median ns/mount (raw / RuneScript inline / RuneScript extracted, ${runsCount} × ${iterations} mounts): ${samples.map(median).join(" / ")}`);
