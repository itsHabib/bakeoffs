# 60-second walkthrough

Run `make demo`.

“The assurance oracle already decided what each collection establishes; I do
not decide it again. Proofline shows what that report came from and what later
consumed it. First, H1 and all three receipt branches are current. H2 makes the
H1 assurance and both downstream observations historical at the contract-to-
assurance edge, while a fresh H2 graph is current and the H1 graph remains
queryable. Removing either the supersession or invalidation edge makes the
retraction disappear; the change is coming from the graph, not a caller's
asserted expected identity.

The cheap exact-subject join finds every H2 row and mechanically reports three
impact candidates but zero joined H1 descendants. Then the planted attack
changes the task revision and policy while keeping the H1 SHA. A head-only
index stays current.
The full lineage retracts the assurance and both transitive consumers with
`contract_identity_mismatch`.

The ordered graph and its digest are the portable artifact. The copied
`SUPPORTED`, `3/3`, and `authority=none` are still the oracle's words. This is
genealogy for assurance, not another gate.”
