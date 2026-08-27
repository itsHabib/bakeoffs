# Agent-substrate bakeoff — 4 entries, independent sessions, operator judges

Date: 2026-08-21

Topic: **which missing primitive most strengthens long-running agent work
without weakening reproducibility, evidence, or human authority?** Four
self-contained briefs go to four fresh sessions that never see one another.
Each entry ends in a runnable, hands-free demo the operator can judge in under
five minutes.

This is a promotion-priority bakeoff among complementary boundaries, not a
claim that only one primitive may ever exist. The winner must earn the next
investment. **Promote none** is a valid outcome.

## Existing reference kernel — not an entry

Read `https://github.com/itsHabib/formal-methods/blob/main/evidence-pr/README.md` before building any
entry. That POC already establishes the assurance floor:

- authenticated evidence receipt ingestion rather than trusting agent-authored
  JSON;
- exact-head rejection of stale receipts;
- deterministic `SUPPORTED | INSUFFICIENT | REFUTED` reduction;
- refutation dominance and fail-closed duplicate conflicts;
- authoritative claim, producer, method, and policy definitions;
- an unchanged risk floor and `merge_authority: none`;
- Lean aggregation laws and a Quint stale-report counterexample;
- an exact-head consumer boundary even after report publication.

Its deliberate limitation is one PR subject, one Gate claim, one method, and
one required evidence slot. This wave explores missing substrates around a
future generalized assurance bundle. It does **not** reward rebuilding the
existing reducer under a new name.

The common vocabulary is:

- `VerificationContractV1` — frozen task/repository/base/head/diff and policy
  identity, claims and invariants, negative controls, accepted
  producer/method/environment classes, and mandatory versus optional evidence
  slots. An agent may add obligations; it may not weaken the repository or
  risk-floor contract.
- `EvidenceReceiptV1` — contract/head/claim identity, producer, harness, tool
  and version, reproducible method or command, expected and observed behavior,
  evidence digests, counterexample and limitations, with a result of
  `pass | fail | error | skipped | abstain`. A trusted runner captures and
  authenticates it. Shape-valid raw JSON is not evidence.
- `ChangeAssuranceV1` — deterministic per-claim outcomes, accepted and rejected
  receipts, exact-subject freshness, declared coverage, and named gaps, always
  with `merge_authority: none`.

Coverage is transparent, for example **`2/3 required claims supported;
critical-finding-resolved is missing`**. No entry may replace this with an
“87% confidence” score. Self-confidence may appear only as diagnostic
provenance and never changes an outcome.

Each contestant owns one seam in the eventual flow:

```text
VerificationContract
        |
   Branchroom -- controlled run artifact --> trusted runner --> EvidenceReceipt
                                                               |
                                             reference oracle --> ChangeAssurance
                                                                  |          |
                                                   Proofline lineage   Obligation frontier

operator-signed Mandate --------------------------------------> effect gateway
```

The diagram is a boundary map, not shared implementation. Every entry remains
independent and may win without building the other three.

## Entries

| # | slug | sole law demonstrated |
|---|---|---|
| 1 | `hack-branchroom` | A child with a fresh causal epoch rejects a late parent terminal while preserving a provably identical declared prefix |
| 2 | `hack-proofline` | An assurance descendant cannot survive a head, task-revision, or contract-policy identity change in its ancestry |
| 3 | `hack-obligation` | A late refutation derives new open work, and no agent overlay can remove a mandatory contract obligation |
| 4 | `hack-mandate` | A validly signed delegated child cannot widen its parent's exact action scope |

## Exact launch prompts

Use one fresh session per prompt; do not relay progress between them.

> Read `agent-substrates-08-21/entry-branchroom.md`
> and build it. You are one of four independent entries; you win by demo, not
> by design.

> Read `agent-substrates-08-21/entry-proofline.md`
> and build it. You are one of four independent entries; you win by demo, not
> by design.

