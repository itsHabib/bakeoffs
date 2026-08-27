# Durable Pipeline Kernel — frozen brief

## Bet

Agent pipelines should not depend on one long-lived agent remembering where it
was. A small embedded kernel can make a pipeline's topology, durable state,
effect semantics, and evidence requirements explicit while leaving every real
action behind a replaceable adapter.

The target is the recurring shape behind shipping and maintenance, not a new
job-running service. One process performs one bounded tick, appends what
happened, and may disappear. A later process reconstructs the run from the same
journal and continues safely.

## One job

Execute and resume a versioned, evidence-directed pipeline—including an
intentional loop—without duplicating a committed effect or accepting evidence
for the wrong subject revision.

The primary demonstration is:

```text
ship -> validate -> e2e -> assure -> land
  ^         |        |
  +---------+--------+  failure starts a new attempt with the failure artifact
```

The same unchanged kernel must also execute a maintenance topology:

```text
detect -> diagnose -> repair -> validate -> observe
                      ^           |
                      +-----------+
```

Pipeline definitions choose the transitions. The kernel owns replay and state;
adapters own effects; evidence evaluators own claims; an external authority owns
merge or production mutation.

## Runtime contract

- A run is bound to a pipeline version and an exact subject revision.
- The journal is append-only. Its events carry run, node, attempt, input digest,
  effect identity, outcome, and evidence references.
- A tick reduces the journal, chooses at most one eligible node, invokes its
  adapter, and appends one typed outcome: `completed`, `failed`, `waiting`, or
  `parked`.
- Every effect declares one honest replay class: `idempotent`, `deduplicated`,
  `at_least_once`, or `manual`.
- A loop creates a new attempt and invalidates downstream evidence. It never
  erases the failed attempt or silently reuses success from an older subject.
- A changed pipeline version or incompatible journal shape refuses resume unless
  an explicit migration is supplied.
- Snapshots are disposable caches. The journal remains the source of truth.

Adapters are ordinary Gleam modules supplied in process; there is no dynamic
plugin loader. The portable outputs are NDJSON journal events plus a derived JSON
run view that existing tools can inspect.

## Two kinds of replay

Durable state replay and evidence replay are different promises, and the kernel
must support both:

1. **Execution replay** reduces the append-only journal without invoking effects.
   Any process given the same pipeline definition and journal must derive the
   same current node, attempts, invalidated evidence, and next legal actions.
2. **Evidence reproduction** reruns a validation or verification recipe against
   its exact subject. Another person must be able to take the evidence bundle to
   a clean checkout, run one command, and obtain the same semantic outcome or a
   precise explanation of which unavailable dependency prevents reproduction.

A validation receipt therefore contains more than `passed`: subject and pipeline
digests, tool and adapter versions, input artifact hashes, command, declared
environment inputs, seed, bounds, exit semantics, and output hashes. Secret
values are never captured. Nondeterministic or external validation must emit a
portable minimized fixture when identical re-execution is impossible.

Reproduction creates a new evidence observation; it does not authenticate the
old file or inherit its authority. Replaying a journal proves what the reducer
derives from those bytes. Rerunning the recipe tests the claim independently.
Neither one authorizes `land` without the configured exact-subject evidence and
external authority boundary.

## Evidence boundary

The [formal-methods evidence-carrying PR experiment](https://github.com/itsHabib/formal-methods/pull/1)
supplies the model for stage handoffs: a passing result is useful only for its
exact subject and producer context, a counterexample dominates passing samples,
and publication is not authorization.

This kernel carries and freshness-checks evidence references. It does not decide
that a PR is correct, authenticate arbitrary local JSON, lower a risk floor, or
gain merge authority. Assurance remains a separate adapter and exact-SHA action
remains an external boundary.

## Frozen hard cases

1. The process dies after dispatch succeeds but before the journal commit. A
   deduplicated adapter reuses the same external key; an `at_least_once` adapter
   reports possible duplication instead of claiming exactly-once execution.
2. E2E fails for revision A. The run records the failure, loops to `ship` as a
   new attempt, and refuses to let A's earlier validation unlock revision B.
3. The process dies after a completed journal event but before the next tick.
   Resume must not invoke the completed adapter again.
4. The pipeline definition changes between ticks. Replay refuses the new shape
   rather than interpreting old events under new meaning.
5. The shipping and maintenance examples use the same reducer and journal code;
   only their topology and adapters differ.
6. A second clean checkout consumes the validation bundle, reruns its recorded
   recipe, and produces the same semantic verdict without trusting the first
   runner's derived state.

Each hard case must export a deterministic trace fixture consumable by an
ordinary test suite.

## Gleam thesis

Gleam is load-bearing if algebraic outcomes and exhaustive matching keep every
transition, replay class, and adapter failure explicit while the immutable
reducer stays small enough to audit. BEAM supervision may restart a process, but
the experiment must demonstrate that restart is not durability—the persisted
journal is.

It loses if Ship's TypeScript driver can be extracted into an equally small,
equally explicit kernel with less translation, or if adapter boundaries become
stringly runtime plumbing that erases the type advantage.

## Relationship to existing work

- Ship already proves a fixed durable `dispatch -> poll -> judgment -> land ->
  record` pipeline is useful in practice.
- The Haskell durable-workflows and recovery-DSL entries proved replay and
  effect-class semantics for a closed sequence.
- This experiment tests the missing claim: whether cyclic topology plus
  exact-subject evidence can be reusable across two real pipeline families
  without importing their policy.

If the best result is a modest refactor of Ship rather than a separate kernel,
that is a successful negative result. Promote the seam, not the demo.

## Done

- one command runs both example pipelines and every crash/replay fixture;
- restarting from the journal is observationally equivalent to uninterrupted
  execution for committed effects;
- loop-back invalidates stale downstream evidence and preserves prior attempts;
- effect-before-commit behavior is honest for all four replay classes;
- the derived JSON view identifies the next action or exact parked question;
- every validation result exports a self-contained replay manifest, and a fresh
  process can reproduce its semantic verdict or report an explicit missing input;
- no engine change is required to switch from shipping to maintenance.

## Non-goals

No scheduler, cron service, worker fleet, queue, database, hosted control plane,
distributed consensus, arbitrary remote plugin loading, secrets broker, UI,
exactly-once claim, evidence evaluator, or merge/production authority.
