package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/loader"
)

// app is a component file: use "rune:dom", the given helpers, and App
// with body. App's first body line is line 4 + the lines of extra.
func app(extra, body string) string {
	return "package p;\nuse \"rune:dom\";\n" + extra + "public Element App() {\n" + body + "\n}\nprotected int main() { dom.renderApp(App, \"#app\"); return 0; }\n"
}

func tmpl(markup string) string { return app("", "    template {\n"+markup+"\n    }") }

func TestDOMAccepted(t *testing.T) {
	cases := map[string]string{
		"counter": app("", `    Signal<int> count = signal(0);
    private void inc() { count.set(count() + 1); }
    template {
        <div><p>{{ count() }} / {{ count.get() }}</p><button onclick={{ inc }} disabled={{ count() > 9 }}>+</button></div>
    }`),
		"lambda handler":    tmpl(`<button type="button" onclick={{ () => trace("hi") }}>hi</button>`),
		"svg":               tmpl(`<svg viewBox="0 0 10 10" width="10"><circle cx="5" cy="5" r="4" fill="red" fill-rule="evenodd"/><foreignObject><b>x</b></foreignObject></svg>`),
		"static values":     tmpl(`<form method="post"><input type="number" min="1" maxlength="3" required checked="checked"><td colspan="2"></td><a href="/x?a=1&amp;b=2" target="_blank">x</a><meter value="0.5"></meter></form>`),
		"data and aria":     tmpl(`<div data-user-id="7" aria-label="x" role="button" tabindex="-1">x</div>`),
		"bound values":      app("", "    Signal<string> c = signal(\"a\");\n    Signal<uint> n = signal(2);\n    Signal<float32> f = signal(1.5);\n    template {\n        <div class={{ c() }} title={{ (string) n() }}><td colspan={{ n() }}></td><meter value={{ 0.25 }}></meter><svg><rect width={{ f() }} height={{ \"10px\" }} x={{ 3 }}/></svg>{{ f() }}{{ true }}</div>\n    }"),
		"child component":   app("public Element Badge(string s) {\n    template { <b>{{ s }}</b> }\n}\n", "    template {\n        <p>{{ Badge(\"new\") }}</p>\n    }"),
		"if chain":          app("", "    Signal<int> n = signal(0);\n    template {\n        <p>$if (n() == 0) zero $elseif (n() < 0) <b>neg</b> $else pos $endif</p>\n    }"),
		"rune parameter":    app("public void bump(Signal<int> s) { s.set(s() + 1); }\n", "    Signal<int> n = signal(0);\n    template {\n        <button onclick={{ () => bump(n) }}>{{ n() }}</button>\n    }"),
		"inferred rune":     app("", "    Signal<string> s = signal(\"x\");\n    trace(s());\n    template { <p>{{ s() }}</p> }"),
		"lambda component":  app("", "    Element() inner = () => {\n        template { <i>x</i> }\n    };\n    template { <p>{{ inner() }}</p> }"),
		"void main renders": "package p;\nuse \"rune:dom\";\npublic Element App() { template { <p>x</p> } }\nprotected int main() { dom.renderApp(App, \"#app\"); return 0; }\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if diags := check(t, src); len(diags) != 0 {
				t.Errorf("unexpected diagnostics: %v", diags)
			}
		})
	}
}

