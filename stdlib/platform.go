package stdlib

import (
	"os"
	"runtime"
)

const isWindows = runtime.GOOS == "windows"

// lookupEnv is os.LookupEnv; it is a variable for tests.
var lookupEnv = os.LookupEnv
