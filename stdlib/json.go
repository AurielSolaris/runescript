package stdlib

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/AurielSolaris/runescript/internal/conv"
	"github.com/AurielSolaris/runescript/internal/types"
)

// std:json (OQ-JSON-01): JSON text and json.Value, an opaque type. A Value
// is null, a bool, a number, a string, an array, or an object; programs
// read one with checked functions (asString, get, at, ...) and build one
// with nullValue, fromString, fromArray, fromObject, and so on. Values are
// never changed once made, so they can be shared freely and cannot form
// cycles.
//
// The parser is our own, run over UTF-8 bytes, with a twin in the
// JavaScript runtime (runtime.js, section json), so both targets accept
// the same texts and fail with the same messages and byte offsets:
//
//   - Numbers keep their text: asInt and asInt64 read it exactly, asFloat
//     rounds it to the nearest float64, and stringify writes it back
//     unchanged.
//   - Objects keep their keys in document order. A repeated key keeps its
//     first position and its last value, as JavaScript's JSON.parse does.
//   - A \u escape of a lone surrogate becomes U+FFFD, since RuneScript
//     strings are valid UTF-8.
//   - Nesting deeper than jsonMaxDepth arrays and objects is an error.

// JSONValue is the type json.Value.
var JSONValue = types.NewOpaque("json.Value")

const jsonMaxDepth = 1000

// jsonValue is a json.Value.
type jsonValue struct {
	kind  string // "null", "bool", "number", "string", "array", or "object"
	b     bool
	text  string         // a number's text, or a string's value
	elems []*jsonValue   // an array's elements, or an object's values
	keys  []string       // an object's keys, in order
	index map[string]int // an object's keys, to their positions
}

var (
	jsonNull  = &jsonValue{kind: "null"}
	jsonTrue  = &jsonValue{kind: "bool", b: true}
	jsonFalse = &jsonValue{kind: "bool", b: false}
)

// jsonObject builds an object from keys and values. A repeated key keeps
// its first position and takes its last value.
func jsonObject(keys []string, vals []*jsonValue) *jsonValue {
	o := &jsonValue{kind: "object", index: map[string]int{}}
	for i, k := range keys {
		if at, ok := o.index[k]; ok {
			o.elems[at] = vals[i]
			continue
		}
		o.index[k] = len(o.keys)
		o.keys = append(o.keys, k)
		o.elems = append(o.elems, vals[i])
	}
	return o
}

// ---- parsing ----

type jsonParser struct {
	s     string
	i     int
	depth int
}

func (p *jsonParser) fail(at int, reason string) {
	panic(Fault{Type: "Error", Message: fmt.Sprintf("%s at byte %d", reason, at)})
}

// unexpected fails at the current byte: the end of the input, or a
// character that cannot come here.
func (p *jsonParser) unexpected() {
	if p.i >= len(p.s) {
		p.fail(p.i, "unexpected end of input")
	}
	r, _ := utf8.DecodeRuneInString(p.s[p.i:])
	p.fail(p.i, "unexpected character "+jsonDescribeRune(r))
}

// jsonDescribeRune shows r in a message: 'x', or U+XXXX for a control
// character.
func jsonDescribeRune(r rune) string {
	if r < 0x20 || r >= 0x7F && r < 0xA0 {
		return fmt.Sprintf("U+%04X", r)
	}
	return "'" + string(r) + "'"
}

func (p *jsonParser) space() {
	for p.i < len(p.s) {
		switch p.s[p.i] {
		case ' ', '\t', '\n', '\r':
			p.i++
		default:
			return
		}
	}
}

func (p *jsonParser) value() *jsonValue {
	p.space()
	if p.i >= len(p.s) {
		p.unexpected()
	}
	switch c := p.s[p.i]; {
	case c == '{':
		return p.object()
	case c == '[':
		return p.array()
	case c == '"':
		return &jsonValue{kind: "string", text: p.string()}
	case c == 't':
		p.word("true")
		return jsonTrue
	case c == 'f':
		p.word("false")
		return jsonFalse
	case c == 'n':
		p.word("null")
		return jsonNull
	case c == '-' || c >= '0' && c <= '9':
		return &jsonValue{kind: "number", text: p.number()}
	}
	p.unexpected()
	return nil
}

