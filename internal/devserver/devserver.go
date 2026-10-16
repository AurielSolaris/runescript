// Package devserver is `runescript dev`: a local web server that compiles a
// program to JavaScript, rebuilds it whenever one of its files changes,
// and reloads the browser.
//
// This is live reload: the page reloads after every successful build, and
// a failed build shows its errors in an overlay instead. Hot module
// replacement that keeps component state is planned for rune:dom
// components (OQ-DEV-01).
//
// The server listens on the loopback interface only. It serves the entry
// file's directory: the user's index.html (with the dev client injected)
// or a generated page, the compiled module from memory at the path
// `runescript build` would write it to (app.rsc → /app.js), and any other
// file there (helper modules for extern functions, styles, images).
package devserver

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/watch"
)

//go:embed client.js
var clientJS string

// Paths the server reserves for itself.
const (
	ClientPath = "/__runescript/client.js"
	EventsPath = "/__runescript/events"
)

// Build is the result of compiling the program once.
type Build struct {
	JS       []byte
	Errors   []*diagnostics.Diagnostic // compile errors (JS is nil)
	Warnings []*diagnostics.Diagnostic
	Files    []string // the files the program was read from
	Err      error    // an unreadable entry file or an internal error
}

func (b Build) ok() bool { return b.Err == nil && len(b.Errors) == 0 }

