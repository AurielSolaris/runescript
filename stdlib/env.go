package stdlib

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/AurielSolaris/runescript/internal/types"
)

// Env is what a running program can reach outside itself: its standard
// streams, command-line arguments, and whether error and warning lines
// may be colored. std:io, std:process, std:fs, and std:path use it; the
// JavaScript runtime gets the same things from run()'s options.
type Env struct {
	Stdout io.Writer // trace and io.print
	Stderr io.Writer // io.error and io.warn
	Stdin  *bufio.Reader
	// Args are the program's command-line arguments, without the program
	// name or the runescript command.
	Args []string
	// Color says whether io.error and io.warn color their lines.
	Color bool
	// Getenv and Getwd stand in for os.Getenv and os.Getwd when set
	// (tests use them).
	Getenv func(string) string
	Getwd  func() (string, error)
	// Windows selects Windows path rules for std:path's platform-dependent
	// functions; it defaults to runtime.GOOS == "windows" (see NewEnv).
	Windows bool
}

// NewEnv is an Env with the process's own streams, arguments, and
// platform.
func NewEnv(stdout, stderr io.Writer, stdin io.Reader, args []string) *Env {
	if stdin == nil {
		stdin = strings.NewReader("")
	}
	return &Env{Stdout: stdout, Stderr: stderr, Stdin: bufio.NewReader(stdin), Args: args, Windows: isWindows}
}

func (e *Env) getenv(name string) string {
	if e.Getenv != nil {
		return e.Getenv(name)
	}
	return os.Getenv(name)
}

func (e *Env) getwd() string {
	get := os.Getwd
	if e.Getwd != nil {
		get = e.Getwd
	}
	dir, err := get()
	if err != nil {
		panic(Fault{Type: "Error", Message: "the working directory is unavailable: " + err.Error()})
	}
	return dir
}

// envFn builds a Func whose implementation needs the Env.
func envFn(name string, result *types.Type, impl func(env *Env, args []any) any, params ...*types.Type) *Func {
	return &Func{Name: name, Params: params, Result: result, EnvImpl: impl, JSName: name}
}

// str is a string argument of a native function.
func str(a any) string { return a.(string) }
