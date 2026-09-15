# Restartable work bakeoff

Three independent builds answer one question: after a worker dies, what may its
replacement safely do?

| Entry | Approach | Result |
|---|---|---|
| [Plain Go](plain-go) | Explicit state machine, journal, and effect functions | Adequate baseline; all tested behaviors pass |
| [Warrant](warrant-entry) | Evidence-aware reducer plus explicit recovery adapter | Useful reduction, but substantial recovery machinery remains outside it |
| [Parley](parley) | Protocol compiler and observer beside an ordinary executor | Adds a useful ordering diagnostic; does not establish effect correctness |

All three detected the required duplicate-effect, stale-incarnation, and
stale-evidence mutants. All three also reached the same limit: when an opaque
effect may have committed but exposes no query or receipt, the replacement must
leave the operation unresolved.

Read [the execution record](RESULTS.md) for the evidence and [the original
rules](RULES.md) for the common workload. Operator scoring remains blank.

The Warrant entry preserves its original pinned-local-dependency design. The
historical Warrant source had no redistribution license, so it is not copied
here and that entry requires the exact external source checkout described in
its README.
