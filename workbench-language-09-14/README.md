# Workbench language bakeoff

Three independent proofs of concept tested ways to express the same plan/apply
workflow: describe intended resources and tasks, inspect a diff, apply it, and
recover after interruption.

| Entry | Bet | Status |
|---|---|---|
| [HCL](hcl) | Familiar configuration syntax earns its parser and dependency cost. | reference |
| [Custom language](custom-language) | A tiny `keep`/`run` syntax makes lifecycle clearer. | rejected |
| [Typed Go API](typed-go) | Ordinary Go supplies composition without a new syntax. | reference |

Each entry includes its original brief, runnable demo, deterministic tests,
examples, and direct-tools baseline. Start with [the scorecard](SCORECARD.md)
for measurements and [the utility review](REVIEW.md) for the decision.

## Decision

All three implementations passed the common six-case workload. Most of the
value came from normalized intent, a readable plan, stale-plan refusal, and
retained per-action evidence. It did not come from the choice of syntax.

Direct tools remain the default for one bounded operation. Workbench did not
adopt any entry as a supported interface. The follow-up uses Terraform to test
a smaller declaration of Fleet cards, seats, and compute before Workbench owns
a language or provider.

## Provenance

The entries were copied from their independent build repositories at these
exact revisions:

| Entry | Source revision |
|---|---|
| HCL | `itsHabib/hack-workbench-hcl@ebf32d245a1d885bccfbbf4d0c31610e30100458` |
| Custom language | `itsHabib/hack-workbench-language@dc1c79332128e64f7efdd72db0226a172af7d9c3` |
| Typed Go API | `itsHabib/hack-workbench-typed@6b757a8e0c81c1c2456874c7256a281221d99327` |
