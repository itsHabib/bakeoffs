# context-diet

**An offline replay harness that prices tool-result hygiene rules against
your real Claude Code transcripts — and computes, rule by rule, which
savings are free and which quietly delete things the agent needed.**

Nobody *decided* to spend the tokens in a 900-line file read where 40 lines
mattered, or in a build log full of ANSI codes and progress-bar redraws.
It's default plumbing. This harness replays `~/.claude/projects/**/*.jsonl`
through four mechanical rules and measures the recouped tokens — with **zero
model calls** and **zero changes to Claude Code** (replay-only; no hooks, no
proxy, no patch).

## One command

```bash
make demo
```

Hands-free. It sets up a venv, runs the tests, replays the bundled
fixtures with known planted waste, replays your real corpus (per-rule
savings + computed risk), and prints a threshold sweep. ~45s cold, ~7s warm.
Writes `report.html` (a single static file) alongside.

Run pieces individually: `make test`, `make fixtures`, `make corpus`,
`make sweep`, `make html`. Point at a different corpus with
`make corpus CORPUS=/path/to/projects`.

## What it found on 114 real sessions

```
tool-result tokens    : 3,549,614 (proxy)
removable (all rules) : 852,370 tokens = 24.0% of tool-result tokens
  FREE (SHIP-rated)   : 12,798 = 0.4%   <- ship today
  the rest is reachable only via cap/truncate — REVIEW-rated, not safe

rule                   fires   saved tok  spans  ref-rate  verdict
cap_result>1500tok       466     839,434    466     33.9%   REVIEW
truncate_reads>200       137     277,200    137     30.7%   REVIEW
strip_noise               91      12,612     91      0.0%    SHIP
dedupe_reads               1         186      1      0.0%    SHIP
```

**The finding is the counter-intuitive part.** The obvious move — "just cap
big tool results at N tokens" — would recoup ~24%. But the harness computes
that ~1 in 3 of those removals deletes an identifier the agent *uses later
in the same session* (`grt_8dea…`, `cmd/gate/internal`, `DESIGN.md`). The
threshold sweep shows the risk never falls to zero at any cap: **there is no
safe size-blind cut.** The only genuinely free rule is stripping terminal
noise (0% reference rate, computed). That is exactly the mistake this
harness stops you from shipping.

## The four rules

Each is a pure function `old_result → new_result`, individually toggleable
(`--rules truncate,dedupe,strip,cap`), and lives in
[`diet/rules.py`](diet/rules.py):

- **truncate_reads>N** — keep the first N lines of a file read, drop the tail
  behind a marker (try `--truncate-lines 200`/`500`).
- **dedupe_reads** — a re-read of a file whose bytes are identical to an
  earlier read becomes a one-line "unchanged since #k" stub; a *changed*
  file is left intact.
- **strip_noise** — remove ANSI escapes, carriage-return progress redraws
  (keep the final frame), and collapse runs of blank lines.
- **cap_result>K** — hard-cap any single result at K tokens with a marker
  (`--cap-tokens 1500`).

## How "risk" is computed (not vibed)

For every span a rule removes, [`diet/safety.py`](diet/safety.py) extracts
the salient identifiers (paths, snake_case/camelCase symbols, dotted names,
ids mixing letters and digits — not prose) that the removal **loses** (i.e.
don't also survive in what we kept), then string-searches the assistant text
that came *after* that result in the real session. A rule's **ref-rate** is
the share of its removed spans whose lost identifiers reappear later.

This is a deliberately **conservative upper bound** on risk: string
co-occurrence, not proven causal use. That makes the `strip_noise = 0%`
result strong (safe even under a generous risk definition), and the ~30% on
cap/truncate a ceiling ("up to 1-in-3," not "exactly"). `SHIP` = ref-rate
< 1%.

Token counts use **tiktoken `o200k_base`** — stated everywhere as a *proxy*
for Claude's real (non-public) tokenizer, not the real thing.

## Why not just pick a cheaper model? (token-necessity)

The obvious alternative to any token-saving idea is *flip the model picker to
a cheaper tier and move on.* Context diet is not a substitute for that — it's
**orthogonal and multiplicative**:

- A tier flip **trades quality** for price. The SHIP-rated diet removes bytes
  with a **computed 0% later-reference rate** — no behavior change — and it
  does so **on whatever tier you already run**. The savings stack on top of a
  tier flip, they don't compete with it.
- A tier flip gives you a discount. It does **not** tell you that your
  planned "cap big tool results" hook would silently degrade agents 1-in-3
  times. This harness does — and that's the part you can't buy by changing a
  dropdown.

## Who pays for this

Anyone whose agent bill is dominated by context — i.e. everyone with an
agent bill — and specifically **any team about to ship a context-trimming
hook or a smaller context window.** Money already moves here: prompt/context
compression is a live product category (context-window management, prompt
caching, log/observability trimming). The buyer isn't buying the 0.4%; they
are buying the harness that **prices any hygiene idea in seconds and blocks
the lossy ones before they ship.** Dollar conversion lives in one editable
table — [`prices.md`](prices.md) — values left TODO for real pricing, since
cache reads, cache writes, fresh input, and output are each priced
differently.

## Layout

```
diet/reader.py   walk a transcript -> ordered tool_use/result pairs + the
                 assistant text that follows each result   (mechanism)
diet/rules.py    the four hygiene rules                    (policy)
diet/safety.py   salient-identifier reference detection    (policy)
diet/tokens.py   tiktoken proxy + an injectable fake tokenizer
diet/harness.py  replay + per-rule attribution + threshold sweep
diet/report.py   CLI table + one static HTML file
cli.py           flags struct + a rule table. nothing more.
fixtures/        synthesized sessions with planted waste (no private data)
tests/           27 tests; run with `make test` (zero deps, no pytest)
```

## Scope (what this deliberately is not)

No live integration (no hooks/proxy/patch). No model-based summarization —
mechanical rules only, so the correctness path stays deterministic. No config
framework. No UI beyond the CLI report and one static HTML table. Shipping the
winning rule live is the follow-through, not this demo.
