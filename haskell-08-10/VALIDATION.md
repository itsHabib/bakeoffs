# Validation record

Validated on 2026-08-10 with Stack 3.11.1, resolver `lts-24.50`, and GHC 9.10.3.
Every package enables `-Wall -Werror`.

| entry | `stack test` | `stack run` hard case |
|---|---|---|
| `protocol-compiler` | pass | unobserved branch is rejected at the Gate role |
| `provenance-datalog` | pass | head movement retracts stale readiness but not unrelated impact |
| `bidirectional-artifacts` | pass | legal override survives rename; derived dependency edit is refused |
| `capability-plans` | pass | read-only grant causes zero effects; merge needs distinct authority |
| `durable-workflows` | pass | committed dispatch is reused after crash; changed step ID is refused |

The tests are bounded executable evidence, not formal proofs. Each candidate's
`DEMO.md` states the deterministic reproduction command and expected refusal.
