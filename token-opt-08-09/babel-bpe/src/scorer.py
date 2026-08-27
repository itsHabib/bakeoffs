"""Scorer: real BPE token counts per payload per format, plus amortized
leaderboards at N=1, N=10, N=100 payloads.

The tokenizer is tiktoken's ``o200k_base``. It is a PROXY for the actual Claude
tokenizer — a real, modern BPE vocabulary, but not the one the invoice is
metered against. Every headline in this repo is "real BPE tokens" only in that
proxy sense; see README.
"""

import functools

import tiktoken

import corpus
import encoders as E
import specs


@functools.lru_cache(maxsize=1)
def _enc():
    return tiktoken.get_encoding("o200k_base")


def count(text):
    return len(_enc().encode(text))


def per_payload():
    """Return {format: {payload_name: token_count}} across the corpus."""
    out = {fmt: {} for fmt in E.FORMAT_ORDER}
    for name, p in corpus.CORPUS:
        for fmt in E.FORMAT_ORDER:
            if not E.applies(fmt, p["type"]):
                continue
            out[fmt][name] = count(E.encode(fmt, p))
    return out


def spec_cost():
    return {fmt: count(specs.SPEC[fmt]) for fmt in E.FORMAT_ORDER}


def totals(scores, spec, n):
    """Amortized token total for sending the whole corpus n times over.

    A format that does not cover every shape (CSV/TSV) is scored only on the
    payloads it can encode, and flagged partial so the leaderboard is honest.
    """
    rows = []
    ncov = len(corpus.CORPUS)
    for fmt in E.FORMAT_ORDER:
        msg = sum(scores[fmt].values())
        covered = len(scores[fmt])
        rows.append({
            "format": fmt,
            "msg_tokens": msg,
            "spec_tokens": spec[fmt],
            "covered": covered,
            "partial": covered < ncov,
            "total": spec[fmt] + msg * n,
        })
    rows.sort(key=lambda r: (r["partial"], r["total"]))
    return rows


def leaderboards(ns=(1, 10, 100)):
    scores = per_payload()
    spec = spec_cost()
    return {n: totals(scores, spec, n) for n in ns}, scores, spec
