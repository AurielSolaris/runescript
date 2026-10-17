package stdlib

import (
	"fmt"
	"io"
	"strings"

	"github.com/AurielSolaris/runescript/internal/types"
)

// std:io: standard output without a newline, colored error and warning
// lines on standard error, and standard input. trace stays the everyday
// way to print a line.
//
// Input is read as bytes: readByte returns them one at a time, and
// readLine and readAll decode UTF-8, replacing invalid bytes with U+FFFD
// the way the JavaScript runtime's TextDecoder does (decodeUTF8).

// ANSI colors of io.error and io.warn lines.
const (
	red    = "\x1b[31m"
	yellow = "\x1b[33m"
	reset  = "\x1b[0m"
)

func init() {
	S, I, B, V := types.StringT, types.Int, types.BoolT, types.VoidT
	p := &Package{Path: "io", Qualifier: "io"}
	p.add(
		envFn("print", V, func(env *Env, a []any) any {
			io.WriteString(env.Stdout, str(a[0]))
			return nil
		}, S),
		envFn("error", V, func(env *Env, a []any) any {
			colorLine(env, red, str(a[0]))
			return nil
		}, S),
		envFn("warn", V, func(env *Env, a []any) any {
			colorLine(env, yellow, str(a[0]))
			return nil
		}, S),
		envFn("readLine", S, func(env *Env, a []any) any { return readLine(env) }),
		envFn("readAll", S, func(env *Env, a []any) any {
			data, err := io.ReadAll(env.Stdin)
			if err != nil {
				panic(Fault{Type: "Error", Message: "cannot read standard input: " + err.Error()})
			}
			return decodeUTF8(data)
		}),
		envFn("readByte", I, func(env *Env, a []any) any {
			c, err := env.Stdin.ReadByte()
			if err != nil {
				return int64(-1)
			}
			return int64(c)
		}),
		envFn("atEnd", B, func(env *Env, a []any) any {
			_, err := env.Stdin.Peek(1)
			return err != nil
		}),
	)
	register(p)
}

// colorLine writes text and a newline to standard error, colored when the
// environment allows it.
func colorLine(env *Env, color, text string) {
	if env.Color {
		fmt.Fprint(env.Stderr, color+text+reset+"\n")
		return
	}
	fmt.Fprint(env.Stderr, text+"\n")
}

// readLine reads the next line of standard input without its "\n" or
// "\r\n". At the end of the input it returns what is left, which is ""
// once everything has been read (atEnd tells the two apart).
func readLine(env *Env) string {
	line, err := env.Stdin.ReadBytes('\n')
	if err != nil && err != io.EOF {
		panic(Fault{Type: "Error", Message: "cannot read standard input: " + err.Error()})
	}
	s := decodeUTF8(line)
	s = strings.TrimSuffix(s, "\n")
	return strings.TrimSuffix(s, "\r")
}
