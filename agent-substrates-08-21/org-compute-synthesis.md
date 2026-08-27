# Org + durable compute synthesis

**Status:** preserved design thought, not a design lock or build commitment  
**Date:** 2026-08-22  
**Workbench anchor:** [`itsHabib/workbench#245`](https://github.com/itsHabib/workbench/pull/245), inspected at `c18df7ec26543cbd035eac8025e620cebae1cecd`  
**Bakeoff heads:** Branchroom `5b9ff6e`, Proofline `902d6d5`, Obligation Engine `ad64b24`, Mandate `c768a21`

This note preserves a discussion happening across multiple agent sessions. It
does not record a winner for the bakeoff. Refresh PR #245 and the four local
heads before turning any of this into implementation work.

## The thought

Workbench PR #245 is larger than session continuity, but its current TDD still
primarily defines the **organization plane**: roles, charters, distilled state,
ownership, incarnation, takeover, assignments, and typed messages.

The larger thing we may want is an agent operating substrate with a separate
durable-compute plane. Ownership answers who may continue the work. Durability
and compute must also answer what was attempted, which effects committed, what
is safe to retry, what is still running, what evidence resulted, and how a
replacement continues after a crash.

```text
Org — who owns and understands the work
  │ delegates WorkIntent
  ▼
Compute — durable activities, effects, retry/reconcile, budgets
  │ emits trusted execution/evidence receipts
  ▼
Assurance — what the evidence establishes and what remains open
  │ reports outcomes, lineage, and obligations
  ▼
Authority — what actions may actually occur
```

These are planes with explicit contracts, not four new platforms. Existing
Drive, Ship, Runway, Gate, custody, dossier, and Workbench surfaces should keep
their current responsibilities unless an experiment proves a missing seam.

## What each bakeoff entry contributes

### Branchroom

Branchroom's fresh causal epoch and stale-parent-terminal refusal are directly
relevant to org incarnation and takeover. Its additional contribution is a
controlled experiment artifact: two branches can prove a common declared
prefix, name one perturbation, and compute the first declared divergence.

Likely use: executable counterexamples and conformance fixtures for org and
compute epoch fencing. It is not itself a scheduler, assurance reducer, or
authority system.

### Proofline

Proofline is read-only lineage over verification contracts, receipts, oracle
assurance, and consumers. It answers which identity-changing edge made prior
assurance and downstream observations historical.

Likely use: the assurance/observation plane after compute emits receipts. It is
not the role journal and should not be folded into org ownership semantics.

### Obligation Engine

Obligation Engine derives the open evidence-work frontier from frozen contracts
and immutable assurance snapshots. Late refutation reopens mandatory work;
agent overlays cannot weaken the contract floor.

Likely use: emit typed work that org may assign and compute may execute. It does
not select workers, run activities, judge raw receipts, or grant authority.

### Mandate

Mandate demonstrated the need for exact-work authority that monotonically
narrows across delegation. Its custom signed-JSON format did not beat a standard
caveated capability plus an audit receipt and remains killed as a standalone
format.

Likely use: an authority-profile requirement for org incarnation grants and
compute effects. It does not solve revocation propagation, global replay, or
the takeover-versus-grant-spend fencing problem identified in PR #245's design
review.

## Where Gleam, Lean, and Quint could be load-bearing

### Go — production contracts and gateways

Keep production contract law, canonical validation, append/fencing gateways,
and Workbench/Drive/Gate integration in Go. `contracts/org` and a possible
`contracts/compute` should have closed types, stable refusal codes, pure folds,
conformance fixtures, fuzz tests, and no host decisions.

### Gleam/OTP — live compute, not durability itself

The Switchboard experiment established a crucial correction: a journal plus a
pure reducer supplies crash recovery to either a resident or stateless host.
What the actor uniquely supplied was serialized ownership under concurrent
turns.

Gleam/OTP earns a place only where a live host needs supervised workers,
serialized mailboxes, backpressure, in-flight operation ownership, or restart
trees. Durable truth remains the journal. A chain plus a correctly locked
read-verify-append path may already be enough for cold role incarnation; do not
add a resident actor unless a fair comparator exposes a concurrency or live
compute failure.

Possible seam: a Gleam `compute-host` accepts already-validated work intents,
supervises disposable workers, and submits typed activity/effect events through
the Go contract gateway. It does not reimplement org, assurance, or authority
policy.

### Lean — unbounded pure laws

Lean can prove the laws that should hold for every finite history and every
cut, rather than only sampled traces:

- fold concatenation and checkpoint/resume equivalence;
- replay equivalence across arbitrary cuts;
- fresh incarnation and task-epoch isolation;
- takeover invalidates all later writes from the displaced incarnation;
- duplicate event idempotence where identity is exact;
- monotone authority attenuation and monotone mandatory obligations, if their
  abstract reducers are included.

Lean does not prove filesystem atomicity, fsync behavior, adapter correctness,
event authenticity, or real concurrency. Those boundaries must remain explicit.

### Quint — bounded concurrent counterexamples

Quint should model the interleavings that the prose design currently leaves
dangerous or underspecified:

- two incarnations read one tip and race read-verify-append;
- takeover races a displaced incarnation's grant spend;
- cap revocation and chain takeover diverge across two planes;
- a crash lands between effect preparation, remote completion, and local
  receipt recording;
- retry encounters an effect whose outcome is unknown;
- two supervisors race a takeover;
- sender and receiver chains observe cross-referenced messages in different
  orders.

Known-bad actions should generate minimal normalized counterexample traces.
Go and Gleam tests should consume those fixtures. Quint proves the bounded
model, not production, so the conformance adapter is part of the experiment.

## Existing material to reuse

- `the switchboard project (not published)` — durable journal/reducer parity
  and the serialized-ownership result.
- `the repair-loop-kernel project (not published)` — reference semantics for durable repair
  activities, exact-subject evidence, effect classes, replay, and
  reconciliation. It is a conformance model, not a runtime.
- `https://github.com/itsHabib/formal-methods/blob/main/entries/fm-epoch-replay-laws/` — Lean fold/replay and epoch-isolation
  laws, with explicit claims boundaries.
- `practical-systems-08-10/flow-state-lab/` — Quint
  counterexample export into ordinary regression fixtures.
- `https://github.com/itsHabib/formal-methods/blob/main/evidence-pr/` — the existing assurance oracle
  floor; do not rebuild it inside org or compute.

## One integrated experiment

Use one IC role, one work item, one durable activity, one external effect, and
one supervisor. Avoid starting with a 50-agent organization.

1. The role incarnates and accepts an exact work intent.
2. Compute prepares an external effect and records the attempt generation.
3. The worker crashes at a deliberately chosen cut.
4. The supervisor records takeover and starts a replacement.
5. The displaced incarnation attempts a stale append or grant spend and is
   fenced with a stable reason.
6. The replacement folds org state and compute history, determines whether the
   effect is committed, absent, or uncertain, and reconciles before retry.
7. The trusted runner emits receipts.
8. Assurance reduces them; Proofline records lineage; Obligation Engine exposes
   any mandatory work still open.

The same scenario should have four views:

- Go executes the production contract path.
- Gleam hosts and supervises the live worker if residency wins its comparator.
- Quint explores bounded crash/takeover/effect orderings and exports red traces.
- Lean proves the abstract fold, replay, and epoch laws.

The experiment wins only if the languages agree through ordinary artifacts and
each language provides leverage unavailable from direct Go tests alone.

## Decisions still open

1. Keep durable compute as a sibling contract/TDD, or expand PR #245. Current
   preference: keep `org` focused and add a sibling `compute` design.
2. Define the activity/effect state machine and its exact-subject identity.
3. Decide whether Gleam's resident host beats a stateless durable host for this
   specific compute scenario.
4. Name the authoritative fencing point for takeover versus grant spend.
5. Define one conformance artifact shared by Go production tests, Gleam host
   tests, Quint counterexamples, and Lean claims mapping without pretending the
   four implementations are automatically equivalent.
6. Decide what state is portable across machines and how chain, journal, and
   anchor material are synchronized.
7. Record the bakeoff promotion decision. This note deliberately does not make
   it for the operator.

## Guardrails

- Do not call actor memory durability.
- Do not build a second generalized runtime before one real activity proves a
  missing execution seam.
- Do not let org decide evidence quality or merge authority.
- Do not let assurance grant authority.
- Do not turn Quint counterexamples or Lean theorems into claims about adapters,
  filesystems, authentication, or production without conformance evidence.
- Do not make the Gleam host a second policy implementation.
- A negative result is useful: direct Go plus a journal may beat the proposed
  host, proof, or model for a given seam.

## Sources

- Workbench PR #245: <https://github.com/itsHabib/workbench/pull/245>
- Exact reviewed org TDD: <https://github.com/itsHabib/workbench/blob/c18df7ec26543cbd035eac8025e620cebae1cecd/docs/features/org/spec.md>
- PR #245 Claude design review: <https://github.com/itsHabib/workbench/pull/245#issuecomment-5377739802>
- `switchboard's gate-c2 comparator evidence (not published)`
- `the repair-loop-kernel project (not published)`
- `https://github.com/itsHabib/formal-methods/blob/main/entries/fm-epoch-replay-laws/CLAIMS.md`
- `practical-systems-08-10/flow-state-lab/README.md`
