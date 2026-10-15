package types

import "testing"

func TestFuncTypes(t *testing.T) {
	a := NewFunc([]*Type{Int, Int}, Int)
	b := NewFunc([]*Type{Sint, Int}, Sint)
	c := NewFunc([]*Type{Int}, Int)
	d := NewFunc([]*Type{Uint, Int}, Int)
	if a.Name != "int(int, int)" {
		t.Errorf("name = %q", a.Name)
	}
	if !Identical(a, b) {
		t.Error("int and sint spellings should not make function types differ")
	}
	if Identical(a, c) || Identical(a, d) || Identical(a, Int) {
		t.Error("different function types reported identical")
	}
	hof := NewFunc([]*Type{StringT}, c)
	if hof.Name != "int(int)(string)" {
		t.Errorf("higher-order name = %q", hof.Name)
	}
	if !Identical(hof, NewFunc([]*Type{StringT}, NewFunc([]*Type{Sint}, Int))) {
		t.Error("nested function types should compare structurally")
	}
}
