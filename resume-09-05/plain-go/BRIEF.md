# Entry 1: Plain Go — the serious baseline

Read [`../RULES.md`](../RULES.md) first. All cases, effect contracts, invariants,
and finish-line requirements apply.

This entry was built in a fresh repository without inspecting its competitors.

## The bet

One small explicit state machine, a journal, and a few effect functions may provide the required recovery behavior with less maintained machinery than a reusable kernel or protocol language. This entry gets every substantive check the other entries get. It must be a competent implementation, not intentionally naive code.

## What to build

Use idiomatic Go, preferably the standard library. Implement the common artifact/check/control/delivery workflow, its journal loader, fresh-process resume, and both sink modes. Keep state transitions explicit and effect handling separate. Name unresolved outcomes and identity refusals.

The test controller owns disposable worker processes, waits for explicit boundary handshakes, terminates them, and launches replacements. Assert correctness from the retained journal, content identities, and sink effect records. The normal command must run the good case and demonstrate the difference between deduplicated recovery and an opaque unresolved effect.

Implement the common mutants and show the tests catching them. Do not import Warrant or Parley. Small helpers are fine; a generic workflow framework is unnecessary.

## What not to build

No scheduler, queue service, DSL, actor framework, dashboard, cross-host replication, live agent integration, or authority system. Scenario incarnation values are test inputs, not a replacement for Org. Do not model exactly-once behavior where the effect contract does not supply it.

## Canned demo and story

“The worker delivered the artifact, but died before learning that delivery committed. Its replacement asks the sink and finds the existing receipt: one effect, one completed logical operation. In the opaque case it cannot ask, so it stops with the unresolved operation. A stale worker result and a changed candidate cannot sneak through either.”

## Finish

One-command README, hands-free demo, DEMO.md, tests green, artifacts with reproducible commands. Report maintained code and operational steps honestly. Keep the build local.
