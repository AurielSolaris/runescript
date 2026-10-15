package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The installers (install.sh, install.ps1) are tested in scratch
// locations: a temporary install directory, a temporary HOME for the
// shell startup files, and a scratch registry key instead of the user's
// environment. They build the CLI, so they are skipped with -short.

func installerSkip(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("builds the CLI; skipped with -short")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func runCmd(t *testing.T, cmd *exec.Cmd) (string, int) {
	t.Helper()
	out, err := cmd.CombinedOutput()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return string(out), code
}

func withEnv(cmd *exec.Cmd, kv ...string) *exec.Cmd {
	cmd.Env = append(os.Environ(), kv...)
	return cmd
}

func TestInstallSh(t *testing.T) {
	installerSkip(t)
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not on PATH")
	}
	home := t.TempDir()
	script := filepath.Join(repoRoot(t), "install.sh")
	run := func(args ...string) (string, int) {
		cmd := withEnv(exec.Command(sh, append([]string{script}, args...)...), "HOME="+home, "SHELL=/bin/bash")
		return runCmd(t, cmd)
	}

	out, code := run()
	if code != 0 || !strings.Contains(out, "RuneScript is installed.") {
		t.Fatalf("install: exit %d\n%s", code, out)
	}
	bin := filepath.Join(home, ".runescript", "bin", "runescript")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if v, code := runCmd(t, exec.Command(bin, "version")); code != 0 || !strings.HasPrefix(v, "runescript ") {
		t.Fatalf("installed CLI: exit %d, %q", code, v)
	}
	// Installing again doesn't add a second PATH entry.
	if out, code := run(); code != 0 {
		t.Fatalf("reinstall: exit %d\n%s", code, out)
	}
	for _, rc := range []string{".profile", ".bashrc"} {
		b, err := os.ReadFile(filepath.Join(home, rc))
		if err != nil {
			t.Fatal(err)
		}
		if n := strings.Count(string(b), "# >>> runescript >>>"); n != 1 {
			t.Errorf("%s has %d runescript blocks:\n%s", rc, n, b)
		}
	}

	if out, code := run("--uninstall"); code != 0 {
		t.Fatalf("uninstall: exit %d\n%s", code, out)
	}
	if _, err := os.Stat(bin); !os.IsNotExist(err) {
		t.Errorf("binary still installed: %v", err)
	}
	if b, _ := os.ReadFile(filepath.Join(home, ".profile")); strings.Contains(string(b), "runescript") {
		t.Errorf(".profile still mentions runescript:\n%s", b)
	}
}

func TestInstallShRejectsBadGo(t *testing.T) {
	installerSkip(t)
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not on PATH")
	}
	dir := t.TempDir()
	old := filepath.Join(dir, "oldgo")
	if err := os.WriteFile(old, []byte("#!/bin/sh\necho 'go version go1.24.9 linux/amd64'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(repoRoot(t), "install.sh")
	for _, c := range []struct{ goPath, want string }{
		{filepath.Join(dir, "missing"), "not an executable file"},
		{old, "Go 1.24.9 is too old: RuneScript needs Go 1.25 or newer"},
	} {
		cmd := withEnv(exec.Command(sh, script, "--prefix", filepath.Join(dir, "rs"), "--no-modify-path"),
			"HOME="+dir, "RUNESCRIPT_GO="+filepath.ToSlash(c.goPath))
		out, code := runCmd(t, cmd)
		if code != 1 || !strings.Contains(out, c.want) {
			t.Errorf("RUNESCRIPT_GO=%s: exit %d, want 1 and %q\n%s", c.goPath, code, c.want, out)
		}
	}
}

func powershell(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("install.ps1 is for Windows")
	}
	ps, err := exec.LookPath("powershell")
	if err != nil {
		t.Skip("powershell is not on PATH")
	}
	return ps
}

