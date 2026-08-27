# Protocol compiler — frozen brief

## The bet

A global protocol should be one value from which a compiler can derive each
role's local obligations and reject conversations whose sends, receives,
choices, or terminal states cannot agree. Haskell earns its seat if the
protocol algebra and its interpreters make projection and validation direct,
total, and difficult to accidentally extend inconsistently.

## User and concrete workflow

An author of an agentic workflow spanning a collector, verifier, human judge,
and merge gate. Today the order and branch obligations live across prose,
handlers, and tests. The useful output is a role-by-role contract plus a precise
compiler refusal before any agent runs.

## One job

Project a global, branching protocol into local role traces and verify that
every communication has compatible sender/receiver obligations on every path.

## Build

- A small global protocol algebra: message, sequence, binary choice, and end.
- Projection into local `Send`/`Receive`/`Choose`/`Offer`/`Done` obligations.
- Validation with path-aware diagnostics for incompatible or uninvolved roles.
- A deterministic rendering of the local contracts.
- One valid evidence-carrying-PR flow and one subtly invalid variant.

## Required hard cases

- A role uninvolved in a choice must observe compatible continuations.
- Both branches must terminate or continue coherently for every role.
- A message cannot target its sender.
- Projection errors identify the role and branch path.

## Haskell thesis

The global protocol is an algebraic program value. Projection, validation, and
rendering are separate total interpreters over the same closed syntax. The
comparison implementation is an idiomatic Go AST plus visitors; the candidate
wins only if Haskell makes illegal interpreter drift or partial case handling
materially harder, not merely shorter.

## Demo

Compile the valid collector → kernel → human/gate protocol and print four local
contracts. Then compile a variant where an uninvolved role cannot distinguish
which branch occurred; reject it at the exact role and branch.

## Kill condition

Kill if the useful guarantee reduces to ordinary runtime graph validation, or
if adding a protocol form still requires manually synchronized ad-hoc cases in
every interpreter without meaningful compiler help.

## Non-goals

No network runtime, code generation, parser, schema registry, MCP adapter,
temporal logic, or general theorem prover.
