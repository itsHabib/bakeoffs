"""Single-host recovery executor. Parley observes its facts; it is not an effect gate."""
import argparse
import json
import select
import subprocess
import sys
from pathlib import Path

import common as c


def emit(value):
    print(json.dumps(value, sort_keys=True), flush=True)


def boundary(name, pause):
    emit({"boundary": name})
    if pause == name and sys.stdin.readline().strip() != "continue":
        raise c.Refusal("controller_closed_handshake")


def sink_query(directory, mode, op):
    response = subprocess.run([sys.executable, str(c.ROOT / "src/sink.py"),
                               str(directory), mode, "query"],
                              input=json.dumps({"op": op}) + "\n", text=True,
                              capture_output=True, timeout=10, check=False)
    value = json.loads(response.stdout)
    if response.returncode or "refused" in value:
        raise c.Refusal(value.get("refused", "sink_query_failed"))
    return value["receipt"]


def sink_put(directory, mode, payload, op, pause):
    process = subprocess.Popen([sys.executable, str(c.ROOT / "src/sink.py"),
                                str(directory), mode, "put"], stdin=subprocess.PIPE,
                               stdout=subprocess.PIPE, text=True)
    process.stdin.write(json.dumps({"op": op, "payload": payload,
                                   "payload_digest": c.identity(payload)}) + "\n")
    process.stdin.flush()
    response = sink_reply(process)
    if "refused" in response:
        process.stdin.close()
        process.wait(timeout=10)
        raise c.Refusal(response["refused"])
    if response != {"boundary": "sink_committed_reply_withheld"}:
        raise c.Refusal("sink_handshake_invalid")
    boundary("sink_committed_reply_withheld", pause)
    process.stdin.write("release\n")
    process.stdin.flush()
    response = sink_reply(process)
    process.stdin.close()
    process.wait(timeout=10)
    return response["receipt"]


def sink_reply(process):
    if not select.select([process.stdout], [], [], 10)[0]:
        process.kill()
        process.wait(timeout=10)
        raise c.Refusal("sink_reply_timeout")
    return json.loads(process.stdout.readline())


def reconcile(directory, mode, intent):
    if mode == "opaque":
        raise c.Refusal("unresolved_opaque_outcome:" + intent["op"])
    return sink_query(directory, mode, intent["op"])


def verify_receipt(receipt, intent):
    if receipt["op"] != intent["op"] or receipt["payload_digest"] != c.identity(intent["payload"]):
        raise c.Refusal("receipt_subject_mismatch")


def finish(journal, terminal, source, content, evidence, receipt):
    current = journal.latest("incarnation")
    c.check_incarnation(terminal, current)
    if terminal["run"] != journal.events[0]["data"]["run"]:
        raise c.Refusal("terminal_run_mismatch")
    if current["source"] != c.identity(source) or terminal["source"] != current["source"]:
        raise c.Refusal("terminal_source_mismatch")
    c.check_candidate_binding(content, source, evidence)
    if terminal["candidate"] != c.digest(content) or terminal["evidence"] != c.identity(evidence):
        raise c.Refusal("terminal_evidence_mismatch")
    if terminal["receipt"] != receipt:
        raise c.Refusal("terminal_receipt_mismatch")
    intent = journal.latest("delivery.intent", c.identity(source))
    if not intent or not journal.latest("negative.detected", c.identity(source)):
        raise c.Refusal("terminal_missing_obligations")
    verify_receipt(receipt, intent)
    journal.append("completed", **terminal)
    return {"outcome": "completed", **terminal}


