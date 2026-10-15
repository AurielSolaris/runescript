// Package types defines RuneScript's semantic types.
//
// Only the v0.1 subset exists: int/sint (one signed type, two spellings),
// uint, bool, string, and void. Integer ranges are the provisional default
// representations (int32 / uint32); widening is not implemented
// (OQ-NUM-01, OQ-NUM-02).
package types

import "math"

// Kind is the semantic identity of a type.
type Kind int

const (
	Invalid Kind = iota
	Signed
	Unsigned
	Bool
	String
	Void
	// UntypedInt is the type of an integer constant expression before it is
	// given a destination type (OQ-NUM-03).
	UntypedInt
	// Func is a function type such as int(int, int).
	Func
)

// Type is a semantic type. Name preserves the spelling used in source so
// that diagnostics can say 'sint' when the programmer wrote 'sint'.
type Type struct {
	Kind Kind
	Name string
	// Params and Result are set for Func types only.
	Params []*Type
	Result *Type
}

// NewFunc returns the function type result(params...).
func NewFunc(params []*Type, result *Type) *Type {
	name := result.Name + "("
	for i, p := range params {
		if i > 0 {
			name += ", "
		}
		name += p.Name
	}
	return &Type{Kind: Func, Name: name + ")", Params: params, Result: result}
}

func (t *Type) String() string { return t.Name }

var (
	Int          = &Type{Kind: Signed, Name: "int"}
	Sint         = &Type{Kind: Signed, Name: "sint"}
	Uint         = &Type{Kind: Unsigned, Name: "uint"}
	BoolT        = &Type{Kind: Bool, Name: "bool"}
	StringT      = &Type{Kind: String, Name: "string"}
	VoidT        = &Type{Kind: Void, Name: "void"}
	UntypedConst = &Type{Kind: UntypedInt, Name: "untyped integer constant"}
	InvalidT     = &Type{Kind: Invalid, Name: "invalid"}
)

var byName = map[string]*Type{
	"int": Int, "sint": Sint, "uint": Uint, "bool": BoolT, "string": StringT, "void": VoidT,
}

// Lookup returns the built-in type with the given name, or nil.
func Lookup(name string) *Type { return byName[name] }

// Identical reports whether a and b are the same type. int and sint are
// identical: sint is only an alternative spelling (CLAUDE.md §4.3).
// Function types are identical when their parameter and result types are.
func Identical(a, b *Type) bool {
	if a.Kind != b.Kind {
		return false
	}
	if a.Kind != Func {
		return true
	}
	if len(a.Params) != len(b.Params) || !Identical(a.Result, b.Result) {
		return false
	}
	for i := range a.Params {
		if !Identical(a.Params[i], b.Params[i]) {
			return false
		}
	}
	return true
}

// IsInteger reports whether t is a signed, unsigned, or untyped integer.
func IsInteger(t *Type) bool {
	return t.Kind == Signed || t.Kind == Unsigned || t.Kind == UntypedInt
}

// Representable ranges of the v0.1 default integer representations.
const (
	SignedMin   int64 = math.MinInt32
	SignedMax   int64 = math.MaxInt32
	UnsignedMax int64 = math.MaxUint32
)

// InRange reports whether v is representable in integer type t.
func InRange(t *Type, v int64) bool {
	switch t.Kind {
	case Signed:
		return v >= SignedMin && v <= SignedMax
	case Unsigned:
		return v >= 0 && v <= UnsignedMax
	}
	return false
}
