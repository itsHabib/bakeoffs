# Agent-substrate bakeoff — scorecard

The build and acceptance facts are filled below. Scores, ranking, and the
promotion decision are intentionally blank for the operator.

Frozen input deck: `sha256:ec742e0129e44d0529bbf5051125b33cfa6bea3c665463062c5da0f1f8947402`

| Criterion (weight) | Branchroom | Proofline | Obligation Engine | Mandate |
|---|---|---|---|---|
| Finish line (repo / README command / canned demo / DEMO.md / local green) | PASS — 5/5 at `5b9ff6e` | PASS — 5/5 at `902d6d5` | PASS — 5/5 at `ad64b24` | PASS — 5/5 at `c768a21` |
| 60-second demo (30) |  |  |  |  |
| Would someone pay or adopt it (25) |  |  |  |  |
| Deterministic share (20) |  |  |  |  |
| Agent-substrate necessity (15) |  |  |  |  |
| Restraint (10) |  |  |  |  |
| **Total (100)** |  |  |  |  |
| Notes |  |  |  |  |

## Verified demo commands

```sh
cd agent-substrates-08-21/branchroom && make demo
cd agent-substrates-08-21/proofline && make demo
cd agent-substrates-08-21/obligation && ./demo.sh
cd agent-substrates-08-21/mandate && ./demo.sh
```

All four copied fixtures match the frozen deck. Their formatting, build or vet,
tests, race tests, golden normalized artifacts, and repeat-run determinism were
independently reproduced at the heads shown above. All four repositories are
clean local repositories with no remote configured.

## Qualification facts — not scores

- Branchroom, Proofline, and Obligation Engine have no known qualification-floor
  blocker at their verified heads.
- Mandate's implementation finish line is green, and its attenuation law is
  demonstrated. Its own strongest-alternative result is `standard caveated
  capability + receipt: MATCH`, however. That activates the brief's explicit
  kill criterion for the custom signed-JSON format; do not promote that format
  as a new substrate on this evidence.
- Assurance remains separate from authority in every entry. None grants merge
  authority or replaces the `evidence-pr` reference reducer.

Tie-breaker — which repository would you attach to the next failed agent run: ____

Winner / promoted primitive (or `none`): ____

Decision rationale: ____
