# Demo — provenance Datalog

Run:

```powershell
stack run
stack test
```

The initial query shows `ready(pr-42, sha-a)` with two complete derivations:
Gate + checks, or human override + checks. A recursive dependency rule also
shows why `workbench/contracts` impacts `agent-console` through Ship.

The fixture then moves the PR to `sha-b` and removes the old head's evidence.
Readiness for `sha-a` disappears, `sha-b` remains unready because checks have
not run, and the unrelated dependency explanation survives.

The tests exercise the provenance semiring laws, recursive fixed-point,
selective retraction, and refusal of a rule whose conclusion contains an
unbound variable.
