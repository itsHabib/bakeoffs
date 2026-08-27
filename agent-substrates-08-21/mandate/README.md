# hack-mandate

`hack-mandate` is a local proof that signed delegation has to verify more than
signatures. A root mandate binds authority to one exact task revision,
repository, base, head, and diff. Its one child may shorten the lifetime,
decrement depth, and remove actions; it cannot change the subject, change an
opaque artifact digest, or add an action. The audience signs the exact request.

Run the entire entry:

```sh
./demo.sh
```

That one command verifies formatting and the frozen input digest, runs vet and
the race-enabled tests, then prints the positive request, the planted
`signatures-only` bug, the production refusal, a deterministic decision
receipt, its digest, the common envelope, and the source-line count.

## What the demo proves

The frozen `contract-h2` oracle snapshot is copied unchanged: `SUPPORTED`,
`3/3 required claims supported`, no gaps, and `merge_authority: none`. With no
mandate, the simulated gateway returns `DENY no_authority`. A synthetic root
permits inspect and review; its child narrows to review for H2; the worker-key
request is allowed. A second correctly signed child adds publish. The mutant
that skips only the subset check allows it, while production denies the same
bytes with `scope_inflation`.

`MandateDecisionReceiptV1` records accepted chain digests, the signed request
digest, exact subject, action, decision, and stable reason. JSON is serialized
from closed Go structs in declared field order. External mandate JSON can be
decoded with `DecodeStrict`, which rejects unknown fields; shape validation
rejects unknown actions, duplicate or unsorted scopes, malformed keys, and
non-canonical timestamps. Logical time is injected.

The checked-in Ed25519 seeds are conspicuously synthetic fixture authority.
They are not an operator grant and must never be reused.

## Buyer and insertion seam

The buyer hypothesis is an agent-platform or security team whose runtime-scoped
credential survives a delegation hop while its prompt restriction does not.
The smallest insertion seam is an effect gateway: before a named publish or
review effect, verify a portable mandate chain and the caller-key signature,
then retain the receipt alongside the attempted effect. No scheduler, identity
platform, secret broker, or assurance evaluator has to move.

A credible shadow field test records receipts at one existing gateway without
performing new effects. Compare the work/action restrictions the gateway would
enforce with the restrictions humans thought they delegated, including a
planted correctly signed widening attempt. Measure mismatches and verification
overhead; do not call the shadow decision authorization.

## Strongest cheap alternative and kill result

A standard caveated capability such as a macaroon or Biscuit-like token, plus
an audit receipt, enforces the same attenuation law and can carry exact work
caveats. The demo reports that comparator as `MATCH`. This entry therefore
establishes the need for work-bound attenuated authority, but it does **not**
establish that a custom signed-JSON format deserves promotion over a standard
capability profile. If the standard primitive yields the same portable receipt
at lower cost in the shadow test, kill the custom format and keep the schema as
an interoperability profile only.

## Honest boundary

Offline verification proves signature ancestry, one-hop attenuation, audience
binding, exact request scope, and why this local decision was made. It does not
provide revocation, caller key custody, global replay prevention, distributed
consumption, exactly-once effects, evidence quality, landability, or merge
authority. The demo gateway performs no effect.

The frozen input deck is copied byte-for-byte from the bakeoff. Golden files
lock the normalized allow receipt and envelope. Tests cover canonical payload,
signature, and digest stability; every required attenuation failure; unknown
fields and actions; the oracle/authority separation; and byte-identical repeat
runs.
