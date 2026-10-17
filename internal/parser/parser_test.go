package parser

import (
	"fmt"
	"testing"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

func mustParse(t *testing.T, src string) *ast.File {
	t.Helper()
	f, diags := Parse([]byte(src))
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	return f
}

func TestParseCanonicalEntryPoint(t *testing.T) {
	f := mustParse(t, "package main;\n\nprotected int main() {\n    // do stuff\n    return 0;\n}\n")
	if got := f.Package.Name; len(got) != 1 || got[0] != "main" {
		t.Errorf("package = %v", got)
	}
	if len(f.Funcs) != 1 {
		t.Fatalf("got %d funcs", len(f.Funcs))
	}
	fn := f.Funcs[0]
	if fn.Visibility != ast.VisProtected || fn.Result.Name != "int" || fn.Name.Name != "main" || len(fn.Params) != 0 {
		t.Errorf("unexpected function header: %+v", fn)
	}
	if ret, ok := fn.Body.Stmts[0].(*ast.ReturnStmt); !ok || ret.Value.(*ast.IntLit).Value != 0 {
		t.Errorf("unexpected body: %#v", fn.Body.Stmts)
	}
}

func TestParseDottedPackageAndParams(t *testing.T) {
	f := mustParse(t, "package org.apache.royale;\nprivate void sayHi(string message, int times) { trace(message, times); }")
	if len(f.Package.Name) != 3 || f.Package.Name[2] != "royale" {
		t.Errorf("package = %v", f.Package.Name)
	}
	fn := f.Funcs[0]
	if fn.Visibility != ast.VisPrivate || len(fn.Params) != 2 || fn.Params[1].Type.Name != "int" || fn.Params[1].Name.Name != "times" {
		t.Errorf("unexpected header %+v", fn)
	}
	call := fn.Body.Stmts[0].(*ast.ExprStmt).Call
	if call.Fn.Name != "trace" || len(call.Args) != 2 {
		t.Errorf("unexpected call %+v", call)
	}
}

func TestParseStatements(t *testing.T) {
	f := mustParse(t, `package p;
public void f() {
    int x = 1;
    x = 2;
    x += 3;
    x++;
    if (x > 1) { } else if (x < 0) { } else { }
    while (x < 10) { x++; break; }
    for (int i = 0; i < 3; i++) { continue; }
    for (;;) { break; }
    return;
}`)
	stmts := f.Funcs[0].Body.Stmts
	kinds := []string{"*ast.VarDecl", "*ast.AssignStmt", "*ast.AssignStmt", "*ast.IncDecStmt",
		"*ast.IfStmt", "*ast.WhileStmt", "*ast.ForStmt", "*ast.ForStmt", "*ast.ReturnStmt"}
	if len(stmts) != len(kinds) {
		t.Fatalf("got %d statements", len(stmts))
	}
	for i, k := range kinds {
		if got := fmt.Sprintf("%T", stmts[i]); got != k {
			t.Errorf("stmt %d: got %s, want %s", i, got, k)
		}
	}
	elif := stmts[4].(*ast.IfStmt).Else.(*ast.IfStmt)
	if _, ok := elif.Else.(*ast.Block); !ok {
		t.Error("else-if chain not parsed")
	}
	empty := stmts[7].(*ast.ForStmt)
	if empty.Init != nil || empty.Cond != nil || empty.Post != nil {
		t.Error("empty for clauses should be nil")
	}
}

func TestPrecedence(t *testing.T) {
	f := mustParse(t, "package p;\npublic bool f() { return 1 + 2 * 3 == 7 && !false || 1 < 2; }")
	or := f.Funcs[0].Body.Stmts[0].(*ast.ReturnStmt).Value.(*ast.BinaryExpr)
	if or.Op != "||" {
		t.Fatalf("top op = %s", or.Op)
	}
	and := or.X.(*ast.BinaryExpr)
	eq := and.X.(*ast.BinaryExpr)
	plus := eq.X.(*ast.BinaryExpr)
	mul := plus.Y.(*ast.BinaryExpr)
	if and.Op != "&&" || eq.Op != "==" || plus.Op != "+" || mul.Op != "*" {
		t.Errorf("wrong tree: %s %s %s %s", and.Op, eq.Op, plus.Op, mul.Op)
	}
	if _, ok := and.Y.(*ast.UnaryExpr); !ok {
		t.Error("expected unary ! on right of &&")
	}
}

func TestLeftAssociativity(t *testing.T) {
	f := mustParse(t, "package p;\npublic int f() { return 10 - 4 - 3; }")
	e := f.Funcs[0].Body.Stmts[0].(*ast.ReturnStmt).Value.(*ast.BinaryExpr)
	if _, ok := e.X.(*ast.BinaryExpr); !ok {
		t.Error("subtraction should be left-associative")
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		code diagnostics.Code
		pos  string
	}{
		{"missing package", "protected int main() { return 0; }", diagnostics.ParseUnexpected, "1:1"},
		{"package block", "package main { }", diagnostics.ParseUnexpected, "1:14"},
		{"missing semicolon", "package p;\nprotected int main() {\n    return 0\n}", diagnostics.ParseUnexpected, "4:1"},
		{"function keyword", "package p;\nfunction main() {}", diagnostics.ParseUnexpected, "2:1"},
		{"two modifiers", "package p;\npublic private void f() {}", diagnostics.ParseUnexpected, "2:8"},
		{"reserved class", "package p;\nclass Foo {}", diagnostics.ParseNotImplemented, "2:1"},
		{"use after function", "package p;\npublic void f() { }\nuse \"std:string\";", diagnostics.ParseUnexpected, "3:1"},
		{"use without string", "package p;\nuse std;", diagnostics.ParseUnexpected, "2:5"},
		{"catch without name", "package p;\npublic void f() {\n    try { } catch (Error) { }\n}", diagnostics.ParseUnexpected, "3:25"},
		{"try alone", "package p;\npublic void f() {\n    try { }\n}", diagnostics.ParseUnexpected, "4:1"},
		{"struct without modifier", "package p;\nstruct Point {\n    int x;\n}", diagnostics.ParseUnexpected, "2:1"},
		{"struct field default", "package p;\npublic struct Point {\n    int x = 0;\n}", diagnostics.ParseUnexpected, "3:11"},
		{"struct in function", "package p;\npublic void f() {\n    public struct P {\n        int x;\n    }\n}", diagnostics.ParseUnexpected, "3:5"},
		{"unclosed struct literal", "package p;\npublic void f() {\n    trace(P{x: 1);\n}", diagnostics.ParseUnexpected, "3:17"},
		{"generic function", "package p;\npublic T first<T>(int n) {\n    return n;\n}", diagnostics.ParseUnexpected, "2:15"},
		{"unclosed index", "package p;\npublic void f() {\n    trace(a[1);\n}", diagnostics.ParseUnexpected, "3:14"},
		{"unclosed array", "package p;\npublic void f() {\n    trace([1, 2);\n}", diagnostics.ParseUnexpected, "3:16"},
		{"index increment", "package p;\npublic void f() {\n    a[0]++;\n}", diagnostics.ParseUnexpected, "3:9"},
		{"empty hole", "package p;\npublic void f() {\n    trace($\"a{}\");\n}", diagnostics.ParseUnexpected, "3:15"},
		{"two values in a hole", "package p;\npublic void f() {\n    trace($\"{a b}\");\n}", diagnostics.ParseUnexpected, "3:16"},
		{"cast of nothing", "package p;\npublic void f() {\n    trace((int));\n}", diagnostics.ParseUnexpected, "3:16"},
		{"reserved type in body", "package p;\npublic void f() {\n    array x = 1;\n}", diagnostics.ParseNotImplemented, "3:5"},
		{"declaration missing = or ;", "package p;\npublic void f() {\n    int x 5;\n}", diagnostics.ParseUnexpected, "3:11"},
		{"expression statement", "package p;\npublic void f() {\n    1 + 2;\n}", diagnostics.ParseUnexpected, "3:5"},
		{"standalone block", "package p;\npublic void f() {\n    { }\n}", diagnostics.ParseNotImplemented, "3:5"},
		{"if without braces", "package p;\npublic void f() {\n    if (true) trace(1);\n}", diagnostics.ParseUnexpected, "3:15"},
		{"decl in for post", "package p;\npublic void f() {\n    for (;; int i = 0) { }\n}", diagnostics.ParseUnexpected, "3:13"},
		{"unclosed body", "package p;\npublic void f() {\n", diagnostics.ParseUnexpected, "3:1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, diags := Parse([]byte(c.src))
			if len(diags) != 1 {
				t.Fatalf("got %d diagnostics, want 1: %v", len(diags), diags)
			}
			if d := diags[0]; d.Code != c.code || d.Span.Start.String() != c.pos {
				t.Errorf("got %s at %s (%s), want %s at %s", d.Code, d.Span.Start, d.Message, c.code, c.pos)
			}
		})
	}
}

