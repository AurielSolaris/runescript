// Package types defines RuneScript's semantic types.
//
// The types are int/sint (one signed type, two spellings), uint, the
// floating-point types float, float32, and float64, bool, string, void,
// function types, the built-in error types (Error and its subtypes), and
// rune:dom's Element and Signal<T>.
// Integer ranges are the provisional default representations (int32 /
// uint32); integer widening is not implemented (OQ-NUM-01, OQ-NUM-02).
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
	// Float is a floating-point type; Bits tells which one.
	Float
	// UntypedFloat is the type of a floating-point literal before it is
	// given a destination type.
	UntypedFloat
	// Error is a built-in error type: Error or one of its subtypes.
	Error
	// Element is a DOM element built by a template (rune:dom).
	Element
	// Signal is a rune: a reactive cell holding a value of type Elem
	// (rune:dom). Signal<T> is a built-in; general generics are open
	// (OQ-TYPE-01).
	Signal
)

// Type is a semantic type. Name preserves the spelling used in source so
// that diagnostics can say 'sint' when the programmer wrote 'sint'.
type Type struct {
	Kind Kind
	Name string
	// Params and Result are set for Func types only.
	Params []*Type
	Result *Type
	// Bits is set for Float types: 0 for the generic 'float', which starts
	// as float32 and widens (see Float), or 32 or 64.
	Bits int
	// Super is the supertype of an error type other than Error.
	Super *Type
	// Elem is the value type of a Signal.
	Elem *Type
}

// NewSignal returns the type Signal<elem>.
func NewSignal(elem *Type) *Type {
	return &Type{Kind: Signal, Name: "Signal<" + elem.Name + ">", Elem: elem}
}

// ElementT is the type of DOM elements built by templates.
var ElementT = &Type{Kind: Element, Name: "Element"}

// IsScalar reports whether t is a number, bool, or string type: the types
// a Signal can hold.
func IsScalar(t *Type) bool {
	switch t.Kind {
	case Signed, Unsigned, Float, Bool, String:
		return true
	}
	return false
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

	// FloatT is the generic 'float'. Its default representation is
	// float32, and a value widens to float64 whenever float32 cannot hold
	// it exactly, in magnitude or precision; it never contracts (CLAUDE.md
	// §4.4). Widening never changes a value, so a float's value is always
	// the one float64 arithmetic gives; the representation becomes
	// observable only once floats can be stored in memory layouts (arrays,
	// structs), which is open (OQ-NUM-01, OQ-FLOAT-01).
	FloatT = &Type{Kind: Float, Name: "float"}
	// Floats follow IEEE 754, as ActionScript 3's Number does: overflow
	// gives Infinity, and 0/0 gives NaN.
	//
	// Float32 and Float64 are exact IEEE 754 binary32 and binary64 types.
	// They never widen; float32 arithmetic rounds to float32.
	Float32           = &Type{Kind: Float, Name: "float32", Bits: 32}
	Float64           = &Type{Kind: Float, Name: "float64", Bits: 64}
	UntypedFloatConst = &Type{Kind: UntypedFloat, Name: "untyped float constant"}

	// The built-in error types, modelled on ActionScript 3's. They become
	// ordinary classes once classes exist.
	ErrorT         = &Type{Kind: Error, Name: "Error"}
	RangeErrorT    = &Type{Kind: Error, Name: "RangeError", Super: ErrorT}
	TypeErrorT     = &Type{Kind: Error, Name: "TypeError", Super: ErrorT}
	ArgumentErrorT = &Type{Kind: Error, Name: "ArgumentError", Super: ErrorT}
)

var byName = map[string]*Type{
	"int": Int, "sint": Sint, "uint": Uint, "bool": BoolT, "string": StringT, "void": VoidT,
	"float": FloatT, "float32": Float32, "float64": Float64,
	"Error": ErrorT, "RangeError": RangeErrorT, "TypeError": TypeErrorT, "ArgumentError": ArgumentErrorT,
}

// ErrorTypes lists the built-in error types, Error first.
var ErrorTypes = []*Type{ErrorT, RangeErrorT, TypeErrorT, ArgumentErrorT}

// IsSubtype reports whether a value of type s can be used where t is
// expected: s and t are identical, or s is an error type below t.
func IsSubtype(s, t *Type) bool {
	if Identical(s, t) {
		return true
	}
	if s.Kind != Error || t.Kind != Error {
		return false
	}
	for ; s != nil; s = s.Super {
		if s == t {
			return true
		}
	}
	return false
}

// IsNumeric reports whether t is an integer or floating-point type,
// including untyped constants.
func IsNumeric(t *Type) bool {
	return IsInteger(t) || t.Kind == Float || t.Kind == UntypedFloat
}

// Float32Max is the largest finite float32.
const Float32Max = math.MaxFloat32

// Lookup returns the built-in type with the given name, or nil.
func Lookup(name string) *Type { return byName[name] }

// Identical reports whether a and b are the same type. int and sint are
// identical: sint is only an alternative spelling (CLAUDE.md §4.3).
// Function types are identical when their parameter and result types are.
func Identical(a, b *Type) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case Float:
		return a.Bits == b.Bits
	case Error:
		return a == b
	case Signal:
		return Identical(a.Elem, b.Elem)
	case Func:
	default:
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
