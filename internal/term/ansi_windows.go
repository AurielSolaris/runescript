package term

import (
	"os"
	"syscall"
)

const enableVirtualTerminalProcessing = 0x0004

var setConsoleMode = syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode")

// enableANSI switches on ANSI escape sequences in the Windows console f.
// It reports false if f is not a console that accepts them (older
// consoles), so plain text is written instead.
func enableANSI(f *os.File) bool {
	h := syscall.Handle(f.Fd())
	var mode uint32
	if err := syscall.GetConsoleMode(h, &mode); err != nil {
		return false
	}
	if mode&enableVirtualTerminalProcessing != 0 {
		return true
	}
	ok, _, _ := setConsoleMode.Call(uintptr(h), uintptr(mode|enableVirtualTerminalProcessing))
	return ok != 0
}
