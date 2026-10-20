package stdlib

import (
	"errors"
	"io"

	"io/fs"
	"os"
	"path/filepath"

	"github.com/AurielSolaris/runescript/internal/types"
)

// std:fs: files and directories. Relative paths are relative to the
// working directory. Text is read and written as UTF-8; invalid UTF-8
// read from a file becomes U+FFFD (decodeUTF8). readBytes, writeBytes, and
// appendBytes move bytes (Array<uint8>) unchanged.
//
// A failure throws an Error whose message names the operation, the path,
// and one of a few reasons with the same wording on every target and
// platform: "no such file or directory", "already exists", "is a
// directory", "not a directory", "directory not empty", or "permission
// denied" (anything else keeps the host's own wording). The checks behind
// the first five are made here, before calling the host, so that both
// targets agree.

// fsFail throws the Error for a failed operation op ("read", ...) on path.
func fsFail(op, path, reason string) {
	panic(Fault{Type: "Error", Message: "cannot " + op + " " + path + ": " + reason})
}

// fsErr throws the Error for err, a failure of op on path.
func fsErr(op, path string, err error) {
	reason := err.Error()
	var pe *fs.PathError
	var le *os.LinkError
	switch {
	case errors.Is(err, fs.ErrNotExist):
		reason = "no such file or directory"
	case errors.Is(err, fs.ErrExist):
		reason = "already exists"
	case errors.Is(err, fs.ErrPermission):
		reason = "permission denied"
	case errors.As(err, &pe):
		reason = pe.Err.Error()
	case errors.As(err, &le):
		reason = le.Err.Error()
	}
	fsFail(op, path, reason)
}

// statOf returns path's file information, following symbolic links, or nil
// if it does not exist.
func statOf(op, path string) fs.FileInfo {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		fsErr(op, path, err)
	}
	return info
}

// mustFile checks that path is an existing file.
func mustFile(op, path string) fs.FileInfo {
	info := statOf(op, path)
	switch {
	case info == nil:
		fsFail(op, path, "no such file or directory")
	case info.IsDir():
		fsFail(op, path, "is a directory")
	}
	return info
}

// mustDir checks that path is an existing directory.
func mustDir(op, path string) {
	info := statOf(op, path)
	switch {
	case info == nil:
		fsFail(op, path, "no such file or directory")
	case !info.IsDir():
		fsFail(op, path, "not a directory")
	}
}

// notDir checks that path, which is about to be written, is not a
// directory.
func notDir(op, path string) {
	if info := statOf(op, path); info != nil && info.IsDir() {
		fsFail(op, path, "is a directory")
	}
}

// readAll reads the file path. readFile decodes it as UTF-8; readBytes
// returns the bytes as they are.
func readAll(path string) []byte {
	mustFile("read", path)
	data, err := os.ReadFile(path)
	if err != nil {
		fsErr("read", path, err)
	}
	return data
}

// writeAll creates or replaces the file path with data.
func writeAll(path string, data []byte) any {
	notDir("write", path)
	if err := os.WriteFile(path, data, 0o666); err != nil {
		fsErr("write", path, err)
	}
	return nil
}

// appendAll adds data to the end of the file path, creating it if needed.
func appendAll(path string, data []byte) any {
	notDir("append to", path)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		fsErr("append to", path, err)
	}
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		fsErr("append to", path, err)
	}
	return nil
}

// names lists the entries of directory path, sorted by name.
func names(op, path string) []string {
	mustDir(op, path)
	entries, err := os.ReadDir(path) // sorted by name
	if err != nil {
		fsErr(op, path, err)
	}
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name()
	}
	return out
}

// walk appends the paths under dir, relative to root and joined with the
// platform's separator: each entry in name order, a directory followed by
// its own contents. Symbolic links are listed but not followed.
func walk(root, rel string, out []string) []string {
	entries, err := os.ReadDir(filepath.Join(root, rel))
	if err != nil {
		fsErr("walk", filepath.Join(root, rel), err)
	}
	for _, e := range entries {
		p := e.Name()
		if rel != "" {
			p = rel + string(filepath.Separator) + p
		}
		out = append(out, p)
		if e.IsDir() {
			out = walk(root, p, out)
		}
	}
	return out
}

