package main

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuffer is a strings.Builder safe for the watch loop's goroutine.
type syncBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestRunWatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prog.rsc")
	lib := filepath.Join(dir, "lib.rsm")
	src := func(word string) string {
		return "package main;\nuse \"./lib.rsm\";\npublic int main() {\n    trace(\"" + word + "\", lib.n());\n    return 0;\n}\n"
	}
	if err := os.WriteFile(path, []byte(src("first")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lib, []byte("package lib;\npublic int n() { return 1; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	in, err := input(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	in.Dir = dir // uses are resolved against the working directory
	ctx, cancel := context.WithCancel(context.Background())
	var stdout, stderr syncBuffer
	done := make(chan int)
	go func() { done <- watchRun(ctx, in, &stdout, &stderr) }()

	waitFor(t, "the first run", func() bool { return strings.Contains(stderr.String(), "watching for changes") })
	if stdout.String() != "first 1\n" {
		t.Errorf("first run: %q", stdout.String())
	}
	later := time.Now().Add(2 * time.Second) // a new modification time, even on coarse clocks
	write := func(p, text string) {
		t.Helper()
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, later, later); err != nil {
			t.Fatal(err)
		}
	}

	// The entry file changes.
	write(path, src("second"))
	waitFor(t, "the second run", func() bool { return strings.Contains(stdout.String(), "second 1") })
	if !strings.Contains(stderr.String(), "prog.rsc changed; running again") {
		t.Errorf("stderr: %s", stderr.String())
	}

	// A used file changes; then a compile error is reported, not fatal.
	write(lib, "package lib;\npublic int n() { return 2; }\n")
	waitFor(t, "the run after the library changed", func() bool { return strings.Contains(stdout.String(), "second 2") })
	later = later.Add(2 * time.Second)
	write(path, "package main;\npublic int main() { return x; }\n")
	waitFor(t, "the compile error", func() bool { return strings.Contains(stderr.String(), "ASX-NAME-001") })
	if !strings.Contains(stderr.String(), "exit status 1; watching") {
		t.Errorf("stderr: %s", stderr.String())
	}

	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Errorf("exit %d", code)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("watch did not stop")
	}
}

func TestListen(t *testing.T) {
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()
	port := busy.Addr().(*net.TCPAddr).Port
	if _, err := listen(port, true); err == nil {
		t.Error("an explicit busy port was accepted")
	}
	ln, err := listen(port, false)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	host, p, _ := net.SplitHostPort(ln.Addr().String())
	if host != "127.0.0.1" || p == strconv.Itoa(port) {
		t.Errorf("listening on %s", ln.Addr())
	}
}

func TestDevUsage(t *testing.T) {
	var stdout, stderr strings.Builder
	if code := cli([]string{"dev"}, &stdout, &stderr); code != exitUsage || !strings.Contains(stderr.String(), "runescript dev") {
		t.Errorf("exit %d: %s", code, stderr.String())
	}
	if code := cli([]string{"check", "--watch", "x.rsc"}, &stdout, &stderr); code != exitUsage {
		t.Errorf("check --watch accepted: %d", code)
	}
}
