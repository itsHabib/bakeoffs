"""Table tests for the deterministic verifiers — the graded correctness path.

Run: .venv/bin/python -m pytest tests/ -q
 or: .venv/bin/python -m unittest discover -s tests -q   (stdlib, no deps)

Each verifier is exercised on: the happy path, structural violations, and the
adversarial shape that the demo replays. If any of these regress, the gate can
no longer be trusted, so these are the tests the entry is judged on.
"""

import json
import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from hack3_offload.verifiers import (  # noqa: E402
    verify_classify,
    verify_extract,
    verify_narrow,
)

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def load_input(task, name):
    with open(os.path.join(REPO, "fixtures", task, f"{name}.input.json")) as f:
        return json.load(f)


def cassette_first(task, name):
    with open(os.path.join(REPO, "cassettes", task, f"{name}.cassette.json")) as f:
        return json.loads(json.load(f)["responses"][0])


class TestNarrow(unittest.TestCase):
    CASES = [
        # name, input, output, want_ok, reason_substr
        (
            "exact test files",
            {"criterion": "go test files",
             "paths": ["a.go", "a_test.go", "b_test.go"]},
            {"matched": ["a_test.go", "b_test.go"]},
            True, None,
        ),
        (
            "invented path (not a subset)",
            {"criterion": "go test files", "paths": ["a_test.go"]},
            {"matched": ["a_test.go", "ghost_test.go"]},
            False, "not a subset",
        ),
        (
            "wrongly included non-match",
            {"criterion": "markdown files", "paths": ["a.md", "b.go"]},
            {"matched": ["a.md", "b.go"]},
            False, "do not match",
        ),
        (
            "wrongly excluded a match",
            {"criterion": "markdown files", "paths": ["a.md", "b.md"]},
            {"matched": ["a.md"]},
            False, "missed items",
        ),
        (
            "missing matched key",
            {"criterion": "markdown files", "paths": ["a.md"]},
            {"result": ["a.md"]},
            False, "matched",
        ),
        (
            "non-syntactic criterion cannot be verified",
            {"criterion": "files that look important", "paths": ["a.go"]},
            {"matched": ["a.go"]},
            False, "not syntactic",
        ),
        (
            "explicit regex override",
            {"criterion": "anything", "spot_check_regex": r"\.py$",
             "paths": ["x.py", "y.go"]},
            {"matched": ["x.py"]},
            True, None,
        ),
    ]

    def test_table(self):
        for name, inp, out, want_ok, sub in self.CASES:
            with self.subTest(name=name):
                ok, reason = verify_narrow(inp, out)
                self.assertEqual(ok, want_ok, f"{name}: reason={reason}")
                if sub:
                    self.assertIn(sub, reason)

    def test_adversarial_fixture_rejects(self):
        inp = load_input("narrow", "adversarial_cmd_dir")
        out = cassette_first("narrow", "adversarial_cmd_dir")
        ok, reason = verify_narrow(inp, out)
        self.assertFalse(ok)
        self.assertIn("docs/cmd-notes.md", reason)

    def test_clean_fixtures_pass(self):
        for name in ("clean_test_files", "clean_markdown"):
            inp = load_input("narrow", name)
            out = cassette_first("narrow", name)
            ok, reason = verify_narrow(inp, out)
            self.assertTrue(ok, f"{name}: {reason}")


class TestExtract(unittest.TestCase):
    CASES = [
        (
            "all values verbatim",
            {"schema": {"v": "string", "n": "int"}, "text": "ver=1.2.3 count=7"},
            {"v": "1.2.3", "n": 7},
            True, None,
        ),
        (
            "hallucinated string value",
            {"schema": {"v": "string"}, "text": "no version here"},
            {"v": "9.9.9"},
            False, "hallucinated",
        ),
        (
            "hallucinated number value",
            {"schema": {"n": "int"}, "text": "count=7"},
            {"n": 42},
            False, "not found verbatim",
        ),
        (
            "wrong type",
            {"schema": {"n": "int"}, "text": "n=7"},
            {"n": "7"},
            False, "expected type int",
        ),
        (
            "missing field",
            {"schema": {"a": "string", "b": "string"}, "text": "a=x b=y"},
            {"a": "x"},
            False, "missing fields",
        ),
        (
            "extra field",
            {"schema": {"a": "string"}, "text": "a=x"},
            {"a": "x", "b": "y"},
            False, "unexpected fields",
        ),
        (
            "bool is not int",
            {"schema": {"n": "int"}, "text": "True"},
            {"n": True},
            False, "expected type int",
        ),
    ]

    def test_table(self):
        for name, inp, out, want_ok, sub in self.CASES:
            with self.subTest(name=name):
                ok, reason = verify_extract(inp, out)
                self.assertEqual(ok, want_ok, f"{name}: reason={reason}")
                if sub:
                    self.assertIn(sub, reason)

    def test_adversarial_fixture_rejects(self):
        inp = load_input("extract", "adversarial_docker_digest")
        out = cassette_first("extract", "adversarial_docker_digest")
        ok, reason = verify_extract(inp, out)
        self.assertFalse(ok)
        self.assertIn("digest", reason)
        self.assertIn("hallucinated", reason)

    def test_clean_fixtures_pass(self):
        for name in ("clean_gotest", "clean_npm"):
            inp = load_input("extract", name)
            out = cassette_first("extract", name)
            ok, reason = verify_extract(inp, out)
            self.assertTrue(ok, f"{name}: {reason}")


class TestClassify(unittest.TestCase):
    CASES = [
        (
            "all correct with planted",
            {"labels": ["A", "B"], "lines": ["x", "y"], "planted": {"0": "A"}},
            {"assignments": ["A", "B"]},
            True, None,
        ),
        (
            "label outside set",
            {"labels": ["A", "B"], "lines": ["x"], "planted": {}},
            {"assignments": ["C"]},
            False, "outside the allowed set",
        ),
        (
            "wrong length",
            {"labels": ["A"], "lines": ["x", "y"], "planted": {}},
            {"assignments": ["A"]},
            False, "expected 2 assignments",
        ),
        (
            "planted ground truth violated",
            {"labels": ["A", "B"], "lines": ["x", "y"], "planted": {"1": "A"}},
            {"assignments": ["A", "B"]},
            False, "planted line 1",
        ),
        (
            "missing assignments key",
            {"labels": ["A"], "lines": ["x"], "planted": {}},
            {"labels": ["A"]},
            False, "assignments",
        ),
    ]

    def test_table(self):
        for name, inp, out, want_ok, sub in self.CASES:
            with self.subTest(name=name):
                ok, reason = verify_classify(inp, out)
                self.assertEqual(ok, want_ok, f"{name}: reason={reason}")
                if sub:
                    self.assertIn(sub, reason)

    def test_adversarial_fixture_rejects(self):
        inp = load_input("classify", "adversarial_security")
        out = cassette_first("classify", "adversarial_security")
        ok, reason = verify_classify(inp, out)
        self.assertFalse(ok)
        self.assertIn("planted line 0", reason)
        self.assertIn("SECURITY", reason)

    def test_clean_fixtures_pass(self):
        for name in ("clean_logs", "clean_gotest"):
            inp = load_input("classify", name)
            out = cassette_first("classify", name)
            ok, reason = verify_classify(inp, out)
            self.assertTrue(ok, f"{name}: {reason}")


if __name__ == "__main__":
    unittest.main()
