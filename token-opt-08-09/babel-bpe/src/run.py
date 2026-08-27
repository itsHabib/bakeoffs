"""CLI: the bake-off harness.

    python src/run.py leaderboard     token leaderboard (offline, deterministic)
    python src/run.py comprehension   run the live rig, write cached/comprehension.json
    python src/run.py demo            the canned 60-second demo (see DEMO.md)
"""

import json
import os
import sys

import comprehension as C
import corpus
import encoders as E
import scorer

HERE = os.path.dirname(os.path.abspath(__file__))
CACHE = os.path.join(HERE, "..", "cached", "comprehension.json")

BOLD, DIM, GRN, YEL, RED, RST = "\033[1m", "\033[2m", "\033[32m", "\033[33m", "\033[31m", "\033[0m"


def _bar(pct, width=20):
    fill = round(pct / 100 * width)
    return "█" * fill + "·" * (width - fill)


def print_leaderboard():
    boards, scores, spec = scorer.leaderboards()
    print(f"\n{BOLD}TOKEN LEADERBOARD{RST}  (tiktoken o200k_base — a proxy for the Claude tokenizer)")
    print(f"{DIM}amortized total = spec_tokens + N × corpus_tokens{RST}\n")
    for n in (1, 10, 100):
        print(f"  {BOLD}N={n:<3}{RST}  payloads sent")
        print(f"  {'format':14}{'total':>9}{'msg':>8}{'spec':>6}   share")
        best = min(r["total"] for r in boards[n] if not r["partial"])
        for r in boards[n]:
            tag = f" {DIM}(tables only){RST}" if r["partial"] else ""
            mark = f"{GRN}◆{RST}" if (not r["partial"] and r["total"] == best) else " "
            print(f" {mark}{r['format']:14}{r['total']:>9}{r['msg_tokens']:>8}"
                  f"{r['spec_tokens']:>6}   {r['total']/best:>4.2f}×{tag}")
        print()


def print_shape_spotlight():
    """The 2025 flagship shape (org chart), in the unit the invoice uses."""
    _, scores, _ = scorer.leaderboards()
    print(f"  {BOLD}Spotlight — the org-chart shape the 2025 report headlined at \"12 tokens\":{RST}")
    order = sorted((f for f in E.FORMAT_ORDER if "org_chart" in scores[f]),
                   key=lambda f: scores[f]["org_chart"])
    worst = max(scores[f]["org_chart"] for f in order)
    for f in order:
        t = scores[f]["org_chart"]
        print(f"    {f:14}{t:>4} tok  {_bar(t / worst * 100)}")
    print(f"  {DIM}2025 counted whitespace tokens and called Babel the winner. In real BPE"
          f"\n  tokens, un-packing Babel's underscores (babel-nl) wins — same grammar,"
          f"\n  the tokenizer just keeps whole words whole.{RST}\n")


def load_cache():
    with open(CACHE) as f:
        return json.load(f)


def save_cache(results):
    payload = {"model": C.MODEL, "results": results}
    with open(CACHE, "w") as f:
        json.dump(payload, f, indent=2)


def _progress(i, total, fmt, name, ok):
    mark = f"{GRN}✓{RST}" if ok else f"{RED}✗{RST}"
    sys.stdout.write(f"\r  [{i:>3}/{total}] {mark} {fmt:14} {name:16}")
    sys.stdout.flush()
    if i == total:
        sys.stdout.write("\n")


def cmd_comprehension():
    if not C.ollama_up():
        print(f"{RED}ollama is not reachable at {C.OLLAMA}{RST} — start it and retry.")
        sys.exit(1)
    print(f"{BOLD}Running comprehension rig{RST} against {C.MODEL} (the subject under test)…")
    results = C.run_full(progress=_progress)
    save_cache(results)
    print(f"Wrote {os.path.relpath(CACHE)}")
    print_frontier(results)


def print_frontier(results):
    _, scores, spec = scorer.leaderboards()
    print(f"\n{BOLD}THE FRONTIER — tokens vs comprehension{RST}  (qwen2.5:7b reading each format)\n")
    print(f"  {'format':14}{'msg tok':>8}{'accuracy':>11}   comprehension")
    rows = []
    for fmt, t in results.items():
        acc = 100 * t["correct"] / t["total"]
        rows.append((fmt, sum(scores[fmt].values()), acc, t["correct"], t["total"]))
    for fmt, tok, acc, c, n in sorted(rows, key=lambda r: r[1]):
        col = GRN if acc >= 90 else (YEL if acc >= 75 else RED)
        print(f"  {fmt:14}{tok:>8}{col}{acc:>9.0f}%{RST}  {c}/{n}  {_bar(acc)}")
    print(f"\n  {DIM}The densest wire format is not automatically the right one: read down the"
          f"\n  token column and across to accuracy — the house format is the knee, not"
          f"\n  the floor.{RST}")
    if "babel" in results and "babel-nl" in results:
        b, n = results["babel"], results["babel-nl"]
        ba = 100 * b["correct"] / b["total"]
        na = 100 * n["correct"] / n["total"]
        bt, nt = sum(scores["babel"].values()), sum(scores["babel-nl"].values())
        print(f"\n  {BOLD}Controlled result{RST} — babel vs babel-nl differ in {BOLD}one thing only{RST}: "
              f"underscore packing.")
        print(f"    babel     {bt} tok  {ba:.0f}% comprehension   {DIM}(the 2025 winner){RST}")
        print(f"    babel-nl  {nt} tok  {na:.0f}% comprehension   {GRN}← cheaper AND clearer{RST}")
        print(f"  {DIM}Un-packing the underscores is a free lunch on both axes. That is the"
              f"\n  whole 2025 headline, reversed, in the unit the invoice actually uses.{RST}\n")


def cmd_demo():
    print(f"{BOLD}════ hack3-babel-bpe — the Babel bake-off, in the unit the invoice uses ════{RST}")
    print_leaderboard()
    print_shape_spotlight()
    if os.path.exists(CACHE):
        cache = load_cache()
        print(f"{DIM}(comprehension: cached run of {cache['model']}){RST}")
        print_frontier(cache["results"])
    else:
        print(f"{YEL}No cached comprehension run found — run: make comprehension{RST}")
    if C.ollama_up():
        _live_smoke()
    else:
        print(f"{DIM}ollama down — leaderboard + cached frontier above are fully offline.{RST}")


def _live_smoke():
    print(f"{BOLD}Live check{RST} {DIM}(proving the rig is real, not a fixture){RST}: "
          f"same question, dense vs plain —")
    q, exp = "What is the headcount of Lead Platform?", "9"
    for fmt in ("babel", "yaml"):
        resp = C.ask(C.prompt_for(fmt, "org_chart", q)).strip().replace("\n", " ")
        ok = C.grade(exp, resp, True)
        mark = f"{GRN}✓{RST}" if ok else f"{RED}✗{RST}"
        print(f"    {mark} {fmt:10} answered: {resp[:40]!r}  (expected {exp})")
    print()


def main():
    cmd = sys.argv[1] if len(sys.argv) > 1 else "demo"
    if cmd == "leaderboard":
        print_leaderboard()
        print_shape_spotlight()
    elif cmd == "comprehension":
        cmd_comprehension()
    elif cmd == "demo":
        cmd_demo()
    else:
        print(__doc__)
        sys.exit(2)


if __name__ == "__main__":
    main()