// word reads the literal w (true, false, or null).
func (p *jsonParser) word(w string) {
	for j := 0; j < len(w); j++ {
		if p.i >= len(p.s) || p.s[p.i] != w[j] {
			p.unexpected()
		}
		p.i++
	}
}

func (p *jsonParser) nest() {
	if p.depth++; p.depth > jsonMaxDepth {
		p.fail(p.i, fmt.Sprintf("more than %d nested arrays and objects", jsonMaxDepth))
	}
}

func (p *jsonParser) array() *jsonValue {
	p.nest()
	p.i++ // [
	a := &jsonValue{kind: "array", elems: []*jsonValue{}}
	p.space()
	if p.i < len(p.s) && p.s[p.i] == ']' {
		p.i++
		p.depth--
		return a
	}
	for {
		a.elems = append(a.elems, p.value())
		p.space()
		if p.i < len(p.s) && p.s[p.i] == ',' {
			p.i++
			continue
		}
		if p.i < len(p.s) && p.s[p.i] == ']' {
			p.i++
			p.depth--
			return a
		}
		p.unexpected()
	}
}

func (p *jsonParser) object() *jsonValue {
	p.nest()
	p.i++ // {
	var keys []string
	var vals []*jsonValue
	p.space()
	if p.i < len(p.s) && p.s[p.i] == '}' {
		p.i++
		p.depth--
		return jsonObject(nil, nil)
	}
	for {
		p.space()
		if p.i >= len(p.s) || p.s[p.i] != '"' {
			p.unexpected()
		}
		keys = append(keys, p.string())
		p.space()
		if p.i >= len(p.s) || p.s[p.i] != ':' {
			p.unexpected()
		}
		p.i++
		vals = append(vals, p.value())
		p.space()
		if p.i < len(p.s) && p.s[p.i] == ',' {
			p.i++
			continue
		}
		if p.i < len(p.s) && p.s[p.i] == '}' {
			p.i++
			p.depth--
			return jsonObject(keys, vals)
		}
		p.unexpected()
	}
}

// string reads a string literal, the current byte being its '"'.
func (p *jsonParser) string() string {
	p.i++
	var b strings.Builder
	for {
		if p.i >= len(p.s) {
			p.fail(p.i, "unexpected end of input")
		}
		c := p.s[p.i]
		switch {
		case c == '"':
			p.i++
			return b.String()
		case c == '\\':
			b.WriteRune(p.escape())
		case c < 0x20:
			p.fail(p.i, fmt.Sprintf("control character U+%04X in a string", c))
		default:
			b.WriteByte(c)
			p.i++
		}
	}
}

// escape reads an escape sequence, the current byte being its '\'.
func (p *jsonParser) escape() rune {
	at := p.i
	p.i++
	if p.i >= len(p.s) {
		p.fail(p.i, "unexpected end of input")
	}
	c := p.s[p.i]
	p.i++
	switch c {
	case '"', '\\', '/':
		return rune(c)
	case 'b':
		return '\b'
	case 'f':
		return '\f'
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	case 'u':
		u := p.hex4(at)
		if u < 0xD800 || u > 0xDFFF {
			return u
		}
		// A high surrogate followed by an escaped low one is one
		// character; any other surrogate is U+FFFD.
		if u < 0xDC00 && strings.HasPrefix(p.s[p.i:], `\u`) {
			save := p.i
			p.i += 2
			if lo := p.hex4(save); lo >= 0xDC00 && lo <= 0xDFFF {
				return 0x10000 + (u-0xD800)<<10 + (lo - 0xDC00)
			}
			p.i = save // read the next escape on its own
		}
		return utf8.RuneError
	}
	p.fail(at, "invalid escape sequence")
	return 0
}

// hex4 reads the four hex digits of a \u escape that starts at at.
func (p *jsonParser) hex4(at int) rune {
	if p.i+4 > len(p.s) {
		p.fail(len(p.s), "unexpected end of input")
	}
	n, err := strconv.ParseUint(p.s[p.i:p.i+4], 16, 32)
	if err != nil || strings.ContainsAny(p.s[p.i:p.i+4], "+-_") {
		p.fail(at, "invalid escape sequence")
	}
	p.i += 4
	return rune(n)
}

