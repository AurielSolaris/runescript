package lsp

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestReferencesRenameAndSignature(t *testing.T) {
	const src = `package app;
public int sum(int a, int b);
public int sum(int a, int b) { return a + b; }
protected int main() {
    int n = 1;
    if (true) {
        int n = 2;
        trace(n);
    }
    trace(sum(n, 4));
    return 0;
}
`
	dir := t.TempDir()
	c := startServer(t, dir, false)
	defer c.shutdown()
	uri := c.open(filepath.Join(dir, "main.rsc"), src)
	at := func(offset int) textDocumentPositionParams {
		return textDocumentPositionParams{TextDocument: textDocumentIdentifier{URI: uri}, Position: toPosition([]byte(src), offset, true)}
	}
	call := strings.Index(src, "sum(n, 4)")
	refs := referenceParams{textDocumentPositionParams: at(call)}
	refs.Context.IncludeDeclaration = true
	var locs []Location
	c.request("textDocument/references", refs, &locs)
	if len(locs) != 3 {
		t.Fatalf("sum references: %+v", locs)
	}
	var edit workspaceEdit
	c.request("textDocument/rename", renameParams{textDocumentPositionParams: at(call + 4), NewName: "count"}, &edit)
	if len(edit.Changes[uri]) != 2 || edit.Changes[uri][1].NewText != "count" {
		t.Fatalf("rename should reach only the outer binding: %+v", edit)
	}
	var help signatureHelp
	c.request("textDocument/signatureHelp", at(call+7), &help)
	if len(help.Signatures) != 1 || help.ActiveParameter != 1 || help.Signatures[0].Parameters[1].Label != "int b" {
		t.Fatalf("signature at second argument: %+v", help)
	}
	incomplete := strings.Replace(src, "sum(n, 4)", "sum(n, ", 1)
	c.change(uri, incomplete, 2)
	c.request("textDocument/signatureHelp", at(call+7), &help)
	if len(help.Signatures) != 1 || help.ActiveParameter != 1 {
		t.Fatalf("signature in incomplete source: %+v", help)
	}
}

func TestFormattingAndCodeActions(t *testing.T) {
	const src = "package app;\nuse \"std:string\";\nprotected int main() {\n  if (true) {\ntrace(\"x\");\n }\n return 0;\n}\n"
	dir := t.TempDir()
	c := startServer(t, dir, false)
	defer c.shutdown()
	uri := c.open(filepath.Join(dir, "main.rsc"), src)
	var p formattingParams
	p.TextDocument.URI, p.Options.TabSize, p.Options.InsertSpaces = uri, 2, true
	var edits []textEdit
	c.request("textDocument/formatting", p, &edits)
	if len(edits) != 3 || edits[0].Range.Start.Line != 4 || edits[0].NewText != "    " {
		t.Fatalf("indent edits: %+v", edits)
	}
	var actions []codeAction
	c.request("textDocument/codeAction", codeActionParams{
		TextDocument: textDocumentIdentifier{URI: uri},
		Range:        Range{Start: Position{Line: 1}, End: Position{Line: 1, Character: 17}},
	}, &actions)
	if len(actions) != 1 || actions[0].Kind != "quickfix" || len(actions[0].Edit.Changes[uri]) != 1 {
		t.Fatalf("unused use action: %+v", actions)
	}
}
