# Delegation DSL — frozen brief

## Bet

Multi-agent work should declare artifacts and handoffs once, then compile into
role-local briefs and a validated dependency graph. Missing producers and
branch-dependent handoffs should fail before prompts are dispatched.

## Authoring target

```haskell
researcher `produces` Findings
implementer `requires` Findings `andProduces` Patch
reviewer    `requires` Patch    `andProduces` Verdict
```

## One job

Compile an authored delegation contract into one brief per role while proving
that every required artifact has a reachable, unambiguous producer.

## Hard case

Two agents produce the same singleton artifact, or a reviewer requires an
artifact no reachable role produces. Both must be precise compile errors.

## Haskell thesis

Artifacts and roles form a compositional contract algebra with graph and
role-projection interpreters. It loses if it becomes a verbose dependency list
with prettier constructors.

## Non-goals

No agent spawning, prompt optimization, message bus, runtime supervision,
provider adapters, or distributed protocol proof.
