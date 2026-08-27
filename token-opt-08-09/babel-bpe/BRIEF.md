# Entry 2: hack3-babel-bpe — the Babel bake-off, in the right unit this time

Read ~/dev/workbench/docs/hackathon-token-opt/README.md first — house rules
and judging apply verbatim. Repo: `~/dev/hack3-babel-bpe`. You never see the
other 4 entries.

## The bet

In 2025 the operator ran `babel-protocol`
(https://github.com/itsHabib/babel-protocol — clone it into
`reference/` for the spec, task shapes, and prior results): agent teams
invented a compressed wire language and claimed 88% compression. The claim
was measured in whitespace-split tokens. Remeasured in real BPE tokens, the
flagship example fell from 19x to 2.4x — and plain indented YAML beat Babel
outright (101 vs 119 BPE tokens) with zero spec overhead, because Babel's
underscore-and-colon packing fragments words the tokenizer would keep whole.
The bake-off was never actually run in the unit that bills. Whoever runs it
honestly gets to declare the house wire format for every subagent report and
inter-agent payload in the portfolio. The buyer is anyone paying per token
for multi-agent pipelines — payload format is a config-level change with
recurring savings.

## What to build

A deterministic format bake-off harness plus a comprehension-tax rig.

1. **Corpus:** ~10 structured payloads spanning the shapes agents actually
   exchange — flat records, multi-row tables, org-chart hierarchies, graphs,
   constraint sets, verification reports. Reuse the task shapes from
   `reference/docs/experiment-design/content/task-library.md`, but write
   fresh data values so nothing is memorized text.
2. **Encoders:** one per candidate format, each a pure function from a
   common in-memory representation: compact JSON, pretty JSON, YAML, CSV/TSV
   (where shape allows), TOON (TODO: pin the exact TOON spec revision you
   implement), Babel v3 (spec in `reference/README.md`), and at least one
   BPE-tuned mutation of your own design (e.g. Babel without underscores,
   or YAML with schema-line headers).
3. **Scorer:** tiktoken `o200k_base` counts per payload per format, plus
   amortized totals that include each format's in-context spec/legend cost
   (0 for JSON/YAML/CSV; measured for Babel and any invented format).
   Leaderboard: tokens at N=1, N=10, N=100 payloads.
4. **Comprehension tax rig:** for each format, `qwen2.5:7b` receives spec +
   encoded payload and must answer ~5 extraction/reasoning questions per
   payload ("what is X's headcount", "which constraint binds Y"). Grading is
   deterministic exact-match against known answers — the model is the
   subject under test, never the grader. Report accuracy-per-format next to
   tokens-per-format: the interesting output is the frontier, not a single
   winner.

The moment the demo turns on: the leaderboard renders and the 2025
experiment's winner is visibly not the 2026 winner — compression and
comprehension traded off in a way whitespace counting could never see.

## What NOT to build

- No language-genesis agent loop, no multi-round evolution — this is a
  bake-off of fixed formats, not a breeding program.
- No frontier API calls; qwen2.5:7b is the only decoder (note the upgrade
  path in the README).
- No claim about the actual Claude tokenizer — label o200k_base as a proxy.
- No streaming/wire-protocol machinery; formats are strings in files.
- Do not port or "fix" the old repo — it's reference data only.

## Canned demo (required)

`make demo` runs the full pipeline offline-deterministic except the model
step: encodes the corpus in every format, prints the token leaderboard, then
(if ollama is up) runs the comprehension rig and prints the
accuracy-vs-tokens table; with ollama down it prints the cached results from
the last real run (commit those). Tests assert encoder correctness
round-trip (encode → parse back → equal) and exact token counts for pinned
fixture payloads.

## The 60-second demo story

"Last year agents invented their own language and we celebrated 88%
compression — measured in the wrong unit. This is the same bake-off in the
unit the invoice uses. Here's the leaderboard: [winner] wins at N=1 and
[maybe another] at N=100 once spec cost amortizes. And here's the twist —
the densest format cost [X] points of comprehension accuracy on a small
model, so the right house format is the knee of this curve, and it's now a
one-line change in every subagent prompt."
