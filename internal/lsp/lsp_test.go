package lsp

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// client drives a Server over pipes, as an editor does.
type client struct {
	t      *testing.T
	w      io.WriteCloser
	msgs   chan map[string]json.RawMessage
	done   chan error
	nextID int
	// diags holds the latest diagnostics published for each URI.
	diags map[string][]Diagnostic
}

func startServer(t *testing.T, root string, utf8 bool) *client {
	t.Helper()
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	c := &client{t: t, w: inW, msgs: make(chan map[string]json.RawMessage, 100), done: make(chan error, 1), diags: map[string][]Diagnostic{}}
	s := NewServer(inR, outW)
	go func() {
		c.done <- s.Run()
		outW.Close()
	}()
	go func() {
		r := bufio.NewReader(outR)
		for {
			body, err := readMessage(r)
			if err != nil {
				close(c.msgs)
				return
			}
			var m map[string]json.RawMessage
			json.Unmarshal(body, &m)
			c.msgs <- m
		}
	}()
	caps := map[string]any{}
	if utf8 {
		caps["general"] = map[string]any{"positionEncodings": []string{"utf-8", "utf-16"}}
	}
	var res struct {
		Capabilities map[string]any `json:"capabilities"`
	}
	c.request("initialize", map[string]any{"rootUri": pathToURI(root), "capabilities": caps}, &res)
	if res.Capabilities["hoverProvider"] != true {
		t.Fatalf("initialize: capabilities %v", res.Capabilities)
	}
	c.notify("initialized", map[string]any{})
	return c
}

func (c *client) send(v any) {
	c.t.Helper()
	if err := writeMessage(c.w, v); err != nil {
		c.t.Fatal(err)
	}
}

func (c *client) notify(method string, params any) {
	c.send(notification{JSONRPC: "2.0", Method: method, Params: params})
}

// request sends a request and decodes its result into out. Notifications
// that arrive first are recorded.
func (c *client) request(method string, params any, out any) {
	c.t.Helper()
	c.nextID++
	id := c.nextID
	c.send(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	for {
		m := c.next()
		if m["id"] == nil {
			continue
		}
		var got int
		json.Unmarshal(m["id"], &got)
		if got != id {
			c.t.Fatalf("response to %d, want %d", got, id)
		}
		if m["error"] != nil {
			c.t.Fatalf("%s: error %s", method, m["error"])
		}
		if out != nil {
			if err := json.Unmarshal(m["result"], out); err != nil {
				c.t.Fatalf("%s: %v in %s", method, err, m["result"])
			}
		}
		return
	}
}

// next returns the next message, recording published diagnostics.
func (c *client) next() map[string]json.RawMessage {
	c.t.Helper()
	select {
	case m, ok := <-c.msgs:
		if !ok {
			c.t.Fatal("the server closed its output")
		}
		var method string
		json.Unmarshal(m["method"], &method)
		if method == "textDocument/publishDiagnostics" {
			var p publishDiagnosticsParams
			json.Unmarshal(m["params"], &p)
			c.diags[p.URI] = p.Diagnostics
		}
		return m
	case <-time.After(10 * time.Second):
		c.t.Fatal("timed out waiting for the server")
	}
	return nil
}

// sync waits until every earlier notification has been handled: the
// server handles messages in order, so a request's response comes after
// the diagnostics of everything before it.
func (c *client) sync() {
	c.t.Helper()
	c.request("textDocument/documentSymbol", map[string]any{"textDocument": map[string]any{"uri": "file:///nothing"}}, nil)
}

func (c *client) open(path, text string) string {
	uri := pathToURI(path)
	c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "runescript", "version": 1, "text": text}})
	c.sync()
	return uri
}

func (c *client) change(uri, text string, version int) {
	c.notify("textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": uri, "version": version},
		"contentChanges": []map[string]any{{"text": text}},
	})
	c.sync()
}

