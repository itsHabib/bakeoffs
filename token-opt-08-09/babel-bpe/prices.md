# Price table (operator fills — do NOT trust model memory)

Token counts become dollars here. The bake-off measures **tokens**; this table
is the only place tokens turn into money, so it stays hand-editable and the
values start as TODO. Fill from the official pricing page — cache reads, cache
writes, fresh input, and output are each priced differently:

<https://docs.claude.com/en/docs/about-claude/pricing>

| tier            | $/MTok (TODO) | notes |
|-----------------|---------------|-------|
| fresh input     | TODO          | payload + spec tokens sent uncached |
| cache write     | TODO          | first send that seeds the prefix cache |
| cache read      | TODO          | repeated prefix (a stable spec/legend rides here) |
| output          | TODO          | not exercised by this bake-off (payloads are input) |

## Why this matters for the format choice

Payloads are almost always **input** tokens. A format's spec/legend, if it sits
at a stable position in the prompt, is paid once at the cache-write rate and
thereafter at the (much cheaper) cache-read rate — which is exactly why the
amortized leaderboard separates `spec_tokens` (paid ~once) from `msg_tokens`
(paid every send). Once you fill the table you can multiply:

    monthly_$ ≈ (msg_tokens × sends × input_$/MTok
               + spec_tokens × cache_write_$/MTok) / 1e6

The bake-off deliberately stops at tokens: the moment a real $/MTok number is
in this table, the winning-format delta becomes a dollar figure without
touching any other file.
