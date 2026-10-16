// Package types defines RuneScript's semantic types.
//
// The types are int/sint (one signed type, two spellings), uint, the
// floating-point types float, float32, and float64, bool, string, void,
// function types, the built-in error types (Error and its subtypes), and
// rune:dom's Element and Signal<T>.
// Integer ranges are the provisional default representations (int32 /
// uint32); integer widening is not implemented (OQ-NUM-01, OQ-NUM-02).
package types

import (
	"math"
	"math/big"
)

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
	// Char is one Unicode scalar value: 0 to 0x10FFFF, except the
	// surrogates 0xD800 to 0xDFFF (OQ-CHAR-01).
	Char
	// Array is Array<Elem>: a growable list, held by reference
	// (OQ-TYPE-02a).
	Array
	// Struct is a user-defined struct: a value type with named Fields
	// (OQ-STRUCT-01). Each declaration is its own type.
	Struct
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
	// as float32 and widens (see Float), or 32 or 64. For integer types it
	// is 0 for the generic 'int'/'sint' and 'uint' (32-bit for now,
	// OQ-NUM-01), or 8, 16, 32, or 64 for the explicit-width types.
	Bits int
	// Super is the supertype of an error type other than Error.
	Super *Type
	// Elem is the value type of a Signal or an Array.
	Elem *Type
	// Fields are the fields of a Struct, in declaration order.
	Fields []*Field
}

// Field is a field of a struct type.
type Field struct {
	Name string
	Type *Type
}

// NewStruct returns a new struct type named name. Its fields are set once
// the declaration's field types are resolved.
func NewStruct(name string) *Type { return &Type{Kind: Struct, Name: name} }

// FieldIndex returns the index of the field named name in struct type t,
// or -1.
func FieldIndex(t *Type, name string) int {
	for i, f := range t.Fields {
		if f.Name == name {
			return i
		}
	}
	return -1
}

// Comparable reports whether == and != apply to values of type t: numbers,
// bools, strings, chars, and structs whose fields are all comparable.
func Comparable(t *Type) bool {
	switch t.Kind {
	case Signed, Unsigned, Float, Bool, String, Char:
		return true
	case Struct:
		for _, f := range t.Fields {
			if !Comparable(f.Type) {
				return false
			}
		}
		return true
	}
	return false
}

// Traceable reports whether trace can print values of type t: anything
// but void, functions, arrays, the rune:dom types, and structs holding
// one of those.
func Traceable(t *Type) bool {
	switch t.Kind {
	case Signed, Unsigned, Float, Bool, String, Char, Error:
		return true
	case Struct:
		for _, f := range t.Fields {
			if !Traceable(f.Type) {
				return false
			}
		}
		return true
	}
	return false
}

// NewSignal returns the type Signal<elem>.
func NewSignal(elem *Type) *Type {
	return &Type{Kind: Signal, Name: "Signal<" + elem.Name + ">", Elem: elem}
}

// NewArray returns the type Array<elem>.
func NewArray(elem *Type) *Type {
	return &Type{Kind: Array, Name: "Array<" + elem.Name + ">", Elem: elem}
}

