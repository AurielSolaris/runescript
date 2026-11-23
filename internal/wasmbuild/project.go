package wasmbuild

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AurielSolaris/runescript/internal/loader"
)

func stageProject(prog *loader.Program, assets []string, dst string) (string, error) {
	file, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	tw := tar.NewWriter(file)
	entry, err := writeProject(tw, prog, assets)
	return entry, errors.Join(err, tw.Close(), file.Close())
}

func writeProject(tw *tar.Writer, prog *loader.Program, assets []string) (string, error) {
	entry, err := projectName(prog.Dir, prog.Entry.Source.Name)
	if err != nil {
		return "", err
	}
	added := map[string]bool{}
	add := func(name string, size int64, content io.Reader) error {
		if added[name] {
			return nil
		}
		added[name] = true
		if err := tw.WriteHeader(&tar.Header{Name: name, Size: size, Mode: 0o644}); err != nil {
			return err
		}
		_, err := io.Copy(tw, content)
		return err
	}
	for _, file := range prog.Files {
		for _, use := range file.Uses {
			if use.Kind == loader.UseFile && filepath.IsAbs(filepath.FromSlash(use.Decl.Path)) {
				return "", fmt.Errorf("use a project-relative path in %q when building for WebAssembly", use.Decl.Path)
			}
		}
		name, err := projectName(prog.Dir, file.Source.Name)
		if err != nil {
			return "", err
		}
		if err := add(name, int64(len(file.Source.Text)), bytes.NewReader(file.Source.Text)); err != nil {
			return "", err
		}
	}
	for _, asset := range assets {
		name, err := projectName(prog.Dir, asset)
		if err != nil {
			return "", err
		}
		if added[name] {
			continue
		}
		path := filepath.Join(prog.Dir, filepath.FromSlash(name))
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("WebAssembly asset %q must be a file", asset)
		}
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		if err := errors.Join(add(name, info.Size(), f), f.Close()); err != nil {
			return "", err
		}
	}
	return entry, nil
}

func projectName(dir, name string) (string, error) {
	path := filepath.FromSlash(name)
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("WebAssembly input %q is outside the project directory %q", name, dir)
	}
	return filepath.ToSlash(rel), nil
}