> Read `agent-substrates-08-21/entry-obligation.md`
> and build it. You are one of four independent entries; you win by demo, not
> by design.

> Read `agent-substrates-08-21/entry-mandate.md`
> and build it. You are one of four independent entries; you win by demo, not
> by design.

## Shared input deck — same bytes, different native artifacts

Every entry must copy, unchanged, the frozen input deck at
`agent-substrates-08-21/fixtures/exact-head-lifecycle-v1.json`.
Its expected byte digest is recorded in `fixtures/SHA256SUMS`.

Verify it with:

```sh
(cd agent-substrates-08-21/fixtures && shasum -a 256 -c SHA256SUMS)
```

The synthetic deck contains:

- work source `task-17` at revision `4`;
- base `H0 = 1111111111111111111111111111111111111111`;
- first candidate `H1 = 2222222222222222222222222222222222222222`;
- repaired candidate `H2 = 3333333333333333333333333333333333333333`;
- a frozen verification contract for each candidate;
- three mandatory claims, trusted-runner receipts, and immutable oracle
  assurance snapshots;
- a critical refutation on `H1`;
- an `H2` repair for which every `H1` receipt is historical but stale;
- partial `H2` assurance of `2/3`, followed by refreshed `3/3` support;
- one planted single-law attack per entry plus raw-JSON and correlated-duplicate
  controls.

This is an **input deck**, not a shared acceptance protocol. An entry may ignore
irrelevant events and attacks. Its brief names the one positive behavior, one
single-law mutant, and one native artifact it owns. Whenever it displays an
oracle assurance snapshot, it must echo the snapshot's summary, coverage, gaps,
and `merge_authority` unchanged.

Every demo emits a small normalized JSON envelope containing:

- `schema_version`, `entry`, and `case_id`;
- `input_bundle_digest` and exact subject identity;
- an entry-native status and stable reason codes;
- the native artifact digest;
- any displayed oracle assurance copied unchanged.

The native status refers only to the contestant's claimed law. It is never a
merge, landability, or quality decision.

## Idea boundaries

- **Branchroom owns controlled causal experiments.** It proves a common
  declared prefix, isolates one perturbation, and produces runner-capturable
  artifacts. It does not assess evidence or carry authority.
- **Proofline owns lineage over oracle outputs.** It indexes contracts,
  receipts, assurance collections, and consumers across subject revisions. It
  never recomputes a claim outcome, schedules missing work, or says work is
  landable.
- **Obligation Engine owns the open evidence frontier.** It consumes frozen
  contract slots and oracle gaps to derive typed work. It never judges raw
  receipts, runs workers, or authorizes effects.
- **Mandate owns delegated authority.** It verifies offline signature ancestry,
  attenuation, audience-key request binding, and exact scope. If the demo
  consumes an effect, one named gateway owns that state. It never treats
  assurance as authority or decides work quality.

## Qualification floor — an entry that crosses this cannot win

- Its native positive control must succeed. “Always refuse” is not safety.
- Its single-law mutant and production mechanism must run over byte-identical
  input and identify the same golden first-failure step every time.
- Two identical runs must produce byte-identical normalized JSON.
- The mechanism must exercise its reducer, verifier, or lineage law. Hardcoding
  case IDs or snapshotting expected output earns no deterministic credit.
- Raw agent-authored JSON must never become trusted provenance.
- No entry may recompute the assurance oracle, average away a refutation, turn
  `error`, `skipped`, `abstain`, missing, stale, or unknown into support, inflate
  coverage with duplicate/correlated receipts, emit a probability/confidence
  score, weaken a frozen mandatory contract or risk floor, or claim merge
  authority.
- The correctness path must remain offline, keyless, model-free, and free of
  mutable portfolio state.
- The brief-specific cheapest comparator must not produce the same native
  artifact with materially less mechanism.

## House rules — identical for all four entries

