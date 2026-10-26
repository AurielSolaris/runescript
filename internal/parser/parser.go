// Package parser builds an AST from RuneScript tokens.
//
// The parser stops at the first syntax error; error recovery is future work.
package parser

import (
	"unicode/utf8"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/lexer"
)

type parser struct {
	toks []lexer.Token
	pos  int
}

// bailout carries the first syntax error out of the recursive descent.
type bailout struct{ d *diagnostics.Diagnostic }

// Parse lexes and parses src. On error the returned file is nil.
func Parse(src []byte) (file *ast.File, diags []*diagnostics.Diagnostic) {
	return ParseSource(diagnostics.NewSource("", src))
}

// ParseSource lexes and parses one source file of a program.
func ParseSource(src *diagnostics.Source) (file *ast.File, diags []*diagnostics.Diagnostic) {
	toks, lexDiags := lexer.LexSource(src)
	if len(lexDiags) > 0 {
		return nil, lexDiags
	}
	p := &parser{toks: toks}
	defer func() {
		if r := recover(); r != nil {
			b, ok := r.(bailout)
			if !ok {
				panic(r)
			}
			file, diags = nil, []*diagnostics.Diagnostic{b.d}
		}
	}()
	return p.file(), nil
}

// ---- token helpers ----

func (p *parser) tok() lexer.Token { return p.toks[p.pos] }
func (p *parser) peekTok(n int) lexer.Token {
	if p.pos+n < len(p.toks) {
		return p.toks[p.pos+n]
	}
	return p.toks[len(p.toks)-1]
}

func (p *parser) next() lexer.Token {
	t := p.toks[p.pos]
	if t.Kind != lexer.EOF {
		p.pos++
	}
	return t
}

func (p *parser) is(kind lexer.Kind, text string) bool {
	t := p.tok()
	return t.Kind == kind && (text == "" || t.Text == text)
}

func (p *parser) isKeyword(text string) bool { return p.is(lexer.Keyword, text) }

func (p *parser) fail(code diagnostics.Code, span diagnostics.Span, format string, args ...any) {
	panic(bailout{diagnostics.New(code, span, format, args...)})
}

func (p *parser) failHelp(code diagnostics.Code, span diagnostics.Span, help, format string, args ...any) {
	panic(bailout{diagnostics.New(code, span, format, args...).WithHelp(help)})
}

func (p *parser) expect(kind lexer.Kind, what string) lexer.Token {
	t := p.tok()
	if t.Kind != kind {
		p.unexpected(what)
	}
	return p.next()
}

func (p *parser) expectKeyword(text string) lexer.Token {
	if !p.isKeyword(text) {
		p.unexpected("'" + text + "'")
	}
	return p.next()
}

func (p *parser) unexpected(what string) {
	t := p.tok()
	if t.Kind == lexer.Reserved {
		p.notImplemented(t)
	}
	p.fail(diagnostics.ParseUnexpected, t.Span, "expected %s, found %s", what, t.Describe())
}

func (p *parser) notImplemented(t lexer.Token) {
	p.failHelp(diagnostics.ParseNotImplemented, t.Span,
		"this word is reserved for a planned feature; see docs/ROADMAP.md",
		"'%s' is reserved but not implemented yet", t.Text)
}

func span(start, end diagnostics.Span) diagnostics.Span {
	return diagnostics.Span{Start: start.Start, End: end.End}
}

// ---- declarations ----

func (p *parser) file() *ast.File {
	f := &ast.File{}
	if !p.isKeyword("package") {
		t := p.tok()
		p.failHelp(diagnostics.ParseUnexpected, t.Span,
			"every RuneScript file currently starts with `package name;`",
			"expected package declaration, found %s", t.Describe())
	}
	f.Package = p.packageDecl()
	for p.isKeyword("use") {
		f.Uses = append(f.Uses, p.useDecl())
	}
	for !p.is(lexer.EOF, "") {
		if p.isKeyword("use") {
			p.fail(diagnostics.ParseUnexpected, p.tok().Span, "use declarations must come before any function")
		}
		if p.isKeyword("extern") {
			f.Externs = append(f.Externs, p.externDecl())
			continue
		}
		if p.isStructDecl() {
			f.Structs = append(f.Structs, p.structDecl())
			continue
		}
		f.Funcs = append(f.Funcs, p.funcDecl())
	}
	return f
}

// isStructDecl reports whether a struct declaration starts here:
// `visibility struct`, or `struct` without the required modifier.
func (p *parser) isStructDecl() bool {
	if p.isKeyword("struct") {
		return true
	}
	_, vis := visibilities[p.tok().Text]
	t := p.peekTok(1)
	return vis && p.tok().Kind == lexer.Keyword && t.Kind == lexer.Keyword && t.Text == "struct"
}

