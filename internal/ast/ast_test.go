package ast

import "testing"

func TestTypeRefString(t *testing.T) {
	i := &TypeRef{Name: "int"}
	s := &TypeRef{Name: "string"}
	fn := &TypeRef{Func: &FuncTypeRef{Result: i, Params: []*TypeRef{i, i}}}
	hof := &TypeRef{Func: &FuncTypeRef{Result: fn, Params: []*TypeRef{s}}}
	none := &TypeRef{Func: &FuncTypeRef{Result: &TypeRef{Name: "void"}}}
	for _, c := range []struct {
		ref  *TypeRef
		want string
	}{{i, "int"}, {fn, "int(int, int)"}, {hof, "int(int, int)(string)"}, {none, "void()"}} {
		if got := c.ref.String(); got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}
}

func TestVisibilityString(t *testing.T) {
	if VisPublic.String() != "public" || VisPrivate.String() != "private" || VisProtected.String() != "protected" {
		t.Error("visibility names wrong")
	}
}
