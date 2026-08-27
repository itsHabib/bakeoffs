# DEMO — the 60-second walkthrough

## Run it (hands-free)

```bash
make demo
```

That's the whole demo. No live input, no keys, no network beyond the
one-time `pip install tiktoken`. ~45s from a cold checkout, ~7s warm.

## What the judge sees, in order

1. **Tests green** — `27 passed, 0 failed`. Every rule's transform and the
   reference-detection logic are table-tested (see below).
2. **Fixtures with *planted* waste** — 3 synthesized sessions. Watch
   `truncate_reads` come back **REVIEW at 100% ref-rate**, and the flagged
   identifier printed is `handleAuthRetry_v2` — the exact symbol the fixture
   plants in the truncated tail and then references later. The harness caught
   it by computation, not by luck.
3. **Real corpus** — 114 of your own sessions, 3.5M tool-result tokens. The
   per-rule table with savings and computed risk. Writes `report.html`.
4. **Threshold sweep** — the harness pricing `cap` and `truncate` across five
   settings each. The ref-rate column never drops to "free."

## The 60-second script

> "Nobody decided to spend these tokens — full-file reads, ANSI-soaked build
> logs, the same file read three times. It's default plumbing. So I replayed
> my last 114 real sessions through four dumb hygiene rules — zero model
> calls, nothing touched in Claude Code.
>
> The obvious move, 'just cap big tool results,' would recoup **24%**. But
> watch the last column: I *compute*, for every removed span, whether the
> agent quoted one of those bytes later in the same session. About **1 in 3**
> of those cap/truncate removals deletes an identifier the agent came back
> for — and the sweep shows that never goes to zero at any threshold. **There
> is no safe blind cut.** The one genuinely free rule is stripping terminal
> noise: **0% reference rate, computed.**
>
> The product isn't the 0.4% of free savings — it's this harness. It priced
> four rules and a ten-point sweep in seconds and told me *don't ship the
> blind cap.* That's the mistake it stops you from shipping."

## Would someone pay (assert it — push back welcome)

**Buyer:** any team about to ship a context-trimming hook, a smaller context
window, or a "summarize old tool output" step. Context/prompt compression is
a live product category with real spend behind it. They are not paying for
the 0.4%; they are paying for the computed **safety gate** that blocks a
lossy 24% cut before it silently degrades their agents. The alternative —
ship it and hope — is exactly what the ~30% ref-rate makes visible.

**Objection you'll raise:** "0.4% free savings is nothing." Right — the value
is the *disproof*: the harness shows the tempting 24% is unsafe, per computed
evidence, in seconds. Preventing one bad hook rollout pays for it.

**Objection:** "Just use a cheaper model." Covered in the README — orthogonal
and multiplicative: the diet removes provably-unused bytes on *whatever* tier
you run, and a tier flip can't tell you your cap hook is lossy.

## Deterministic share (show the tests)

The entire correctness path is code with tests — no model anywhere.

- [`tests/test_rules.py`](tests/test_rules.py) — every rule's transform,
  fire and no-op cases.
- [`tests/test_safety.py`](tests/test_safety.py) — reference detection: a
  lost symbol that reappears is flagged; one that survives in the kept text
  is not; ANSI residue scores zero salient tokens.
- [`tests/test_fixtures.py`](tests/test_fixtures.py) — the planted-waste
  numbers asserted exactly (3× re-read → 2 dupes; the escape-heavy log; the
  risky truncation naming `handleAuthRetry_v2`).
- [`tests/test_reader.py`](tests/test_reader.py) — transcript parsing and
  "text that follows a result" attribution.

```bash
make test    # 27 passed, 0 failed
```
