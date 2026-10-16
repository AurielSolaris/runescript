package stdlib

import (
	"strings"
	"unicode/utf8"

	"github.com/AurielSolaris/runescript/internal/types"
)

// std:string mirrors Go's strings package. Positions and lengths are UTF-8
// byte offsets, exactly as in Go. Names are camelCase versions of Go's
// (ToUpper → toUpper).
//
// Functions whose signatures need types RuneScript does not have yet are
// omitted; docs/guide/10-std-string.md lists them with the reason.
func init() {
	S, I, B, C := types.StringT, types.Int, types.BoolT, types.CharT
	str := func(a any) string { return a.(string) }
	num := func(a any) int { return int(a.(int64)) }
	// Callbacks take a char (OQ-CHAR-01). The runtime passes them as Go
	// functions of these types.
	pred := types.NewFunc([]*types.Type{C}, B)
	mapper := types.NewFunc([]*types.Type{C}, C)
	isFn := func(a any) func(rune) bool { return a.(func(rune) bool) }
	// Arrays of strings cross as []string (OQ-TYPE-02a).
	SA := types.NewArray(S)
	list := func(v []string) any {
		if v == nil {
			return []string{} // an empty array, never nil
		}
		return v
	}

	funcs := []*Func{
		fn("clone", S, func(a []any) any { return strings.Clone(str(a[0])) }, S),
		fn("compare", I, func(a []any) any { return int64(strings.Compare(str(a[0]), str(a[1]))) }, S, S),
		fn("contains", B, func(a []any) any { return strings.Contains(str(a[0]), str(a[1])) }, S, S),
		fn("containsFunc", B, func(a []any) any { return strings.ContainsFunc(str(a[0]), isFn(a[1])) }, S, pred),
		fn("containsRune", B, func(a []any) any { return strings.ContainsRune(str(a[0]), a[1].(rune)) }, S, C),
		fn("indexFunc", I, func(a []any) any { return int64(strings.IndexFunc(str(a[0]), isFn(a[1]))) }, S, pred),
		fn("indexRune", I, func(a []any) any { return int64(strings.IndexRune(str(a[0]), a[1].(rune))) }, S, C),
		fn("lastIndexFunc", I, func(a []any) any { return int64(strings.LastIndexFunc(str(a[0]), isFn(a[1]))) }, S, pred),
		fn("map", S, func(a []any) any { return strings.Map(a[0].(func(rune) rune), str(a[1])) }, mapper, S),
		fn("trimFunc", S, func(a []any) any { return strings.TrimFunc(str(a[0]), isFn(a[1])) }, S, pred),
		fn("trimLeftFunc", S, func(a []any) any { return strings.TrimLeftFunc(str(a[0]), isFn(a[1])) }, S, pred),
		fn("trimRightFunc", S, func(a []any) any { return strings.TrimRightFunc(str(a[0]), isFn(a[1])) }, S, pred),
		fn("containsAny", B, func(a []any) any { return strings.ContainsAny(str(a[0]), str(a[1])) }, S, S),
		fn("count", I, func(a []any) any { return int64(strings.Count(str(a[0]), str(a[1]))) }, S, S),
		fn("equalFold", B, func(a []any) any { return strings.EqualFold(str(a[0]), str(a[1])) }, S, S),
		fn("hasPrefix", B, func(a []any) any { return strings.HasPrefix(str(a[0]), str(a[1])) }, S, S),
		fn("hasSuffix", B, func(a []any) any { return strings.HasSuffix(str(a[0]), str(a[1])) }, S, S),
		fn("index", I, func(a []any) any { return int64(strings.Index(str(a[0]), str(a[1]))) }, S, S),
		fn("indexAny", I, func(a []any) any { return int64(strings.IndexAny(str(a[0]), str(a[1]))) }, S, S),
		fn("lastIndex", I, func(a []any) any { return int64(strings.LastIndex(str(a[0]), str(a[1]))) }, S, S),
		fn("len", I, func(a []any) any { return int64(len(str(a[0]))) }, S),
		fn("runeCount", I, func(a []any) any { return int64(utf8.RuneCountInString(str(a[0]))) }, S),
		fn("slice", S, func(a []any) any { return slice(str(a[0]), num(a[1]), num(a[2])) }, S, I, I),
		fn("fields", SA, func(a []any) any { return list(strings.Fields(str(a[0]))) }, S),
		fn("fieldsFunc", SA, func(a []any) any { return list(strings.FieldsFunc(str(a[0]), isFn(a[1]))) }, S, pred),
		fn("join", S, func(a []any) any { return strings.Join(a[0].([]string), str(a[1])) }, SA, S),
		fn("split", SA, func(a []any) any { return list(strings.Split(str(a[0]), str(a[1]))) }, S, S),
		fn("splitAfter", SA, func(a []any) any { return list(strings.SplitAfter(str(a[0]), str(a[1]))) }, S, S),
		fn("splitAfterN", SA, func(a []any) any { return list(strings.SplitAfterN(str(a[0]), str(a[1]), num(a[2]))) }, S, S, I),
		fn("splitN", SA, func(a []any) any { return list(strings.SplitN(str(a[0]), str(a[1]), num(a[2]))) }, S, S, I),
		fn("lastIndexAny", I, func(a []any) any { return int64(strings.LastIndexAny(str(a[0]), str(a[1]))) }, S, S),
		fn("repeat", S, func(a []any) any { return strings.Repeat(str(a[0]), num(a[1])) }, S, I),
		fn("replace", S, func(a []any) any { return strings.Replace(str(a[0]), str(a[1]), str(a[2]), num(a[3])) }, S, S, S, I),
		fn("replaceAll", S, func(a []any) any { return strings.ReplaceAll(str(a[0]), str(a[1]), str(a[2])) }, S, S, S),
		fn("toLower", S, func(a []any) any { return strings.ToLower(str(a[0])) }, S),
		fn("toTitle", S, func(a []any) any { return strings.ToTitle(str(a[0])) }, S),
		fn("toUpper", S, func(a []any) any { return strings.ToUpper(str(a[0])) }, S),
		fn("toValidUTF8", S, func(a []any) any { return strings.ToValidUTF8(str(a[0]), str(a[1])) }, S, S),
		fn("trim", S, func(a []any) any { return strings.Trim(str(a[0]), str(a[1])) }, S, S),
		fn("trimLeft", S, func(a []any) any { return strings.TrimLeft(str(a[0]), str(a[1])) }, S, S),
		fn("trimPrefix", S, func(a []any) any { return strings.TrimPrefix(str(a[0]), str(a[1])) }, S, S),
		fn("trimRight", S, func(a []any) any { return strings.TrimRight(str(a[0]), str(a[1])) }, S, S),
		fn("trimSpace", S, func(a []any) any { return strings.TrimSpace(str(a[0])) }, S),
		fn("trimSuffix", S, func(a []any) any { return strings.TrimSuffix(str(a[0]), str(a[1])) }, S, S),
	}
	p := &Package{Path: "string", Qualifier: "string", Funcs: map[string]*Func{}}
	for _, f := range funcs {
		p.Funcs[f.Name] = f
	}
	register(p)
}

// slice is s[start:end] in bytes (OQ-STR-03). Unlike Go, it never cuts a
// character in two, so the result stays valid UTF-8.
func slice(s string, start, end int) string {
	switch {
	case start < 0 || end > len(s) || start > end:
		rangeFault("bounds [%d:%d] are out of range for a string of %d bytes", start, end, len(s))
	case !boundary(s, start):
		rangeFault("byte offset %d is inside a character", start)
	case !boundary(s, end):
		rangeFault("byte offset %d is inside a character", end)
	}
	return s[start:end]
}

func boundary(s string, i int) bool { return i == len(s) || utf8.RuneStart(s[i]) }
