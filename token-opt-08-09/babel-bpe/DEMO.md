# DEMO.md — the 60-second walkthrough

```bash
cd ~/dev/hack3-babel-bpe
make demo
```

Runs in ~1.5s. Fully offline-deterministic except a 2-call live check at the
end (skipped cleanly if ollama is down). No keys, nothing leaves the machine.

## What the judge sees, in order

**1. The token leaderboard (N=1 / 10 / 100).** Eight formats, amortized totals
that fold in each format's in-context spec cost. `babel-nl` — my BPE-tuned
mutation — wins at every N. `babel` (the 2025 winner) is second. Plain
`json-compact` costs ~1.6× the leader; `json-pretty` ~2.9×.

**2. The org-chart spotlight.** The exact shape the 2025 report headlined at
"12 tokens." In real BPE tokens: `babel-nl` 103, `toon` 108, `babel` 119,
`yaml` 169, `json-compact` 178, `json-pretty` 318. The "12" was whitespace
tokens; the invoice never saw it.

**3. The frontier — tokens vs comprehension** (qwen2.5:7b as the subject under
test, deterministic exact-match grading):

| format   | msg tok | comprehension |
|----------|--------:|--------------:|
| tsv      |     206 |          100% |
| csv      |     210 |          100% |
| babel-nl |     499 |           90% |
| babel    |     500 |           75% |
| toon     |     608 |           95% |
| json     |     791 |           90% |
| yaml     |     835 |           95% |

**4. The controlled result.** `babel` and `babel-nl` differ in exactly one
thing — underscore packing. babel: 500 tok, 75%. babel-nl: 499 tok, 90%.
Un-packing the underscores is **cheaper AND clearer**. Same grammar, same
questions, same model — so the 15-point gap is a clean measurement of the
packing tax, not noise.

## The spoken story (~60s)

> "Last year agents invented their own language and we celebrated 88%
> compression — counted in whitespace tokens, which is not the unit the invoice
> uses. Here's the same bake-off in real BPE tokens.
>
> The leaderboard: the 2025 winner, Babel, loses. It packs names with
> underscores — `Alice_Nakamura` — and that fragments whole words for the
> tokenizer. Un-pack those underscores and you get babel-nl, which wins at
> every scale.
>
> Now the twist token-counting can't see. We hand each format to a small model
> and ask it to read the payload back. Babel doesn't just cost more tokens — it
> costs 15 points of comprehension versus its own un-packed twin, because the
> model literally echoes `qf_442` when you asked for `QF 442`. The densest
> packing hurt the reader.
>
> So the house wire format isn't the densest one — it's the knee of this curve.
> Here that's babel-nl or plain TOON: near-top compression, no comprehension
> tax. And it's a one-line change in every subagent prompt."

## Would someone pay? (assert, and expect pushback)

Buyer: anyone running a multi-agent pipeline metered per token — the operator's
own portfolio has subagent reports and inter-agent payloads on every run. This
is a **config-level** change (swap the serializer, edit one line in a subagent
prompt) with **recurring** savings and, unlike a model downgrade, **zero
quality loss on the send side** — the round-trip tests prove every format is
losslessly reconstructable.

Pushback the judge will raise, answered:
- *"Just pick a cheaper model."* That cuts $/token; this cuts tokens/payload.
  They multiply — format savings compound on top of any tier. See README
  "token-necessity."
- *"o200k_base isn't Claude's tokenizer."* Correct, and stated everywhere — it's
  a proxy. The *direction* (underscore packing fragments words; density trades
  against comprehension) is tokenizer-agnostic.
- *"The 7b model is weak."* On purpose (keyless/local). The accuracy numbers are
  a floor; the finding is the *shape* of the tradeoff. Upgrade path noted.

## Proof it's real code, not vibes

```bash
make test
```

10 tests, all deterministic, no model in the loop:
`encode → decode → equal` for all 68 (format, payload) pairs, exact pinned
token counts for fixture payloads, and the grader's numeric-boundary /
string-match rules. `src/comprehension.py` calls the model; `src/run.py` and the
scorer never let it grade itself.
