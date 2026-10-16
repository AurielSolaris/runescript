package dom

import (
	"sort"
	"strings"
	"testing"
)

// These lists are written independently of elements.go, from the HTML 4.01
// element index and the HTML living standard's element index, so that a
// slip in one shows up as a difference.
const (
	html401 = `a abbr acronym address applet area b base basefont bdo big blockquote body br button
		caption center cite code col colgroup dd del dfn dir div dl dt em fieldset font form frame
		frameset h1 h2 h3 h4 h5 h6 head hr html i iframe img input ins isindex kbd label legend li
		link map menu meta noframes noscript object ol optgroup option p param pre q s samp script
		select small span strike strong style sub sup table tbody td textarea tfoot th thead title
		tr tt u ul var`
	living = `a abbr address area article aside audio b base bdi bdo blockquote body br button canvas
		caption cite code col colgroup data datalist dd del details dfn dialog div dl dt em embed
		fieldset figcaption figure footer form h1 h2 h3 h4 h5 h6 head header hgroup hr html i iframe
		img input ins kbd label legend li link main map mark menu meta meter nav noscript object ol
		optgroup option output p picture pre progress q rp rt ruby s samp script search section
		select slot small source span strong style sub summary sup table tbody td template textarea
		tfoot th thead time title tr track u ul var video wbr`
	voidList = `area base br col embed hr img input link meta source track wbr basefont frame isindex param`
)

func set(list string) map[string]bool {
	m := map[string]bool{}
	for _, n := range strings.Fields(list) {
		m[n] = true
	}
	return m
}

