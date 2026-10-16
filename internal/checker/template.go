package checker

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/dom"
	"github.com/AurielSolaris/runescript/internal/types"
)

// Templates (OQ-DOM-01): a component is a function returning Element whose
// last statement is `template { … }`. Every element and attribute is
// checked against internal/dom: unknown names are errors, obsolete ones
// warn, static attribute values must fit the attribute, and {{ }} values
// must have a type the attribute (or the text) can take.

// templateFor returns the template statement that may end a function
// body: its last statement, if that is one.
func templateFor(body []ast.Stmt) *ast.TemplateStmt {
	if len(body) == 0 {
		return nil
	}
	t, _ := body[len(body)-1].(*ast.TemplateStmt)
	return t
}

func (c *checker) templateStmt(s *ast.TemplateStmt) {
	switch {
	case !c.dom:
		c.errorf(diagnostics.TmplNeedsDOM, s.Span, "templates need rune:dom").
			WithHelp(`add use "rune:dom"; after the package declaration`)
		return
	case s != c.lastTemplate:
		c.errorf(diagnostics.TmplPlacement, s.Span, "a template must be the last statement of a function returning Element").
			WithHelp("a component is `public Element Name() { … template { … } }`")
	case c.result == nil: // a lambda whose result type is inferred
		c.result = types.ElementT
	case c.result.Kind != types.Element && c.result.Kind != types.Invalid:
		c.errorf(diagnostics.TmplPlacement, s.Span, "%s returns '%s', but a template makes it return 'Element'", c.fnName, c.result).
			WithHelp("declare the function's result type as Element")
	}
	c.domUsed[c.file] = true
	c.markupElement(s.Root, false)
}

// markupNodes checks content in the HTML (svg false) or SVG namespace.
func (c *checker) markupNodes(nodes []ast.Markup, svg bool) {
	for _, n := range nodes {
		switch n := n.(type) {
		case *ast.MElement:
			c.markupElement(n, svg)
		case *ast.MInterp:
			c.textValue(n.X)
		case *ast.MIf:
			for i, cond := range n.Conds {
				c.cond(cond)
				c.markupNodes(n.Branches[i], svg)
			}
			c.markupNodes(n.Else, svg)
		}
	}
}

func (c *checker) markupElement(e *ast.MElement, svg bool) {
	var el *dom.Element
	if svg || e.Name == "svg" {
		el = dom.LookupSVG(e.Name)
		svg = true
	} else {
		el = dom.Lookup(e.Name)
	}
	if el == nil {
		c.unknownElement(e, svg)
	} else {
		c.info.TemplateElems[e] = el
		if el.Obsolete {
			c.warnf(diagnostics.WarnObsolete, e.NameSpan, "<%s> is obsolete in %s", e.Name, current(el)).
				WithHelp(el.Hint)
		}
	}
	seen := map[string]*ast.MAttr{}
	for _, a := range e.Attrs {
		if prev := seen[a.Name]; prev != nil {
			c.errorf(diagnostics.TmplDuplicateAttr, a.NameSpan, "attribute '%s' is already set at %s", a.Name, prev.NameSpan.Start)
		}
		seen[a.Name] = a
		if el != nil {
			c.markupAttr(el, a)
		} else if a.Expr != nil {
			c.expr(a.Expr)
		}
	}
	inner := svg && !(e.Name == "foreignObject")
	c.markupNodes(e.Children, inner)
}

func current(el *dom.Element) string {
	if el.Namespace == dom.SVG {
		return "SVG 2"
	}
	return "HTML5"
}

