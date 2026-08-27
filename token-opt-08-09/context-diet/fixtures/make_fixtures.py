#!/usr/bin/env python3
"""Synthesize fixture transcripts with *known* planted waste. Everything
here is fabricated — no private transcript data — so the fixtures are safe
to commit (house-rule privacy). Tests assert the planted numbers exactly.

Run: python fixtures/make_fixtures.py   (re-emits the .jsonl next to this).

Planted waste, asserted in tests/test_fixtures.py:
  fixture_dedupe.jsonl  : one 30-line file read 3x identically  -> 2 dupes
  fixture_ansi.jsonl    : a build log with ANSI + \\r redraws + blank runs
  fixture_bigread.jsonl : a 400-line read (truncate drops 200, incl. a
                          symbol the agent references later => RISKY) and a
                          ~4000-token bash dump (cap fires)
"""

from __future__ import annotations

import json
from pathlib import Path

HERE = Path(__file__).parent

# --- planted constants (imported by tests) ---------------------------------
DEDUP_PATH = "/repo/pkg/config.go"
DEDUP_LINES = 30
DEDUP_READS = 3
BIGREAD_PATH = "/repo/internal/server.go"
BIGREAD_LINES = 260  # >200 so truncate fires; short lines keep it under the cap
TRUNCATE_N = 200
REFERENCED_SYMBOL = "handleAuthRetry_v2"  # lives in the truncated tail
BIGREAD_SYMBOL_LINE = 250


def _asst_tool_use(tool_id, name, inp, text_before=""):
    content = []
    if text_before:
        content.append({"type": "text", "text": text_before})
    content.append({"type": "tool_use", "id": tool_id, "name": name, "input": inp})
    return {
        "type": "assistant",
        "message": {"role": "assistant", "content": content, "usage": {}},
    }


def _tool_result(tool_id, text):
    return {
        "type": "user",
        "message": {
            "role": "user",
            "content": [
                {"type": "tool_result", "tool_use_id": tool_id, "content": text}
            ],
        },
    }


def _asst_text(text):
    return {
        "type": "assistant",
        "message": {"role": "assistant", "content": [{"type": "text", "text": text}]},
    }


def _numbered(lines):
    """cat -n style, matching how Read results look in real transcripts."""
    return "\n".join(f"{i + 1}\t{ln}" for i, ln in enumerate(lines))


def _write(name, objs):
    path = HERE / name
    with path.open("w") as fh:
        for o in objs:
            fh.write(json.dumps(o) + "\n")
    return path


def make_dedupe():
    body = _numbered([f"line {i} of config" for i in range(DEDUP_LINES)])
    objs = []
    for k in range(DEDUP_READS):
        tid = f"tool_dedupe_{k}"
        objs.append(_asst_tool_use(tid, "Read", {"file_path": DEDUP_PATH}))
        objs.append(_tool_result(tid, body))
        objs.append(_asst_text(f"Checked config, read #{k}. Looks fine, moving on."))
    return _write("fixture_dedupe.jsonl", objs)


def make_ansi():
    # a build log: color codes, a progress bar redrawn via \r, blank runs
    log = (
        "\x1b[32mCompiling\x1b[0m module core\n"
        "downloading [                    ] 0%\r"
        "downloading [==========          ] 50%\r"
        "downloading [====================] 100% done\n"
        "\x1b[1m\x1b[31mERROR\x1b[0m nope just kidding\n"
        "\n\n\n\n"
        "\x1b[32mBUILD OK\x1b[0m in 4.2s\n"
    )
    tid = "tool_ansi_0"
    objs = [
        _asst_tool_use(tid, "Bash", {"command": "make build"}),
        _tool_result(tid, log),
        _asst_text("Build passed. The ANSI/progress noise is irrelevant to me."),
    ]
    return _write("fixture_ansi.jsonl", objs)


def make_bigread():
    lines = []
    for i in range(BIGREAD_LINES):
        if i + 1 == BIGREAD_SYMBOL_LINE:
            lines.append(f"func {REFERENCED_SYMBOL}(c) {{")
        else:
            lines.append(f"s{i}")
    body = _numbered(lines)
    tid = "tool_big_0"

    # a big bash dump to trip the token cap (well over 1500 proxy tokens)
    dump = "\n".join(f"2026-08-09T00:00:{i:02d} log line payload {i} xyz" for i in range(400))
    tid2 = "tool_big_1"

    objs = [
        _asst_tool_use(tid, "Read", {"file_path": BIGREAD_PATH}),
        _tool_result(tid, body),
        # the agent LATER references a symbol that lived in the truncated tail
        _asst_text(
            f"I need to call {REFERENCED_SYMBOL} from the auth path; it was "
            f"defined near line {BIGREAD_SYMBOL_LINE}."
        ),
        _asst_tool_use(tid2, "Bash", {"command": "cat huge.log"}),
        _tool_result(tid2, dump),
        _asst_text("Scanned the log, nothing actionable."),
    ]
    return _write("fixture_bigread.jsonl", objs)


def main():
    for fn in (make_dedupe, make_ansi, make_bigread):
        print("wrote", fn().name)


if __name__ == "__main__":
    main()
