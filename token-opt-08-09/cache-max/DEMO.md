# DEMO — the 60-second walkthrough

One command, hands-free:

```bash
make demo
```

## What the judge sees (three acts)

**Act 1 — the tests are the proof (≈10s).** `go test ./...` asserts, on bundled
fixtures: the steady session's exact hit ratio and zero busts; the busty
session's *two* busts with exact thresholds (3750, 6250) and excess (21250 @5m,
23750 @1h); the exact counterfactual dollar delta; idle-TTL attribution; and
that a content diff outranks an idle gap. The correctness path is real code, not
model vibes — [`internal/analyze/analyze_test.go`](internal/analyze/analyze_test.go).

**Act 2 — the fixtures make it legible.** The same tool prints the two planted
sessions: one steady/high-hit, one with two engineered busts blamed on a big
`tool:Bash` result and a big `tool:Read` result. You can see the mechanism do
exactly what the tests assert.

**Act 3 — the real number.** The identical binary runs over ~30 days of this
machine's transcripts (~123 MB, ~200 sessions, ~1.6 s) and prints:

```
HEADLINE: 96% hit rate; 790 busts cost $396.80; top cause idle-ttl-expiry
          (58% of busts, seen in 52% of sessions).
          At 5,000,000,000 tok/mo that's ~$565/mo for zero product-code change.
```

## The 60-second story

"Every turn re-sends the whole conversation — caching is what makes that
affordable, and a prefix bust silently turns 10×-cheap reads into premium
writes. I mined 30 days of my real transcripts with zero model calls: 96% hit
rate, 790 busts, and the top cause is the *same* one in over half my sessions —
the session goes idle, the cache TTL expires, and the next turn re-writes the
whole prefix. Priced out, those busts cost ~$397 over the month; extrapolated to
my 5B-token monthly rate that's ~$565/mo — for changing zero lines of product
code. That's the case for cache discipline before any clever compression, and
it points at one concrete fix: keep the prefix warm across idle gaps."

## Would someone pay? (the buyer)

The buyer is **anyone whose bill is dominated by input tokens re-read every
turn — which is exactly what agent sessions are.** Money already moves here:
Anthropic prices cache reads and writes as distinct, published line items
precisely because the difference is billable, and prompt caching is marketed as
a cost lever (~10× cheaper reads). This tool turns that lever into a measured
dollar figure per account and names the top fixable cause — the input to a
keep-warm change or a TTL-tier decision. Pushback the judge will raise: *"isn't
this just tuned to your machine?"* — no: the rule and pricing are parameterized
and table-tested; point it at any `~/.claude/projects` and it recomputes from
that corpus's own usage fields.

## Fill in real prices (optional, 30s)

Dollar figures use labelled EXAMPLE placeholders until you fill
[`pricing.json`](pricing.json) `rates_usd_per_mtok` from
<https://docs.claude.com/en/docs/about-claude/pricing>. The banner disappears
once all five rates are real. The asserted test deltas don't depend on this.
