package lsp

import (
	"bytes"
	"slices"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/lexer"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/parser"
)

func (s *Server) formatting(p *formattingParams) []textEdit {
	d := s.docs[p.TextDocument.URI]
	if d == nil {
		return nil
	}
	src := diagnostics.NewSource(d.path, d.text)
	f, ds := parser.ParseSource(src)
	if len(ds) > 0 {
		return nil
	}
	unit := "\t"
	if p.Options.InsertSpaces {
		size := p.Options.TabSize
		if size < 1 || size > 16 {
			size = 4
		}
		unit = strings.Repeat(" ", size)
	}
	var protected []ast.Span
	var protect func(node)
	protect = func(n node) {
		switch n := n.(type) {
		case *ast.TemplateStmt:
			protected = append(protected, n.Span)
		case *ast.InterpString:
			protected = append(protected, n.Span)
		default:
			for _, child := range children(n) {
				protect(child)
			}
		}
	}
	protect(f)
	slices.SortFunc(protected, func(a, b ast.Span) int { return a.Start.Offset - b.Start.Offset })

	tokens, _ := lexer.LexSource(src)
	var out []textEdit
	depth := 0
	for i := 0; i < len(tokens)-1; {
		first := tokens[i]
		line := first.Span.Start.Line
		j := i
		indent := depth
		leading := true
		skip := false
		for ; j < len(tokens)-1 && tokens[j].Span.Start.Line == line; j++ {
			t := tokens[j]
			for len(protected) > 0 && protected[0].End.Line < line {
				protected = protected[1:]
			}
			if len(protected) > 0 && protected[0].Start.Line <= line {
				skip = true
				// Interpolations and markup may contain their own braces.
				// Count ordinary code on the same line around them.
				n := t.Span.Start.Offset
				k := 0
				for k < len(protected) && protected[k].End.Offset <= n {
					k++
				}
				if k < len(protected) && protected[k].Start.Offset <= n {
					leading = false
					continue
				}
			}
			switch t.Kind {
			case lexer.RBrace:
				depth = max(depth-1, 0)
				if leading {
					indent = depth
				}
			case lexer.LBrace:
				depth++
				leading = false
			default:
				leading = false
			}
		}
		if !skip {
			off := first.Span.Start.Offset
			start := bytes.LastIndexByte(src.Text[:off], '\n') + 1
			want := strings.Repeat(unit, indent)
			have := string(src.Text[start:off])
			// The lexer skips comments. Leave any line whose first token
			// follows a comment alone.
			if strings.Trim(have, " \t") == "" && have != want {
				out = append(out, textEdit{Range: Range{
					Start: Position{Line: line - 1},
					End:   toPosition(src.Text, off, s.utf16),
				}, NewText: want})
			}
		}
		i = j
	}
	return out
}

func (s *Server) codeActions(p *codeActionParams) []codeAction {
	d := s.docs[p.TextDocument.URI]
	if d == nil || d.last == nil || d.last.file == nil || d.last.file.AST == nil ||
		len(p.Context.Only) > 0 && !slices.Contains(p.Context.Only, "quickfix") {
		return nil
	}
	var out []codeAction
	for _, diag := range d.last.diags {
		if diag.Code != diagnostics.WarnUnusedUse {
			continue
		}
		r := s.toRange(diag.Span, d.last.file.Source.Text)
		if posBefore(r.End, p.Range.Start) || posBefore(p.Range.End, r.Start) {
			continue
		}
		for _, u := range d.last.file.AST.Uses {
			// File and package uses can load an entry point even when
			// their qualifier is unused.
			if kind, _ := loader.Classify(u.Path); kind != loader.UseStd || u.Span != diag.Span {
				continue
			}
			out = append(out, codeAction{Title: "Remove unused use " + u.Path, Kind: "quickfix",
				Edit: workspaceEdit{Changes: map[string][]textEdit{d.uri: {{Range: r}}}}})
		}
	}
	return out
}
