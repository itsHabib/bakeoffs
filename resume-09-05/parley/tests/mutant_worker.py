"""Deliberately broken policies. Only the test controller invokes this launcher."""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))
import common
import worker

mutant = sys.argv.pop(1)
if mutant == "blind_retry":
    worker.reconcile = lambda directory, mode, intent: None
elif mutant == "old_terminal":
    common.check_incarnation = lambda terminal, current: None
elif mutant == "changed_candidate":
    common.check_candidate_binding = lambda content, source, evidence: None
else:
    raise ValueError(mutant)
sys.exit(worker.main())