// structDecl parses `visibility struct Name { Type field; ... }`.
func (p *parser) structDecl() *ast.StructDecl {
	start := p.tok().Span
	d := &ast.StructDecl{}
	if p.isKeyword("struct") {
		p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
			"write `public|private|protected struct Name { ... }`, e.g. `public struct Point { int x; int y; }`",
			"a struct declaration needs an access modifier")
	}
	d.Visibility = visibilities[p.next().Text]
	p.expectKeyword("struct")
	d.Name = p.ident("struct name")
	p.noTypeParams()
	p.expect(lexer.LBrace, "'{' to begin the struct's fields")
	for !p.is(lexer.RBrace, "") {
		if p.is(lexer.EOF, "") {
			p.unexpected("'}' to end the struct")
		}
		fd := &ast.FieldDecl{Type: p.typeRef("field type")}
		fd.Name = p.ident("field name")
		if p.is(lexer.Assign, "") {
			p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
				"every field's value is given by the struct literal, e.g. Point{x: 1, y: 2}",
				"struct fields have no default values")
		}
		p.expect(lexer.Semicolon, "';' after the field")
		d.Fields = append(d.Fields, fd)
	}
	end := p.next()
	d.Span = span(start, end.Span)
	return d
}

// externDecl parses `extern "module" Type name(Type p, ...);`.
func (p *parser) externDecl() *ast.ExternDecl {
	start := p.expectKeyword("extern")
	if !p.is(lexer.String, "") {
		p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
			`name the JavaScript module as a string, e.g. extern "rune:fs" string readFile(string path);`,
			"expected module string after 'extern', found %s", p.tok().Describe())
	}
	mod := p.next()
	d := &ast.ExternDecl{Module: mod.Text, ModuleSpan: mod.Span}
	if _, ok := visibilities[p.tok().Text]; ok && p.tok().Kind == lexer.Keyword {
		p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
			"an extern function is visible only in the file that declares it",
			"extern declarations take no access modifier")
	}
	d.Result = p.typeRef("return type")
	d.Name = p.ident("function name")
	p.noTypeParams()
	p.expect(lexer.LParen, "'(' after function name")
	if !p.is(lexer.RParen, "") {
		for {
			prm := &ast.Param{Type: p.typeRef("parameter type")}
			if p.is(lexer.Ident, "") {
				prm.Name = p.ident("parameter name")
			}
			d.Params = append(d.Params, prm)
			if !p.is(lexer.Comma, "") {
				break
			}
			p.next()
		}
	}
	p.expect(lexer.RParen, "')' after parameters")
	if p.is(lexer.LBrace, "") {
		p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
			"the function is implemented in JavaScript by the module; end the declaration with ';'",
			"an extern declaration has no body")
	}
	end := p.expect(lexer.Semicolon, "';' after extern declaration")
	d.Span = span(start.Span, end.Span)
	return d
}

func (p *parser) useDecl() *ast.UseDecl {
	start := p.expectKeyword("use")
	if !p.is(lexer.String, "") {
		p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
			`the package or file is written as a string, e.g. use "std:string";`,
			"expected package path string after 'use', found %s", p.tok().Describe())
	}
	path := p.next()
	u := &ast.UseDecl{Path: path.Text, PathSpan: path.Span}
	// `as` is contextual: it is a keyword only here, so it stays usable as
	// a name everywhere else.
	if p.is(lexer.Ident, "as") {
		p.next()
		u.Alias = p.ident("name after 'as'")
	}
	end := p.expect(lexer.Semicolon, "';' after use declaration")
	u.Span = span(start.Span, end.Span)
	return u
}

func (p *parser) packageDecl() *ast.PackageDecl {
	start := p.expectKeyword("package")
	d := &ast.PackageDecl{}
	d.Name = append(d.Name, p.expect(lexer.Ident, "package name").Text)
	for p.is(lexer.Dot, "") {
		p.next()
		d.Name = append(d.Name, p.expect(lexer.Ident, "package name segment").Text)
	}
	if p.is(lexer.LBrace, "") {
		p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
			"RuneScript packages are declarations terminated by ';', not ActionScript-style blocks",
			"expected ';' after package name, found '{'")
	}
	end := p.expect(lexer.Semicolon, "';' after package name")
	d.Span = span(start.Span, end.Span)
	return d
}

var visibilities = map[string]ast.Visibility{
	"public": ast.VisPublic, "protected": ast.VisProtected, "private": ast.VisPrivate,
}

// funcDecl parses `visibility Type name(params) { body }` or the prototype
// form ending in ';'. The access modifier is required on every function.
func (p *parser) funcDecl() *ast.FuncDecl {
	start := p.tok().Span
	fn := &ast.FuncDecl{}
	if p.is(lexer.Ident, "function") {
		p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
			"write `public|private|protected <return type> <name>(...)`, e.g. `public int add(int a, int b) { ... }`",
			"RuneScript does not use the 'function' keyword")
	}
	if p.tok().Kind == lexer.Reserved {
		p.notImplemented(p.tok())
	}
	v, ok := visibilities[p.tok().Text]
	if !ok || p.tok().Kind != lexer.Keyword {
		p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
			"functions are declared `public|private|protected <return type> <name>(...)`",
			"expected access modifier (public, private, or protected) to begin a function, found %s", p.tok().Describe())
	}
	fn.Visibility = v
	p.next()
	if _, again := visibilities[p.tok().Text]; again && p.tok().Kind == lexer.Keyword {
		p.fail(diagnostics.ParseUnexpected, p.tok().Span, "multiple access modifiers")
	}
	fn.Result = p.typeRef("return type")
	fn.Name = p.ident("function name")
	p.noTypeParams()
	p.expect(lexer.LParen, "'(' after function name")
	if !p.is(lexer.RParen, "") {
		for {
			prm := &ast.Param{Type: p.typeRef("parameter type")}
			if p.is(lexer.Ident, "") {
				prm.Name = p.ident("parameter name")
			}
			fn.Params = append(fn.Params, prm)
			if !p.is(lexer.Comma, "") {
				break
			}
			p.next()
		}
	}
	p.expect(lexer.RParen, "')' after parameters")
	if p.is(lexer.Semicolon, "") {
		end := p.next()
		fn.Span = span(start, end.Span)
		return fn // prototype
	}
	for _, prm := range fn.Params {
		if prm.Name == nil {
			p.failHelp(diagnostics.ParseUnexpected, prm.Type.Span,
				"parameter names may be left out only in prototypes (declarations ending in ';')",
				"parameter of type '%s' needs a name", prm.Type)
		}
	}
	fn.Body = p.block()
	fn.Span = span(start, fn.Body.Span)
	return fn
}

