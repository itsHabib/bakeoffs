# Recovery DSL

A typed activity language whose retry declarations change replay behavior.
Deduplicated activities reuse an external key after an effect-before-commit
crash; idempotent activities may safely repeat; at-least-once activities repeat
and report that fact. Compatible journal entries are never executed again.

```powershell
stack test
stack run
```

An exhaustive Rust state machine is the strongest alternative. Haskell makes
stage composition and per-activity codecs compact, but this candidate loses if
real persistence reduces the language to stringly runtime validation or if it
merely rebuilds a workflow engine.
