#!/usr/bin/env python3
"""Table tests for the deterministic core. Exact numbers, no model, no network.

The expected sums below are hand-computed from fixtures/_gen.py — the whole
point of the fixtures is that a human can add the usage tuples up and check.
"""
import os
import unittest

import audit

HERE = os.path.dirname(__file__)
FIX = os.path.join(HERE, "fixtures")


class LenEstimator:
    """Deterministic stand-in for tiktoken: one 'token' per character."""

    method = "len (test)"

    def count(self, text: str) -> int:
        return len(text)


def load():
    paths = audit.collect_paths(FIX)
    usage, volumes, stats = audit.parse(paths, LenEstimator())
    return usage, volumes, stats, audit.aggregate(usage, volumes)


class ParseStats(unittest.TestCase):
    def test_counts(self):
        _, _, stats, _ = load()
        self.assertEqual(stats.files, 5)
        self.assertEqual(stats.usage_msgs, 8)  # a1,a2,a3,b1,c1,c2,d1,e1
        self.assertEqual(stats.dup_msgs, 2)    # a2 identical + a3 stub partial
        self.assertEqual(stats.tool_results, 5)
        self.assertEqual(stats.orphan_tool_results, 1)


class Aggregation(unittest.TestCase):
    def setUp(self):
        _, _, _, self.agg = load()

    def test_by_category(self):
        t = self.agg.by_category.tokens
        self.assertEqual(t["input"], 950)
        self.assertEqual(t["cache_read"], 79000)
        self.assertEqual(t["cache_write_5m"], 8500)
        self.assertEqual(t["cache_write_1h"], 1500)
        self.assertEqual(t["output"], 7900)  # a3 kept at 500, not the 2 stub
        self.assertEqual(self.agg.by_category.total_tokens, 97850)

    def test_by_day(self):
        d = {k: v.total_tokens for k, v in self.agg.by_day.items()}
        self.assertEqual(d["2026-08-01"], 30210)
        self.assertEqual(d["2026-08-02"], 37820)
        self.assertEqual(d["2026-08-03"], 29820)

    def test_by_project_rolls_up_worktrees(self):
        p = {k: v.total_tokens for k, v in self.agg.by_project.items()}
        # session-b's worktree cwd must fold into /home/dev/alpha
        self.assertEqual(p["/home/dev/alpha"], 54910)
        self.assertEqual(p["/home/dev/beta"], 23140)
        self.assertEqual(p["/home/dev/gamma"], 19800)
        self.assertNotIn("/home/dev/alpha/.claude/worktrees/foo-123", p)

    def test_by_model(self):
        m = {k: v.total_tokens for k, v in self.agg.by_model.items()}
        self.assertEqual(m["claude-opus-5"], 54910)
        self.assertEqual(m["claude-fable-5"], 23140)
        self.assertEqual(m["claude-opus-4-8"], 19800)

    def test_by_session(self):
        s = {k: v.total_tokens for k, v in self.agg.by_session.items()}
        self.assertEqual(s["session-a"], 30210)
        self.assertEqual(s["session-b"], 24700)
        self.assertEqual(s["session-e"], 19800)

    def test_tool_volume_attribution(self):
        tv = self.agg.tool_volume
        self.assertEqual(tv["Bash"], 10)  # "AAAA"(4) + "CCCCCC"(6)
        self.assertEqual(tv["Read"], 8)   # "BBBBBBBB"
        self.assertEqual(tv["Grep"], 2)   # list block text "DD"
        self.assertEqual(tv["(unknown tool)"], 5)  # orphan "EEEEE"

    def test_totals_are_consistent(self):
        # every breakdown must sum to the same grand total
        total = self.agg.by_category.total_tokens
        for groups in (self.agg.by_day, self.agg.by_project,
                       self.agg.by_session, self.agg.by_model):
            self.assertEqual(sum(s.total_tokens for s in groups.values()), total)


class Pricing(unittest.TestCase):
    def test_unfilled_table_is_detected(self):
        prices = audit.load_prices(os.path.join(HERE, "prices.toml"))
        self.assertFalse(audit.prices_filled(prices))

    def test_filled_table_is_detected(self):
        self.assertTrue(audit.prices_filled({"m": {"input": 1.0}}))

    def test_per_model_dollar_math(self):
        usage, volumes, _, agg = load()
        # only opus-5 cache_read is priced, at $2/MTok; everything else free.
        prices = {
            "claude-opus-5": {"cache_read": 2.0},
            "__default__": {},
        }
        audit.price(agg, usage, prices)
        # opus-5 cache_read tokens = 10000+12000+3000+20000 = 45000 -> *2/1e6
        self.assertAlmostEqual(agg.by_model["claude-opus-5"].dollars["cache_read"],
                               45000 * 2 / 1_000_000)
        self.assertAlmostEqual(agg.by_model["claude-fable-5"].total_dollars, 0.0)
        # category total dollars equals the one priced slice
        self.assertAlmostEqual(agg.by_category.total_dollars, 0.090)


if __name__ == "__main__":
    unittest.main()
