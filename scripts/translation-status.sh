#!/bin/sh
# Lists translations whose English source has changed since they were
# translated. Each translated file names its source and commit in a comment
# on its first line:  <!-- source: docs/guide/01-getting-started.md @ 4f1c2ab -->
#
# usage: scripts/translation-status.sh
set -eu
cd "$(dirname "$0")/.."

stale=0
for f in $(find localization -name '*.md' ! -path 'localization/README.md' | sort); do
    line=$(head -n 1 "$f")
    src=$(printf '%s\n' "$line" | sed -n 's/^<!-- source: \(.*\) @ \([0-9a-f]*\) -->$/\1/p')
    rev=$(printf '%s\n' "$line" | sed -n 's/^<!-- source: \(.*\) @ \([0-9a-f]*\) -->$/\2/p')
    if [ -z "$src" ] || [ -z "$rev" ]; then
        echo "no source comment: $f"
        stale=1
        continue
    fi
    if ! git cat-file -e "$rev^{commit}" 2>/dev/null; then
        echo "unknown commit $rev: $f"
        stale=1
        continue
    fi
    if [ -n "$(git log --oneline "$rev..HEAD" -- "$src")" ]; then
        echo "out of date: $f ($src changed since $rev)"
        stale=1
    fi
done
[ "$stale" -eq 0 ] && echo "all translations are up to date"
exit "$stale"
