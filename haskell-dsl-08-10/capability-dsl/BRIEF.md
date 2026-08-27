# Capability DSL — frozen brief

## Bet

A useful bounded agent plan can return typed results while retaining its entire
operation tree, allowing exact permission and mutation preflight before the
first effect.

## Authoring target

```haskell
(,,) <$> readRepo "src/Gate.hs"
     <*> runCheck "unit"
     <*> commentOn 42 "ready"
```

## One job

Analyze and authorize the complete effect envelope of the same plan value that
will execute.

## Hard case

A grant missing only comment authority must refuse the whole plan with zero
effects. Adding merge must require a distinct capability and mutation tier.

## Haskell thesis

The free-applicative shape combines heterogeneous typed results without hiding
future operations. It loses if real plans immediately require monadic branching
or an ordinary command array offers the same useful contract.

## Non-goals

No credential broker, tool proxy, policy language, network calls, dynamic agent
loop, or generalized effects library.
