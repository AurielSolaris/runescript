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
)

// Type is a semantic type. Name preserves the spelling used in source so
// that diagnostics can say 'sint' when the programmer wrote 'sint'.
type Type struct {
	Kind Kind
	Name string
}

func (t *Type) String() string { return t.Name }

var (
	Int          = &Type{Signed, "int"}
	Sint         = &Type{Signed, "sint"}
	Uint         = &Type{Unsigned, "uint"}
	BoolT        = &Type{Bool, "bool"}
	StringT      = &Type{String, "string"}
	VoidT        = &Type{Void, "void"}
	UntypedConst = &Type{UntypedInt, "untyped integer constant"}
	InvalidT     = &Type{Invalid, "invalid"}
)

var byName = map[string]*Type{
	"int": Int, "sint": Sint, "uint": Uint, "bool": BoolT, "string": StringT, "void": VoidT,
}

// Lookup returns the built-in type with the given name, or nil.
func Lookup(name string) *Type { return byName[name] }

// Identical reports whether a and b are the same type. int and sint are
// identical: sint is only an alternative spelling (CLAUDE.md §4.3).
func Identical(a, b *Type) bool { return a.Kind == b.Kind }

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
