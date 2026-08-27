"""Integration tests: replay each bundled fixture and assert the *planted*
waste is caught exactly. These are the numbers the demo claims.

Run under the venv (tiktoken present) so CapResult's default encoder loads.
"""

import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

import fixtures.make_fixtures as fx
from diet.harness import run
from diet.reader import read_session
from diet.rules import CapResult, DedupeReads, StripNoise, TruncateReads

FX = os.path.join(os.path.dirname(__file__), "..", "fixtures")

RULES = [
    lambda: TruncateReads(fx.TRUNCATE_N),
    lambda: DedupeReads(),
    lambda: StripNoise(),
    lambda: CapResult(1500),
]


def _report(name):
    session = read_session(os.path.join(FX, name))
    return run([session], RULES)


def test_dedupe_fixture_catches_two_rereads():
    rep = _report("fixture_dedupe.jsonl")
    stat = rep.rules["dedupe_reads"]
    assert stat.fires == fx.DEDUP_READS - 1  # 3 reads -> 2 redundant
    assert stat.saved_tokens > 0
    assert stat.ref_rate == 0.0  # the stub content is never referenced later


def test_ansi_fixture_strips_escape_heavy_log():
    rep = _report("fixture_ansi.jsonl")
    stat = rep.rules["strip_noise"]
    assert stat.fires == 1
    assert stat.saved_tokens > 0
    assert stat.ref_rate == 0.0  # escape codes are never referenced


def test_bigread_truncate_is_flagged_risky():
    rep = _report("fixture_bigread.jsonl")
    stat = rep.rules[f"truncate_reads>{fx.TRUNCATE_N}"]
    assert stat.fires == 1
    # the truncated tail held a symbol the agent referenced later => risky
    assert stat.ref_rate == 1.0
    assert fx.REFERENCED_SYMBOL in stat.examples


def test_bigread_cap_fires_on_dump():
    rep = _report("fixture_bigread.jsonl")
    stat = rep.rules["cap_result>1500tok"]
    assert stat.fires == 1
    assert stat.saved_tokens > 0


def test_combined_never_exceeds_total():
    rep = _report("fixture_bigread.jsonl")
    assert 0 < rep.combined_saved_tokens <= rep.total_result_tokens
