"""Tests for the replay reader: pairing tool_use -> tool_result and
attributing the assistant text that follows each result."""

import json
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from diet.reader import read_session


def _write_session(tmp, objs):
    p = os.path.join(tmp, "s.jsonl")
    with open(p, "w") as fh:
        for o in objs:
            fh.write(json.dumps(o) + "\n")
    return p


def test_pairs_use_with_result_and_input(tmp_path):
    objs = [
        {"type": "assistant", "message": {"content": [
            {"type": "tool_use", "id": "t1", "name": "Read",
             "input": {"file_path": "/x.go"}}]}},
        {"type": "user", "message": {"content": [
            {"type": "tool_result", "tool_use_id": "t1", "content": "file body"}]}},
    ]
    s = read_session(_write_session(str(tmp_path), objs))
    assert len(s.results) == 1
    r = s.results[0]
    assert r.tool_name == "Read"
    assert r.file_path == "/x.go"
    assert r.content == "file body"


def test_later_text_only_captures_text_after_result(tmp_path):
    objs = [
        {"type": "assistant", "message": {"content": [
            {"type": "text", "text": "BEFORE the read"},
            {"type": "tool_use", "id": "t1", "name": "Read", "input": {}}]}},
        {"type": "user", "message": {"content": [
            {"type": "tool_result", "tool_use_id": "t1", "content": "body"}]}},
        {"type": "assistant", "message": {"content": [
            {"type": "text", "text": "AFTER the read"}]}},
    ]
    s = read_session(_write_session(str(tmp_path), objs))
    assert "AFTER the read" in s.results[0].later_text
    assert "BEFORE the read" not in s.results[0].later_text


def test_result_content_list_form_is_normalized(tmp_path):
    objs = [
        {"type": "assistant", "message": {"content": [
            {"type": "tool_use", "id": "t1", "name": "Bash", "input": {}}]}},
        {"type": "user", "message": {"content": [
            {"type": "tool_result", "tool_use_id": "t1",
             "content": [{"type": "text", "text": "chunk one "},
                         {"type": "text", "text": "chunk two"}]}]}},
    ]
    s = read_session(_write_session(str(tmp_path), objs))
    assert s.results[0].content == "chunk one chunk two"


def test_malformed_lines_are_skipped(tmp_path):
    p = os.path.join(str(tmp_path), "s.jsonl")
    with open(p, "w") as fh:
        fh.write("not json\n")
        fh.write(json.dumps({"type": "assistant", "message": {"content": [
            {"type": "tool_use", "id": "t1", "name": "Read", "input": {}}]}}) + "\n")
        fh.write(json.dumps({"type": "user", "message": {"content": [
            {"type": "tool_result", "tool_use_id": "t1", "content": "ok"}]}}) + "\n")
    s = read_session(p)
    assert len(s.results) == 1 and s.results[0].content == "ok"
