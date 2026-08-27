# Entry 4: hack-mandate — prove delegated authority only narrows

Read `agent-substrates-08-21/README.md` first. Its house
rules, assurance boundary, frozen input deck, and judging apply verbatim.

Repo: `agent-substrates-08-21/mandate`. You never see the other three entries.

## The bet

Agents are commonly handed ambient credentials plus prose describing what they
may do. The credential travels; the restriction does not. The bet is that a
small signed work mandate can carry exact work identity and monotonically
narrow across one delegation hop. Any offline verifier can check signature
ancestry, attenuation, and a request signed by the audience key without
trusting the presenting agent.

Buyer hypothesis to validate before promotion: agent-platform and security
teams will adopt a portable work-specific authority chain when existing scoped
credentials stop at a runtime boundary and leave delegation restrictions in
prompts.

## One law

A validly signed child mandate may only narrow its parent's allowed action set.
A child signed by the right delegate but adding one action is invalid even
though every signature verifies.

## What to build

Build one local CLI/library implementing one-hop `WorkMandateV1` with Go stdlib
Ed25519 or an equally standard local primitive. Use checked-in fixture keys,
injected logical time, exact resources, and canonical serialization.

The signed payload contains:

- mandate id, issuer public key, audience public key, and parent mandate digest;
- exact task id/revision, repository, base, head, and diff digest;
- a closed allowed-action list: `inspect_candidate`, `review_candidate`, or
  `publish_candidate`;
- issued-at, expiry, and remaining delegation depth;
- optional opaque artifact digests, compared only for equality and never
  interpreted as evidence or authority.

Implement:

- root signature verification;
- one child delegation where child issuer equals parent audience, parent digest
  matches, exact work identity cannot change, actions are a subset, expiry
  cannot grow, and depth decrements;
- request verification where the request is signed by the child audience key
  and its exact action/work identity matches;
- a deterministic `MandateDecisionReceiptV1` that records the verified chain,
  request digest, allow/deny, and stable reason code.

The fixture root is synthetic test authority, never an operator grant. The
demo may simulate a named effect gateway, but it performs no effect. Offline
verification does **not** promise global replay prevention, revocation, or
exactly-once consumption.

Use `contract-h2` to show that even an oracle `SUPPORTED` snapshot with
`merge_authority: none` grants no authority. The root permits
`inspect_candidate` and `review_candidate`; its child narrows to one exact
`review_candidate` request. The valid request receives an allow receipt.

## Single-law mutant

Implement `signatures-only`: verify both signatures and the parent link but
skip only the child-action subset check. Feed it a correctly signed child that
adds `publish_candidate`. The mutant must allow; production must deny the same
chain with `scope_inflation`. Do not corrupt a signature or disable another
attenuation rule.

## Strongest cheap alternative

A standard caveated capability—macaroon- or Biscuit-like attenuation—with an
audit receipt, not a bearer token plus good intentions. Mandate earns
substrate-necessity points only if its exact work binding and decision artifact
add something the standard capability does not provide at lower cost.

Kill the bet if signed JSON is the whole novelty, if a standard caveated token
plus receipt gives the same result, if assurance becomes authority, or if the
entry claims decentralized replay/theft prevention without an authoritative
gateway and caller authentication.

## Required tests

- fixed-key canonical payload, signature, and digest stability;
- valid root, narrowed child, audience-signed exact request, and allow receipt;
- validly signed action widening fails `scope_inflation`;
- wrong child issuer, parent digest, audience request signature, subject, or
  expiry fails closed;
- unknown fields and unknown actions fail closed;
- oracle assurance alone never satisfies mandate verification;
- normalized demo JSON and decision receipt match checked-in golden files.

## What NOT to build

- No use counter, replay ledger, concurrency race, distributed consumption,
  revocation service, or exactly-once claim.
- No OAuth/OIDC, PKI discovery, KMS, secret broker, blockchain, payment
  protocol, custom cryptography, identity platform, wallet, or approval UI.
- No MCP/A2A transport, network, GitHub integration, account system, database,
  workflow planning, scheduling, evidence evaluation, or quality judgment.

## Canned demo

One command runs tests and the standard-capability/signatures-only/production
comparison:

```text
oracle H2: SUPPORTED  coverage=3/3  merge_authority=none
without mandate: DENY no_authority
root -> child review_candidate: signatures valid
audience-signed review request: ALLOW receipt sha256:...
validly signed child adds publish_candidate:
  signatures-only mutant: ALLOW  <-- planted bug
  mandate: DENY scope_inflation
```

Then print `MandateDecisionReceiptV1`, its digest, common envelope, and source
line count.

## The 60-second demo story

“A supported assurance report still says it has no merge authority. The worker
acts only through a signed child mandate bound to this exact work and its own
public key. A signature-only verifier accepts a child that adds publish; the
attenuation law rejects the same valid signatures. This proves what authority
was delegated and why this request fits. It does not pretend offline signatures
solve global replay or revocation.”

