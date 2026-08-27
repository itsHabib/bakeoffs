# Durable workflows

A closed, typed workflow value for dispatch → verify → record with deterministic
journal replay. A completed, committed activity is reused after a crash; a
workflow whose version, step identity, order, input, or encoded result no longer
matches is refused.

The implementation states the unavoidable boundary plainly: if an external
effect succeeds and the process dies before the journal commit, replay is
at-least-once and may duplicate that effect. The POC does not claim exactly-once
execution.

```powershell
stack test
stack run
```

There is no scheduler, worker service, database, timer system, or distributed
protocol here—only the replay semantic kernel.

## Why Haskell

`Workflow a b` and its GADT activities make stage compatibility part of the
program value: dispatch produces the exact input accepted by verify, which
produces the exact input accepted by record. The same closed value supports
signature inspection, journal validation, fresh execution, and replay. Adding
an activity forces every interpreter to account for its input, output, kind,
and codec.

The journal is still bytes-at-rest in miniature, so replay dynamically checks
version, shape, input, and decoding. The POC does not pretend serialization
preserves static types across process boundaries.

## Strongest alternative

An exhaustive Rust state machine can preserve much of the stage safety; a Go
engine can preserve the journal protocol with more runtime checks. Temporal is
the mature operational alternative. This project earns a separate existence
only as a much smaller agent-delivery replay kernel whose semantics are visible
in one typed program rather than dispersed across callbacks and worker code.
