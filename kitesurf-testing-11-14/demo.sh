#!/bin/sh
set -eu
cd "$(dirname "$0")"
: "${KITESURF_SOURCE:?Set KITESURF_SOURCE to your built Kitesurf checkout}"
export KITESURF_SOURCE
KITESURF_REVISION=$(git -C "$KITESURF_SOURCE" rev-parse HEAD)
export KITESURF_REVISION
owned_pids=''
cleanup() {
  for pid in $owned_pids; do kill "$pid" 2>/dev/null || true; done
}
trap cleanup EXIT
trap 'exit 130' INT TERM
is_kitesurf() {
  node -e 'fetch("http://127.0.0.1:"+process.argv[1]+"/json/version", {signal:AbortSignal.timeout(1500)}).then(r=>r.json()).then(d=>{if(!/^Kitesurf\//.test(d.Browser))process.exit(1)}).catch(()=>process.exit(1))' "$1"
}
for port in 18891 18892 18893; do
  if is_kitesurf "$port"; then continue; fi
  ./start-runtime.sh "$port" > "runtime-$port.log" 2>&1 &
  owned_pids="$owned_pids $!"
  attempt=0
  until is_kitesurf "$port"; do
    attempt=$((attempt + 1))
    if [ "$attempt" -ge 20 ]; then
      echo "Kitesurf $port failed to start; see runtime-$port.log" >&2
      exit 1
    fi
    sleep 1
  done
done
result=0
for entry in fixture-pruner needle proofgap; do
  (cd "$entry" && ./demo.sh) || result=1
done
exit "$result"
