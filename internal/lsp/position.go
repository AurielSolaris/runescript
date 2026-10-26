package lsp

import (
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"
)

// The compiler's positions are byte offsets into UTF-8 text. LSP positions
// are a line and a character, counted in UTF-16 code units unless the
// client accepts UTF-8 (positionEncoding "utf-8"), in which case they are
// bytes.

// toPosition converts the byte offset off in text to an LSP position.
func toPosition(text []byte, off int, utf16 bool) Position {
	if off > len(text) {
		off = len(text)
	}
	if off < 0 {
		off = 0
	}
	line, start := 0, 0
	for i := 0; i < off; i++ {
		if text[i] == '\n' {
			line++
			start = i + 1
		}
	}
	if !utf16 {
		return Position{Line: line, Character: off - start}
	}
	return Position{Line: line, Character: utf16Len(text[start:off])}
}

// toOffset converts an LSP position to a byte offset in text, clamped to
// the line's end.
func toOffset(text []byte, p Position, utf16 bool) int {
	off := 0
	for line := 0; line < p.Line; line++ {
		i := strings.IndexByte(string(text[off:]), '\n')
		if i < 0 {
			return len(text)
		}
		off += i + 1
	}
	end := off
	for end < len(text) && text[end] != '\n' {
		end++
	}
	if !utf16 {
		return min(off+p.Character, end)
	}
	units := 0
	for off < end && units < p.Character {
		r, size := utf8.DecodeRune(text[off:])
		if r >= 0x10000 {
			units += 2
		} else {
			units++
		}
		off += size
	}
	return off
}

func utf16Len(b []byte) int {
	n := 0
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
		b = b[size:]
	}
	return n
}

// uriToPath converts a file URI to a file system path. Other URIs (an
// unsaved "untitled:" document) give "".
func uriToPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return ""
	}
	p := u.Path
	if runtime.GOOS == "windows" {
		// file:///c:/x → c:/x; file://server/share/x → //server/share/x
		if u.Host != "" {
			p = "//" + u.Host + p
		} else if len(p) >= 3 && p[0] == '/' && p[2] == ':' {
			p = strings.ToUpper(p[1:2]) + p[2:] // as filepath.Abs writes it
		}
	}
	return filepath.Clean(filepath.FromSlash(p))
}

// pathToURI converts an absolute path to a file URI, written as VS Code
// writes them: a lowercase drive letter, with its ':' escaped.
func pathToURI(path string) string {
	p := filepath.ToSlash(path)
	if strings.HasPrefix(p, "//") { // a UNC path: //server/share/x
		return "file:" + p
	}
	var b strings.Builder
	b.WriteString("file://")
	if !strings.HasPrefix(p, "/") {
		b.WriteByte('/')
	}
	for i, seg := range strings.Split(p, "/") {
		if i > 0 {
			b.WriteByte('/')
		}
		if i == 0 && len(seg) == 2 && seg[1] == ':' {
			b.WriteString(strings.ToLower(seg[:1]) + "%3A")
			continue
		}
		b.WriteString(url.PathEscape(seg))
	}
	return b.String()
}

// pathKey is path as a map key: Windows paths compare without regard to
// case.
func pathKey(path string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}
