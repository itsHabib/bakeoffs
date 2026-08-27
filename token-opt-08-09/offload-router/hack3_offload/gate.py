"""The gate — run one offload call through verify-or-reject, then ledger it.

Policy (this file) is separated from mechanism (model.py, ledger.py, tokens.py):
the gate decides; the others plumb.

Flow, per the brief:
  1. render prompt, call the model, parse + verify.
  2. on verifier failure: retry ONCE with the failure reason appended.
  3. second failure: REJECT with the reason. Never release unverified output.
"""

from __future__ import annotations

import json
import time
from dataclasses import dataclass
from typing import Any

from . import ledger, tasks, tokens
from .verifiers import VERIFIERS

MAX_ATTEMPTS = 2  # initial try + one retry


@dataclass
class Outcome:
    verdict: str            # "PASS" | "REJECT"
    output: Any | None      # verified output object on PASS, else None
    reason: str | None      # failure reason on REJECT
    record: ledger.Record


def _parse_and_verify(task_class: str, task_input: dict, raw: str):
    """Return (ok, parsed_or_None, reason)."""
    try:
        parsed = json.loads(raw)
    except json.JSONDecodeError as e:
        return False, None, f"model output is not valid JSON: {e}"
    ok, reason = VERIFIERS[task_class](task_input, parsed)
    return ok, (parsed if ok else None), reason


def run(task_class: str, task_input: dict, model, *, live: bool,
        fixture: str | None = None) -> Outcome:
    if task_class not in VERIFIERS:
        raise ValueError(f"unknown task class {task_class!r}")

    input_bytes = len(json.dumps(task_input))
    first_prompt = tasks.render_prompt(task_class, task_input)

    reason: str | None = None
    last_raw = ""
    verified: Any | None = None
    attempts = 0

    t0 = time.perf_counter()
    for attempt in range(1, MAX_ATTEMPTS + 1):
        attempts = attempt
        prompt = tasks.render_prompt(task_class, task_input, reason if attempt > 1 else None)
        last_raw = model.generate(prompt)
        ok, parsed, reason = _parse_and_verify(task_class, task_input, last_raw)
        if ok:
            verified = parsed
            break
    wall_ms = int((time.perf_counter() - t0) * 1000)

    passed = verified is not None
    verdict = "PASS" if passed else "REJECT"
    output_str = json.dumps(verified) if passed else last_raw

    in_tok = tokens.count(first_prompt)
    out_tok = tokens.count(output_str)
    record = ledger.Record(
        task=task_class,
        verdict=verdict,
        attempts=attempts,
        input_bytes=input_bytes,
        output_bytes=len(output_str),
        input_tokens_est=in_tok,
        output_tokens_est=out_tok,
        frontier_tokens_est=in_tok + out_tok,
        local_wall_ms=wall_ms,
        model=model.label(),
        live=live,
        token_backend=tokens.backend(),
        reject_reason=None if passed else reason,
        fixture=fixture,
    )
    return Outcome(verdict=verdict, output=verified, reason=None if passed else reason,
                   record=record)