def work(args):
    directory = args.directory
    with c.locked(directory):
        journal = c.Journal(directory)
        if not journal.events:
            journal.append("opened", definition=c.definition(), run=args.run,
                           mode=args.mode, sources=c.SOURCES)
        header = journal.events[0]["data"]
        if header["run"] != args.run or header["mode"] != args.mode:
            raise c.Refusal("run_or_sink_mode_mismatch")
        source = c.SOURCES[args.source]
        subject = c.identity(source)
        incarnation = (journal.latest("incarnation") or {}).get("incarnation", 0) + 1
        journal.append("incarnation", source=subject, incarnation=incarnation)
        boundary("incarnation_durable", args.pause)
        path = directory / f"candidate-{subject}.json"
        candidate_event = journal.latest("candidate.ready", subject)
        if candidate_event is None:
            c.durable_write(path, c.encoded(c.candidate(source)))
            journal.append("candidate.ready", source=subject, incarnation=incarnation,
                           candidate=c.digest(path.read_bytes()), path=path.name)
        content = path.read_bytes()
        checked = journal.latest("candidate.checked", subject)
        if checked is None:
            if not c.check(content, source):
                raise c.Refusal("checker_rejected_candidate")
            evidence = c.binding_evidence(source, content)
            journal.append("candidate.checked", source=subject, incarnation=incarnation,
                           evidence=evidence, evidence_digest=c.identity(evidence))
        else:
            evidence = checked["evidence"]
        c.check_candidate_binding(content, source, evidence)
        negative = journal.latest("negative.detected", subject)
        if negative is None:
            defective = json.loads(content)
            defective["squares"][0] += 1
            defect_bytes = c.encoded(defective)
            if c.check(defect_bytes, source):
                raise c.Refusal("negative_control_not_detected")
            defect_path = directory / f"negative-{subject}.json"
            c.durable_write(defect_path, defect_bytes)
            journal.append("negative.detected", source=subject, incarnation=incarnation,
                           path=defect_path.name, digest=c.digest(defect_bytes),
                           checker="square-list-v1")
        boundary("checks_durable", args.pause)
        complete = journal.latest("completed", subject)
        if complete:
            return {"outcome": "already_completed", **complete}
        intent = journal.latest("delivery.intent", subject)
        prior_intent = intent is not None
        payload = c.delivery(args.run, source, content, evidence)
        if intent is None:
            intent = {"source": subject, "incarnation": incarnation,
                      "op": c.operation(payload), "payload": payload}
            journal.append("delivery.intent", **intent)
            boundary("intent_durable", args.pause)
        if intent["payload"] != payload or intent["op"] != c.operation(payload):
            raise c.Refusal("intent_subject_mismatch")
        retained = journal.latest("delivery.receipt", subject)
        receipt = retained["receipt"] if retained else None
        if receipt is None and prior_intent:
            try:
                receipt = reconcile(directory / "sink", args.mode, intent)
            except c.Refusal as exc:
                journal.append("unresolved", source=subject, incarnation=incarnation,
                               op=intent["op"], reason=str(exc))
                return {"outcome": "unresolved", "op": intent["op"], "reason": str(exc)}
        if receipt is None:
            receipt = sink_put(directory / "sink", args.mode, payload, intent["op"], args.pause)
        verify_receipt(receipt, intent)
        if retained is None:
            journal.append("delivery.receipt", source=subject, incarnation=incarnation,
                           receipt=receipt, op=intent["op"])
        boundary("receipt_durable", args.pause)
        terminal = {"run": args.run, "source": subject, "incarnation": incarnation,
                    "candidate": c.digest(content), "evidence": c.identity(evidence),
                    "receipt": receipt}
        return finish(journal, terminal, source, path.read_bytes(), evidence, receipt)


def terminal_message(args):
    with c.locked(args.directory):
        journal = c.Journal(args.directory)
        terminal = json.loads(args.message.read_text())
        source = c.SOURCES[args.source]
        subject = c.identity(source)
        try:
            checked = journal.latest("candidate.checked", subject)
            retained = journal.latest("delivery.receipt", subject)
            if not checked or not retained:
                raise c.Refusal("terminal_missing_evidence")
            return finish(journal, terminal, source,
                          (args.directory / f"candidate-{subject}.json").read_bytes(),
                          checked["evidence"], retained["receipt"])
        except c.Refusal as exc:
            journal.append("terminal.rejected", source=subject, raw=terminal, reason=str(exc))
            return {"outcome": "refused", "reason": str(exc)}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("directory", type=Path)
    parser.add_argument("--run", default="run-demo")
    parser.add_argument("--mode", choices=["queryable", "opaque"], default="queryable")
    parser.add_argument("--source", choices=list(c.SOURCES), default="S1")
    parser.add_argument("--pause", default="")
    parser.add_argument("--message", type=Path)
    args = parser.parse_args()
    args.directory.mkdir(parents=True, exist_ok=True)
    try:
        result = terminal_message(args) if args.message else work(args)
    except (c.Refusal, OSError, ValueError, KeyError, subprocess.TimeoutExpired) as exc:
        result = {"outcome": "refused", "reason": str(exc)}
    emit(result)
    return 0 if result["outcome"] in ("completed", "already_completed", "unresolved") else 2


if __name__ == "__main__":
    sys.exit(main())
