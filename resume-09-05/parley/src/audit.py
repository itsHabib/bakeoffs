"""Lossless sidecar to the pinned Parley observer. Audits; never gates execution."""
import argparse
import json
import subprocess
from pathlib import Path

import common as c

MESSAGES = {
    "candidate.ready": ("worker", "checker", "candidate.ready"),
    "candidate.checked": ("checker", "worker", "candidate.checked"),
    "negative.detected": ("checker", "recorder", "negative.detected"),
    "delivery.intent": ("worker", "recorder", "delivery.intent"),
    "delivery.receipt": ("recorder", "worker", "delivery.receipt"),
    "completed": ("worker", "recorder", "completed"),
}


def observe(directory, subject, rows):
    trace = directory / (subject + ".trace")
    trace.write_text("run " + subject + "\n" + "".join(" ".join(r["normalized"]) + "\n" for r in rows))
    result = subprocess.run([str(c.ROOT / "vendor/parley/compiler/bin/parleyc"), "observe",
                             str(c.ROOT / "protocol.parley"), str(trace)],
                            capture_output=True, text=True, timeout=10, check=True)
    first = result.stdout.splitlines()[0]
    classification = "deviating" if "DEVIATES" in first else (
        "stalled" if "stalled" in first else "complete")
    return {"classification": classification, "output": result.stdout}


def audit(directory):
    directory = Path(directory)
    events = c.Journal(directory).events
    output = directory / "audit"
    output.mkdir(exist_ok=True)
    groups = {}
    excluded = []
    for event in events:
        if event["kind"] not in MESSAGES:
            if event["kind"] not in {"opened", "incarnation", "unresolved", "terminal.rejected"}:
                raise c.Refusal("unknown_event_during_normalization:" + event["id"])
            excluded.append({"raw": event, "reason": "execution/identity fact outside order protocol"})
            continue
        subject = event["data"]["source"]
        rows = groups.setdefault(subject, [])
        rows.append({"index": len(rows) + 1, "raw": event,
                     "normalized": MESSAGES[event["kind"]]})
    report = {"mode": "observer_only", "subjects": {}, "excluded": excluded}
    for subject, rows in groups.items():
        outcome = observe(output, subject, rows)
        if outcome["classification"] == "deviating":
            for index in range(1, len(rows) + 1):
                prefix = observe(output, "prefix", rows[:index])
                if prefix["classification"] == "deviating":
                    outcome["first_deviation"] = rows[index - 1]
                    break
        c.save_json(output / (subject + ".identity-map.json"), rows)
        report["subjects"][subject] = outcome
    c.save_json(output / "report.json", report)
    return report


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("directory", type=Path)
    args = parser.parse_args()
    print(json.dumps(audit(args.directory), indent=2))
