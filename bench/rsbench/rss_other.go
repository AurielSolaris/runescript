//go:build !unix

package main

import "os"

// peakRSS is not available here: on Windows it needs GetProcessMemoryInfo,
// which the standard library doesn't wrap, and calling it directly needs
// unsafe (not approved for this project). The memory column stays empty.
func peakRSS(*os.ProcessState) int64 { return -1 }
