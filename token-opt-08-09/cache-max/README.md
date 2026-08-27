# hack3-cache-max — the silent dollars in every cache miss

Prompt caching bills cache **reads** at a fraction of fresh input and cache
**writes** at a premium. Every agent turn re-sends the whole conversation, so
the same tokens cost wildly different dollars depending on whether the prefix
stayed cached. When a prefix **busts**, cheap reads silently become premium
writes — and nobody on this machine knew the hit rate, the bust count, or the
bill.

`cache-max` is an offline analyzer over `~/.claude/projects/**/*.jsonl`. It is
**deterministic and model-free**: every number comes from the `message.usage`
fields the transcripts already record.

## One command

```bash
make demo
```

That (1) runs the policy tests that assert exact hit ratios / bust counts /
dollar deltas on bundled fixtures, (2) prints the fixture report with two
planted busts, then (3) runs the identical tool over 30 days of this machine's
real transcripts and prints the headline what-if number.

Real-corpus run today (≈123 MB, 198 sessions, in ~1.6 s):

```
HEADLINE: 96% hit rate; 790 busts cost $396.80; top cause idle-ttl-expiry
          (58% of busts, seen in 52% of sessions).
          At 5,000,000,000 tok/mo that's ~$565/mo for zero product-code change.
```

## What it computes

1. **Per-message cache accounting** — for every assistant message: cache read,
   cache creation split into 1h vs 5m, fresh input, output. Per session: a hit
   ratio `read / (read + write + fresh)` and totals; rolled up corpus-wide.
2. **Bust detection, mechanical** — a message busts when its
   `cache_creation_input_tokens` exceeds `--bust-frac` (default **0.25**) of the
   *previous* message's `cache_read`, past the first message and above a
   `--min-prev-read` floor (you cannot bust a prefix that was never cached).
   Reported per session and as corpus busts/hour.
3. **Bust attribution, computed not vibed** — at each bust we diff what entered
   the context window since the previous message and blame the **largest new
   contributor** (a tool result by tool name, a user message, a
   system-reminder, an attachment). When *nothing* new entered but a long
   wall-clock gap did, it is attributed to **idle cache-TTL expiry** — the
   session sat idle and the prefix fell out of cache. Output: a ranked table by
   total re-written tokens. **Attribution is correlational** — it names what
   changed most, not a proven cause; it is still computed from transcript
   facts, never guessed.
4. **What-if simulator** — each session is priced twice: as-happened, and a
   counterfactual where every bust's cache-creation tokens *above the
   threshold* are re-billed at the cache-read rate. The delta, summed over the
   corpus and extrapolated to the 5B-tokens/month rate, is the headline.

## Prices — the one editable table

All `$/MTok` rates live in [`pricing.json`](pricing.json), one table, five
columns (fresh input, output, cache read, cache write 5m, cache write 1h). The
real rates ship as `0` = **TODO**: fill them from
<https://docs.claude.com/en/docs/about-claude/pricing>. Until you do, the demo
falls back to clearly-labelled **example** placeholders and prints an
EXAMPLE-PRICES banner, so no printed dollar figure is ever mistaken for real.
The bundled tests use their own fixed rate constants, so the asserted deltas are
independent of whatever you fill in.

## Why not just flip the model picker?

The obvious token-cost alternative is "pick a cheaper model tier and move on."
That is orthogonal here: cache discipline multiplies **whatever** tier you run.
A prefix bust turns a 10×-cheap read into a premium write *at the current
model's prices* — dropping to a cheaper tier drops the base rate but keeps the
same wasteful read→write conversion on every bust. The savings this finds are
**on top of** any tier choice, for zero change to product code or model
selection. And the top cause here — idle-TTL expiry — is a workflow/keep-alive
lever a cheaper model does nothing for.

## Assumptions (stated, not hidden)

- A stable prefix would have stayed cache-resident across the session's TTL, so
  a bust's excess creation is repriced as a read. Real server-side eviction is
  not modelled beyond what the usage fields state.
- Attribution is correlational (see above).
- Extrapolation assumes the same cache behavior at the target monthly volume.

## Layout

```
cmd/cache-max        CLI: walk, parse, analyze, report
internal/transcript  mechanism: JSONL -> ordered typed turns (no policy)
internal/analyze     policy: accounting, bust rule, attribution, what-if  <- tests here
internal/pricing     the one rate table + TODO/example fallback
internal/report      text + single-file HTML rendering
fixtures/            synthesized sessions with planted, known cache behavior
pricing.json         the editable $/MTok table
```

## Checks

```bash
gofmt -l . && go vet ./... && go test ./...
```

Privacy: everything runs locally; nothing is uploaded. The bundled fixtures are
synthesized (`fixtures/gen.py`), not scrubbed real data.
