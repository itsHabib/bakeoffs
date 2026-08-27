# Braid — agentic delivery compiler

**Status:** design note, not a build commitment

**Date:** 2026-08-12

**Lineage:** [`work-driver-dsl`](work-driver-dsl), [`PROMOTION.md`](PROMOTION.md)

## Thesis

Braid is not valuable because a human or agent would rather write a Haskell DSL
than `driver.md`. An agent can already generate a manifest. Braid earns a place
only if it compiles messy, mutable work intent into a frozen delivery program and
rejects unsafe programs before Ship executes them.

The useful product is an **agentic delivery compiler**:

```text
Dossier task / Jira ticket / GitHub issue / spec / raw prompt
                              +
portfolio floor / repository floor / task and PR requirements
                              |
                              v
                     Braid compiler
                              |
          +-------------------+-------------------+
          |                   |                   |
          v                   v                   v
     source lock       ShippingPlanV1       explanation
                              |
                              v
                   driver.md today
                   direct Ship import later
```

The human interface may remain conversational. The operator says what should be
shipped; an agent authors the structured source; Braid resolves, checks, and
compiles it. The operator should not need to write Haskell or another programming
language by hand.

## Why an agent would use it

Direct `driver.md` generation remains the strongest alternative. Braid must do
work that generation alone cannot safely do:

1. Resolve heterogeneous work references and freeze exactly what was read.
2. Detect when two references name the same underlying work.
3. Build and check dependency, evidence, and authority graphs.
4. Compute safe parallel batches and explain forced serialization.
5. Compose portfolio, repository, task, and actual-PR policy without allowing a
   task or agent to weaken a safety floor.
6. Bind review, CI, and Gate evidence to the exact candidate head they certify.
7. Refuse to lower a requirement that the target runtime cannot represent.

If one real comparison shows that an agent-authored `driver.md` is equally safe
and legible, kill Braid and improve the existing manifest generator instead.

## Source language

The source describes intent, not mutable execution state. A conceptual form:

```text
delivery "driver-hardening" using portfolio_pr_v1 {
  parser = task dossier("tsk_01K...") {
    touches ["packages/driver/src/manifest.ts"]
  }

  engine = task jira("SHIP-142") after parser {
    touches ["packages/driver/src/engine.ts"]
    require review("security")
  }

  docs = task prompt("Document the recovery boundary") after parser

  verify [engine, docs] {
    check "make check"
    review configured_panel
    exact_head
  }

  land after verified([engine, docs]) {
    gate max_tier T2
  }
}
```

The syntax is illustrative. JSON, YAML, a small external DSL, or a typed API are
all viable. Semantic leverage matters more than surface notation.

### Work sources

```text
TaskSource =
    Dossier(task_id)
  | Jira(ticket_key)
  | GitHub(issue_or_pr)
  | Spec(path)
  | Prompt(text)
```

Resolution produces an immutable lock entry:

```yaml
kind: jira
locator: SHIP-142
source_revision: "17"
content_digest: sha256:...
resolved_at: 2026-08-12T00:00:00Z
captured_payload: ...
```

Where an external system exposes no trustworthy revision, Braid canonicalizes
the relevant payload and hashes it. Resume uses the captured source. A later
source change is drift requiring an explicit re-resolve and new plan revision,
never a silent change to a running program.

## Policy belongs in the delivery program

Repository-wide `.ship.json` remains useful for non-negotiable floors and local
dispatch constraints, but it is too blunt to express the entire shipping policy.
A docs PR and an authentication PR in the same repository may require different
checks, reviewers, and authority.

Braid should support task- and PR-specific requirements while composing policy
monotonically:

```text
portfolio floor
  + repository floor
  + compiled task requirements
  + actual PR risk escalation
  = effective policy for one exact PR head
```

Each dimension needs an explicit composition law:

