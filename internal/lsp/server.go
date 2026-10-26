// Package lsp is RuneScript's language server: `runescript lsp` speaks the
// Language Server Protocol on standard input and output, so that editors
// (VS Code, Vim, Neovim, Emacs, ...) show the compiler's diagnostics as
// you type, and offer hover, go-to-definition, completion, and an outline.
//
// It reuses the compiler: every change re-checks the program around each
// open document with the loader and checker, reading open files from the
// editor's buffers (loader.Options.Overlay). Nothing is cached between
// checks except the last successful analysis of each document, which
// completion falls back on while the document has a syntax error.
package lsp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"sync"
)

// Server is one language server session.
type Server struct {
	in  *bufio.Reader
	out io.Writer
	mu  sync.Mutex // serializes writes to out

	// roots are the workspace folders, as file system paths; root is the
	// first one, or "".
	roots []string
	root  string
	// utf16 is true unless the client agreed to UTF-8 positions.
	utf16 bool

	docs        map[string]*document // open documents, by URI
	initialized bool
	shutdown    bool
	// Log, if set, receives a line for each request that failed.
	Log io.Writer
}

// document is an open text document.
type document struct {
	uri     string
	path    string // "" for an unsaved document without a file
	text    []byte
	version int
	// last is the latest analysis; good is the latest one whose file
	// parsed, for completion while the document has a syntax error.
	last, good *analysis
}

// NewServer returns a server reading requests from in and writing
// responses to out.
func NewServer(in io.Reader, out io.Writer) *Server {
	return &Server{in: bufio.NewReader(in), out: out, utf16: true, docs: map[string]*document{}}
}

// Serve runs a language server on in and out until the client exits. It
// returns nil after a shutdown request and the exit notification.
func Serve(in io.Reader, out io.Writer) error {
	return NewServer(in, out).Run()
}

// errExit ends Run.
var errExit = errors.New("exit")

// Run reads and handles messages until the exit notification or the end
// of the input.
func (s *Server) Run() error {
	for {
		body, err := readMessage(s.in)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		var msg message
		if err := json.Unmarshal(body, &msg); err != nil {
			s.reply(nil, nil, &rpcError{Code: codeParseError, Message: err.Error()})
			continue
		}
		if err := s.handle(&msg); err == errExit {
			if !s.shutdown {
				return errors.New("lsp: exit without shutdown")
			}
			return nil
		}
	}
}

// handle dispatches one message.
func (s *Server) handle(msg *message) error {
	isRequest := len(msg.ID) > 0 && msg.Method != ""
	if msg.Method == "" {
		return nil // a response to a request of ours; the server sends none
	}
	if !s.initialized && msg.Method != "initialize" && msg.Method != "exit" {
		if isRequest {
			s.reply(msg.ID, nil, &rpcError{Code: codeServerNotInitialized, Message: "the server is not initialized"})
		}
		return nil
	}
	var result any
	var rerr *rpcError
	decode := func(v any) bool {
		if err := json.Unmarshal(msg.Params, v); err != nil {
			rerr = &rpcError{Code: codeInvalidParams, Message: err.Error()}
			return false
		}
		return true
	}
	switch msg.Method {
	case "initialize":
		var p initializeParams
		if decode(&p) {
			result = s.initialize(&p)
		}
	case "initialized":
	case "shutdown":
		s.shutdown = true
	case "exit":
		return errExit
	case "textDocument/didOpen":
		var p didOpenParams
		if decode(&p) {
			d := &document{uri: p.TextDocument.URI, path: uriToPath(p.TextDocument.URI), text: []byte(p.TextDocument.Text), version: p.TextDocument.Version}
			s.docs[d.uri] = d
			s.refresh()
		}
	case "textDocument/didChange":
		var p didChangeParams
		if decode(&p) {
			if d := s.docs[p.TextDocument.URI]; d != nil && len(p.ContentChanges) > 0 {
				// Full synchronization: the last change is the whole text.
				d.text = []byte(p.ContentChanges[len(p.ContentChanges)-1].Text)
				d.version = p.TextDocument.Version
				s.refresh()
			}
		}
	case "textDocument/didSave":
		s.refresh()
	case "textDocument/didClose":
		var p didCloseParams
		if decode(&p) {
			delete(s.docs, p.TextDocument.URI)
			s.publish(p.TextDocument.URI, nil, nil)
			s.refresh()
		}
	case "textDocument/hover":
		var p textDocumentPositionParams
		if decode(&p) {
			if h := s.hover(&p); h != nil {
				result = h
			}
		}
	case "textDocument/definition":
		var p textDocumentPositionParams
		if decode(&p) {
			if locs := s.definition(&p); locs != nil {
				result = locs
			}
		}
	case "textDocument/completion":
		var p textDocumentPositionParams
		if decode(&p) {
			result = s.completion(&p)
		}
	case "textDocument/documentSymbol":
		var p struct {
			TextDocument textDocumentIdentifier `json:"textDocument"`
		}
		if decode(&p) {
			result = s.documentSymbols(p.TextDocument.URI)
		}
	default:
		if isRequest {
			rerr = &rpcError{Code: codeMethodNotFound, Message: "unsupported method " + msg.Method}
		}
	}
	if isRequest {
		if rerr != nil && s.Log != nil {
			fmt.Fprintf(s.Log, "runescript lsp: %s: %s\n", msg.Method, rerr.Message)
		}
		s.reply(msg.ID, result, rerr)
	}
	return nil
}

