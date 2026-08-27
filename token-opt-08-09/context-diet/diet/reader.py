"""Replay reader: walk a Claude Code session transcript (JSONL) and
reconstruct, in order, every tool_use -> tool_result pair together with the
assistant text that follows each result in the same session.

Read-only. Never modifies the source transcript.
"""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from pathlib import Path


@dataclass
class Result:
    """One tool result, with the tool that produced it and the assistant
    reasoning that came after it."""

    index: int  # position among results in this session
    tool_name: str
    tool_input: dict
    content: str
    # concatenation of every assistant text block emitted *after* this
    # result — the evidence for "did the agent later reference what we cut".
    later_text: str = ""

    @property
    def file_path(self) -> str:
        return str(self.tool_input.get("file_path", ""))


@dataclass
class Session:
    path: str
    results: list = field(default_factory=list)


def _content_to_str(content) -> str:
    """tool_result.content is either a plain string or a list of
    {type:text,text} blocks. Normalize to one string."""
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        parts = []
        for c in content:
            if isinstance(c, dict) and c.get("type") == "text":
                parts.append(c.get("text", ""))
            elif isinstance(c, str):
                parts.append(c)
        return "".join(parts)
    return "" if content is None else str(content)


def _iter_lines(path: Path):
    with path.open("r", errors="replace") as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            try:
                yield json.loads(line)
            except json.JSONDecodeError:
                continue


def read_session(path) -> Session:
    """Parse one transcript into an ordered Session of Results.

    Two passes conceptually, one loop: tool_use inputs are recorded as we
    see them (they precede their result), assistant text is accumulated in
    file order, and each result snapshots how much assistant text existed
    *before* it — everything after that snapshot is its `later_text`.
    """
    path = Path(path)
    uses: dict = {}  # tool_use_id -> (name, input)
    texts: list = []  # assistant text blocks, in file order
    pending: list = []  # (Result, text_cursor) awaiting later_text fill-in
    results: list = []

    for obj in _iter_lines(path):
        typ = obj.get("type")
        msg = obj.get("message") or {}
        content = msg.get("content")

        if typ == "assistant" and isinstance(content, list):
            for c in content:
                if not isinstance(c, dict):
                    continue
                if c.get("type") == "tool_use":
                    uses[c.get("id")] = (c.get("name", ""), c.get("input") or {})
                elif c.get("type") == "text" and c.get("text"):
                    texts.append(c["text"])

        if typ == "user" and isinstance(content, list):
            for c in content:
                if not (isinstance(c, dict) and c.get("type") == "tool_result"):
                    continue
                name, inp = uses.get(c.get("tool_use_id"), ("", {}))
                r = Result(
                    index=len(results),
                    tool_name=name,
                    tool_input=inp,
                    content=_content_to_str(c.get("content")),
                )
                results.append(r)
                pending.append((r, len(texts)))

    for r, cursor in pending:
        r.later_text = "\n".join(texts[cursor:])

    return Session(path=str(path), results=results)


def iter_sessions(root):
    """Yield a Session for every *.jsonl under root."""
    for p in sorted(Path(root).rglob("*.jsonl")):
        try:
            yield read_session(p)
        except (OSError, ValueError):
            continue
