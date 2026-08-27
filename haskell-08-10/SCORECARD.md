# Scorecard — haskell-08-10

Rubric from [RULES.md](RULES.md). Score only after running each `DEMO.md` and
its tests.

| entry | Haskell /30 | Utility /25 | Depth /20 | Demo /15 | Restraint /10 | **Total /100** | promote case | kill case |
|---|---:|---:|---:|---:|---:|---:|---|---|
| protocol-compiler | 23 | 19 | 17 | 14 | 9 | **82** | projection makes branch knowledge an explicit compiler check | Rust enums retain most exhaustiveness; runtime adoption needs a protocol boundary |
| provenance-datalog | 22 | 23 | 19 | 14 | 9 | **87** | provenance is computed with the fact, so exact explanations and retraction cannot drift | established Datalog engines are stronger and the algebra ports cleanly |
| bidirectional-artifacts | 17 | 21 | 17 | 14 | 10 | **79** | lawful regeneration solves a real convention-plus-escape-hatch problem | two Rust conversion functions plus property tests retain the demonstrated guarantee |
| capability-plans | 26 | 23 | 18 | 15 | 9 | **91** | a typed result-bearing plan exposes its entire effect envelope before execution | applicative shape excludes result-dependent agent work; plain command batches are competitive |
| durable-workflows | 27 | 24 | 19 | 15 | 9 | **94** | typed stages plus one replayable program value make crash semantics unusually explicit | risks rebuilding Ship/Temporal; serialized results still require dynamic checks |

Minimum graduation bar: **75 total and 20 Haskell-necessity points**.

Tie-breaker: which semantic core loses the most when ported idiomatically to Go
or Rust?

## Verdict

- Winner: **durable-workflows — 94/100**
- Promote to: a standalone Haskell repository for a narrowly scoped
  agent-delivery replay kernel (working name: **reprise**).
- Why: it has the strongest combination of immediate portfolio relevance and
  Haskell-shaped semantics. The indexed workflow rules out invalid stage
  composition; the closed program supports signature, validation, execution,
  and replay; and the demo distinguishes the two crash windows without an
  exactly-once fairy tale. The production responsibility remains one sentence:
  **never re-run a compatibly journaled activity, and refuse ambiguous replay.**
- Runner-up: **capability-plans — 91/100**
- Why it lost: its static preflight is the most intrinsically applicative idea
  in the wave, but that is also its product constraint. General agent work is
  result-dependent; once the workflow becomes monadic, the complete-envelope
  promise disappears. It remains an excellent bounded subsystem for approval
  packets, not the best flagship project.
- Wave-level negative result: **bidirectional-artifacts fails the 20/30 Haskell
  floor despite scoring 79 overall.** The laws are useful, but the POC confirmed
  its own kill condition: ordinary conversion functions are equally clear.

## Ranking and decision notes

1. `durable-workflows` — graduate the semantic bet, not this archived folder.
2. `capability-plans` — retain as the best alternate; consider folding its
   preflight algebra into the winner only after a real workflow demands it.
3. `provenance-datalog` — valuable reference model; prefer an existing Datalog
   implementation unless exact provenance becomes a proven portfolio need.
4. `protocol-compiler` — compelling compiler demo, but lacks an immediate
   adoption boundary and Rust narrows the language advantage.
5. `bidirectional-artifacts` — useful problem, rejected as a Haskell flagship.

No mechanisms are shared between the winner and runner-up at promotion time.
Combining capability authorization, provenance, protocols, reconciliation, and
replay now would recreate the overengineered platform this wave was designed to
avoid.
