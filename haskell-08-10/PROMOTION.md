# Promotion handoff — `durable-workflows` → `reprise`

## Decision

Promote the semantic bet, not the bakeoff implementation. `reprise` gets one
job:

> Resume a versioned agent-delivery workflow from an append-only journal,
> reusing every compatible committed activity result and refusing replay when
> identity, order, input, version, or result encoding is ambiguous.

This is intentionally not “a workflow platform.”

## First real vertical slice

Use one existing delivery sequence—dispatch → verify → record—as the only
production workflow. Give each activity a stable ID, typed input/output codec,
and declared retry behavior. Persist the journal durably, kill the process at
every effect/commit boundary, and demonstrate:

1. a committed dispatch is never performed twice;
2. an effect that happened before its missing commit is reported and retried as
   at-least-once, never mislabeled exactly-once;
3. changed workflow identity or input refuses replay with a useful diagnostic;
4. a clean run and every compatible resumed run produce the same final receipt.

The first adapter may invoke an existing Ship-like command boundary. It should
not absorb Ship's scheduling, policy, review, or merge responsibilities.

## What must become production-grade

- durable append/commit semantics with recovery tests, likely behind one small
  journal interface;
- explicit activity retry classes and deduplication keys where the external
  boundary supports them;
- property/model tests that inject crashes at every interpreter transition;
- schema/version refusal and a deliberate migration story before any workflow
  version changes in place;
- structured explanations for “reused,” “executed,” “retrying at-least-once,”
  and “refused.”

## Deliberate exclusions

No scheduler, server, worker fleet, timers, arbitrary DAGs, UI, plugin system,
policy DSL, generalized effect framework, protocol compiler, provenance engine,
or bidirectional artifact editor. Do not fold in `capability-plans` unless the
real adopter first produces an approval problem that requires static preflight.

## Kill the promotion if

- the real workflow needs so much dynamic code that its serializable typed value
  becomes ceremonial;
- journal codecs and migrations erase the stage-safety advantage;
- integrating the Haskell boundary costs more complexity than the replay
  semantics remove;
- an existing small durable-execution library supplies the same visible
  semantics without importing a platform.

The next decision is therefore not “which features should `reprise` have?” It
is “can this one real delivery sequence survive exhaustive crash injection with
the POC's semantics intact?”
