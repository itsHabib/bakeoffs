# Entry 2: Warrant — make the reducer earn its place

Read [`../RULES.md`](../RULES.md) first. All cases, effect contracts, invariants,
and finish-line requirements apply.

This entry was built in a fresh repository without inspecting its competitors.

## The bet

Warrant's subject-bound evidence and recovery decisions can remove consumer bookkeeping while preserving understandable execution. The comparison must include the missing journal durability and restart boundary, not just the existing pure reducer.

## Existing source

The entry used the Warrant README, runner, reducer, and tests at commit
`33094f0`. The inspected runner starts from a fresh journal and accepts an
`io.Writer`; neither behavior constitutes a complete durable resume implementation.

Vendor the minimal required source with its license into the fresh repo, or use an explicitly pinned local dependency with reproducible instructions. Any experimental changes belong in the entry. Do not edit or publish the canonical project.

## What to build

Express the common loop with Warrant's definition and evidence vocabulary. Implement the durable load/append boundary, replacement execution, and effect reconciliation needed to satisfy the common cases. Preserve the distinction between replaying a journal and executing a continuation.

Use the reducer for the decisions it actually owns. If the common scenario needs incarnation fencing or sink bookkeeping outside it, make that code and responsibility visible. Do not reinterpret protocol fields or fabricate evidence to make the kernel appear sufficient.

Tests must inspect the actual effect count and identity, not only the final reducer view. Prove both a useful successful continuation and an honest unresolved opaque-effect case, then run all other common controls.

## What not to build

No fluent API, generic persistence layer, scheduler, live RoxIQ integration, Gate replacement, schema migration platform, or fork of the entire portfolio substrate. Keep one workflow and state the crash durability assumptions.

## Canned demo and story

“Warrant determines what the evidence allows next. The executor survives this lost acknowledgement by reconciling the exact operation. Change the candidate and the old support no longer discharges the requirement. Where the effect is unknowable, the run remains unresolved rather than inventing certainty.”

## Finish

One-command README, canned demo, DEMO.md, green tests, exact reused source identity, and a short account of adapter/recovery code added. Count relevant reused machinery as well as the consumer. Keep the build local.