var typeKeywords = map[string]bool{
	"int": true, "sint": true, "uint": true, "string": true, "bool": true, "void": true,
	"float": true, "float32": true, "float64": true, "char": true,
	"int8": true, "int16": true, "int32": true, "int64": true,
	"uint8": true, "uint16": true, "uint32": true, "uint64": true,
}

// errorTypes are the built-in error type names.
var errorTypes = map[string]bool{"Error": true, "RangeError": true, "TypeError": true, "ArgumentError": true}

// castTypes are the types a `(Type) x` conversion can name.
var castTypes = map[string]bool{
	"int": true, "sint": true, "uint": true, "string": true, "bool": true,
	"float": true, "float32": true, "float64": true, "char": true,
	"int8": true, "int16": true, "int32": true, "int64": true,
	"uint8": true, "uint16": true, "uint32": true, "uint64": true,
}

func (p *parser) isTypeStart() bool {
	t := p.tok()
	return (t.Kind == lexer.Keyword && typeKeywords[t.Text] && p.peekTok(1).Kind != lexer.Dot) ||
		(t.Kind == lexer.Ident && p.peekTok(1).Kind == lexer.Ident) ||
		(t.Kind == lexer.Ident && p.isNamedFuncType()) ||
		(t.Kind == lexer.Ident && p.genericEnd(1) > 0 && p.peekTok(p.genericEnd(1)).Kind == lexer.Ident) ||
		// a type of another package: geom.Point p
		(t.Kind == lexer.Ident && p.peekTok(1).Kind == lexer.Dot && p.peekTok(2).Kind == lexer.Ident && p.peekTok(3).Kind == lexer.Ident) ||
		// a function type returning one: json.Value(json.Value) f
		(t.Kind == lexer.Ident && p.peekTok(1).Kind == lexer.Dot && p.peekTok(2).Kind == lexer.Ident && p.isNamedFuncTypeAt(3))
}

// genericEnd returns the index, relative to the current token, of the
// token after a type-argument list `<...>` starting at index i, or 0 if
// there is none. `Signal<int> count` is a declaration; an expression
// statement is always a call, so `a < b > c` cannot be confused with it.
func (p *parser) genericEnd(i int) int {
	if !p.isOp(i, "<") {
		return 0
	}
	depth := 0
	for ; ; i++ {
		t := p.peekTok(i)
		switch {
		case p.isOp(i, "<"):
			depth++
		case p.isOp(i, ">"), p.isOp(i, ">>"):
			depth -= len(t.Text) // `>>` closes two lists: Array<Array<int>>
			if depth == 0 {
				return i + 1
			}
			if depth < 0 {
				return 0
			}
		case t.Kind == lexer.Ident, t.Kind == lexer.Comma, t.Kind == lexer.LParen, t.Kind == lexer.RParen, t.Kind == lexer.Dot,
			t.Kind == lexer.Keyword && typeKeywords[t.Text]:
		default:
			return 0
		}
	}
}

func (p *parser) isOp(i int, op string) bool {
	t := p.peekTok(i)
	return t.Kind == lexer.Op && t.Text == op
}

// isNamedFuncType reports whether the tokens ahead are a function type
// whose result is a named type, followed by a variable name, such as
// `Error(string) make`. A call is never followed by a name, so this does
// not mistake `f(x)` for a type.
func (p *parser) isNamedFuncType() bool { return p.isNamedFuncTypeAt(1) }

// isNamedFuncTypeAt is isNamedFuncType for parameter lists starting at
// token i.
func (p *parser) isNamedFuncTypeAt(start int) bool {
	i := start
	for p.peekTok(i).Kind == lexer.LParen {
		depth := 0
		for {
			switch p.peekTok(i).Kind {
			case lexer.LParen:
				depth++
			case lexer.RParen:
				depth--
			case lexer.EOF, lexer.Semicolon, lexer.LBrace:
				return false
			}
			i++
			if depth == 0 {
				break
			}
		}
	}
	return i > start && p.peekTok(i).Kind == lexer.Ident
}

// isQualifiedStart reports whether the next tokens begin `pkg.name`. A type
// keyword may be a qualifier: `string.toUpper(s)` uses package std:string.
// So may the reserved words in reservedQualifiers.
func (p *parser) isQualifiedStart() bool {
	t := p.tok()
	return (t.Kind == lexer.Ident || (t.Kind == lexer.Keyword && typeKeywords[t.Text]) ||
		(t.Kind == lexer.Reserved && reservedQualifiers[t.Text])) &&
		p.peekTok(1).Kind == lexer.Dot
}

