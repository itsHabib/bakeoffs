# Haskell-shaped project bakeoff — five semantic kernels

Topic: **find one substantial agentic-development project whose central job is
materially better expressed in Haskell, then promote only the winner.**

This is not a language-tour wave. An entry loses if Haskell is merely a pleasant
implementation choice around a generic service. Each entry must expose a
semantic core—types, laws, interpreters, fixed points, or replay semantics—that
would become less direct, less checkable, or less compositional in the obvious
Go/Rust/TypeScript implementation.

## Entries

| # | slug | one job |
|---|---|---|
| 1 | `protocol-compiler` | project a global multi-role protocol into local obligations and reject incompatible conversations |
| 2 | `provenance-datalog` | derive engineering facts while retaining an exact, recomputable explanation of why each fact is true |
| 3 | `bidirectional-artifacts` | round-trip compact intent and editable generated artifacts without silently losing human overrides |
| 4 | `capability-plans` | analyze a plan's complete effect envelope before allowing the same plan value to execute |
| 5 | `durable-workflows` | replay a serializable workflow after interruption without duplicating completed effects |

## House rules

- **One responsibility, done extraordinarily well.** Each entry chooses one
  semantic job and carries it through happy path, refusal path, adversarial
  case, and laws. Small means cohesive, not shallow.
- **The semantic kernel is the demo.** No web UI, daemon, plugin system,
  marketplace, deployment layer, generic adapter framework, or persistence
  abstraction unless that mechanism is inseparable from the stated job.
- **Haskell must earn its seat.** Every README names the strongest conventional
  implementation and identifies the specific guarantee or clarity lost in that
  port. "ADTs are nice" is not enough.
- **Useful, not scholastic.** Each entry uses one artifact or workflow shape
  recognizable from the operator's agentic-development portfolio. No arithmetic
  expression evaluators, toy bank accounts, or placeholder domains.
- **Same finish line.** Every entry ships:
  1. `BRIEF.md`—the frozen bet, scope, and kill condition.
  2. `README.md`—the result, Haskell thesis, and one command to run.
  3. `DEMO.md`—a deterministic walkthrough under two minutes.
  4. A runnable semantic kernel and dependency-light tests.
  5. At least one law/property and one adversarial fixture.
- **No false proof language.** Tests establish the declared bounded property;
  they do not prove a system correct. Static guarantees must say exactly what
  the type checker rules out and what remains dynamic.
- **No shared candidate code.** Repetition is allowed. A prematurely extracted
  framework hides which candidate actually has the cleanest natural core.
- **Dependency restraint.** Prefer `base`, `containers`, and other GHC-bundled
  packages. Add a library only when it is part of the thesis, not to save ten
  lines.
- **Stop when the bet is answered.** A convincing negative result is valid. Do
  not rescue a weak thesis by adding product surface.

## Common evaluation protocol

For each entry:

1. Run its deterministic demo from a clean checkout.
2. Run its test command.
3. Inspect the semantic kernel, not total repository polish.
4. Attempt the adversarial case documented in `DEMO.md`.
5. Score it against the rubric below.
6. Record the strongest reason to promote and the strongest reason to kill.

## Judging — 100 points

- **30 — Haskell necessity.** Does the project genuinely exploit a Haskell
  semantic advantage—typed EDSLs, algebraic interpretation, laziness/fixed
  points, lawful bidirectionality, or replayable program values? Would the
  obvious non-Haskell version lose a meaningful property rather than syntax?
- **25 — real agentic utility.** Is there a named user, concrete workflow, and
  useful output today? Can it attach to an existing portfolio boundary without
  first creating a platform?
- **20 — depth of responsibility.** Does its one job cover the hard semantic
  cases and refusals extraordinarily well? Laws, diagnostics, and adversarial
  behavior count; breadth does not.
- **15 — falsifiable demonstration.** Does the demo make the bet obvious, and
  can a specific result kill the project rather than merely suggest more work?
- **10 — restraint and coherence.** Is every mechanism load-bearing for the
  thesis? Thin-but-useless loses here alongside overbuilt.

Tie-breaker: **which candidate's core loses the most if rewritten idiomatically
in Go or Rust?**

## Promotion rule

Only the highest-scoring candidate may graduate into its own repository. A
winner below **75/100**, or below **20/30 on Haskell necessity**, does not
graduate; the honest outcome is that this wave did not find the project.
