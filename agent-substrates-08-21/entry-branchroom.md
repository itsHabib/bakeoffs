# Entry 1: hack-branchroom — fork a declared causal prefix, not hidden state

Read `agent-substrates-08-21/README.md` first. Its house
rules, assurance boundary, frozen input deck, and judging apply verbatim.

Repo: `agent-substrates-08-21/branchroom`. You never see the other three entries.

## The bet

Rerunning an agent from a prompt is not a replay: effective context, tool
observations, environment, and causal position may already have drifted. The
bet is that a small semantic capsule can prove the declared state two runs
share, fork it into fresh causal epochs, and identify the first meaningful
divergence. This is not a snapshot of model activations or a running process.

Buyer hypothesis to validate before promotion: teams evaluating or debugging
long-running agents will pay to turn irreproducible “try it again” failures into
fair, attributable counterfactual runs.

## One law

After a fork, a child with a fresh epoch accepts its own correlated terminal
result and rejects a late terminal from the parent epoch. Both children retain
the exact same content-addressed declared prefix.

## What to build

Build one local CLI/library with three small values:

- `AgentCapsuleV1`: frozen input-bundle digest, verification-contract digest,
  repository tree, instructions, tool schema, harness, environment, terminal
  prefix sequence, and prefix digest;
- `BranchV1`: capsule digest, branch label, parent epoch, and deterministic
  fresh child epoch;
- `RunEventV1`: branch, epoch, sequence, call id, event kind, payload digest,
  and parent event digest.

Implement canonical serialization, content-derived ids, and one pure reducer.
The reducer verifies prefix continuity and accepts a tool terminal only when
branch, epoch, next sequence, and pending call id all match.

From the frozen input deck:

1. Record a common prefix immediately before the trusted fixture runner would
   execute one verification method.
2. Publish its capsule and fork `control` and `counterfactual` with fresh
   deterministic epochs.
3. Change exactly one declared input in `counterfactual`—the fixture's planted
   negative-control mode—and accept each child's own terminal.
4. Deliver one delayed parent-epoch terminal to both children and reject it at
   the same golden step.
5. Emit `ExperimentForkV1`: capsule/prefix identity, declared perturbation,
   branch event digests, first divergence, and rejected terminal.

The experiment artifact is suitable for later capture by a trusted runner. It
is not an `EvidenceReceipt`, an assurance outcome, or authority.

## Single-law mutant

Implement `retain-parent-epoch`: forked children keep the parent's epoch. Run it
over byte-identical input. It must accept the delayed parent terminal; the
production reducer must reject the same event with `parent_epoch_terminal`.
Do not disable any other check.

## Strongest cheap alternative

`git worktree add`, copy the prompt and JSONL log, add a `branch_id`, and rerun.
Branchroom earns substrate-necessity points only if `ExperimentForkV1` proves
common-prefix identity and causal separation that this recipe does not.

Kill the bet if ordinary event replay plus a branch id produces the same
artifact, if a fresh epoch adds no safety beyond a correlated call id, or if
the useful claim requires hidden model state.

## Required tests

- canonical capsule bytes and ids are stable;
- branch creation order does not affect child ids;
- both branches share the exact prefix digest;
- each child accepts its own next terminal;
- both reject the late parent terminal at the same step;
- changed instructions, tool schema, harness, or environment change the
  capsule digest;
- duplicate/gapped sequences fail closed;
- normalized demo JSON matches a checked-in golden file.

## What NOT to build

- No approval, grant, mandate, evidence assessment, assurance reducer, or
  action authorization.
- No Firecracker, VM/CRIU snapshot, provider API, live model, or hidden-state
  claim.
- No workflow scheduler, event-sourcing framework, agent registry, database,
  distributed log, compaction, UI, or branch ranking/merging.

## Canned demo

One command runs tests and the byte-identical mutant/production trace:

```text
capsule sha256:...  common-prefix sha256:...  terminal-seq 6
control epoch 13        own terminal: ACCEPT
counterfactual epoch 14 own terminal: ACCEPT
mutant epoch 12         late parent terminal: ACCEPT  <-- planted bug
branchroom              late parent terminal: REFUSE parent_epoch_terminal
first divergence: event 7
```

Then print `ExperimentForkV1`, its digest, the common normalized envelope, and
the disclosed source line count.

## The 60-second demo story

“These runs share a proven declared history, not merely a similar prompt. I
fork once and perturb one input. The naive fork retains the parent epoch, so a
late result is accepted into a child and corrupts the comparison. Branchroom
gives each child a fresh causal epoch, accepts its own result, rejects the
parent result, and points to the first divergence. It captures only state the
harness can know—and it makes that limit explicit.”

