# Entry 3: hack-obligation — coordinate the mandatory evidence frontier

Read `agent-substrates-08-21/README.md` first. Its house
rules, assurance boundary, frozen input deck, and judging apply verbatim.

Repo: `agent-substrates-08-21/obligation`. You never see the other three entries.

## The bet

Named-node workflows encode who should run and when, but evidence can create
work after every planned node returned: a counterexample refutes a claim, a
repair changes the subject, or a contract gains a mandatory slot. The bet is
that coordination should center on a monotone, typed evidence frontier.
Replaceable workers may add work or submit oracle outputs; they cannot declare
the repository's mandatory obligations gone.

Buyer hypothesis to validate before promotion: multi-agent assurance and
incident workflows will adopt a deterministic frontier if it prevents “all
agents returned” from being confused with “all mandatory evidence exists.”

## One law

Mandatory evidence work is monotone under later information and agent overlays:
a late refutation can reopen a settled goal, and an agent may add but never
remove or downgrade a mandatory contract obligation.

## What to build

Build one local CLI/library with a closed, fixture-specific
`ObligationFrontierV1`:

- exact goal identity: task revision plus contract and subject;
- three obligation kinds only: `collect_claim`, `resolve_refutation`, and
  `refresh_subject`;
- stable obligation id, claim id, subject, origin (`contract | derived | agent`),
  mandatory flag, state, and prerequisite ids;
- inputs consisting only of frozen contract slots, immutable oracle
  `ChangeAssurance` snapshots, and an additive agent overlay;
- a pure reducer that derives the canonically ordered open frontier and an
  explanation for every open, discharged, reopened, or superseded item;
- append-only events from which the same frontier is reconstructed.

Do not ingest or judge raw receipts. The oracle has already decided supported,
insufficient, and refuted. In the frozen deck:

1. Instantiate three mandatory `contract-h1` collection obligations.
2. Consume `aligned-h1` and reach an empty evidence frontier.
3. Consume the later `refuted-h1`; reopen work as one
   `resolve_refutation(critical-finding-resolved)` obligation.
4. Apply the `H2` repair; supersede the old subject, open
   `refresh_subject`, then derive only the gaps named by `partial-h2`.
5. Consume `refreshed-h2` and return to an empty evidence frontier.
6. Reject an agent overlay that deletes or downgrades
   `critical-finding-resolved`; accept an additive optional diagnostic.

An empty evidence frontier is not authorization. Do not add a human-approval or
effect-execution phase.

## Single-law mutant

Implement `settled-is-terminal`: once a goal first reaches an empty frontier,
later assurance snapshots cannot reopen it. Run the late-refutation attack over
byte-identical input. The mutant must remain empty; production must derive the
mandatory repair obligation with reason `mandatory_obligation_reopened`. Do not
disable contract-floor enforcement in this mutant.

## Strongest cheap alternative

A named-node scheduler with conditional branches, event hooks, and an explicit
reopen transition—not a four-box checklist designed to lose. Obligation Engine
wins only if contract-derived identities, monotone overlays, and the replayable
frontier make the late work materially safer or simpler.

Kill the bet if that scheduler produces the same frontier and explanation with
less machinery, if a manager/model must interpret evidence, or if this entry
evaluates receipts instead of consuming oracle gaps.

## Required tests

- deterministic contract-to-obligation ids and frontier ordering;
- `aligned-h1` empties the native evidence frontier;
- a late refutation reopens exactly one mandatory repair item;
- head/contract supersession prevents stale discharge;
- `partial-h2` opens exactly its named gap and `refreshed-h2` closes it;
- delete/downgrade overlays reject while additive overlays survive replay;
- raw receipt input cannot discharge anything;
- event replay and normalized demo JSON match checked-in golden files.

## What NOT to build

- No agent launcher, queue, worker registry, process supervisor, message bus,
  marketplace, bidding, model router, or actual scheduling.
- No receipt parser/evaluator, review-content adjudicator, assurance reducer,
  authorization, effect execution, database, network, or UI.
- No general predicate language, workflow DSL, Datalog/Prolog interpreter,
  theorem prover, plugin architecture, or natural-language planning.

## Canned demo

One command runs tests and the scheduler/mutant/production comparison:

```text
aligned H1 frontier: empty
late REFUTED snapshot arrives:
  settled-is-terminal mutant: empty  <-- planted bug
  obligation engine: OPEN resolve_refutation(critical-finding-resolved)
repair -> H2 partial: OPEN collect_claim(critical-finding-resolved)
agent remove mandatory claim: REFUSE mandatory_contract_weakened
refreshed H2 frontier: empty  (evidence only; no authority)
```

Then print `ObligationFrontierV1`, its digest, common envelope, and source line
count.

## The 60-second demo story

“Every planned worker returned, and the frontier was empty. Then the oracle
recorded a counterexample. A terminal scheduler ignores it; the Obligation
Engine derives new mandatory work from the frozen contract. The repair changes
the subject, so old work cannot discharge the new gap. Agents may add a useful
diagnostic, but cannot delete the repository's floor. Empty means evidence work
is satisfied—not that anyone may act.”

