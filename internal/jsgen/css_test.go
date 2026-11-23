package jsgen

import (
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/ir"
)

func TestExtractedStyle(t *testing.T) {
	g := &gen{styles: map[string]string{}}
	makeAttr := func(s string) *ir.TAttr { return &ir.TAttr{Name: "style", Static: s} }
	s := "color: red; padding: 4px"
	element := &ir.TElement{Attrs: []*ir.TAttr{makeAttr(s)}}
	id := g.extractedStyle(element)
	if id == "" || g.extractedStyle(element) != id || strings.Count(g.css.String(), s) != 1 {
		t.Fatalf("style wasn't extracted and deduplicated: %q", g.css.String())
	}
	for _, a := range []*ir.TAttr{
		makeAttr("background: url(./image.png)"),
		makeAttr("color: red !important"),
		makeAttr("color: red} body {display:none"),
		{Name: "style", Value: &ir.StringConst{Value: s}},
	} {
		if g.extractedStyle(&ir.TElement{Attrs: []*ir.TAttr{a}}) != "" {
			t.Errorf("extracted a style that must stay inline: %+v", a)
		}
	}
}
