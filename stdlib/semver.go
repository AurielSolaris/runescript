package stdlib

import (
	"fmt"
	"strings"

	"github.com/AurielSolaris/runescript/internal/types"
)

// std:semver compares version strings (OQ-MOD-01): versions follow
// SemVer 2.0.0 exactly, and comparison and ranges follow npm's semver
// package (node-semver) without its loose and includePrerelease options.
//
//   - A version is MAJOR.MINOR.PATCH, then optionally -PRERELEASE and
//     +BUILD. Numbers have no leading zeros and any number of digits.
//     There is no "v" prefix and no surrounding space: "v1.2.3" and
//     " 1.2.3" are not versions.
//   - Precedence: numbers compare as numbers; a version with a prerelease
//     is lower than the same version without one; prerelease identifiers
//     compare one by one (numeric ones as numbers, below alphanumeric ones,
//     which compare in ASCII order), and a longer list is higher when the
//     shorter one is its prefix. Build metadata is ignored.
//   - Ranges are npm's: comparators (<, <=, >, >=, =) joined by spaces,
//     sets joined by ||, x-ranges (1.2.x, 1.x, *, and partial versions
//     such as 1.2), tilde (~1.2.3, also ~>), caret (^1.2.3), and hyphen
//     ranges (1.2.3 - 2.3.4). Each is first rewritten into plain
//     comparators exactly as node-semver does. A version with a prerelease
//     satisfies a set only if some comparator of the set names the same
//     MAJOR.MINOR.PATCH with a prerelease.
//
// Every function throws an ArgumentError for an invalid version or range,
// except valid and validRange, which test for one.
//
// The JavaScript twin is section semver in internal/jsgen/runtime.js.
func init() {
	S, B, I := types.StringT, types.BoolT, types.Int
	str := func(a any) string { return a.(string) }
	cmp := func(name string, ok func(int) bool) *Func {
		return fn(name, B, func(a []any) any { return ok(svCompare(svMust(str(a[0])), svMust(str(a[1])))) }, S, S)
	}
	p := &Package{Path: "semver", Qualifier: "semver"}
	p.add(
		fn("valid", B, func(a []any) any { _, ok := svParse(str(a[0])); return ok }, S),
		fn("validRange", B, func(a []any) any { _, err := svParseRange(str(a[0])); return err == "" }, S),
		fn("compare", I, func(a []any) any { return int64(svCompare(svMust(str(a[0])), svMust(str(a[1])))) }, S, S),
		cmp("eq", func(c int) bool { return c == 0 }),
		cmp("neq", func(c int) bool { return c != 0 }),
		cmp("gt", func(c int) bool { return c > 0 }),
		cmp("gte", func(c int) bool { return c >= 0 }),
		cmp("lt", func(c int) bool { return c < 0 }),
		cmp("lte", func(c int) bool { return c <= 0 }),
		fn("satisfies", B, func(a []any) any {
			v := svMust(str(a[0]))
			r, err := svParseRange(str(a[1]))
			if err != "" {
				argFault("%s", err)
			}
			return svSatisfies(v, r)
		}, S, S),
		fn("major", I, func(a []any) any { return svInt(svMust(str(a[0])).major, "major") }, S),
		fn("minor", I, func(a []any) any { return svInt(svMust(str(a[0])).minor, "minor") }, S),
		fn("patch", I, func(a []any) any { return svInt(svMust(str(a[0])).patch, "patch") }, S),
		fn("prerelease", S, func(a []any) any { return strings.Join(svMust(str(a[0])).pre, ".") }, S),
		fn("build", S, func(a []any) any { return strings.Join(svMust(str(a[0])).build, ".") }, S),
	)
	register(p)
}

// semVersion is a parsed version. The numbers are kept as digit strings,
// so any number of digits compares exactly.
type semVersion struct {
	major, minor, patch string
	pre, build          []string
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// svNumber reports whether s is a SemVer number: digits, no leading zero.
func svNumber(s string) bool {
	return isDigits(s) && (s == "0" || s[0] != '0')
}

// svIdentChars reports whether s is a non-empty run of [0-9A-Za-z-].
func svIdentChars(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '-') {
			return false
		}
	}
	return true
}

