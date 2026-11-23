package main

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
)

func cmdWeb(fs *flag.FlagSet, args []string, w io.Writer) error {
	runs := fs.Int("runs", 9, "number of timed samples per implementation")
	iterations := fs.Int("iterations", 100000, "DOM mounts per sample")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 || *runs < 1 || *iterations < 1 {
		return fmt.Errorf("web needs positive -runs and -iterations and takes no benchmark names")
	}
	tc, err := setup(true)
	if err != nil {
		return err
	}
	defer tc.Close()
	fmt.Fprintln(w, tc.RSVersion)
	cmd := exec.Command(tc.Node, filepath.Join(tc.BenchRoot, "tools", "web.mjs"),
		tc.RuneScript, tc.Work, strconv.Itoa(*runs), strconv.Itoa(*iterations))
	cmd.Dir = tc.Root
	out, err := cmd.CombinedOutput()
	fmt.Fprint(w, string(out))
	return err
}