func TestDOMRejected(t *testing.T) {
	cases := []struct {
		name, src string
		code      diagnostics.Code
		msg       string // in the message or help
	}{
		{"no rune:dom", "package p;\npublic Element App() { template { <p>x</p> } }\nprotected int main() { return 0; }\n", diagnostics.TypeUnknown, `use "rune:dom"`},
		{"template not last", app("", "    template { <p>x</p> }\n    trace(1);"), diagnostics.TmplPlacement, "last statement"},
		{"template in void function", "package p;\nuse \"rune:dom\";\npublic void f() { template { <p>x</p> } }\nprotected int main() { f(); return 0; }\n", diagnostics.TmplPlacement, "returns 'void'"},
		{"template in if", app("", "    if (true) { template { <p>x</p> } }\n    template { <p>y</p> }"), diagnostics.TmplPlacement, "last statement"},
		{"unknown element", tmpl("<dvi>x</dvi>"), diagnostics.TmplUnknownElement, "unknown element <dvi>"},
		{"uppercase html", tmpl("<DIV>x</DIV>"), diagnostics.TmplUnknownElement, "lowercase: <div>"},
		{"svg outside svg", tmpl("<circle r=\"1\"/>"), diagnostics.TmplUnknownElement, "inside <svg>"},
		{"html inside svg", tmpl("<svg><div>x</div></svg>"), diagnostics.TmplUnknownElement, "foreignObject"},
		{"svg case", tmpl("<svg><lineargradient/></svg>"), diagnostics.TmplUnknownElement, "<linearGradient>"},
		{"custom element", tmpl("<my-widget></my-widget>"), diagnostics.TmplUnknownElement, "custom elements"},
		{"unknown attribute", tmpl(`<div href="/x">x</div>`), diagnostics.TmplUnknownAttr, "<div> has no attribute 'href'"},
		{"uppercase attribute", tmpl(`<div Class="x">x</div>`), diagnostics.TmplUnknownAttr, "lowercase: class"},
		{"bad data name", tmpl(`<div data-userId="x">x</div>`), diagnostics.TmplUnknownAttr, "lowercase: data-userid"}, {"bool with value", tmpl(`<input disabled="false">`), diagnostics.TmplBadValue, "boolean attribute"},
		{"enum value", tmpl(`<input type="textarea">`), diagnostics.TmplBadValue, `"email"`},
		{"uint value", tmpl(`<td colspan="two"></td>`), diagnostics.TmplBadValue, "non-negative integer"},
		{"length value", tmpl(`<svg><circle r="big"/></svg>`), diagnostics.TmplBadValue, "length"},
		{"script url", tmpl(`<a href=" javascript:alert(1)">x</a>`), diagnostics.TmplBadValue, "would run script"},
		{"static event", tmpl(`<button onclick="go()">x</button>`), diagnostics.TmplBadValue, "RuneScript function"},
		{"missing value", tmpl(`<a href>x</a>`), diagnostics.TmplBadValue, "needs a value"},
		{"duplicate attribute", tmpl(`<p id="a" id="b">x</p>`), diagnostics.TmplDuplicateAttr, "already set"},
		{"rune not called", app("", "    Signal<int> n = signal(0);\n    template { <p>{{ n }}</p> }"), diagnostics.TmplBadInterp, "{{ name() }}"},
		{"function as text", app("", "    template { <p>{{ main }}</p> }"), diagnostics.TmplBadInterp, "cannot be shown"},
		{"int for string attribute", app("", "    template { <p title={{ 3 }}>x</p> }"), diagnostics.TmplBadInterp, "(string) n"},
		{"string for bool", app("", "    template { <input disabled={{ \"yes\" }}> }"), diagnostics.TypeMismatch, "'bool'"},
		{"handler type", app("", "    template { <button onclick={{ (int x) => trace(x) }}>x</button> }"), diagnostics.TypeMismatch, "void()"},
		{"int condition", app("", "    template { <p>$if (1) x $endif</p> }"), diagnostics.TypeCondition, "bool"},
		{"rune of error", app("", "    Signal<Error> e = signal(new Error(\"x\"));\n    template { <p>x</p> }"), diagnostics.TypeArgs, "number, bool, or string"},
		{"signal without arg", app("", "    Signal<int> n = signal();\n    template { <p>x</p> }"), diagnostics.TypeArgCount, "initial value"},
		{"set wrong type", app("", "    Signal<int> n = signal(0);\n    n.set(\"x\");\n    template { <p>x</p> }"), diagnostics.TypeMismatch, "'int'"},
		{"unknown rune method", app("", "    Signal<int> n = signal(0);\n    n.update(1);\n    template { <p>x</p> }"), diagnostics.TypeNoField, "no method 'update'"},
		{"capture non-rune", app("", "    int step = 1;\n    void() f = () => trace(step);\n    template { <p>x</p> }"), diagnostics.NameAcrossBoundary, "pass 'step'"},
		{"assign captured rune", app("", "    Signal<int> n = signal(0);\n    void() f = () => { n = signal(1); };\n    template { <p>x</p> }"), diagnostics.NameAcrossBoundary, ""},
		{"wrong component", "package p;\nuse \"rune:dom\";\npublic Element App(int x) { template { <p>x</p> } }\nprotected int main() { dom.renderApp(App, \"#app\"); return 0; }\n", diagnostics.TypeMismatch, "Element()"},
		{"unknown dom function", "package p;\nuse \"rune:dom\";\nprotected int main() { dom.render(\"#app\"); return 0; }\n", diagnostics.NameNotInPackage, "renderApp"},
		{"unknown rune module", "package p;\nuse \"rune:net\";\nprotected int main() { return 0; }\n", diagnostics.UseUnknownPackage, "rune:dom"},
		{"type args on Error", "package p;\nprotected int main() { Error<int> e = new Error(\"x\"); trace(e.message); return 0; }\n", diagnostics.TypeArgs, "no type arguments"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			diags := check(t, c.src)
			for _, d := range diags {
				if d.Code == c.code && strings.Contains(d.Message+" "+d.Help, c.msg) {
					return
				}
			}
			t.Errorf("got %v, want %s with %q", diags, c.code, c.msg)
		})
	}
}

