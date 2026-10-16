package devserver

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

// fake is a build function whose result the test controls.
type fake struct {
	result Build
	builds int
}

func (f *fake) build() Build { f.builds++; return f.result }

func setup(t *testing.T) (*Server, *fake, *httptest.Server, string) {
	t.Helper()
	dir := t.TempDir()
	entry := filepath.Join(dir, "app.rsc")
	if err := os.WriteFile(entry, []byte("package main;"), 0o644); err != nil {
		t.Fatal(err)
	}
	f := &fake{result: Build{JS: []byte("export const run = () => 0;"), Files: []string{entry}}}
	s := New(entry, f.build, io.Discard)
	s.Rebuild()
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return s, f, ts, dir
}

func get(t *testing.T, url string) (int, string, http.Header) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body), resp.Header
}

func TestDefaultPageAndModule(t *testing.T) {
	s, _, ts, _ := setup(t)
	if s.Module() != "/app.js" {
		t.Errorf("module path %s", s.Module())
	}
	code, page, h := get(t, ts.URL+"/")
	if code != 200 || !strings.Contains(page, clientTag) || !strings.Contains(page, `from "/app.js"`) ||
		!strings.Contains(page, `<div id="app"></div>`) || !strings.Contains(page, "<title>app.rsc · runescript dev</title>") ||
		h.Get("Cache-Control") != "no-store" {
		t.Errorf("page %d %s\n%s", code, h, page)
	}
	code, js, h := get(t, ts.URL+"/app.js")
	if code != 200 || js != "export const run = () => 0;" || !strings.HasPrefix(h.Get("Content-Type"), "text/javascript") {
		t.Errorf("module %d %q %s", code, js, h)
	}
	if code, js, _ := get(t, ts.URL+ClientPath); code != 200 || !strings.Contains(js, "EventSource") {
		t.Errorf("client %d", code)
	}
}

func TestUserPage(t *testing.T) {
	_, _, ts, dir := setup(t)
	page := "<!doctype html><HTML><Head><title>Mine</title></Head><body><div id=app></div></body></HTML>"
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "style.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, got, _ := get(t, ts.URL+"/")
	want := "<!doctype html><HTML><Head><title>Mine</title>" + clientTag + "\n</Head><body><div id=app></div></body></HTML>"
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	if code, css, _ := get(t, ts.URL+"/style.css"); code != 200 || css != "body{}" {
		t.Errorf("static file %d %q", code, css)
	}
	if got := string(inject([]byte("<p>no head</p>"))); got != clientTag+"\n<p>no head</p>" {
		t.Errorf("no head: %s", got)
	}
}

func TestFilesOutsideTheDirectoryAreNotServed(t *testing.T) {
	_, _, ts, dir := setup(t)
	secret := filepath.Join(filepath.Dir(dir), "secret.txt")
	if err := os.WriteFile(secret, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(secret)
	for _, p := range []string{"/../secret.txt", "/..%2fsecret.txt", "/%2e%2e/secret.txt"} {
		if code, body, _ := get(t, ts.URL+p); strings.Contains(body, "secret") && code == 200 {
			t.Errorf("%s served the parent directory", p)
		}
	}
}

func TestFailedBuild(t *testing.T) {
	s, f, ts, _ := setup(t)
	d := diagnostics.New(diagnostics.TypeMismatch, diagnostics.Span{}, "cannot assign")
	f.result = Build{Errors: []*diagnostics.Diagnostic{d}}
	s.Rebuild()
	if code, _, _ := get(t, ts.URL+"/app.js"); code != http.StatusServiceUnavailable {
		t.Errorf("module served after a failed build: %d", code)
	}
	ev := s.event()
	if ev.OK || len(ev.Errors) != 1 || !strings.Contains(ev.Errors[0], "error ASX-TYPE-001: cannot assign") || ev.Version != 2 {
		t.Errorf("event %+v", ev)
	}
}

// readEvent reads the next build event from an event stream.
func readEvent(t *testing.T, r *bufio.Reader) Event {
	t.Helper()
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if data, ok := strings.CutPrefix(strings.TrimRight(line, "\r\n"), "data: "); ok {
			var ev Event
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				t.Fatal(err)
			}
			return ev
		}
	}
}

func TestEvents(t *testing.T) {
	s, f, ts, dir := setup(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL+EventsPath, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("content type %s", ct)
	}
	r := bufio.NewReader(resp.Body)
	first := readEvent(t, r)
	if !first.OK || first.Version != 1 || first.Server == "" {
		t.Errorf("first event %+v", first)
	}

	// A RuneScript file rebuilds; anything else only reloads.
	w := diagnostics.NewWarning(diagnostics.WarnUnusedVar, diagnostics.Span{}, "variable 'x' is never used")
	f.result.Warnings = []*diagnostics.Diagnostic{w}
	s.Changed([]string{filepath.Join(dir, "lib.rsm")})
	ev := readEvent(t, r)
	if ev.Version != 2 || !ev.OK || len(ev.Warnings) != 1 || f.builds != 2 || ev.Server != first.Server {
		t.Errorf("after a source change: %+v (builds %d)", ev, f.builds)
	}
	s.Changed([]string{filepath.Join(dir, "style.css")})
	if ev := readEvent(t, r); ev.Version != 3 || f.builds != 2 {
		t.Errorf("after a style change: %+v (builds %d)", ev, f.builds)
	}
}

func TestDirs(t *testing.T) {
	s, f, _, dir := setup(t)
	lib := filepath.Join(t.TempDir(), "lib.rsm")
	f.result.Files = append(f.result.Files, lib)
	s.Rebuild()
	got := s.Dirs()
	if len(got) != 2 || got[0] != filepath.Dir(lib) && got[1] != filepath.Dir(lib) {
		t.Errorf("dirs %v (served %s)", got, dir)
	}
}