func TestParseUseAndQualifiedCalls(t *testing.T) {
	f := mustParse(t, `package p;
use "std:string";
use "lib/x.rsc" as x;
public void f() {
    string s = string.toUpper("a");
    string.trimSpace(s);
    trace(string.index(s, "A") + 1);
}`)
	if len(f.Uses) != 2 || f.Uses[0].Path != "std:string" || f.Uses[1].Path != "lib/x.rsc" {
		t.Fatalf("uses = %+v", f.Uses)
	}
	if f.Uses[0].Alias != nil || f.Uses[1].Alias == nil || f.Uses[1].Alias.Name != "x" || f.Uses[1].Alias.Span.Start.Col != 20 {
		t.Errorf("aliases: %+v, %+v", f.Uses[0].Alias, f.Uses[1].Alias)
	}
	stmts := f.Funcs[0].Body.Stmts
	decl := stmts[0].(*ast.VarDecl)
	call := decl.Init.(*ast.CallExpr)
	if decl.Type.Name != "string" || call.Pkg == nil || call.Pkg.Name != "string" || call.Fn.Name != "toUpper" {
		t.Errorf("unexpected declaration %+v / %+v", decl, call)
	}
	if call.Span.Start.Col != 16 {
		t.Errorf("qualified call should start at the qualifier, got col %d", call.Span.Start.Col)
	}
	if s := stmts[1].(*ast.ExprStmt); s.Call.Pkg.Name != "string" || s.Call.Fn.Name != "trimSpace" {
		t.Errorf("unexpected statement %+v", s.Call)
	}
}