func init() {
	S, B, V := types.StringT, types.BoolT, types.VoidT
	SA, BA := types.NewArray(S), types.NewArray(types.Uint8)
	p := &Package{Path: "fs", Qualifier: "fs"}
	p.add(
		fn("readFile", S, func(a []any) any { return decodeUTF8(readAll(str(a[0]))) }, S),
		fn("writeFile", V, func(a []any) any { return writeAll(str(a[0]), []byte(str(a[1]))) }, S, S),
		fn("appendFile", V, func(a []any) any { return appendAll(str(a[0]), []byte(str(a[1]))) }, S, S),
		fn("readBytes", BA, func(a []any) any { return readAll(str(a[0])) }, S),
		fn("writeBytes", V, func(a []any) any { return writeAll(str(a[0]), a[1].([]byte)) }, S, BA),
		fn("appendBytes", V, func(a []any) any { return appendAll(str(a[0]), a[1].([]byte)) }, S, BA),
		fn("exists", B, func(a []any) any { return statOf("check", str(a[0])) != nil }, S),
		fn("isFile", B, func(a []any) any {
			info := statOf("check", str(a[0]))
			return info != nil && info.Mode().IsRegular()
		}, S),
		fn("isDir", B, func(a []any) any {
			info := statOf("check", str(a[0]))
			return info != nil && info.IsDir()
		}, S),
		fn("size", types.Int64, func(a []any) any { return mustFile("measure", str(a[0])).Size() }, S),
		fn("copyFile", V, func(a []any) any {
			from, to := str(a[0]), str(a[1])
			info := mustFile("copy", from)
			notDir("copy to", to)
			src, err := os.Open(from)
			if err != nil {
				fsErr("copy", from, err)
			}
			defer src.Close()
			dst, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
			if err != nil {
				fsErr("copy to", to, err)
			}
			_, err = io.Copy(dst, src)
			if cerr := dst.Close(); err == nil {
				err = cerr
			}
			if err != nil {
				fsErr("copy to", to, err)
			}
			return nil
		}, S, S),
		fn("rename", V, func(a []any) any {
			from, to := str(a[0]), str(a[1])
			if _, err := os.Lstat(from); err != nil {
				fsErr("rename", from, err)
			}
			if err := os.Rename(from, to); err != nil {
				fsErr("rename", from, err)
			}
			return nil
		}, S, S),
		fn("remove", V, func(a []any) any {
			path := str(a[0])
			info, err := os.Lstat(path)
			if err != nil {
				fsErr("remove", path, err)
			}
			if info.IsDir() {
				if entries, err := os.ReadDir(path); err == nil && len(entries) > 0 {
					fsFail("remove", path, "directory not empty")
				}
			}
			if err := os.Remove(path); err != nil {
				fsErr("remove", path, err)
			}
			return nil
		}, S),
		fn("removeAll", V, func(a []any) any {
			path := str(a[0])
			if err := os.RemoveAll(path); err != nil {
				fsErr("remove", path, err)
			}
			return nil
		}, S),
		fn("mkdir", V, func(a []any) any {
			path := str(a[0])
			if statOf("create", path) != nil {
				fsFail("create", path, "already exists")
			}
			if err := os.Mkdir(path, 0o777); err != nil {
				fsErr("create", path, err)
			}
			return nil
		}, S),
		fn("mkdirAll", V, func(a []any) any {
			path := str(a[0])
			if info := statOf("create", path); info != nil && !info.IsDir() {
				fsFail("create", path, "already exists")
			}
			if err := os.MkdirAll(path, 0o777); err != nil {
				fsErr("create", path, err)
			}
			return nil
		}, S),
		fn("readDir", SA, func(a []any) any { return names("list", str(a[0])) }, S),
		fn("walk", SA, func(a []any) any {
			path := str(a[0])
			mustDir("walk", path)
			return walk(path, "", []string{})
		}, S),
		fn("tempDir", S, func(a []any) any { return os.TempDir() }),
	)
	register(p)
}
