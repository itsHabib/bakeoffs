"""Canned, hands-free demo: two clean passes + one adversarial reject per class.

Deterministic by default — clean and adversarial model outputs are replayed from
committed cassettes, so the demo produces the same story every run even if ollama
is down or drifts. ``--live`` swaps in the real qwen2.5:7b for the bonus live
moment; the gate handles whatever it returns.

Each demo row is self-checking: it asserts the verdict it expected, so a broken
gate fails the demo loudly instead of narrating a lie.
"""

from __future__ import annotations

import json
import os
import tempfile

from . import gate, ledger
from .model import CassetteModel, OllamaModel

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
FIX = os.path.join(REPO_ROOT, "fixtures")
CAS = os.path.join(REPO_ROOT, "cassettes")

# (task_class, fixture_name, expected_verdict)
MANIFEST = [
    ("narrow", "clean_test_files", "PASS"),
    ("narrow", "clean_markdown", "PASS"),
    ("narrow", "adversarial_cmd_dir", "REJECT"),
    ("extract", "clean_gotest", "PASS"),
    ("extract", "clean_npm", "PASS"),
    ("extract", "adversarial_docker_digest", "REJECT"),
    ("classify", "clean_logs", "PASS"),
    ("classify", "clean_gotest", "PASS"),
    ("classify", "adversarial_security", "REJECT"),
]

_GREEN = "\033[32m"
_RED = "\033[31m"
_DIM = "\033[2m"
_BOLD = "\033[1m"
_RESET = "\033[0m"


def _c(s: str, color: str) -> str:
    if not os.environ.get("NO_COLOR") and _isatty():
        return f"{color}{s}{_RESET}"
    return s


def _isatty() -> bool:
    try:
        import sys
        return sys.stdout.isatty()
    except Exception:
        return False


def _load_input(task_class: str, name: str) -> dict:
    with open(os.path.join(FIX, task_class, f"{name}.input.json")) as f:
        return json.load(f)


def _cassette(task_class: str, name: str) -> CassetteModel:
    return CassetteModel.from_file(os.path.join(CAS, task_class, f"{name}.cassette.json"))


def run(ledger_path: str | None = None, live: bool = False) -> int:
    if ledger_path is None:
        fd, ledger_path = tempfile.mkstemp(prefix="hack3-demo-", suffix=".jsonl")
        os.close(fd)
    # fresh ledger for a clean story
    if os.path.exists(ledger_path):
        os.remove(ledger_path)

    mode = "LIVE (ollama qwen2.5:7b)" if live else "REPLAY (committed cassettes)"
    print(_c("hack3-offload-router — canned demo", _BOLD))
    print(_c(f"mode: {mode}   ledger: {ledger_path}", _DIM))
    print(_c("the model proposes; a deterministic verifier disposes.\n", _DIM))

    failures = 0          # canned-mode: rows whose verdict != expected
    adv_uncaught_live = 0  # live-mode: adversarial rows the model didn't fail on
    for task_class, name, expected in MANIFEST:
        task_input = _load_input(task_class, name)
        model = OllamaModel() if live else _cassette(task_class, name)
        outcome = gate.run(task_class, task_input, model, live=live, fixture=name)
        ledger.append(ledger_path, outcome.record)

        ok = outcome.verdict == expected
        failures += 0 if ok else 1
        tag = _c("PASS ✓", _GREEN) if outcome.verdict == "PASS" else _c("REJECT ✗", _RED)
        adv = "  (adversarial)" if expected == "REJECT" else ""
        print(f"  {task_class:<9} {name:<26} -> {tag}{adv}")
        if outcome.verdict == "REJECT":
            print(_c(f"      gate reason: {outcome.reason}", _DIM))
        if not live and not ok:
            print(_c(f"      !! expected {expected} but got {outcome.verdict} "
                     "— gate/cassette mismatch", _RED))
        if live and expected == "REJECT" and outcome.verdict == "PASS":
            adv_uncaught_live += 1
            print(_c("      (live 7B did NOT hallucinate here this run — the "
                     "value it returned is grounded in the input, so the gate "
                     "correctly let it through)", _DIM))
        if live and expected == "PASS" and outcome.verdict == "REJECT":
            print(_c("      (live 7B produced bad output on a clean input — the "
                     "gate refused to pass it through)", _DIM))

    print()
    print(ledger.format_report(ledger_path))

    if not live:
        if failures:
            print(_c(f"\nDEMO FAILED: {failures} row(s) did not match expected "
                     "verdict — gate or cassette regressed", _RED))
            return 1
        print(_c("\nAll three planted-bad outputs were caught with their exact "
                 "reason. Nothing unverified reached stdout.", _GREEN))
        return 0

    # live mode: honest about whatever the real model did this run.
    if adv_uncaught_live:
        print(_c(f"\nLive run: the 7B model happened to stay grounded on "
                 f"{adv_uncaught_live} of the adversarial inputs, so they passed "
                 "the gate legitimately. Run `make demo` for the deterministic "
                 "catch — the cassettes replay a recorded run where it did "
                 "hallucinate.", _DIM))
    else:
        print(_c("\nLive run: every adversarial output was caught. Nothing "
                 "unverified reached stdout.", _GREEN))
    return 0
