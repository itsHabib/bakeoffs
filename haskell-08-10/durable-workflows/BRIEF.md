# Durable workflows — frozen brief

## The bet

A durable workflow should be a serializable program value with deterministic
replay semantics, not an opaque callback whose correctness depends on scattered
idempotency conventions. Haskell earns its seat if one workflow syntax can be
interpreted for fresh execution, crash replay, and explanation while making the
effect boundary and stable step identity explicit.

## User and concrete workflow

An agentic delivery engine performing dispatch → wait → record, where a process
may die after an external effect succeeds but before later steps run. The useful
output is recovery that reuses completed results and never repeats a recorded
effect.

## One job

Execute and replay a serializable sequence of named activities so that a
recorded completed activity is not performed again after interruption.

## Build

- A typed workflow AST with pure values, named activities, sequencing, and
  explicit failure/compensation boundaries.
- A journal containing stable activity IDs and encoded deterministic results.
- Fresh and replay interpreters over the same syntax.
- Crash injection after a chosen activity commit.
- Diagnostics for changed workflow shape or conflicting recorded results.

## Required hard cases

- Crash after commit but before continuation does not duplicate the activity.
- Crash before commit may retry and is described honestly as at-least-once.
- Reordered or duplicate journal entries are rejected.
- Workflow-version drift is detected instead of replaying against new meaning.

## Haskell thesis

The workflow is a closed program algebra with multiple interpreters and explicit
result types; deterministic replay is defined over that value rather than an
ambient effectful function. The candidate loses if serializability forces an
awkward untyped encoding or if it merely recreates an ordinary state machine.

## Demo

Run dispatch → verify → record, crash after dispatch is committed, replay, and
show dispatch executed once while later activities complete. Then change the
workflow's stable step identity and show replay refusal rather than accidental
reuse.

## Kill condition

Kill if useful workflows cannot remain serializable without a bespoke compiler,
if result encoding destroys type safety, or if the model duplicates Ship or
Temporal without yielding a sharper semantic advantage.

## Non-goals

No scheduler, worker fleet, database, timers, parallel branches, network API,
multi-version migration framework, or agent-provider integration.