// IsElemType reports whether an array can hold values of type t: any
// value type except the rune:dom types, which belong to the web layer.
func IsElemType(t *Type) bool {
	switch t.Kind {
	case Signed, Unsigned, Float, Bool, String, Char, Func, Error, Array, Struct:
		return true
	}
	return false
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
	Int  = &Type{Kind: Signed, Name: "int"}
	Sint = &Type{Kind: Signed, Name: "sint"}
	Uint = &Type{Kind: Unsigned, Name: "uint"}

	// The explicit-width integers (OQ-NUM-07). They never widen, and each
	// is a type of its own: an int32 is not an int, because the generic
	// int may widen one day (CLAUDE.md §4.4), so moving a value between
	// them needs a conversion.
	Int8         = &Type{Kind: Signed, Name: "int8", Bits: 8}
	Int16        = &Type{Kind: Signed, Name: "int16", Bits: 16}
	Int32        = &Type{Kind: Signed, Name: "int32", Bits: 32}
	Int64        = &Type{Kind: Signed, Name: "int64", Bits: 64}
	Uint8        = &Type{Kind: Unsigned, Name: "uint8", Bits: 8}
	Uint16       = &Type{Kind: Unsigned, Name: "uint16", Bits: 16}
	Uint32       = &Type{Kind: Unsigned, Name: "uint32", Bits: 32}
	Uint64       = &Type{Kind: Unsigned, Name: "uint64", Bits: 64}
	BoolT        = &Type{Kind: Bool, Name: "bool"}
	StringT      = &Type{Kind: String, Name: "string"}
	CharT        = &Type{Kind: Char, Name: "char"}
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
	"int": Int, "sint": Sint, "uint": Uint, "bool": BoolT, "string": StringT, "char": CharT, "void": VoidT,
	"float": FloatT, "float32": Float32, "float64": Float64,
	"int8": Int8, "int16": Int16, "int32": Int32, "int64": Int64,
	"uint8": Uint8, "uint16": Uint16, "uint32": Uint32, "uint64": Uint64,
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
	case Float, Signed, Unsigned:
		return a.Bits == b.Bits
	case Error, Struct:
		return a == b
	case Signal, Array:
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

// ValidChar reports whether n is a Unicode scalar value, the values a
// char can hold.
func ValidChar(n int64) bool {
	return n >= 0 && n <= 0x10FFFF && (n < 0xD800 || n > 0xDFFF)
}

// Representable ranges of the v0.1 default integer representations.
const (
	SignedMin   int64 = math.MinInt32
	SignedMax   int64 = math.MaxInt32
	UnsignedMax int64 = math.MaxUint32
)

// IntBits is the width of integer type t: 32 for the generic int and uint.
func IntBits(t *Type) int {
	if t.Bits == 0 {
		return 32
	}
	return t.Bits
}

// IsWide reports whether t is a 64-bit integer type, whose values do not
// all fit in a float64 (on the JavaScript target they are BigInts).
func IsWide(t *Type) bool { return IntBits(t) == 64 && (t.Kind == Signed || t.Kind == Unsigned) }

// IntRange returns the smallest and largest values of integer type t.
func IntRange(t *Type) (lo, hi *big.Int) {
	bits := uint(IntBits(t))
	if t.Kind == Unsigned {
		hi = new(big.Int).Lsh(big.NewInt(1), bits)
		return big.NewInt(0), hi.Sub(hi, big.NewInt(1))
	}
	hi = new(big.Int).Lsh(big.NewInt(1), bits-1)
	lo = new(big.Int).Neg(hi)
	return lo, hi.Sub(hi, big.NewInt(1))
}

// Fits reports whether the exact value v is representable in integer type t.
func Fits(t *Type, v *big.Int) bool {
	lo, hi := IntRange(t)
	return v.Cmp(lo) >= 0 && v.Cmp(hi) <= 0
}

// An integer value of any type is held in an int64: its value for every
// type but uint64, whose values above math.MaxInt64 are held as their
// two's-complement bit pattern. Big and FromBig convert between that
// representation and the exact value.

// Big returns the exact value of v, an integer of type t.
func Big(t *Type, v int64) *big.Int {
	if t.Kind == Unsigned && IntBits(t) == 64 {
		return new(big.Int).SetUint64(uint64(v))
	}
	return big.NewInt(v)
}

// FromBig returns the int64 representation of v, which fits in t.
func FromBig(t *Type, v *big.Int) int64 {
	if t.Kind == Unsigned && IntBits(t) == 64 {
		return int64(v.Uint64())
	}
	return v.Int64()
}

// InRange reports whether v, an exact value (never a uint64 bit pattern),
// is representable in integer type t.
func InRange(t *Type, v int64) bool {
	bits := IntBits(t)
	switch t.Kind {
	case Signed:
		return bits == 64 || (v >= -(1<<(bits-1)) && v <= 1<<(bits-1)-1)
	case Unsigned:
		return v >= 0 && (bits == 64 || v <= 1<<bits-1)
	}
	return false
}
