"""Table tests for each rule's transform. No tiktoken needed — rules are
pure string transforms (CapResult takes an injected tokenizer)."""

import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from diet.reader import Result
from diet.rules import CapResult, DedupeReads, StripNoise, TruncateReads
from diet.tokens import WordTokenizer


def _read(content, path="/f.go", index=0, later=""):
    return Result(index=index, tool_name="Read", tool_input={"file_path": path},
                  content=content, later_text=later)


def _bash(content, index=0, later=""):
    return Result(index=index, tool_name="Bash", tool_input={"command": "x"},
                  content=content, later_text=later)


# --- TruncateReads ---------------------------------------------------------

def test_truncate_fires_beyond_n():
    content = "\n".join(f"line{i}" for i in range(500))
    out = TruncateReads(200).apply(_read(content))
    assert out.fired
    assert out.content.count("\n") < 500
    assert "truncated 300 lines beyond 200" in out.content
    assert "line499" in out.removed[0] and "line499" not in out.content


def test_truncate_noop_under_n():
    content = "\n".join(f"line{i}" for i in range(50))
    out = TruncateReads(200).apply(_read(content))
    assert not out.fired and out.content == content


def test_truncate_ignores_non_read():
    content = "\n".join(f"line{i}" for i in range(500))
    out = TruncateReads(200).apply(_bash(content))
    assert not out.fired


# --- DedupeReads -----------------------------------------------------------

def test_dedupe_stubs_identical_reread():
    r = DedupeReads()
    body = "package main\nfunc x() {}"
    first = r.apply(_read(body, path="/a.go", index=0))
    second = r.apply(_read(body, path="/a.go", index=1))
    assert not first.fired
    assert second.fired
    assert "unchanged since read #0" in second.content
    assert second.removed[0] == body


def test_dedupe_keeps_changed_content():
    r = DedupeReads()
    r.apply(_read("v1", path="/a.go", index=0))
    out = r.apply(_read("v2 changed", path="/a.go", index=1))
    assert not out.fired  # different content at same path is not redundant


def test_dedupe_per_path():
    r = DedupeReads()
    r.apply(_read("same", path="/a.go", index=0))
    out = r.apply(_read("same", path="/b.go", index=1))
    assert not out.fired  # different path, first sighting


# --- StripNoise ------------------------------------------------------------

def test_strip_removes_ansi():
    out = StripNoise().apply(_bash("\x1b[32mgreen\x1b[0m text"))
    assert out.fired and out.content == "green text"
    assert "\x1b" not in out.content


def test_strip_collapses_progress_redraws():
    out = StripNoise().apply(_bash("prog [   ]\rprog [=  ]\rprog [===] done"))
    assert out.content == "prog [===] done"


def test_strip_collapses_blank_runs():
    out = StripNoise().apply(_bash("a\n\n\n\n\nb"))
    assert out.content == "a\n\nb"


def test_strip_noop_on_clean_text():
    out = StripNoise().apply(_bash("clean single line"))
    assert not out.fired


# --- CapResult (injected tokenizer) ---------------------------------------

def test_cap_fires_over_budget():
    tok = WordTokenizer()
    content = " ".join(f"w{i}" for i in range(100))  # 100 "tokens"
    out = CapResult(10, tokenizer=tok).apply(_bash(content))
    assert out.fired
    assert len(tok.encode(out.content.split("\n")[0])) <= 11  # 10 + marker word
    assert "capped at 10 tokens, 90 dropped" in out.content


def test_cap_noop_under_budget():
    tok = WordTokenizer()
    out = CapResult(1000, tokenizer=tok).apply(_bash("a b c"))
    assert not out.fired