func (s *Server) initialize(p *initializeParams) any {
	s.initialized = true
	for _, f := range p.WorkspaceFolders {
		if path := uriToPath(f.URI); path != "" {
			s.roots = append(s.roots, path)
		}
	}
	if len(s.roots) == 0 {
		if path := uriToPath(p.RootURI); path != "" {
			s.roots = append(s.roots, path)
		} else if p.RootPath != "" {
			s.roots = append(s.roots, filepath.Clean(p.RootPath))
		}
	}
	if len(s.roots) > 0 {
		s.root = s.roots[0]
	}
	encoding := "utf-16"
	for _, e := range p.Capabilities.General.PositionEncodings {
		if e == "utf-8" {
			encoding, s.utf16 = "utf-8", false
		}
	}
	return map[string]any{
		"capabilities": map[string]any{
			"positionEncoding": encoding,
			"textDocumentSync": map[string]any{
				"openClose": true,
				"change":    1, // full
				"save":      true,
			},
			"hoverProvider":          true,
			"definitionProvider":     true,
			"documentSymbolProvider": true,
			"completionProvider": map[string]any{
				"triggerCharacters": []string{".", ":", "\""},
			},
		},
		"serverInfo": map[string]any{"name": "runescript"},
	}
}

// refresh re-checks every open document and publishes its diagnostics.
// One change can affect other files (a function renamed in a package), so
// all of them are checked.
func (s *Server) refresh() {
	uris := make([]string, 0, len(s.docs))
	for uri := range s.docs {
		uris = append(uris, uri)
	}
	sort.Strings(uris)
	for _, uri := range uris {
		d := s.docs[uri]
		a := s.analyze(d)
		d.last = a
		if a.file != nil && a.file.AST != nil && a.info != nil {
			d.good = a
		}
		var text []byte
		if a.file != nil {
			text = a.file.Source.Text
		}
		s.publish(uri, &d.version, a.diagsFor(s, text))
	}
}

func (a *analysis) diagsFor(s *Server, text []byte) []Diagnostic {
	out := []Diagnostic{}
	for _, d := range a.diags {
		out = append(out, s.toDiagnostic(d, text))
	}
	return out
}

func (s *Server) publish(uri string, version *int, diags []Diagnostic) {
	if diags == nil {
		diags = []Diagnostic{}
	}
	s.notify("textDocument/publishDiagnostics", publishDiagnosticsParams{URI: uri, Version: version, Diagnostics: diags})
}

// overlay gives the loader the text of open documents.
func (s *Server) overlay(abs string) ([]byte, bool) {
	key := pathKey(filepath.Clean(abs))
	for _, d := range s.docs {
		if d.path != "" && pathKey(d.path) == key {
			return d.text, true
		}
	}
	return nil, false
}

func (s *Server) reply(id json.RawMessage, result any, err *rpcError) {
	if id == nil {
		id = json.RawMessage("null")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		writeMessage(s.out, errorResponse{JSONRPC: "2.0", ID: id, Error: err})
		return
	}
	writeMessage(s.out, response{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *Server) notify(method string, params any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeMessage(s.out, notification{JSONRPC: "2.0", Method: method, Params: params})
}
