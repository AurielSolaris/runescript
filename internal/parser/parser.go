// Package parser builds an AST from RuneScript tokens.
//
// The parser stops at the first syntax error; error recovery is future work.
package parser

import (
	"strconv"

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
	toks, lexDiags := lexer.Lex(src)
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
	for !p.is(lexer.EOF, "") {
		f.Funcs = append(f.Funcs, p.funcDecl())
	}
	return f
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

func (p *parser) funcDecl() *ast.FuncDecl {
	start := p.tok().Span
	fn := &ast.FuncDecl{}
	if v, ok := visibilities[p.tok().Text]; ok && p.tok().Kind == lexer.Keyword {
		fn.Visibility = v
		p.next()
		if _, again := visibilities[p.tok().Text]; again && p.tok().Kind == lexer.Keyword {
			p.fail(diagnostics.ParseUnexpected, p.tok().Span, "multiple visibility modifiers")
		}
	}
	if p.is(lexer.Ident, "function") {
		p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
			"write the return type before the name, e.g. `int add(int a, int b) { ... }`",
			"RuneScript does not use the 'function' keyword")
	}
	fn.Result = p.typeRef("return type")
	fn.Name = p.ident("function name")
	p.expect(lexer.LParen, "'(' after function name")
	if !p.is(lexer.RParen, "") {
		for {
			prm := &ast.Param{Type: p.typeRef("parameter type")}
			prm.Name = p.ident("parameter name")
			fn.Params = append(fn.Params, prm)
			if !p.is(lexer.Comma, "") {
				break
			}
			p.next()
		}
	}
	p.expect(lexer.RParen, "')' after parameters")
	fn.Body = p.block()
	fn.Span = span(start, fn.Body.Span)
	return fn
}

var typeKeywords = map[string]bool{
	"int": true, "sint": true, "uint": true, "string": true, "bool": true, "void": true,
}

func (p *parser) isTypeStart() bool {
	t := p.tok()
	return (t.Kind == lexer.Keyword && typeKeywords[t.Text]) ||
		(t.Kind == lexer.Ident && p.peekTok(1).Kind == lexer.Ident)
}

// typeRef accepts a built-in type keyword or an identifier; unknown names
// are reported by the checker.
func (p *parser) typeRef(what string) *ast.TypeRef {
	t := p.tok()
	if (t.Kind == lexer.Keyword && typeKeywords[t.Text]) || t.Kind == lexer.Ident {
		p.next()
		return &ast.TypeRef{Name: t.Text, Span: t.Span}
	}
	p.unexpected(what)
	return nil
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
		return p.forStmt()
	case t.Kind == lexer.Keyword && (t.Text == "break" || t.Text == "continue"):
		p.next()
		end := p.expect(lexer.Semicolon, "';' after '"+t.Text+"'")
		return &ast.BranchStmt{Keyword: t.Text, Span: span(t.Span, end.Span)}
	case t.Kind == lexer.LBrace:
		p.failHelp(diagnostics.ParseNotImplemented, t.Span,
			"standalone blocks are not part of v0.1; blocks follow functions, if, while, and for",
			"unexpected '{'")
	case t.Kind == lexer.Reserved:
		p.notImplemented(t)
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
	}
}

var assignOps = map[string]bool{"+=": true, "-=": true, "*=": true, "/=": true, "%=": true}

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
		if !p.is(lexer.Assign, "") {
			p.failHelp(diagnostics.ParseUnexpected, p.tok().Span,
				"variables must be initialized when declared in v0.1, e.g. `int x = 0;`",
				"expected '=' after variable name, found %s", p.tok().Describe())
		}
		p.next()
		init := p.expr()
		return &ast.VarDecl{Type: typ, Name: name, Init: init, Span: span(typ.Span, init.NodeSpan())}
	}
	if t.Kind != lexer.Ident {
		p.unexpected("statement")
	}
	next := p.peekTok(1)
	switch {
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
		e := p.expr()
		if call, ok := e.(*ast.CallExpr); ok {
			return &ast.ExprStmt{Call: call}
		}
		p.fail(diagnostics.ParseUnexpected, e.NodeSpan(), "only function calls can be used as statements")
	}
	p.fail(diagnostics.ParseUnexpected, t.Span, "expected statement, found %s", t.Describe())
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

// ---- expressions ----

// precedence of binary operators; higher binds tighter.
var precedence = map[string]int{
	"||": 1,
	"&&": 2,
	"==": 3, "!=": 3,
	"<": 4, "<=": 4, ">": 4, ">=": 4,
	"+": 5, "-": 5,
	"*": 6, "/": 6, "%": 6,
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
	if t.Kind == lexer.Op && (t.Text == "-" || t.Text == "!") {
		p.next()
		x := p.unary()
		return &ast.UnaryExpr{Op: t.Text, X: x, Span: span(t.Span, x.NodeSpan())}
	}
	return p.primary()
}

func (p *parser) primary() ast.Expr {
	t := p.tok()
	switch t.Kind {
	case lexer.Int:
		p.next()
		v, _ := strconv.ParseUint(t.Text, 10, 64) // range already checked by the lexer
		return &ast.IntLit{Value: v, Text: t.Text, Span: t.Span}
	case lexer.String:
		p.next()
		return &ast.StringLit{Value: t.Text, Span: t.Span}
	case lexer.Keyword:
		if t.Text == "true" || t.Text == "false" {
			p.next()
			return &ast.BoolLit{Value: t.Text == "true", Span: t.Span}
		}
	case lexer.LParen:
		p.next()
		x := p.expr()
		end := p.expect(lexer.RParen, "')'")
		return &ast.ParenExpr{X: x, Span: span(t.Span, end.Span)}
	case lexer.Ident:
		id := p.ident("name")
		if !p.is(lexer.LParen, "") {
			return id
		}
		p.next()
		call := &ast.CallExpr{Fn: id}
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
		call.Span = span(id.Span, end.Span)
		return call
	}
	p.unexpected("expression")
	return nil
}
