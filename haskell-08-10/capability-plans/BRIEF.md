# Capability plans — frozen brief

## The bet

An automation plan should be analyzable before it is executable. The exact same
plan value should support an interpreter that computes its complete capability,
resource, and mutation envelope and another interpreter that performs the
effects only after a grant covers that envelope. Haskell earns its seat if a
free applicative/selective-style plan algebra preserves analyzability without
turning execution into a second hand-maintained program.

## User and concrete workflow

An agent proposing repository work: read files, run checks, post a PR comment,
or request a merge. The operator or Gate-like boundary needs to know the full
effect envelope before authorizing execution. The useful output is a precise
preflight plus fail-closed execution under a scoped grant.

## One job

Represent a statically analyzable plan once, compute its complete effect
envelope, and execute it only when a grant covers every required capability and
resource.

## Build

- A typed operation GADT for a narrow repository workflow.
- A free applicative plan value supporting independent result composition.
- An analysis interpreter producing capabilities, resources, mutation class,
  and a bounded cost count without running effects.
- An execution interpreter against a deterministic in-memory fixture.
- A grant check with explicit missing-capability diagnostics.

## Required hard cases

- Analysis cannot execute or inspect runtime results.
- Independent operations retain their complete combined envelope.
- Insufficient grants fail before the first effect.
- Execution and dry-run render the same ordered operation set.

## Haskell thesis

Applicative structure is the product constraint: it permits static analysis of
the whole plan, while unrestricted monadic dependence would hide future effects
behind runtime values. Multiple interpreters consume one typed syntax. The
candidate loses if an ordinary list of command structs provides the same useful
guarantee with less ceremony.

## Demo

Preflight a read/test/comment plan, reject it under a read-only grant before any
fixture mutation, authorize it under the exact grant, then show that a plan
containing merge requires a distinct capability and mutation class.

## Kill condition

Kill if useful agent plans immediately require unrestricted dynamic branching,
if resource scopes cannot be known until execution, or if the free structure
adds no practical guarantee over a command list plus validation pass.

## Non-goals

No real GitHub calls, credential broker, policy language, network proxy,
long-running workflow engine, LLM planner, or generalized effect library.
