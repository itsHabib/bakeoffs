"""Process-crash controller and independent, artifact/effect-based assertions."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import queue
import shlex
import sqlite3
import subprocess
import sys
import tempfile
import threading

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "src"))
import audit


def wire(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":")).encode()


def sha(data):
    return hashlib.sha256(data).hexdigest()


def save(path, value):
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n")


def records(directory):
    return [json.loads(line) for line in (directory / "journal.jsonl").read_text().splitlines()]


def latest(directory, kind, source=None):
    return next((e["data"] for e in reversed(records(directory)) if e["kind"] == kind and
                 (source is None or e["data"].get("source") == source)), None)


def effects(directory):
    # Controller-only ground truth. The worker never opens this database.
    path = directory / "sink/private.sqlite"
    if not path.exists():
        return []
    with sqlite3.connect(f"file:{path}?mode=ro", uri=True) as db:
        return [{"id": row[0], "mode": row[1], "op": row[2], "digest": row[3],
                 "payload": json.loads(row[4]), "receipt": json.loads(row[5])}
                for row in db.execute("SELECT * FROM effects ORDER BY id")]


def independent_completion(directory, revision="S1", count=1):
    """Does not trust outcome labels, executor checker, key helper, or protocol order."""
    inputs = {"S1": {"revision": "S1", "numbers": [2, 3, 5]},
              "S2": {"revision": "S2", "numbers": [7, 11]}}
    source = inputs[revision]
    subject = sha(wire(source))
    content = (directory / f"candidate-{subject}.json").read_bytes()
    parsed = json.loads(content)
    assert parsed == {"source": subject, "squares": [x ** 2 for x in source["numbers"]]}, "candidate semantics"
    checked = latest(directory, "candidate.checked", subject)
    evidence = checked["evidence"]
    assert evidence == {"source": subject, "candidate": sha(content), "checker": "square-list-v1"}, "stale evidence"
    assert checked["evidence_digest"] == sha(wire(evidence)), "evidence digest"
    negative = latest(directory, "negative.detected", subject)
    defect = (directory / negative["path"]).read_bytes()
    assert sha(defect) == negative["digest"], "negative bytes identity"
    assert json.loads(defect) != parsed, "undetected negative"
    assert json.loads(defect)["squares"][0] == parsed["squares"][0] + 1, "wrong planted defect"
    rows = effects(directory)
    assert len(rows) == count, f"effect count {len(rows)}, expected {count}"
    matching = [row for row in rows if row["payload"]["source"] == subject]
    assert len(matching) == 1, "subject effect count"
    row = matching[0]
    payload = row["payload"]
    assert payload == {"run": "run-demo", "step": "deliver", "source": subject,
                       "candidate": sha(content), "evidence": sha(wire(evidence)),
                       "content": content.decode()}, "effect payload identity"
    assert row["digest"] == sha(wire(payload)), "sink payload digest"
    assert row["op"] == "delivery-" + sha(wire(payload)), "stable operation key"
    assert row["receipt"] == {"effect_id": row["id"], "op": row["op"],
                              "payload_digest": row["digest"]}, "sink receipt identity"
    completed = latest(directory, "completed", subject)
    assert completed, "no named completion"
    assert completed["run"] == "run-demo" and completed["source"] == subject, "terminal subject"
    assert completed["candidate"] == sha(content) and completed["evidence"] == sha(wire(evidence)), "terminal evidence"
    assert completed["receipt"] == row["receipt"], "terminal receipt"
    terminal_event = next(e for e in reversed(records(directory)) if e["kind"] == "completed" and e["data"]["source"] == subject)
    prior = [e["data"] for e in records(directory) if e["kind"] == "incarnation" and e["seq"] < terminal_event["seq"]]
    assert completed["incarnation"] == prior[-1]["incarnation"], "stale incarnation advanced run"


class Controller:
    def __init__(self, base):
        self.base = base
        self.results = []
        self.invocations = {}

    def case(self, name, expectation):
        directory = self.base / name
        directory.mkdir()
        save(directory / "expected.json", expectation)
        return directory

    def command(self, directory, mode="queryable", source="S1", pause="", mutant=None, message=None):
        command = [sys.executable, str(ROOT / "src/worker.py")]
        if mutant:
            command = [sys.executable, str(ROOT / "tests/mutant_worker.py"), mutant]
        command += [str(directory), "--mode", mode, "--source", source]
        if pause:
            command += ["--pause", pause]
        if message:
            command += ["--message", str(message)]
        with (directory / "commands.txt").open("a") as stream:
            stream.write(shlex.join(command) + (f" # controller SIGKILL at {pause}" if pause else "") + "\n")
        return command

    def run(self, directory, **options):
        command = self.command(directory, **options)
        number = self.invocations.get(directory, 0) + 1
        self.invocations[directory] = number
        if options.get("pause"):
            process = subprocess.Popen(command, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                       stderr=subprocess.STDOUT, text=True)
            lines = []
            inbox = queue.Queue()

            def read():
                for line in process.stdout:
                    inbox.put(line)
                inbox.put(None)

            thread = threading.Thread(target=read, daemon=True)
            thread.start()
            try:
                while True:
                    line = inbox.get(timeout=15)
                    assert line is not None, "worker exited before handshake"
                    lines.append(line)
                    event = json.loads(line)
                    if event.get("boundary") == options["pause"]:
                        break
                process.kill()  # real SIGKILL, only this controller's worker PID
                process.wait(timeout=10)
                process.stdin.close()
                thread.join(timeout=10)
                assert not thread.is_alive(), "sink child failed to release pipe"
                assert process.returncode == -9
                result = {"outcome": "killed", "signal": 9, "pid": process.pid,
                          "boundary": options["pause"]}
            finally:
                if process.poll() is None:
                    process.kill()
                    process.wait(timeout=10)
        else:
            response = subprocess.run(command, text=True, capture_output=True, timeout=20)
            lines = response.stdout.splitlines(keepends=True)
            assert lines, response.stderr
            result = json.loads(lines[-1])
            assert response.returncode in (0, 2), response.stderr
        (directory / f"worker-{number}.stdout").write_text("".join(lines))
        save(directory / f"worker-{number}.result.json", result)
        return result

    def end(self, directory, outcome, observer=True, mutant=None):
        observed = effects(directory)
        save(directory / "controller-effects.json", observed)
        report = audit.audit(directory) if observer else None
        if report:
            expected = "complete"
            if outcome in ("unresolved", "refused"):
                expected = "stalled"
            if outcome == "observer_deviation_only":
                expected = "deviating"
            assert all(item["classification"] == expected for item in report["subjects"].values()), "unexpected protocol classification"
            raw = [item["raw"] for item in report["excluded"]]
            for subject in report["subjects"]:
                mapped = json.loads((directory / "audit" / (subject + ".identity-map.json")).read_text())
                raw.extend(item["raw"] for item in mapped)
            assert sorted(raw, key=lambda event: event["seq"]) == records(directory), "normalization lost raw identity"
        entry = {"case": directory.name, "passed": True, "effects": len(observed),
                 "outcome": outcome, "mutant_detected": mutant,
                 "protocol": {s: r["classification"] for s, r in report["subjects"].items()} if report else "refused_before_audit"}
        save(directory / "actual.json", entry)
        self.results.append(entry)
        print(f"PASS {directory.name:37} effects={len(observed)} outcome={outcome}", flush=True)


def prepare_old_terminal(controller, directory):
    controller.run(directory, pause="receipt_durable")
    receipt = latest(directory, "delivery.receipt")
    evidence = latest(directory, "candidate.checked")["evidence"]
    old = {"run": "run-demo", "source": evidence["source"], "candidate": evidence["candidate"],
           "evidence": sha(wire(evidence)), "incarnation": 1, "receipt": receipt["receipt"]}
    save(directory / "late-terminal.json", old)
    controller.run(directory, pause="incarnation_durable")
    return directory / "late-terminal.json"


def mutate_journal(directory, transform):
    events = records(directory)
    transform(events)
    previous = "0" * 64
    for index, event in enumerate(events, 1):
        event.update(seq=index, id=f"event-{index}", previous=previous)
        event.pop("hash", None)
        event["hash"] = sha(wire(event))
        previous = event["hash"]
    (directory / "journal.jsonl").write_bytes(b"".join(wire(e) + b"\n" for e in events))


def all_cases(controller):
    for mode in ("queryable", "opaque"):
        d = controller.case(f"good-{mode}", {"outcome": "completed", "effects": 1})
        assert controller.run(d, mode=mode)["outcome"] == "completed"
        independent_completion(d)
        controller.end(d, "completed")

    for mode in ("queryable", "opaque"):
        for cut in ("intent_durable", "sink_committed_reply_withheld", "receipt_durable"):
            unresolved = mode == "opaque" and cut != "receipt_durable"
            expected_effects = 0 if mode == "opaque" and cut == "intent_durable" else 1
            d = controller.case(f"crash-{mode}-{cut}", {"outcome": "unresolved" if unresolved else "completed", "effects": expected_effects})
            controller.run(d, mode=mode, pause=cut)
            before = effects(d)
            assert len(before) == (0 if cut == "intent_durable" else 1)
            result = controller.run(d, mode=mode)
            assert len(effects(d)) == expected_effects
            if unresolved:
                assert result["outcome"] == "unresolved"
                assert result["op"] == latest(d, "delivery.intent")["op"]
                assert latest(d, "completed") is None
                # Another restart must also remain parked without repeating.
                assert controller.run(d, mode=mode)["outcome"] == "unresolved"
                assert effects(d) == before
            else:
                assert result["outcome"] == "completed"
                independent_completion(d)
                if before:
                    assert effects(d) == before, "repeated effect"
                    assert result["receipt"] == before[0]["receipt"], "receipt not retained"
            controller.end(d, result["outcome"])

    d = controller.case("late-terminal", {"outcome": "refused_then_completed", "effects": 1})
    message = prepare_old_terminal(controller, d)
    result = controller.run(d, message=message)
    assert result == {"outcome": "refused", "reason": "old_incarnation_terminal"}
    assert latest(d, "completed") is None
    assert latest(d, "terminal.rejected")["raw"] == json.loads(message.read_text())
    controller.run(d)
    independent_completion(d)
    controller.end(d, "refused_then_completed")

    d = controller.case("source-S2", {"outcome": "fresh_S2_completed", "effects": 2})
    controller.run(d)
    original = effects(d)[0]
    old_terminal = latest(d, "completed")
    save(d / "S1-terminal.json", old_terminal)
    controller.run(d, source="S2", pause="receipt_durable")
    # Give the stale source message the CURRENT incarnation to isolate subject fencing.
    old_terminal["incarnation"] = latest(d, "incarnation")["incarnation"]
    save(d / "S1-terminal.json", old_terminal)
    result = controller.run(d, source="S2", message=d / "S1-terminal.json")
    assert result == {"outcome": "refused", "reason": "terminal_source_mismatch"}
    # Relabeling the subject cannot turn S1 evidence into certification of S2.
    s2_checked = latest(d, "candidate.checked")["evidence"]
    transplanted = {**old_terminal, "source": s2_checked["source"],
                    "candidate": s2_checked["candidate"],
                    "receipt": latest(d, "delivery.receipt")["receipt"]}
    save(d / "transplanted-evidence-terminal.json", transplanted)
    result = controller.run(d, source="S2", message=d / "transplanted-evidence-terminal.json")
    assert result == {"outcome": "refused", "reason": "terminal_evidence_mismatch"}
    controller.run(d, source="S2")
    independent_completion(d, "S1", count=2)
    independent_completion(d, "S2", count=2)
    assert effects(d)[0] == original, "historical S1 effect changed"
    assert latest(d, "candidate.checked", old_terminal["source"])["evidence_digest"] != latest(d, "candidate.checked")["evidence_digest"]
    controller.end(d, "fresh_S2_completed")

    d = controller.case("changed-candidate", {"outcome": "refused", "effects": 0})
    controller.run(d, pause="checks_durable")
    path = next(d.glob("candidate-*.json"))
    changed = json.loads(path.read_text())
    changed["squares"][0] += 8
    path.write_bytes(wire(changed))
    result = controller.run(d)
    assert result == {"outcome": "refused", "reason": "candidate_changed_after_check"}
    assert not effects(d) and latest(d, "completed") is None
    controller.end(d, "refused")

    for mode in ("queryable", "opaque"):
        d = controller.case(f"truncated-tail-{mode}", {"outcome": "completed" if mode == "queryable" else "unresolved", "effects": 1})
        controller.run(d, mode=mode, pause="sink_committed_reply_withheld")
        prefix = (d / "journal.jsonl").read_bytes()
        tail = b'{"seq":99,"kind":"delivery.receipt"'
        with (d / "journal.jsonl").open("ab") as stream:
            stream.write(tail)
            stream.flush()
            os.fsync(stream.fileno())
        result = controller.run(d, mode=mode)
        assert (d / "discarded-tail.bin").read_bytes() == tail
        assert (d / "journal.jsonl").read_bytes().startswith(prefix)
        assert len(effects(d)) == 1
        if mode == "queryable":
            independent_completion(d)
        else:
            assert result["outcome"] == "unresolved" and latest(d, "completed") is None
        controller.end(d, result["outcome"])

    for corruption in ("complete-record", "definition-mismatch"):
        d = controller.case(corruption, {"outcome": "refused", "effects": 0})
        controller.run(d, pause="intent_durable")
        if corruption == "complete-record":
            with (d / "journal.jsonl").open("ab") as stream:
                stream.write(b'{"bad": true}\n')
        else:
            mutate_journal(d, lambda events: events[0]["data"].update(definition="different-definition"))
        damaged = (d / "journal.jsonl").read_bytes()
        result = controller.run(d)
        assert result["outcome"] == "refused"
        assert ("corrupt_complete_record" if corruption == "complete-record" else "definition_mismatch") in result["reason"]
        assert not effects(d) and damaged == (d / "journal.jsonl").read_bytes()
        controller.end(d, "refused", observer=False)

    d = controller.case("sink-dedup-and-conflict", {"outcome": "deduplicated_then_refused", "effects": 1})
    controller.run(d)
    original = effects(d)
    row = original[0]
    command = [sys.executable, str(ROOT / "src/sink.py"), str(d / "sink"), "queryable", "put"]
    request = {"op": row["op"], "payload": row["payload"], "payload_digest": row["digest"]}
    save(d / "sink-repeat-request.json", request)
    response = subprocess.run(command, input=json.dumps(request) + "\nrelease\n", capture_output=True, text=True, timeout=10)
    assert json.loads(response.stdout.splitlines()[-1])["receipt"] == row["receipt"]
    request["payload"] = {**row["payload"], "candidate": "changed"}
    request["payload_digest"] = sha(wire(request["payload"]))
    save(d / "sink-conflict-request.json", request)
    response = subprocess.run(command, input=json.dumps(request) + "\nrelease\n", capture_output=True, text=True, timeout=10)
    assert response.returncode == 2 and json.loads(response.stdout)["refused"] == "conflicting_payload"
    assert effects(d) == original
    independent_completion(d)
    controller.end(d, "deduplicated_then_refused")

    d = controller.case("ordering-deviation", {"outcome": "observer_deviation_only", "effects": 1})
    controller.run(d)
    (d / "original-journal.jsonl").write_bytes((d / "journal.jsonl").read_bytes())
    def reorder(events):
        first = next(i for i, e in enumerate(events) if e["kind"] == "candidate.checked")
        second = next(i for i, e in enumerate(events) if e["kind"] == "negative.detected")
        events[first], events[second] = events[second], events[first]
    mutate_journal(d, reorder)
    independent_completion(d)  # Same valid effects and bindings: deliberately ignores order.
    report = audit.audit(d)
    result = next(iter(report["subjects"].values()))
    assert result["classification"] == "deviating"
    assert result["first_deviation"]["index"] == 2
    assert result["first_deviation"]["raw"]["kind"] == "negative.detected"
    controller.end(d, "observer_deviation_only")

    d = controller.case("mutant-blind-retry", {"defect": "repeats opaque effect", "safe_effects": 1})
    controller.run(d, mode="opaque", pause="sink_committed_reply_withheld")
    result = controller.run(d, mode="opaque", mutant="blind_retry")
    assert result["outcome"] == "completed" and len(effects(d)) == 2
    try:
        independent_completion(d)
        raise RuntimeError("mutant escaped effect assertion")
    except AssertionError as exc:
        assert "effect count" in str(exc)
    controller.end(d, "false_completion", mutant="effect_count")

    d = controller.case("mutant-old-terminal", {"defect": "old incarnation advances run"})
    message = prepare_old_terminal(controller, d)
    result = controller.run(d, message=message, mutant="old_terminal")
    assert result["outcome"] == "completed"
    try:
        independent_completion(d)
        raise RuntimeError("mutant escaped incarnation assertion")
    except AssertionError as exc:
        assert "stale incarnation" in str(exc)
    controller.end(d, "false_completion", mutant="incarnation")

    d = controller.case("mutant-changed-candidate", {"defect": "old evidence certifies changed bytes"})
    controller.run(d, pause="checks_durable")
    path = next(d.glob("candidate-*.json"))
    # Whitespace preserves candidate semantics but changes its exact checked bytes.
    path.write_bytes(path.read_bytes() + b"\n")
    result = controller.run(d, mutant="changed_candidate")
    assert result["outcome"] == "completed"
    try:
        independent_completion(d)
        raise RuntimeError("mutant escaped evidence assertion")
    except AssertionError as exc:
        assert "stale evidence" in str(exc)
    controller.end(d, "false_completion", mutant="candidate_digest")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    if args.output:
        args.output.mkdir(parents=True, exist_ok=False)
        base = args.output.resolve()
    else:
        base = Path(tempfile.mkdtemp(prefix="resume-parley-"))
    controller = Controller(base)
    print(f"ARTIFACTS {base}", flush=True)
    all_cases(controller)
    save(base / "summary.json", controller.results)
    (ROOT / "artifacts").mkdir(exist_ok=True)
    save(ROOT / "artifacts/latest-run.json", {"directory": str(base), "cases": len(controller.results),
                                              "summary": str(base / "summary.json")})
    print(f"\n{len(controller.results)} cases passed; 3 broken policies caught; protocol is OBSERVER ONLY.")
    print(f"Replay all cases: {sys.executable} {ROOT / 'tests/run.py'}")


if __name__ == "__main__":
    main()
