#!/usr/bin/env python3
"""context-diet CLI — offline replay of tool-result hygiene rules over
Claude Code transcripts.

    python cli.py --corpus ~/.claude/projects        # real corpus
    python cli.py --fixtures                          # bundled fixtures
    python cli.py --corpus DIR --html report.html     # + static HTML

No flags needed beyond a source. A flags struct and a rule table, nothing
more — per the house rules.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

from diet import report as report_mod
from diet.harness import run, run_corpus, sweep
from diet.reader import iter_sessions
from diet.rules import CapResult, DedupeReads, StripNoise, TruncateReads

FIXTURES = Path(__file__).parent / "fixtures"


def build_rules(args):
    """The rule table. Each entry is toggleable from the CLI; default all."""
    table = {
        "truncate": lambda: TruncateReads(args.truncate_lines),
        "dedupe": lambda: DedupeReads(),
        "strip": lambda: StripNoise(),
        "cap": lambda: CapResult(args.cap_tokens),
    }
    picked = args.rules.split(",") if args.rules else list(table)
    unknown = [r for r in picked if r not in table]
    if unknown:
        sys.exit(f"unknown rule(s): {', '.join(unknown)}; have {list(table)}")
    return [table[r] for r in picked]


def _sweep_table(title, unit, rows):
    out = [f"\n  {title}", "  " + "-" * 56,
           f"  {unit:>10}{'saved tok':>13}{'saved%':>9}{'ref-rate':>11}{'fires':>8}",
           "  " + "-" * 56]
    for v, saved, pct, ref, fires in rows:
        out.append(f"  {v:>10}{saved:>13,}{100 * pct:>8.1f}%"
                   f"{100 * ref:>10.1f}%{fires:>8}")
    return "\n".join(out)


def run_sweep(args):
    if not args.corpus:
        sys.exit("--sweep needs --corpus DIR")
    sessions = [s for s in iter_sessions(args.corpus) if s.results]
    print("  Pricing each rule across its parameter — the harness as a")
    print("  cost model. Watch ref-rate: it never drops to 'free' for a")
    print("  size-blind cut, at any threshold. (proxy tokens)")
    print(_sweep_table("cap_result — max tokens per result", "max-tok",
                       sweep(sessions, lambda k: CapResult(k),
                             [1000, 1500, 2000, 4000, 8000])))
    print(_sweep_table("truncate_reads — max lines per file read", "max-lines",
                       sweep(sessions, lambda n: TruncateReads(n),
                             [200, 500, 1000, 2000])))
    return 0


def main(argv=None):
    p = argparse.ArgumentParser(description=__doc__)
    src = p.add_mutually_exclusive_group(required=True)
    src.add_argument("--corpus", metavar="DIR", help="root of *.jsonl transcripts")
    src.add_argument(
        "--fixtures", action="store_true", help="run bundled fixture sessions"
    )
    p.add_argument("--rules", help="comma list: truncate,dedupe,strip,cap")
    p.add_argument("--truncate-lines", type=int, default=200)
    p.add_argument("--cap-tokens", type=int, default=1500)
    p.add_argument("--html", metavar="FILE", help="also write a static HTML report")
    p.add_argument("--sweep", action="store_true",
                   help="price cap/truncate across thresholds (needs --corpus)")
    args = p.parse_args(argv)

    if args.sweep:
        return run_sweep(args)

    rules = build_rules(args)

    if args.fixtures:
        label = "bundled fixtures"
        rep = run(iter_sessions(FIXTURES), rules)
    else:
        label = args.corpus
        rep = run_corpus(args.corpus, rules)

    print(report_mod.render_text(rep, corpus_label=label))

    if args.html:
        Path(args.html).write_text(report_mod.render_html(rep, corpus_label=label))
        print(f"\n  wrote {args.html}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
