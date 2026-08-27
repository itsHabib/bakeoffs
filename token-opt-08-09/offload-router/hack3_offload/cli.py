"""hack3-offload CLI.

Contract:
    hack3-offload <task-class> [--replay FILE] [--live] [--ledger FILE] < in.json > out.json
    hack3-offload report [--ledger FILE]
    hack3-offload demo   [--live] [--ledger FILE]

Task classes: narrow | extract | classify.

Exit codes:
    0  PASS  — verified output written to stdout
    2  REJECT — gate caught bad model output; reason on stderr, nothing on stdout
    3  usage / input error
    4  model unavailable (live mode, ollama down)
"""

from __future__ import annotations

import argparse
import json
import os
import sys

from . import demo as demo_mod
from . import gate, ledger
from .model import CassetteModel, ModelUnavailable, OllamaModel
from .verifiers import TASK_CLASSES

DEFAULT_LEDGER = os.environ.get("HACK3_LEDGER", ".hack3-ledger.jsonl")

EXIT_PASS = 0
EXIT_REJECT = 2
EXIT_USAGE = 3
EXIT_MODEL = 4


def _run_task(args) -> int:
    try:
        task_input = json.load(sys.stdin)
    except json.JSONDecodeError as e:
        print(f"error: stdin is not valid JSON: {e}", file=sys.stderr)
        return EXIT_USAGE
    if not isinstance(task_input, dict):
        print("error: input must be a JSON object", file=sys.stderr)
        return EXIT_USAGE

    if args.replay:
        model = CassetteModel.from_file(args.replay)
        live = False
    else:
        model = OllamaModel()
        live = True

    try:
        outcome = gate.run(args.task_class, task_input, model, live=live)
    except ModelUnavailable as e:
        print(f"error: {e}", file=sys.stderr)
        return EXIT_MODEL

    ledger.append(args.ledger, outcome.record)

    if outcome.verdict == "PASS":
        json.dump(outcome.output, sys.stdout)
        sys.stdout.write("\n")
        return EXIT_PASS

    print(f"REJECT [{args.task_class}]: {outcome.reason}", file=sys.stderr)
    return EXIT_REJECT


def _report(args) -> int:
    print(ledger.format_report(args.ledger))
    return EXIT_PASS


def _demo(args) -> int:
    return demo_mod.run(ledger_path=args.ledger, live=args.live)


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(prog="hack3-offload", description=__doc__,
                                formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = p.add_subparsers(dest="command", required=True)

    for tc in TASK_CLASSES:
        sp = sub.add_parser(tc, help=f"run the {tc} task class over stdin JSON")
        sp.add_argument("--replay", metavar="FILE",
                        help="replay a cassette instead of calling ollama")
        sp.add_argument("--live", action="store_true",
                        help="(default) call the live ollama model")
        sp.add_argument("--ledger", default=DEFAULT_LEDGER, help="ledger JSONL path")
        sp.set_defaults(func=_run_task, task_class=tc)

    rp = sub.add_parser("report", help="totals of displaced tokens + rejection rates")
    rp.add_argument("--ledger", default=DEFAULT_LEDGER, help="ledger JSONL path")
    rp.set_defaults(func=_report)

    dp = sub.add_parser("demo", help="run the canned demo (all classes, replay)")
    dp.add_argument("--live", action="store_true",
                    help="use live ollama instead of committed cassettes")
    dp.add_argument("--ledger", default=None, help="ledger JSONL path (default: temp)")
    dp.set_defaults(func=_demo)
    return p


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
