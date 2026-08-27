# Branchroom

Branchroom turns an agent rerun into a controlled causal fork over state the
harness can actually declare. It hashes a common prefix into an `AgentCapsuleV1`,
forks deterministic child epochs, and uses one pure reducer to accept each
child's correlated terminal while refusing a delayed terminal from the parent
epoch.

Run the entire entry:

```sh
make demo
```

That command verifies the frozen fixture bytes, formatting, vet, and tests,
then prints the production/mutant trace, `ExperimentForkV1`, its digest, the
normalized bakeoff envelope, and weighted source line count. It is offline,
keyless, model-free, and has no dependencies outside Go's standard library.

## What the artifact proves

The capsule commits to the frozen input bundle and verification contract,
repository tree, instructions, tool schema, harness, environment, and six-event
declared prefix. Both branches retain that exact capsule and prefix digest.
Their child epochs are allocated from sorted labels, so creation order cannot
change identity. The terminal reducer checks epoch, branch, next sequence,
event ancestry, and pending call correlation before accepting.

The exported API has no retain-parent switch: `ForkBranches` always creates a
fresh epoch, and `StartBranch` rejects a forged non-fresh child. The planted
mutant is package-private demo code and cannot be selected by a library caller.

The `retain-parent-epoch` mutant changes only epoch allocation. Given the same
late terminal bytes, it accepts the result; production refuses it with
`parent_epoch_terminal` at event 7. This is a controlled-run artifact suitable
for capture by a trusted runner. It is not an evidence receipt, assurance
outcome, approval, or authority.

## Buyer test

Buyer hypothesis: teams already paying engineers to reproduce intermittent
failures in long-running coding-agent and evaluation runs will adopt a small
capture boundary that makes two reruns comparable and identifies the first
declared divergence.

The smallest insertion seam is immediately before a trusted harness executes
one tool call: record the canonical prefix and pending call, fork fresh epochs,
then wrap terminal delivery with the reducer. No scheduler, provider change,
database, or agent-platform migration is required.

A credible shadow field test is to instrument one repeatedly failing agent
verification loop without changing its decisions. For each retry, capture a
capsule and two branches, compare Branchroom's first declared divergence with
the engineer's eventual diagnosis, and record capture overhead and cases where
undeclared or hidden state prevents attribution. Promotion fails if plain event
replay plus a branch ID produces the same artifact with less mechanism, if call
correlation alone prevents the observed stale-terminal failures, or if useful
attribution requires hidden model state.

## Why a worktree and copied log are weaker

`git worktree add`, a copied prompt/JSONL log, and a `branch_id` separate files,
but do not prove which execution-relevant prefix bytes are shared or give a
reused correlated call a new causal generation. Adding a canonical prefix
digest, a fresh deterministic epoch, parent-event continuity, and fail-closed
terminal reduction closes those gaps; at that point the cheap alternative has
implemented Branchroom's one law.

## Scope

Branchroom captures declared harness-visible state only. It does not snapshot
model activations or processes, judge evidence, recompute the fixture's oracle,
schedule work, rank branches, authorize effects, or claim merge authority.