// reservedQualifiers are reserved words that may still qualify a call:
// `matrix.mul(a, b)` uses std:matrix. The owner chose a library over a
// matrix type (OQ-TYPE-03a), but `matrix` stays reserved.
var reservedQualifiers = map[string]bool{"matrix": true}

// typeRef parses a type: a built-in type keyword or an identifier (unknown
// names are reported by the checker), followed by any number of function
// type suffixes: `int(int, int)` is a function taking two ints and
// returning int; `int(int)(string)` takes a string and returns an int(int).
func (p *parser) typeRef(what string) *ast.TypeRef {
	t := p.tok()
	if !((t.Kind == lexer.Keyword && typeKeywords[t.Text]) || t.Kind == lexer.Ident) {
		p.unexpected(what)
	}
	p.next()
	ref := &ast.TypeRef{Name: t.Text, Span: t.Span}
	if t.Kind == lexer.Ident && p.is(lexer.Dot, "") && p.peekTok(1).Kind == lexer.Ident {
		// A type of another package: geom.Point.
		p.next()
		name := p.next()
		ref = &ast.TypeRef{Pkg: &ast.Ident{Name: t.Text, Span: t.Span}, Name: name.Text, Span: span(t.Span, name.Span)}
	}
	if t.Kind == lexer.Ident && p.is(lexer.Op, "<") {
		p.next()
		for {
			ref.Args = append(ref.Args, p.typeRef("type argument"))
			if !p.is(lexer.Comma, "") {
				break
			}
			p.next()
		}
		p.splitShift()
		if !p.is(lexer.Op, ">") {
			p.unexpected("'>' to close the type arguments")
		}
		ref.Span = span(t.Span, p.next().Span)
	}
	for p.is(lexer.LParen, "") {
		p.next()
		fn := &ast.FuncTypeRef{Result: ref}
		if !p.is(lexer.RParen, "") {
			for {
				fn.Params = append(fn.Params, p.typeRef("parameter type"))
				if !p.is(lexer.Comma, "") {
					break
				}
				p.next()
			}
		}
		end := p.expect(lexer.RParen, "')' to close function type")
		ref = &ast.TypeRef{Func: fn, Span: span(t.Span, end.Span)}
	}
	return ref
}

// noTypeParams explains `T first<T>(...)`: functions cannot be generic.
func (p *parser) noTypeParams() {
	if t := p.tok(); t.Kind == lexer.Op && t.Text == "<" {
		p.failHelp(diagnostics.ParseUnexpected, t.Span,
			"RuneScript has no user-defined generics: write the function for each concrete type; Array<T> is built in (guide, chapter 17)",
			"functions cannot have type parameters")
	}
}

// splitShift turns a `>>` token into two `>` tokens, so that the first
// can close a type-argument list: Array<Array<int>>.
func (p *parser) splitShift() {
	t := p.tok()
	if t.Kind != lexer.Op || t.Text != ">>" {
		return
	}
	mid := t.Span.Start
	mid.Offset++
	mid.Col++
	first := lexer.Token{Kind: lexer.Op, Text: ">", Span: diagnostics.Span{Start: t.Span.Start, End: mid}}
	second := lexer.Token{Kind: lexer.Op, Text: ">", Span: diagnostics.Span{Start: mid, End: t.Span.End}}
	p.toks = append(p.toks[:p.pos], append([]lexer.Token{first, second}, p.toks[p.pos+1:]...)...)
}

func (p *parser) ident(what string) *ast.Ident {
	t := p.expect(lexer.Ident, what)
	return &ast.Ident{Name: t.Text, Span: t.Span}
}

// ---- statements ----

func (p *parser) block() *ast.Block {
	lb := p.expect(lexer.LBrace, "'{'")
	b := &ast.Block{}
	for !p.is(lexer.RBrace, "") {
		if p.is(lexer.EOF, "") {
			p.unexpected("'}'")
		}
		b.Stmts = append(b.Stmts, p.stmt())
	}
	rb := p.next()
	b.Span = span(lb.Span, rb.Span)
	return b
}

