package webbuild

import (
	"bytes"
	"cmp"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.yaml.in/yaml/v3"
)

type docsConfig struct {
	Content     string `json:"content"`
	Layout      string `json:"layout"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type chapterLink struct {
	URL, Title, Description, Summary, Section, Source string
	Order                                             int32
}

type headingLink struct {
	ID, Text string
	Level    int
}

type docsPage struct {
	chapterLink
	IsIndex  bool
	Chapters []chapterLink
	Headings []headingLink
	Content  template.HTML
	Styles   template.HTML
}

var chapterSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func buildDocs(root, stage string, cfg docsConfig, styles []string) error {
	if !portablePath.MatchString(cfg.Content) || !fs.ValidPath(cfg.Content) ||
		cfg.Content == "." || !validPath(cfg.Layout, ".html") {
		return fmt.Errorf("docs content and layout must be project-relative paths")
	}
	if strings.TrimSpace(cfg.Title) == "" || strings.TrimSpace(cfg.Description) == "" {
		return fmt.Errorf("docs index needs a title and description")
	}
	project, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer project.Close()
	layoutText, err := project.ReadFile(cfg.Layout)
	if err != nil {
		return err
	}
	if len(styles) > 0 {
		if bytes.Count(layoutText, []byte(marker)) != 1 {
			return fmt.Errorf("%s needs exactly one %s in its head for shared styles", cfg.Layout, marker)
		}
		layoutText = bytes.Replace(layoutText, []byte(marker), []byte("{{.Styles}}"), 1)
	}
	layout, err := template.New(cfg.Layout).Parse(string(layoutText))
	if err != nil {
		return err
	}
	entries, err := fs.ReadDir(project.FS(), cfg.Content)
	if err != nil {
		return err
	}
	markdown := goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithParserOptions(parser.WithAutoHeadingID()))
	index := docsPage{
		chapterLink: chapterLink{URL: "/docs", Title: cfg.Title, Description: cfg.Description},
		IsIndex:     true,
	}
	var chapters []docsPage
	for _, entry := range entries {
		ext := path.Ext(entry.Name())
		if !strings.EqualFold(ext, ".md") {
			continue
		}
		slug := strings.TrimSuffix(entry.Name(), ext)
		name := path.Join(cfg.Content, entry.Name())
		if ext != ".md" || !chapterSlug.MatchString(slug) || slug == "index" || !validPath(slug+".html", ".html") || !entry.Type().IsRegular() {
			return fmt.Errorf("%s: expected a regular <lowercase-slug>.md chapter other than index.md", name)
		}
		source, err := project.ReadFile(name)
		if err != nil {
			return err
		}
		chapter, err := readChapter(name, source, markdown)
		if err != nil {
			return err
		}
		chapter.URL = "/docs/" + slug
		chapters = append(chapters, chapter)
	}
	if len(chapters) == 0 {
		return fmt.Errorf("%s: no Markdown chapters found", cfg.Content)
	}
	slices.SortFunc(chapters, func(a, b docsPage) int {
		if order := cmp.Compare(a.Order, b.Order); order != 0 {
			return order
		}
		return strings.Compare(a.URL, b.URL)
	})
	for _, chapter := range chapters {
		index.Chapters = append(index.Chapters, chapter.chapterLink)
	}

	public := make(map[string]string)
	err = fs.WalkDir(os.DirFS(stage), ".", func(name string, entry fs.DirEntry, err error) error {
		if err == nil {
			public[strings.ToLower(name)] = name
		}
		return err
	})
	if err != nil {
		return err
	}
	pages := append([]docsPage{index}, chapters...)
	for _, page := range pages {
		name := strings.TrimPrefix(page.URL, "/")
		for _, candidate := range []string{name + ".html", path.Join(name, "index.html")} {
			if existing, found := public[candidate]; found {
				return fmt.Errorf("docs route %s conflicts with public path %q", page.URL, existing)
			}
		}
		if existing, found := public[name]; found {
			info, err := fs.Stat(os.DirFS(stage), existing)
			if err != nil {
				return err
			}
			if !page.IsIndex || existing != name || !info.IsDir() {
				return fmt.Errorf("docs route %s conflicts with public path %q", page.URL, existing)
			}
		}
	}
	var output bytes.Buffer
	for _, page := range pages {
		page.Chapters = index.Chapters
		name := strings.TrimPrefix(page.URL, "/") + ".html"
		var tags strings.Builder
		writeStyleLinks(&tags, name, styles)
		// The link URLs are generated from portable paths and content hashes.
		page.Styles = template.HTML(tags.String())
		output.Reset()
		if err := layout.Execute(&output, page); err != nil {
			return fmt.Errorf("render %s: %w", page.URL, err)
		}
		if err := writeOutput(stage, name, output.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func readChapter(name string, source []byte, markdown goldmark.Markdown) (docsPage, error) {
	var page docsPage
	if !utf8.Valid(source) {
		return page, fmt.Errorf("%s: Markdown must be UTF-8", name)
	}
	normalized := strings.ReplaceAll(string(source), "\r\n", "\n")
	rest, ok := strings.CutPrefix(normalized, "---\n")
	metadata, body, closed := strings.Cut(rest, "\n---\n")
	if !ok || !closed {
		return page, fmt.Errorf("%s:1: expected YAML front matter delimited by ---", name)
	}
	if err := readChapterMetadata(metadata, &page.chapterLink); err != nil {
		return page, fmt.Errorf("%s:%w", name, err)
	}
	src := []byte(body)
	doc := markdown.Parser().Parse(text.NewReader(src))
	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if h, ok := node.(*ast.Heading); ok && entering && (h.Level == 2 || h.Level == 3) {
			rawID, _ := h.AttributeString("id")
			if id, ok := rawID.([]byte); ok {
				page.Headings = append(page.Headings, headingLink{ID: string(id), Text: string(h.Text(src)), Level: h.Level})
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return page, fmt.Errorf("%s: %w", name, err)
	}
	var rendered bytes.Buffer
	if err := markdown.Renderer().Render(&rendered, src, doc); err != nil {
		return page, fmt.Errorf("%s: %w", name, err)
	}
	// Goldmark renders text and omits raw HTML and dangerous link destinations.
	page.Content = template.HTML(rendered.String())
	return page, nil
}

func readChapterMetadata(source string, link *chapterLink) error {
	decoder := yaml.NewDecoder(strings.NewReader(source))
	var doc, extra yaml.Node
	if err := decoder.Decode(&doc); err != nil {
		return fmt.Errorf("2: invalid YAML front matter: %w", err)
	}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode || decoder.Decode(&extra) != io.EOF {
		return fmt.Errorf("2: expected a single mapping in front matter")
	}
	fields := map[string]*string{
		"title": &link.Title, "description": &link.Description, "summary": &link.Summary,
		"section": &link.Section, "source": &link.Source,
	}
	seen := make(map[string]bool)
	for value := range slices.Chunk(doc.Content[0].Content, 2) {
		field, node := value[0], value[1]
		if field.Kind != yaml.ScalarNode || field.ShortTag() != "!!str" || seen[field.Value] {
			return fmt.Errorf("%d: expected a unique front-matter field name", field.Line+1)
		}
		seen[field.Value] = true
		if field.Value == "order" {
			if node.Kind != yaml.ScalarNode || node.ShortTag() != "!!int" || node.Decode(&link.Order) != nil {
				return fmt.Errorf("%d: order must be a signed 32-bit integer", node.Line+1)
			}
			continue
		}
		target, ok := fields[field.Value]
		if !ok {
			return fmt.Errorf("%d: unknown front-matter field %q", field.Line+1, field.Value)
		}
		if node.Kind != yaml.ScalarNode || node.ShortTag() != "!!str" || strings.TrimSpace(node.Value) == "" {
			return fmt.Errorf("%d: %s must be a non-empty string", node.Line+1, field.Value)
		}
		if field.Value == "section" && node.Value != "Guide" && node.Value != "Reference" && node.Value != "Project" {
			return fmt.Errorf("%d: section must be Guide, Reference, or Project", node.Line+1)
		}
		*target = node.Value
	}
	for _, required := range []string{"title", "description", "summary", "section", "order", "source"} {
		if !seen[required] {
			return fmt.Errorf("2: missing front-matter field %q", required)
		}
	}
	return nil
}
