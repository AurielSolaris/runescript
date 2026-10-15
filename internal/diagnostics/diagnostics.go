// Package diagnostics defines source positions and compiler/runtime
// diagnostics shared by every RuneScript compiler stage.
package diagnostics

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Prefix is the provisional diagnostic-code prefix. It is taken from the
// example in CLAUDE.md §4.5 and is NOT final (open question OQ-DIAG-01).
// Change it here only.
const Prefix = "ASX"

// Code identifies a class of diagnostic, without the prefix.
type Code string

// String returns the full code, e.g. "ASX-TYPE-004".
func (c Code) String() string { return Prefix + "-" + string(c) }

// Diagnostic codes. Codes are stable identifiers; messages may change.
const (
	LexUnexpectedChar     Code = "LEX-001"
	LexUnterminatedString Code = "LEX-002"
	LexInvalidEscape      Code = "LEX-003"
	LexUnterminatedBlock  Code = "LEX-004"
	LexInvalidNumber      Code = "LEX-005"
	LexUnsupportedLiteral Code = "LEX-006"

	ParseUnexpected     Code = "PARSE-001"
	ParseNotImplemented Code = "PARSE-002"

	NameUndefined     Code = "NAME-001"
	NameRedeclared    Code = "NAME-002"
	NameNotCallable   Code = "NAME-003"
	NameNotAValue     Code = "NAME-004"
	NameBuiltinReused Code = "NAME-005"

	TypeMismatch           Code = "TYPE-001"
	TypeVoidValue          Code = "TYPE-002"
	TypeArgCount           Code = "TYPE-003"
	TypeSignedToUnsigned   Code = "TYPE-004"
	TypeUnsignedToSigned   Code = "TYPE-005"
	TypeConstOutOfRange    Code = "TYPE-006"
	TypeNegativeToUnsigned Code = "TYPE-007"
	TypeMissingReturn      Code = "TYPE-008"
	TypeBadReturn          Code = "TYPE-009"
	TypeUnknown            Code = "TYPE-010"
	TypeMixedSignedness    Code = "TYPE-011"
	TypeBadOperand         Code = "TYPE-012"
	TypeCondition          Code = "TYPE-013"
	TypeConstDivByZero     Code = "TYPE-014"

	FlowBranchOutsideLoop Code = "FLOW-001"
	NameAssignToParam     Code = "NAME-006"

	EntryMissing      Code = "ENTRY-001"
	EntryBadSignature Code = "ENTRY-002"

	RunCallDepth    Code = "RUN-001"
	RunOverflow     Code = "RUN-002"
	RunDivideByZero Code = "RUN-003"
)

// Pos is a position in a source file. Line and Col are 1-based; Col counts
// Unicode code points. Offset is a 0-based byte offset.
type Pos struct {
	Offset int
	Line   int
	Col    int
}

func (p Pos) String() string { return fmt.Sprintf("%d:%d", p.Line, p.Col) }

// Span is a half-open source range [Start, End).
type Span struct {
	Start Pos
	End   Pos
}

// Diagnostic is a single error reported against a source span.
type Diagnostic struct {
	Code    Code
	Span    Span
	Message string
	Help    string
}

// Error implements error so a diagnostic can be returned directly.
func (d *Diagnostic) Error() string {
	return fmt.Sprintf("%s: error %s: %s", d.Span.Start, d.Code, d.Message)
}

// New creates a diagnostic.
func New(code Code, span Span, format string, args ...any) *Diagnostic {
	return &Diagnostic{Code: code, Span: span, Message: fmt.Sprintf(format, args...)}
}

// WithHelp attaches a help line and returns d.
func (d *Diagnostic) WithHelp(help string) *Diagnostic {
	d.Help = help
	return d
}

// Format renders d with the offending source line and a caret underline:
//
//	file.rs:3:10: error ASX-TYPE-004: message
//	    uint u = s;
//	             ^
//	  help: ...
func Format(d *Diagnostic, filename string, src []byte) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s:%s: error %s: %s\n", filename, d.Span.Start, d.Code, d.Message)
	if line, ok := sourceLine(src, d.Span.Start.Offset); ok {
		width := 1
		if d.Span.End.Line == d.Span.Start.Line && d.Span.End.Offset > d.Span.Start.Offset {
			width = utf8.RuneCount(src[d.Span.Start.Offset:d.Span.End.Offset])
		}
		fmt.Fprintf(&b, "    %s\n", strings.ReplaceAll(line, "\t", " "))
		fmt.Fprintf(&b, "    %s%s\n", strings.Repeat(" ", d.Span.Start.Col-1), strings.Repeat("^", width))
	}
	if d.Help != "" {
		fmt.Fprintf(&b, "  help: %s\n", d.Help)
	}
	return b.String()
}

// sourceLine returns the full line of src containing byte offset off,
// without its line terminator.
func sourceLine(src []byte, off int) (string, bool) {
	if off < 0 || off > len(src) {
		return "", false
	}
	start := off
	for start > 0 && src[start-1] != '\n' {
		start--
	}
	end := off
	for end < len(src) && src[end] != '\n' {
		end++
	}
	return strings.TrimRight(string(src[start:end]), "\r"), true
}