func (c *client) shutdown() {
	c.t.Helper()
	c.request("shutdown", nil, nil)
	c.notify("exit", nil)
	select {
	case err := <-c.done:
		if err != nil {
			c.t.Fatalf("server: %v", err)
		}
	case <-time.After(5 * time.Second):
		c.t.Fatal("server did not exit")
	}
}

func pos(line, char int) map[string]any {
	return map[string]any{"line": line, "character": char}
}

func at(uri string, line, char int) map[string]any {
	return map[string]any{"textDocument": map[string]any{"uri": uri}, "position": pos(line, char)}
}

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func codes(diags []Diagnostic) string {
	var out []string
	for _, d := range diags {
		out = append(out, d.Code)
	}
	return strings.Join(out, " ")
}

const signedProgram = `package main;

protected int main() {
    sint s = -10;
    uint u = s;
    return 0;
}
`

func TestDiagnosticsFollowEdits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rsc")
	write(t, path, signedProgram)
	c := startServer(t, dir, false)
	uri := c.open(path, signedProgram)
	ds := c.diags[uri]
	if len(ds) != 1 || ds[0].Code != "ASX-TYPE-004" || ds[0].Severity != severityError || ds[0].Source != "runescript" {
		t.Fatalf("diagnostics %+v", ds)
	}
	if r := ds[0].Range; r.Start != (Position{4, 13}) || r.End != (Position{4, 14}) {
		t.Errorf("range %+v, want 4:13-4:14", r)
	}
	if !strings.Contains(ds[0].Message, "help: ") {
		t.Errorf("message without help: %q", ds[0].Message)
	}
	// The fix is checked as typed, unsaved.
	c.change(uri, strings.Replace(signedProgram, "uint u = s;", "trace(s);", 1), 2)
	if ds := c.diags[uri]; len(ds) != 0 {
		t.Errorf("after the fix: %+v", ds)
	}
	// A syntax error, then a warning.
	c.change(uri, "package main;\nprotected int main() {\n    int x = ;\n    return 0;\n}\n", 3)
	if got := codes(c.diags[uri]); got != "ASX-PARSE-001" {
		t.Errorf("syntax error: %s", got)
	}
	c.change(uri, "package main;\nprotected int main() {\n    int unused = 1;\n    return 0;\n}\n", 4)
	if ds := c.diags[uri]; len(ds) != 1 || ds[0].Severity != severityWarning || ds[0].Code != "ASX-WARN-001" {
		t.Errorf("warning: %+v", ds)
	}
	// Closing a document clears its diagnostics.
	c.notify("textDocument/didClose", map[string]any{"textDocument": map[string]any{"uri": uri}})
	c.sync()
	if ds, ok := c.diags[uri]; !ok || len(ds) != 0 {
		t.Errorf("after close: %+v", ds)
	}
	c.shutdown()
}

func TestHelperWithoutMain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "helpers.rsc")
	text := "package main;\npublic int twice(int x) {\n    return x * 2;\n}\n"
	write(t, path, text)
	c := startServer(t, dir, false)
	uri := c.open(path, text)
	if ds := c.diags[uri]; len(ds) != 0 {
		t.Errorf("a file without main: %+v", ds)
	}
	c.shutdown()
}

