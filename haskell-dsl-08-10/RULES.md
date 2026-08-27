# Haskell DSL bakeoff — executable languages for real work

Topic: **find a small language that makes an agentic or day-job workflow more
pleasant to author and materially safer to run.** The winning project must be a
useful language, not a Haskell technique wearing a domain costume.

## Entries

> **One entry is withheld from this public archive.** It was built against
> day-job domain rules and named an employer product, so it is not mine to
> publish. Its scores, validation result and judging notes stay in place —
> a scorecard quietly edited down to five entries would misrepresent the
> round.

| # | slug | one job |
|---|---|---|
| 1 | `assurance-dsl` | define exact-head evidence policies and explain readiness decisions |
| 2 | `delegation-dsl` | define typed agent handoffs and compile role-local briefs |
| 3 | `capability-dsl` | author executable plans whose complete authority envelope is inspectable first |
| 4 | `recovery-dsl` | author durable activities with explicit retry semantics and crash behavior |
| 5 | *withheld — see note* | define day-job domain rules once while exposing executable logic and snapshot inputs |
| 6 | `work-driver-dsl` | compile project intent into dependency-correct, conflict-free execution batches |

## House rules

- **Authoring is part of the product.** Each entry starts with a readable DSL
  example that a real user could plausibly maintain.
- **At least two load-bearing interpretations.** Rendering and pretty-printing
  do not count by themselves. Examples: analyze + execute, evaluate + enumerate
  requirements, validate + compile.
- **One responsibility.** No scheduler, daemon, agent runtime, UI, generic
  plugin framework, policy platform, or storage abstraction.
- **Haskell must change the design.** A candidate loses if a JSON schema and a
  loop provide the same guarantees and authoring experience.
- **A real hard case is mandatory.** Every demo must show one refusal or
  counterexample that ordinary happy-path configuration tends to miss.
- **No false proof language.** Types rule out only what the implementation
  actually encodes. Runtime data, persistence, and external effects remain
  dynamic unless demonstrated otherwise.
- **No shared candidate code.** Repetition is allowed so abstractions cannot
  hide which language is naturally smallest.
- **Dependency restraint.** Prefer `base` and `containers`; add packages only
  when they are part of the DSL thesis.
- **Stop when the language bet is answered.** Do not add adapters or product
  surface to rescue weak syntax or semantics.

## Required experiment for every entry

1. Show the authored DSL value.
2. Run its primary interpreter.
3. Run its independent analysis/compiler interpreter.
4. Exercise the frozen adversarial case.
5. Run laws or structural properties appropriate to the language.
6. Compare the same job to the strongest ordinary Rust/Go/TypeScript shape.

## Judging — 100 points

- **25 — Haskell necessity.** Do typed syntax, algebraic composition, or
  interpretation materially improve the language?
- **20 — authoring joy.** Is the common case concise, legible, and difficult to
  misuse without becoming clever punctuation?
- **20 — real utility.** Does the language attach to a workflow used today?
- **15 — semantic leverage.** Do multiple interpreters or static analyses
  produce genuinely different useful outputs from one definition?
- **10 — falsifiable hard cases.** Does the demo expose a specific reason to
  promote or kill the candidate?
- **10 — restraint and fit.** Is every mechanism load-bearing, and can the DSL
  integrate without becoming a platform?

Graduation requires **78/100**, at least **18/25 Haskell necessity**, and at
least **14/20 authoring joy**.

Tie-breaker: which language would the operator most want to author by hand next
week while still trusting its compiled result?
