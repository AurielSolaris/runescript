package ir

import (
	"strings"

	"fmt"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/dom"
	"github.com/AurielSolaris/runescript/internal/types"
)

// Lower converts a program that checked without errors into IR. It
// panics if info is incomplete, which would be a checker bug, not a user
// error.
func Lower(info *checker.Info) *Program {
	p := &Program{EntrySite: info.Entry.Name.Span}
	l := &lowerer{info: info, prog: p, funcs: map[*ast.FuncDecl]*Func{}, syms: map[*checker.FuncSym]*Func{},
		externs: map[*checker.ExternSym]*Extern{}}
	for _, file := range info.Program.Files {
		for _, fd := range file.AST.Funcs {
			if fd.Body != nil {
				f := l.funcDecl(fd)
				f.Pkg = file.Package.Name
				p.Funcs = append(p.Funcs, f)
			}
		}
	}
	p.Entry = l.funcs[info.Entry]
	return p
}

type lowerer struct {
	info    *checker.Info
	prog    *Program
	externs map[*checker.ExternSym]*Extern
	// funcs maps definitions to their IR function, created on first
	// reference so calls may precede definitions.
	funcs map[*ast.FuncDecl]*Func
	syms  map[*checker.FuncSym]*Func
	// scopes is the variable scope chain of the function being lowered.
	// A function boundary starts a fresh chain: functions do not capture,
	// except the runes of the functions around them (OQ-DOM-01), which are
	// found in outer, the chains of the enclosing functions.
	scopes []map[string]*Var
	outer  [][]map[string]*Var
}

func (l *lowerer) ice(format string, args ...any) {
	panic("ir: internal compiler error: " + fmt.Sprintf(format, args...))
}

func (l *lowerer) fn(sym *checker.FuncSym) *Func {
	if f, ok := l.syms[sym]; ok {
		return f
	}
	if sym.Def == nil {
		l.ice("function %q has no definition", sym.Name)
	}
	f := l.declOf(sym.Def)
	if f.Type == nil {
		f.Type = sym.Type // the body may not have been lowered yet
	}
	l.syms[sym] = f
	return f
}

func (l *lowerer) declOf(fd *ast.FuncDecl) *Func {
	if f, ok := l.funcs[fd]; ok {
		return f
	}
	f := &Func{Name: fd.Name.Name, Span: fd.Span}
	l.funcs[fd] = f
	return f
}

// funcDecl lowers a function definition's signature and body.
func (l *lowerer) funcDecl(fd *ast.FuncDecl) *Func {
	f := l.declOf(fd)
	params := make([]*types.Type, len(fd.Params))
	defer l.boundary()()
	for i, p := range fd.Params {
		f.Params = append(f.Params, l.declare(p.Name.Name, l.typeOf(p.Type), true))
		params[i] = f.Params[i].Type
	}
	f.Type = types.NewFunc(params, l.typeOf(fd.Result))
	f.Body = l.block(fd.Body)
	return f
}

func (l *lowerer) lambda(e *ast.LambdaExpr) *Func {
	t := l.concrete(e, l.info.Types[e])
	f := &Func{Type: t, Span: e.Span}
	defer l.boundary()()
	for i, p := range e.Params {
		f.Params = append(f.Params, l.declare(p.Name.Name, t.Params[i], true))
	}
	if e.Body != nil {
		f.Body = l.block(e.Body)
		return f
	}
	// An expression body returns its value, or, when the lambda's result
	// is void, is evaluated for its effect.
	if t.Result.Kind == types.Void {
		f.Body = &Block{Stmts: []Stmt{&Eval{X: l.expr(e.Expr)}}}
	} else {
		f.Body = &Block{Stmts: []Stmt{&Return{Value: l.expr(e.Expr)}}}
	}
	return f
}

