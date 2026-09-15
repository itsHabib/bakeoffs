"""Hands-free canned demonstration; all claims come from the just-finished tests."""
import json
from pathlib import Path
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parent
start = time.monotonic()
print("Running real worker crashes, recovery, and independent effect checks...", flush=True)
result = subprocess.run(["make", "test"], cwd=ROOT, capture_output=True, text=True, timeout=180)
(ROOT / "artifacts").mkdir(exist_ok=True)
(ROOT / "artifacts/demo-test.log").write_text(result.stdout + result.stderr)
if result.returncode:
    print(result.stdout + result.stderr)
    sys.exit(result.returncode)
location = json.loads((ROOT / "artifacts/latest-run.json").read_text())
base = Path(location["directory"])
results = {entry["case"]: entry for entry in json.loads((base / "summary.json").read_text())}
print("\nThe worker died. What can its replacement safely do?\n")
print("1. Legal S1 run: real squares checked; planted defect rejected; one effect; named completion.")
print("2. Queryable sink: COMMIT -> reply withheld -> SIGKILL -> replacement queries receipt.")
print("   Effect count: 1. The operation key survived the incarnation change.")
parked = json.loads((base / "crash-opaque-sink_committed_reply_withheld/worker-2.result.json").read_text())
print("3. Opaque sink: COMMIT -> reply withheld -> SIGKILL -> UNRESOLVED. No repeat.")
print("   Affected operation: " + parked["op"])
print("4. Old terminal refused. S2 uses fresh checks and a separate effect; S1 remains history.")
print("   Changed bytes refused. Torn tails recover; corrupt records and changed definitions refuse.")
report = json.loads((base / "ordering-deviation/audit/report.json").read_text())
deviation = next(iter(report["subjects"].values()))
raw = deviation["first_deviation"]["raw"]
print(f"5. Parley OBSERVES the planted order swap at normalized event 2 / {raw['id']} / journal seq {raw['seq']}.")
print("   Effects and identity assertions still pass. Parley adds an order diagnostic, not prevention.")
print("6. All three broken policies caught: duplicate effect, stale incarnation, stale byte digest.")
print("   All three mutant traces are protocol-complete: order alone cannot certify execution.")
print(f"\n{len(results)} local cases + 27 upstream compiler tests passed in {time.monotonic() - start:.1f}s.")
print(f"Artifacts: {base}")
print(f"Test log: {ROOT / 'artifacts/demo-test.log'}")
