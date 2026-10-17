package stdlib

import (
	"runtime"

	"github.com/AurielSolaris/runescript/internal/types"
)

// std:process: the running program's command-line arguments,
// environment variables, working directory, and platform.
func init() {
	S, B := types.StringT, types.BoolT
	p := &Package{Path: "process", Qualifier: "process"}
	p.add(
		envFn("args", types.NewArray(S), func(env *Env, a []any) any {
			return append([]string{}, env.Args...) // a new array each call
		}),
		envFn("env", S, func(env *Env, a []any) any { return env.getenv(str(a[0])) }, S),
		envFn("hasEnv", B, func(env *Env, a []any) any {
			if env.Getenv != nil {
				return env.Getenv(str(a[0])) != ""
			}
			_, ok := lookupEnv(str(a[0]))
			return ok
		}, S),
		envFn("cwd", S, func(env *Env, a []any) any { return env.getwd() }),
		envFn("platform", S, func(env *Env, a []any) any { return platformName(runtime.GOOS) }),
	)
	register(p)
}

// platformName is process.platform()'s name for a Go GOOS, the same names
// the JavaScript runtime gives Node's process.platform values: "windows",
// "macos", "linux", and the GOOS itself for others ("freebsd", ...).
func platformName(goos string) string {
	switch goos {
	case "darwin":
		return "macos"
	}
	return goos
}
