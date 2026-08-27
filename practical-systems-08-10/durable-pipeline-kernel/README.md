# Durable Pipeline Kernel

A small Gleam experiment in making agent pipelines resumable without making the
agent remember the pipeline. The kernel is a pure reducer over an append-only
NDJSON journal. Pipeline definitions supply topology and evidence requirements;
adapters remain outside the kernel.

The primary fixture executes a real loop:

```text
ship -> validate -> e2e -> assure -> land
  ^         |        |
  +---------+--------+
```

E2E fails for `rev-a`, records a refuting receipt, and starts attempt two at
`ship`. The new ship effect produces `rev-b`. Validation and E2E must run again;
the passing receipt for `rev-a` cannot unlock any `rev-b` step. The same reducer
also executes `detect -> diagnose -> repair -> validate -> observe` for a
maintenance pipeline.

## Run everything

Docker is the only host dependency. The script pins both Gleam 1.18.1 and the
container digest:

```powershell
.\check.ps1
```

On macOS or Linux:

```sh
./check.sh
```

The command formats and checks the Gleam source, runs the full test suite,
executes the shipping loop, reconstructs its final view from disk, then launches
a separate evidence-reproduction process.

## What is implemented

- Typed pipeline definitions composed from step specs, outcome rules, and
  exact-subject evidence requirements.
- Durable, fsync-backed NDJSON events with monotone sequence checking and
  torn-tail recovery.
- Pure journal replay: rebuilding state invokes no pipeline effects.
- Honest recovery for `idempotent`, `deduplicated`, `at_least_once`, and
  `manual` effect classes. A prepared but uncommitted at-least-once effect parks
  for reconciliation rather than claiming exactly-once execution.
- Cyclic transitions that preserve prior attempts while making their evidence
  stale for a new subject revision.
- Counterexample dominance: a refuting receipt beats any passing receipt for the
  same claim and subject.
- Pipeline name, version, and digest binding; definition drift refuses replay.
- A portable evidence manifest plus input fixture. A fresh process rereads the
  input, canonicalizes and hashes it, reruns its checks, and compares the new
  verdict with the recorded expectation.
- A derived JSON run view suitable for a CLI, test harness, or existing
  observability tool.

## The replay distinction

There are two independent operations:

1. `gleam run` replays execution history and proves that the journal derives the
   same pipeline state as the uninterrupted reducer.
2. `gleam run -m replay_evidence` reruns the validation recipe from its input
   fixture and emits a new `EvidenceReproductionV1` result.

The first trusts the journal bytes as input. The second tests the underlying
claim again. Neither authenticates an arbitrary receipt or grants authority to
land a change.

## File map

- `src/pipeline.gleam` — pipeline algebra, reducer, evidence gating, recovery.
- `src/journal.gleam` — durable event codec, append, replay, torn-tail recovery.
- `src/examples.gleam` — shipping and maintenance topologies and traces.
- `src/evidence.gleam` — independently rerunnable evidence bundle.
- `fixtures/evidence/` — committed replay manifest and its exact input.
- `test/` — crash, replay, loop-back, stale-evidence, drift, and reproduction
  checks.

## What this does not claim

This is not a scheduler, queue, worker service, dynamic plugin loader, database,
or merge authority. It does not make an effect exactly once. It tests whether
the durable reducer and adapter seam are worth extracting into Ship; if a direct
Ship refactor is smaller and equally clear, that is the preferred result.
