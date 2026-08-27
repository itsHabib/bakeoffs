# hack-obligation

`hack-obligation` is a local, deterministic evidence-work frontier. It consumes
frozen verification contracts, immutable oracle `ChangeAssurance` snapshots,
and additive agent overlays. It does not inspect receipts, run agents, execute
effects, or grant authority.

Run the complete check and hands-free demo:

```sh
./demo.sh
```

The command verifies formatting, vets and tests the Go code, runs the
byte-stable lifecycle, prints the settled-is-terminal mutant beside production,
and emits the final `ObligationFrontierV1`, its digest, the common envelope, and
the nonblank weighted production source line count.

## The law

Mandatory evidence work is monotone under later information and agent overlays.
A late refutation can reopen a settled goal. An agent may add work, but cannot
delete or downgrade a mandatory contract obligation.

The reducer is pure over an append-only event sequence. Stable obligation IDs
bind kind and claim to task revision, contract, repository, base, head, diff,
and policy. Subject repair supersedes H1 obligations, opens `refresh_subject`,
then lets the H2 oracle name the remaining gaps. Transition histories explain
every open, discharged, reopened, and superseded state.

An empty frontier means only that mandatory evidence work is satisfied.
Every displayed oracle snapshot retains its original summary, coverage, gaps,
and `merge_authority: none`.

## Buyer test

The buyer hypothesis is narrow: teams already paying for multi-agent assurance
or incident review will attach a deterministic frontier if they have seen “all
workers returned” mistaken for “all mandatory evidence exists.” The smallest
insertion seam is after an existing assurance reducer: append its immutable
snapshot and read the resulting open obligations.

A credible shadow test is one authorized long-running review loop. Record
contracts and oracle snapshots without controlling scheduling or merge, then
compare frontier reopenings and subject refreshes with later CI failures,
review dispositions, fixes, and reverts. Measure missed mandatory follow-up and
operator overhead; do not report a confidence score.

## Strongest cheap alternative

The demo gives a named-node scheduler conditional branches, an event hook, and
an explicit reopen transition. It settles, replays deterministically, and opens
`repair-critical-finding-resolved` after the late refutation. It does not
produce the same native artifact: contract-derived identities, monotone overlay
enforcement, and typed transition explanations would need additional policy.

Kill the bet if a real scheduler already produces the same typed frontier and
explanations with less machinery. Also kill it if a manager or model must
interpret evidence, or if the frontier ever accepts raw receipts.

## Scope

The implementation is Go 1.26 standard library only, offline, keyless, and
self-contained. `fixtures/exact-head-lifecycle-v1.json` is a byte-for-byte copy of
the frozen bakeoff deck. Golden files pin the final replayed frontier and the
normalized demo JSON.
