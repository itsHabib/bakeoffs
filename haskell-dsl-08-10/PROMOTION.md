# Promotion handoff — `work-driver-dsl` → `braid`

The promotion thesis was refined on 2026-08-12 into an agent-facing compiler
over heterogeneous work sources, task-specific shipping policy, and a frozen
`ShippingPlanV1`. See [`AGENTIC-DELIVERY-COMPILER.md`](AGENTIC-DELIVERY-COMPILER.md).

## One job

`braid` compiles an authored project plan into deterministic, dependency-correct,
file-conflict-safe batches for the existing work-driver boundary.

It does not dispatch, poll, retry, review, merge, or own task state.

## First real experiment

Choose one upcoming Dossier phase that would normally go through
`work-driver-prep`. Author it in both the current flow and `braid`, then compare:

1. resolved dependencies and maximum parallel batches;
2. every file-scope conflict and the explanation for serialization;
3. the critical path and validation-before-landing check;
4. the exact manifest consumed by work-driver;
5. authoring time and how many corrections were needed.

The vertical slice succeeds only if the generated manifest can enter the
existing driver unchanged or through one deliberately tiny adapter.

## Production responsibilities

- a stable, documented project DSL with opaque in-plan task references;
- deterministic compilation independent of map/set iteration order;
- a precise scope-overlap model with visible conservative decisions;
- validation of duplicate names and required landing gates;
- a versioned output format matching the existing work-driver consumer;
- golden and property tests for batching, determinism, and diagnostics.

## Deliberate exclusions

No task database, Dossier replacement, Ship replacement, agent runtime,
worktree manager, retry engine, scheduler service, dashboard, GitHub client,
capability policy, assurance engine, or generic workflow framework.

## Kill the promotion if

- real plans need so many escape hatches that the DSL stops being clearer than
  work-driver-prep's current documents;
- declared file scopes are too inaccurate to make batching decisions useful;
- requiring a Haskell toolchain creates more friction than the compiler removes;
- stable work-driver output requires importing its whole execution model;
- users must understand implementation details to predict a batch.

The next question is not which feature to add. It is whether one real phase is
more pleasant to author and produces a safer manifest than today's flow.
