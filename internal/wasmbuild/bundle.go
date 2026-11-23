// Package wasmbuild packages the Go RuneScript runtime and raylib's
// Emscripten WebAssembly module for a browser.
package wasmbuild

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AurielSolaris/runescript/internal/loader"
)

const module = "github.com/AurielSolaris/runescript"
const raylib = "github.com/gen2brain/raylib-go/raylib"
const raygui = "github.com/gen2brain/raylib-go/raygui"

// Build writes a new web directory. Project paths are relative to the
// directory used for RuneScript imports on the host.
func Build(prog *loader.Program, out string, assets []string) error {
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		if err != nil {
			return err
		}
		return fmt.Errorf("WebAssembly output %q already exists; choose a new directory", out)
	}
	stage, err := os.MkdirTemp(filepath.Dir(out), ".runescript-web-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	stage, err = filepath.Abs(stage)
	if err != nil {
		return err
	}
	entry, err := stageProject(prog, assets, filepath.Join(stage, "project.tar"))
	if err != nil {
		return err
	}
	root, err := sourceRoot()
	if err != nil {
		return err
	}
	if _, err := goTool(root, "build", "-buildvcs=false", "-trimpath", "-ldflags=-s -w",
		"-tags", "runescript_gfx", "-o", filepath.Join(stage, "app.wasm"), "./cmd/runescript"); err != nil {
		return err
	}
	goRoot, err := goTool(root, "env", "GOROOT")
	if err != nil {
		return err
	}
	deps, err := goTool(root, "list", "-m", "-f", "{{.Dir}}", raylib, raygui)
	if err != nil {
		return err
	}
	rlRoot, guiRoot, hasGUI := strings.Cut(deps, "\n")
	if !hasGUI || rlRoot == "" || guiRoot == "" {
		return fmt.Errorf("could not find raylib-go and raygui modules")
	}
	rlRoot, guiRoot = strings.TrimSpace(rlRoot), strings.TrimSpace(guiRoot)
	rlWasm := filepath.Join(rlRoot, "libs", "wasm")
	if err := os.Mkdir(filepath.Join(stage, "licenses"), 0o755); err != nil {
		return err
	}
	for _, asset := range []struct{ name, source string }{
		{"wasm_exec.js", filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js")},
		{"raylib-go.js", filepath.Join(rlWasm, "raylib-go.js")},
		{"raylib.js", filepath.Join(rlWasm, "raylib.js.gz")},
		{"raylib.wasm", filepath.Join(rlWasm, "raylib.wasm.gz")},
		{"licenses/go.txt", filepath.Join(goRoot, "LICENSE")},
		{"licenses/raylib-go.txt", filepath.Join(rlRoot, "LICENSE")},
		{"licenses/raygui.txt", filepath.Join(guiRoot, "LICENSE")},
		{"licenses/runescript.txt", filepath.Join(root, "LICENSE")},
	} {
		if err := copyAsset(filepath.Join(stage, asset.name), asset.source); err != nil {
			return fmt.Errorf("copy %s: %w", asset.name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(stage, "index.html"), []byte(indexHTML), 0o644); err != nil {
		return err
	}
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	main := strings.Replace(mainJS, "ENTRY_FILE", string(entryJSON), 1)
	if err := os.WriteFile(filepath.Join(stage, "main.js"), []byte(main), 0o644); err != nil {
		return err
	}
	return os.Rename(stage, out)
}

func goTool(root string, args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0")
	out, err := cmd.Output()
	if err != nil {
		var failed *exec.ExitError
		if errors.As(err, &failed) {
			return "", fmt.Errorf("go %s: %w\n%s", args[0], err, strings.TrimSpace(string(failed.Stderr)))
		}
		return "", fmt.Errorf("go %s: %w", args[0], err)
	}
	return strings.TrimSpace(string(out)), nil
}

func sourceRoot() (string, error) {
	if explicit := os.Getenv("RUNESCRIPT_SOURCE"); explicit != "" {
		return filepath.Abs(explicit)
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if mod, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
			for _, line := range strings.Split(string(mod), "\n") {
				if strings.TrimSpace(line) == "module "+module {
					return dir, nil
				}
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("WebAssembly builds need the RuneScript source checkout; set RUNESCRIPT_SOURCE to its directory")
}

func copyAsset(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	var reader io.Reader = in
	if strings.HasSuffix(src, ".gz") {
		gzipReader, err := gzip.NewReader(in)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, reader)
	return errors.Join(err, out.Close())
}

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>RuneScript</title>
  <style>
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: #141923; color: #fff; font: 1rem system-ui; }
    canvas { display: block; max-width: 100vw; max-height: 100vh; object-fit: contain; }
    [role="alert"] { max-width: 40rem; padding: 1rem; overflow-wrap: anywhere; }
  </style>
</head>
<body>
  <canvas id="canvas" tabindex="0" aria-label="Game"></canvas>
  <p id="error" role="alert" hidden></p>
  <script src="./wasm_exec.js"></script>
  <script type="module" src="./main.js"></script>
</body>
</html>
`

const mainJS = `import { run } from "./raylib-go.js";

const canvas = document.getElementById("canvas");
globalThis.process.cwd = () => "/";
canvas.addEventListener("pointerdown", () => canvas.focus());

try {
  await run({ canvas, wasm: "./app.wasm", assets: "./project.tar", args: ["run", "/" + ENTRY_FILE] });
} catch (error) {
  const message = document.getElementById("error");
  message.textContent = "Could not start RuneScript: " + error.message;
  message.hidden = false;
  console.error(error);
}
`
