# Flow State Lab — frozen brief

## Bet

A small model of a CAM client handing work to a factory workflow service can
find retry and custody failures that are hard to see when task state, remote
state, local receipts, and durable commands are reviewed one file at a time.

The useful boundary is not the model alone: every counterexample must be
exported as plain JSON that an ordinary application test suite can ingest.

## One job

Model-check one upload-and-complete protocol, then compile unsafe traces into
portable regression fixtures.

## State in scope

- local programming task status, including check-in/check-out history semantics;
- local epoch status;
- snapshot readiness;
- remote payload/write status;
- the durable upload receipt stored locally;
- durable completion-command custody and recovery availability;
- the number of remote saves performed for one logical upload.

## Required invariants

1. A local upload receipt reflects a saved or completed remote payload.
2. A locally completed task has a durable upload receipt.
3. A closed epoch is complete locally and remotely.
4. One logical upload performs at most one remote save.
5. A retained completion command exposes a recovery path.
6. Recovery availability is scoped to a retained command.
7. Local invalidation cannot orphan remote work.

## Frozen hard cases

- remote save succeeded, response/receipt was lost, blind retry duplicated it;
- a durable completion command was retained without resume or release;
- a local task completed before the remote-save receipt was durable.
- local invalidation abandoned a saved remote payload.

## Quint thesis

Quint is load-bearing because the failure is an ordering problem across several
independently durable facts. The checker explores reachable interleavings and
returns minimal counterexample traces. It loses if hand-authored scenario tests
find the same failures with less model/implementation drift and equal coverage.

## Done

- the safe model typechecks and verifies for every state reachable within the
  declared bound;
- every unsafe transition produces a counterexample;
- counterexamples are normalized into stable JSON fixtures;
- ordinary runtime tests assert the semantic shape of every generated fixture;
- one command reruns the complete experiment.

## Non-goals

No production service, generic workflow engine, remote API client, database,
incident-data ingestion, scheduler, UI, or claim that the model proves a remote
implementation.