// A module is checked with the rest of its package, and open files are
// read from the editor, not the disk.
func TestModulesAndUnsavedFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "geom", "a.rsm")
	b := filepath.Join(dir, "geom", "b.rsm")
	main := filepath.Join(dir, "main.rsc")
	aText := "package geom;\npublic int area(int w, int h) {\n    return scale(w * h);\n}\n"
	bText := "package geom;\nprotected int scale(int x) {\n    return x;\n}\n"
	mainText := "package main;\nuse \"geom\";\nprotected int main() {\n    trace(geom.area(2, 3));\n    return 0;\n}\n"
	write(t, a, aText)
	write(t, b, bText)
	write(t, main, mainText)
	c := startServer(t, dir, false)
	aURI := c.open(a, aText)
	if ds := c.diags[aURI]; len(ds) != 0 {
		t.Fatalf("a.rsm uses b.rsm's function: %+v", ds)
	}
	mainURI := c.open(main, mainText)
	if ds := c.diags[mainURI]; len(ds) != 0 {
		t.Fatalf("main.rsc: %+v", ds)
	}
	// Rename area in the editor without saving: main.rsc sees it.
	c.change(aURI, strings.Replace(aText, "area", "surface", 1), 2)
	if got := codes(c.diags[mainURI]); got != "ASX-NAME-007" && !strings.Contains(got, "NAME") {
		t.Errorf("main.rsc after renaming area: %s %+v", got, c.diags[mainURI])
	}
	// An error inside the module is reported in the module.
	c.change(aURI, strings.Replace(aText, "scale(w * h)", "scale(\"x\")", 1), 3)
	if got := codes(c.diags[aURI]); got != "ASX-TYPE-001" {
		t.Errorf("a.rsm with a type error: %s", got)
	}
	c.shutdown()
}

const featureProgram = `package main;

use "std:csv";

public struct Point {
    int x;
    int y;
}

public int sum(Array<int> xs) {
    int total = 0;
    for each (int v in xs) {
        total = total + v;
    }
    return total;
}

protected int main() {
    Array<int> nums = [1, 2, 3];
    Point p = Point{x: 1, y: sum(nums)};
    Array<Array<string>> rows = csv.parse("a,b");
    trace(p.y, rows.length);
    return 0;
}
`

func TestHoverDefinitionCompletionSymbols(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rsc")
	write(t, path, featureProgram)
	c := startServer(t, dir, false)
	uri := c.open(path, featureProgram)
	if ds := c.diags[uri]; len(ds) != 0 {
		t.Fatalf("diagnostics: %+v", ds)
	}
	lines := strings.Split(featureProgram, "\n")
	find := func(line int, word string) map[string]any {
		col := strings.Index(lines[line], word)
		if col < 0 {
			t.Fatalf("%q not on line %d", word, line)
		}
		return at(uri, line, col+1)
	}

	hoverCases := []struct {
		line int
		word string
		want string
	}{
		{19, "sum", "public int sum(Array<int> xs)"},
		{19, "nums", "Array<int> nums"},
		{19, "Point{", "public struct Point {"},
		{20, "parse", "Array<Array<string>> csv.parse(string)"},
		{20, "csv", "std:csv"},
		{12, "total +", "int total"},
		{11, "v in", "int v"},
		{9, "xs", "Array<int> xs"},
		{21, "y,", "int y"},
	}
	for _, hc := range hoverCases {
		var h hover
		c.request("textDocument/hover", find(hc.line, hc.word), &h)
		if !strings.Contains(h.Contents.Value, hc.want) {
			t.Errorf("hover %q on line %d: %q, want %q", hc.word, hc.line, h.Contents.Value, hc.want)
		}
	}

	defCases := []struct {
		line int
		word string
		want Position
	}{
		{19, "sum", Position{9, 11}},    // the function's name
		{19, "nums)", Position{18, 15}}, // the variable's declaration
		{19, "Point{", Position{4, 14}}, // the struct's name
		{12, "total +", Position{10, 8}},
		{21, "y,", Position{6, 8}}, // the field
	}
	for _, dc := range defCases {
		var locs []Location
		c.request("textDocument/definition", find(dc.line, dc.word), &locs)
		if len(locs) != 1 || locs[0].Range.Start != dc.want || pathKey(uriToPath(locs[0].URI)) != pathKey(path) {
			t.Errorf("definition of %q on line %d: %+v, want %v", dc.word, dc.line, locs, dc.want)
		}
	}

	labels := func(params map[string]any) map[string]bool {
		var items []completionItem
		c.request("textDocument/completion", params, &items)
		out := map[string]bool{}
		for _, it := range items {
			out[it.Label] = true
		}
		return out
	}
	// Edit the document to complete at the end of a line.
	edit := func(line int, text string) map[string]any {
		ls := append([]string{}, lines...)
		ls[line] = text
		c.change(uri, strings.Join(ls, "\n"), 10+line)
		return at(uri, line, len(text))
	}
	if got := labels(edit(21, "    trace(csv.")); !got["parse"] || !got["formatWith"] || got["while"] {
		t.Errorf("after csv.: %v", got)
	}
	if got := labels(edit(21, "    nums.")); !got["push"] || !got["length"] {
		t.Errorf("after nums.: %v", got)
	}
	if got := labels(edit(21, "    p.")); !got["x"] || !got["y"] || len(got) != 2 {
		t.Errorf("after p.: %v", got)
	}
	if got := labels(edit(21, "    tr")); !got["trace"] || !got["nums"] || !got["rows"] || !got["sum"] || !got["Point"] || !got["while"] || !got["csv"] || got["total"] {
		t.Errorf("in main: %v", got)
	}
	if got := labels(edit(2, `use "std:`)); !got["csv"] || !got["semver"] || got["while"] {
		t.Errorf("in a use: %v", got)
	}
	c.change(uri, featureProgram, 99)

	var syms []documentSymbol
	c.request("textDocument/documentSymbol", map[string]any{"textDocument": map[string]any{"uri": uri}}, &syms)
	var names []string
	for _, s := range syms {
		names = append(names, s.Name)
	}
	if strings.Join(names, " ") != "Point sum main" || len(syms[0].Children) != 2 {
		t.Errorf("symbols: %v", syms)
	}
	c.shutdown()
}

