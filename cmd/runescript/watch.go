package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	goruntime "runtime"
	"strconv"

	"github.com/AurielSolaris/runescript/internal/devserver"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/driver"
	"github.com/AurielSolaris/runescript/internal/jsgen"
	"github.com/AurielSolaris/runescript/internal/watch"
)

// watchRun implements `runescript run --watch`: it runs the program, then
// runs it again whenever one of its files changes, until ctx is done. A
// run that is still going when a file changes finishes first.
func watchRun(ctx context.Context, in driver.Input, stdout, stderr io.Writer) int {
	var files []string
	in.Files = func(paths []string) { files = paths }
	entry, err := filepath.Abs(in.Path)
	if err != nil {
		entry = in.Path
	}
	once := func() {
		code := runOnce(in, stdout, stderr)
		fmt.Fprintf(stderr, "runescript: exit status %d; watching for changes (Ctrl+C to stop)\n", code)
	}
	once()
	dirs := func() []string { return watch.Dirs(append([]string{entry}, files...)...) }
	watch.Poll(ctx, dirs, func(changed []string) {
		fmt.Fprintf(stderr, "\nrunescript: %s changed; running again\n", display(changed[0], in.Dir))
		once()
	})
	return 0
}

// display shows path relative to dir when it is inside it.
func display(path, dir string) string {
	if rel, err := filepath.Rel(dir, path); err == nil && filepath.IsLocal(rel) {
		return filepath.ToSlash(rel)
	}
	return path
}

// dev implements `runescript dev`.
func dev(args []string, stderr io.Writer) int {
	fs, policy := flags("dev", stderr)
	port := fs.Int("port", 5173, "")
	open := fs.Bool("open", false, "")
	if err := fs.Parse(args); err != nil || fs.NArg() != 1 || policy.err != nil {
		usageError(stderr, policy.err)
		return exitUsage
	}
	path := fs.Arg(0)
	explicitPort := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "port" {
			explicitPort = true
		}
	})
	build := func() devserver.Build {
		var b devserver.Build
		in, err := input(path, io.Discard) // conversion notes go to the build log below
		if err != nil {
			b.Err = err
			return b
		}
		p := policy.WarningPolicy
		p.Report = func(ws []*diagnostics.Diagnostic) { b.Warnings = ws }
		in.Warnings = p
		in.Files = func(paths []string) { b.Files = paths }
		b.JS, b.Errors, b.Err = driver.BuildJSInput(in, jsgen.Options{Version: Version})
		return b
	}
	s := devserver.New(path, build, stderr)
	s.Rebuild()

	ln, err := listen(*port, explicitPort)
	if err != nil {
		fmt.Fprintf(stderr, "runescript dev: %v\n", err)
		return exitUsage
	}
	url := "http://" + ln.Addr().String() + "/"
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	srv := &http.Server{Handler: s.Handler()}
	go watch.Poll(ctx, s.Dirs, s.Changed)
	go func() {
		<-ctx.Done()
		srv.Close() // also ends the browsers' event streams
	}()
	fmt.Fprintf(stderr, "runescript dev: serving %s at %s (Ctrl+C to stop)\n", path, url)
	if *open {
		openBrowser(url, stderr)
	}
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintf(stderr, "runescript dev: %v\n", err)
		return exitUsage
	}
	return 0
}

// listen listens on the loopback interface. Without an explicit port, the
// next free port after the default is used, as Vite does.
func listen(port int, explicit bool) (net.Listener, error) {
	tries := 20
	if explicit {
		tries = 1
	}
	var err error
	for i := 0; i < tries; i++ {
		var ln net.Listener
		if ln, err = net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port+i))); err == nil {
			return ln, nil
		}
	}
	return nil, err
}

// openBrowser opens url in the default browser.
func openBrowser(url string, stderr io.Writer) {
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(stderr, "runescript dev: could not open a browser (%v); open %s yourself\n", err, url)
	}
}
