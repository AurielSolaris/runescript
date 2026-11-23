package jsgen

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ir"
)

const styleAttribute = "data-runescript-style"

// Extract only declaration lists made of simple tokens. Functions, strings,
// comments, escapes, and !important stay inline. In particular, relative
// CSS URLs must keep resolving against the original document.
var simpleDeclaration = regexp.MustCompile(`^[A-Za-z-][A-Za-z0-9-]*[ \t]*:[ \t]*[A-Za-z0-9#.,%+/ \t\r\n-]+$`)

func simpleStyle(value string) string {
	value = strings.TrimSuffix(strings.TrimSpace(value), ";")
	if value == "" {
		return ""
	}
	for decl := range strings.SplitSeq(value, ";") {
		if !simpleDeclaration.MatchString(strings.TrimSpace(decl)) {
			return ""
		}
	}
	return value
}

func (g *gen) extractedStyle(e *ir.TElement) string {
	if g.styles == nil {
		return ""
	}
	var declarations string
	for _, a := range e.Attrs {
		switch {
		case strings.EqualFold(a.Name, styleAttribute):
			return ""
		case a.Name == "style" && a.Value == nil && !a.Present:
			declarations = simpleStyle(a.Static)
		}
	}
	if declarations == "" {
		return ""
	}
	if name, ok := g.styles[declarations]; ok {
		return name
	}
	sum := sha256.Sum256([]byte(declarations))
	name := fmt.Sprintf("%x", sum[:12])
	g.styles[declarations] = name
	fmt.Fprintf(&g.css, "[%s=\"%s\"] { %s; }\n", styleAttribute, name, declarations)
	return name
}
