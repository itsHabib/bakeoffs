# Entry 5: hack3-cache-max — the silent dollars in every cache miss

Read ~/dev/workbench/docs/hackathon-token-opt/README.md first — house rules
and judging apply verbatim. Repo: `~/dev/hack3-cache-max`. You never see the
other 4 entries.

## The bet

Prompt caching bills reads at a fraction of fresh-input price and writes at
a premium — so the same 5B tokens can cost wildly different dollars
depending on cache hit rate, and a mid-session prefix bust silently converts
cheap reads into expensive writes. Nobody on this machine knows the actual
hit rate, how many busts happen, or what they cost. The bet: the transcripts
already contain everything needed to answer this (`cache_read_input_tokens`,
`cache_creation_input_tokens`, the 1h/5m ephemeral split, per message), and
the dollar delta between "cache behavior as it happened" and "cache behavior
with stable prefixes" is the single biggest no-code-change number in this
hackathon. Buyer: anyone whose bill is dominated by input tokens re-read
every turn — which is what agent sessions are.

## What to build

An offline analyzer over `~/.claude/projects/**/*.jsonl`. Deterministic
core, no model:

1. **Per-message cache accounting:** for every assistant message, record
   cache read, cache creation (1h vs 5m), fresh input, output. Per session:
   hit ratio = cache read / (cache read + cache creation + fresh input),
   trend over the session, and totals.
2. **Bust detection, defined mechanically:** a bust is a message whose
   `cache_creation_input_tokens` exceeds a threshold (parameterized;
   default: > 25% of the previous message's cache read) after the session
   is past its first message. Report busts per session and corpus-wide
   busts/hour.
3. **Bust attribution, computed not vibed:** at each bust, diff what
   entered the context window between the previous message and this one —
   which tool results, how large, any system-reminder-shaped entries — and
   bucket busts by the largest new contributor. Output: ranked table of
   bust causes by total re-written tokens. (Attribution is correlational;
   label it as such. The ranking is still computed from transcript facts,
   never guessed.)
4. **What-if simulator:** recompute each session's dollar cost under the
   TODO-priced $/MTok table twice — as-happened vs a counterfactual where
   every bust's cache-creation tokens above the threshold are re-priced as
   cache reads. That delta, summed over 30 days and extrapolated to the 5B
   monthly rate, is the headline number.

The moment the demo turns on: "your 30-day hit rate is X%; busts cost you
$Y; the top cause is [Z] — and it's the same cause in 80% of sessions."

## What NOT to build

- No changes to Claude Code, no hooks, no live proxy — analysis only.
  (Fixing the top bust cause is the winner's follow-through.)
- No prescriptive prompt-restructuring engine — one short "observations"
  section in the report is the ceiling for advice.
- No dashboard framework — CLI report first; one static HTML page with
  inline data as optional sugar.
- No attempt to model Anthropic's server-side cache internals beyond what
  the usage fields directly state; where the simulator assumes (e.g. a
  stable prefix would have stayed cached across the 1h TTL), state the
  assumption in one line in the report.

## Canned demo (required)

`make demo` runs on bundled synthesized fixtures with planted cache
behavior — one steady high-hit session, one session with two engineered
busts of known size — and tests assert the exact hit ratios, bust
detections, and counterfactual dollar deltas. Then the same command runs
against the real corpus and prints the 30-day report with the headline
what-if number.

## The 60-second demo story

"Every turn, the whole conversation gets re-sent — caching is what makes
that affordable, and a prefix bust silently turns 10x-cheap reads into
premium writes. I mined 30 days of my real transcripts: hit rate [X]%,
[N] busts, top cause [Z]. The what-if number: with stable prefixes those
same sessions cost $[Y] less — extrapolated to my 5B monthly rate, that's
$[Y×k] a month for changing zero lines of product code. That's why cache
discipline comes before every clever compression scheme."
