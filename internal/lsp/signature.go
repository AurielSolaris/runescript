package lsp

import (
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/lexer"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

type argumentList struct {
	open      int
	parameter int
}

// currentArguments counts only the commas at the call's own nesting level.
func currentArguments(tokens []lexer.Token, offset int) argumentList {
	var stack []argumentList
	for i, t := range tokens {
		if t.Span.Start.Offset >= offset {
			break
		}
		switch t.Kind {
		case lexer.LParen, lexer.LBracket, lexer.LBrace:
			stack = append(stack, argumentList{open: i})
		case lexer.RParen, lexer.RBracket, lexer.RBrace:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case lexer.Comma:
			if len(stack) > 0 {
				stack[len(stack)-1].parameter++
			}
		}
	}
	for i := len(stack) - 1; i >= 0; i-- {
		open := stack[i].open
		if open > 0 && tokens[open].Kind == lexer.LParen {
			switch tokens[open-1].Kind {
			case lexer.Ident, lexer.RParen, lexer.RBracket:
				return stack[i]
			}
		}
	}
	return argumentList{}
}

func (s *Server) signatureHelp(p *textDocumentPositionParams) *signatureHelp {
	d := s.docs[p.TextDocument.URI]
	if d == nil {
		return nil
	}
	offset := toOffset(d.text, p.Position, s.utf16)
	tokens, _ := lexer.Lex(d.text)
	args := currentArguments(tokens, offset)
	if args.open == 0 {
		return nil
	}
	a, off := s.at(p.TextDocument.URI, p.Position)
	var sigs []signatureInformation
	if a != nil {
		path := pathAt(a.file.AST, off)
		for i := len(path) - 1; i >= 0; i-- {
			if call, ok := path[i].(*ast.CallExpr); ok {
				sigs = a.callSignatures(call)
				break
			}
			if expr, ok := path[i].(*ast.NewExpr); ok {
				if member := a.info.Members[a.info.Types[expr]]; member != nil && member.Ctor != nil {
					if fn := member.Ctor.Def; fn != nil {
						params := labels(fn.Params)
						sigs = []signatureInformation{signature(fn.Name.Name+"("+strings.Join(params, ", ")+")", params)}
					} else {
						sigs = []signatureInformation{signature(member.Type.Name+"()", nil)}
					}
				}
				break
			}
		}
	}
	if len(sigs) == 0 {
		sigs = incompleteSignatures(d.good, tokens, args.open)
	}
	if len(sigs) == 0 {
		return nil
	}
	return &signatureHelp{Signatures: sigs, ActiveParameter: min(args.parameter, max(0, len(sigs[0].Parameters)-1))}
}

func signature(label string, names []string) signatureInformation {
	params := make([]parameterInformation, len(names))
	for i, name := range names {
		params[i] = parameterInformation{Label: name}
	}
	return signatureInformation{Label: label, Parameters: params}
}

func labels(params []*ast.Param) []string {
	out := make([]string, len(params))
	for i, p := range params {
		out[i] = p.Type.String()
		if p.Name != nil {
			out[i] += " " + p.Name.Name
		}
	}
	return out
}

func typedLabels(params []*types.Type) []string {
	out := make([]string, len(params))
	for i, p := range params {
		out[i] = p.String()
	}
	return out
}

func (a *analysis) callSignatures(call *ast.CallExpr) []signatureInformation {
	info := a.info
	fn := info.Callees[call]
	if method := info.MethodCalls[call]; method != nil {
		fn = method.Method
	}
	if fn != nil {
		decl := fn.Def
		if decl == nil {
			decl = fn.Proto
		}
		return []signatureInformation{signature(declSig(decl), labels(decl.Params))}
	}
	if nf := info.Natives[call]; nf != nil {
		return nativeSignatures(call.Pkg.Name, nf, stdlib.Lookup(nf.Package))
	}
	if ext := info.Externs[call]; ext != nil {
		return []signatureInformation{signature(externSig(ext), labels(ext.Decl.Params))}
	}
	var t *types.Type
	var name string
	switch {
	case call.Callee != nil:
		t = info.Types[call.Callee]
	case call.Fn != nil:
		t, name = info.Types[call.Fn], call.Fn.Name
	}
	if t != nil && t.Kind == types.Func {
		params := typedLabels(t.Params)
		return []signatureInformation{signature(t.Result.String()+" "+name+"("+strings.Join(params, ", ")+")", params)}
	}
	return nil
}

func nativeSignatures(qual string, first *stdlib.Func, pkg *stdlib.Package) []signatureInformation {
	sigs := []signatureInformation{signature(nativeSig(qual, first), typedLabels(first.Params))}
	for _, fn := range pkg.Overloads[first.Name] {
		if fn != first {
			sigs = append(sigs, signature(nativeSig(qual, fn), typedLabels(fn.Params)))
		}
	}
	return sigs
}

func incompleteSignatures(a *analysis, tokens []lexer.Token, open int) []signatureInformation {
	name := tokens[open-1].Text
	if name == "trace" {
		return []signatureInformation{signature("void trace(...)", []string{"..."})}
	}
	if a == nil || a.file == nil {
		return nil
	}
	if open >= 3 && tokens[open-2].Kind == lexer.Dot {
		q := tokens[open-3].Text
		for _, u := range a.file.Uses {
			if qualifier(u) != q {
				continue
			}
			if u.Kind == loader.UseStd {
				if pkg := stdlib.Lookup(u.Name); pkg != nil && pkg.Funcs[name] != nil {
					return nativeSignatures(q, pkg.Funcs[name], pkg)
				}
			}
			if u.Target != nil {
				return packageSignatures(u.Target, a.file, name)
			}
		}
		return nil
	}
	return packageSignatures(a.file.Package, a.file, name)
}

func packageSignatures(pkg *loader.Package, file *loader.File, name string) []signatureInformation {
	if pkg == nil {
		return nil
	}
	for _, f := range pkg.Files {
		if f.AST == nil {
			continue
		}
		for _, fn := range f.AST.Funcs {
			if fn.Name.Name == name && (f == file || fn.Visibility == ast.VisPublic ||
				fn.Visibility == ast.VisProtected && pkg == file.Package) {
				return []signatureInformation{signature(declSig(fn), labels(fn.Params))}
			}
		}
	}
	return nil
}