func (c *checker) unknownElement(e *ast.MElement, svg bool) {
	d := c.errorf(diagnostics.TmplUnknownElement, e.NameSpan, "unknown element <%s>", e.Name)
	lower := strings.ToLower(e.Name)
	switch {
	case svg && dom.Lookup(e.Name) != nil:
		d.WithHelp(fmt.Sprintf("<%s> is an HTML element; inside <svg>, put HTML in a <foreignObject>", e.Name))
	case !svg && dom.LookupSVG(e.Name) != nil:
		d.WithHelp(fmt.Sprintf("<%s> is an SVG element: put it inside <svg>", e.Name))
	case !svg && dom.Lookup(lower) != nil:
		d.WithHelp(fmt.Sprintf("HTML element names are lowercase: <%s>", lower))
	case svg:
		for _, s := range dom.SVGElements() {
			if strings.EqualFold(s.Name, e.Name) {
				d.WithHelp(fmt.Sprintf("SVG element names are case-sensitive: <%s>", s.Name))
			}
		}
	case strings.Contains(e.Name, "-"):
		d.WithHelp("custom elements are not supported yet")
	}
}

func (c *checker) markupAttr(el *dom.Element, a *ast.MAttr) {
	at := el.Attr(a.Name)
	if at == nil {
		d := c.errorf(diagnostics.TmplUnknownAttr, a.NameSpan, "<%s> has no attribute '%s'", el.Name, a.Name)
		if el.Namespace == dom.HTML && el.Attr(strings.ToLower(a.Name)) != nil {
			d.WithHelp(fmt.Sprintf("HTML attribute names are lowercase: %s", strings.ToLower(a.Name)))
		} else if strings.HasPrefix(a.Name, "data-") || strings.HasPrefix(a.Name, "aria-") {
			d.WithHelp("after data- and aria-, use lowercase letters, digits, '-', '_', and '.'")
		}
		if a.Expr != nil {
			c.expr(a.Expr)
		}
		return
	}
	c.info.TemplateAttrs[a] = at
	if at.Obsolete {
		c.warnf(diagnostics.WarnObsolete, a.NameSpan, "attribute '%s' of <%s> is obsolete in %s", a.Name, el.Name, current(el)).
			WithHelp(at.Hint)
	}
	if a.Expr != nil {
		c.attrValue(at, a)
		return
	}
	c.staticAttr(at, a)
}

// attrValue checks name={{ expr }}.
func (c *checker) attrValue(at *dom.Attr, a *ast.MAttr) {
	switch at.Kind {
	case dom.Event:
		c.assign(types.NewFunc(nil, types.VoidT), a.Expr, "pass")
	case dom.Bool:
		c.assign(types.BoolT, a.Expr, "pass")
	case dom.Int:
		c.assign(types.Int, a.Expr, "pass")
	case dom.UInt:
		c.assign(types.Uint, a.Expr, "pass")
	case dom.Float, dom.Length:
		x := c.expr(a.Expr)
		switch {
		case isUntyped(x.typ):
			c.convertNum(a.Expr, x, types.Float64)
		case x.typ.Kind == types.Float, x.typ.Kind == types.Invalid:
		case at.Kind == dom.Length && (types.IsInteger(x.typ) || x.typ.Kind == types.String):
		default:
			want := "a float"
			if at.Kind == dom.Length {
				want = "a number or a string such as \"10px\""
			}
			c.errorf(diagnostics.TmplBadInterp, a.Expr.NodeSpan(), "'%s' takes %s, found '%s'", a.Name, want, x.typ)
		}
	default:
		x := c.expr(a.Expr)
		if x.typ.Kind != types.String && x.typ.Kind != types.Invalid {
			d := c.errorf(diagnostics.TmplBadInterp, a.Expr.NodeSpan(), "'%s' takes a string, found '%s'", a.Name, x.typ)
			if types.IsNumeric(x.typ) || x.typ.Kind == types.Bool {
				d.WithHelp("convert it explicitly, e.g. {{ (string) n }}")
			} else if x.typ.Kind == types.Signal {
				d.WithHelp("read the rune: {{ name() }}")
			}
		}
	}
}

var (
	intRE    = regexp.MustCompile(`^-?[0-9]+$`)
	uintRE   = regexp.MustCompile(`^[0-9]+$`)
	lengthRE = regexp.MustCompile(`^[-+]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][-+]?[0-9]+)?(%|[a-zA-Z]+)?$`)
)