// svIdents splits a prerelease (pre) or build list and checks it.
func svIdents(s string, pre bool) ([]string, bool) {
	ids := strings.Split(s, ".")
	for _, id := range ids {
		if !svIdentChars(id) || (pre && isDigits(id) && !svNumber(id)) {
			return nil, false
		}
	}
	return ids, true
}

// svTail parses the -PRERELEASE+BUILD after a version's numbers.
func svTail(s string, v *semVersion) bool {
	if rest, build, ok := strings.Cut(s, "+"); ok {
		ids, ok := svIdents(build, false)
		if !ok {
			return false
		}
		v.build, s = ids, rest
	}
	if s == "" {
		return true
	}
	if s[0] != '-' {
		return false
	}
	ids, ok := svIdents(s[1:], true)
	v.pre = ids
	return ok
}

// svParse parses a SemVer 2.0.0 version.
func svParse(s string) (semVersion, bool) {
	var v semVersion
	end := strings.IndexAny(s, "-+")
	if end < 0 {
		end = len(s)
	}
	nums := strings.Split(s[:end], ".")
	if len(nums) != 3 || !svNumber(nums[0]) || !svNumber(nums[1]) || !svNumber(nums[2]) {
		return v, false
	}
	v.major, v.minor, v.patch = nums[0], nums[1], nums[2]
	return v, svTail(s[end:], &v)
}

func svMust(s string) semVersion {
	v, ok := svParse(s)
	if !ok {
		argFault("%s is not a valid version: SemVer 2.0.0 is MAJOR.MINOR.PATCH, e.g. 1.2.3 or 1.2.3-rc.1", svQuote(s))
	}
	return v
}

// svQuote quotes s for an error message.
func svQuote(s string) string {
	return `"` + s + `"`
}

// svInt returns a version number as an int, which it must fit.
func svInt(digits, what string) int64 {
	if len(digits) > 10 || (len(digits) == 10 && digits > "2147483647") {
		panic(Fault{Type: "RangeError", Message: fmt.Sprintf("the %s version %s does not fit in 'int'", what, digits)})
	}
	var n int64
	for i := 0; i < len(digits); i++ {
		n = n*10 + int64(digits[i]-'0')
	}
	return n
}

