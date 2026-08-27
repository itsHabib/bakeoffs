# DEMO — 60 seconds, hands-free

## The one command

```bash
make demo
```

Runs the deterministic pipeline over 5 bundled fixtures (totals asserted in
`test_audit.py`), writes `out/report.txt` + `out/report.html`, and opens the
dashboard. No input, no network, no model. Then, for the real payload:

```bash
make real
```

Same pipeline, your actual `~/.claude/projects` corpus — ~1.7B tokens, ~2s.

## The 60-second script

> "Everyone's been arguing how to cut agent token spend — compression,
> caching, offloading — with **zero data**. Here's my real usage — every
> session on this machine — itemized. One command."
>
> *(run `make real`; the dashboard opens)*
>
> "This is the invoice nobody could show me. Spend by category, by project, by
> model, by day, by tool — and once the price table's filled, in dollars.
>
> Now watch the category breakdown. **Cache reads are 96.7% of every token.**
> Output is 0.4%. Fresh input rounds to zero. So every compression argument on
> this machine has been fighting over **less than one percent** of the bill —
> which is exactly why that old A/B test moved total spend by only 4%.
>
> The lever that matters is the 96.7%: prefix stability, keeping the cache
> alive. And because this prices **per model**, it can tell me whether just
> flipping to a cheaper tier would help — it wouldn't change that shape, only
> the rate. This is the measurement that decides where the next 40% comes from,
> instead of guessing."

## What the judge is looking at

- **Category donut** — the money slide. Cache read dwarfs everything.
- **Stacked bars by day** — 30 days, cache-read-dominated every day.
- **Top projects / sessions / models** — where it concentrates (`~/dev/gate`,
  `~/dev/workbench` lead; opus-5 is the biggest model slice).
- **Top tools by result volume** — which tool outputs fatten context later
  (browser control, Bash, Read).

## The live moment that works

`make demo` is fully canned and asserted, so it can't miss. `make real` is the
live moment — it runs in ~2 seconds on 123MB and the 96.7% headline is real,
not staged.

## Correctness you can check in one minute

```bash
make test
```

11 table tests, exact numbers, hand-computable from `fixtures/_gen.py`. They
lock the two things that actually move the totals: **streamed-partial dedupe**
(keep the richest line — worth ~18% of output tokens on the real corpus) and
**worktree roll-up** (worktrees fold into their repo). The model is used for
**nothing**; correctness is 100% deterministic code.
