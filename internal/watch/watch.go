// Package watch detects changes to a program's files, for watch mode
// (`runescript run --watch`) and the dev server (`runescript dev`).
//
// It polls: Go's standard library has no file-notification API, and polling
// a few directories a few times a second is cheap. A directory is watched
// non-recursively, so a new file in a package directory is noticed too.
package watch

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Interval is how often Poll looks for changes.
const Interval = 250 * time.Millisecond

type stamp struct {
	mod  time.Time
	size int64
}

// Snapshot is the state of the watched files at one moment.
type Snapshot map[string]stamp

// Take records the regular files directly inside dirs. Hidden files and
// editor backups (name~, .name.swp) are ignored: saving creates and removes
// them, which would trigger spurious rebuilds.
func Take(dirs []string) Snapshot {
	s := Snapshot{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // a missing directory: nothing to watch
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || strings.HasPrefix(name, ".") || strings.HasSuffix(name, "~") {
				continue
			}
			info, err := e.Info()
			if err != nil || !info.Mode().IsRegular() {
				continue
			}
			s[filepath.Join(dir, name)] = stamp{info.ModTime(), info.Size()}
		}
	}
	return s
}

// Changed returns the files that differ between s and t: changed, added,
// or removed, sorted.
func (s Snapshot) Changed(t Snapshot) []string {
	var out []string
	for p, a := range s {
		if b, ok := t[p]; !ok || !a.mod.Equal(b.mod) || a.size != b.size {
			out = append(out, p)
		}
	}
	for p := range t {
		if _, ok := s[p]; !ok {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// Dirs returns the directories holding files, without duplicates, sorted.
func Dirs(files ...string) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range files {
		d := filepath.Dir(f)
		if !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	sort.Strings(out)
	return out
}

// Poll calls changed with the changed files whenever the files in the
// directories dirs() returns change, until ctx is done. dirs is called
// again after each change, since a program may start using files in a
// new directory.
func Poll(ctx context.Context, dirs func() []string, changed func(files []string)) {
	prev := Take(dirs())
	t := time.NewTicker(Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		cur := Take(dirs())
		if files := prev.Changed(cur); len(files) > 0 {
			changed(files)
			cur = Take(dirs()) // the directories may have changed
		}
		prev = cur
	}
}
