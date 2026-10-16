package parser

import (
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/dom"
	"github.com/AurielSolaris/runescript/internal/lexer"
)

// templateStmt parses `template { markup }` (OQ-DOM-01). The markup has
// exactly one root element. Element and attribute names are checked by
// the checker; the parser checks the structure: matching tags, void
// elements (HTML only: inside <svg> every element may be self-closed),
// and complete $if ... $endif blocks.
func (p *parser) templateStmt() *ast.TemplateStmt {
	start := p.next() // template
	p.expect(lexer.LBrace, "'{' after template")
	nodes := p.markupNodes(false)
	end := p.tok()
	switch end.Kind {
	case lexer.MTagClose:
		p.fail(diagnostics.ParseUnexpected, end.Span, "</%s> closes no open element", end.Text)
	case lexer.MDirective:
		p.fail(diagnostics.ParseUnexpected, end.Span, "%s without $if", end.Text)
	}
	p.expect(lexer.RBrace, "'}' to close the template")
	var root *ast.MElement
	for _, n := range nodes {
		if t, ok := n.(*ast.MText); ok && strings.TrimSpace(t.Text) == "" {
			continue // white space around the root element
		}
		e, ok := n.(*ast.MElement)
		if !ok || root != nil {
			p.failHelp(diagnostics.ParseUnexpected, n.NodeSpan(),
				"wrap the content in one element, e.g. <div> ... </div>",
				"a template has exactly one root element")
		}
		root = e
	}
	if root == nil {
		p.failHelp(diagnostics.ParseUnexpected, start.Span, "e.g. template { <p>Hello</p> }", "the template is empty: it needs one root element")
	}
	return &ast.TemplateStmt{Root: root, Span: span(start.Span, end.Span)}
}

// markupNodes parses content up to a closing tag, a directive that
// continues or ends an $if, or the end of the template; it stops before it.
func (p *parser) markupNodes(svg bool) []ast.Markup {
	var nodes []ast.Markup
	for {
		t := p.tok()
		switch t.Kind {
		case lexer.MText:
			p.next()
			if s := normText(t.Text); s != "" {
				nodes = append(nodes, &ast.MText{Text: s, Span: t.Span})
			}
		case lexer.MTagOpen:
			nodes = append(nodes, p.element(svg))
		case lexer.MInterpOpen:
			p.next()
			x := p.expr()
			end := p.expect(lexer.MInterpClose, "'}}' to close the interpolation")
			nodes = append(nodes, &ast.MInterp{X: x, Span: span(t.Span, end.Span)})
		case lexer.MDirective:
			if t.Text != "$if" {
				return nodes
			}
			nodes = append(nodes, p.markupIf(svg))
		default:
			return nodes
		}
	}
}

// element parses `<name attrs>children</name>` or `<name attrs/>`.
func (p *parser) element(svg bool) *ast.MElement {
	open := p.next()
	e := &ast.MElement{Name: open.Text, NameSpan: open.Span}
	e.NameSpan.Start.Offset++ // after '<'
	e.NameSpan.Start.Col++
	inner := svg || e.Name == "svg"
	if svg && e.Name == "foreignObject" {
		inner = false
	}
	for p.is(lexer.MAttr, "") {
		e.Attrs = append(e.Attrs, p.attribute())
	}
	switch t := p.tok(); t.Kind {
	case lexer.MSelfClose:
		p.next()
		e.Span = span(open.Span, t.Span)
		return e
	case lexer.MTagEnd:
		p.next()
		if !svg {
			if d := dom.Lookup(e.Name); d != nil && d.Void {
				e.Span = span(open.Span, t.Span)
				return e
			}
		}
	default:
		p.unexpected("'>' or '/>' to end the tag")
	}
	e.Children = p.markupNodes(inner)
	end := p.tok()
	switch {
	case end.Kind == lexer.MTagClose && end.Text == e.Name:
		p.next()
		e.Span = span(open.Span, end.Span)
		return e
	case end.Kind == lexer.MTagClose:
		help := ""
		if d := dom.Lookup(end.Text); d != nil && d.Void && !svg {
			help = "<" + end.Text + "> is a void element: it has no closing tag"
		}
		p.failHelp(diagnostics.ParseUnexpected, end.Span, help, "</%s> does not close <%s>, opened at %s", end.Text, e.Name, open.Span.Start)
	case end.Kind == lexer.MDirective:
		p.fail(diagnostics.ParseUnexpected, end.Span, "%s inside <%s>, which its $if does not contain: close <%s> first", end.Text, e.Name, e.Name)
	default:
		p.failHelp(diagnostics.ParseUnexpected, open.Span, "close it with </"+e.Name+">, or write <"+e.Name+" />",
			"<%s> is not closed", e.Name)
	}
	return nil
}

func (p *parser) attribute() *ast.MAttr {
	t := p.next()
	a := &ast.MAttr{Name: t.Text, NameSpan: t.Span, Span: t.Span}
	if !p.is(lexer.Assign, "") {
		return a
	}
	p.next()
	switch v := p.tok(); v.Kind {
	case lexer.MAttrValue:
		p.next()
		s := v.Text
		a.Value, a.ValueSpan = &s, v.Span
		a.Span = span(t.Span, v.Span)
	case lexer.MInterpOpen:
		p.next()
		a.Expr = p.expr()
		end := p.expect(lexer.MInterpClose, "'}}' to close the interpolation")
		a.ValueSpan = span(v.Span, end.Span)
		a.Span = span(t.Span, end.Span)
	default:
		p.unexpected("an attribute value")
	}
	return a
}

// markupIf parses `$if (c) … [$elseif (c) …]… [$else …] $endif`.
func (p *parser) markupIf(svg bool) *ast.MIf {
	start := p.next()
	n := &ast.MIf{}
	cond := func() ast.Expr {
		p.expect(lexer.LParen, "'(' after "+p.toks[p.pos-1].Text)
		x := p.expr()
		p.expect(lexer.RParen, "')' to close the condition")
		return x
	}
	n.Conds = append(n.Conds, cond())
	n.Branches = append(n.Branches, p.markupNodes(svg))
	for {
		t := p.tok()
		if t.Kind != lexer.MDirective {
			p.failHelp(diagnostics.ParseUnexpected, start.Span, "end it with $endif", "$if has no $endif")
		}
		p.next()
		switch t.Text {
		case "$elseif":
			if n.HasElse {
				p.fail(diagnostics.ParseUnexpected, t.Span, "$elseif after $else")
			}
			n.Conds = append(n.Conds, cond())
			n.Branches = append(n.Branches, p.markupNodes(svg))
		case "$else":
			if n.HasElse {
				p.fail(diagnostics.ParseUnexpected, t.Span, "a second $else")
			}
			n.HasElse = true
			n.Else = p.markupNodes(svg)
		case "$endif":
			n.Span = span(start.Span, t.Span)
			return n
		case "$if":
			p.fail(diagnostics.ParseUnexpected, t.Span, "unexpected $if") // unreachable: markupNodes parses it
		}
	}
}

// normText applies the white-space rule: white space that contains a line
// break is removed at the start and end of the text and collapsed to one
// space inside it. Other white space is kept.
func normText(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if !isSpace(s[i]) {
			b.WriteByte(s[i])
			i++
			continue
		}
		j := i
		for j < len(s) && isSpace(s[j]) {
			j++
		}
		run := s[i:j]
		switch {
		case !strings.ContainsAny(run, "\n\r"):
			b.WriteString(run)
		case i > 0 && j < len(s):
			b.WriteByte(' ')
		}
		i = j
	}
	return b.String()
}

func isSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }
