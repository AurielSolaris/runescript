package ir

import (
	"fmt"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/types"
)

// Lower converts a program that checked without errors into IR. It
// panics if info is incomplete, which would be a checker bug, not a user
// error.
func Lower(info *checker.Info) *Program {
	p := &Program{EntryVoid: info.EntryVoid, EntrySite: info.Entry.Name.Span}
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
	// A function boundary starts a fresh chain: functions do not capture.
	scopes []map[string]*Var
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
	if ref.Func != nil {
		params := make([]*types.Type, len(ref.Func.Params))
		for i, p := range ref.Func.Params {
			params[i] = l.typeOf(p)
		}
		return types.NewFunc(params, l.typeOf(ref.Func.Result))
	}
	t := types.Lookup(ref.Name)
	if t == nil {
		l.ice("unknown type %q", ref.Name)
	}
	return t
}

func (l *lowerer) boundary() (restore func()) {
	saved := l.scopes
	l.scopes = []map[string]*Var{{}}
	return func() { l.scopes = saved }
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
		init := l.expr(s.Init) // the initializer cannot see the new variable
		return &Decl{Var: l.declare(s.Name.Name, l.typeOf(s.Type), false), Init: init}
	case *ast.AssignStmt:
		v := l.lookup(s.Target.Name)
		value := l.expr(s.Value)
		if s.Op != "=" {
			value = &Arith{Op: s.Op[:1], T: v.Type, X: &Load{Var: v}, Y: value, Site: s.Span}
		}
		return &Assign{Var: v, Value: value}
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
		return &Field{X: l.expr(e.X), Name: e.Name.Name, T: l.concrete(e, l.info.Types[e])}
	case *ast.ParenExpr:
		return l.expr(e.X)
	case *ast.StringLit:
		return &StringConst{Value: e.Value, Span: e.Span}
	case *ast.BoolLit:
		return &BoolConst{Value: e.Value}
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
		return &Arith{Op: e.Op, T: l.concrete(e, l.info.Types[e]), X: x, Y: y, Site: e.OpPos}
	}
	l.ice("unexpected expression %T", e)
	return nil
}

func (l *lowerer) call(e *ast.CallExpr) Expr {
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
