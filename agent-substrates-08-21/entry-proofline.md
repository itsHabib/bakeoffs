# Entry 2: hack-proofline — retain assurance lineage without recomputing truth

Read `agent-substrates-08-21/README.md` first. Its house
rules, assurance boundary, frozen input deck, and judging apply verbatim.

Repo: `agent-substrates-08-21/proofline`. You never see the other three entries.

## The bet

An assurance report can be correct and still become hard to reason about after
its task, contract, policy, candidate, or consumer changes. The existing
`evidence-pr` kernel evaluates one exact collection; it does not retain a
cross-collection explanation of which later artifacts descended from it and
which identity change retracted them. The bet is that a small, read-only
lineage index makes that history mechanically explainable without becoming a
second assurance engine.

Buyer hypothesis to validate before promotion: release, compliance, and agent
platform teams will adopt a portable lineage artifact when an exact-subject
database query cannot explain transitive impact across regenerated contracts
and downstream consumers.

## One law

An assurance descendant is current only while every identity edge in its
ancestry—task revision, repository/base/head/diff, contract, and policy—still
matches. A changed ancestor retracts descendants but never deletes history or
changes the oracle's recorded outcome.

## What to build

Build a read-only local CLI/library over the frozen input deck. Define a closed
`ProoflineV1` graph:

- opaque nodes for verification contract, trusted receipt, immutable oracle
  assurance snapshot, and downstream consumer observation;
- exact canonical identity on every node;
- edges `evaluates`, `includes`, `derived_from`, `supersedes`, `invalidates`,
  and `consumed_by`;
- validation for dangling edges, duplicate-id conflicts, forbidden cycles, and
  identity mismatches;
- a query returning one canonically ordered **lineage subgraph**, not a single
  path, for a selected assurance collection;
- transitive retraction that identifies the first stale edge and every affected
  descendant while keeping the historical subgraph queryable.

Use the deck to show:

1. the `aligned-h1` assurance with its exact contract and receipt ancestry;
2. the `H2` subject change retracting all `H1` descendants;
3. fresh `H2` assurance creating a new current subgraph without deleting `H1`;
4. the planted same-head contract drift—task revision and policy change while
   head stays `H1`—retracting the old assurance and consumer observation.

Whenever Proofline prints `SUPPORTED`, `INSUFFICIENT`, `REFUTED`, coverage, or
gaps, copy those fields byte-for-byte from the oracle snapshot. Proofline may
label lineage `current`, `historical`, or `invalid_input`; it may not derive a
claim outcome, missing-work frontier, confidence score, or landability verdict.

## Single-law mutant

Implement `head-only-identity`: lineage equality checks repository and head but
not task revision, contract identity, or policy digest. Run the same-head drift
attack over byte-identical input. The mutant must retain the descendant; the
production index must retract it with `contract_identity_mismatch`. Do not
disable another graph check.

## Strongest cheap alternative

An exact-subject relational join over contract, receipt, assurance, and
consumer tables—not a knowingly unsafe “latest green” query. Proofline wins
only if the ordered subgraph and transitive impact explanation answer a real
cross-revision question that the join does not answer as directly.

Kill the bet if one exact-subject SQL query produces the same explanation with
less machinery, if the graph recomputes an oracle outcome, or if it describes
any candidate as landable or authorized.

## Required tests

- canonical node/edge order and byte-stable graph digest;
- complete `H1` and fresh `H2` lineage subgraphs;
- first stale edge on head change and on same-head contract drift;
- transitive descendant retraction without historical deletion;
- dangling edge, forbidden cycle, duplicate conflict, and missing identity;
- oracle fields are emitted unchanged;
- normalized demo JSON matches a checked-in golden file.

## What NOT to build

- No assurance reducer, receipt authenticator, evidence rescoring, confidence
  metric, missing-obligation query, Gate clone, or authority edge.
- No database, daemon, dashboard, connector, watcher, mutation of source
  artifacts, general graph language, Datalog engine, ontology, or plugins.
- No judgment about whether a review finding or receipt is true.

## Canned demo

One command runs tests, the exact-subject comparator, and the byte-identical
mutant/production trace:

```text
H1 lineage: current  oracle=SUPPORTED  coverage=3/3  authority=none
H2 created: H1 lineage historical  first stale=edge contract-h1 -> assurance-h1
fresh H2 lineage: current  oracle=SUPPORTED  coverage=3/3  authority=none
same head, task/policy changed:
  head-only mutant: current  <-- planted bug
  proofline: retracted contract_identity_mismatch
```

Then print the ordered `ProoflineV1` subgraph, retraction set, artifact digest,
common envelope, and source line count.

## The 60-second demo story

“The assurance oracle already decided what each collection establishes; I do
not decide it again. Proofline shows what that report came from and what later
consumed it. A head-only index survives a policy change on the same SHA. The
full lineage retracts the report and its descendants at the first mismatched
identity while preserving the old history. This is genealogy for assurance,
not another gate.”

