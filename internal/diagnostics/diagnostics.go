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
	LexUnexpectedChar       Code = "LEX-001"
	LexUnterminatedString   Code = "LEX-002"
	LexInvalidEscape        Code = "LEX-003"
	LexUnterminatedBlock    Code = "LEX-004"
	LexInvalidNumber        Code = "LEX-005"
	LexBadChar              Code = "LEX-006" // a character literal without exactly one character
	LexInvalidUTF8          Code = "LEX-007"
	LexUnterminatedTemplate Code = "LEX-008"

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
	TypeBadConversion      Code = "TYPE-017"
	TypeInexactFloat       Code = "TYPE-018"
	TypeBadNew             Code = "TYPE-019"
	TypeNoField            Code = "TYPE-020"
	TypeNotError           Code = "TYPE-021"
	TypeUnreachableCatch   Code = "TYPE-022"

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
	RunUncaught     Code = "RUN-005" // a thrown error that nothing caught
	RunConversion   Code = "RUN-006" // a (Type) x conversion that failed
	RunExternResult Code = "RUN-007" // a JavaScript function returned a value of the wrong type
	RunExternThrew  Code = "RUN-008" // a JavaScript function threw

	// Extern (JavaScript) function declarations.
	TypeExternType     Code = "TYPE-023" // a type that cannot cross the JavaScript boundary
	TypeExternMismatch Code = "TYPE-024" // a prototype that differs from the rune: function
	UseExternUnknown   Code = "USE-009"  // an unknown rune: module or function
	TargetNeedsJS      Code = "TARGET-002"

	// Templates (rune:dom).
	TmplPlacement      Code = "TMPL-001" // a template that does not end a function returning Element
	TmplUnknownElement Code = "TMPL-002"
	TmplUnknownAttr    Code = "TMPL-003"
	TmplBadValue       Code = "TMPL-004" // a static attribute value the attribute cannot take
	TmplBadInterp      Code = "TMPL-005" // a {{ }} value of the wrong type
	TmplDuplicateAttr  Code = "TMPL-006"
	TmplNeedsDOM       Code = "TMPL-007" // a template without use "rune:dom"
	TypeArgs           Code = "TYPE-025" // wrong type arguments, e.g. Signal<Error>
	RunNoMountTarget   Code = "RUN-009"  // dom.renderApp found no element
	RunDOMValue        Code = "RUN-010"  // a {{ }} value an attribute cannot take
	RunIndex           Code = "RUN-011"  // an array index out of range, or pop of an empty array
	RunShift           Code = "RUN-012"  // a shift count outside 0 to the width minus 1

	// Warnings. A warning does not stop a program from running unless
	// warnings are treated as errors (-Werror).
	WarnUnusedVar     Code = "WARN-001" // a local variable that is never read
	WarnUnusedParam   Code = "WARN-002" // a parameter that is never read
	WarnUnusedUse     Code = "WARN-003" // a use whose qualifier is never referenced
	WarnUnreachable   Code = "WARN-004" // a statement that can never run
	WarnDiscardResult Code = "WARN-005" // a call whose non-void result is ignored
	WarnHostModule    Code = "WARN-006" // an extern from one host's module, e.g. node:fs
	WarnObsolete      Code = "WARN-007" // an obsolete HTML4 or SVG 1.1 element or attribute
	WarnMountSelector Code = "WARN-008" // dom.renderApp with a selector that is not an id
)

// Severity says whether a diagnostic is an error or a warning.
type Severity int

const (
	Error Severity = iota
	Warning
)

func (s Severity) String() string {
	if s == Warning {
		return "warning"
	}
	return "error"
}

// Source is one source file of a program.
type Source struct {
	// Name is the file name shown in diagnostics, e.g. "lib/geom.rsm".
	Name string
	Text []byte
	// Index is the file's position in the program's load order; it orders
	// diagnostics from different files.
	Index int
	// ConvertedFrom names the encoding the file was converted from to
	// UTF-8 when it was read (see internal/srcenc), or is "".
	ConvertedFrom string
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

// Diagnostic is a single error or warning reported against a source span.
type Diagnostic struct {
	Code     Code
	Severity Severity
	Span     Span
	Message  string
	Help     string
	// Promoted marks a warning reported as an error because warnings are
	// treated as errors (-Werror). Its Severity is Error.
	Promoted bool
}

// Error implements error so a diagnostic can be returned directly.
func (d *Diagnostic) Error() string {
	return fmt.Sprintf("%s: %s %s: %s", d.Span.Start, d.Severity, d.Code, d.Message)
}

// New creates an error diagnostic.
func New(code Code, span Span, format string, args ...any) *Diagnostic {
	return &Diagnostic{Code: code, Span: span, Message: fmt.Sprintf(format, args...)}
}

// NewWarning creates a warning diagnostic.
func NewWarning(code Code, span Span, format string, args ...any) *Diagnostic {
	d := New(code, span, format, args...)
	d.Severity = Warning
	return d
}

// ParseCode parses a diagnostic code written with or without the prefix,
// in any case: "ASX-WARN-001", "asx-warn-001", or "WARN-001".
func ParseCode(s string) (Code, bool) {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, Prefix+"-")
	stage, num, ok := strings.Cut(s, "-")
	if !ok || stage == "" || len(num) != 3 || strings.Trim(num, "0123456789") != "" || strings.Trim(stage, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
		return "", false
	}
	return Code(s), true
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
//
// Warnings say "warning" instead of "error"; a warning promoted to an
// error by -Werror ends its first line with "[-Werror]".
func Format(d *Diagnostic, filename string, src []byte) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s:%s: %s %s: %s", filename, d.Span.Start, d.Severity, d.Code, d.Message)
	if d.Promoted {
		b.WriteString(" [-Werror]")
	}
	b.WriteString("\n")
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