func diff(t *testing.T, what string, got, want map[string]bool) {
	t.Helper()
	var missing, extra []string
	for n := range want {
		if !got[n] {
			missing = append(missing, n)
		}
	}
	for n := range got {
		if !want[n] {
			extra = append(extra, n)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing)+len(extra) > 0 {
		t.Errorf("%s: missing %v, unexpected %v", what, missing, extra)
	}
}

func TestElementLists(t *testing.T) {
	h4, h5, void, obsolete := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, e := range Elements() {
		if e.HTML4 {
			h4[e.Name] = true
		}
		if e.HTML5 {
			h5[e.Name] = true
		}
		if e.Void {
			void[e.Name] = true
		}
		if e.Obsolete {
			obsolete[e.Name] = true
		}
	}
	diff(t, "HTML 4.01", h4, set(html401))
	diff(t, "HTML5", h5, set(living))
	diff(t, "void", void, set(voidList))
	// Obsolete means exactly: in HTML4, not in HTML5.
	removedSet := map[string]bool{}
	for n := range set(html401) {
		if !set(living)[n] {
			removedSet[n] = true
		}
	}
	diff(t, "obsolete", obsolete, removedSet)
}

func TestElementsAreConsistent(t *testing.T) {
	list := Elements()
	if len(list) != len(set(html401+" "+living)) {
		t.Errorf("%d elements", len(list))
	}
	for i, e := range list {
		if i > 0 && list[i-1].Name >= e.Name {
			t.Errorf("not sorted at %s", e.Name)
		}
		if Lookup(e.Name) != e || Lookup(strings.ToUpper(e.Name)) != nil {
			t.Errorf("Lookup(%s)", e.Name)
		}
		if e.Obsolete && e.Hint == "" || !e.Obsolete && e.Hint != "" {
			t.Errorf("%s: hint %q", e.Name, e.Hint)
		}
		for name, a := range e.Attrs {
			if a.Name != name || (a.Kind == Enum) != (len(a.Values) > 0) || a.Obsolete != (a.Hint != "") {
				t.Errorf("%s.%s: %+v", e.Name, name, a)
			}
		}
	}
}

func TestAttributes(t *testing.T) {
	cases := []struct {
		elem, attr string
		kind       Kind
		obsolete   bool
	}{
		{"a", "href", URL, false},
		{"a", "rel", Tokens, false},
		{"a", "name", String, true},
		{"input", "type", Enum, false},
		{"input", "required", Bool, false},
		{"input", "maxlength", UInt, false},
		{"input", "list", IDRef, false},
		{"td", "colspan", UInt, false},
		{"td", "headers", IDRefs, false},
		{"td", "bgcolor", String, true},
		{"meter", "value", Float, false},
		{"label", "for", IDRef, false},
		{"body", "bgcolor", String, true},
		{"body", "onbeforeunload", Event, false},
		{"frameset", "onunload", Event, false},
		{"font", "color", String, false}, // the element is obsolete, not its attributes
		{"div", "id", String, false},
		{"div", "tabindex", Int, false},
		{"div", "hidden", Enum, false},
		{"div", "class", Tokens, false},
		{"center", "onclick", Event, false},
		{"div", "data-user-id", String, false},
		{"div", "aria-label", String, false},
		{"div", "align", Enum, true},
		{"h3", "align", Enum, true},
	}
	for _, c := range cases {
		a := Lookup(c.elem).Attr(c.attr)
		if a == nil || a.Kind != c.kind || a.Obsolete != c.obsolete {
			t.Errorf("%s.%s = %+v, want kind %v obsolete %v", c.elem, c.attr, a, c.kind, c.obsolete)
		}
	}
	for _, c := range [][2]string{
		{"div", "href"}, {"div", "onbeforeunload"}, {"span", "colspan"}, {"div", "data-"},
		{"div", "data-User"}, {"div", "data-xmlish"}, {"div", "aria-"}, {"div", "Class"}, {"img", "onClick"},
	} {
		if a := Lookup(c[0]).Attr(c[1]); a != nil {
			t.Errorf("%s.%s accepted: %+v", c[0], c[1], a)
		}
	}
}

func TestEnumValues(t *testing.T) {
	has := func(elem, attr, v string) bool {
		for _, x := range Lookup(elem).Attr(attr).Values {
			if x == v {
				return true
			}
		}
		return false
	}
	for _, c := range [][3]string{
		{"input", "type", "email"}, {"input", "type", "datetime-local"}, {"button", "type", "reset"},
		{"form", "method", "dialog"}, {"ol", "type", "I"}, {"img", "crossorigin", ""},
		{"img", "loading", "lazy"}, {"div", "dir", "auto"}, {"th", "scope", "colgroup"},
	} {
		if !has(c[0], c[1], c[2]) {
			t.Errorf("%s.%s lacks %q", c[0], c[1], c[2])
		}
	}
	if has("input", "type", "datetime") || has("button", "type", "text") {
		t.Error("unexpected keyword")
	}
}

func TestParseAttrs(t *testing.T) {
	a, err := parseAttrs("href:url; !align(left|right)~use CSS; type(|a|b); x-y", HTML)
	if err != nil {
		t.Fatal(err)
	}
	if a["href"].Kind != URL || !a["align"].Obsolete || a["align"].Hint != "use CSS" || a["align"].Kind != Enum ||
		strings.Join(a["type"].Values, ",") != ",a,b" || a["x-y"].Kind != String {
		t.Errorf("%+v", a)
	}
	for _, bad := range []string{"x:nope", "x(a|b", "!x", "x~hint", "x; x", "X", "x:bool(a)", "x:enum"} {
		if _, err := parseAttrs(bad, HTML); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

// Written independently of svg.go, from the element indexes of SVG 1.1
// (second edition) and SVG 2 (without the HTML elements SVG 2 allows, the
// unknown element, and the draft-only ones), plus feDropShadow.
const (
	svg11 = `a altGlyph altGlyphDef altGlyphItem animate animateColor animateMotion animateTransform
		circle clipPath color-profile cursor defs desc ellipse feBlend feColorMatrix feComponentTransfer
		feComposite feConvolveMatrix feDiffuseLighting feDisplacementMap feDistantLight feFlood feFuncA
		feFuncB feFuncG feFuncR feGaussianBlur feImage feMerge feMergeNode feMorphology feOffset
		fePointLight feSpecularLighting feSpotLight feTile feTurbulence filter font font-face
		font-face-format font-face-name font-face-src font-face-uri foreignObject g glyph glyphRef hkern
		image line linearGradient marker mask metadata missing-glyph mpath path pattern polygon polyline
		radialGradient rect script set stop style svg switch symbol text textPath title tref tspan use
		view vkern`
	svg2 = `a animate animateMotion animateTransform circle clipPath defs desc ellipse feBlend
		feColorMatrix feComponentTransfer feComposite feConvolveMatrix feDiffuseLighting
		feDisplacementMap feDistantLight feDropShadow feFlood feFuncA feFuncB feFuncG feFuncR
		feGaussianBlur feImage feMerge feMergeNode feMorphology feOffset fePointLight
		feSpecularLighting feSpotLight feTile feTurbulence filter foreignObject g image line
		linearGradient marker mask metadata mpath path pattern polygon polyline radialGradient rect
		script set stop style svg switch symbol text textPath title tspan use view`
)

func TestSVGElementLists(t *testing.T) {
	v11, v2, obsolete := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, e := range SVGElements() {
		if e.Namespace != SVG || e.HTML4 || e.HTML5 || e.Void {
			t.Errorf("%s: %+v", e.Name, e)
		}
		if e.SVG11 {
			v11[e.Name] = true
		}
		if e.SVG2 {
			v2[e.Name] = true
		}
		if e.Obsolete {
			obsolete[e.Name] = true
			if e.Hint == "" {
				t.Errorf("%s: no hint", e.Name)
			}
		}
		if LookupSVG(e.Name) != e {
			t.Errorf("LookupSVG(%s)", e.Name)
		}
	}
	diff(t, "SVG 1.1", v11, set(svg11))
	diff(t, "SVG 2", v2, set(svg2))
	removedSet := map[string]bool{}
	for n := range set(svg11) {
		if !set(svg2)[n] {
			removedSet[n] = true
		}
	}
	diff(t, "SVG obsolete", obsolete, removedSet)
}

// HTML and SVG are separate namespaces: names they share are different
// elements, and SVG names are case-sensitive.
func TestNamespaces(t *testing.T) {
	for _, n := range []string{"a", "script", "style", "title"} {
		h, s := Lookup(n), LookupSVG(n)
		if h == nil || s == nil || h == s || h.Namespace != HTML || s.Namespace != SVG {
			t.Errorf("%s: %v %v", n, h, s)
		}
	}
	if Lookup("circle") != nil || LookupSVG("div") != nil || LookupSVG("lineargradient") != nil || LookupSVG("linearGradient") == nil {
		t.Error("namespace leak or case-insensitive lookup")
	}
	if Lookup("a").Attr("viewBox") != nil || Lookup("div").Attr("fill") != nil {
		t.Error("SVG attributes on HTML elements")
	}
}

func TestSVGAttributes(t *testing.T) {
	cases := []struct {
		elem, attr string
		kind       Kind
		obsolete   bool
	}{
		{"circle", "r", Length, false},
		{"circle", "fill", String, false},
		{"circle", "fill-rule", Enum, false},
		{"rect", "rx", Length, false},
		{"svg", "viewBox", String, false},
		{"svg", "version", String, true},
		{"use", "href", URL, false},
		{"use", "xlink:href", URL, true},
		{"g", "xml:space", Enum, true},
		{"g", "class", Tokens, false},
		{"g", "onclick", Event, false},
		{"g", "data-id", String, false},
		{"g", "aria-hidden", String, false},
		{"g", "kerning", String, true},
		{"linearGradient", "gradientUnits", Enum, false},
		{"feTurbulence", "numOctaves", UInt, false},
		{"feComposite", "k1", Float, false},
		{"animate", "fill", Enum, false}, // freeze|remove, not the paint
		{"path", "d", String, false},
		{"text", "lengthAdjust", Enum, false},
		{"font", "horiz-adv-x", String, false}, // the element is obsolete, not its attributes
	}
	for _, c := range cases {
		a := LookupSVG(c.elem).Attr(c.attr)
		if a == nil || a.Kind != c.kind || a.Obsolete != c.obsolete {
			t.Errorf("%s.%s = %+v, want kind %v obsolete %v", c.elem, c.attr, a, c.kind, c.obsolete)
		}
	}
	if a := LookupSVG("animate").Attr("fill"); strings.Join(a.Values, "|") != "freeze|remove" {
		t.Errorf("animate fill = %v", a.Values)
	}
	if a := LookupSVG("path").Attr("stroke-linecap"); a == nil || !strings.Contains(strings.Join(a.Values, "|"), "round|square|inherit") {
		t.Errorf("stroke-linecap = %+v", a)
	}
	for _, c := range [][2]string{{"circle", "href"}, {"circle", "viewbox"}, {"g", "width"}, {"rect", "colspan"}, {"svg", "xlink:foo"}} {
		if a := LookupSVG(c[0]).Attr(c[1]); a != nil {
			t.Errorf("%s.%s accepted: %+v", c[0], c[1], a)
		}
	}
}

func TestParseSVGAttrs(t *testing.T) {
	a, err := parseAttrs("viewBox; xlink:href:url; !xml:space(default|preserve)~use CSS", SVG)
	if err != nil {
		t.Fatal(err)
	}
	if a["viewBox"] == nil || a["xlink:href"].Kind != URL || a["xml:space"].Kind != Enum || !a["xml:space"].Obsolete {
		t.Errorf("%+v", a)
	}
	for _, bad := range []string{"foo:href", "xlink:", "viewBox; viewBox"} {
		if _, err := parseAttrs(bad, SVG); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if _, err := parseAttrs("viewBox", HTML); err == nil {
		t.Error("camelCase accepted in HTML")
	}
}
