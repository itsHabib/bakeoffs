"""Tests for the gate loop: verify-or-retry-once-then-reject, and ledger math.

Uses CassetteModel so no ollama is required. Confirms the behavior the brief
demands: never silently pass unverified output; retry once with the reason;
second failure -> REJECT.
"""

import json
import os
import sys
import tempfile
import unittest

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from hack3_offload import gate, ledger  # noqa: E402
from hack3_offload.model import CassetteModel  # noqa: E402


class TestGate(unittest.TestCase):
    narrow_in = {"criterion": "go test files",
                 "paths": ["a.go", "a_test.go"]}

    def test_pass_first_attempt(self):
        m = CassetteModel(['{"matched": ["a_test.go"]}'])
        out = gate.run("narrow", self.narrow_in, m, live=False)
        self.assertEqual(out.verdict, "PASS")
        self.assertEqual(out.output, {"matched": ["a_test.go"]})
        self.assertEqual(out.record.attempts, 1)

    def test_retry_then_pass(self):
        # first answer wrong (includes non-test file), second answer correct.
        m = CassetteModel([
            '{"matched": ["a.go", "a_test.go"]}',
            '{"matched": ["a_test.go"]}',
        ])
        out = gate.run("narrow", self.narrow_in, m, live=False)
        self.assertEqual(out.verdict, "PASS")
        self.assertEqual(out.record.attempts, 2)

    def test_reject_after_two_failures(self):
        m = CassetteModel([
            '{"matched": ["a.go", "a_test.go"]}',
            '{"matched": ["a.go", "a_test.go"]}',
        ])
        out = gate.run("narrow", self.narrow_in, m, live=False)
        self.assertEqual(out.verdict, "REJECT")
        self.assertIsNone(out.output)
        self.assertIn("do not match", out.reason)
        self.assertEqual(out.record.attempts, 2)

    def test_invalid_json_is_a_failure(self):
        m = CassetteModel(["not json at all", "still not json"])
        out = gate.run("narrow", self.narrow_in, m, live=False)
        self.assertEqual(out.verdict, "REJECT")
        self.assertIn("not valid JSON", out.reason)

    def test_ledger_counts_only_pass_as_displaced(self):
        path = tempfile.mktemp(suffix=".jsonl")
        try:
            good = CassetteModel(['{"matched": ["a_test.go"]}'])
            bad = CassetteModel(['{"matched": ["a.go"]}', '{"matched": ["a.go"]}'])
            p = gate.run("narrow", self.narrow_in, good, live=False)
            r = gate.run("narrow", self.narrow_in, bad, live=False)
            ledger.append(path, p.record)
            ledger.append(path, r.record)

            roll = ledger.rollup(path)["narrow"]
            self.assertEqual(roll.calls, 2)
            self.assertEqual(roll.passed, 1)
            self.assertEqual(roll.rejected, 1)
            # displaced tokens come only from the PASS record.
            self.assertEqual(roll.displaced_tokens, p.record.frontier_tokens_est)
            self.assertGreater(roll.displaced_tokens, 0)
        finally:
            if os.path.exists(path):
                os.remove(path)

    def test_reject_record_has_reason(self):
        path = tempfile.mktemp(suffix=".jsonl")
        try:
            bad = CassetteModel(['{"matched": ["a.go"]}', '{"matched": ["a.go"]}'])
            out = gate.run("narrow", self.narrow_in, bad, live=False)
            ledger.append(path, out.record)
            with open(path) as f:
                rec = json.loads(f.read().strip())
            self.assertEqual(rec["verdict"], "REJECT")
            self.assertIsNotNone(rec["reject_reason"])
        finally:
            if os.path.exists(path):
                os.remove(path)


if __name__ == "__main__":
    unittest.main()
