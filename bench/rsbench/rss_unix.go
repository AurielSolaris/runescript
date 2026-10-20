//go:build unix

package main

import (
	"os"
	"runtime"
	"syscall"
)

// peakRSS returns the finished process's peak resident memory in bytes.
// Linux reports ru_maxrss in kilobytes, macOS and the BSDs in bytes.
func peakRSS(ps *os.ProcessState) int64 {
	ru, ok := ps.SysUsage().(*syscall.Rusage)
	if !ok {
		return -1
	}
	if runtime.GOOS == "linux" {
		return int64(ru.Maxrss) * 1024
	}
	return int64(ru.Maxrss)
}