// typeOf resolves a written type. The checker has accepted it, so it is
// one of the built-in types or a function type over them.
func (l *lowerer) typeOf(ref *ast.TypeRef) *types.Type {
	if t, ok := l.info.TypeRefs[ref]; ok {
		return l.concreteType(ref, t)
	}
	if ref.Func != nil {
		params := make([]*types.Type, len(ref.Func.Params))
		for i, p := range ref.Func.Params {
			params[i] = l.typeOf(p)
		}
		return types.NewFunc(params, l.typeOf(ref.Func.Result))
	}
	switch ref.Name {
	case "char":
		return types.CharT
	case "Array":
		return types.NewArray(l.typeOf(ref.Args[0]))
	case "Element":
		return types.ElementT
	case "Signal":
		return types.NewSignal(l.typeOf(ref.Args[0]))
	}
	t := types.Lookup(ref.Name)
	if t == nil {
		l.ice("unknown type %q", ref.Name)
	}
	return t
}

func (l *lowerer) boundary() (restore func()) {
	saved := l.scopes
	l.outer = append(l.outer, saved)
	l.scopes = []map[string]*Var{{}}
	return func() {
		l.scopes = saved
		l.outer = l.outer[:len(l.outer)-1]
	}
}

func (l *lowerer) declare(name string, t *types.Type, param bool) *Var {
	v := &Var{Name: name, Type: t, Param: param}
	l.scopes[len(l.scopes)-1][name] = v
	return v
}

func (l *lowerer) lookup(name string) *Var {
	for i := len(l.scopes) - 1; i >= 0; i-- {
		if v, ok := l.scopes[i][name]; ok {
			return v
		}
	}
	for j := len(l.outer) - 1; j >= 0; j-- {
		chain := l.outer[j]
		for i := len(chain) - 1; i >= 0; i-- {
			if v, ok := chain[i][name]; ok && v.Type.Kind == types.Signal {
				return v // a captured rune
			}
		}
	}
	l.ice("unresolved variable %q", name)
	return nil
}

// ---- statements ----

func (l *lowerer) block(b *ast.Block) *Block {
	l.scopes = append(l.scopes, map[string]*Var{})
	defer func() { l.scopes = l.scopes[:len(l.scopes)-1] }()
	out := &Block{}
	for _, s := range b.Stmts {
		if st := l.stmt(s); st != nil {
			out.Stmts = append(out.Stmts, st)
		}
	}
	return out
}

func (l *lowerer) stmt(s ast.Stmt) Stmt {
	switch s := s.(type) {
	case *ast.FuncDecl:
		if s.Body == nil {
			return nil // prototypes only matter to the checker
		}
		return &FuncStmt{Func: l.funcDecl(s)}
	case *ast.VarDecl:
		t := l.typeOf(s.Type)
		var init Expr
		if s.Init != nil {
			init = l.expr(s.Init) // the initializer cannot see the new variable
		} else {
			init = zero(t)
		}
		return &Decl{Var: l.declare(s.Name.Name, t, false), Init: init}
	case *ast.AssignStmt:
		v := l.lookup(s.Target.Name)
		value := l.expr(s.Value)
		if s.Op == "+=" && v.Type.Kind == types.String {
			return &Assign{Var: v, Value: concat(&Load{Var: v}, value)}
		}
		if s.Op != "=" {
			value = binary(strings.TrimSuffix(s.Op, "="), v.Type, &Load{Var: v}, value, s.Span)
		}
		return &Assign{Var: v, Value: value}
	case *ast.SetStmt:
		value := l.expr(s.Value)
		t := l.concrete(s.Target, l.info.Types[s.Target])
		op := strings.TrimSuffix(s.Op, "=")
		return l.store(s.Target, s.Op != "=", func(old Expr) Expr {
			if s.Op == "=" {
				return value
			}
			return binary(op, t, old, value, s.Span)
		}, s.Span)
	case *ast.IncDecStmt:
		v := l.lookup(s.Target.Name)
		one := &IntConst{T: v.Type, Value: 1}
		return &Assign{Var: v, Value: &Arith{Op: s.Op[:1], T: v.Type, X: &Load{Var: v}, Y: one, Site: s.Span}}
	case *ast.ExprStmt:
		return &Eval{X: l.expr(s.Call)}
	case *ast.ReturnStmt:
		if s.Value == nil {
			return &Return{}
		}
		return &Return{Value: l.expr(s.Value)}
	case *ast.IfStmt:
		return l.ifStmt(s)
	case *ast.WhileStmt:
		return &While{Cond: l.expr(s.Cond), Body: l.block(s.Body)}
	case *ast.ForStmt:
		l.scopes = append(l.scopes, map[string]*Var{})
		defer func() { l.scopes = l.scopes[:len(l.scopes)-1] }()
		f := &For{}
		if s.Init != nil {
			f.Init = l.stmt(s.Init)
		}
		if s.Cond != nil {
			f.Cond = l.expr(s.Cond)
		}
		if s.Post != nil {
			f.Post = l.stmt(s.Post)
		}
		f.Body = l.block(s.Body)
		return f
	case *ast.BranchStmt:
		if s.Keyword == "break" {
			return &Break{}
		}
		return &Continue{}
	case *ast.ThrowStmt:
		return &Throw{X: l.expr(s.Value), Site: s.Span}
	case *ast.TemplateStmt:
		return &Return{Value: &Template{Root: l.telement(s.Root)}}
	case *ast.TryStmt:
		t := &Try{Body: l.block(s.Body)}
		for _, cl := range s.Catches {
			l.scopes = append(l.scopes, map[string]*Var{})
			ct := l.info.CatchTypes[cl]
			c := &Catch{T: ct, Var: l.declare(cl.Name.Name, ct, false)}
			c.Body = l.block(cl.Body)
			l.scopes = l.scopes[:len(l.scopes)-1]
			t.Catches = append(t.Catches, c)
		}
		if s.Finally != nil {
			t.Finally = l.block(s.Finally)
		}
		return t
	}
	l.ice("unexpected statement %T", s)
	return nil
}

// store lowers an assignment to the place target, whose new value is
// update(old); useOld reports whether update reads old. A variable is
// assigned, an element is set with IndexSet (reading the old element
// first when it is needed), and a struct field becomes an assignment of
// a changed copy to the place holding the struct: p.a.b = v is
// p = WithField(p, a, WithField(p.a, b, v)).
func (l *lowerer) store(target ast.Expr, useOld bool, update func(old Expr) Expr, site Span) Stmt {
	switch t := unparenAST(target).(type) {
	case *ast.Ident:
		v := l.lookup(t.Name)
		return &Assign{Var: v, Value: update(&Load{Var: v})}
	case *ast.IndexExpr:
		set := &IndexSet{X: l.expr(t.X), I: l.expr(t.Index), Site: site, Compound: useOld}
		set.Value = update(&IndexOld{T: l.concrete(t, l.info.Types[t])})
		return set
	case *ast.SelectorExpr:
		st := l.concrete(t.X, l.info.Types[t.X])
		i := types.FieldIndex(st, t.Name.Name)
		if i < 0 {
			l.ice("struct %s has no field %s", st, t.Name.Name)
		}
		return l.store(t.X, true, func(old Expr) Expr {
			return &WithField{X: old, Index: i, Value: update(&GetField{X: old, Index: i})}
		}, site)
	}
	l.ice("cannot store to %T", target)
	return nil
}

func unparenAST(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}

func (l *lowerer) ifStmt(s *ast.IfStmt) *If {
	out := &If{Cond: l.expr(s.Cond), Then: l.block(s.Then)}
	switch e := s.Else.(type) {
	case *ast.Block:
		out.Else = l.block(e)
	case *ast.IfStmt:
		out.Else = l.ifStmt(e)
	}
	return out
}

// ---- expressions ----

func (l *lowerer) concreteType(ref *ast.TypeRef, t *types.Type) *types.Type {
	if t.Kind == types.Invalid {
		l.ice("type %s at %s was not resolved", ref, ref.Span.Start)
	}
	return t
}

// concrete returns t, which must be a fully resolved type.
func (l *lowerer) concrete(e ast.Expr, t *types.Type) *types.Type {
	if t == nil || t.Kind == types.Invalid || t.Kind == types.UntypedInt || t.Kind == types.UntypedFloat {
		l.ice("expression at %s has no concrete type", e.NodeSpan().Start)
	}
	return t
}

func (l *lowerer) expr(e ast.Expr) Expr {
	if v, ok := l.info.Consts[e]; ok {
		return &IntConst{T: l.concrete(e, l.info.Types[e]), Value: v}
	}
	if v, ok := l.info.FloatConsts[e]; ok {
		return &FloatConst{T: l.concrete(e, l.info.Types[e]), Value: v}
	}
	switch e := e.(type) {
	case *ast.FloatLit:
		l.ice("float literal at %s was not given a type", e.Span.Start)
	case *ast.CastExpr:
		return &Convert{X: l.expr(e.X), To: l.concrete(e, l.info.Types[e]), Site: e.Span}
	case *ast.NewExpr:
		n := &NewError{T: l.concrete(e, l.info.Types[e])}
		if len(e.Args) > 0 {
			n.Message = l.expr(e.Args[0])
		}
		if len(e.Args) > 1 {
			n.ID = l.expr(e.Args[1])
		}
		return n
	case *ast.SelectorExpr:
		if sym, ok := l.info.QualRefs[e]; ok {
			return &FuncRef{Func: l.fn(sym)}
		}
		x := l.expr(e.X)
		if x.Type().Kind == types.Array {
			return &ArrayLen{X: x}
		}
		if x.Type().Kind == types.Struct {
			return &GetField{X: x, Index: types.FieldIndex(x.Type(), e.Name.Name)}
		}
		return &Field{X: x, Name: e.Name.Name, T: l.concrete(e, l.info.Types[e])}
	case *ast.ParenExpr:
		return l.expr(e.X)
	case *ast.StringLit:
		return &StringConst{Value: e.Value, Span: e.Span}
	case *ast.InterpString:
		var parts []Expr
		for i, x := range e.Exprs {
			if e.Parts[i] != "" {
				parts = append(parts, &StringConst{Value: e.Parts[i], Span: e.Span})
			}
			v := l.expr(x)
			if v.Type().Kind != types.String {
				v = &Convert{X: v, To: types.StringT, Site: x.NodeSpan()}
			}
			parts = append(parts, v)
		}
		if last := e.Parts[len(e.Exprs)]; last != "" {
			parts = append(parts, &StringConst{Value: last, Span: e.Span})
		}
		return concat(parts...)
	case *ast.BoolLit:
		return &BoolConst{Value: e.Value}
	case *ast.CharLit:
		return &CharConst{Value: e.Value}
	case *ast.StructLit:
		lit := &StructLit{T: l.concrete(e, l.info.Types[e])}
		for _, fi := range e.Fields {
			lit.Inits = append(lit.Inits, FieldInit{Index: types.FieldIndex(lit.T, fi.Name.Name), Value: l.expr(fi.Value)})
		}
		return lit
	case *ast.ArrayLit:
		lit := &ArrayLit{T: l.concrete(e, l.info.Types[e])}
		for _, x := range e.Elems {
			lit.Elems = append(lit.Elems, l.expr(x))
		}
		return lit
	case *ast.IndexExpr:
		return &Index{X: l.expr(e.X), I: l.expr(e.Index), Site: e.Span}
	case *ast.IntLit:
		l.ice("integer literal at %s was not given a type", e.Span.Start)
	case *ast.Ident:
		if sym, ok := l.info.FuncRefs[e]; ok {
			return &FuncRef{Func: l.fn(sym)}
		}
		return &Load{Var: l.lookup(e.Name)}
	case *ast.LambdaExpr:
		return &Lambda{Func: l.lambda(e)}
	case *ast.CallExpr:
		return l.call(e)
	case *ast.UnaryExpr:
		x := l.expr(e.X)
		if e.Op == "!" {
			return &Not{X: x}
		}
		if e.Op == "~" {
			return &Complement{X: x}
		}
		t := x.Type()
		if t.Kind == types.Float {
			return &Neg{X: x}
		}
		return &Arith{Op: "-", T: t, X: &IntConst{T: t}, Y: x, Site: e.Span}
	case *ast.BinaryExpr:
		x, y := l.expr(e.X), l.expr(e.Y)
		switch e.Op {
		case "&&", "||":
			return &Logic{Op: e.Op, X: x, Y: y}
		case "==", "!=", "<", "<=", ">", ">=":
			return &Compare{Op: e.Op, X: x, Y: y}
		}
		return binary(e.Op, l.concrete(e, l.info.Types[e]), x, y, e.OpPos)
	}
	l.ice("unexpected expression %T", e)
	return nil
}

