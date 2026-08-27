"""Encoder correctness: every format round-trips every payload it covers, and
pinned fixtures keep their exact BPE token counts. This is the deterministic
core — no model is involved anywhere in this file.
"""

import os
import sys
import unittest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))

import corpus  # noqa: E402
import encoders as E  # noqa: E402
import scorer  # noqa: E402


class TestRoundTrip(unittest.TestCase):
    def test_every_format_round_trips_every_payload(self):
        for name, p in corpus.CORPUS:
            shape, sch, canon = p["type"], E.schema(p), E.canonical(p)
            for fmt in E.FORMAT_ORDER:
                if not E.applies(fmt, shape):
                    continue
                with self.subTest(payload=name, fmt=fmt):
                    back = E.decode(fmt, E.encode(fmt, p), shape, sch)
                    self.assertTrue(E.roundtrip_equal(back, canon),
                                    f"{fmt}/{name} did not round-trip")

    def test_int_vs_string_coercion_survives(self):
        # "40782"-style all-digit values come back as ints; "07:20" stays str.
        self.assertEqual(E.coerce("6"), 6)
        self.assertEqual(E.coerce("-3"), -3)
        self.assertEqual(E.coerce("07:20"), "07:20")
        self.assertEqual(E.coerce("d41ab9c"), "d41ab9c")

    def test_csv_tsv_only_apply_to_tables(self):
        self.assertFalse(E.applies("csv", "hierarchy"))
        self.assertTrue(E.applies("csv", "table"))
        self.assertTrue(E.applies("babel", "graph"))


class TestPinnedTokenCounts(unittest.TestCase):
    """Exact o200k_base counts for fixture payloads. A drift here means either
    the tokenizer version moved or an encoder changed — both worth catching."""

    PINS = {
        "deploy_record": {"json-compact": 45, "yaml": 47, "babel": 38, "babel-nl": 45},
        "flight_table": {"csv": 72, "tsv": 72, "toon": 82, "babel": 66, "babel-nl": 74},
        "org_chart": {"json-compact": 178, "yaml": 169, "toon": 108,
                      "babel": 119, "babel-nl": 103},
        "network_graph": {"babel": 28, "babel-nl": 28, "yaml": 54},
    }
    SPEC_PINS = {"toon": 71, "babel": 109, "babel-nl": 92,
                 "json-compact": 0, "yaml": 0}

    def test_pinned_payload_counts(self):
        by_name = dict(corpus.CORPUS)
        for name, fmts in self.PINS.items():
            p = by_name[name]
            for fmt, expected in fmts.items():
                with self.subTest(payload=name, fmt=fmt):
                    self.assertEqual(scorer.count(E.encode(fmt, p)), expected)

    def test_pinned_spec_costs(self):
        spec = scorer.spec_cost()
        for fmt, expected in self.SPEC_PINS.items():
            with self.subTest(fmt=fmt):
                self.assertEqual(spec[fmt], expected)

    def test_babel_nl_beats_babel_on_org_chart(self):
        # The headline claim, as an assertion: un-packing underscores wins.
        p = dict(corpus.CORPUS)["org_chart"]
        self.assertLess(scorer.count(E.encode("babel-nl", p)),
                        scorer.count(E.encode("babel", p)))


if __name__ == "__main__":
    unittest.main()
