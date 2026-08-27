# Recovery DSL — frozen brief

## Bet

Durable agent activities should declare retry meaning beside their typed stage,
so crash simulation and replay can distinguish idempotent, deduplicated, and
honestly at-least-once effects.

## Authoring target

```haskell
activity "dispatch" Dispatch deduplicated
  >>> activity "verify" Verify idempotent
  >>> activity "record" Record idempotent
  >>> Done
```

## One job

Interpret a versioned activity program across crash boundaries without
misrepresenting whether an uncommitted external effect may repeat.

## Hard case

Crash after effect but before journal commit: a deduplicated dispatch reuses its
external key, while an at-least-once notification visibly duplicates. Changed
step identity must refuse replay.

## Haskell thesis

Indexed stages and a closed retry algebra make compatibility and replay
semantics part of the program. It loses if policies are decorative labels over
an ordinary state machine.

## Non-goals

No scheduler, worker fleet, database, timers, arbitrary DAG, Temporal clone,
or agent-provider integration.
