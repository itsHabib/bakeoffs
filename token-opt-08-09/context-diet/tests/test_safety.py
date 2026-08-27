"""Tests for the reference-detection logic — the computed safety score."""

import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from diet.safety import SpanScore, salient_tokens, score_span


def test_salient_extracts_identifiers_not_prose():
    toks = salient_tokens("call handleAuthRetry_v2 at server.go line 350")
    assert "handleAuthRetry_v2" in toks
    assert "server.go" in toks
    # plain prose words carry no reference signal and are dropped
    assert "call" not in toks and "line" not in toks


def test_referenced_when_symbol_reappears_later():
    removed = "func handleAuthRetry_v2(ctx Context) error {"
    later = "I should call handleAuthRetry_v2 from the auth path."
    sc = score_span(removed, later)
    assert sc.referenced
    assert "handleAuthRetry_v2" in sc.hits


def test_not_referenced_when_symbol_absent():
    removed = "func handleAuthRetry_v2(ctx Context) error {"
    later = "Build passed, nothing else to do here."
    sc = score_span(removed, later)
    assert not sc.referenced and sc.hits == ()


def test_ansi_residue_scores_zero_salient():
    # what StripNoise removes (escape codes / whitespace) has no identifiers
    sc = score_span("\x1b[32m \x1b[0m \r \n\n", "later text mentions nothing")
    assert sc.salient == 0 and not sc.referenced


def test_token_surviving_in_kept_is_not_counted_lost():
    # identifier appears in the removed tail AND in what we kept above the
    # cut — truncating the tail did not cost the agent access to it
    removed = "call handleAuthRetry_v2 again down here"
    kept = "func handleAuthRetry_v2(ctx Context) error {"
    later = "later I use handleAuthRetry_v2"
    sc = score_span(removed, later, kept=kept)
    assert not sc.referenced  # not lost, so not risky


def test_partial_word_is_not_a_false_hit():
    # substring match is intentional but salient tokens are long+codey, so a
    # short common word can't sneak a hit
    sc = score_span("tmpBuffer_42", "the buffer was fine")
    assert not sc.referenced
