# Validation record

> **One entry is withheld from this public archive.** It was built against
> day-job domain rules and named an employer product, so it is not mine to
> publish. Its scores, validation result and judging notes stay in place —
> a scorecard quietly edited down to five entries would misrepresent the
> round.

Validated on 2026-08-10 with Stack 3.11.1, resolver `lts-24.50`, and GHC 9.10.3.
All packages enable `-Wall -Werror`.

| entry | tests | demonstrated hard case |
|---|---|---|
| `assurance-dsl` | pass | previous-head evidence is reported stale |
| `delegation-dsl` | pass | missing and ambiguous artifact producers are rejected |
| `capability-dsl` | pass | insufficient grant produces zero effects |
| `recovery-dsl` | pass | deduplicated and at-least-once effects diverge honestly after the same crash |
| *withheld — see note* | pass | missing snapshot field blocks evaluation and appears in the manifest |
| `work-driver-dsl` | pass | overlapping parallel work serializes; ungated landing is rejected |

These are bounded executable properties, not proofs of real external systems.
