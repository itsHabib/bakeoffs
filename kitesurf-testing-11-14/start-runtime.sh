#!/bin/sh
set -eu
PORT="${1:?Usage: start-runtime.sh PORT}"
case "$PORT" in 18891|18892|18893) ;; *) echo 'Supported ports: 18891, 18892, 18893' >&2; exit 2;; esac
RUNTIME="${KITESURF_SOURCE:?Set KITESURF_SOURCE to an absolute path to your built Kitesurf checkout}"
if lsof -nP -iTCP:"$PORT" -sTCP:LISTEN >/dev/null 2>&1; then
  echo "Port $PORT already has a listener; refusing to replace it." >&2
  exit 1
fi
cd "$RUNTIME/packages/engine"
export WRANGLER_SEND_METRICS=false
export DO_NOT_TRACK=1
exec ./node_modules/.bin/wrangler dev --env local \
  -c wrangler.jsonc -c test/fixtures-worker/wrangler.jsonc \
  --port "$PORT" --inspector-port "$((PORT + 1000))" \
  --persist-to "$RUNTIME/.wrangler/hackathon-$PORT"
