"""The hygiene rules. Each is a pure transform of one tool result into a
smaller one, plus the exact spans it removed (so safety.py can ask whether
the agent later needed any of them).

Design: a rule is a small object with `.name`, a fresh-per-session
`.reset()`, and `.apply(result) -> Applied`. Only DedupeReads carries
cross-result state (the set of reads already seen); the other three are
stateless. Rules never call a tokenizer except CapResult, which takes one
by injection — so the whole module is unit-testable with no venv.

`Applied.removed` is the list of text spans deleted from the original. An
empty list means the rule did not fire on this result.
"""

from __future__ import annotations

import hashlib
import re
from dataclasses import dataclass


@dataclass
class Applied:
    content: str  # the dieted result
    removed: list  # spans of text removed (for safety scoring)

    @property
    def fired(self) -> bool:
        return bool(self.removed)


def _noop(content: str) -> Applied:
    return Applied(content=content, removed=[])


class TruncateReads:
    """Keep only the first `n` lines of a file read; drop the tail behind a
    marker. Applies to Read results only — the agent asked for a file, and
    beyond a couple hundred lines it almost never used the rest."""

    def __init__(self, n: int = 200):
        self.n = n
        self.name = f"truncate_reads>{n}"

    def reset(self):
        pass

    def apply(self, result) -> Applied:
        if result.tool_name != "Read":
            return _noop(result.content)
        lines = result.content.split("\n")
        if len(lines) <= self.n:
            return _noop(result.content)
        kept = lines[: self.n]
        dropped = lines[self.n :]
        marker = f"\n... [diet: truncated {len(dropped)} lines beyond {self.n}]"
        return Applied("\n".join(kept) + marker, removed=["\n".join(dropped)])


class DedupeReads:
    """Replace a re-read of a file whose content is byte-identical to an
    earlier read in the same session with a one-line stub. A changed file
    (different content at the same path) is left intact — the diet only
    removes what is provably redundant."""

    def __init__(self):
        self.name = "dedupe_reads"
        self._seen: dict = {}  # file_path -> (content_hash, first_index)

    def reset(self):
        self._seen = {}

    def apply(self, result) -> Applied:
        if result.tool_name != "Read":
            return _noop(result.content)
        path = result.file_path or f"anon:{result.index}"
        digest = hashlib.sha1(result.content.encode("utf-8", "replace")).hexdigest()
        prev = self._seen.get(path)
        if prev is None:
            self._seen[path] = (digest, result.index)
            return _noop(result.content)
        prev_hash, first_index = prev
        if prev_hash != digest:
            # content changed since the earlier read — keep it, refresh anchor
            self._seen[path] = (digest, result.index)
            return _noop(result.content)
        stub = f"[diet: {path} unchanged since read #{first_index} above]"
        return Applied(stub, removed=[result.content])


_ANSI = re.compile(r"\x1b\[[0-9;?]*[ -/]*[@-~]")
_BLANKS = re.compile(r"\n[ \t]*\n(?:[ \t]*\n)+")


class StripNoise:
    """Strip terminal noise from any tool result: ANSI escape sequences,
    carriage-return progress-bar redraws (keep only the final frame of each
    line), and runs of 3+ blank lines collapsed to one. Removes rendering
    artifacts, never content."""

    def __init__(self):
        self.name = "strip_noise"

    def reset(self):
        pass

    def apply(self, result) -> Applied:
        text = result.content
        if not text:
            return _noop(text)
        out = _ANSI.sub("", text)
        # progress-bar redraws: "a\rb\rc" on one line means only "c" survived
        if "\r" in out:
            out = "\n".join(seg.split("\r")[-1] for seg in out.split("\n"))
        out = _BLANKS.sub("\n\n", out)
        if out == text:
            return _noop(text)
        # removed span = the characters that disappeared, as a single blob
        removed = _diff_removed(text, out)
        return Applied(out, removed=[removed] if removed else [])


class CapResult:
    """Hard cap on any single result at `max_tokens`, measured by an
    injected tokenizer, with a marker. The backstop for the pathological
    result no other rule caught (a 40k-token dumped log)."""

    def __init__(self, max_tokens: int = 1500, tokenizer=None):
        self.max_tokens = max_tokens
        self.name = f"cap_result>{max_tokens}tok"
        self._tok = tokenizer

    def reset(self):
        pass

    def _tokenizer(self):
        if self._tok is None:
            from .tokens import encoder

            self._tok = encoder()
        return self._tok

    def apply(self, result) -> Applied:
        tok = self._tokenizer()
        ids = tok.encode(result.content)
        if len(ids) <= self.max_tokens:
            return _noop(result.content)
        kept = tok.decode(ids[: self.max_tokens])
        dropped = tok.decode(ids[self.max_tokens :])
        marker = f"\n... [diet: capped at {self.max_tokens} tokens, {len(ids) - self.max_tokens} dropped]"
        return Applied(kept + marker, removed=[dropped])


def _diff_removed(original: str, dieted: str) -> str:
    """Best-effort reconstruction of what StripNoise deleted: the ANSI/CR
    artifacts. We report the residue by stripping the *kept* characters from
    a normalized view — good enough to feed identifier extraction, which is
    all safety scoring needs."""
    # Characters unique to the original (artifacts) — cheap set diff keeps
    # this dependency-free; safety only needs identifier candidates from it.
    kept = set(dieted)
    return "".join(ch for ch in original if ch not in kept)


DEFAULT_RULES = [
    lambda: TruncateReads(200),
    lambda: DedupeReads(),
    lambda: StripNoise(),
    lambda: CapResult(1500),
]