// staticAttr checks name and name="value".
func (c *checker) staticAttr(at *dom.Attr, a *ast.MAttr) {
	bad := func(format string, args ...any) *diagnostics.Diagnostic {
		span := a.ValueSpan
		if a.Value == nil {
			span = a.NameSpan
		}
		return c.errorf(diagnostics.TmplBadValue, span, format, args...)
	}
	if at.Kind == dom.Event {
		bad("'%s' takes a RuneScript function, not text", a.Name).
			WithHelp(fmt.Sprintf("write %s={{ handler }} or %s={{ () => count.set(count() + 1) }}", a.Name, a.Name))
		return
	}
	if a.Value == nil {
		if at.Kind != dom.Bool && !(at.Kind == dom.Enum && contains(at.Values, "")) {
			bad("attribute '%s' needs a value", a.Name).WithHelp(fmt.Sprintf(`write %s="…" or %s={{ expression }}`, a.Name, a.Name))
		}
		return
	}
	v := *a.Value
	switch at.Kind {
	case dom.Bool:
		if v != "" && v != a.Name {
			bad("'%s' is a boolean attribute: it is on when present", a.Name).
				WithHelp(fmt.Sprintf("write %s alone, or %s={{ condition }}", a.Name, a.Name))
		}
	case dom.Enum:
		if !contains(at.Values, v) {
			bad("%q is not a value of '%s'", v, a.Name).WithHelp("its values are " + quoteAll(at.Values))
		}
	case dom.Int:
		if !intRE.MatchString(v) {
			bad("'%s' takes an integer, found %q", a.Name, v)
		}
	case dom.UInt:
		if !uintRE.MatchString(v) {
			bad("'%s' takes a non-negative integer, found %q", a.Name, v)
		}
	case dom.Float:
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			bad("'%s' takes a number, found %q", a.Name, v)
		}
	case dom.Length:
		if !lengthRE.MatchString(strings.TrimSpace(v)) {
			bad("'%s' takes a length such as 10, 10px, or 50%%, found %q", a.Name, v)
		}
	case dom.IDRef:
		if v == "" || strings.ContainsAny(v, " \t\n") {
			bad("'%s' takes one element id, found %q", a.Name, v)
		}
	case dom.URL:
		if unsafeURL(v) {
			bad("refusing the URL %q in '%s': it would run script", v, a.Name).
				WithHelp("handle the event in RuneScript instead, e.g. onclick={{ handler }}")
		}
	}
}

var unsafeURLRE = regexp.MustCompile(`(?i)^[\x00-\x20]*(javascript|vbscript|data):`)

// unsafeURL reports whether a URL would run script when followed. The
// JavaScript runtime applies the same rule to {{ }} values.
func unsafeURL(v string) bool {
	return unsafeURLRE.MatchString(strings.NewReplacer("\t", "", "\n", "", "\r", "").Replace(v))
}

// textValue checks {{ expr }} in content: text, or a child element.
func (c *checker) textValue(e ast.Expr) {
	x := c.defaulted(e)
	switch x.typ.Kind {
	case types.String, types.Signed, types.Unsigned, types.Float, types.Bool, types.Element, types.Invalid:
	case types.Signal:
		c.errorf(diagnostics.TmplBadInterp, e.NodeSpan(), "a rune is not text: read it").
			WithHelp("write {{ name() }} to show its value; the text then follows it")
	case types.Void:
		c.errorf(diagnostics.TmplBadInterp, e.NodeSpan(), "expression has type 'void' and cannot be shown")
	default:
		c.errorf(diagnostics.TmplBadInterp, e.NodeSpan(), "values of type '%s' cannot be shown", x.typ).
			WithHelp("{{ }} shows strings, numbers, bools, and elements (child components)")
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func quoteAll(list []string) string {
	q := make([]string, len(list))
	for i, s := range list {
		q[i] = strconv.Quote(s)
	}
	return strings.Join(q, ", ")
}
