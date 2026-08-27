# Entry 2: hack-readback — voice-gated irreversible actions

Read ~/dev/interject/docs/hackathon/README.md first — house rules and judging
apply verbatim. Repo: `~/dev/hack-readback`. You never see the other three
entries.

## The bet

Aviation made irreversible actions safe over a lossy audio channel with the
readback/hearback loop: order → canonical readback → explicit confirm →
execute. Nobody has brought that discipline to voice-driven ops. The demo:
you SPEAK an order, the system reads back the exact canonical action it
parsed — IDs normalized, dangerous parameters spelled out — and NOTHING
executes until you speak the confirm phrase. Mishears become harmless:
a wrong readback just doesn't get confirmed.

## What to build

A browser console for a SIMULATED ops world (fake but plausible: a few
repos with PRs, a couple of deploy targets, some feature flags — seed data
in a JSON file).

- Command grammar in code: a small closed set of verbs
  (`merge <repo> pr <n>`, `deploy <service> to <env>`, `flag <name> on|off`,
  `rollback <service>`). Parse with deterministic matching (fuzzy on repo
  names is fine, but the RESULT is exact) — this is roll-call
  matcher.go's trick, grown up. No LLM parses commands.
- Readback: the system speaks (speechSynthesis) the canonical form with the
  risky bits made unmistakable — digits read individually ("PR one-four"),
  environment stressed ("to PRODUCTION"). Screen shows the parsed action as
  a structured card simultaneously.
- Confirm phrase must echo a changing token (e.g. "confirm one-four" /
  "negative") so a stale "confirm" from three commands ago can't arm this
  one. Timeout → order expires, spoken "order expired".
- Execute = mutate the fake world + append to an audit log the judge can
  read: heard text → parsed action → readback → confirmation → effect,
  timestamped. The audit trail IS the pitch.
- Two-mishear demo: a deliberately ambiguous order ("merge roxiq" when two
  PRs are open) must come back as a spoken CLARIFYING readback, not a guess.

Mic path via Web Speech; typed input must drive everything identically.

## What NOT to build

No real gh/git/deploy integration — the fake world is the point (and the
house write-boundary). No LLM anywhere in the command path; if you want one
at all, phrasing flavor only. No user accounts, no RBAC theater.

## Canned demo (required)

A scripted run: clean order → readback → confirm → executes; then a mishear
(script feeds "merge PR forty" when they said fourteen) → readback exposes
it → "negative" → nothing happened; then the ambiguous order → clarify.
Audit log shown at the end.

## The 60-second demo story

"I just told it to merge PR fourteen. It heard forty. Watch what happens —
it reads back 'PR four-zero', I say negative, and the audit log shows a
mishear that cost nothing. Voice ops isn't unsafe because recognition is
imperfect; it's unsafe without readback. This is the readback."
