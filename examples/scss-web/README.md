# Native SCSS web build

```sh
go run ./cmd/runescript build --target web examples/scss-web
```

Open `dist/index.html` from a static server. The web builder compiles
`styles/app.scss` and writes its content-hashed CSS under `dist/__runescript/`.
Use `--source-map` to publish a linked CSS source map.

Use the optional `styles` array in `runescript.web.json` to list SCSS
entries in cascade order. Each entry must be a project-relative `.scss`
path. File imports are resolved from the importing stylesheet and must
stay in the project. Each managed HTML page or docs layout uses
`<!-- runescript -->` in the head to receive the shared styles.

The compiler preserves CSS `url()` references. Use root-relative URLs
for files copied from `public/` if they would otherwise resolve relative
to the generated CSS directory.

RuneScript embeds a [pure-Go Sass compiler](https://github.com/shyim/go-sass).
Run the example with the regular RuneScript CLI. The library implements
the Sass features used here and has its own compatibility coverage.
