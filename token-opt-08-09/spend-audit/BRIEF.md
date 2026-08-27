# Entry 1: hack3-spend-audit — where do 5 billion tokens actually go?

Read ~/dev/workbench/docs/hackathon-token-opt/README.md first — house rules
and judging apply verbatim. Repo: `~/dev/hack3-spend-audit`. You never see
the other 4 entries.

## The bet

The operator burns ~5B tokens/month across agent sessions and wants to reach
~3B, but no tool on this machine can say what a token was spent ON. Every
optimization argument so far has been a guess; the one measured data point
(an A/B test in an old experiment) showed that the "obvious" lever — payload
compression — moved total spend by only 4%. Whoever renders the real
breakdown first decides where the other four entries' ideas live or die.
The buyer is anyone running agent workloads at scale with a monthly invoice
they can't itemize.

## What to build

A parser + aggregator over `~/.claude/projects/**/*.jsonl` and a single
static HTML dashboard.

Deterministic core (all of it — this entry uses no model at all):

1. **Parse** every transcript line; collect each assistant message's
   `message.usage` (fresh input, cache creation split 1h/5m, cache read,
   output) and each `tool_result`'s content size, attributed to the tool
   name from the matching `tool_use` block.
2. **Aggregate** by: category (cache read / cache write / fresh input /
   output), project dir, session, day, and tool name (for tool-result
   volume). Token counts come straight from usage fields; tool-result sizes
   may be measured in bytes with a clearly-labeled est-tokens conversion via
   tiktoken.
3. **Price** via the TODO-marked $/MTok table (one price per category).
   Headline output: "your 30 days = X tokens = $Y, and here are the top 5
   slices in dollars."
4. **Render** a one-file HTML dashboard (vanilla JS, data inlined at build
   time): stacked bars by day, breakdown donut by category, top-10 tables by
   project / session / tool. Plus a plain-text CLI report for terminals.

The moment the demo turns on: the judge sees, for the first time, the actual
ranked answer to "what would have to shrink for 5B to become 3B" — with the
top slice almost certainly not being the one anyone guessed.

## What NOT to build

- No live API or admin-API polling; the transcript corpus is the universe.
- No recommendations engine — report facts, not advice.
- No database, no persistence beyond generated report files.
- No file watching / daemon mode; it's a batch tool.
- No per-message drill-down UI — top-10 tables are the floor of detail.

## Canned demo (required)

`make demo` (or equivalent single command) runs against a bundled
`fixtures/` directory of ~5 synthesized transcripts with known totals
(asserted in tests), generates the dashboard + CLI report, and opens the
HTML. A second command runs the same pipeline against the real
`~/.claude/projects` corpus. Tests assert exact aggregation numbers on the
fixtures.

## The 60-second demo story

"Everyone's been arguing about how to cut token spend — compression,
caching, offloading — with zero data. This is 30 days of my real agent
usage, itemized. Watch: one command, and here's the invoice nobody could
show me — spend by category, by project, by tool, in dollars. The top slice
is [X], which is why [the matching lever] is where the next 40% comes from
— and why [some popular idea] would've saved almost nothing."
