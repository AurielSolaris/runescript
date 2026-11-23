package lsp

import (
	"sort"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/checker"
	"github.com/AurielSolaris/runescript/internal/lexer"
)

// localName recognizes declarations tracked by the checker's VarRefs.
func localName(n *ast.Ident, parent node) bool {
	switch p := parent.(type) {
	case *ast.Param:
		return n == p.Name
	case *ast.VarDecl:
		return n == p.Name
	case *ast.CatchClause:
		return n == p.Name
	case *ast.ForEachStmt:
		return n == p.Name
	case *ast.MForEach:
		return n == p.Name
	}
	return false
}

func (s *Server) referenceTarget(p *textDocumentPositionParams, localsOnly bool) Location {
	a, off := s.at(p.TextDocument.URI, p.Position)
	if a == nil {
		return Location{}
	}
	path := pathAt(a.file.AST, off)
	if len(path) < 2 {
		return Location{}
	}
	id, ok := path[len(path)-1].(*ast.Ident)
	if !ok {
		return Location{}
	}
	parent := path[len(path)-2]
	var span ast.Span
	if decl := a.info.VarRefs[id]; decl != nil {
		span = decl.Span
	} else if localName(id, parent) {
		span = id.Span
	} else if !localsOnly {
		if sel, ok := parent.(*ast.SelectorExpr); ok && id == sel.Name && len(path) >= 3 {
			if call, ok := path[len(path)-3].(*ast.CallExpr); ok {
				if method := a.info.MethodCalls[call]; method != nil && method.Method != nil {
					span, _ = funcSpan(method.Method)
				}
			}
		}
		if span.Start.File == nil {
			span = a.functionTarget(id, parent)
		}
	}
	loc, _ := s.location(a, span)
	return loc
}

func (a *analysis) functionTarget(id *ast.Ident, parent node) ast.Span {
	info := a.info
	f := info.FuncRefs[id]
	switch p := parent.(type) {
	case *ast.CallExpr:
		if id == p.Fn {
			f = info.Callees[p]
			if call := info.MethodCalls[p]; call != nil {
				f = call.Method
			}
			if ext := info.Externs[p]; ext != nil {
				return ext.Decl.Name.Span
			}
		}
	case *ast.SelectorExpr:
		if id == p.Name {
			f = info.QualRefs[p]
		}
	case *ast.FuncDecl:
		if id == p.Name {
			if fn := info.FuncDecls[p]; fn != nil {
				f = fn
			} else {
				return id.Span
			}
		}
	case *ast.ExternDecl:
		if id == p.Name {
			return id.Span
		}
	}
	if f != nil {
		span, _ := funcSpan(f)
		return span
	}
	return ast.Span{}
}

func posBefore(a, b Position) bool {
	return a.Line < b.Line || a.Line == b.Line && a.Character < b.Character
}

func (s *Server) references(p *referenceParams) []Location {
	target := s.referenceTarget(&p.textDocumentPositionParams, false)
	if target.URI == "" {
		return nil
	}
	matches := map[Location]bool{}
	if p.Context.IncludeDeclaration {
		matches[target] = true
	}
	seen := map[*checker.Info]bool{}
	for _, d := range s.docs {
		a := d.last
		if a == nil || a.info == nil || seen[a.info] {
			continue
		}
		seen[a.info] = true
		add := func(ref, decl ast.Span) {
			if loc, ok := s.location(a, decl); ok && loc == target {
				if use, ok := s.location(a, ref); ok && use != target {
					matches[use] = true
				}
			}
		}
		addFunc := func(ref ast.Span, fn *checker.FuncSym) {
			if def, ok := funcSpan(fn); ok {
				add(ref, def)
			}
		}
		for id, decl := range a.info.VarRefs {
			add(id.Span, decl.Span)
		}
		for id, fn := range a.info.FuncRefs {
			addFunc(id.Span, fn)
		}
		for expr, fn := range a.info.QualRefs {
			addFunc(expr.Name.Span, fn)
		}
		for call, fn := range a.info.Callees {
			addFunc(call.Fn.Span, fn)
		}
		for call, ext := range a.info.Externs {
			add(call.Fn.Span, ext.Decl.Name.Span)
		}
		for call, method := range a.info.MethodCalls {
			if method.Method != nil {
				if call.Fn != nil {
					addFunc(call.Fn.Span, method.Method)
				} else if sel, ok := call.Callee.(*ast.SelectorExpr); ok {
					addFunc(sel.Name.Span, method.Method)
				}
			}
		}
		if p.Context.IncludeDeclaration {
			for decl, fn := range a.info.FuncDecls {
				addFunc(decl.Name.Span, fn)
			}
		}
	}
	out := make([]Location, 0, len(matches))
	for loc := range matches {
		out = append(out, loc)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].URI != out[j].URI {
			return out[i].URI < out[j].URI
		}
		return posBefore(out[i].Range.Start, out[j].Range.Start)
	})
	return out
}

func (s *Server) localRenameTarget(p *textDocumentPositionParams) Location {
	return s.referenceTarget(p, true)
}

func (s *Server) prepareRename(p *textDocumentPositionParams) *Range {
	if s.localRenameTarget(p).URI == "" {
		return nil
	}
	a, off := s.at(p.TextDocument.URI, p.Position)
	path := pathAt(a.file.AST, off)
	sp, _ := spanOf(path[len(path)-1])
	r := s.toRange(sp, a.file.Source.Text)
	return &r
}

func (s *Server) rename(p *renameParams) (*workspaceEdit, *rpcError) {
	target := s.localRenameTarget(&p.textDocumentPositionParams)
	if target.URI == "" {
		return nil, &rpcError{Code: codeInvalidParams, Message: "select a local variable or parameter to rename"}
	}
	if !lexer.IsName(p.NewName) || p.NewName == "trace" {
		return nil, &rpcError{Code: codeInvalidParams, Message: "the new name must be a valid RuneScript identifier"}
	}
	a, _ := s.at(p.TextDocument.URI, p.Position)
	decl := toOffset(a.file.Source.Text, target.Range.Start, s.utf16)
	end := toOffset(a.file.Source.Text, target.Range.End, s.utf16)
	if string(a.file.Source.Text[decl:end]) == p.NewName {
		return &workspaceEdit{Changes: map[string][]textEdit{}}, nil
	}
	// A new name already mentioned in the surrounding function could shadow
	// an existing use. Reject it before producing edits.
	scope := a.file.AST.Package.Span
	scope.End.Offset = len(a.file.Source.Text)
	for _, n := range pathAt(a.file.AST, decl) {
		switch n.(type) {
		case *ast.FuncDecl, *ast.LambdaExpr:
			scope, _ = spanOf(n)
		}
	}
	tokens, _ := lexer.LexSource(a.file.Source)
	for _, token := range tokens {
		if token.Kind == lexer.Ident && token.Text == p.NewName &&
			token.Span.Start.Offset >= scope.Start.Offset && token.Span.Start.Offset < scope.End.Offset {
			return nil, &rpcError{Code: codeInvalidParams, Message: "the new name is already used in this function"}
		}
	}
	params := &referenceParams{textDocumentPositionParams: p.textDocumentPositionParams}
	params.Context.IncludeDeclaration = true
	edits := &workspaceEdit{Changes: map[string][]textEdit{}}
	for _, loc := range s.references(params) {
		edits.Changes[loc.URI] = append(edits.Changes[loc.URI], textEdit{Range: loc.Range, NewText: p.NewName})
	}
	return edits, nil
}
