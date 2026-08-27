"""Deterministic verifiers — the law the model must satisfy.

Each verifier takes the original task ``input`` (a dict) and the model's parsed
``output`` (already JSON-decoded) and returns ``(ok, reason)``:

  - ok == True  -> the output is trustworthy; the gate may release it.
  - ok == False -> reason is a precise, operator-readable failure string.

These functions never call a model, never touch the network, and are pure. They
are the entire correctness path, and they are table-tested in tests/.

Design rule (house invariant): correctness is *computed*, never model-judged.
A verifier that can be sweet-talked is not a verifier. Each class picks a shape
whose answer is checkable by construction:

  narrow   -> subset-only + exact syntactic membership (regex is ground truth)
  extract  -> schema-valid + every value appears verbatim in the input
  classify -> labels in the allowed set + planted ground truth scored exactly
"""

from __future__ import annotations

import re
from typing import Any, Callable

# ---------------------------------------------------------------------------
# narrow
# ---------------------------------------------------------------------------

# Registry of *syntactic* criteria -> the spot-check regex that IS the answer.
# Because these are syntactic, the correct matched set is fully determined:
# exactly the paths the regex matches. The verifier does not trust the model's
# judgment; it recomputes the answer and compares.
NARROW_CRITERIA: dict[str, str] = {
    "go test files": r"_test\.go$",
    "markdown files": r"\.md$",
    "files under a cmd/ directory": r"(^|/)cmd/",
}


def resolve_narrow_regex(task_input: dict) -> str | None:
    """Regex for a narrow criterion: explicit override, else registry, else None."""
    override = task_input.get("spot_check_regex")
    if isinstance(override, str) and override:
        return override
    return NARROW_CRITERIA.get(task_input.get("criterion", ""))


def verify_narrow(task_input: dict, output: Any) -> tuple[bool, str | None]:
    if not isinstance(output, dict) or "matched" not in output:
        return False, 'output must be a JSON object with a "matched" array'
    matched = output["matched"]
    if not isinstance(matched, list) or not all(isinstance(p, str) for p in matched):
        return False, '"matched" must be an array of strings'

    paths = task_input.get("paths", [])
    path_set = set(paths)

    # subset-only: the model may not invent paths that were not offered.
    invented = [p for p in matched if p not in path_set]
    if invented:
        return False, f"output is not a subset of input; invented paths: {invented}"

    regex = resolve_narrow_regex(task_input)
    if regex is None:
        return (
            False,
            f'criterion {task_input.get("criterion")!r} is not syntactic '
            "(no spot-check regex); cannot mechanically verify",
        )
    try:
        pat = re.compile(regex)
    except re.error as e:
        return False, f"invalid spot_check_regex {regex!r}: {e}"

    expected = {p for p in paths if pat.search(p)}
    got = set(matched)

    wrongly_included = sorted(got - expected)  # do not match, should be excluded
    wrongly_excluded = sorted(expected - got)  # do match, should be included
    if wrongly_included:
        return False, (
            f"included items that do not match criterion "
            f"(regex {regex!r}): {wrongly_included}"
        )
    if wrongly_excluded:
        return False, (
            f"missed items that match criterion "
            f"(regex {regex!r}): {wrongly_excluded}"
        )
    return True, None


# ---------------------------------------------------------------------------
# extract
# ---------------------------------------------------------------------------

def _type_ok(value: Any, decl: str) -> bool:
    if decl == "string":
        return isinstance(value, str)
    if decl == "int":
        return isinstance(value, int) and not isinstance(value, bool)
    if decl == "number":
        return isinstance(value, (int, float)) and not isinstance(value, bool)
    return False


def verify_extract(task_input: dict, output: Any) -> tuple[bool, str | None]:
    schema = task_input.get("schema", {})
    text = task_input.get("text", "")
    if not isinstance(output, dict):
        return False, "output must be a JSON object"

    schema_keys = set(schema)
    out_keys = set(output)
    if out_keys != schema_keys:
        missing = sorted(schema_keys - out_keys)
        extra = sorted(out_keys - schema_keys)
        parts = []
        if missing:
            parts.append(f"missing fields {missing}")
        if extra:
            parts.append(f"unexpected fields {extra}")
        return False, "schema mismatch: " + "; ".join(parts)

    for field, decl in schema.items():
        value = output[field]
        if not _type_ok(value, decl):
            return False, (
                f"field {field!r} expected type {decl}, got "
                f"{type(value).__name__} ({value!r})"
            )
        # anti-hallucination: the value must appear verbatim in the input text.
        needle = value if isinstance(value, str) else _num_str(value)
        if needle not in text:
            return False, (
                f"value {needle!r} for field {field!r} not found verbatim in "
                "input (hallucinated value)"
            )
    return True, None


def _num_str(value: Any) -> str:
    """String form of a number as it would appear in text (drop trailing .0)."""
    if isinstance(value, float) and value.is_integer():
        return str(int(value))
    return str(value)


# ---------------------------------------------------------------------------
# classify
# ---------------------------------------------------------------------------

def verify_classify(task_input: dict, output: Any) -> tuple[bool, str | None]:
    labels = task_input.get("labels", [])
    lines = task_input.get("lines", [])
    planted = task_input.get("planted", {}) or {}
    label_set = set(labels)

    if not isinstance(output, dict) or "assignments" not in output:
        return False, 'output must be a JSON object with an "assignments" array'
    assignments = output["assignments"]
    if not isinstance(assignments, list):
        return False, '"assignments" must be an array'

    # every line covered, exactly once, in order.
    if len(assignments) != len(lines):
        return False, (
            f"expected {len(lines)} assignments (one per line), "
            f"got {len(assignments)}"
        )

    bad = sorted({a for a in assignments if a not in label_set})
    if bad:
        return False, f"labels outside the allowed set {sorted(label_set)}: {bad}"

    # planted ground truth scored exactly.
    for idx_str, expected_label in planted.items():
        idx = int(idx_str)
        if idx < 0 or idx >= len(assignments):
            return False, f"planted index {idx} out of range"
        got = assignments[idx]
        if got != expected_label:
            return False, (
                f"planted line {idx}: expected label {expected_label!r}, "
                f"got {got!r}"
            )
    return True, None


# ---------------------------------------------------------------------------
# registry
# ---------------------------------------------------------------------------

VERIFIERS: dict[str, Callable[[dict, Any], tuple[bool, str | None]]] = {
    "narrow": verify_narrow,
    "extract": verify_extract,
    "classify": verify_classify,
}

TASK_CLASSES = tuple(VERIFIERS.keys())
