// Measures generated JavaScript for the size benchmarks: raw bytes, and
// gzip (level 9) and Brotli (quality 11) of both the raw and the minified
// output. Minification uses the pinned esbuild with the same settings for
// every file. Usage: node sizes.mjs <file.js>...  (prints one JSON line per
// file; "min" fields are -1 when esbuild is not installed).
import { readFileSync } from "node:fs";
import { gzipSync, brotliCompressSync, constants } from "node:zlib";

let esbuild = null;
try {
  esbuild = await import("esbuild");
} catch {
  // Not installed: report raw sizes only.
}

const gz = (b) => gzipSync(b, { level: 9 }).length;
const br = (b) =>
  brotliCompressSync(b, {
    params: { [constants.BROTLI_PARAM_QUALITY]: 11, [constants.BROTLI_PARAM_SIZE_HINT]: b.length },
  }).length;

for (const file of process.argv.slice(2)) {
  const raw = readFileSync(file);
  const r = { file, raw: raw.length, raw_gzip: gz(raw), raw_brotli: br(raw), min: -1, min_gzip: -1, min_brotli: -1 };
  if (esbuild) {
    const { code } = await esbuild.transform(raw.toString("utf8"), {
      loader: "js",
      minify: true,
      legalComments: "none",
    });
    const min = Buffer.from(code);
    r.min = min.length;
    r.min_gzip = gz(min);
    r.min_brotli = br(min);
  }
  process.stdout.write(JSON.stringify(r) + "\n");
}