| Dimension | Composition rule |
|---|---|
| required reviewers | union |
| required checks | union |
| allowed runtimes/providers | intersection |
| sensitive paths | union |
| settling period | maximum |
| review-cycle limit | most restrictive cap |
| authority ceiling | most restrictive ceiling |
| human confirmation | required if any layer requires it |
| autonomous action | allowed only if every layer permits it |

An agent may add requirements. It may not remove the repo's required reviewer,
widen a provider allow-list, increase its own Gate tier, or remove an operator
confirmation boundary.

### Two-stage compilation

The final diff and PR do not exist when work is first compiled, so policy cannot
be fully specialized in one pass.

```text
Stage 1: sources + declared requirements -> WorkPlanV1

Stage 2: WorkPlanV1 + PR + exact head + actual diff + risk result
                                      -> EffectiveDeliveryPolicyV1
```

If a task was compiled with a T2 authority ceiling but the actual PR is T3, the
result is a typed park, not an automatic widening. If the head changes after
review or Gate, its evidence is stale and a new specialization is required.
Grant minting remains operator authority; Braid can encode that a valid grant is
required but can never create one.

## Evidence as obligations

The compiler does not claim that future tests or reviews will pass. It proves
that every path to landing carries the required obligations.

Conceptually:

```text
Candidate<repo, sha>
  -> TestedCandidate<repo, sha>
  -> ReviewedCandidate<repo, sha>
  -> GatePassedCandidate<repo, sha>
  -> LandableCandidate<repo, sha>
```

A review receipt for `abc123` cannot produce `ReviewedCandidate<def456>`. A
passing check with no subject identity cannot discharge an exact-head check.
Ship and Gate supply the runtime receipts; Braid declares and checks the graph
of obligations those receipts must satisfy.

## Intermediate representation

The canonical output should be an immutable, content-addressed plan rather than
Markdown that doubles as a resume database.

```yaml
schema: ShippingPlanV1
plan_id: pln_...
plan_digest: sha256:...
compiler:
  name: braid
  version: 0.1.0
sources: [...locked sources...]
work: [...tasks, repositories, pinned bases, declared scopes...]
requirements: [...checks, reviews, Gate and authority obligations...]
control_flow: [...dependencies, parallel requests, bounded loops...]
terminals: [succeeded, exhausted, parked, refused]
```

Four kinds of truth stay distinct:

| Artifact | Owns |
|---|---|
| source lock | what external work was frozen |
| `ShippingPlanV1` | what execution and evidence are required |
| Ship or another runtime store | what actually happened |
| receipts | why a transition was admitted |

Initial outputs should be:

```text
sources.lock.json
shipping-plan.json
generated task specs
driver.md
explain.md
```

`driver.md` is a compatibility target for today's Ship driver. If `driver-v1`
cannot represent a checked requirement, compilation must refuse rather than
hiding that requirement in advisory prose.

## System boundaries

### Braid

Owns source resolution, locks, policy composition, dependency/evidence/authority
analysis, parallelism planning, lowering, and compiler diagnostics. It performs
no agent work and owns no runtime state.

### Ship

Owns durable execution today: import, dispatch, polling, persistence, retries,
judgment parks, review-address flow, landing, and receipts. Braid should compile
to Ship, not rebuild it. See the live [`@ship/driver`](../../ship/packages/driver)
package.

### Drive

Owns the operator's portfolio-shaped view and the attachment of sessions to
work. It should link a Drive scope to a Braid plan, a Ship driver run, worker
sessions, tasks, branches, and PRs. It should not interpret the Braid language or
duplicate Ship's scheduler. See [`drive-v0`](../../drive/docs/features/drive-v0/spec.md).

### Gate and the operator

Gate owns merge authorization semantics. The operator remains the authority
root for grants and irreversible policy changes. Braid describes required
authority; it cannot manufacture it.

### Repair Loop Kernel and RoxIQ

RoxIQ is the first genuinely hard control-flow consumer. Repair is not retry:
a repair creates a new subject, invalidates evidence about the prior subject,
and begins another bounded attempt.

