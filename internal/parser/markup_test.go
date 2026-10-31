package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

func component(markup string) string {
	return "package p;\nuse \"rune:dom\";\npublic Element App() {\n    template {\n" + markup + "\n    }\n}\n"
}

// tree renders markup as a compact s-expression.
func tree(n ast.Markup) string {
	switch n := n.(type) {
	case *ast.MElement:
		var parts []string
		for _, a := range n.Attrs {
			switch {
			case a.Value != nil:
				parts = append(parts, fmt.Sprintf("%s=%q", a.Name, *a.Value))
			case a.Expr != nil:
				parts = append(parts, a.Name+"={{}}")
			default:
				parts = append(parts, a.Name)
			}
		}
		for _, c := range n.Children {
			parts = append(parts, tree(c))
		}
		return "(" + strings.Join(append([]string{n.Name}, parts...), " ") + ")"
	case *ast.MText:
		return fmt.Sprintf("%q", n.Text)
	case *ast.MInterp:
		return "{{}}"
	case *ast.MIf:
		s := "(if"
		for _, b := range n.Branches {
			s += " ["
			for _, c := range b {
				s += tree(c)
			}
			s += "]"
		}
		if n.HasElse {
			s += " else["
			for _, c := range n.Else {
				s += tree(c)
			}
			s += "]"
		}
		return s + ")"
	case *ast.MForEach:
		s := "(foreach " + n.Type.String() + " " + n.Name.Name + " ["
		for _, c := range n.Body {
			s += tree(c)
		}
		return s + "])"
	}
	return "?"
}

func parseTemplate(t *testing.T, markup string) *ast.TemplateStmt {
	t.Helper()
	f := mustParse(t, component(markup))
	return f.Funcs[0].Body.Stmts[0].(*ast.TemplateStmt)
}

func TestParseTemplate(t *testing.T) {
	cases := map[string]string{
		`<p>Hello</p>`: `(p "Hello")`,
		`<div class="card" hidden id={{ x }}>
            <h1>Hi {{ name() }}!</h1>
            <br>
            <img src="a.png" alt="">
        </div>`: `(div class="card" hidden id={{}} (h1 "Hi " {{}} "!") (br) (img src="a.png" alt=""))`,
		`<ul>
            <li>one</li>  <li>two</li>
        </ul>`: `(ul (li "one") "  " (li "two"))`,
		`<p>a
            b</p>`: `(p "a b")`,
		`<svg viewBox="0 0 10 10"><circle r="4"/><g><path d="M0 0"></path></g><foreignObject><br></foreignObject></svg>`: `(svg viewBox="0 0 10 10" (circle r="4") (g (path d="M0 0")) (foreignObject (br)))`,
		`<div>$if (a) <b>A</b> $elseif (b) B $else C $endif</div>`:                                                       `(div (if [" "(b "A")" "] [" B "] else[" C "]))`,
		`<div>$if (a) x $endif</div>`: `(div (if [" x "]))`,
		`<div/>`:                      `(div)`,
		`<ul>$foreach (string s in names)<li>{{ s }}</li>$endforeach</ul>`:                      `(ul (foreach string s [(li {{}})]))`,
		`<p>$foreach (Array<int> r in g)$foreach (int x in r){{ x }}$endforeach$endforeach</p>`: `(p (foreach Array<int> r [(foreach int x [{{}}])]))`,
		`<p>$foreach (int x in a)$if (x > 0)+$endif$endforeach</p>`:                             `(p (foreach int x [(if ["+"])]))`,
	}
	for markup, want := range cases {
		if got := tree(parseTemplate(t, markup).Root); got != want {
			t.Errorf("%s\n got %s\nwant %s", markup, got, want)
		}
	}
}

func TestParseTemplateErrors(t *testing.T) {
	cases := map[string]string{
		`<div><p></div>`:       "</div> does not close <p>",
		`<div>`:                "<div> is not closed",
		`<p>a</p><p>b</p>`:     "exactly one root element",
		`text`:                 "exactly one root element",
		``:                     "the template is empty",
		`<div>$if (a) x</div>`: "$endif",
		`<div>$if (a) x $else y $else z $endif</div>`:         "a second $else",
		`<div>$else</div>`:                                    "$else inside <div>",
		`<p><br></br></p>`:                                    "void element",
		`<p class=></p>`:                                      "",
		`<ul>$foreach (string s in names)<li>x</li></ul>`:     "$endforeach",
		`<ul>$foreach (string s in names) x $endif</ul>`:      "$endif inside $foreach",
		`<ul>$if (a) x $endforeach</ul>`:                      "$endforeach inside an $if",
		`<ul>$endforeach</ul>`:                                "$endforeach inside <ul>, which its $foreach",
		`<ul>$foreach (s in names) x $endforeach</ul>`:        "before 'in'",
		`<ul>$foreach (string s of names) x $endforeach</ul>`: "expected 'in'",
		`<ul><li>$foreach (int x in a)</li>$endforeach</ul>`:  "$foreach has no $endforeach",
	}
	for markup, want := range cases {
		_, diags := Parse([]byte(component(markup)))
		if len(diags) == 0 || (diags[0].Code != diagnostics.ParseUnexpected && diags[0].Code != diagnostics.LexUnexpectedChar) || !strings.Contains(diags[0].Message+" "+diags[0].Help, want) {
			t.Errorf("%q: got %v, want %q", markup, diags, want)
		}
	}
}

func TestParseSignalType(t *testing.T) {
	f := mustParse(t, "package p;\nuse \"rune:dom\";\nprotected int main() {\n    Signal<int> count = signal(0);\n    void(Signal<string>) f = (Signal<string> s) => trace(s());\n    return 0;\n}\n")
	d := f.Funcs[0].Body.Stmts[0].(*ast.VarDecl)
	if d.Type.String() != "Signal<int>" || d.Name.Name != "count" {
		t.Errorf("got %s %s", d.Type, d.Name.Name)
	}
	if got := f.Funcs[0].Body.Stmts[1].(*ast.VarDecl).Type.String(); got != "void(Signal<string>)" {
		t.Errorf("got %s", got)
	}
}
