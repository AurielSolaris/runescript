package ir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/types"
)

func lower(t *testing.T, src string) *Program {
	t.Helper()
	prog, diags := loader.LoadSource(diagnostics.NewSource("t.rsc", []byte(src)), t.TempDir())
	if diags != nil {
		t.Fatal(diags)
	}
	info, diags := checker.Check(prog)
	if diags != nil {
		t.Fatal(diags)
	}
	return Lower(info)
}

// Every example lowers to IR that passes verification.
func TestLowerExamplesVerify(t *testing.T) {
	files, _ := filepath.Glob("../../examples/*.rsc")
	if len(files) == 0 {
		t.Fatal("no examples")
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_error.rsc") {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := Verify(lower(t, string(src))); err != nil {
			t.Errorf("%s: %v", path, err)
		}
	}
}

func TestLowering(t *testing.T) {
	p := lower(t, `package p;
public uint main2(uint u) {
    uint v = u;
    v += 5;
    v++;
    return v;
}
public int main() {
    int x = -3;
    return -x;
}
`)
	f := p.Funcs[0]
	add := f.Body.Stmts[1].(*Assign).Value.(*Arith)
	if add.Op != "+" || add.T.Kind != types.Unsigned || add.Y.(*IntConst).Value != 5 || add.Site.Start.String() != "4:5" {
		t.Errorf("compound assignment lowered to %+v", add)
	}
	inc := f.Body.Stmts[2].(*Assign).Value.(*Arith)
	if inc.Op != "+" || inc.Y.(*IntConst).Value != 1 {
		t.Errorf("increment lowered to %+v", inc)
	}
	m := p.Entry
	if c := m.Body.Stmts[0].(*Decl).Init.(*IntConst); c.Value != -3 || c.T.Kind != types.Signed {
		t.Errorf("constant -3 lowered to %+v", c)
	}
	neg := m.Body.Stmts[1].(*Return).Value.(*Arith)
	if neg.Op != "-" || neg.X.(*IntConst).Value != 0 {
		t.Errorf("negation lowered to %+v", neg)
	}
}

// Verify catches IR that breaks the typing rules, whatever produced it.
func TestVerifyRejects(t *testing.T) {
	u := &Var{Name: "u", Type: types.Uint}
	s := &Var{Name: "s", Type: types.Int, Param: true}
	cases := map[string]Stmt{
		"signed to unsigned":  &Decl{Var: u, Init: &Load{Var: s}},
		"mixed signedness":    &Eval{X: &Arith{Op: "+", T: types.Int, X: &Load{Var: s}, Y: &IntConst{T: types.Uint, Value: 1}}},
		"constant range":      &Decl{Var: u, Init: &IntConst{T: types.Uint, Value: -1}},
		"condition not bool":  &If{Cond: &Load{Var: s}, Then: &Block{}},
		"assign to parameter": &Assign{Var: s, Value: &IntConst{T: types.Int}},
		"break outside loop":  &Break{},
		"string arithmetic":   &Eval{X: &Arith{Op: "+", T: types.StringT, X: &StringConst{}, Y: &StringConst{}}},
		"join of an int":      &Eval{X: &Concat{Parts: []Expr{&StringConst{}, &Load{Var: s}}}},
		"empty join":          &Eval{X: &Concat{}},
		"surrogate char":      &Eval{X: &CharConst{Value: 0xD800}},
		"char arithmetic":     &Eval{X: &Arith{Op: "+", T: types.CharT, X: &CharConst{Value: 'a'}, Y: &CharConst{Value: 'b'}}},
		"char to float":       &Eval{X: &Convert{X: &CharConst{Value: 'a'}, To: types.FloatT}},
		"mixed array":         &Eval{X: &ArrayLit{T: types.NewArray(types.Int), Elems: []Expr{&StringConst{}}}},
		"index a string":      &Eval{X: &Index{X: &StringConst{}, I: &IntConst{T: types.Int}}},
		"string index":        &Eval{X: &Index{X: &ArrayLit{T: types.NewArray(types.Int)}, I: &StringConst{}}},
		"push wrong type":     &Eval{X: &ArrayPush{X: &ArrayLit{T: types.NewArray(types.Int)}, V: &StringConst{}}},
		"store wrong type":    &IndexSet{X: &ArrayLit{T: types.NewArray(types.Int)}, I: &IntConst{T: types.Int}, Value: &BoolConst{}},
		"bool attr bound to string": &Eval{X: &Template{Root: &TElement{Tag: "input",
			Attrs: []*TAttr{{Name: "disabled", Mode: AttrBool, Value: &StringConst{}}}}}},
		"keyword attr bound to int": &Eval{X: &Template{Root: &TElement{Tag: "input",
			Attrs: []*TAttr{{Name: "type", Mode: AttrText, Values: []string{"text"}, Value: &Load{Var: s}}}}}},
		"event without handler": &Eval{X: &Template{Root: &TElement{Tag: "button",
			Attrs: []*TAttr{{Name: "onclick", Mode: AttrEvent, Value: &StringConst{}}}}}},
	}
	for name, stmt := range cases {
		f := &Func{Name: "main", Type: types.NewFunc([]*types.Type{types.Int}, types.VoidT), Params: []*Var{s},
			Body: &Block{Stmts: []Stmt{stmt}}}
		err := Verify(&Program{Funcs: []*Func{f}, Entry: f})
		if err == nil || !strings.Contains(err.Error(), "internal compiler error") {
			t.Errorf("%s: Verify returned %v", name, err)
		}
	}
}
