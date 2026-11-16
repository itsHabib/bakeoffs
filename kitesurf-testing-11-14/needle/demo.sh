#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
export KITESURF_ENDPOINT="${KITESURF_ENDPOINT:-ws://127.0.0.1:18892/devtools/browser}"
export KITESURF_REVISION="${KITESURF_REVISION:-4a2903813a22b08d4f3cc64a4b35339d0293a5d1}"
mkdir -p artifacts
PORT=8892 node src/server.mjs > artifacts/server.log 2>&1 &
server_pid=$!
trap 'kill "$server_pid" 2>/dev/null || true' EXIT
for i in {1..30}; do
  if ! kill -0 "$server_pid" 2>/dev/null; then cat artifacts/server.log; exit 1; fi
  if curl -fs http://127.0.0.1:8892/health >/dev/null; then break; fi
  sleep 0.1
done
npm test
npm run check
for workload in search cart; do
  echo "Reducing $workload"
  node src/cli.mjs reduce "examples/$workload.json" "artifacts/$workload"
  node src/cli.mjs replay "artifacts/$workload/minimal.json" > "artifacts/$workload/replay.json"
  EXPECT_STATUS=pass node src/cli.mjs replay "examples/$workload-fixed.json" > "artifacts/$workload/fixed-control.json"
done
echo 'Done. Read artifacts/search/report.md and artifacts/cart/report.md.'
