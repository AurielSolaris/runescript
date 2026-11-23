// Package webbuild builds static, multi-page sites from RuneScript programs.
package webbuild

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/jsgen"
)

const (
	Manifest  = "runescript.web.json"
	generated = "__runescript"
	marker    = "<!-- runescript -->"
)

type manifest struct {
	Pages  map[string]string `json:"pages"`
	Docs   *docsConfig       `json:"docs"`
	Styles []string          `json:"styles"`
}

type Options struct {
	Version    string
	Warnings   driver.WarningPolicy
	Notes      io.Writer
	SourceMap  bool
	ExtractCSS bool
}

// Build assembles a site in a new output directory. On a failed build it
// leaves the destination untouched.
func Build(root, out string, opts Options) ([]*diagnostics.Diagnostic, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	public := filepath.Join(root, "public")
	publicDir, err := os.Lstat(public)
	if err != nil {
		return nil, err
	}
	if !publicDir.IsDir() {
		return nil, fmt.Errorf("web project public path must be a directory: %q", public)
	}
	config, err := readManifest(filepath.Join(root, Manifest))
	if err != nil {
		return nil, err
	}
	for page, entry := range config.Pages {
		if !validPath(page, ".html") || !validPath(entry, ".rsc") || strings.EqualFold(strings.Split(page, "/")[0], generated) {
			return nil, fmt.Errorf("invalid web page %q or entry %q: use project-relative .html and .rsc paths", page, entry)
		}
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return nil, err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(out))
	if err != nil {
		return nil, err
	}
	out = filepath.Join(parent, filepath.Base(out))
	if rel, err := filepath.Rel(public, out); err == nil && (rel == "." || filepath.IsLocal(rel)) {
		return nil, fmt.Errorf("web output %q must be outside public", out)
	}
	if _, err := os.Lstat(out); err == nil {
		return nil, fmt.Errorf("web output %q already exists; choose a new directory", out)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	stage, err := os.MkdirTemp(parent, ".runescript-web-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stage)
	if err := copyPublic(public, stage); err != nil {
		return nil, err
	}
	seenPages := map[string]bool{}
	assets := map[string]bool{}
	styles, err := buildStyles(root, stage, config.Styles, opts, assets)
	if err != nil {
		return nil, err
	}
	var pages []string
	for page := range config.Pages {
		if key := strings.ToLower(page); seenPages[key] {
			return nil, fmt.Errorf("web pages differ only in case: %q", page)
		} else {
			seenPages[key] = true
		}
		pages = append(pages, page)
	}
	slices.Sort(pages)
	for _, page := range pages {
		if ds, err := buildPage(root, stage, out, page, config.Pages[page], opts, assets, styles); err != nil || len(ds) > 0 {
			return ds, err
		}
	}
	if config.Docs != nil {
		if err := buildDocs(root, stage, *config.Docs, styles); err != nil {
			return nil, err
		}
	}
	return nil, os.Rename(stage, out)
}

var portablePath = regexp.MustCompile(`^[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*$`)

func validPath(name, ext string) bool {
	return path.Ext(name) == ext && path.Clean(name) == name && portablePath.MatchString(name) && filepath.IsLocal(filepath.FromSlash(name))
}

func readManifest(name string) (manifest, error) {
	f, err := os.Open(name)
	if err != nil {
		return manifest{}, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	var m manifest
	if err := d.Decode(&m); err != nil {
		return m, fmt.Errorf("%s: %w", name, err)
	}
	if err := d.Decode(new(json.RawMessage)); err != io.EOF {
		return m, fmt.Errorf("%s: expected a single JSON object", name)
	}
	if len(m.Pages) == 0 && m.Docs == nil {
		return m, fmt.Errorf("%s: configure pages or a docs collection", name)
	}
	return m, nil
}

func buildPage(root, stage, out, page, entry string, opts Options, assets map[string]bool, styles []string) ([]*diagnostics.Diagnostic, error) {
	html, err := os.ReadFile(filepath.Join(stage, filepath.FromSlash(page)))
	if err != nil {
		return nil, fmt.Errorf("page %s: %w", page, err)
	}
	if bytes.Count(html, []byte(marker)) != 1 {
		return nil, fmt.Errorf("page %s needs exactly one %s in its head", page, marker)
	}
	base := path.Join(generated, "pages", strings.TrimSuffix(page, ".html"))
	module := base + ".js"
	output, ds, err := driver.BuildJSMapInput(driver.Input{
		Path: filepath.Join(root, filepath.FromSlash(entry)), Dir: root,
		Notes: opts.Notes, Warnings: opts.Warnings,
	}, jsgen.Options{
		Version: opts.Version, SourceMap: opts.SourceMap, File: path.Base(module), MapURL: path.Base(module) + ".map",
		SourcePath: func(name string) string {
			source := filepath.FromSlash(name)
			if !filepath.IsAbs(source) {
				source = filepath.Join(root, source)
			}
			rel, err := filepath.Rel(filepath.Join(out, filepath.FromSlash(path.Dir(module))), source)
			if err != nil {
				return filepath.ToSlash(source)
			}
			return filepath.ToSlash(rel)
		},
		RuntimePrefix: strings.TrimSuffix(relativeURL(module, generated+"/runtime.js"), "runtime.js"),
		ImportPath: func(spec string) string {
			if strings.HasPrefix(spec, "./") {
				return strings.Repeat("../", strings.Count(module, "/")) + strings.TrimPrefix(spec, "./")
			}
			return spec
		},
		BrowserEntry: true, ExtractCSS: opts.ExtractCSS,
	})
	if err != nil || len(ds) > 0 {
		return ds, err
	}
	runtimePath := path.Join(generated, output.RuntimeFile)
	if err := writeAsset(stage, runtimePath, output.Runtime, assets); err != nil {
		return nil, err
	}
	if err := writeOutput(stage, module, output.JS); err != nil {
		return nil, err
	}
	if opts.SourceMap {
		if err := writeOutput(stage, module+".map", output.Map); err != nil {
			return nil, err
		}
	}
	var tags strings.Builder
	writeStyleLinks(&tags, page, styles)
	if len(output.CSS) > 0 {
		sum := sha256.Sum256(output.CSS)
		css := path.Join(generated, fmt.Sprintf("styles-%x.css", sum[:12]))
		if err := writeAsset(stage, css, output.CSS, assets); err != nil {
			return nil, err
		}
		fmt.Fprintf(&tags, `<link rel="stylesheet" href="%s">`+"\n", relativeURL(page, css))
	}
	fmt.Fprintf(&tags, `<script type="module" src="%s"></script>`, relativeURL(page, module))
	return nil, writeOutput(stage, page, bytes.Replace(html, []byte(marker), []byte(tags.String()), 1))
}

func writeAsset(root, name string, content []byte, assets map[string]bool) error {
	if assets[name] {
		return nil
	}
	if err := writeOutput(root, name, content); err != nil {
		return err
	}
	assets[name] = true
	return nil
}

func relativeURL(from, to string) string {
	rel, err := filepath.Rel(filepath.FromSlash(path.Dir(from)), filepath.FromSlash(to))
	if err != nil {
		panic(err) // both paths are relative and use portable path names
	}
	url := filepath.ToSlash(rel)
	if !strings.HasPrefix(url, "../") {
		url = "./" + url
	}
	return url
}

func writeOutput(root, name string, content []byte) error {
	dst := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0o644)
}

func copyPublic(root, out string) error {
	return filepath.WalkDir(root, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, file)
		if err != nil {
			return err
		}
		if strings.EqualFold(strings.Split(filepath.ToSlash(rel), "/")[0], generated) {
			return fmt.Errorf("public path %q is reserved for generated modules", rel)
		}
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(out, rel), 0o755)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("public file %q must be a regular file", rel)
		}
		in, err := os.Open(file)
		if err != nil {
			return err
		}
		defer in.Close()
		dst, err := os.Create(filepath.Join(out, rel))
		if err != nil {
			return err
		}
		_, err = io.Copy(dst, in)
		return errors.Join(err, dst.Close())
	})
}
