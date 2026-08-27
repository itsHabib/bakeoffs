# Work-driver DSL — frozen brief

## Bet

Project intent should be pleasant to author once and compile deterministically
into dependency-correct, file-conflict-safe execution batches plus an
explanation of what can run next.

## Authoring target

```haskell
project "review-kernel" $ do
  spec <- task "spec" `touches` ["docs/**"]
  (impl, fixtures) <- parallel
    (task "implement" `after` spec `touches` ["src/**"])
    (task "fixtures"  `after` spec `touches` ["test/**"])
  green <- validate "local-green" `afterAll` [impl, fixtures]
  land "merge" `after` green
```

## One job

Compile a work graph into the maximum deterministic parallel batches allowed by
dependencies and declared file scopes.

## Hard case

Nominally parallel tasks whose scopes overlap must be serialized with an
explanation. Opaque in-plan references must make missing dependencies and
forward-reference cycles unrepresentable; a landing path without a validation
gate must still be rejected dynamically.

## Haskell thesis

Sequential/parallel composition, graph analysis, and batch compilation operate
over one authored algebra. It loses if the do-notation lies about concurrency or
a plain task manifest plus topological sort is clearer.

## Non-goals

No agent dispatch, polling, retries, Git worktrees, Jira/Dossier writes, Ship
replacement, dashboard, or general project-management system.
