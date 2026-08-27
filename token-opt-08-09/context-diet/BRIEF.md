# Entry 3: hack3-context-diet — recoup the tool-result tokens nobody meant to spend

Read ~/dev/workbench/docs/hackathon-token-opt/README.md first — house rules
and judging apply verbatim. Repo: `~/dev/hack3-context-diet`. You never see
the other 4 entries.

## The bet

In coding-agent sessions, the fattest raw-token slice is usually tool
results: full-file reads where 40 lines were needed, the same file read
three times, command output full of ANSI codes and progress bars, verbose
test logs. None of it was a decision anyone made — it's default plumbing.
The bet: a small set of mechanical hygiene rules, replayed against real
transcripts, proves double-digit-percent savings with zero behavior change
— and the replay harness itself is the product, because it prices every
future hygiene idea before anyone ships it. Buyer: anyone whose agent bill
is dominated by context, i.e. everyone with an agent bill.

## What to build

An offline replay harness over `~/.claude/projects/**/*.jsonl` — no hooks,
no live interception, no modification of Claude Code.

Deterministic core (no model anywhere):

1. **Replay reader:** walk each session transcript; reconstruct the sequence
   of `tool_use` → `tool_result` pairs with their content and the assistant
   text that follows them.
2. **Rule set** (each rule a pure function old-result → new-result, each
   individually toggleable):
   - truncate file reads beyond N lines (parameterized; try 200/500);
   - dedupe re-reads — an identical read of the same path later in the
     session is replaced by a one-line "unchanged since above" stub;
   - strip ANSI escapes, progress-bar redraws, and repeated blank lines from
     command output;
   - cap any single tool result at K tokens with a marker.
3. **Recount:** tiktoken (o200k_base, labeled a proxy) over original vs
   dieted transcripts. Attribute savings per rule, per session, and in
   aggregate across the 172-session corpus.
4. **Safety check, computed not vibed:** for every span a rule removed,
   string-search whether the *removed content* (identifiers, line ranges,
   error strings) is referenced by any later assistant message in the real
   session. Report each rule's "later-referenced" hit-rate as its risk
   score. A rule that saves 8% with 0.2% reference-hits is a shippable
   default; one that saves 12% with 9% hits is not — the report must make
   that tradeoff visible per rule.

The moment the demo turns on: the aggregate table — "these four dumb rules
would have saved X% of last month's input tokens, and here's the computed
evidence for which are safe."

## What NOT to build

- No live integration: no PreToolUse/PostToolUse hooks, no proxy, no patch
  to Claude Code. Replay-only. (Shipping the winning rules live is the
  winner's follow-through, not this demo.)
- No model-based summarization of tool results — mechanical rules only;
  keep the correctness path deterministic.
- No config framework — a flags struct and a table of rules is plenty.
- No UI beyond a clean CLI report (a static HTML table is optional sugar).

## Canned demo (required)

`make demo` runs the harness on bundled synthesized fixture sessions with
known planted waste (asserted exactly in tests: the dedupe fixture has a
known 3x re-read, the ANSI fixture a known escape-heavy build log), then on
the real corpus, printing the per-rule savings + risk table. Tests cover
every rule's transform and the reference-detection logic.

## The 60-second demo story

"Nobody decided to spend these tokens — they're default plumbing. I replayed
my last 172 real sessions through four mechanical hygiene rules. Result:
[X]% of input tokens gone, zero model calls, and — because 'would it have
broken anything' is computed, not guessed — I can show you rule by rule
which are free wins and which actually delete things the agent later needed.
This harness prices any future context-diet idea in minutes."
