package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"runtime"
)

func resultsDir(tc *Toolchain, out string) (string, error) {
	if out == "" {
		out = filepath.Join(tc.BenchRoot, "results")
	}
	return out, os.MkdirAll(out, 0o755)
}

func writeCSV(tc *Toolchain, out, name string, header []string, rows [][]string) (string, error) {
	dir, err := resultsDir(tc, out)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write(header)
	w.WriteAll(rows)
	return path, w.Error()
}

func writeFile(tc *Toolchain, out, name, text string) (string, error) {
	dir, err := resultsDir(tc, out)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	return path, os.WriteFile(path, []byte(text), 0o644)
}

func goos() string   { return runtime.GOOS }
func goarch() string { return runtime.GOARCH }
