#!/bin/sh
# RuneScript installer for Linux and macOS (and other POSIX systems).
#
# Builds the runescript CLI (compiler and runtime) with your Go toolchain,
# installs it in ~/.runescript/bin, and adds that directory to PATH in your
# shell startup files.
#
#   sh install.sh                      # from a RuneScript checkout: builds that source
#   curl -fsSL https://raw.githubusercontent.com/AurielSolaris/runescript/main/install.sh | sh
#                                      # anywhere else: builds the latest version from GitHub
#
# Options (or the environment variables in brackets):
#   --prefix DIR        install into DIR/bin              [RUNESCRIPT_HOME]
#   --version REF       version to build from GitHub: a tag, branch, or commit
#                       (default: latest)                 [RUNESCRIPT_VERSION]
#   --remote            build from GitHub even inside a checkout
#   --no-modify-path    don't change shell startup files
#   --uninstall         remove RuneScript and its PATH entries
#   -h, --help          show this help
#
# Go: the installer needs Go 1.25 or newer (https://go.dev/dl/). It uses the
# `go` on PATH, or the one named by RUNESCRIPT_GO, or a standard location
# such as /usr/local/go/bin/go.

set -eu

MODULE="github.com/AurielSolaris/runescript"
MIN_GO="1.25" # keep in sync with go.mod
MARK_BEGIN="# >>> runescript >>>"
MARK_END="# <<< runescript <<<"

if [ -t 1 ]; then
    bold=$(printf '\033[1m'); red=$(printf '\033[31m'); green=$(printf '\033[32m'); yellow=$(printf '\033[33m'); reset=$(printf '\033[0m')
else
    bold=""; red=""; green=""; yellow=""; reset=""
fi
say()  { printf '%s\n' "$*"; }
step() { printf '%s==>%s %s\n' "$bold" "$reset" "$*"; }
ok()   { printf '%s  ok%s %s\n' "$green" "$reset" "$*"; }
warn() { printf '%swarning:%s %s\n' "$yellow" "$reset" "$*" >&2; }
die()  { printf '%serror:%s %s\n' "$red" "$reset" "$*" >&2; exit 1; }

usage() { sed -n '2,/^$/p' "$0" 2>/dev/null | sed 's/^# \{0,1\}//'; }

# ---- options ----

prefix="${RUNESCRIPT_HOME:-$HOME/.runescript}"
version="${RUNESCRIPT_VERSION:-latest}"
force_remote=0
modify_path=1
uninstall=0
while [ $# -gt 0 ]; do
    case "$1" in
        --prefix) [ $# -ge 2 ] || die "--prefix needs a directory"; prefix=$2; shift 2 ;;
        --prefix=*) prefix=${1#*=}; shift ;;
        --version) [ $# -ge 2 ] || die "--version needs a value"; version=$2; shift 2 ;;
        --version=*) version=${1#*=}; shift ;;
        --remote) force_remote=1; shift ;;
        --no-modify-path) modify_path=0; shift ;;
        --uninstall) uninstall=1; shift ;;
        -h|--help) usage; exit 0 ;;
        *) die "unknown option: $1 (see --help)" ;;
    esac
done
case "$prefix" in
    /*) ;;
    *) prefix="$(pwd)/$prefix" ;;
esac
bindir="$prefix/bin"

# ---- shell startup files ----

# rc_files lists the startup files that set PATH for the user's shells.
rc_files() {
    say "$HOME/.profile"
    [ -f "$HOME/.bash_profile" ] && say "$HOME/.bash_profile"
    if [ -f "$HOME/.bashrc" ] || [ "${SHELL##*/}" = bash ]; then say "$HOME/.bashrc"; fi
    if [ -f "$HOME/.zshrc" ] || [ "${SHELL##*/}" = zsh ]; then say "$HOME/.zshrc"; fi
    return 0
}

remove_block() { # file: removes our block, and reports whether there was one
    [ -f "$1" ] || return 1
    if grep -q "^$MARK_BEGIN\$" "$1"; then
        tmp="$1.runescript.$$"
        awk -v b="$MARK_BEGIN" -v e="$MARK_END" '$0 == b { skip = 1; next } $0 == e { skip = 0; next } !skip' "$1" > "$tmp"
        cat "$tmp" > "$1" && rm -f "$tmp"
        return 0
    fi
    return 1
}

