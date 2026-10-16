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
	a, err := parseAttrs("href:url; !align(left|right)~use CSS; type(|a|b); x-y")
	if err != nil {
		t.Fatal(err)
	}
	if a["href"].Kind != URL || !a["align"].Obsolete || a["align"].Hint != "use CSS" || a["align"].Kind != Enum ||
		strings.Join(a["type"].Values, ",") != ",a,b" || a["x-y"].Kind != String {
		t.Errorf("%+v", a)
	}
	for _, bad := range []string{"x:nope", "x(a|b", "!x", "x~hint", "x; x", "X", "x:bool(a)", "x:enum"} {
		if _, err := parseAttrs(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}