func TestParseFunctionFeatures(t *testing.T) {
	f := mustParse(t, `package p;
public int isEven(int);
private int(int)(string) pick(int(int, int) op, void() cb) {
    protected int square(int n) { return n * n; }
    public int later(int n);
    int(int, int) add = (int a, int b) => a + b;
    void() hello = () => { trace("hi"); };
    trace(add(1, 2)(3), (op)(4, 5));
    return square;
}`)
	proto := f.Funcs[0]
	if proto.Body != nil || proto.Params[0].Name != nil || proto.Visibility != ast.VisPublic {
		t.Errorf("prototype parsed wrong: %+v", proto)
	}
	fn := f.Funcs[1]
	if got := fn.Result.String(); got != "int(int)(string)" {
		t.Errorf("result type = %s", got)
	}
	if got := fn.Params[0].Type.String(); got != "int(int, int)" {
		t.Errorf("param type = %s", got)
	}
	if got := fn.Params[1].Type.String(); got != "void()" {
		t.Errorf("param type = %s", got)
	}
	stmts := fn.Body.Stmts
	if nested, ok := stmts[0].(*ast.FuncDecl); !ok || nested.Name.Name != "square" || nested.Visibility != ast.VisProtected || nested.Body == nil {
		t.Errorf("nested function: %#v", stmts[0])
	}
	if nested, ok := stmts[1].(*ast.FuncDecl); !ok || nested.Body != nil {
		t.Errorf("nested prototype: %#v", stmts[1])
	}
	lam := stmts[2].(*ast.VarDecl).Init.(*ast.LambdaExpr)
	if len(lam.Params) != 2 || lam.Expr == nil || lam.Body != nil {
		t.Errorf("expression lambda: %+v", lam)
	}
	block := stmts[3].(*ast.VarDecl).Init.(*ast.LambdaExpr)
	if len(block.Params) != 0 || block.Body == nil {
		t.Errorf("block lambda: %+v", block)
	}
	args := stmts[4].(*ast.ExprStmt).Call.Args
	chained := args[0].(*ast.CallExpr)
	if chained.Callee == nil || chained.Callee.(*ast.CallExpr).Fn.Name != "add" {
		t.Errorf("chained call: %+v", chained)
	}
	if paren := args[1].(*ast.CallExpr); paren.Callee == nil {
		t.Errorf("call of parenthesized expression: %+v", paren)
	}
}

func TestParenthesizedQualifiedCallIsNotALambda(t *testing.T) {
	f := mustParse(t, "package p;\nuse \"std:string\";\npublic void f() { trace((string.toUpper(\"a\"))); }")
	arg := f.Funcs[0].Body.Stmts[0].(*ast.ExprStmt).Call.Args[0]
	if _, ok := arg.(*ast.ParenExpr); !ok {
		t.Errorf("got %T", arg)
	}
}

func TestParseFunctionErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		pos  string
	}{
		{"missing modifier", "package p;\nint main() { return 0; }", "2:1"},
		{"missing nested modifier", "package p;\npublic void f() {\n    int g() { return 1; }\n}", "3:10"},
		{"unnamed definition param", "package p;\npublic void f(int) { }", "2:15"},
		{"lambda missing arrow", "package p;\npublic void f() {\n    void() g = () { };\n}", "3:19"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, diags := Parse([]byte(c.src))
			if len(diags) != 1 || diags[0].Code != diagnostics.ParseUnexpected || diags[0].Span.Start.String() != c.pos {
				t.Errorf("got %v, want ASX-PARSE-001 at %s", diags, c.pos)
			}
		})
	}
}

