# Assurance DSL — frozen brief

## Bet

Exact-head delivery policy should read like the evidence argument it represents,
then support evaluation, requirement discovery, and explanation from the same
syntax tree.

## Authoring target

```haskell
ready "merge" $
  exactHead
    .&&. checksPassed ["unit", "lint"]
    .&&. (gateReceipt .||. humanJudgment)
```

## One job

Decide whether a named action is ready for one immutable subject revision and
return a structured explanation of the decision.

## Hard case

Evidence for the previous revision must be reported as stale, not silently
counted as missing or accepted. Alternative authorization paths must preserve
which path succeeded.

## Haskell thesis

A closed proposition algebra supports independent evaluation, evidence-demand,
and explanation interpreters. It loses if a boolean expression over a map is
equally legible and informative.

## Non-goals

No policy server, identity system, cryptography, GitHub client, receipt format,
or merge implementation.