// number reads a number and returns its text.
func (p *jsonParser) number() string {
	start := p.i
	digits := func() int {
		n := 0
		for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
			p.i++
			n++
		}
		return n
	}
	if p.s[p.i] == '-' {
		p.i++
	}
	switch {
	case p.i < len(p.s) && p.s[p.i] == '0':
		p.i++
	case digits() == 0:
		p.fail(start, "invalid number")
	}
	if p.i < len(p.s) && p.s[p.i] == '.' {
		p.i++
		if digits() == 0 {
			p.fail(start, "invalid number")
		}
	}
	if p.i < len(p.s) && (p.s[p.i] == 'e' || p.s[p.i] == 'E') {
		p.i++
		if p.i < len(p.s) && (p.s[p.i] == '+' || p.s[p.i] == '-') {
			p.i++
		}
		if digits() == 0 {
			p.fail(start, "invalid number")
		}
	}
	return p.s[start:p.i]
}

func jsonParse(s string) *jsonValue {
	p := &jsonParser{s: s}
	v := p.value()
	p.space()
	if p.i < len(p.s) {
		p.unexpected()
	}
	return v
}

// ---- writing ----

// jsonQuote writes s as a JSON string literal, escaping as JavaScript's
// JSON.stringify does.
func jsonQuote(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}

// jsonWrite writes v. With an indent, each element and member goes on a
// line of its own, indented once more than prefix, as JSON.stringify(v,
// null, indent) does.
func jsonWrite(b *strings.Builder, v *jsonValue, indent, prefix string) {
	switch v.kind {
	case "null":
		b.WriteString("null")
	case "bool":
		b.WriteString(strconv.FormatBool(v.b))
	case "number":
		b.WriteString(v.text)
	case "string":
		jsonQuote(b, v.text)
	case "array", "object":
		open, close := "[", "]"
		if v.kind == "object" {
			open, close = "{", "}"
		}
		b.WriteString(open)
		inner := prefix + indent
		for i, e := range v.elems {
			if i > 0 {
				b.WriteByte(',')
			}
			if indent != "" {
				b.WriteString("\n" + inner)
			}
			if v.kind == "object" {
				jsonQuote(b, v.keys[i])
				b.WriteByte(':')
				if indent != "" {
					b.WriteByte(' ')
				}
			}
			jsonWrite(b, e, indent, inner)
		}
		if indent != "" && len(v.elems) > 0 {
			b.WriteString("\n" + prefix)
		}
		b.WriteString(close)
	}
}

func jsonString(v *jsonValue, indent string) string {
	var b strings.Builder
	jsonWrite(&b, v, indent, "")
	return b.String()
}

// ---- reading values ----

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func jsonArg(a any) *jsonValue { return a.(*jsonValue) }

// jsonWant checks that v is of kind want ("array or object" accepts both).
func jsonWant(v *jsonValue, want string) {
	if v.kind == want || want == "array or object" && (v.kind == "array" || v.kind == "object") {
		return
	}
	panic(Fault{Type: "TypeError", Message: "expected " + jsonArticle(want) + ", found " + jsonArticle(v.kind)})
}

func jsonArticle(kind string) string {
	switch kind {
	case "null":
		return "null"
	case "array", "array or object", "object":
		return "an " + kind
	}
	return "a " + kind
}

// jsonInt reads an integer number, which must fit in t (int or int64).
func jsonInt(v *jsonValue, t *types.Type) int64 {
	jsonWant(v, "number")
	if strings.ContainsAny(v.text, ".eE") {
		panic(Fault{Type: "TypeError", Message: v.text + " is not an integer"})
	}
	n, _ := new(big.Int).SetString(v.text, 10)
	if !n.IsInt64() || !types.InRange(t, n.Int64()) {
		panic(Fault{Type: "RangeError", Message: fmt.Sprintf("%s does not fit in '%s'", v.text, t)})
	}
	return n.Int64()
}

