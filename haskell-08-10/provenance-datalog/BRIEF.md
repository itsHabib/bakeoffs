# Provenance Datalog — frozen brief

## The bet

Engineering facts should carry a recomputable explanation of why they are true.
A compact rule engine whose annotations form a provenance semiring can answer
both a query and its alternative derivations, then retract conclusions when an
input becomes stale. Haskell earns its seat if the algebra of explanations and
the fixed-point evaluator stay small, lawful, and extensible without coupling
query logic to presentation.

## User and concrete workflow

An agent deciding whether a PR is ready or what depends on a changed contract.
Today the answer is reconstructed from Gate verdicts, checks, receipts, docs,
and repository relationships. The useful output is a fact plus the exact base
facts and rule paths supporting it.

## One job

Compute a monotone fixed point of engineering rules while preserving all
alternative provenance for each derived fact.

## Build

- Typed facts for exact-head evidence and dependency relationships.
- Rules with conjunctive premises and one conclusion.
- A fixed-point evaluator over a provenance expression algebra.
- Normalization sufficient to make duplicate derivations deterministic.
- Retraction by recomputation from a changed base-fact set.
- Human-readable explanations for one readiness query and one impact query.

## Required hard cases

- Conjunction records that all premises were required.
- Alternative derivations remain alternatives rather than being collapsed.
- Recursive rules terminate at a stable fixed point.
- Removing a stale exact-head fact removes only conclusions that depended on it.

## Haskell thesis

The core is a generic algebra: rule evaluation combines evidence with product
and alternatives with sum, independent of rendering. Laziness/recursive data
and typeclass-like interpretation should make the fixed-point and explanation
semantics clearer than a conventional mutable graph walker. It loses if this is
just a map of strings with pretty labels.

## Demo

Derive `Ready(pr-42, sha-a)` through two legitimate paths, print both
explanations, retract the Gate fact after the head moves to `sha-b`, and show
that the stale readiness disappears while unrelated impact facts survive.

## Kill condition

Kill if provenance annotations add little beyond logging rule names, recursive
queries require a general database before becoming useful, or the explanation
algebra cannot express a real portfolio question cleanly.

## Non-goals

No durable database, parser, service catalog crawler, UI, embeddings, natural
language query layer, or incremental database optimizer.