func TestUTF8Positions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rsc")
	text := "package main;\nprotected int main() {\n    string e = \"é😀\"; sint s = -1; uint u = s;\n    return 0;\n}\n"
	write(t, path, text)
	for _, utf8 := range []bool{false, true} {
		c := startServer(t, dir, utf8)
		uri := c.open(path, text)
		ds := c.diags[uri]
		if len(ds) == 0 {
			t.Fatal("no diagnostics")
		}
		// `s` in `= s;`: after "    string é = "😀"; sint s = -1; uint u = ".
		before := "    string e = \"é😀\"; sint s = -1; uint u = "
		want := utf16Len([]byte(before))
		if utf8 {
			want = len(before)
		}
		if got := ds[len(ds)-1].Range.Start; got != (Position{2, want}) {
			t.Errorf("utf8=%v: %+v, want 2:%d", utf8, got, want)
		}
		c.shutdown()
	}
}

func TestPositionsAndURIs(t *testing.T) {
	text := []byte("ab\né😀x\n")
	for off, want := range map[int]Position{0: {0, 0}, 3: {1, 0}, 5: {1, 1}, 9: {1, 3}, 10: {1, 4}} {
		if got := toPosition(text, off, true); got != want {
			t.Errorf("toPosition(%d) = %v, want %v", off, got, want)
		}
		if got := toOffset(text, want, true); got != off {
			t.Errorf("toOffset(%v) = %d, want %d", want, got, off)
		}
	}
	if got := toOffset(text, Position{0, 99}, true); got != 2 {
		t.Errorf("past the end of a line: %d", got)
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "a dir", "x.rsc")
	if back := uriToPath(pathToURI(p)); back != p {
		t.Errorf("round trip %q -> %q -> %q", p, pathToURI(p), back)
	}
	if uriToPath("untitled:Untitled-1") != "" {
		t.Error("untitled URI")
	}
}

func TestRequestsBeforeInitialize(t *testing.T) {
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	s := NewServer(inR, outW)
	go s.Run()
	go writeMessage(inW, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "textDocument/hover", "params": map[string]any{}})
	body, err := readMessage(bufio.NewReader(outR))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"code":-32002`) {
		t.Errorf("got %s", body)
	}
}