The [`Repair Loop Kernel`](../../repair-loop-kernel/SPEC.md) defines the reusable
semantics: attempt identity, stale-evidence refusal, counterexample dominance,
negative controls, honest replay classes, named terminals, and drift refusal.
Braid may eventually type-check a bounded `repair_loop`; Ship or another runtime
executes it; RoxIQ supplies product-specific inject, repair, verify, audit, and
publish adapters.

RoxIQ's Go gauntlet is currently a deliberate single forward pass with one
repair attempt. Its next claim is not "we added a loop" but approximately fifty
repairs with verified recovery in the disposable environment while production
actions remain human-confirmed.

### Possible Gleam/OTP runtime

A sophisticated Gleam/OTP runner is a plausible later interpreter for the same
`ShippingPlanV1`, not the next step. OTP supplies isolation, supervision, timers,
registries, and bounded concurrency, but not durable effect semantics for free.

The existing [`agents-as-processes-gleam`](../../agents-as-processes-gleam)
experiments established the required shape in narrower domains: the journal,
not actor memory, owns durable truth; `decide` produces events; `evolve` is a
pure replay fold; effects dispatch only after commit; stable IDs make late and
duplicate completions decidable.

Only after Ship has executed real compiled plans should a Gleam backend consume
the same IR and conformance corpus. It earns promotion only if crash injection
shows equivalent semantics with materially clearer operation than Ship.

## First falsifiable experiment

Choose one upcoming Dossier phase that would ordinarily go through
`work-driver-prep`.

Run two arms:

1. An agent directly authors `driver.md`.
2. The same agent authors Braid source, and Braid compiles it to `driver.md`.

Inject these cases before execution:

- one external source changes after resolution;
- the same task is referenced through two systems;
- a landing path lacks an exact-head verification ancestor;
- two nominally parallel tasks have overlapping scopes;
- a task tries to weaken a repository review floor;
- actual PR risk exceeds the compiled authority ceiling;
- the source asks for a verification stage `driver-v1` cannot represent.

Success requires all of:

- Braid catches at least one material defect the direct-manifest path misses;
- the generated manifest imports into Ship unchanged or through one tiny,
  explicit adapter;
- the explanation makes every serialization and policy escalation legible;
- the agent needs fewer corrections or produces a measurably safer artifact;
- no model call occurs inside compilation.

Otherwise stop. Do not proceed to a generalized language or a Gleam runtime.

## Development ladder

1. **Compiler experiment:** Dossier + spec + prompt sources, locks, graph checks,
   policy floors, `ShippingPlanV1`, `driver-v1`, and explanation.
2. **Artifact specialization:** bind one real PR/head/diff to an effective
   task-specific review and Gate policy; prove stale-head refusal.
3. **Drive integration:** render one scope from linked plan, driver run, workers,
   PRs, and evidence without copying their state.
4. **RoxIQ advanced control flow:** make the Go gauntlet consume the repair-loop
   conformance cases; only then add bounded repair to the plan language.
5. **Runtime comparison:** interpret the same frozen plans in a shadow Gleam/OTP
   backend under an equivalent crash matrix.

This sequence preserves the useful ambition without starting from a generalized
agent platform. Each rung must earn the next one.

## Kill conditions

Kill or narrow Braid if:

- it is only nicer syntax for `driver.md`;
- agents need to understand compiler implementation details to predict output;
- declared scopes are too inaccurate to improve batching;
- connectors cannot provide stable enough source identity to make locking useful;
- task-specific policy makes repository floors harder to audit;
- the target silently drops checked requirements;
- the compiler imports Ship's runtime model instead of producing a clean plan;
- a small Go or TypeScript preprocessor supplies the same safety and diagnostics
  with less operational friction.

The question is not whether this can become a sophisticated platform. It can.
The question is whether compiling one real delivery scope makes an agent's work
safer, more reproducible, and easier for the operator to trust than directly
writing the artifact Ship already understands.
