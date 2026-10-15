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
	TypeProtoMismatch      Code = "TYPE-015"
	TypeNotCallable        Code = "TYPE-016"

	FlowBranchOutsideLoop Code = "FLOW-001"
	NameAssignToParam     Code = "NAME-006"
	NamePackageNotUsed    Code = "NAME-007"
	NameNotInPackage      Code = "NAME-008"
	NameAcrossBoundary    Code = "NAME-009"
	NameNeverDefined      Code = "NAME-010"

	UseUnknownPackage  Code = "USE-001"
	UseDuplicate       Code = "USE-002"
	UseUnsupported     Code = "USE-003"
	UseNotFound        Code = "USE-004"
	UseCycle           Code = "USE-005"
	UseBadName         Code = "USE-006"
	UsePackageMismatch Code = "USE-007"
	UseQualifierClash  Code = "USE-008"

	// NameNotVisible reports a use of a private or protected function from
	// outside the file or package it belongs to.
	NameNotVisible Code = "NAME-011"

	EntryMissing      Code = "ENTRY-001"
	EntryBadSignature Code = "ENTRY-002"
	EntryDuplicate    Code = "ENTRY-003"

	RunCallDepth    Code = "RUN-001"
	RunOverflow     Code = "RUN-002"
	RunDivideByZero Code = "RUN-003"
	RunNativeFailed Code = "RUN-004"

	// TargetUnsupported reports a valid program that a code-generation
	// target cannot represent.
	TargetUnsupported Code = "TARGET-001"
)

// Source is one source file of a program.
type Source struct {
	// Name is the file name shown in diagnostics, e.g. "lib/geom.rsm".
	Name string
	Text []byte
	// Index is the file's position in the program's load order; it orders
	// diagnostics from different files.
	Index int
}

// NewSource returns a source with the given display name.
func NewSource(name string, text []byte) *Source {
	return &Source{Name: name, Text: text}
}

// Pos is a position in a source file. Line and Col are 1-based; Col counts
// Unicode code points. Offset is a 0-based byte offset. File is the source
// the position is in; it is nil only for positions built by hand in tests.
type Pos struct {
	File   *Source
	Offset int
	Line   int
	Col    int
}

func (p Pos) String() string { return fmt.Sprintf("%d:%d", p.Line, p.Col) }

// Where renders p as seen from file from: "line:col" in the same file, and
// "name:line:col" in another one.
func (p Pos) Where(from *Source) string {
	if p.File == nil || p.File == from {
		return p.String()
	}
	return p.File.Name + ":" + p.String()
}

// Before reports whether p comes before q in load order, then source order.
func (p Pos) Before(q Pos) bool {
	pi, qi := -1, -1
	if p.File != nil {
		pi = p.File.Index
	}
	if q.File != nil {
		qi = q.File.Index
	}
	if pi != qi {
		return pi < qi
	}
	return p.Offset < q.Offset
}

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
//	file.rsc:3:10: error ASX-TYPE-004: message
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

// Render formats d against the source file its span is in.
func Render(d *Diagnostic) string {
	if f := d.Span.Start.File; f != nil {
		return Format(d, f.Name, f.Text)
	}
	return Format(d, "", nil)
}

// sourceLine returns the full line of src containing byte offset off,
// without its line terminator.
func sourceLine(src []byte, off int) (string, bool) {
	if src == nil || off < 0 || off > len(src) {
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
