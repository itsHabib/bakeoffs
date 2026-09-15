#!/bin/sh
set -eu
cd "$(dirname "$0")"
expected=33094f0b99a5a200fdcfebc87c28b39efab5ef8f
actual=$(git -C ../warrant rev-parse HEAD)
test "$actual" = "$expected" || { echo "Warrant source identity mismatch: $actual" >&2; exit 1; }
test -z "$(git -C ../warrant status --porcelain --untracked-files=normal)" || { echo 'Warrant source must match its clean pinned checkout' >&2; exit 1; }
shasum -a 256 -c SOURCE.sha256 >/dev/null