func (p *parser) stmt() ast.Stmt {
	t := p.tok()
	switch {
	case t.Kind == lexer.Keyword && t.Text == "return":
		p.next()
		r := &ast.ReturnStmt{}
		if !p.is(lexer.Semicolon, "") {
			r.Value = p.expr()
		}
		end := p.expect(lexer.Semicolon, "';' after return")
		r.Span = span(t.Span, end.Span)
		return r
	case t.Kind == lexer.Keyword && t.Text == "if":
		return p.ifStmt()
	case t.Kind == lexer.Keyword && t.Text == "while":
		p.next()
		p.expect(lexer.LParen, "'(' after 'while'")
		cond := p.expr()
		p.expect(lexer.RParen, "')' after condition")
		body := p.block()
		return &ast.WhileStmt{Cond: cond, Body: body, Span: span(t.Span, body.Span)}
	case t.Kind == lexer.Keyword && t.Text == "for":
		if n := p.peekTok(1); n.Kind == lexer.Ident && n.Text == "each" {
			return p.forEachStmt()
		}
		return p.forStmt()
	case t.Kind == lexer.Keyword && t.Text == "throw":
		p.next()
		v := p.expr()
		end := p.expect(lexer.Semicolon, "';' after throw")
		return &ast.ThrowStmt{Value: v, Span: span(t.Span, end.Span)}
	case t.Kind == lexer.Keyword && t.Text == "try":
		return p.tryStmt()
	case t.Kind == lexer.Ident && t.Text == "template" && p.peekTok(1).Kind == lexer.LBrace:
		return p.templateStmt()
	case t.Kind == lexer.Keyword && t.Text == "extern":
		p.failHelp(diagnostics.ParseUnexpected, t.Span,
			"declare extern functions at the top level of the file, outside any function",
			"extern declarations are not allowed inside a function")
	case t.Kind == lexer.Keyword && (t.Text == "break" || t.Text == "continue"):
		p.next()
		end := p.expect(lexer.Semicolon, "';' after '"+t.Text+"'")
		return &ast.BranchStmt{Keyword: t.Text, Span: span(t.Span, end.Span)}
	case t.Kind == lexer.LBrace:
		p.failHelp(diagnostics.ParseNotImplemented, t.Span,
			"standalone blocks are not part of v0.1; blocks follow functions, if, while, and for",
			"unexpected '{'")
	case t.Kind == lexer.Reserved && !p.isQualifiedStart():
		p.notImplemented(t)
	case t.Kind == lexer.Keyword && visibilities[t.Text] != ast.VisNone && p.isStructDecl():
		p.failHelp(diagnostics.ParseUnexpected, t.Span, "declare structs at the top level of the file, outside any function",
			"struct declarations are not allowed inside a function")
	case t.Kind == lexer.Keyword && visibilities[t.Text] != ast.VisNone:
		return p.funcDecl() // nested function or prototype
	}
	s := p.simpleStmt(true)
	end := p.expect(lexer.Semicolon, "';' after statement")
	setEnd(s, end.Span)
	return s
}

// setEnd extends a simple statement's span to include its terminator.
func setEnd(s ast.Stmt, end diagnostics.Span) {
	switch s := s.(type) {
	case *ast.VarDecl:
		s.Span.End = end.End
	case *ast.AssignStmt:
		s.Span.End = end.End
	case *ast.IncDecStmt:
		s.Span.End = end.End
	case *ast.SetStmt:
		s.Span.End = end.End
	}
}

var assignOps = map[string]bool{"+=": true, "-=": true, "*=": true, "/=": true, "%=": true,
	"&=": true, "|=": true, "^=": true, "<<=": true, ">>=": true}

// simpleStmt parses a declaration, assignment, increment, or call without
// its terminating ';'. Declarations are allowed only when allowDecl is set.
func (p *parser) simpleStmt(allowDecl bool) ast.Stmt {
	t := p.tok()
	if p.isTypeStart() {
		if !allowDecl {
			p.fail(diagnostics.ParseUnexpected, t.Span, "a variable declaration is not allowed here")
		}
		typ := p.typeRef("type")
		name := p.ident("variable name")
		if p.is(lexer.LParen, "") {
			p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
				"functions are declared `public|private|protected <return type> <name>(...)`, also inside other functions",
				"function '%s' needs an access modifier", name.Name)
		}
		if p.is(lexer.Semicolon, "") {
			// `int x;` starts at the type's zero value (OQ-VAR-01).
			return &ast.VarDecl{Type: typ, Name: name, Span: span(typ.Span, name.Span)}
		}
		if !p.is(lexer.Assign, "") {
			p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
				"a declaration is `int x = 1;`, or `int x;` for the type's zero value",
				"expected '=' or ';' after variable name, found %s", p.tok().Describe())
		}
		p.next()
		init := p.expr()
		return &ast.VarDecl{Type: typ, Name: name, Init: init, Span: span(typ.Span, init.NodeSpan())}
	}
	if p.isQualifiedStart() || t.Kind == lexer.LParen {
		return p.exprOrSet()
	}
	if t.Kind != lexer.Ident {
		p.unexpected("statement")
	}
	next := p.peekTok(1)
	switch {
	case next.Kind == lexer.LBracket:
		// a[i] = v; a[i].x += v; or a call such as a[i].push(v);
		return p.exprOrSet()
	case next.Kind == lexer.Assign || (next.Kind == lexer.Op && assignOps[next.Text]):
		target := p.ident("variable name")
		op := p.next().Text
		val := p.expr()
		return &ast.AssignStmt{Target: target, Op: op, Value: val, Span: span(target.Span, val.NodeSpan())}
	case next.Kind == lexer.Op && (next.Text == "++" || next.Text == "--"):
		target := p.ident("variable name")
		op := p.next()
		return &ast.IncDecStmt{Target: target, Op: op.Text, Span: span(target.Span, op.Span)}
	case next.Kind == lexer.LParen:
		return p.exprOrSet() // a call, or f().x = v (rejected by the checker)
	}
	p.fail(diagnostics.ParseUnexpected, t.Span, "expected statement, found %s", t.Describe())
	return nil
}