// binary is x op y for an arithmetic, bitwise, or shift operator, or + on
// strings, with result type t.
func binary(op string, t *types.Type, x, y Expr, site Span) Expr {
	switch op {
	case "&", "|", "^":
		return &Bits{Op: op, T: t, X: x, Y: y}
	case "<<", ">>":
		return &Shift{Op: op, X: x, Count: y, Site: site}
	case "+":
		if t.Kind == types.String {
			return concat(x, y)
		}
	}
	return &Arith{Op: op, T: t, X: x, Y: y, Site: site}
}

// concat joins string expressions, flattening nested joins.
func concat(parts ...Expr) *Concat {
	out := &Concat{}
	for _, p := range parts {
		if c, ok := p.(*Concat); ok {
			out.Parts = append(out.Parts, c.Parts...)
		} else {
			out.Parts = append(out.Parts, p)
		}
	}
	return out
}

func (l *lowerer) call(e *ast.CallExpr) Expr {
	if method, ok := l.info.ArrayCalls[e]; ok {
		var recv Expr
		if e.Pkg != nil {
			recv = &Load{Var: l.lookup(e.Pkg.Name)}
		} else {
			recv = l.expr(e.Callee.(*ast.SelectorExpr).X)
		}
		if method == "push" {
			return &ArrayPush{X: recv, V: l.expr(e.Args[0])}
		}
		return &ArrayPop{X: recv, Site: e.Span}
	}
	var callee Expr
	if e.Callee != nil {
		callee = l.expr(e.Callee) // evaluated before the arguments
	}
	args := make([]Expr, len(e.Args))
	for i, a := range e.Args {
		args[i] = l.expr(a)
	}
	if nf, ok := l.info.Natives[e]; ok {
		return &CallNative{Pkg: e.Pkg.Name, Func: nf, Args: args, Site: e.Span}
	}
	if op, ok := l.info.Signals[e]; ok {
		switch op.Op {
		case "new":
			return &NewSignal{T: op.T, Init: args[0]}
		case "get":
			return &SignalGet{X: l.runeOf(e)}
		default:
			return &SignalSet{X: l.runeOf(e), Value: args[0]}
		}
	}
	if l.info.DOMCalls[e] == "renderApp" {
		return &RenderApp{Component: args[0], Selector: args[1], Site: e.Span}
	}
	if x, ok := l.info.Externs[e]; ok {
		return &CallExtern{Extern: l.extern(x), Args: args, Site: e.Span}
	}
	if callee != nil {
		return &CallValue{Callee: callee, Args: args, Site: e.Span}
	}
	if sym, ok := l.info.Callees[e]; ok {
		return &CallFunc{Func: l.fn(sym), Args: args, Site: e.Span}
	}
	if name, ok := l.info.Builtins[e]; ok {
		return &FloatTest{Name: name, X: args[0]}
	}
	if e.Fn.Name == "trace" {
		return &Trace{Args: args}
	}
	// A variable holding a function value.
	return &CallValue{Callee: &Load{Var: l.lookup(e.Fn.Name)}, Args: args, Site: e.Span}
}

