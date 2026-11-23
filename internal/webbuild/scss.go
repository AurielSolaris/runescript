package webbuild

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	sass "github.com/shyim/go-sass"
)

type scssImporter struct {
	*sass.FsImporter
	root *os.Root
}

func (i scssImporter) Load(url sass.CanonicalURL) (*sass.ImporterResult, bool, error) {
	name, err := filepath.Rel(i.root.Name(), url.String())
	if err != nil || !filepath.IsLocal(name) {
		return nil, false, fmt.Errorf("stylesheet import is outside the project: %s", url.String())
	}
	source, err := readStylesheet(i.root, name)
	if err != nil {
		return nil, false, err
	}
	syntax := sass.SyntaxSCSS
	switch strings.ToLower(filepath.Ext(name)) {
	case ".sass":
		syntax = sass.SyntaxSass
	case ".css":
		syntax = sass.SyntaxCSS
	}
	return &sass.ImporterResult{Contents: source, Syntax: syntax}, true, nil
}

func readStylesheet(root *os.Root, name string) (string, error) {
	info, err := root.Stat(name)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s: stylesheet must be a regular file", name)
	}
	source, err := root.ReadFile(name)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(source) {
		return "", fmt.Errorf("%s: stylesheet must be UTF-8", name)
	}
	return string(source), nil
}

func buildStyles(root, stage string, entries []string, opts Options, assets map[string]bool) ([]string, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	project, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer project.Close()
	importer := scssImporter{FsImporter: sass.NewFsImporter(nil), root: project}
	var styles []string
	seen := make(map[string]bool)
	for _, entry := range entries {
		if !validPath(entry, ".scss") {
			return nil, fmt.Errorf("invalid stylesheet %q: expected a project-relative .scss path", entry)
		}
		if seen[entry] {
			continue
		}
		seen[entry] = true
		source, err := readStylesheet(project, entry)
		if err != nil {
			return nil, err
		}
		sourcePath := filepath.Join(root, filepath.FromSlash(entry))
		options := sass.NewOptions().WithStyle(sass.StyleCompressed).
			WithImporter(importer).WithURL(sourcePath)
		if opts.Notes != nil {
			options.WithWarnHandler(func(event *sass.WarnEvent) { fmt.Fprintln(opts.Notes, event.Message) })
		}
		var css string
		if opts.SourceMap {
			// Expanded output preserves mappings across rules from different
			// files. The compiler's compressed output collapses them to one line.
			result, err := sass.CompileWithSourceMap(source, options.WithStyle(sass.StyleExpanded).WithSourceMapIncludeSources(true))
			if err != nil {
				return nil, fmt.Errorf("compile %s: %w", entry, err)
			}
			result.SourceMap.HasFile = false
			for j, name := range result.SourceMap.Sources {
				if !filepath.IsAbs(name) {
					name = filepath.Join(filepath.Dir(sourcePath), filepath.FromSlash(name))
				}
				relative, err := filepath.Rel(root, name)
				if err == nil && filepath.IsLocal(relative) {
					result.SourceMap.Sources[j] = filepath.ToSlash(relative)
				}
			}
			data := []byte(result.SourceMap.ToJSON())
			hash := sha256.Sum256(data)
			name := fmt.Sprintf("scss-%x.css.map", hash[:12])
			if err := writeAsset(stage, path.Join(generated, name), data, assets); err != nil {
				return nil, err
			}
			css = result.CSS + "\n/*# sourceMappingURL=" + name + " */\n"
		} else {
			css, err = sass.Compile(source, options)
			if err != nil {
				return nil, fmt.Errorf("compile %s: %w", entry, err)
			}
		}
		hash := sha256.Sum256([]byte(css))
		name := path.Join(generated, fmt.Sprintf("scss-%x.css", hash[:12]))
		if err := writeAsset(stage, name, []byte(css), assets); err != nil {
			return nil, err
		}
		styles = append(styles, name)
	}
	return styles, nil
}

func writeStyleLinks(tags *strings.Builder, page string, styles []string) {
	for _, style := range styles {
		fmt.Fprintf(tags, `<link rel="stylesheet" href="%s">`+"\n", relativeURL(page, style))
	}
}