// exprOrSet parses a call statement, or an assignment to an element or a
// field: a[i] = v, p.x += 1, shapes[i].origin.x = 0.
func (p *parser) exprOrSet() ast.Stmt {
	e := p.expr()
	if call, ok := e.(*ast.CallExpr); ok {
		return &ast.ExprStmt{Call: call}
	}
	op := p.tok()
	switch e.(type) {
	case *ast.IndexExpr, *ast.SelectorExpr:
		if op.Kind == lexer.Assign || (op.Kind == lexer.Op && assignOps[op.Text]) {
			p.next()
			val := p.expr()
			return &ast.SetStmt{Target: e, Op: op.Text, Value: val, Span: span(e.NodeSpan(), val.NodeSpan())}
		}
		if op.Kind == lexer.Op && (op.Text == "++" || op.Text == "--") {
			p.failHelp(diagnostics.ParseUnexpected, op.Span, "write x += 1 or x -= 1 instead",
				"%s works on variables only", op.Text)
		}
	}
	p.fail(diagnostics.ParseUnexpected, e.NodeSpan(), "only calls and assignments can be used as statements")
	return nil
}

func (p *parser) ifStmt() *ast.IfStmt {
	start := p.expectKeyword("if")
	p.expect(lexer.LParen, "'(' after 'if'")
	cond := p.expr()
	p.expect(lexer.RParen, "')' after condition")
	s := &ast.IfStmt{Cond: cond, Then: p.block()}
	s.Span = span(start.Span, s.Then.Span)
	if p.isKeyword("else") {
		p.next()
		if p.isKeyword("if") {
			elif := p.ifStmt()
			s.Else, s.Span = elif, span(start.Span, elif.Span)
		} else {
			b := p.block()
			s.Else, s.Span = b, span(start.Span, b.Span)
		}
	}
	return s
}

// tryStmt parses `try { } catch (Type name) { } ... finally { }`.
func (p *parser) tryStmt() *ast.TryStmt {
	start := p.expectKeyword("try")
	s := &ast.TryStmt{Body: p.block()}
	end := s.Body.Span
	for p.isKeyword("catch") {
		c := &ast.CatchClause{Span: p.next().Span}
		p.expect(lexer.LParen, "'(' after 'catch'")
		if p.is(lexer.Ident, "") && p.peekTok(1).Kind == lexer.RParen {
			// catch (e) catches every error. catch (Error) is a mistake.
			if errorTypes[p.tok().Text] {
				p.failHelp(diagnostics.ParseUnexpected, p.peekTok(1).Span,
					"name the caught error, e.g. catch ("+p.tok().Text+" e)",
					"expected a name after '%s' in catch clause", p.tok().Text)
			}
		} else {
			c.Type = p.typeRef("error type")
		}
		c.Name = p.ident("name of the caught error")
		p.expect(lexer.RParen, "')' after catch parameter")
		c.Body = p.block()
		c.Span = span(c.Span, c.Body.Span)
		s.Catches = append(s.Catches, c)
		end = c.Body.Span
	}
	if p.isKeyword("finally") {
		p.next()
		s.Finally = p.block()
		end = s.Finally.Span
	}
	if len(s.Catches) == 0 && s.Finally == nil {
		p.fail(diagnostics.ParseUnexpected, p.tok().Span, "expected 'catch' or 'finally' after try block, found %s", p.tok().Describe())
	}
	s.Span = span(start.Span, end)
	return s
}

func (p *parser) forStmt() *ast.ForStmt {
	start := p.expectKeyword("for")
	p.expect(lexer.LParen, "'(' after 'for'")
	s := &ast.ForStmt{}
	if !p.is(lexer.Semicolon, "") {
		s.Init = p.simpleStmt(true)
	}
	p.expect(lexer.Semicolon, "';' after for-loop initializer")
	if !p.is(lexer.Semicolon, "") {
		s.Cond = p.expr()
	}
	p.expect(lexer.Semicolon, "';' after for-loop condition")
	if !p.is(lexer.RParen, "") {
		s.Post = p.simpleStmt(false)
	}
	p.expect(lexer.RParen, "')' after for-loop clauses")
	s.Body = p.block()
	s.Span = span(start.Span, s.Body.Span)
	return s
}

// forEachStmt parses `for each (T name in x) { }`.
func (p *parser) forEachStmt() *ast.ForEachStmt {
	start := p.expectKeyword("for")
	p.next() // each
	p.expect(lexer.LParen, "'(' after 'for each'")
	s := &ast.ForEachStmt{Type: p.typeRef("the element type of a for-each loop")}
	if t := p.tok(); t.Kind == lexer.Ident && t.Text == "in" && !(p.peekTok(1).Kind == lexer.Ident && p.peekTok(1).Text == "in") {
		p.fail(diagnostics.ParseUnexpected, s.Type.Span, "expected a type and a variable name before 'in', e.g. for each (int x in a)")
	}
	s.Name = p.ident("a variable name after the element type")
	if t := p.tok(); t.Kind != lexer.Ident || t.Text != "in" {
		p.fail(diagnostics.ParseUnexpected, t.Span, "expected 'in' after the for-each variable, found %s", t.Describe())
	}
	p.next()
	s.X = p.expr()
	p.expect(lexer.RParen, "')' after the array of a for-each loop")
	s.Body = p.block()
	s.Span = span(start.Span, s.Body.Span)
	return s
}

// ---- expressions ----

// precedence of binary operators; higher binds tighter.
var precedence = map[string]int{
	"||": 1,
	"&&": 2,
	"|":  3,
	"^":  4,
	"&":  5,
	"==": 6, "!=": 6,
	"<": 7, "<=": 7, ">": 7, ">=": 7,
	"<<": 8, ">>": 8,
	"+": 9, "-": 9,
	"*": 10, "/": 10, "%": 10,
}

