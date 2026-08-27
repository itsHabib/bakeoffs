"""Safety scoring — computed, never vibed.

For every span a rule removed, we ask a deterministic question: does any
*salient* string in the removed text (a code identifier, a path, a numeric
line reference, an error token) reappear in the assistant text that came
*after* that result in the real session? If it does, the agent used
something we deleted — the removal was risky.

A rule's risk score is the fraction of its removed spans that are
later-referenced. A rule that saves tokens with a ~0% reference rate is a
free win; a high reference rate means the rule deletes things the agent
needed. The report makes that tradeoff visible per rule.
"""

from __future__ import annotations

import re
from dataclasses import dataclass

# candidate strings: runs of identifier-ish characters. We then strip
# surrounding punctuation and keep only those with a real identifier signal,
# so "table." / "compression." (prose) fall out while "server.go" survives.
_CANDIDATE = re.compile(r"[A-Za-z0-9_./\-]+")
_DOTTED = re.compile(r"[A-Za-z][A-Za-z0-9]*\.[A-Za-z][A-Za-z0-9]*")  # foo.bar


def _is_salient(tok: str) -> bool:
    """A token worth tracking is something an agent quotes back verbatim: a
    path, a snake_case or camelCase symbol, a dotted name (server.go), or an
    id mixing letters and digits. Plain words — even ending in a period —
    are not, and neither are bare numbers."""
    tok = tok.strip("._-/")
    if len(tok) < 5 or not any(c.isalpha() for c in tok):
        return False
    if "/" in tok or "_" in tok:
        return True
    if any(tok[i].islower() and tok[i + 1].isupper() for i in range(len(tok) - 1)):
        return True  # camelCase boundary
    if any(c.isdigit() for c in tok):
        return True  # id mixing letters + digits (v2, sha1ab…)
    return bool(_DOTTED.fullmatch(tok))  # dotted path, both sides alpha


def salient_tokens(text: str) -> set:
    """Extract the set of salient identifier strings from a removed span."""
    out = set()
    for raw in _CANDIDATE.findall(text or ""):
        t = raw.strip("._-/")
        if _is_salient(t):
            out.add(t)
    return out


@dataclass
class SpanScore:
    salient: int  # count of salient tokens in the removed span
    referenced: bool  # any *lost* token reappears in later assistant text
    hits: tuple  # a few example lost-and-referenced tokens (for the report)


def score_span(removed: str, later_text: str, kept: str = "") -> SpanScore:
    """A removal is risky only if it *loses* an identifier the agent later
    uses. A token that survives in `kept` (e.g. it appears both above and
    below a truncation cut) was not lost — removing the tail cost nothing —
    so we exclude it. Without this, string-search over a whole session's
    later text blames a rule for identifiers it never actually removed.
    """
    toks = salient_tokens(removed)
    if not toks:
        return SpanScore(salient=0, referenced=False, hits=())
    lost = [t for t in toks if t not in kept]
    hits = tuple(sorted(t for t in lost if t in later_text))
    return SpanScore(salient=len(toks), referenced=bool(hits), hits=hits[:5])