add_block() { # file
    remove_block "$1" || true
    {
        say "$MARK_BEGIN"
        say "export RUNESCRIPT_HOME=\"$prefix\""
        say 'case ":$PATH:" in *":$RUNESCRIPT_HOME/bin:"*) ;; *) export PATH="$RUNESCRIPT_HOME/bin:$PATH" ;; esac'
        say "$MARK_END"
    } >> "$1"
}

fish_file="$HOME/.config/fish/conf.d/runescript.fish"

# ---- uninstall ----

if [ "$uninstall" -eq 1 ]; then
    step "Uninstalling RuneScript from $prefix"
    if [ -d "$prefix" ]; then
        rm -f "$bindir/runescript" "$bindir/runescript.exe"
        rmdir "$bindir" 2>/dev/null || true
        rmdir "$prefix" 2>/dev/null || warn "$prefix is not empty; left in place"
    fi
    for f in "$HOME/.profile" "$HOME/.bash_profile" "$HOME/.bashrc" "$HOME/.zshrc"; do
        if remove_block "$f"; then ok "removed the PATH entry from $f"; fi
    done
    [ -f "$fish_file" ] && rm -f "$fish_file" && ok "removed $fish_file"
    ok "RuneScript is uninstalled. Open a new terminal to update PATH."
    exit 0
fi

# ---- Go ----

step "Checking Go"

if [ -n "${RUNESCRIPT_GO:-}" ] && ! [ -x "$RUNESCRIPT_GO" ]; then
    die "RUNESCRIPT_GO is set to '$RUNESCRIPT_GO', which is not an executable file"
fi

find_go() {
    if [ -n "${RUNESCRIPT_GO:-}" ]; then
        say "$RUNESCRIPT_GO"
        return 0
    fi
    if command -v go >/dev/null 2>&1; then
        command -v go
        return 0
    fi
    for g in /usr/local/go/bin/go /opt/homebrew/bin/go /usr/lib/go/bin/go /snap/bin/go "$HOME/go/bin/go" "$HOME"/sdk/go*/bin/go; do
        if [ -x "$g" ]; then
            say "$g"
            return 0
        fi
    done
    return 1
}

go_missing() {
    die "Go was not found.
RuneScript is built from source, so the installer needs Go $MIN_GO or newer.
Install it from https://go.dev/dl/ (or with your package manager), open a new
terminal, and run the installer again. If Go is installed in an unusual place,
set RUNESCRIPT_GO to the path of its 'go' program."
}

GO=$(find_go) || go_missing
command -v go >/dev/null 2>&1 || [ -n "${RUNESCRIPT_GO:-}" ] || warn "using $GO, which is not on PATH"

gv=$(GOTOOLCHAIN=local "$GO" version 2>&1) || die "'$GO version' failed:
$gv
The Go installation looks broken. Reinstall Go from https://go.dev/dl/."
# "go version go1.25.11 linux/amd64" (or "go version devel go1.26-abc ...")
ver=$(say "$gv" | sed -n 's/.*go\([0-9][0-9]*\.[0-9][0-9]*\)\(\.[0-9][0-9]*\)\{0,1\}.*/\1\2/p' | head -n 1)
[ -n "$ver" ] || die "could not read the Go version from '$gv'"

version_ge() { # a b: is a >= b, for dotted versions
    a=$1; b=$2
    while [ -n "$a$b" ]; do
        x=${a%%.*}; y=${b%%.*}
        x=${x:-0}; y=${y:-0}
        [ "$x" -gt "$y" ] && return 0
        [ "$x" -lt "$y" ] && return 1
        case "$a" in *.*) a=${a#*.} ;; *) a="" ;; esac
        case "$b" in *.*) b=${b#*.} ;; *) b="" ;; esac
    done
    return 0
}

# Inside a checkout, go.mod says which Go it needs.
src=""
script_dir=""
case "$0" in
    */install.sh|install.sh) script_dir=$(cd "$(dirname "$0")" 2>/dev/null && pwd) || script_dir="" ;;