// extern returns the IR form of an extern function, adding it to the
// program on first use.
func (l *lowerer) extern(x *checker.ExternSym) *Extern {
	if e, ok := l.externs[x]; ok {
		return e
	}
	e := &Extern{Name: x.Name, Module: x.Module, Type: x.Type, Host: x.Host}
	l.externs[x] = e
	l.prog.Externs = append(l.prog.Externs, e)
	return e
}

// runeOf is the rune a get or set call acts on: count in count() and in
// count.set(1).
func (l *lowerer) runeOf(e *ast.CallExpr) Expr {
	if e.Pkg != nil {
		return &Load{Var: l.lookup(e.Pkg.Name)}
	}
	return &Load{Var: l.lookup(e.Fn.Name)}
}

// ---- templates ----

// attrMode is the runtime behaviour of an attribute of the given kind:
// the one place where DOM metadata becomes IR.
func attrMode(k dom.Kind) AttrMode {
	switch k {
	case dom.Bool:
		return AttrBool
	case dom.URL:
		return AttrURL
	case dom.Event:
		return AttrEvent
	}
	return AttrText
}

func (l *lowerer) telement(e *ast.MElement) *TElement {
	el := l.info.TemplateElems[e]
	if el == nil {
		l.ice("element <%s> was not resolved", e.Name)
	}
	t := &TElement{Tag: e.Name, SVG: el.Namespace == dom.SVG}
	for _, a := range e.Attrs {
		at := l.info.TemplateAttrs[a]
		if at == nil {
			l.ice("attribute %s of <%s> was not resolved", a.Name, e.Name)
		}
		ta := &TAttr{Name: a.Name, Mode: attrMode(at.Kind), Site: a.Span}
		switch {
		case a.Expr != nil:
			ta.Value = l.expr(a.Expr)
			if at.Kind == dom.Enum {
				ta.Values = at.Values
			}
		case a.Value == nil, at.Kind == dom.Bool:
			ta.Present = true
		default:
			ta.Static = *a.Value
		}
		t.Attrs = append(t.Attrs, ta)
	}
	t.Children = l.tnodes(e.Children)
	return t
}

func (l *lowerer) tnodes(nodes []ast.Markup) []TNode {
	var out []TNode
	for _, n := range nodes {
		switch n := n.(type) {
		case *ast.MElement:
			out = append(out, l.telement(n))
		case *ast.MText:
			out = append(out, &TText{Text: n.Text})
		case *ast.MInterp:
			out = append(out, &TInterp{X: l.expr(n.X)})
		case *ast.MIf:
			t := &TIf{}
			for i, c := range n.Conds {
				t.Conds = append(t.Conds, l.expr(c))
				t.Branches = append(t.Branches, l.tnodes(n.Branches[i]))
			}
			t.Else = l.tnodes(n.Else)
			out = append(out, t)
		}
	}
	return out
}

// zero is the zero value of t, the initial value of `T x;` (OQ-VAR-01).
// The checker allows that only for types.HasZero types. An array's zero
// value is a new empty array each time the declaration runs.
func zero(t *types.Type) Expr {
	switch t.Kind {
	case types.Signed, types.Unsigned:
		return &IntConst{T: t}
	case types.Float:
		return &FloatConst{T: t}
	case types.Char:
		return &CharConst{}
	case types.String:
		return &StringConst{}
	case types.Bool:
		return &BoolConst{}
	case types.Array:
		return &ArrayLit{T: t}
	case types.Struct:
		lit := &StructLit{T: t}
		for i, f := range t.Fields {
			lit.Inits = append(lit.Inits, FieldInit{Index: i, Value: zero(f.Type)})
		}
		return lit
	}
	panic("ir: no zero value for " + t.String())
}
