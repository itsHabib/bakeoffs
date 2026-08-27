# Bidirectional artifacts — frozen brief

## The bet

Generated agent-workflow artifacts need a lawful escape hatch: regenerate from
compact intent without erasing reviewed human overrides, and interpret allowed
edits back into the intent model without pretending every generated field is
editable. Haskell earns its seat if a bidirectional transformation makes the
round-trip laws executable and keeps refusal behavior local to the mapping.

## User and concrete workflow

An operator generating a Ship driver manifest or agent kickoff from a compact
project intent, then reviewing and changing model, concurrency, or escalation
policy in the generated artifact. The useful output is safe regeneration with
preserved overrides and explicit rejection of edits to derived fields.

## One job

Round-trip structured intent and an editable generated workflow artifact while
enforcing which fields are authoritative, derived, or legal overrides.

## Build

- An intent type and a rendered workflow-artifact type.
- A bidirectional mapping with `get` (generate) and `put` (reconcile edits).
- Explicit edit authority: derived identity/dependency fields are read-only;
  bounded execution-policy fields are editable.
- Structured conflict diagnostics.
- Executable GetPut, PutGet, and PutPut laws over a finite adversarial corpus.

## Required hard cases

- Regeneration preserves a legal human override.
- Editing a derived task identity is rejected, not silently overwritten.
- An upstream intent change plus an old override reconciles deterministically.
- Applying the same edit twice is stable.

## Haskell thesis

The product is the pair of directions plus their laws, not a generator and a
separate reverse parser. Haskell's first-class functions, immutable values, and
equational testing should make the bidirectional contract the implementation's
center. It loses if two ordinary conversion functions are equally clear.

## Demo

Generate a three-stream driver artifact, apply a legal concurrency/model
override, change the upstream project name, regenerate with the override
preserved, then attempt to edit a derived dependency and show the precise
refusal. Run the three round-trip laws.

## Kill condition

Kill if realistic artifacts are too lossy or presentation-sensitive for useful
round trips, if overrides inevitably become an untyped patch language, or if
the lens laws do not catch a failure a normal golden test would miss.

## Non-goals

No Markdown/YAML parser, natural-language intent inference, file watcher,
template registry, editor integration, or general-purpose lens library.
