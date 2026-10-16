# Localization

English is the original language of RuneScript's documentation. The
English files ([`README.md`](../README.md) and [`docs/`](../docs/)) are the
source of truth; this folder holds translations of them, one folder per
language, named with a short code chosen by the project: `bn` for Bengali
and `heb` for Hebrew. (Web pages still declare Hebrew with the standard
language tag, `lang="he"`, which browsers and screen readers expect.)

| Language | Folder | Translated | Maintained by |
|---|---|---|---|
| বাংলা (Bengali) | [`bn/`](bn/README.md) | README, the guide (`docs/guide/`) | AI, checked by hand by the author |
| עברית (Hebrew) | [`heb/`](heb/README.md) | README, the guide (`docs/guide/`) | AI, checked by hand by the author |

**Status:** the README, the guide index, and all 15 chapters are translated
in both languages.

**Help is appreciated.** If you speak one of these languages, corrections
and better wording are very welcome; so are new languages. Open an issue or
a pull request.

## What is translated

- The README and every chapter of the guide.
- Not the [language specification](../docs/LANGUAGE_SPEC.md) or the
  [roadmap](../docs/ROADMAP.md): the specification is normative and changes
  often, so it stays in English, and the translations link to it.
- Not code: programs, commands, file names, keywords, diagnostic codes, and
  compiler messages stay exactly as in English. Every code block in a
  translation must be identical to the English one, in the same order;
  `go test ./tests` checks this, and also compiles every program in the
  translated guides.

## How a translation is laid out

A translation mirrors the English paths: `docs/guide/03-types-and-variables.md`
is translated in `localization/bn/docs/guide/03-types-and-variables.md`.
Links between translated pages stay relative; links to untranslated files
point at the English ones.

Each translated file starts with a comment naming the English file and the
commit it was translated from:

```html
<!-- source: docs/guide/03-types-and-variables.md @ 4f1c2ab -->
```

and with a short note, in the translation's language, that the English
version is authoritative.

## Keeping translations up to date

```sh
scripts/translation-status.sh        # lists translations whose English source changed since
```

When an English file changes, update its translations and the commit in
their `source:` comment. If a translation can't be updated right away, it
stays as it is: the note at its top points readers to the English original.

## Adding a language

1. Create `localization/<code>/` and translate `README.md` and
   `docs/guide/` (keeping every code block unchanged).
2. Add the language to the tables here and in the main README.
3. Run `go test ./tests`.

Right-to-left languages such as Hebrew need nothing special in the
Markdown; code stays left to right.
