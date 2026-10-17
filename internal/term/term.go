// Package term decides whether output to a terminal may be colored, for
// the CLI's diagnostics and for std:io's error and warning lines.
package term

import (
	"fmt"
	"os"
)

// Mode is the --color setting.
type Mode int

const (
	Auto   Mode = iota // color when the stream is a terminal and NO_COLOR is unset
	Always             // always color
	Never              // never color
)

// ParseMode parses "auto", "always", or "never".
func ParseMode(s string) (Mode, error) {
	switch s {
	case "auto":
		return Auto, nil
	case "always":
		return Always, nil
	case "never":
		return Never, nil
	}
	return Auto, fmt.Errorf("--color takes auto, always, or never, not %q", s)
}

// Color reports whether output to f should be colored under mode m. In
// auto mode, that is when f is a terminal, the NO_COLOR environment
// variable (https://no-color.org) is unset or empty, and TERM is not
// "dumb". A nil f is never a terminal. Coloring a Windows console also
// switches on its support for ANSI escape sequences.
func Color(m Mode, f *os.File) bool {
	switch m {
	case Always:
		if f != nil {
			enableANSI(f)
		}
		return true
	case Never:
		return false
	}
	if f == nil || os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" || !IsTerminal(f) {
		return false
	}
	return enableANSI(f)
}

// IsTerminal reports whether f is a terminal (a character device).
func IsTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
