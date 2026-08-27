# readback — voice-gated irreversible actions

Aviation made irreversible actions safe over a lossy audio channel with the
readback/hearback loop. This is that loop for ops:

**order → canonical readback → spoken confirm token → execute.**

You say `merge roxiq pr fourteen`. The system reads back the exact action it
parsed — digits spelled individually ("P R one four"), environment stressed
("to PRODUCTION") — and **nothing executes** until you speak the confirm
token it gave you. A mishear becomes harmless: the readback exposes it, you
say "negative", the audit log records a mishear that cost nothing.

## Run

```bash
go run .
```

Open http://localhost:8014 (Chrome or Safari for voice). Click
**▶ canned demo (no mic)** — a fully scripted run with spoken readbacks:
clean order, a mishear caught by readback, an ambiguous order answered with
a clarifying question. No microphone needed. `DEMO.md` has the 60-second
walkthrough.

Live use: type orders in the input box, or hit **🎤 speak** (Web Speech,
Chrome/Safari). Typed and spoken input drive the identical path.

## The world

Simulated ops estate seeded from [world.json](world.json): three repos with
PRs, two services with staging/production versions, two feature flags. No
real git/gh/deploy anywhere — the fake world is the write boundary.

## Grammar (closed set, no model anywhere)

```
merge <repo> [pr <n>]        deploy <service> to <env>
flag <name> on|off           rollback <service>
confirm <token>              negative
```

Parsing, matching, gating, expiry, and world mutation are deterministic Go
in [policy/](policy/) — table-tested, zero model calls, zero dependencies
beyond the stdlib. Repo names match fuzzily (closed-set edit distance), but
the *result* is always exact; a tie is a spoken clarifying question, never a
guess. Spoken numbers ("fourteen", "one four", "40") normalize before
matching, and the readback's own spelling always round-trips.

The confirm token is the risky digits themselves (PR number, version), so
confirming forces you to re-speak the exact value that could have been
misheard — that's the hearback half of the loop. Orders without digits get a
random two-digit challenge. Tokens change per order, so a stale "confirm one
four" from three commands ago can never arm the current one. Unconfirmed
orders expire after 25 seconds, out loud.

Every exchange lands in the audit trail: heard text → parsed action →
readback → confirmation → effect, timestamped. The audit trail is the
product.

## Tests

```bash
go test ./...
```

All correctness lives in `policy/` and is table-tested: the grammar
([parse_test.go](policy/parse_test.go)), number/name normalization
([numbers_test.go](policy/numbers_test.go)), and the gate's safety
properties ([gate_test.go](policy/gate_test.go)) — wrong token never
executes, stale token can't arm the next order, expiry and negatives leave
the world untouched, parse never mutates anything.
