package stdlib

import (
	"fmt"
	"strings"

	"github.com/AurielSolaris/runescript/internal/types"
)

// std:csv reads and writes comma-separated values (RFC 4180) as
// Array<Array<string>>: one array of fields per record (OQ-CSV-01).
//
// Reading:
//   - Fields are separated by the separator (',' unless given) and records
//     by "\n" or "\r\n". A lone '\r' is part of a field.
//   - A field that starts with '"' is quoted: it ends at the next single
//     '"', and "" inside it stands for one '"'. A quoted field may contain
//     separators and line breaks, which are kept exactly ("\r\n" stays
//     "\r\n").
//   - Empty lines are skipped, so an empty text has no records. A record
//     may have any number of fields (records need not be the same length).
//   - Spaces are kept: " a " is a field of three characters.
//   - A UTF-8 byte order mark at the start is skipped.
//   - A '"' inside an unquoted field, a character other than a separator or
//     a line break after a closing '"', and a quoted field with no closing
//     '"' are errors. Their messages give the line and the column, counted
//     from 1; the column is in bytes, like string offsets.
//
// Writing ends every record with "\n" and quotes a field only when it has
// to: when it contains the separator, '"', '\r', or '\n', or when it is
// the only field of its record and is empty (an empty line would be
// skipped when read back). So parse(format(rows)) == rows, for rows whose
// records each have at least one field.
//
// The JavaScript twin (section csv in internal/jsgen/runtime.js) follows
// this code line by line, over the text's UTF-8 bytes.
func init() {
	S, C := types.StringT, types.CharT
	SAA := types.NewArray(types.NewArray(S))
	p := &Package{Path: "csv", Qualifier: "csv"}
	p.add(
		fn("parse", SAA, func(a []any) any { return csvParse(a[0].(string), ",") }, S),
		fn("parseWith", SAA, func(a []any) any { return csvParse(a[0].(string), csvSep(a[1].(rune))) }, S, C),
		fn("format", S, func(a []any) any { return csvFormat(a[0].([][]string), ",") }, SAA),
		fn("formatWith", S, func(a []any) any { return csvFormat(a[0].([][]string), csvSep(a[1].(rune))) }, SAA, C),
	)
	register(p)
}

// csvSep checks a separator and returns it as UTF-8.
func csvSep(r rune) string {
	switch r {
	case '"':
		panic(Fault{Type: "ArgumentError", Message: `the separator cannot be '"'`})
	case '\r':
		panic(Fault{Type: "ArgumentError", Message: `the separator cannot be '\r'`})
	case '\n':
		panic(Fault{Type: "ArgumentError", Message: `the separator cannot be '\n'`})
	}
	return string(r)
}

// csvParse splits s into records of fields, separated by sep.
func csvParse(s, sep string) [][]string {
	records := [][]string{}
	i, n := 0, len(s)
	if strings.HasPrefix(s, string(rune(0xFEFF))) {
		i = 3
	}
	line, lineStart := 1, i
	fail := func(l, at int, format string, args ...any) {
		panic(Fault{Type: "Error", Message: fmt.Sprintf("line %d, column %d: ", l, at-lineStart+1) + fmt.Sprintf(format, args...)})
	}
	// lineBreak returns the length of the line break at j: 1 for "\n", 2
	// for "\r\n", 0 for none.
	lineBreak := func(j int) int {
		switch {
		case j < n && s[j] == '\n':
			return 1
		case j+1 < n && s[j] == '\r' && s[j+1] == '\n':
			return 2
		}
		return 0
	}
	for i < n {
		if k := lineBreak(i); k > 0 { // an empty line
			i += k
			line++
			lineStart = i
			continue
		}
		record := []string{}
		for {
			var field strings.Builder
			if i < n && s[i] == '"' {
				qline, qstart, qat := line, lineStart, i
				i++
				for {
					if i >= n {
						lineStart = qstart
						fail(qline, qat, `quoted field has no closing '"'`)
					}
					c := s[i]
					if c == '"' {
						if i+1 < n && s[i+1] == '"' {
							field.WriteByte('"')
							i += 2
							continue
						}
						i++
						break
					}
					field.WriteByte(c)
					i++
					if c == '\n' {
						line++
						lineStart = i
					}
				}
				if i < n && !strings.HasPrefix(s[i:], sep) && lineBreak(i) == 0 {
					fail(line, i, `expected a separator or a line break after the closing '"'`)
				}
			} else {
				for i < n && !strings.HasPrefix(s[i:], sep) && lineBreak(i) == 0 {
					if s[i] == '"' {
						fail(line, i, `'"' in an unquoted field; quote the field and write the '"' twice`)
					}
					field.WriteByte(s[i])
					i++
				}
			}
			record = append(record, field.String())
			if i < n && strings.HasPrefix(s[i:], sep) {
				i += len(sep)
				continue
			}
			break
		}
		records = append(records, record)
		if k := lineBreak(i); k > 0 {
			i += k
			line++
			lineStart = i
		}
	}
	return records
}

// csvFormat writes records as CSV text, separated by sep.
func csvFormat(records [][]string, sep string) string {
	var b strings.Builder
	for r, record := range records {
		if len(record) == 0 {
			panic(Fault{Type: "ArgumentError", Message: fmt.Sprintf("rows[%d] has no fields; a CSV record has at least one", r)})
		}
		for i, field := range record {
			if i > 0 {
				b.WriteString(sep)
			}
			if strings.Contains(field, sep) || strings.ContainsAny(field, "\"\r\n") || (len(record) == 1 && field == "") {
				b.WriteByte('"')
				b.WriteString(strings.ReplaceAll(field, `"`, `""`))
				b.WriteByte('"')
			} else {
				b.WriteString(field)
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}
