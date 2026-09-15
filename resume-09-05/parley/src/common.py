"""Local journal and content contract. No external service or credential access."""
import contextlib
import fcntl
import hashlib
import json
import os
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SOURCES = {"S1": {"revision": "S1", "numbers": [2, 3, 5]},
           "S2": {"revision": "S2", "numbers": [7, 11]}}


class Refusal(Exception):
    pass


def encoded(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":")).encode()


def digest(data):
    return hashlib.sha256(data).hexdigest()


def identity(value):
    return digest(encoded(value))


def definition():
    return digest((ROOT / "protocol.parley").read_bytes() +
                  b"square-list-v1;journal-v1;delivery-v1;checker-v1")


def durable_write(path, data):
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    temp = path.with_suffix(path.suffix + ".tmp")
    with temp.open("wb") as stream:
        stream.write(data)
        stream.flush()
        os.fsync(stream.fileno())
    os.replace(temp, path)


def save_json(path, value):
    durable_write(path, encoded(value) + b"\n")


@contextlib.contextmanager
def locked(directory):
    with (directory / "worker.lock").open("a") as stream:
        try:
            fcntl.flock(stream, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError as exc:
            raise Refusal("worker_already_active") from exc
        yield


class Journal:
    def __init__(self, directory, expected_definition=None):
        self.directory = Path(directory)
        self.path = self.directory / "journal.jsonl"
        self.events = []
        data = self.path.read_bytes() if self.path.exists() else b""
        prefix_end = data.rfind(b"\n") + 1
        self.tail = data[prefix_end:]
        for number, line in enumerate(data[:prefix_end].splitlines(), 1):
            try:
                event = json.loads(line)
                checksum = event.pop("hash")
                assert checksum == identity(event)
                assert event["seq"] == number
                assert event["previous"] == (self.events[-1]["hash"] if self.events else "0" * 64)
                assert event["id"] == f"event-{number}"
                assert isinstance(event["kind"], str) and isinstance(event["data"], dict)
                event["hash"] = checksum
                self.events.append(event)
            except (ValueError, KeyError, TypeError, AssertionError, AttributeError) as exc:
                raise Refusal(f"corrupt_complete_record:{number}") from exc
        if self.events:
            first = self.events[0]
            if first["kind"] != "opened":
                raise Refusal("missing_definition_record")
            expected = expected_definition or definition()
            if first["data"]["definition"] != expected:
                raise Refusal("definition_mismatch")
            if first["data"]["sources"] != SOURCES:
                raise Refusal("input_definition_mismatch")
        # Validate the complete prefix BEFORE changing even an incomplete tail.
        if self.tail:
            durable_write(self.directory / "discarded-tail.bin", self.tail)
            with self.path.open("r+b") as stream:
                stream.truncate(prefix_end)
                stream.flush()
                os.fsync(stream.fileno())

    def append(self, kind, **data):
        number = len(self.events) + 1
        event = {"seq": number, "id": f"event-{number}", "kind": kind, "data": data,
                 "previous": self.events[-1]["hash"] if self.events else "0" * 64}
        event["hash"] = identity(event)
        with self.path.open("ab") as stream:
            stream.write(encoded(event) + b"\n")
            stream.flush()
            os.fsync(stream.fileno())
        self.events.append(event)
        return event

    def latest(self, kind, source=None):
        return next((e["data"] for e in reversed(self.events)
                     if e["kind"] == kind and
                     (source is None or e["data"].get("source") == source)), None)


def candidate(source):
    return {"source": identity(source), "squares": [n * n for n in source["numbers"]]}


def check(content, source):
    """Same actual-byte checker for the candidate and planted negative control."""
    try:
        actual = json.loads(content)
    except (ValueError, UnicodeDecodeError):
        return False
    return actual == candidate(source)


def check_candidate_binding(content, source, evidence):
    if digest(content) != evidence["candidate"] or not check(content, source):
        raise Refusal("candidate_changed_after_check")
    if evidence["source"] != identity(source):
        raise Refusal("evidence_subject_mismatch")


def check_incarnation(terminal, current):
    if terminal["incarnation"] != current["incarnation"]:
        raise Refusal("old_incarnation_terminal")


def binding_evidence(source, content):
    return {"source": identity(source), "candidate": digest(content),
            "checker": "square-list-v1"}


def delivery(run, source, content, evidence):
    return {"run": run, "step": "deliver", "source": identity(source),
            "candidate": digest(content), "evidence": identity(evidence),
            "content": content.decode()}


def operation(payload):
    # Run, step, subject, candidate and evidence all contribute; incarnation does not.
    return "delivery-" + identity(payload)
