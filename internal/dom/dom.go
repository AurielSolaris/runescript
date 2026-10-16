// Package dom is RuneScript's internal list of DOM elements and their
// attributes (OQ-JS-09). Templates will be checked against it: an element
// or attribute that is not in the list is a compile-time error, an
// attribute value must have the attribute's type, and elements and
// attributes that HTML5 made obsolete are accepted with a warning.
//
// The list covers HTML 4.01 (strict, transitional, and frameset) and the
// HTML living standard ("HTML5"). SVG is planned. It is transcribed by hand
// from the specifications (elements.go) and is not exhaustive; the tests
// cross-check it against independently written element lists.
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
)

var kindNames = map[string]Kind{
	"string": String, "url": URL, "bool": Bool, "int": Int, "uint": UInt, "float": Float,
	"enum": Enum, "event": Event, "idref": IDRef, "idrefs": IDRefs, "tokens": Tokens,
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

// Element is an element of HTML4, HTML5, or both.
type Element struct {
	Name  string
	HTML4 bool // in HTML 4.01
	HTML5 bool // in the HTML living standard
	// Void elements have no content and no end tag, e.g. br and img.
	Void bool
	// Obsolete elements are HTML4 elements that HTML5 removed. They are
	// accepted with a warning; Hint says what to use instead.
	Obsolete bool
	Hint     string
	// Attrs are the element's own attributes. Global attributes and
	// event handlers are not repeated here; see Attr.
	Attrs map[string]*Attr
	// Events are event handler attributes only this element has, e.g.
	// body's onbeforeunload.
	Events map[string]*Attr
}

// Lookup returns the element called name, or nil. Names are lowercase:
// RuneScript templates do not accept other spellings.
func Lookup(name string) *Element { return elements[name] }

// Elements returns every element, sorted by name.
func Elements() []*Element {
	list := make([]*Element, 0, len(elements))
	for _, e := range elements {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

// Attr returns the attribute name of e: one of e's own attributes, a
// global attribute, an event handler, a data-* attribute (String), or an
// aria-* attribute (String). It returns nil for anything else.
func (e *Element) Attr(name string) *Attr {
	if a := e.Attrs[name]; a != nil {
		return a
	}
	if a := globals[name]; a != nil {
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
// handler.
func IsGlobal(name string) bool { return globals[name] != nil || events[name] != nil }

// ---- the attribute notation used in elements.go ----
//
// An attribute list is written as entries separated by ';'. Each entry is
//
//	[!]name[:kind][(value|value|...)][~hint]
//
// The kind defaults to string; an enum lists its keywords in parentheses,
// where an empty keyword is the empty value. '!' marks an obsolete
// attribute and '~' starts the hint shown for it.

// parseAttrs parses an attribute list. An error is a mistake in
// elements.go, not a user error.
func parseAttrs(spec string) (map[string]*Attr, error) {
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
		name, kind, hasKind := strings.Cut(entry, ":")
		a.Name = name
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
		case !validAttrName(a.Name):
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

func validAttrName(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}

func mustAttrs(spec string) map[string]*Attr {
	attrs, err := parseAttrs(spec)
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