func TestInstallPs1(t *testing.T) {
	installerSkip(t)
	ps := powershell(t)
	dir := filepath.Join(t.TempDir(), "RuneScript")
	key := fmt.Sprintf(`HKCU:\Software\RuneScriptInstallerTest%d`, time.Now().UnixNano())
	t.Cleanup(func() {
		exec.Command(ps, "-NoProfile", "-Command", "Remove-Item -LiteralPath '"+key+"' -Recurse -Force -ErrorAction SilentlyContinue").Run()
	})
	script := filepath.Join(repoRoot(t), "install.ps1")
	run := func(args ...string) (string, int) {
		base := []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script, "-InstallDir", dir, "-EnvKey", key}
		return runCmd(t, exec.Command(ps, append(base, args...)...))
	}
	userPath := func() string {
		out, _ := runCmd(t, exec.Command(ps, "-NoProfile", "-Command",
			"(Get-Item -LiteralPath '"+key+"').GetValue('Path', '', 'DoNotExpandEnvironmentNames')"))
		return strings.TrimSpace(out)
	}
	bin := filepath.Join(dir, "bin")

	out, code := run()
	if code != 0 || !strings.Contains(out, "RuneScript is installed.") {
		t.Fatalf("install: exit %d\n%s", code, out)
	}
	if v, code := runCmd(t, exec.Command(filepath.Join(bin, "runescript.exe"), "version")); code != 0 || !strings.HasPrefix(v, "runescript ") {
		t.Fatalf("installed CLI: exit %d, %q", code, v)
	}
	if p := userPath(); !strings.EqualFold(p, bin) {
		t.Errorf("user Path = %q, want %q", p, bin)
	}
	if out, code := run(); code != 0 || !strings.Contains(out, "already on PATH") {
		t.Errorf("reinstall: exit %d\n%s", code, out)
	}
	if p := userPath(); !strings.EqualFold(p, bin) {
		t.Errorf("after reinstall, user Path = %q", p)
	}

	if out, code := run("-Uninstall"); code != 0 {
		t.Fatalf("uninstall: exit %d\n%s", code, out)
	}
	if p := userPath(); p != "" {
		t.Errorf("after uninstall, user Path = %q", p)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("install directory still exists: %v", err)
	}
}

func TestInstallPs1RejectsBadGo(t *testing.T) {
	installerSkip(t)
	ps := powershell(t)
	dir := t.TempDir()
	old := filepath.Join(dir, "oldgo.cmd")
	broken := filepath.Join(dir, "badgo.cmd")
	if err := os.WriteFile(old, []byte("@echo go version go1.21.3 windows/amd64\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(broken, []byte("@echo boom 1>&2\r\n@exit /b 3\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(repoRoot(t), "install.ps1")
	args := []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script, "-InstallDir", filepath.Join(dir, "rs"), "-NoModifyPath"}
	for _, c := range []struct{ goPath, want string }{
		{filepath.Join(dir, "missing.exe"), "which is not a file"},
		{old, "Go 1.21.3 is too old: RuneScript needs Go 1.25 or newer"},
		{broken, "The Go installation looks broken"},
	} {
		out, code := runCmd(t, withEnv(exec.Command(ps, args...), "RUNESCRIPT_GO="+c.goPath))
		if code != 1 || !strings.Contains(out, c.want) {
			t.Errorf("RUNESCRIPT_GO=%s: exit %d, want 1 and %q\n%s", c.goPath, code, c.want, out)
		}
	}
	// No Go anywhere: only Windows' own directories on PATH.
	sysroot := os.Getenv("SystemRoot")
	cmd := exec.Command(ps, args...)
	cmd.Env = []string{"PATH=" + filepath.Join(sysroot, "System32") + ";" + filepath.Join(sysroot, `System32\WindowsPowerShell\v1.0`),
		"SystemRoot=" + sysroot, "LOCALAPPDATA=" + dir, "USERPROFILE=" + dir, "ProgramFiles=" + dir, "TEMP=" + dir, "TMP=" + dir}
	if out, code := runCmd(t, cmd); code != 1 || !strings.Contains(out, "Go was not found") {
		t.Errorf("without Go: exit %d\n%s", code, out)
	}
}