esac
if [ "$force_remote" -eq 0 ] && [ -n "$script_dir" ] && [ -f "$script_dir/go.mod" ] &&
    grep -q "^module $MODULE\$" "$script_dir/go.mod"; then
    src=$script_dir
    need=$(sed -n 's/^go \([0-9][0-9.]*\).*/\1/p' "$src/go.mod" | head -n 1)
    [ -n "$need" ] && MIN_GO=$need
fi

version_ge "$ver" "$MIN_GO" || die "Go $ver is too old: RuneScript needs Go $MIN_GO or newer.
Update Go from https://go.dev/dl/ and run the installer again."

goroot=$(GOTOOLCHAIN=local "$GO" env GOROOT 2>/dev/null) || die "'$GO env GOROOT' failed; the Go installation looks broken"
[ -d "$goroot" ] || die "Go's GOROOT ($goroot) does not exist; reinstall Go from https://go.dev/dl/"
exe=$(GOTOOLCHAIN=local "$GO" env GOEXE 2>/dev/null || true)
ok "Go $ver ($GO)"

# ---- build ----

mkdir -p "$bindir" || die "cannot create $bindir"
target="$bindir/runescript$exe"
# A pure-Go build with this toolchain: no C compiler needed, and no
# toolchain download that would bypass the version check above.
export GOTOOLCHAIN=local CGO_ENABLED=0
if [ -n "$src" ]; then
    step "Building RuneScript from $src"
    (cd "$src" && "$GO" build -trimpath -o "$target" ./cmd/runescript) || die "the build failed (see the messages above)"
else
    command -v git >/dev/null 2>&1 || warn "git is not on PATH; 'go install' may not be able to fetch a branch or commit"
    step "Building RuneScript $version from $MODULE"
    tmpbin=$(mktemp -d 2>/dev/null || mktemp -d -t runescript)
    GOBIN="$tmpbin" "$GO" install -trimpath "$MODULE/cmd/runescript@$version" || {
        rm -rf "$tmpbin"
        die "'go install $MODULE/cmd/runescript@$version' failed (see the messages above).
Check your network connection, or that '$version' is a tag, branch, or commit."
    }
    mv -f "$tmpbin/runescript$exe" "$target"
    rm -rf "$tmpbin"
fi
chmod +x "$target" 2>/dev/null || true
ok "installed $target"

# ---- check ----

step "Checking the installation"
installed=$("$target" version 2>&1) || die "the installed CLI does not run: $installed"
smoke=$(mktemp -d 2>/dev/null || mktemp -d -t runescript)
cat > "$smoke/hello.rsc" <<'EOF'
package main;

public int main() {
    trace("Hello from RuneScript");
    return 0;
}
EOF
out=$(cd "$smoke" && "$target" run hello.rsc 2>&1) || { rm -rf "$smoke"; die "the installed CLI could not run a test program: $out"; }
rm -rf "$smoke"
[ "$out" = "Hello from RuneScript" ] || die "the test program printed '$out'"
ok "$installed"

if command -v node >/dev/null 2>&1; then
    ok "Node.js found: modules from 'runescript build' can run with node"
else
    say "  Node.js not found. It's needed only to run JavaScript from 'runescript build'."
fi

# ---- PATH ----

if [ "$modify_path" -eq 1 ]; then
    step "Adding $bindir to PATH"
    for f in $(rc_files); do
        add_block "$f"
        ok "updated $f"
    done
    if command -v fish >/dev/null 2>&1 || [ -d "$HOME/.config/fish" ]; then
        mkdir -p "$(dirname "$fish_file")"
        {
            say "$MARK_BEGIN"
            say "set -gx RUNESCRIPT_HOME \"$prefix\""
            say "fish_add_path -g \"$bindir\""
            say "$MARK_END"
        } > "$fish_file"
        ok "updated $fish_file"
    fi
fi

other=$(command -v runescript 2>/dev/null || true)
if [ -n "$other" ] && [ "$other" != "$target" ]; then
    warn "another runescript is earlier on PATH: $other"
fi

say ""
say "${green}RuneScript is installed.${reset}"
case ":$PATH:" in
    *":$bindir:"*) say "Run: runescript version" ;;
    *)
        if [ "$modify_path" -eq 1 ]; then
            say "Open a new terminal (or run: export PATH=\"$bindir:\$PATH\"), then: runescript version"
        else
            say "Add $bindir to PATH to run 'runescript' from anywhere."
        fi
        ;;
esac
