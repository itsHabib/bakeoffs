# Practical systems bakeoff — language leverage without an infrastructure hobby

Topic: **build small tools that remain useful when placed beside real production
systems.** The language must contribute something material; the project must not
need a new service fleet, scheduler, database, or control plane to justify itself.

## Entries

| # | slug | language | one job |
|---|---|---|---|
| 1 | `mcp-contract-lab` | Gleam | compile MCP contracts and recorded exchanges into compatibility checks |
| 2 | `streaming-fixture-lab` | Gleam | capture, normalize, redact, and replay large streaming fixtures |
| 3 | `flow-state-lab` | Quint | find unsafe CAM-to-factory workflow interleavings and export regression traces |
| 4 | `durable-pipeline-kernel` | Gleam | resume evidence-gated shipping and maintenance loops from one append-only journal |

## House rules

- **Attach to work that exists.** Every entry names the real integration seam it
  improves and produces an artifact an existing test suite or review can consume.
- **No infrastructure alibi.** A CLI, library, model, or fixture pack is enough.
  No daemon, queue, database, hosted dashboard, or deployment is allowed.
- **The language is load-bearing.** Remove Gleam's typed result/data model or
  Quint's exhaustive state exploration and the project must become meaningfully
  worse, not merely differently spelled.
- **One command to judge.** A clean checkout must expose a single local command
  that runs the important behavior and its tests.
- **Failure is a product.** Each entry freezes at least one adversarial case as a
  deterministic artifact, rather than only demonstrating a happy path.
- **Evidence is a replay recipe, not a screenshot.** A result carries the exact
  subject, versioned tool identity, inputs, seeds, bounds, and commands needed
  for another clean checkout to reproduce it. If an external dependency prevents
  full reproduction, the artifact names that limitation and still provides a
  locally replayable semantic fixture.
- **Production claims stay honest.** A model proves the model; a fixture tests the
  adapter. Neither is described as proving a remote production implementation.
- **Employer-specific material stays out.** Public examples use generic names and
  synthetic identifiers. No internal URLs, tickets, people, or copied source.
- **Stop at the seam.** Do not build a generic platform to make a narrow result
  appear larger.

## Required experiment for every entry

1. Show the smallest useful authored input.
2. Run the primary analysis or transformation.
3. Produce an artifact consumable outside the project.
4. Reproduce one frozen hard case.
5. Replay the produced artifact in a fresh process and compare its semantic result.
6. Run the full local check from a clean checkout.
7. State exactly what the result does and does not establish.

## Judging — 100 points

- **30 — real utility.** Would this remove toil or prevent a plausible failure in
  a workflow used today?
- **20 — production realism.** Does it handle retries, partial data, malformed
  input, scale, or integration drift instead of assuming the demo is the world?
- **20 — language leverage.** Does the selected language materially improve the
  guarantees or authoring model?
- **15 — artifact portability.** Can an ordinary existing test suite consume the
  output without adopting the whole project?
- **10 — falsifiability.** Is there a crisp promote/kill result and a reproduced
  adversarial case?
- **5 — restraint.** Is the project the smallest useful thing at this seam?

Graduation requires **78/100**, at least **22/30 real utility**, and at least
**14/20 language leverage**.

Tie-breaker: which entry would still be used after the novelty of its language
wore off?
