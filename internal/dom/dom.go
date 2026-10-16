// Package dom is RuneScript's internal list of DOM elements and their
// attributes (OQ-JS-09). Templates will be checked against it: an element
// or attribute that is not in the list is a compile-time error, an
// attribute value must have the attribute's type, and elements and
// attributes that HTML5 made obsolete are accepted with a warning.
//
// The list covers HTML 4.01 (strict, transitional, and frameset), the HTML
// living standard ("HTML5"), SVG 1.1, and SVG 2, in two namespaces: HTML
// (elements.go) and SVG (svg.go). It is transcribed by hand from the
// specifications and is not exhaustive; the tests cross-check it against
// independently written element lists.
package dom

import (
	"fmt"
	"sort"
	"strings"
)

// Kind is the type of an attribute's value.
type Kind int

const (
	String Kind = iota // any text
	URL                // a URL
	Bool               // a boolean attribute: present or absent
	Int                // an integer, e.g. tabindex
	UInt               // a non-negative integer, e.g. colspan
	Float              // a number, e.g. meter's value
	Enum               // one of Attr.Values
	Event              // an event handler
	IDRef              // the id of another element, e.g. label's for
	IDRefs             // a space-separated list of ids, e.g. headers
	Tokens             // a space-separated list of words, e.g. class, rel
	Length             // a number, or a length or percentage, e.g. 10, 10px, 50%
)

var kindNames = map[string]Kind{
	"string": String, "url": URL, "bool": Bool, "int": Int, "uint": UInt, "float": Float,
	"enum": Enum, "event": Event, "idref": IDRef, "idrefs": IDRefs, "tokens": Tokens, "length": Length,
}

func (k Kind) String() string {
	for name, kk := range kindNames {
		if kk == k {
			return name
		}
	}
	return fmt.Sprintf("Kind(%d)", int(k))
}

// Attr is an attribute of an element, or a global attribute.
type Attr struct {
	Name   string
	Kind   Kind
	Values []string // the keywords of an Enum; "" means the empty value
	// Obsolete attributes are accepted with a warning; Hint says what to
	// use instead.
	Obsolete bool
	Hint     string
}

// Namespace says whether an element is an HTML or an SVG element. A few
// names exist in both (a, script, style, title) and differ.
type Namespace int

const (
	HTML Namespace = iota
	SVG
)

func (n Namespace) String() string {
	if n == SVG {
		return "SVG"
	}
	return "HTML"
}

// Element is an element of HTML4 and/or HTML5, or of SVG 1.1 and/or SVG 2.
type Element struct {
	Name      string
	Namespace Namespace
	HTML4     bool // in HTML 4.01
	HTML5     bool // in the HTML living standard
	SVG11     bool // in SVG 1.1 (second edition)
	SVG2      bool // in SVG 2
	// Void elements have no content and no end tag, e.g. br and img.
	Void bool
	// Obsolete elements are HTML4 elements that HTML5 removed, or SVG 1.1
	// elements that SVG 2 removed. They are accepted with a warning; Hint
	// says what to use instead.
	Obsolete bool
	Hint     string
	// Attrs are the element's own attributes. Global attributes and
	// event handlers are not repeated here; see Attr.
	Attrs map[string]*Attr
	// Events are event handler attributes only this element has, e.g.
	// body's onbeforeunload.
	Events map[string]*Attr
}

// Lookup returns the HTML element called name, or nil. HTML names are
// lowercase: RuneScript templates do not accept other spellings.
func Lookup(name string) *Element { return elements[name] }

// LookupSVG returns the SVG element called name, or nil. SVG names are
// case-sensitive, e.g. linearGradient.
func LookupSVG(name string) *Element { return svgElements[name] }

// Elements returns every HTML element, sorted by name.
func Elements() []*Element { return sorted(elements) }

// SVGElements returns every SVG element, sorted by name.
func SVGElements() []*Element { return sorted(svgElements) }