func init() {
	S, B, I, F, V := types.StringT, types.BoolT, types.Int, types.Float64, JSONValue
	SA, VA := types.NewArray(S), types.NewArray(V)
	p := &Package{Path: "json", Qualifier: "json"}
	p.add(
		fn("parse", V, func(a []any) any { return jsonParse(str(a[0])) }, S),
		fn("stringify", S, func(a []any) any { return jsonString(jsonArg(a[0]), "") }, V),
		fn("stringifyIndent", S, func(a []any) any { return jsonString(jsonArg(a[0]), str(a[1])) }, V, S),

		fn("kind", S, func(a []any) any { return jsonArg(a[0]).kind }, V),
		fn("isNull", B, func(a []any) any { return jsonArg(a[0]).kind == "null" }, V),
		fn("asBool", B, func(a []any) any {
			v := jsonArg(a[0])
			jsonWant(v, "bool")
			return v.b
		}, V),
		fn("asString", S, func(a []any) any {
			v := jsonArg(a[0])
			jsonWant(v, "string")
			return v.text
		}, V),
		fn("asInt", I, func(a []any) any { return jsonInt(jsonArg(a[0]), types.Int) }, V),
		fn("asInt64", types.Int64, func(a []any) any { return jsonInt(jsonArg(a[0]), types.Int64) }, V),
		fn("asFloat", F, func(a []any) any {
			v := jsonArg(a[0])
			jsonWant(v, "number")
			f, _ := strconv.ParseFloat(v.text, 64)
			if math.IsInf(f, 0) {
				panic(Fault{Type: "RangeError", Message: v.text + " is out of range for float64"})
			}
			return f
		}, V),

		fn("length", I, func(a []any) any {
			v := jsonArg(a[0])
			jsonWant(v, "array or object")
			return int64(len(v.elems))
		}, V),
		fn("at", V, func(a []any) any {
			v, i := jsonArg(a[0]), a[1].(int64)
			jsonWant(v, "array")
			if i < 0 || i >= int64(len(v.elems)) {
				rangeFault("index %d is out of range for an array of length %d", i, len(v.elems))
			}
			return v.elems[i]
		}, V, I),
		fn("items", VA, func(a []any) any {
			v := jsonArg(a[0])
			jsonWant(v, "array")
			out := make([]any, len(v.elems))
			for i, e := range v.elems {
				out[i] = e
			}
			return out
		}, V),
		fn("has", B, func(a []any) any {
			v := jsonArg(a[0])
			jsonWant(v, "object")
			_, ok := v.index[str(a[1])]
			return ok
		}, V, S),
		fn("get", V, func(a []any) any {
			v, k := jsonArg(a[0]), str(a[1])
			jsonWant(v, "object")
			at, ok := v.index[k]
			if !ok {
				var b strings.Builder
				jsonQuote(&b, k)
				panic(Fault{Type: "Error", Message: "no key " + b.String()})
			}
			return v.elems[at]
		}, V, S),
		fn("keys", SA, func(a []any) any {
			v := jsonArg(a[0])
			jsonWant(v, "object")
			return append([]string{}, v.keys...)
		}, V),

		fn("nullValue", V, func(a []any) any { return jsonNull }),
		fn("fromBool", V, func(a []any) any {
			if a[0].(bool) {
				return jsonTrue
			}
			return jsonFalse
		}, B),
		fn("fromString", V, func(a []any) any { return &jsonValue{kind: "string", text: str(a[0])} }, S),
		fn("fromInt", V, func(a []any) any { return &jsonValue{kind: "number", text: strconv.FormatInt(a[0].(int64), 10)} }, I),
		fn("fromInt64", V, func(a []any) any { return &jsonValue{kind: "number", text: strconv.FormatInt(a[0].(int64), 10)} }, types.Int64),
		fn("fromFloat", V, func(a []any) any {
			f := a[0].(float64)
			if math.IsNaN(f) || math.IsInf(f, 0) {
				panic(Fault{Type: "ArgumentError", Message: conv.FormatFloat(f, 64) + " cannot be written in JSON"})
			}
			return &jsonValue{kind: "number", text: conv.FormatFloat(f, 64)}
		}, F),
		fn("fromArray", V, func(a []any) any {
			in := a[0].([]any)
			v := &jsonValue{kind: "array", elems: make([]*jsonValue, len(in))}
			for i, e := range in {
				v.elems[i] = jsonArg(e)
			}
			return v
		}, VA),
		fn("fromObject", V, func(a []any) any {
			keys, in := a[0].([]string), a[1].([]any)
			if len(keys) != len(in) {
				panic(Fault{Type: "ArgumentError", Message: fmt.Sprintf("%d %s but %d %s", len(keys), plural(len(keys), "key", "keys"), len(in), plural(len(in), "value", "values"))})
			}
			vals := make([]*jsonValue, len(in))
			for i, e := range in {
				vals[i] = jsonArg(e)
			}
			return jsonObject(keys, vals)
		}, SA, VA),
	)
	p.Types = map[string]*types.Type{"Value": JSONValue}
	register(p)
}
