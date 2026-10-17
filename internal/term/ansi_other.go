//go:build !windows

package term

import "os"

// enableANSI reports true: terminals outside Windows accept ANSI escape
// sequences.
func enableANSI(*os.File) bool { return true }