func sorted(m map[string]*Element) []*Element {
	list := make([]*Element, 0, len(m))
	for _, e := range m {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

// Attr returns the attribute name of e: one of e's own attributes, a
// global attribute (for SVG, also a presentation attribute), an event
// handler, a data-* attribute (String), or an aria-* attribute (String).
// It returns nil for anything else.
func (e *Element) Attr(name string) *Attr {
	if a := e.Attrs[name]; a != nil {
		return a
	}
	shared := globals
	if e.Namespace == SVG {
		shared = svgGlobals
		if a := presentation[name]; a != nil {
			return a
		}
	}
	if a := shared[name]; a != nil {
		return a
	}
	if a := events[name]; a != nil {
		return a
	}
	if a := e.Events[name]; a != nil {
		return a
	}
	switch {
	case strings.HasPrefix(name, "data-") && validDataName(name[len("data-"):]):
		return &Attr{Name: name, Kind: String}
	case strings.HasPrefix(name, "aria-") && validDataName(name[len("aria-"):]):
		return &Attr{Name: name, Kind: String}
	}
	return nil
}

// validDataName reports whether s may follow "data-" or "aria-": at least
// one character, lowercase letters, digits, '-', '_', and '.', and not
// starting with "xml" (HTML reserves those names).
func validDataName(s string) bool {
	if s == "" || strings.HasPrefix(s, "xml") {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}

// IsGlobal reports whether name is a global attribute or a global event
// handler of HTML elements.
func IsGlobal(name string) bool { return globals[name] != nil || events[name] != nil }

// ---- the attribute notation used in elements.go ----
//
// An attribute list is written as entries separated by ';'. Each entry is
//
//	[!]name[:kind][(value|value|...)][~hint]
//
// SVG names may be camelCase (viewBox) and may start with "xlink:" or
// "xml:" (xlink:href).
// The kind defaults to string; an enum lists its keywords in parentheses,
// where an empty keyword is the empty value. '!' marks an obsolete
// attribute and '~' starts the hint shown for it.

// parseAttrs parses an attribute list of namespace ns. An error is a
// mistake in elements.go or svg.go, not a user error.
func parseAttrs(spec string, ns Namespace) (map[string]*Attr, error) {
	attrs := map[string]*Attr{}
	for _, entry := range strings.Split(spec, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		a := &Attr{}
		if strings.HasPrefix(entry, "!") {
			a.Obsolete = true
			entry = entry[1:]
		}
		if i := strings.Index(entry, "~"); i >= 0 {
			a.Hint = strings.TrimSpace(entry[i+1:])
			entry = strings.TrimSpace(entry[:i])
		}
		if i := strings.Index(entry, "("); i >= 0 {
			if !strings.HasSuffix(entry, ")") {
				return nil, fmt.Errorf("%q: unclosed keyword list", entry)
			}
			a.Values = strings.Split(entry[i+1:len(entry)-1], "|")
			entry = entry[:i]
		}
		prefix := ""
		for _, p := range []string{"xlink:", "xml:"} {
			if ns == SVG && strings.HasPrefix(entry, p) {
				prefix, entry = p, entry[len(p):]
			}
		}
		name, kind, hasKind := strings.Cut(entry, ":")
		a.Name = prefix + name
		switch {
		case !hasKind && a.Values != nil:
			a.Kind = Enum
		case !hasKind:
			a.Kind = String
		default:
			k, ok := kindNames[kind]
			if !ok {
				return nil, fmt.Errorf("%q: unknown kind %q", entry, kind)
			}
			a.Kind = k
		}
		switch {
		case !validAttrName(strings.TrimPrefix(a.Name, prefix), ns):
			return nil, fmt.Errorf("%q: bad attribute name", a.Name)
		case (a.Kind == Enum) != (a.Values != nil):
			return nil, fmt.Errorf("%q: keyword lists go with the enum kind only", a.Name)
		case a.Obsolete != (a.Hint != ""):
			return nil, fmt.Errorf("%q: an obsolete attribute needs a hint, and only it", a.Name)
		case attrs[a.Name] != nil:
			return nil, fmt.Errorf("%q: listed twice", a.Name)
		}
		attrs[a.Name] = a
	}
	return attrs, nil
}

// validAttrName reports whether s is an attribute name: lowercase letters,
// digits, and '-'; SVG names may also have uppercase letters.
func validAttrName(s string, ns Namespace) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || ns == SVG && c >= 'A' && c <= 'Z') {
			return false
		}
	}
	return true
}

func mustAttrs(spec string) map[string]*Attr { return mustAttrsNS(spec, HTML) }

func mustAttrsNS(spec string, ns Namespace) map[string]*Attr {
	attrs, err := parseAttrs(spec, ns)
	if err != nil {
		panic("dom: " + err.Error())
	}
	return attrs
}

// eventAttrs makes Event attributes from space-separated names.
func eventAttrs(names string) map[string]*Attr {
	attrs := map[string]*Attr{}
	for _, n := range strings.Fields(names) {
		attrs[n] = &Attr{Name: n, Kind: Event}
	}
	return attrs
}
