package jsgen

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

// Source maps (Source Map Revision 3). While generating, the line of each
// statement and function starts with a marker naming its source position:
// markOpen, the index of the position in gen.marks, markClose. Generated
// JavaScript never contains these control characters otherwise (quote
// escapes them), so markers survive code that is generated into separate
// buffers and pasted in later, such as lambdas. sourceMap removes them
// and records where they were.

const (
	markOpen  = '\x00'
	markClose = '\x01'
)

// unmapped is a marker that ends the previous mapping: the code after it
// comes from no RuneScript source.
const unmapped = string(markOpen) + "-" + string(markClose)

// marker returns the marker for the start of span.
func (g *gen) marker(span diagnostics.Span) string {
	if span.Start.File == nil {
		return ""
	}
	g.marks = append(g.marks, span.Start)
	return string(markOpen) + strconv.Itoa(len(g.marks)-1) + string(markClose)
}

// mapping is one segment of a source map: generated column and line
// (0-based, UTF-16), and the source position (no File: unmapped).
type mapping struct {
	line, col int
	pos       diagnostics.Pos
}

// unmark removes the markers from js and returns where they were.
func unmark(js string, marks []diagnostics.Pos) (string, []mapping) {
	if strings.IndexByte(js, markOpen) < 0 {
		return js, nil
	}
	var out strings.Builder
	var maps []mapping
	line, lineStart := 0, 0 // lineStart: offset in out of the current line
	for i := 0; i < len(js); {
		c := js[i]
		switch c {
		case markOpen:
			end := strings.IndexByte(js[i:], markClose)
			col := utf16Len(out.String()[lineStart:])
			mp := mapping{line: line, col: col}
			if n, err := strconv.Atoi(js[i+1 : i+end]); err == nil {
				mp.pos = marks[n]
			}
			maps = append(maps, mp)
			i += end + 1
			continue
		case '\n':
			line++
			out.WriteByte(c)
			lineStart = out.Len()
		default:
			out.WriteByte(c)
		}
		i++
	}
	return out.String(), maps
}

// utf16Len is the length of s in UTF-16 code units.
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n++
		if r >= 0x10000 {
			n++
		}
	}
	return n
}

// sourceCol is the 0-based UTF-16 column of pos in its file.
func sourceCol(pos diagnostics.Pos) int {
	text := pos.File.Text
	if pos.Offset > len(text) {
		return pos.Col - 1
	}
	start := bytes.LastIndexByte(text[:pos.Offset], '\n') + 1
	if !utf8.Valid(text[start:pos.Offset]) {
		return pos.Col - 1
	}
	return utf16Len(string(text[start:pos.Offset]))
}

type sourceMapJSON struct {
	Version        int      `json:"version"`
	File           string   `json:"file,omitempty"`
	Sources        []string `json:"sources"`
	SourcesContent []string `json:"sourcesContent"`
	Names          []string `json:"names"`
	Mappings       string   `json:"mappings"`
}

// buildMap encodes maps as a source map. file is the generated file's
// name; sourcePath, if not nil, turns a source's name into the path the
// map records.
func buildMap(maps []mapping, file string, sourcePath func(string) string) []byte {
	m := sourceMapJSON{Version: 3, File: file, Sources: []string{}, SourcesContent: []string{}, Names: []string{}}
	index := map[*diagnostics.Source]int{}
	var b strings.Builder
	line := 0
	prevCol, prevSrc, prevLine, prevSrcCol := 0, 0, 0, 0
	first := true
	for _, mp := range maps {
		for line < mp.line {
			b.WriteByte(';')
			line++
			prevCol, first = 0, true
		}
		if !first {
			b.WriteByte(',')
		}
		first = false
		vlq(&b, mp.col-prevCol)
		prevCol = mp.col
		if mp.pos.File == nil {
			continue // a one-field segment: unmapped
		}
		src, ok := index[mp.pos.File]
		if !ok {
			src = len(m.Sources)
			index[mp.pos.File] = src
			name := mp.pos.File.Name
			if sourcePath != nil {
				name = sourcePath(name)
			}
			m.Sources = append(m.Sources, name)
			m.SourcesContent = append(m.SourcesContent, string(mp.pos.File.Text))
		}
		srcLine, srcCol := mp.pos.Line-1, sourceCol(mp.pos)
		vlq(&b, src-prevSrc)
		vlq(&b, srcLine-prevLine)
		vlq(&b, srcCol-prevSrcCol)
		prevSrc, prevLine, prevSrcCol = src, srcLine, srcCol
	}
	m.Mappings = b.String()
	out, _ := json.Marshal(m)
	return out
}

const base64Digits = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// vlq appends n in the source map's base64 VLQ encoding.
func vlq(b *strings.Builder, n int) {
	v := n << 1
	if n < 0 {
		v = (-n << 1) | 1
	}
	for {
		digit := v & 31
		v >>= 5
		if v > 0 {
			digit |= 32
		}
		b.WriteByte(base64Digits[digit])
		if v == 0 {
			return
		}
	}
}

// inlineMapURL is a data: URL holding the map, for sourceMappingURL.
func inlineMapURL(m []byte) string {
	return "data:application/json;charset=utf-8;base64," + base64.StdEncoding.EncodeToString(m)
}
