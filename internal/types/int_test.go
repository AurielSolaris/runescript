package types

import (
	"math"
	"math/big"
	"testing"
)

func TestIntRanges(t *testing.T) {
	cases := []struct {
		t      *Type
		lo, hi string
	}{
		{Int, "-2147483648", "2147483647"},
		{Uint, "0", "4294967295"},
		{Int8, "-128", "127"},
		{Int16, "-32768", "32767"},
		{Int32, "-2147483648", "2147483647"},
		{Int64, "-9223372036854775808", "9223372036854775807"},
		{Uint8, "0", "255"},
		{Uint16, "0", "65535"},
		{Uint32, "0", "4294967295"},
		{Uint64, "0", "18446744073709551615"},
	}
	for _, c := range cases {
		lo, hi := IntRange(c.t)
		if lo.String() != c.lo || hi.String() != c.hi {
			t.Errorf("%s: range %s..%s, want %s..%s", c.t, lo, hi, c.lo, c.hi)
		}
		// InRange agrees with Fits at and around the bounds.
		for _, v := range []*big.Int{lo, hi, new(big.Int).Sub(lo, big.NewInt(1)), new(big.Int).Add(hi, big.NewInt(1))} {
			if !v.IsInt64() {
				continue
			}
			if got, want := InRange(c.t, v.Int64()), Fits(c.t, v); got != want {
				t.Errorf("%s: InRange(%s) = %v, Fits = %v", c.t, v, got, want)
			}
		}
	}
}

func TestUint64Representation(t *testing.T) {
	max := new(big.Int).SetUint64(math.MaxUint64)
	if r := FromBig(Uint64, max); r != -1 || Big(Uint64, r).Cmp(max) != 0 {
		t.Errorf("uint64 max round trip: %d -> %s", r, Big(Uint64, r))
	}
	if Identical(Int, Int32) || Identical(Uint, Uint32) || !Identical(Int, Sint) || Identical(Int8, Uint8) {
		t.Error("integer type identity is wrong")
	}
}
