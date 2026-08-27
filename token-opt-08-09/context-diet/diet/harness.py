"""The replay harness: run the rule set over sessions and accumulate, per
rule and in aggregate, the tokens saved and the safety (later-referenced)
score. This is the product — it prices any hygiene rule against real
transcripts before anyone ships it live.

Per-rule attribution is *independent*: each rule is run alone over the
original results, so its saved-token number and risk score are its own,
not confounded by rule ordering. The `combined` number applies every rule
in sequence to each result and is the real end-to-end total (<= the sum of
independent savings, since rules can overlap).
"""

from __future__ import annotations

from dataclasses import dataclass, field

from . import safety
from .reader import iter_sessions, read_session
from .tokens import count, encoder


@dataclass
class RuleStat:
    name: str
    orig_tokens: int = 0  # tokens in results this rule touched
    saved_tokens: int = 0
    fires: int = 0  # results the rule fired on
    spans: int = 0
    spans_referenced: int = 0
    examples: list = field(default_factory=list)  # sample referenced hits

    @property
    def ref_rate(self) -> float:
        return self.spans_referenced / self.spans if self.spans else 0.0

    @property
    def saved_pct_of_touched(self) -> float:
        return self.saved_tokens / self.orig_tokens if self.orig_tokens else 0.0


@dataclass
class Report:
    sessions: int = 0
    results: int = 0
    total_result_tokens: int = 0  # proxy tokens over every tool result
    combined_saved_tokens: int = 0
    rules: dict = field(default_factory=dict)  # name -> RuleStat


def _stat(report: Report, name: str) -> RuleStat:
    return report.rules.setdefault(name, RuleStat(name=name))


def run(paths, rule_factories, enc=None) -> Report:
    """Replay `paths` (Session objects or file paths) through the rules.

    `rule_factories` is a list of zero-arg callables each producing a fresh
    rule (state is per-session). `enc` is an injectable tokenizer for tests.
    """
    enc = enc or encoder()
    report = Report()

    for src in paths:
        session = src if hasattr(src, "results") else read_session(src)
        if not session.results:
            continue
        report.sessions += 1

        # independent per-rule pass
        rules = [f() for f in rule_factories]
        for rule in rules:
            rule.reset()

        combined_rules = [f() for f in rule_factories]
        for rule in combined_rules:
            rule.reset()

        for result in session.results:
            report.results += 1
            orig_tok = count(result.content, enc)
            report.total_result_tokens += orig_tok

            for rule in rules:
                applied = rule.apply(result)
                if not applied.fired:
                    continue
                stat = _stat(report, rule.name)
                stat.fires += 1
                stat.orig_tokens += orig_tok
                stat.saved_tokens += orig_tok - count(applied.content, enc)
                for span in applied.removed:
                    stat.spans += 1
                    sc = safety.score_span(span, result.later_text, applied.content)
                    if sc.referenced:
                        stat.spans_referenced += 1
                        if len(stat.examples) < 5:
                            stat.examples.append(sc.hits[0])

            # combined pass: rules chained on the shrinking content
            content = result.content
            for rule in combined_rules:
                # feed the running content through each rule in turn
                proxy = _one_shot(result, content)
                applied = rule.apply(proxy)
                content = applied.content
            report.combined_saved_tokens += orig_tok - count(content, enc)

    return report


class _one_shot:
    """A lightweight Result-shaped view carrying updated content while
    preserving tool_name/input/later_text, for the chained combined pass."""

    def __init__(self, result, content):
        self._r = result
        self.content = content

    def __getattr__(self, k):
        return getattr(self._r, k)


def run_corpus(root, rule_factories, enc=None) -> Report:
    return run(iter_sessions(root), rule_factories, enc=enc)


def sweep(sessions, rule_factory, values, enc=None):
    """Price one rule across a range of its parameter — the harness's whole
    point: cost every hygiene idea before shipping it. Returns a list of
    (value, saved_tokens, saved_pct, ref_rate, fires). `sessions` is
    materialized once so each setting replays the same corpus.
    """
    sessions = list(sessions)
    rows = []
    for v in values:
        rep = run(sessions, [lambda v=v: rule_factory(v)], enc=enc)
        stat = next(iter(rep.rules.values()), None)
        saved = stat.saved_tokens if stat else 0
        pct = saved / rep.total_result_tokens if rep.total_result_tokens else 0.0
        rows.append((v, saved, pct, stat.ref_rate if stat else 0.0,
                     stat.fires if stat else 0))
    return rows
