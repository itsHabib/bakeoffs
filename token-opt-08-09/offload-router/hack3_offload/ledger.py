"""The displacement ledger — proof, in JSONL, that offload was worth it.

One append-only record per gate call. The whole point of the router is turning
"use a cheaper model sometimes" from vibes into a number, so the ledger is the
product-facing artifact: displaced frontier tokens and rejection rate per class.

A record:
  {
    "task": "extract", "verdict": "PASS" | "REJECT", "attempts": 1|2,
    "input_bytes": .., "output_bytes": ..,
    "input_tokens_est": .., "output_tokens_est": .., "frontier_tokens_est": ..,
    "local_wall_ms": .., "model": "cassette:..|ollama:qwen2.5:7b",
    "live": false, "reject_reason": null | "...", "token_backend": "..."
  }

All token figures are estimates (tiktoken o200k_base proxy); labeled as such.
"""

from __future__ import annotations

import json
import os
from dataclasses import asdict, dataclass, field

from . import prices


@dataclass
class Record:
    task: str
    verdict: str
    attempts: int
    input_bytes: int
    output_bytes: int
    input_tokens_est: int
    output_tokens_est: int
    frontier_tokens_est: int
    local_wall_ms: int
    model: str
    live: bool
    token_backend: str
    reject_reason: str | None = None
    fixture: str | None = None

    def to_json(self) -> str:
        return json.dumps(asdict(self))


def append(path: str, record: Record) -> None:
    os.makedirs(os.path.dirname(os.path.abspath(path)), exist_ok=True)
    with open(path, "a") as f:
        f.write(record.to_json() + "\n")


def _load(path: str) -> list[dict]:
    if not os.path.exists(path):
        return []
    out = []
    with open(path) as f:
        for line in f:
            line = line.strip()
            if line:
                out.append(json.loads(line))
    return out


@dataclass
class ClassRollup:
    calls: int = 0
    passed: int = 0
    rejected: int = 0
    displaced_tokens: int = 0  # frontier tokens displaced (PASS only)
    local_wall_ms: int = 0

    @property
    def rejection_rate(self) -> float:
        return self.rejected / self.calls if self.calls else 0.0


def rollup(path: str) -> dict[str, ClassRollup]:
    """Aggregate the ledger by task class. PASS calls displace frontier tokens;
    REJECT calls displace nothing (they would have fallen back)."""
    by_class: dict[str, ClassRollup] = {}
    for r in _load(path):
        c = by_class.setdefault(r["task"], ClassRollup())
        c.calls += 1
        c.local_wall_ms += r.get("local_wall_ms", 0)
        if r["verdict"] == "PASS":
            c.passed += 1
            c.displaced_tokens += r.get("frontier_tokens_est", 0)
        else:
            c.rejected += 1
    return by_class


def format_report(path: str) -> str:
    by_class = rollup(path)
    if not by_class:
        return f"no ledger records at {path}"

    lines: list[str] = []
    lines.append("Displacement ledger — frontier tokens saved by verified offload")
    lines.append(f"  source: {path}")
    lines.append("  (token figures are estimates; tiktoken o200k_base is a proxy "
                 "for the Claude tokenizer)")
    lines.append("")
    header = f"  {'class':<10} {'calls':>6} {'pass':>5} {'reject':>7} " \
             f"{'rej.rate':>9} {'displaced tok':>14}"
    lines.append(header)
    lines.append("  " + "-" * (len(header) - 2))

    tot = ClassRollup()
    for name in sorted(by_class):
        c = by_class[name]
        tot.calls += c.calls
        tot.passed += c.passed
        tot.rejected += c.rejected
        tot.displaced_tokens += c.displaced_tokens
        lines.append(
            f"  {name:<10} {c.calls:>6} {c.passed:>5} {c.rejected:>7} "
            f"{c.rejection_rate:>8.0%} {c.displaced_tokens:>14,}"
        )
    lines.append("  " + "-" * (len(header) - 2))
    lines.append(
        f"  {'TOTAL':<10} {tot.calls:>6} {tot.passed:>5} {tot.rejected:>7} "
        f"{tot.rejection_rate:>8.0%} {tot.displaced_tokens:>14,}"
    )
    lines.append("")
    lines.append(
        f"  Net: {tot.displaced_tokens:,} frontier tokens displaced to a $0 "
        f"local model across {tot.passed} verified calls; "
        f"{tot.rejection_rate:.0%} of calls were caught and rejected "
        "(NOT silently passed through)."
    )

    dollars = prices.dollars(  # only the displaced (PASS) input+output matter
        _displaced_input(path), _displaced_output(path)
    )
    if dollars is not None:
        lines.append(
            f"  Est. frontier $ displaced: ${dollars:,.2f} "
            f"(model {prices.FRONTIER_MODEL}, prices from prices.py)"
        )
    else:
        lines.append(
            "  Est. frontier $ displaced: (fill prices.py to price the tokens)"
        )
    return "\n".join(lines)


def _displaced_input(path: str) -> int:
    return sum(r.get("input_tokens_est", 0)
               for r in _load(path) if r["verdict"] == "PASS")


def _displaced_output(path: str) -> int:
    return sum(r.get("output_tokens_est", 0)
               for r in _load(path) if r["verdict"] == "PASS")