func TestDOMWarnings(t *testing.T) {
	cases := map[string]string{
		tmpl(`<center>x</center>`):                               "ASX-WARN-007",
		tmpl(`<table bgcolor="red"></table>`):                    "ASX-WARN-007",
		tmpl(`<svg><tref/></svg>`):                               "ASX-WARN-007",
		tmpl(`<svg><use xlink:href="#a"/></svg>`):                "ASX-WARN-007",
		strings.Replace(tmpl(`<p>x</p>`), `"#app"`, `".app"`, 1): "ASX-WARN-008",
	}
	for src, want := range cases {
		got := strings.Join(warnings(t, src), ",")
		if !strings.Contains(got, want) {
			t.Errorf("%s\ngot %s, want %s", src, got, want)
		}
	}
	// A file that uses rune:dom only for its types still uses it.
	got := warnings(t, "package p;\nuse \"rune:dom\";\npublic Element App() { template { <p>x</p> } }\nprotected int main() { return 0; }\n")
	if strings.Contains(strings.Join(got, ","), "WARN-003") {
		t.Errorf("use rune:dom reported unused: %v", got)
	}
	obsolete := check(t, tmpl(`<center>x</center>`))
	if len(obsolete) != 0 {
		t.Errorf("an obsolete element is not an error: %v", obsolete)
	}
}

func TestQualifiedFunctionValues(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "geom"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(dir, "geom/a.rsm", "package geom;\npublic int area(int w, int h) { return w * h; }\nprotected int hidden() { return 0; }\n"); err != nil {
		t.Fatal(err)
	}
	for src, want := range map[string]string{
		"package p;\nuse \"geom\";\nprotected int main() { int(int, int) f = geom.area; trace(f(1, 2)); return 0; }\n":              "",
		"package p;\nuse \"geom\";\nprotected int main() { int() f = geom.hidden; trace(f()); return 0; }\n":                        "ASX-NAME-011",
		"package p;\nuse \"geom\";\nprotected int main() { int() f = geom.nope; trace(f()); return 0; }\n":                          "ASX-NAME-008",
		"package p;\nuse \"geom\";\nprotected int main() { int() f = geom.area; trace(f()); return 0; }\n":                          "ASX-TYPE-001",
		"package p;\nuse \"std:string\";\nprotected int main() { string(string) f = string.toUpper; trace(f(\"a\")); return 0; }\n": "ASX-NAME-004",
	} {
		prog, diags := loader.LoadSource(diagnostics.NewSource("t.rsc", []byte(src)), dir)
		if len(diags) != 0 {
			t.Fatal(diags)
		}
		_, diags = Check(prog)
		var codes []string
		for _, d := range diags {
			codes = append(codes, d.Code.String())
		}
		if got := strings.Join(codes, ","); !strings.Contains(got, want) || (want == "" && got != "") {
			t.Errorf("%s: got %v, want %s", src, diags, want)
		}
	}
}