- **One session, one fresh repo.** Scaffold at `~/dev/<slug>`. Entries
  never share context, code, libraries, or progress reports.
- **Small local stack.** Go 1.26 stdlib is preferred; plain Node 26 or Python
  3.14 stdlib is acceptable when it makes the core clearer. No web framework,
  database, daemon, container, or new dependency unless the demo is impossible
  without it.
- **Keyless and hands-free.** No API key, GitHub login, cloud service, live
  model, Firecracker host, or network access may be required. Ollama
  `qwen2.5:7b` exists but the demo must remain correct with Ollama stopped. A
  model may phrase output or be the subject under test; it may never grade,
  authorize, resolve, authenticate, or establish correctness.
- **Synthetic fixtures only.** Do not copy private transcripts, grants, issue
  bodies, or repository data into the entry. Existing portfolio repos may be
  read for mechanism patterns, but entry code is self-contained and imports
  none of them.
- **Correctness is computed, never model-judged.** Every state transition,
  digest, match, invalidation, attenuation, refusal, and score is deterministic
  and table-tested.
- **No spec and no design doc.** Deliver a README, working code, focused policy
  tests, and the required `DEMO.md`. Simplify until the one claim is undeniable.
- **Target at most 700 weighted source lines** excluding tests and the frozen
  fixture. LOC is disclosure, not a cross-language score; excess mechanisms
  lose restraint.
- **Required finish line:**
  1. README with exactly one obvious command to run;
  2. canned demo requiring no live input;
  3. `DEMO.md` containing the exact 60-second walkthrough;
  4. local green: formatting plus build/vet/tests appropriate to the language;
  5. normalized JSON envelope, golden output test, and final source line count.

## Judging — operator only, after all four land

100 points:

- **30 — the 60-second demo.** Does the entry's native positive case,
  single-law mutant, exact refusal or retraction, and artifact make the
  primitive obvious without reading its architecture?
- **25 — would someone pay or adopt it.** Name the buyer hypothesis, the
  already-funded pain, the smallest insertion seam, and a credible shadow
  field test. The judge will push back on invented markets or adoption that
  first requires a platform replacement.
- **20 — deterministic share.** How much of the claimed invariant is executable
  adversarial mechanism rather than prose or model judgment? Show the tests,
  byte-stable artifact, and one-field negative control.
- **15 — agent-substrate necessity.** Demonstrate an incremental invariant
  beyond `evidence-pr` and beat the brief's strongest cheap alternative. A
  renamed oracle, exact-subject join, worktree, conditional scheduler, or
  standard caveated capability does not earn these points.
- **10 — restraint.** One law, small surface, precise refusals, few additional
  trusted components, honest non-goals, and no framework seams.

Tie-breaker: **which repository would the operator actually attach to the next
failed agent run?**

## After judging

The winner, if there is one, earns a reviewed TDD. The current landing
hypothesis is deliberately split:

- `formal-methods` retains proofs and reference semantics;
- Workbench would own versioned production schemas and evaluation;
- `cc-skills` `/validation-card` would capture and render the bundle;
- Drive would consume exact-head outcome, declared coverage, and named gaps;
- Gate and operator-minted mandates remain the authority boundary.

Before production, the promoted primitive must survive a shadow-only field
experiment. A strong external testbed is Alexandre Malucelli's seven-loop
setup, especially its `counterpart-drift` loop: instrument one authorized loop
to emit contracts, trusted receipts, and assurance, then compare recorded gaps
and refutations with CI, high-effort review, human disposition, and later fixes
or reverts. The experiment observes; it never merges or expands agent
authority. See [Malucelli's 12-day loop write-up](https://malucelli.net/posts/2026-08-18-claude-code-12-days-straight/).

The other entries remain independent reference experiments. No automatic
platform merge occurs, and the shadow stage reports observed delta and overhead
rather than an “accuracy” or confidence percentage.