// Event is what the browser receives after each build.
type Event struct {
	// Server identifies this server run, so a client that reconnects to a
	// restarted server reloads.
	Server   string   `json:"server"`
	Version  int      `json:"version"`
	OK       bool     `json:"ok"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// Server is a dev server for one program.
type Server struct {
	root   string // the entry file's directory
	module string // URL path of the compiled module, e.g. "/app.js"
	title  string
	build  func() Build
	log    io.Writer
	id     string

	mu      sync.Mutex
	last    Build
	version int
	subs    map[chan Event]struct{}
}

// New returns a server for the program whose entry file is entry. build
// compiles it; log receives a line per build, and the diagnostics.
func New(entry string, build func() Build, log io.Writer) *Server {
	abs, err := filepath.Abs(entry)
	if err != nil {
		abs = entry
	}
	base := filepath.Base(abs)
	return &Server{
		root:   filepath.Dir(abs),
		module: "/" + strings.TrimSuffix(base, filepath.Ext(base)) + ".js",
		title:  base,
		build:  build,
		log:    log,
		id:     fmt.Sprint(time.Now().UnixNano()),
		subs:   map[chan Event]struct{}{},
	}
}

// Module returns the URL path the compiled module is served at.
func (s *Server) Module() string { return s.module }

// Rebuild compiles the program and tells every browser.
func (s *Server) Rebuild() {
	start := time.Now()
	b := s.build()
	s.mu.Lock()
	s.last = b
	s.version++
	ev := s.event()
	s.mu.Unlock()
	s.report(b, time.Since(start))
	s.broadcast(ev)
}

// Reload tells every browser to reload without rebuilding, e.g. after a
// style sheet changed.
func (s *Server) Reload() {
	s.mu.Lock()
	s.version++
	ev := s.event()
	s.mu.Unlock()
	fmt.Fprintln(s.log, "runescript dev: reloading")
	s.broadcast(ev)
}

// Changed handles changed files: a RuneScript source rebuilds, anything
// else only reloads.
func (s *Server) Changed(files []string) {
	for _, f := range files {
		if ext := filepath.Ext(f); ext == ".rsc" || ext == ".rsm" {
			s.Rebuild()
			return
		}
	}
	s.Reload()
}

// Dirs returns the directories to watch: those of the program's files and
// the served directory.
func (s *Server) Dirs() []string {
	s.mu.Lock()
	files := append([]string{filepath.Join(s.root, "index.html")}, s.last.Files...)
	s.mu.Unlock()
	return watch.Dirs(files...)
}

// event describes the last build. s.mu is held.
func (s *Server) event() Event {
	ev := Event{Server: s.id, Version: s.version, OK: s.last.ok(), Errors: []string{}, Warnings: []string{}}
	for _, d := range s.last.Errors {
		ev.Errors = append(ev.Errors, diagnostics.Render(d))
	}
	if s.last.Err != nil {
		ev.Errors = append(ev.Errors, "runescript: "+s.last.Err.Error())
	}
	for _, d := range s.last.Warnings {
		ev.Warnings = append(ev.Warnings, diagnostics.Render(d))
	}
	return ev
}

func (s *Server) report(b Build, took time.Duration) {
	for _, d := range b.Warnings {
		fmt.Fprint(s.log, diagnostics.Render(d))
	}
	for _, d := range b.Errors {
		fmt.Fprint(s.log, diagnostics.Render(d))
	}
	switch {
	case b.Err != nil:
		fmt.Fprintf(s.log, "runescript dev: %v\n", b.Err)
	case len(b.Errors) > 0:
		fmt.Fprintf(s.log, "runescript dev: build failed with %d error(s); the browser shows them\n", len(b.Errors))
	default:
		fmt.Fprintf(s.log, "runescript dev: built %s in %s\n", s.title, took.Round(time.Millisecond))
	}
}

func (s *Server) broadcast(ev Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.subs {
		select {
		case ch <- ev:
		default: // a slow client gets the next event
		}
	}
}

// Handler serves the page, the module, the dev client, the event stream,
// and the files of the entry file's directory.
func (s *Server) Handler() http.Handler {
	files := http.FileServer(http.Dir(s.root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		switch r.URL.Path {
		case "/", "/index.html":
			s.page(w)
		case s.module:
			s.serveModule(w)
		case ClientPath:
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			io.WriteString(w, clientJS)
		case EventsPath:
			s.events(w, r)
		case "/favicon.ico":
			// Browsers ask for it; without one, answer quietly instead of
			// logging a 404 on every page load.
			if _, err := os.Stat(filepath.Join(s.root, "favicon.ico")); err != nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			files.ServeHTTP(w, r)
		default:
			files.ServeHTTP(w, r)
		}
	})
}

func (s *Server) serveModule(w http.ResponseWriter) {
	s.mu.Lock()
	b := s.last
	s.mu.Unlock()
	if !b.ok() {
		// The overlay shows why; the page's import fails.
		http.Error(w, "the program does not compile; see the terminal or the page overlay", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Write(b.JS)
}

const clientTag = `<script type="module" src="` + ClientPath + `"></script>`

// page serves index.html from the entry file's directory with the dev
// client injected, or a generated page that runs the program and shows
// its trace output.
func (s *Server) page(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if text, err := os.ReadFile(filepath.Join(s.root, "index.html")); err == nil {
		w.Write(inject(text))
		return
	}
	module, _ := json.Marshal(s.module)
	fmt.Fprintf(w, defaultPage, html.EscapeString(s.title), clientTag, module)
}

// inject adds the dev client to an HTML page, before </head> if it has
// one, else at the start.
func inject(page []byte) []byte {
	if i := bytes.Index(bytes.ToLower(page), []byte("</head>")); i >= 0 {
		return append(append(append([]byte{}, page[:i]...), clientTag+"\n"...), page[i:]...)
	}
	return append([]byte(clientTag+"\n"), page...)
}

const defaultPage = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s · runescript dev</title>
%s
<style>
  :root { color-scheme: light dark; }
  body { margin: 16px; font: 14px/1.5 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
</style>
</head>
<body>
<div id="app"></div>
<pre id="runescript-trace"></pre>
<script type="module">
import { run, RuneScriptError } from %s;
const out = document.getElementById("runescript-trace");
try {
  const code = run({ trace: (line) => { out.textContent += line + "\n"; } });
  out.textContent += "\n(main returned " + code + ")";
} catch (e) {
  if (!(e instanceof RuneScriptError)) throw e;
  window.__runescript.showError("Runtime error", e.message);
}
</script>
</body>
</html>
`

// events streams build events to one browser.
func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	ch := make(chan Event, 1)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	first := s.event()
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.subs, ch)
		s.mu.Unlock()
	}()
	send := func(ev Event) bool {
		data, _ := json.Marshal(ev)
		if _, err := fmt.Fprintf(w, "event: build\ndata: %s\n\n", data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}
	if !send(first) {
		return
	}
	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-ch:
			if !send(ev) {
				return
			}
		case <-keepAlive.C:
			if _, err := io.WriteString(w, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
