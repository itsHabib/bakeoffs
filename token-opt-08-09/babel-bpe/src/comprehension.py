"""Comprehension-tax rig: does compressing the payload cost the reader accuracy?

For every (format, payload, question) we hand qwen2.5:7b the format's legend +
the encoded payload + one question, and grade its answer by deterministic
exact-match against the known value. The model is the SUBJECT under test; all
grading is table-tested code. Output: accuracy per format, to sit next to
tokens per format — the frontier, not a single winner.

Determinism knobs: temperature 0, fixed seed. The model is still a model, so
runs can vary slightly; that is exactly why the real run is cached and
committed, and the demo reads the cache by default.
"""

import json
import re
import urllib.error
import urllib.request

import corpus
import encoders as E
import questions
import specs

OLLAMA = "http://localhost:11434"
MODEL = "qwen2.5:7b"

_CORPUS = dict(corpus.CORPUS)


def ollama_up():
    try:
        urllib.request.urlopen(OLLAMA + "/api/tags", timeout=2)
        return True
    except Exception:
        return False


def ask(prompt):
    body = json.dumps({
        "model": MODEL,
        "prompt": prompt,
        "stream": False,
        "options": {"temperature": 0, "seed": 7, "num_predict": 48},
    }).encode()
    req = urllib.request.Request(OLLAMA + "/api/generate", data=body,
                                 headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=120) as r:
        return json.loads(r.read())["response"]


def _norm(s):
    return re.sub(r"\s+", " ", s.strip().lower())


def grade(expected, response, numeric):
    exp, resp = _norm(expected), _norm(response)
    if numeric:
        return re.search(rf"(?<!\d){re.escape(exp)}(?!\d)", resp) is not None
    return exp in resp


def prompt_for(fmt, name, question):
    p = _CORPUS[name]
    legend = specs.SPEC[fmt]
    lead = f"You are reading one structured payload in the {fmt} format.\n"
    if legend:
        lead += f"Format legend: {legend}\n"
    return (
        lead
        + "Payload:\n"
        + E.encode(fmt, p).rstrip("\n")
        + f"\n\nQuestion: {question}\n"
        + "Answer with ONLY the value, no explanation."
    )


def _items():
    for name, qs in questions.QUESTIONS.items():
        shape = _CORPUS[name]["type"]
        for fmt in E.FORMAT_ORDER:
            if not E.applies(fmt, shape):
                continue
            for q, exp, numeric in qs:
                yield fmt, name, q, exp, numeric


def run_full(progress=None):
    """Run every (format, payload, question); return per-format accuracy."""
    tally = {fmt: {"correct": 0, "total": 0, "misses": []} for fmt in E.FORMAT_ORDER}
    items = list(_items())
    for i, (fmt, name, q, exp, numeric) in enumerate(items):
        resp = ask(prompt_for(fmt, name, q))
        ok = grade(exp, resp, numeric)
        t = tally[fmt]
        t["total"] += 1
        t["correct"] += int(ok)
        if not ok:
            t["misses"].append({"payload": name, "q": q, "expected": exp,
                                 "got": _norm(resp)[:60]})
        if progress:
            progress(i + 1, len(items), fmt, name, ok)
    return {fmt: t for fmt, t in tally.items() if t["total"] > 0}
