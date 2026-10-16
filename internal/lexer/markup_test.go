package lexer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

// show renders tokens compactly: kind:text.
func show(toks []Token) string {
	var parts []string
	for _, t := range toks {
		name := t.Kind.String()
		switch t.Kind {
		case MText, MTagOpen, MTagClose, MAttr, MAttrValue, MDirective, Ident, Keyword, String, Int, Op:
			name = fmt.Sprintf("%s:%q", strings.Trim(name, "'"), t.Text)
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, " ")
}

func TestMarkup(t *testing.T) {
	src := `template {
  <div class="a &amp; b" hidden id={{ x }}>Hi {{ name() }} &lt;3<br/>
    $if (n > 0) <b>yes</b> $elseif (f(1)) no $else maybe $endif
    <!-- a comment -->
    <button onclick={{ () => { go(); } }}>$5</button>
  </div>
}
int after = 1;`
	toks, diags := Lex([]byte(src))
	if len(diags) != 0 {
		t.Fatal(diags)
	}
	got := show(toks)
	for _, want := range []string{
		`identifier:"template" '{' template text:"\n  " <:"div" attribute:"class" '=' attribute value:"a & b" attribute:"hidden" attribute:"id" '=' '{{' identifier:"x" '}}' '>'`,
		`template text:"Hi " '{{' identifier:"name" '(' ')' '}}' template text:" <3"`,
		`<:"br" '/>'`,
		`directive:"$if" '(' identifier:"n" operator:">" integer literal:"0" ')' template text:" " <:"b" '>' template text:"yes" closing tag:"b" template text:" " directive:"$elseif" '(' identifier:"f" '(' integer literal:"1" ')' ')' template text:" no " directive:"$else" template text:" maybe " directive:"$endif"`,
		`attribute:"onclick" '=' '{{' '(' ')' operator:"=>" '{' identifier:"go" '(' ')' ';' '}' '}}' '>' template text:"$5" closing tag:"button"`,
		`closing tag:"div" template text:"\n" '}' keyword:"int" identifier:"after"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing\n  %s\nin\n  %s", want, got)
		}
	}
	if strings.Contains(got, "comment") {
		t.Error("comment not skipped")
	}
}

// `template` is an ordinary name when no '{' follows.
func TestTemplateIsContextual(t *testing.T) {
	toks, diags := Lex([]byte("int template = 1; f(template);"))
	if len(diags) != 0 || strings.Contains(show(toks), "template text") {
		t.Errorf("%v %s", diags, show(toks))
	}
}

func TestMarkupErrors(t *testing.T) {
	cases := []struct {
		src  string
		code diagnostics.Code
	}{
		{"template { <div>", diagnostics.LexUnterminatedTemplate},
		{"template { <div class=a> }", diagnostics.LexUnexpectedChar},
		{"template { a { b }", diagnostics.LexUnexpectedChar},
		{"template { <p>{{ x </p> }", diagnostics.LexUnterminatedTemplate},
		{`template { <p title="x> }`, diagnostics.LexUnterminatedString},
		{"template { $if x }", diagnostics.LexUnexpectedChar},
		{"template { <p # > }", diagnostics.LexUnexpectedChar},
		{"template { <!-- x }", diagnostics.LexUnterminatedBlock},
	}
	for _, c := range cases {
		_, diags := Lex([]byte(c.src))
		if len(diags) == 0 || diags[0].Code != c.code {
			t.Errorf("%q: got %v, want %s", c.src, diags, c.code)
		}
	}
}