func (p *parser) expr() ast.Expr { return p.binary(1) }

func (p *parser) binary(minPrec int) ast.Expr {
	x := p.unary()
	for {
		t := p.tok()
		prec, ok := precedence[t.Text]
		if t.Kind != lexer.Op || !ok || prec < minPrec {
			return x
		}
		p.next()
		y := p.binary(prec + 1)
		x = &ast.BinaryExpr{Op: t.Text, X: x, Y: y, OpPos: t.Span, Span: span(x.NodeSpan(), y.NodeSpan())}
	}
}

func (p *parser) unary() ast.Expr {
	t := p.tok()
	if t.Kind == lexer.Op && (t.Text == "-" || t.Text == "!" || t.Text == "~") {
		p.next()
		x := p.unary()
		return &ast.UnaryExpr{Op: t.Text, X: x, Span: span(t.Span, x.NodeSpan())}
	}
	if p.isCastStart() {
		p.next()
		typ := p.typeRef("type")
		p.expect(lexer.RParen, "')'")
		x := p.unary()
		return &ast.CastExpr{Type: typ, X: x, Span: span(t.Span, x.NodeSpan())}
	}
	return p.postfix(p.primary())
}

// isCastStart reports whether the next tokens are `(Type)`. A type keyword
// is never an expression, so `(int)` can only begin a conversion.
func (p *parser) isCastStart() bool {
	t1, t2 := p.peekTok(1), p.peekTok(2)
	return p.is(lexer.LParen, "") && t1.Kind == lexer.Keyword && castTypes[t1.Text] && t2.Kind == lexer.RParen
}

// postfix parses calls, field selections, and indexing applied to an
// expression: f(1)(2), (g)(x), e.message, parts[0].
func (p *parser) postfix(x ast.Expr) ast.Expr {
	for p.is(lexer.LParen, "") || p.is(lexer.Dot, "") || p.is(lexer.LBracket, "") {
		if p.is(lexer.LBracket, "") {
			p.next()
			i := p.expr()
			end := p.expect(lexer.RBracket, "']' after the index")
			x = &ast.IndexExpr{X: x, Index: i, Span: span(x.NodeSpan(), end.Span)}
			continue
		}
		if p.is(lexer.Dot, "") {
			p.next()
			name := p.ident("field name after '.'")
			x = &ast.SelectorExpr{X: x, Name: name, Span: span(x.NodeSpan(), name.Span)}
			continue
		}
		call := p.callArgs(nil)
		call.Callee = x
		call.Span.Start = x.NodeSpan().Start
		x = call
	}
	return x
}

// isLambdaStart reports whether '(' begins a lambda parameter list rather
// than a parenthesized expression: `()` or `(Type name`.
func (p *parser) isLambdaStart() bool {
	if !p.is(lexer.LParen, "") {
		return false
	}
	t1, t2 := p.peekTok(1), p.peekTok(2)
	switch {
	case t1.Kind == lexer.RParen:
		return true // `()` is never a parenthesized expression
	case t1.Kind == lexer.Keyword && typeKeywords[t1.Text]:
		return t2.Kind != lexer.Dot // `(string.trim(x))` is a parenthesized call
	case t1.Kind == lexer.Ident:
		return t2.Kind == lexer.Ident || p.genericEnd(2) > 0 && p.peekTok(p.genericEnd(2)).Kind == lexer.Ident ||
			t2.Kind == lexer.Dot && p.peekTok(3).Kind == lexer.Ident && p.peekTok(4).Kind == lexer.Ident // (geom.Point p)
	}
	return false
}

// lambda parses `(Type a, ...) => expr` or `(Type a, ...) => { ... }`.
func (p *parser) lambda() *ast.LambdaExpr {
	start := p.expect(lexer.LParen, "'('")
	lam := &ast.LambdaExpr{}
	if !p.is(lexer.RParen, "") {
		for {
			prm := &ast.Param{Type: p.typeRef("parameter type")}
			prm.Name = p.ident("parameter name")
			lam.Params = append(lam.Params, prm)
			if !p.is(lexer.Comma, "") {
				break
			}
			p.next()
		}
	}
	p.expect(lexer.RParen, "')' after lambda parameters")
	if !p.is(lexer.Op, "=>") {
		p.unexpected("'=>' after lambda parameters")
	}
	p.next()
	if p.is(lexer.LBrace, "") {
		lam.Body = p.block()
		lam.Span = span(start.Span, lam.Body.Span)
	} else {
		lam.Expr = p.expr()
		lam.Span = span(start.Span, lam.Expr.NodeSpan())
	}
	return lam
}

