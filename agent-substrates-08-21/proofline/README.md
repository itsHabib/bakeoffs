# hack-proofline

Proofline is a read-only lineage index over frozen verification contracts,
trusted receipts, immutable oracle assurance snapshots, and downstream
observations. It answers one question: which exact identity edge made an old
assurance and everything that consumed it historical?

Run the whole entry:

```sh
make demo
```

That command formats-checks, vets, and tests the implementation before running
the canned trace, the exact-subject relational comparator, the head-only
mutant, and the normalized JSON artifact. It is offline, keyless, model-free,
and uses only Go's standard library.

## Boundary

Proofline copies oracle `summary`, `coverage`, `gaps`, and `merge_authority`
without interpreting them. It never authenticates receipts, recomputes an
assurance outcome, finds missing work, scores confidence, or calls a candidate
landable or authorized. `current` and `historical` describe lineage only.

The graph is closed: contracts, receipts, assurances, and consumer observations
carry complete canonical identities; the only relations are `evaluates`,
`includes`, `derived_from`, `supersedes`, `invalidates`, and `consumed_by`.
Construction rejects missing identities, dangling edges, duplicate-ID
conflicts, invalid kind relations, identity mismatch, and cycles. A query emits
the canonically ordered ancestry, recorded supersession/invalidation changes,
and all transitive consumers for one assurance. Retraction occurs only when the
graph itself contains both the old-to-new `supersedes` edge and the new-to-old-
assurance `invalidates` edge. Conflicting duplicate source IDs are rejected
before fixture records enter adapter maps.

The planted drift is selected as the unique nonempty fixture attack whose
declared reason is `contract_identity_mismatch`; zero or multiple matches are
invalid input. Its fixture key is carried unchanged into the demo envelope, so
renaming the fixture record changes the emitted case identity without changing
the lineage mechanics.

## Buyer test and comparator

Buyer hypothesis: a release/compliance or agent-platform team already retaining
assurance reports will adopt one portable lineage artifact when regenerated
contracts make incident impact analysis a manual reconstruction.

The smallest insertion seam is after the assurance oracle and before reports
reach release and compliance consumers. A shadow field test records only
digests and exact identities for those existing artifacts, then asks during a
real contract regeneration: “which observations became historical, and at
which first edge?” It does not alter the delivery path.

The strongest cheap alternative in the demo is an exact-subject relational
join. It correctly returns all H2 rows. Because exact equality deliberately
excludes H1, its mechanically computed joined-impact set is empty even though
the recorded invalidation and recursive consumer traversal identify three
impact candidates. Adding those history relations is precisely the mechanism
Proofline packages.
Kill this bet if a team's existing exact-subject SQL already returns the same
ordered subgraph, first stale edge, and transitive retraction set with less
machinery.

The frozen fixture is copied byte-for-byte from the bakeoff deck; its digest is
checked in `fixtures/SHA256SUMS`.