// svCompareNum compares two numbers written without leading zeros.
func svCompareNum(a, b string) int {
	switch {
	case len(a) != len(b):
		return svSign(len(a) - len(b))
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func svSign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}

// svCompareIdent compares two prerelease identifiers.
func svCompareIdent(a, b string) int {
	an, bn := isDigits(a), isDigits(b)
	switch {
	case an && bn:
		return svCompareNum(a, b)
	case an:
		return -1
	case bn:
		return 1
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// svCompare returns -1, 0, or 1 as a is lower than, equal to, or higher
// than b in SemVer precedence.
func svCompare(a, b semVersion) int {
	for _, c := range []int{svCompareNum(a.major, b.major), svCompareNum(a.minor, b.minor), svCompareNum(a.patch, b.patch)} {
		if c != 0 {
			return c
		}
	}
	switch {
	case len(a.pre) == 0 && len(b.pre) == 0:
		return 0
	case len(a.pre) == 0:
		return 1
	case len(b.pre) == 0:
		return -1
	}
	for i := 0; i < len(a.pre) && i < len(b.pre); i++ {
		if c := svCompareIdent(a.pre[i], b.pre[i]); c != 0 {
			return c
		}
	}
	return svSign(len(a.pre) - len(b.pre))
}

// ---- ranges ----

// svComparator is one comparison; op "" matches every version.
type svComparator struct {
	op string
	v  semVersion
}

// svPartial is a version in a range, where parts may be missing or x.
type svPartial struct {
	nums [3]string // "" for a missing or x part
	pre  []string
}

func (p svPartial) x(i int) bool { return p.nums[i] == "" }

// svInc adds one to a number written in digits.
func svInc(d string) string {
	b := []byte(d)
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] < '9' {
			b[i]++
			return string(b)
		}
		b[i] = '0'
	}
	return "1" + string(b)
}

// svParsePartial parses an x-range such as 1, 1.2, 1.x, *, or 1.2.3-rc.1.
func svParsePartial(s string) (svPartial, bool) {
	var p svPartial
	end := strings.IndexAny(s, "-+")
	if end < 0 {
		end = len(s)
	}
	parts := strings.Split(s[:end], ".")
	if len(parts) > 3 {
		return p, false
	}
	wild := false
	for i, part := range parts {
		switch {
		case part == "x" || part == "X" || part == "*":
			wild = true
		case svNumber(part):
			if !wild { // node-semver: after an x, the rest is x too (1.x.3 is 1.x)
				p.nums[i] = part
			}
		default:
			return p, false
		}
	}
	if end < len(s) {
		// node-semver accepts a prerelease and build after three parts
		// even when one is x, and then ignores them: 1.x.0-rc is 1.x.
		if len(parts) != 3 {
			return p, false
		}
		var v semVersion
		if !svTail(s[end:], &v) {
			return p, false
		}
		if !wild {
			p.pre = v.pre
		}
	}
	return p, true
}

// svFull is the version M.m.p (with p's prerelease when pre is set).
func svFull(M, m, pt string, pre []string) semVersion {
	return semVersion{major: M, minor: m, patch: pt, pre: pre}
}

var svZeroPre = []string{"0"}

// svNone is a comparator no version satisfies (node-semver's <0.0.0-0).
var svNone = svComparator{"<", svFull("0", "0", "0", svZeroPre)}

// svDesugar rewrites one range term (an operator and a partial version)
// into plain comparators, as node-semver's replaceTilde, replaceCaret, and
// replaceXRange do.
func svDesugar(op string, p svPartial) []svComparator {
	M, m, pt := p.nums[0], p.nums[1], p.nums[2]
	xM, xm, xp := p.x(0), p.x(0) || p.x(1), p.x(0) || p.x(1) || p.x(2)
	ge := func(v semVersion) svComparator { return svComparator{">=", v} }
	lt := func(v semVersion) svComparator { return svComparator{"<", v} }
	switch op {
	case "~", "~>":
		switch {
		case xM:
			return nil
		case xm:
			return []svComparator{ge(svFull(M, "0", "0", nil)), lt(svFull(svInc(M), "0", "0", svZeroPre))}
		case xp:
			return []svComparator{ge(svFull(M, m, "0", nil)), lt(svFull(M, svInc(m), "0", svZeroPre))}
		}
		return []svComparator{ge(svFull(M, m, pt, p.pre)), lt(svFull(M, svInc(m), "0", svZeroPre))}
	case "^":
		switch {
		case xM:
			return nil
		case xm:
			return []svComparator{ge(svFull(M, "0", "0", nil)), lt(svFull(svInc(M), "0", "0", svZeroPre))}
		case xp:
			if M == "0" {
				return []svComparator{ge(svFull(M, m, "0", nil)), lt(svFull(M, svInc(m), "0", svZeroPre))}
			}
			return []svComparator{ge(svFull(M, m, "0", nil)), lt(svFull(svInc(M), "0", "0", svZeroPre))}
		}
		low := ge(svFull(M, m, pt, p.pre))
		switch {
		case M == "0" && m == "0":
			return []svComparator{low, lt(svFull(M, m, svInc(pt), svZeroPre))}
		case M == "0":
			return []svComparator{low, lt(svFull(M, svInc(m), "0", svZeroPre))}
		}
		return []svComparator{low, lt(svFull(svInc(M), "0", "0", svZeroPre))}
	}
	// A primitive comparator or a plain x-range.
	if !xp {
		if op == "" {
			op = "="
		}
		return []svComparator{{op, svFull(M, m, pt, p.pre)}}
	}
	if op == "=" {
		op = ""
	}
	switch {
	case xM:
		if op == ">" || op == "<" {
			return []svComparator{svNone}
		}
		return nil
	case op != "":
		if xm {
			m = "0"
		}
		pt = "0"
		switch op {
		case ">":
			op = ">="
			if xm {
				M, m = svInc(M), "0"
			} else {
				m = svInc(m)
			}
		case "<=":
			op = "<"
			if xm {
				M = svInc(M)
			} else {
				m = svInc(m)
			}
		}
		var pre []string
		if op == "<" {
			pre = svZeroPre
		}
		return []svComparator{{op, svFull(M, m, pt, pre)}}
	case xm:
		return []svComparator{ge(svFull(M, "0", "0", nil)), lt(svFull(svInc(M), "0", "0", svZeroPre))}
	}
	return []svComparator{ge(svFull(M, m, "0", nil)), lt(svFull(M, svInc(m), "0", svZeroPre))}
}

// svHyphen rewrites the hyphen range from - to, as node-semver's
// hyphenReplace does.
func svHyphen(from, to svPartial) []svComparator {
	var out []svComparator
	switch {
	case from.x(0):
	case from.x(1):
		out = append(out, svComparator{">=", svFull(from.nums[0], "0", "0", nil)})
	case from.x(2):
		out = append(out, svComparator{">=", svFull(from.nums[0], from.nums[1], "0", nil)})
	default:
		out = append(out, svComparator{">=", svFull(from.nums[0], from.nums[1], from.nums[2], from.pre)})
	}
	switch {
	case to.x(0):
	case to.x(1):
		out = append(out, svComparator{"<", svFull(svInc(to.nums[0]), "0", "0", svZeroPre)})
	case to.x(2):
		out = append(out, svComparator{"<", svFull(to.nums[0], svInc(to.nums[1]), "0", svZeroPre)})
	default:
		out = append(out, svComparator{"<=", svFull(to.nums[0], to.nums[1], to.nums[2], to.pre)})
	}
	return out
}

// svAppend adds cs to set, leaving out >=0.0.0, which matches every
// version (node-semver's replaceGTE0).
func svAppend(set, cs []svComparator) []svComparator {
	for _, c := range cs {
		if c.op == ">=" && c.v.major == "0" && c.v.minor == "0" && c.v.patch == "0" && len(c.v.pre) == 0 {
			continue
		}
		set = append(set, c)
	}
	return set
}

var svOps = []string{"~>", "~", "^", "<=", ">=", "<", ">", "="}

// svFields splits s at spaces and tabs.
func svFields(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '\t' })
}

