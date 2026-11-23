# Static English documentation

From the RuneScript checkout, build the example into a new directory.

```sh
go run ./cmd/runescript build --target web examples/static-docs
```

The build writes `dist/docs.html` for `/docs` and
`dist/docs/getting-started.html` for `/docs/getting-started`. Configure
your static host to serve extensionless HTML URLs. Both pages share the
stylesheet from `public/` and work without JavaScript.
The [Astro comparison](../../bench/README.md#static-markdown-benchmark)
builds the same example with both tools.

The optional `docs` section of `runescript.web.json` names the project-relative
Markdown directory and the shared HTML layout, and supplies the index title
and description. Each `<lowercase-slug>.md` becomes one chapter. The YAML
front matter requires non-empty strings for `title`, `description`,
`summary`, `section`, and `source`, plus a signed 32-bit integer `order`.
Use `Guide`, `Reference`, or `Project` for `section`. The index sorts by
`order` and breaks ties by URL.

Layouts use Go's build-time HTML templates. Use `.IsIndex` to choose the
page layout, `.Chapters` for navigation, `.Headings` for links to level-two
and level-three headings, and `.Content` for rendered Markdown. Metadata is
escaped as text, and raw HTML in Markdown is omitted. A project can mix
this collection with existing RuneScript browser entries under `pages`.

## Optional syntax highlighting and copy buttons

From `examples/static-docs/tooling`, run:

```sh
bun install --frozen-lockfile
bun run build
```

The script runs the RuneScript web build and publishes the result to
`dist-enhanced/`. RuneScript compiles SCSS natively to one shared stylesheet.
Shiki highlights the chapter's examples during the build and uses the sibling site's RuneScript
grammar when available. Set `RUNESCRIPT_SITE` to point to that checkout.
Without it, the example uses Shiki's ActionScript grammar.

Bun resolves and bundles the TypeScript modules for the chapter's Alpine
copy buttons. The index uses the shared CSS and loads no JavaScript.
The chapter bundle measured about 20 KiB gzipped in the example build.
Use `bun run build --source-map` to opt into linked CSS and browser source maps.
The plain RuneScript build above needs no JavaScript tooling.
