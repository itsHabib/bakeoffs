"""Token counting. tiktoken/o200k_base is a *proxy* for Claude's real
tokenizer (which is not public) — every number this harness prints is a
proxy measurement, stated as such in the README and the report header.

The rest of the harness depends only on the small `Tokenizer` shape
(encode/decode) and the `count` callable, never on tiktoken directly, so
rules and tests inject a trivial fake and stay dependency-free.
"""

from __future__ import annotations

from typing import Protocol


class Tokenizer(Protocol):
    def encode(self, text: str) -> list: ...
    def decode(self, ids: list) -> str: ...


_ENC = None


def encoder() -> Tokenizer:
    """The real o200k_base encoder. Imported lazily so importing the
    package (and running the rule/safety tests) needs no venv."""
    global _ENC
    if _ENC is None:
        import tiktoken  # deferred: only the recount path needs it

        _ENC = tiktoken.get_encoding("o200k_base")
    return _ENC


def count(text: str, enc: Tokenizer | None = None) -> int:
    if not text:
        return 0
    return len((enc or encoder()).encode(text))


class WordTokenizer:
    """Deterministic stand-in used by unit tests: one token per
    whitespace-delimited word. Satisfies the Tokenizer shape so the
    token-capping rule can be table-tested without tiktoken."""

    def encode(self, text: str) -> list:
        return text.split(" ")

    def decode(self, ids: list) -> str:
        return " ".join(ids)