// svParseRange parses a range into sets of comparators. It returns the
// error message for an invalid range.
func svParseRange(r string) ([][]svComparator, string) {
	bad := func(why string) ([][]svComparator, string) {
		return nil, fmt.Sprintf("%s is not a valid range: %s", svQuote(r), why)
	}
	var sets [][]svComparator
	for _, part := range strings.Split(r, "||") {
		words := svFields(part)
		set := []svComparator{}
		if len(words) == 3 && words[1] == "-" {
			from, ok1 := svParsePartial(words[0])
			to, ok2 := svParsePartial(words[2])
			if !ok1 || !ok2 {
				return bad(fmt.Sprintf("%s - %s is not a hyphen range of versions", words[0], words[2]))
			}
			sets = append(sets, svAppend([]svComparator{}, svHyphen(from, to)))
			continue
		}
		for i := 0; i < len(words); i++ {
			w := words[i]
			op := ""
			for _, o := range svOps {
				if strings.HasPrefix(w, o) {
					op = o
					break
				}
			}
			if op == w && i+1 < len(words) { // an operator written apart: >= 1.2.3
				i++
				w += words[i]
			}
			p, ok := svParsePartial(w[len(op):])
			if !ok {
				return bad(fmt.Sprintf("%s is not a comparator", svQuote(w)))
			}
			set = svAppend(set, svDesugar(op, p))
		}
		sets = append(sets, set)
	}
	// node-semver: when one set matches every version (no comparators
	// left, as for *), the range is that set alone. That matters for
	// prereleases: "^2.0.0-alpha || *" does not match 2.0.0-beta.
	if len(sets) > 1 {
		for _, set := range sets {
			if len(set) == 0 {
				return [][]svComparator{set}, ""
			}
		}
	}
	return sets, ""
}

func (c svComparator) test(v semVersion) bool {
	d := svCompare(v, c.v)
	switch c.op {
	case "":
		return true
	case "=":
		return d == 0
	case "<":
		return d < 0
	case "<=":
		return d <= 0
	case ">":
		return d > 0
	}
	return d >= 0 // ">="
}

// svSatisfies reports whether v is in the range r.
func svSatisfies(v semVersion, r [][]svComparator) bool {
	for _, set := range r {
		if svTestSet(set, v) {
			return true
		}
	}
	return false
}

func svTestSet(set []svComparator, v semVersion) bool {
	for _, c := range set {
		if !c.test(v) {
			return false
		}
	}
	if len(v.pre) == 0 {
		return true
	}
	for _, c := range set {
		if c.op != "" && len(c.v.pre) > 0 && c.v.major == v.major && c.v.minor == v.minor && c.v.patch == v.patch {
			return true
		}
	}
	return false
}
