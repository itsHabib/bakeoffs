# Entry 3: Parley — protocol order with honest execution boundaries

Read [`../RULES.md`](../RULES.md) first. All cases, effect contracts, invariants,
and finish-line requirements apply.

This entry was built in a fresh repository without inspecting its competitors.

## The bet

Deriving sequencing obligations from a protocol may make interrupted multi-role work easier to implement and change. The protocol only owns order. Evidence identity, persistence, incarnation fencing, and effect reconciliation must remain explicit rather than being credited to a theorem that does not cover them.

## Existing source

The entry used Parley's compiler/observer interface and enforcement core at
commit `ef7cce0`. The documented Lean results concern concrete protocols and
traces; do not claim a general correctness theorem.

Use the compiler/observer or the pure enforcement core as appropriate, but state whether the protocol is gating execution or auditing it. If it is only an observer, the ordinary executor must still satisfy every common requirement, and the observer's incremental benefit must be visible. Do not present observation as prevented effects.

Reuse minimal pinned source with its license or a documented pinned local dependency. Keep all adaptations in the new entry. Check available runtimes; do not install a toolchain as part of this round.

## What to build

Model the common worker/checker/recorder interaction with a small protocol. Preserve raw event identity through normalization and retain the exact point of any deviation. Implement or reuse a small local execution shell for persistence, sink interaction, restart, and subject checks.

Demonstrate recovery through the same process-level interruptions as the other entries. Replay must reconstruct the relevant protocol state without pretending an unacknowledged effect is absent. An old incarnation may obey the same message sequence and still be invalid; implement that distinction outside pure order checking.

Include a legal trace, planted ordering deviation, and all common mutants. Show whether the protocol detects anything the independent effect/identity assertions do not already detect.

## What not to build

No general agent bus, new authentication system, general projection proof, replicated actor state, live Org adapter, universal protocol language extension, or dashboard. Narrow or report a missing expressive feature rather than enlarging the algebra without need.

## Canned demo and story

“The protocol says which interaction is legal next. The journal and sink establish what actually committed. After the worker dies, those two views agree on a valid continuation—or expose what is missing. A correctly ordered message from the wrong incarnation still cannot complete the run.”

## Finish

One-command README, hands-free demo, DEMO.md, green checks, retained artifacts, and the exact reused source identity. Report which Haskell/Gleam/Lean components were executed and which were not. Count the ordinary executor and adapters along with Parley. Keep the build local.