func (p *parser) primary() ast.Expr {
	t := p.tok()
	if p.isQualifiedStart() {
		pkg := &ast.Ident{Name: p.next().Text, Span: t.Span}
		p.next() // '.'
		name := p.ident("name after '" + pkg.Name + ".'")
		if p.isStructLitStart() {
			return p.structLit(pkg, name)
		}
		if !p.is(lexer.LParen, "") {
			// A field, such as e.message; the checker rejects a package
			// function used without a call.
			return &ast.SelectorExpr{X: pkg, Name: name, Span: span(pkg.Span, name.Span)}
		}
		call := p.callArgs(name)
		call.Pkg = pkg
		call.Span.Start = pkg.Span.Start
		return call
	}
	switch t.Kind {
	case lexer.Float:
		p.next()
		v, _ := lexer.ParseFloat(t.Text) // range already checked by the lexer
		return &ast.FloatLit{Value: v, Text: t.Text, Span: t.Span}
	case lexer.Int:
		p.next()
		v, _ := lexer.ParseInt(t.Text) // range already checked by the lexer
		return &ast.IntLit{Value: v, Text: t.Text, Span: t.Span}
	case lexer.String:
		p.next()
		return &ast.StringLit{Value: t.Text, Span: t.Span}
	case lexer.StrHead:
		return p.interpString()
	case lexer.Char:
		p.next()
		r, _ := utf8.DecodeRuneInString(t.Text)
		return &ast.CharLit{Value: r, Span: t.Span}
	case lexer.Keyword:
		if t.Text == "true" || t.Text == "false" {
			p.next()
			return &ast.BoolLit{Value: t.Text == "true", Span: t.Span}
		}
		if t.Text == "new" {
			p.next()
			// A built-in type keyword is accepted here so that the checker
			// can explain that only error types can be created.
			nt := p.tok()
			if nt.Kind != lexer.Ident && !(nt.Kind == lexer.Keyword && typeKeywords[nt.Text]) {
				p.unexpected("type name after 'new'")
			}
			p.next()
			typ := &ast.TypeRef{Name: nt.Text, Span: nt.Span}
			call := p.callArgs(nil)
			return &ast.NewExpr{Type: typ, Args: call.Args, Span: diagnostics.Span{Start: t.Span.Start, End: call.Span.End}}
		}
	case lexer.LBracket:
		p.next()
		lit := &ast.ArrayLit{}
		for !p.is(lexer.RBracket, "") {
			lit.Elems = append(lit.Elems, p.expr())
			if !p.is(lexer.Comma, "") {
				break
			}
			p.next()
		}
		end := p.expect(lexer.RBracket, "']' after the array elements")
		lit.Span = span(t.Span, end.Span)
		return lit
	case lexer.LParen:
		if p.isLambdaStart() {
			return p.lambda()
		}
		p.next()
		x := p.expr()
		end := p.expect(lexer.RParen, "')'")
		return &ast.ParenExpr{X: x, Span: span(t.Span, end.Span)}
	case lexer.Ident:
		id := p.ident("name")
		if p.isStructLitStart() {
			return p.structLit(nil, id)
		}
		if !p.is(lexer.LParen, "") {
			return id
		}
		return p.callArgs(id)
	}
	p.unexpected("expression")
	return nil
}

// interpString parses `$"text {x} text"`: a StrHead token, then each
// hole's expression followed by a StrMid or, after the last, a StrTail.
func (p *parser) interpString() *ast.InterpString {
	head := p.next()
	e := &ast.InterpString{Parts: []string{head.Text}}
	for {
		e.Exprs = append(e.Exprs, p.expr())
		t := p.tok()
		if t.Kind != lexer.StrMid && t.Kind != lexer.StrTail {
			p.unexpected("'}' to close the interpolation")
		}
		p.next()
		e.Parts = append(e.Parts, t.Text)
		if t.Kind == lexer.StrTail {
			e.Span = span(head.Span, t.Span)
			return e
		}
	}
}

// isStructLitStart reports whether `{` after a name begins a struct
// literal: `{}` or `{name:`. A name followed by a block never occurs in
// an expression, so this cannot be confused with one.
func (p *parser) isStructLitStart() bool {
	if !p.is(lexer.LBrace, "") {
		return false
	}
	t1, t2 := p.peekTok(1), p.peekTok(2)
	return t1.Kind == lexer.RBrace || (t1.Kind == lexer.Ident && t2.Kind == lexer.Colon)
}

// structLit parses `{field: value, ...}` after the struct's name.
func (p *parser) structLit(pkg, name *ast.Ident) *ast.StructLit {
	lit := &ast.StructLit{Pkg: pkg, Name: name}
	p.expect(lexer.LBrace, "'{'")
	for !p.is(lexer.RBrace, "") {
		fi := &ast.FieldInit{Name: p.ident("field name")}
		p.expect(lexer.Colon, "':' after the field name")
		fi.Value = p.expr()
		lit.Fields = append(lit.Fields, fi)
		if !p.is(lexer.Comma, "") {
			break
		}
		p.next()
	}
	end := p.expect(lexer.RBrace, "'}' to end the struct literal")
	start := name.Span
	if pkg != nil {
		start = pkg.Span
	}
	lit.Span = span(start, end.Span)
	return lit
}

// callArgs parses `(args)` after a callee. fn is nil when the caller sets
// Callee itself.
func (p *parser) callArgs(fn *ast.Ident) *ast.CallExpr {
	p.expect(lexer.LParen, "'('")
	call := &ast.CallExpr{Fn: fn}
	if !p.is(lexer.RParen, "") {
		for {
			call.Args = append(call.Args, p.expr())
			if !p.is(lexer.Comma, "") {
				break
			}
			p.next()
		}
	}
	end := p.expect(lexer.RParen, "')' after arguments")
	call.Span.End = end.Span.End
	if fn != nil {
		call.Span.Start = fn.Span.Start
	}
	return call
}
