"""The grader is policy — it decides right/wrong, so it is table-tested here.
The model never grades itself; this function does, deterministically."""

import os
import sys
import unittest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))

import comprehension as C  # noqa: E402


class TestGrade(unittest.TestCase):
    def test_numeric_word_boundary(self):
        # "6" must not match inside "16" or "60".
        self.assertTrue(C.grade("6", "6 replicas", True))
        self.assertTrue(C.grade("6", "the answer is 6.", True))
        self.assertFalse(C.grade("6", "16", True))
        self.assertFalse(C.grade("6", "60", True))

    def test_string_substring_after_normalization(self):
        self.assertTrue(C.grade("Theo Vance", "Theo Vance (VP Engineering)", False))
        self.assertTrue(C.grade("Theo Vance", "  theo   vance  ", False))
        self.assertFalse(C.grade("Theo Vance", "Lucia Menon", False))

    def test_detail_with_digits_is_string_match(self):
        self.assertTrue(C.grade("2 of 314 failed",
                                "The detail is: 2 of 314 failed.", False))

    def test_yes_no(self):
        self.assertTrue(C.grade("yes", "Yes.", False))
        self.assertFalse(C.grade("yes", "no", False))


if __name__ == "__main__":
    unittest.main()
