#!/usr/bin/env python3
"""Provenance for the bundled fixtures. Synthesizes two sessions with PLANTED,
known cache behavior — no scrubbed real data. `make demo` reads the committed
.jsonl files, not this script; regenerate only if you change the plant.

  steady.jsonl : 4 assistant turns, high hit ratio, ZERO busts.
  busty.jsonl  : 7 assistant turns with exactly TWO engineered busts of known
                 excess (20000 tokens @5m, then 23500 tokens @1h), each blamed
                 on a large tool result (Bash, then Read).
"""
import json

T0 = "2026-08-01T12:00:00.000Z"

def ts(minute):
    return f"2026-08-01T12:{minute:02d}:00.000Z"

def assistant(minute, read, c1h, c5m, inp=0, out=500, tool=None):
    content = [{"type": "text", "text": "ok"}]
    if tool:
        name, tid = tool
        content.append({"type": "tool_use", "id": tid, "name": name, "input": {}})
    return {
        "type": "assistant",
        "timestamp": ts(minute),
        "message": {
            "role": "assistant",
            "content": content,
            "usage": {
                "input_tokens": inp,
                "cache_read_input_tokens": read,
                "cache_creation_input_tokens": c1h + c5m,
                "output_tokens": out,
                "cache_creation": {
                    "ephemeral_1h_input_tokens": c1h,
                    "ephemeral_5m_input_tokens": c5m,
                },
            },
        },
    }

def tool_result(tid, nchars):
    return {
        "type": "user",
        "message": {
            "role": "user",
            "content": [{"type": "tool_result", "tool_use_id": tid,
                         "content": "x" * nchars, "is_error": False}],
        },
    }

def user_text(s):
    return {"type": "user", "message": {"role": "user", "content": s}}

def write(path, rows):
    with open(path, "w") as f:
        for r in rows:
            f.write(json.dumps(r) + "\n")

# --- steady: no busts, high hit ---------------------------------------------
steady = [
    user_text("start"),
    assistant(0, read=15000, c1h=5000, c5m=0),      # idx0 first — excluded
    user_text("next"),
    assistant(1, read=20000, c1h=0, c5m=1000),      # thr .25*15000=3750; ok
    user_text("next"),
    assistant(2, read=21000, c1h=0, c5m=1000),      # thr .25*20000=5000; ok
    user_text("next"),
    assistant(3, read=22000, c1h=0, c5m=1000),      # thr .25*21000=5250; ok
]

# --- busty: exactly two engineered busts ------------------------------------
busty = [
    user_text("start"),
    assistant(0, read=15000, c1h=5000, c5m=0, tool=("Bash", "b1")),  # idx0 first
    tool_result("b1", 40000),                        # big Bash result enters
    assistant(2, read=3000, c1h=0, c5m=25000),       # BUST#1 thr .25*15000=3750
                                                     #   excess 25000-3750=21250 @5m, cause tool:Bash
    user_text("continue"),
    assistant(3, read=25000, c1h=0, c5m=200, tool=("Read", "r1")),   # thr .25*3000=750; ok
    tool_result("r1", 50000),                        # big Read result enters
    assistant(5, read=4000, c1h=30000, c5m=0),       # BUST#2 thr .25*25000=6250
                                                     #   excess 30000-6250=23750 @1h, cause tool:Read
    user_text("done"),
    assistant(6, read=30000, c1h=0, c5m=200),        # thr .25*4000=1000; ok
]

write("steady.jsonl", steady)
write("busty.jsonl", busty)
print("wrote steady.jsonl, busty.jsonl")
