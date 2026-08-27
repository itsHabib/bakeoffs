# Price table — fill from official pricing, do not trust model memory

Source of truth: https://docs.claude.com/en/docs/about-claude/pricing

Values are **TODO** for the operator to fill. Cache reads, cache writes,
fresh input, and output are priced differently — one column each. Leaving
these blank is deliberate: this harness reports *tokens*, and the dollar
conversion is a single multiply the operator does with real numbers.

| model | fresh input $/MTok | cache write $/MTok | cache read $/MTok | output $/MTok |
|-------|-------------------:|-------------------:|------------------:|--------------:|
| claude-opus-4-x   | TODO | TODO | TODO | TODO |
| claude-sonnet-4-x | TODO | TODO | TODO | TODO |
| claude-haiku-4-x  | TODO | TODO | TODO | TODO |

## Turning saved tokens into dollars

The harness prints **saved tool-result tokens** (a tiktoken/o200k_base
proxy). To price a saving:

```
saved_$ ≈ saved_tokens / 1_000_000 × (fresh-input $/MTok)
```

That is the **floor**. Tool-result content sits in the prompt prefix and is
re-sent on every later turn of the session — as a cache *read* each turn.
So a token removed once is a token not re-read on every subsequent turn:
the true saving is `saved_tokens × (avg turns it would have persisted) ×
(cache-read $/MTok)`, which is larger. We report the conservative one-shot
floor and leave the cache-amplification multiplier to a cache-focused
analysis. Either way the unit is tokens first, dollars by one multiply here.
