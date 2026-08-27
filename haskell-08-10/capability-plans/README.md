# Capability plans

A typed, free-applicative plan for bounded repository automation. The same plan
value is interpreted once for a complete capability/resource/mutation envelope
and once for execution against a deterministic fixture. Authorization covers
the whole envelope before the first operation runs.

The applicative restriction is intentional: later operations cannot be hidden
behind runtime results. That makes the plan preflightable, at the cost of ruling
out unrestricted dynamic workflows.

```powershell
stack test
stack run
```

This POC ends at the semantic kernel. There is no credential service, policy
DSL, GitHub client, or daemon.

## Why Haskell

The GADT gives every operation its result type, while the free-applicative
syntax retains the complete operation tree. One plan can therefore return a
typed result tuple, compute its entire authority envelope without effects, and
execute only after authorization. Replacing `Applicative` with unrestricted
`Monad` would make later operations depend on runtime values and destroy that
specific guarantee.

## Strongest alternative

A Rust enum command list followed by validation is simpler and equally good
when callers only need an untyped batch. It loses the typed composition of
heterogeneous results, but that advantage matters only for workflows whose
shape can genuinely remain applicative. This candidate should die if useful
agent plans immediately need result-dependent branching.
