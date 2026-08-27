"""Prompt templates for qwen2.5:7b, one per task class.

The prompt is the ONLY place the model's phrasing lives. Correctness is decided
by verifiers.py, never here. On a retry the gate appends the verifier's failure
reason so the second attempt sees exactly why the first was rejected.

Prompts ask for strict JSON. We also request ollama's JSON mode at the transport
layer (model.py), so the raw response should already parse; when it does not,
that is simply a verifier failure like any other.
"""

from __future__ import annotations

import json

_RETRY_SUFFIX = (
    "\n\nYour previous answer was REJECTED by a deterministic verifier for this "
    "reason:\n  {reason}\nFix exactly that and return corrected JSON. Do not "
    "explain."
)


def _narrow(task_input: dict) -> str:
    return (
        "You filter a list of file paths to those matching a criterion.\n"
        f"Criterion: {task_input.get('criterion')!r}\n"
        "Paths:\n"
        + json.dumps(task_input.get("paths", []), indent=2)
        + '\n\nReturn ONLY JSON of the form {"matched": [<subset of the paths '
        "above, verbatim>]}. Include a path if and only if it satisfies the "
        "criterion. Do not invent paths."
    )


def _extract(task_input: dict) -> str:
    schema = task_input.get("schema", {})
    fields = ", ".join(f"{k} ({v})" for k, v in schema.items())
    return (
        "You extract structured fields from noisy command output.\n"
        f"Extract exactly these fields: {fields}.\n"
        "Every value you return MUST appear verbatim in the text below; never "
        "guess or fabricate a value that is not literally present.\n\n"
        "TEXT:\n"
        + task_input.get("text", "")
        + "\n\nReturn ONLY a JSON object with exactly the requested fields."
    )


def _classify(task_input: dict) -> str:
    labels = task_input.get("labels", [])
    lines = task_input.get("lines", [])
    numbered = "\n".join(f"{i}: {ln}" for i, ln in enumerate(lines))
    return (
        "You classify each log line with exactly one label.\n"
        f"Allowed labels (use only these): {labels}\n\n"
        "LINES:\n"
        + numbered
        + '\n\nReturn ONLY JSON of the form {"assignments": [<one label per '
        "line, in the same order as the lines above>]}. The array length must "
        f"equal {len(lines)}."
    )


_RENDERERS = {
    "narrow": _narrow,
    "extract": _extract,
    "classify": _classify,
}


def render_prompt(task_class: str, task_input: dict, reason: str | None = None) -> str:
    """Render the model prompt for a task, optionally with a retry reason."""
    base = _RENDERERS[task_class](task_input)
    if reason:
        base += _RETRY_SUFFIX.format(reason=reason)
    return base
