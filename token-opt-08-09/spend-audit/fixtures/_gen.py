#!/usr/bin/env python3
"""Synthesize the fixture corpus with KNOWN totals (asserted in test_audit.py).

Run: python3 fixtures/_gen.py   (emits *.jsonl next to this file)

Every usage number here is fake and round on purpose so a human can add them up.
No real transcript data — safe to commit. See test_audit.py for the expected sums.
"""
import json
import os

HERE = os.path.dirname(__file__)


def asst(mid, ts, cwd, model, usage, tool=None):
    content = [{"type": "text", "text": "ok"}]
    if tool:
        content.append({"type": "tool_use", "id": tool[0], "name": tool[1]})
    return {
        "type": "assistant",
        "timestamp": ts,
        "cwd": cwd,
        "sessionId": "sess",
        "message": {
            "role": "assistant",
            "id": mid,
            "model": model,
            "content": content,
            "usage": {
                "input_tokens": usage[0],
                "cache_read_input_tokens": usage[1],
                "output_tokens": usage[4],
                "cache_creation_input_tokens": usage[2] + usage[3],
                "cache_creation": {
                    "ephemeral_5m_input_tokens": usage[2],
                    "ephemeral_1h_input_tokens": usage[3],
                },
            },
        },
    }


def user_result(tuid, content):
    return {
        "type": "user",
        "message": {
            "role": "user",
            "content": [{"type": "tool_result", "tool_use_id": tuid, "content": content}],
        },
    }


# usage tuple = (input, cache_read, cache_write_5m, cache_write_1h, output)
FILES = {
    "session-a.jsonl": [
        asst("a1", "2026-08-01T09:00:00Z", "/home/dev/alpha", "claude-opus-5",
             (100, 10000, 2000, 500, 800), ("tu_a1", "Bash")),
        user_result("tu_a1", "AAAA"),
        asst("a2", "2026-08-01T09:05:00Z", "/home/dev/alpha", "claude-opus-5",
             (50, 12000, 0, 0, 1200), ("tu_a2", "Read")),
        user_result("tu_a2", "BBBBBBBB"),
        # duplicate of a2 (same message id) — must be skipped for usage.
        asst("a2", "2026-08-01T09:05:01Z", "/home/dev/alpha", "claude-opus-5",
             (50, 12000, 0, 0, 1200)),
        # streamed partials of a3: stub output first, final output second.
        # Dedupe must keep the RICHER line (output 500), not the first (2).
        asst("a3", "2026-08-01T09:10:00Z", "/home/dev/alpha", "claude-opus-5",
             (60, 3000, 0, 0, 2)),
        asst("a3", "2026-08-01T09:10:01Z", "/home/dev/alpha", "claude-opus-5",
             (60, 3000, 0, 0, 500)),
    ],
    "session-b.jsonl": [
        # worktree cwd — must roll up into /home/dev/alpha
        asst("b1", "2026-08-02T10:00:00Z",
             "/home/dev/alpha/.claude/worktrees/foo-123", "claude-opus-5",
             (200, 20000, 3000, 0, 1500), ("tu_b1", "Bash")),
        user_result("tu_b1", "CCCCCC"),
    ],
    "session-c.jsonl": [
        asst("c1", "2026-08-02T11:00:00Z", "/home/dev/beta", "claude-fable-5",
             (80, 5000, 1000, 0, 600), ("tu_c1", "Grep")),
        user_result("tu_c1", [{"type": "text", "text": "DD"}]),
        asst("c2", "2026-08-02T11:10:00Z", "/home/dev/beta", "claude-fable-5",
             (40, 6000, 0, 0, 400)),
    ],
    "session-d.jsonl": [
        asst("d1", "2026-08-03T08:00:00Z", "/home/dev/beta", "claude-fable-5",
             (120, 8000, 0, 1000, 900)),
    ],
    "session-e.jsonl": [
        asst("e1", "2026-08-03T14:00:00Z", "/home/dev/gamma", "claude-opus-4-8",
             (300, 15000, 2500, 0, 2000)),
        # synthetic all-zero message — must be ignored entirely.
        asst("syn", "2026-08-03T14:01:00Z", "/home/dev/gamma", "<synthetic>",
             (0, 0, 0, 0, 0)),
        # orphan tool_result — no matching tool_use — must count as unknown tool.
        user_result("tu_missing", "EEEEE"),
    ],
}


def main():
    for name, records in FILES.items():
        with open(os.path.join(HERE, name), "w", encoding="utf-8") as fh:
            for r in records:
                fh.write(json.dumps(r, ensure_ascii=False) + "\n")
    print(f"wrote {len(FILES)} fixtures to {HERE}")


if __name__ == "__main__":
    main()
