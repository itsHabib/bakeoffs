# hack3-spend-audit — the invoice nobody could show you

You burn ~5B agent tokens a month and want ~3B. Every lever anyone's proposed
— compression, caching, offloading — has been argued **without data**. This
itemizes where the tokens actually go, in dollars, from the one source of
ground truth on the machine: your own session transcripts.

No model. No network. No API keys. It reads `~/.claude/projects/**/*.jsonl`,
sums each assistant message's `message.usage`, prices it from an editable
table, and renders one static HTML dashboard + a plain-text report.

## Run it

```bash
make demo      # canned: bundled fixtures with asserted totals, opens the dashboard
make real      # the money shot: your real ~/.claude/projects corpus
make test      # exact-number table tests on the deterministic core
```

`make demo` needs only `python3` (stdlib). `make venv` (optional) installs
`tiktoken` so tool-result *volume* is counted with a real tokenizer instead of
a chars/4 estimate — everything else is exact either way.

## What it measures (and how honest each number is)

| Number | Source | Exactness |
|---|---|---|
| Fresh input / cache read / cache write (1h & 5m) / output tokens | `message.usage` fields | **exact** |
| Tokens by day / project / session / model | grouped from the above | **exact** |
| Dollars | tokens × your `prices.toml` rates | exact once you fill the table |
| Tool-result **volume** by tool | `tool_result` size → tokens | **estimate**, labeled (tiktoken `o200k_base`, a proxy for the Claude tokenizer; or chars/4) |

Two correctness details that change the totals — both locked by tests:

- **Streamed partials are deduped by keeping the richest line.** A message id
  recurs across partial lines with a stub `output_tokens` that grows to its
  final value; summing them double-counts, keeping the first under-counts
  output (the priciest category). On the real corpus this is ~9.5k duplicate
  lines and ~18% of output tokens.
- **Worktrees roll up to their repo.** `…/repo/.claude/worktrees/xyz` is
  attributed to `…/repo`, so "by project" matches how you think about projects.

## Prices live in one table

`prices.toml` — one row per model, one column per category (input, output,
cache read, cache write 5m, cache write 1h), all **`0.0` / TODO**. Nothing is
hardcoded from model memory. Fill it from
<https://docs.claude.com/en/docs/about-claude/pricing>; until you do, the tool
shows tokens only and says so. Pricing is **per model**, because a tier flip
changes rates but not where the tokens are.

## What it does NOT do (by design)

No live/admin-API polling, no recommendations engine, no database, no daemon,
no per-message drill-down. It's a batch tool that reports facts. One Python
file (~600 lines, roughly half of which is the inlined HTML/CSS/JS dashboard
template), one dependency, and that dependency is optional.

## Would someone pay for this?

**Buyer:** anyone running agent fleets with a monthly LLM invoice they can't
itemize — platform/FinOps teams, AI-eng leads. LLM cost-observability is
already a paid category (Helicone, Langfuse, Vantage-style cloud-cost tools);
they bill per-request through a proxy. This entry gets the same
by-category/by-project/by-model breakdown **from transcripts already on disk**
— no proxy, no key, nothing leaves the machine. For a regulated or air-gapped
shop that's the difference between "can measure" and "cannot."

## Why not just flip the model picker to a cheaper tier?

That's the obvious alternative — and this tool is what tells you whether it
would even help. The audit shows the token *distribution* (on the real corpus,
**96.7% cache reads, 0.4% output, ~0% fresh input**). A tier flip rescales
every rate but leaves that shape untouched; the highest-leverage move is prefix
stability, not tier. And because pricing is per-model, the audit **quantifies**
a proposed tier flip exactly instead of guessing. It's the measurement that
makes the decision, not a competitor to it.

## Layout

```
audit.py        parse → aggregate → price → render (CLI + HTML). One file.
prices.toml     the only place dollars come from. TODO until you fill it.
test_audit.py   exact-number table tests on the policy layer.
fixtures/       5 synthesized transcripts with hand-computable totals (_gen.py).
```

## Privacy

Transcripts hold private work data. Everything stays local — no uploads, no
artifacts. The bundled fixtures are synthetic (`fixtures/_gen.py`); `out-real/`
is git-ignored.