func TestParseCastsErrorsAndTry(t *testing.T) {
	f := mustParse(t, `package p;
public void f() {
    int n = (int) -64.7 + 1;
    try {
        throw new RangeError("bad", 7);
    } catch (RangeError e) {
        trace(e.message, (string) 1.5e3);
    } catch (Error e) {
    } finally {
    }
}`)
	stmts := f.Funcs[0].Body.Stmts
	bin := stmts[0].(*ast.VarDecl).Init.(*ast.BinaryExpr)
	cast, ok := bin.X.(*ast.CastExpr) // a cast binds tighter than +
	if !ok || cast.Type.Name != "int" || cast.Span.Start.Col != 13 {
		t.Fatalf("cast = %+v", bin.X)
	}
	if neg, ok := cast.X.(*ast.UnaryExpr); !ok || neg.X.(*ast.FloatLit).Value != 64.7 {
		t.Errorf("cast operand = %+v", cast.X)
	}
	try := stmts[1].(*ast.TryStmt)
	if len(try.Catches) != 2 || try.Finally == nil || try.Catches[1].Type.Name != "Error" {
		t.Fatalf("try = %+v", try)
	}
	n := try.Body.Stmts[0].(*ast.ThrowStmt).Value.(*ast.NewExpr)
	if n.Type.Name != "RangeError" || len(n.Args) != 2 {
		t.Errorf("new = %+v", n)
	}
	args := try.Catches[0].Body.Stmts[0].(*ast.ExprStmt).Call.Args
	if sel, ok := args[0].(*ast.SelectorExpr); !ok || sel.Name.Name != "message" || sel.X.(*ast.Ident).Name != "e" {
		t.Errorf("selector = %+v", args[0])
	}
	if c, ok := args[1].(*ast.CastExpr); !ok || c.Type.Name != "string" {
		t.Errorf("string cast = %+v", args[1])
	}
}

func TestParseExterns(t *testing.T) {
	f := mustParse(t, "package p;\nextern \"rune:fs\" string readFile(string path);\nprotected int main() { return 0; }\nextern \"./x.mjs\" int f(int, float32 y);\n")
	if len(f.Externs) != 2 || len(f.Funcs) != 1 {
		t.Fatalf("got %d externs, %d funcs", len(f.Externs), len(f.Funcs))
	}
	x := f.Externs[0]
	if x.Module != "rune:fs" || x.Result.Name != "string" || x.Name.Name != "readFile" || len(x.Params) != 1 || x.Params[0].Name.Name != "path" {
		t.Errorf("unexpected extern: %+v", x)
	}
	if y := f.Externs[1]; y.Params[0].Name != nil || y.Params[1].Type.Name != "float32" || y.Span.Start.Line != 4 {
		t.Errorf("unexpected extern: %+v", y)
	}
}

func TestParseExternErrors(t *testing.T) {
	cases := []struct{ name, src, pos string }{
		{"no module", "package p;\nextern int f();", "2:8"},
		{"modifier", "package p;\nextern \"m\" public int f();", "2:12"},
		{"body", "package p;\nextern \"m\" int f() { return 1; }", "2:20"},
		{"no semicolon", "package p;\nextern \"m\" int f()\nprotected int main() { return 0; }", "3:1"},
		{"inside a function", "package p;\nprotected int main() {\n    extern \"m\" int f();\n    return 0;\n}", "3:5"},
		{"before use", "package p;\nextern \"m\" int f();\nuse \"std:string\";", "3:1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, diags := Parse([]byte(c.src))
			if len(diags) != 1 || diags[0].Code != diagnostics.ParseUnexpected || diags[0].Span.Start.String() != c.pos {
				t.Errorf("got %v, want ASX-PARSE-001 at %s", diags, c.pos)
			}
		})
	}
}

// `>>` is a shift operator, but it also closes two type-argument lists.
func TestNestedTypeArgumentsAndShifts(t *testing.T) {
	src := "package p;\npublic void f() {\n    Array<Array<int>> g = [];\n    Array<Array<Array<int>>> h = [];\n    int x = 256 >> 2 << 1;\n    x >>= 1;\n}"
	f, diags := Parse([]byte(src))
	if len(diags) != 0 {
		t.Fatal(diags)
	}
	stmts := f.Funcs[0].Body.Stmts
	if got := stmts[0].(*ast.VarDecl).Type.String(); got != "Array<Array<int>>" {
		t.Errorf("got type %s", got)
	}
	if got := stmts[1].(*ast.VarDecl).Type.String(); got != "Array<Array<Array<int>>>" {
		t.Errorf("got type %s", got)
	}
	shl := stmts[2].(*ast.VarDecl).Init.(*ast.BinaryExpr)
	if shl.Op != "<<" || shl.X.(*ast.BinaryExpr).Op != ">>" {
		t.Errorf("shifts parsed as %s(%s)", shl.Op, shl.X.(*ast.BinaryExpr).Op)
	}
	if op := stmts[3].(*ast.AssignStmt).Op; op != ">>=" {
		t.Errorf("got assignment %s", op)
	}
}
