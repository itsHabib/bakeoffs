# provenance-datalog

A tiny fixed-point rule engine where each conclusion retains every alternative
derivation as an idempotent provenance polynomial. It answers both “is this
fact true?” and “which exact base facts and rules make it true?”

```powershell
stack run
stack test
```

## Result

The POC derives exact-head PR readiness through Gate and human-override paths,
derives transitive dependency impact recursively, and retracts only the stale
readiness facts when the PR head changes. No mutable explanation log exists;
the explanation is the value computed by the same rule evaluation.

## Why Haskell

The central type is an idempotent semiring of provenance monomials. Rule
conjunction uses multiplication, alternative derivations use addition, and the
fixed-point evaluator is generic over those operations. Immutable maps/sets and
recursive evaluation keep the semantics close to the algebra.

Rust or Go can implement this engine. The Haskell claim is narrower: the
algebraic laws and the evaluator's composition are the implementation structure,
not conventions spread through a mutable graph walker. The tests pin
associativity, distributivity, identities, recursive closure, and retraction.

## Strongest alternative

Soufflé/Datalog with a separate provenance extension—or Differential Datalog
for real incremental scale. This candidate should graduate only if its
provenance-bearing answers are valuable enough to justify a purpose-built,
typed embedded engine rather than adopting an existing database.
