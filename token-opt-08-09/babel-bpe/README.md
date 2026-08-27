# hack3-babel-bpe — the Babel bake-off, in the unit the invoice uses

In 2025 the `babel-protocol` experiment
([itsHabib/babel-protocol](https://github.com/itsHabib/babel-protocol)) had
agent teams invent a compressed wire language and reported **88% compression**
and a flagship org-chart at **12 tokens**. That number was counted in
*whitespace-split tokens*. The invoice is metered in *tokenizer (BPE) tokens*.
This repo reruns the bake-off in the billed unit against honest baselines, and
adds the thing token-counting alone can't see: a **comprehension tax**.

```bash
make demo        # the 60-second walkthrough — offline-deterministic + a live check
```

One command. No API keys, nothing leaves the machine.

## What it does

1. **Corpus** (`src/corpus.py`) — 10 freshly-synthesized structured payloads
   spanning the shapes agents actually exchange: flat records, multi-row
   tables, an org-chart hierarchy, directed graphs, a constraint set, and a
   verification report. All built from four structural types (record / table /
   hierarchy / graph). No text is reused from the 2025 experiment.
2. **Encoders** (`src/encoders.py`) — one pure function per candidate format:
   `json-compact`, `json-pretty`, `yaml`, `csv`, `tsv`, `toon`, `babel` (v3,
   the 2025 winner), and **`babel-nl`** — my BPE-tuned mutation. Every encoder
   has a matching decoder; the test suite proves `encode → decode → equal` for
   all 68 (format, payload) pairs.
3. **Scorer** (`src/scorer.py`) — real BPE token counts via tiktoken
   `o200k_base`, plus **amortized** leaderboards at N=1, N=10, N=100 payloads
   that fold in each format's in-context spec/legend cost (0 for JSON/YAML/CSV,
   measured for TOON/Babel/babel-nl).
4. **Comprehension rig** (`src/comprehension.py`) — for each format, `qwen2.5:7b`
   gets the legend + encoded payload + a question and must answer; grading is
   deterministic exact-match against known values. The model is the **subject
   under test, never the grader**. Output: accuracy-per-format next to
   tokens-per-format — the frontier.

## The bet, and the buyer

Whoever runs this honestly gets to declare the **house wire format** for every
subagent report and inter-agent payload in the portfolio. The buyer is anyone
paying per token for a multi-agent pipeline: payload format is a config-level
change with recurring savings, and swapping it is a one-line edit in a subagent
prompt. Dollar math lives in one editable table, [`prices.md`](prices.md)
(values TODO — fill from the official pricing page, never model memory).

## The finding (spoiler)

- **The 2025 winner is not the 2026 winner.** Babel packs multi-word values
  with underscores (`Alice_Nakamura`) and separates fields with spaces. That
  underscore packing is exactly what *fragments* whole words for a BPE
  tokenizer. `babel-nl` keeps the identical grammar but un-packs the
  underscores (real spaces, newline/tab field separators) — same information,
  fewer tokens, and it **wins the token leaderboard at every N**.
- **Spec cost decides low-N.** At N=1 an invented format is dragged down by the
  legend it must teach in-context; by N=100 message density dominates. The
  amortized board shows the crossover.
- **Density trades against comprehension.** The densest encodings cost a small
  model some extraction accuracy. The right house format is the **knee** of the
  tokens-vs-accuracy curve, not its floor — see `make demo`.

## Honesty notes (read these)

- **`o200k_base` is a proxy.** It is a real, modern BPE vocabulary, but not the
  actual Claude tokenizer. Every "real BPE tokens" headline here means "real
  BPE tokens *under this proxy*". Directionally sound, not the billed exact.
- **TOON revision is pinned.** This repo implements *TOON r1 (this repo's
  pinned subset)* — the exact grammar in `src/specs.py`'s legend and
  `src/encoders.py`'s `toon_*` functions. It is a reasonable Token-Oriented
  Object Notation, not a claim to track any upstream revision.
- **The comprehension model is small on purpose.** House rules: keyless and
  local. `qwen2.5:7b` is the decoder; the accuracy numbers are a floor, and a
  frontier model would likely erase much of the comprehension gap (the
  **upgrade path**: point `comprehension.MODEL` at a stronger local or hosted
  model). The *shape* of the tradeoff is the finding, not the absolute %.
- **Decoders may use a schema, because that's how these formats really work.**
  A hierarchy decoder is handed the attribute key list (e.g. `headcount`) that
  Babel keeps in its spec/legend rather than in every message — and whose cost
  is already counted in the amortized spec column. It is never handed data
  values, so round-trip can't cheat.

## Why not just pick a cheaper model? (the token-necessity alternative)

The obvious way to cut token spend is to flip the model picker to a cheaper
tier and move on. That changes the **price per token**; this changes the
**tokens per payload** — the two multiply. Format savings compound *on top of*
whatever tier you run, they don't collide with it, and unlike a tier downgrade
they cost zero quality on the sending side (the payload is losslessly
reconstructable — the round-trip tests prove it). Cheaper-tier is the baseline
this beats, not a substitute for it.

## Commands

```bash
make demo           # canned 60-second walkthrough (default)
make leaderboard    # token leaderboard only — fully offline, deterministic
make comprehension  # re-run the live rig against qwen2.5:7b, refresh the cache
make test           # round-trips + pinned token counts + grader (deterministic)
```

`make demo` reads the committed comprehension cache (`cached/comprehension.json`)
so it is fast and works with ollama down; if ollama is up it also fires a tiny
live check to prove the rig is real, not a fixture.

## Layout

```
src/corpus.py         10 payloads, four structural types, fresh data
src/encoders.py       8 encoders + matching decoders + round-trip equality
src/specs.py          in-context legend per format (the amortized spec cost)
src/scorer.py         tiktoken counts + amortized N=1/10/100 leaderboards
src/questions.py      comprehension questions with known answers
src/comprehension.py  the qwen2.5:7b rig + the deterministic grader
src/run.py            CLI: leaderboard | comprehension | demo
tests/                round-trips, pinned counts, grader — the policy layer
prices.md             the one editable $/MTok table (values TODO)
cached/               committed comprehension run for offline demo
```
