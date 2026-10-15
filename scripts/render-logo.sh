#!/bin/sh
# Rasterize the canonical SVG logos into PNGs using headless Microsoft Edge
# (Chromium). The SVGs in assets/ are the source of truth; never edit the
# PNGs by hand.
#
# usage: scripts/render-logo.sh            (run from the repository root)
#
# Requires Microsoft Edge or Chrome. Override the browser path with BROWSER=...
# Any Chromium-based browser accepts the same flags.
set -eu

BROWSER=${BROWSER:-"/c/Program Files (x86)/Microsoft/Edge/Application/msedge.exe"}

# render <svg> <png> <scale>: the SVGs are 256x256, so the PNG is 256*scale px.
render() {
	svg=$(realpath "$1")
	out=$(realpath -m "$2")
	if command -v cygpath >/dev/null 2>&1; then
		svg=$(cygpath -m "$svg")
		out=$(cygpath -w "$out")
	fi
	"$BROWSER" --headless --disable-gpu --hide-scrollbars \
		--default-background-color=00000000 \
		--window-size=256,256 --force-device-scale-factor="$3" \
		--screenshot="$out" "file:///${svg#/}" 2>/dev/null
	echo "wrote $2"
}

render assets/runescript.svg      assets/runescript.png      2    # 512x512, README
render assets/runescript-icon.svg assets/runescript-icon.png 0.25 # 64x64, app icon
