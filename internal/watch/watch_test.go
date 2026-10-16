package watch

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func write(t *testing.T, path, text string, mod time.Time) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mod, mod); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotChanges(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	a, b, c := filepath.Join(dir, "a.rsc"), filepath.Join(dir, "b.rsm"), filepath.Join(dir, "c.rsm")
	write(t, a, "1", base)
	write(t, b, "1", base)
	write(t, filepath.Join(dir, ".a.rsc.swp"), "x", base)
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	s := Take([]string{dir, filepath.Join(dir, "missing")})
	if len(s) != 2 {
		t.Fatalf("snapshot %v", s)
	}
	if got := s.Changed(Take([]string{dir})); len(got) != 0 {
		t.Errorf("no change, got %v", got)
	}

	write(t, a, "2", base.Add(time.Second))           // modified
	write(t, c, "new", base)                          // added
	write(t, filepath.Join(dir, "b.rsm~"), "x", base) // editor backup: ignored
	write(t, filepath.Join(dir, "sub", "x.rsm"), "x", base)
	if err := os.Remove(b); err != nil { // removed
		t.Fatal(err)
	}
	got := s.Changed(Take([]string{dir}))
	if strings.Join(got, ",") != strings.Join([]string{a, b, c}, ",") {
		t.Errorf("got %v", got)
	}
}

// A change in size is a change even when the modification time is the
// same (coarse file-system clocks).
func TestSizeChange(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	p := filepath.Join(dir, "a.rsc")
	write(t, p, "1", base)
	s := Take([]string{dir})
	write(t, p, "12", base)
	if got := s.Changed(Take([]string{dir})); len(got) != 1 {
		t.Errorf("got %v", got)
	}
}

func TestDirs(t *testing.T) {
	got := Dirs(filepath.Join("x", "a.rsc"), filepath.Join("x", "b.rsm"), filepath.Join("y", "c.rsm"))
	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Errorf("got %v", got)
	}
}

func TestPoll(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.rsc")
	write(t, p, "1", time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	changes := make(chan []string, 1)
	go Poll(ctx, func() []string { return []string{dir} }, func(files []string) {
		select {
		case changes <- files:
		default:
		}
	})
	time.Sleep(2 * Interval) // let Poll take its first snapshot
	write(t, p, "22", time.Date(2026, 10, 16, 12, 0, 1, 0, time.UTC))
	select {
	case files := <-changes:
		if len(files) != 1 || files[0] != p {
			t.Errorf("got %v", files)
		}
	case <-ctx.Done():
		t.Fatal("no change reported")
	}
}
