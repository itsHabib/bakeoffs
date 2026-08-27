"""Token estimation.

tiktoken's ``o200k_base`` is a *proxy* for the real Claude tokenizer, not the
tokenizer itself. Every ledger number derived from this module is an estimate,
labeled as such. If tiktoken is unavailable we fall back to a coarse
~4-chars-per-token heuristic so tests still run keyless/offline.
"""

from __future__ import annotations

_encoder = None
_USING_TIKTOKEN = False

try:  # pragma: no cover - exercised via count()
    import tiktoken

    _encoder = tiktoken.get_encoding("o200k_base")
    _USING_TIKTOKEN = True
except Exception:  # tiktoken absent -> heuristic fallback
    _encoder = None
    _USING_TIKTOKEN = False


def count(text: str) -> int:
    """Estimate the token count of ``text``.

    Uses tiktoken ``o200k_base`` when present; otherwise a chars/4 heuristic.
    """
    if not text:
        return 0
    if _encoder is not None:
        return len(_encoder.encode(text))
    return max(1, len(text) // 4)


def backend() -> str:
    """Name the estimator in use, for honest reporting."""
    return "tiktoken:o200k_base (proxy)" if _USING_TIKTOKEN else "heuristic:chars/4"
