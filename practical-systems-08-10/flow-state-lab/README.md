# Flow State Lab

Flow State Lab model-checks one practical protocol: a CAM operator client saves
a programming payload to a factory workflow service, records a local receipt,
completes its local task, and then completes the remote workflow.

The model is deliberately generic and synthetic. It contains no employer code,
identifiers, URLs, or incident data.

## The result

The safe protocol maintains seven invariants across every state reachable within
12 transitions. Four intentionally unsafe transitions produce minimal traces:

| fixture | failure exposed |
|---|---|
| `duplicate-save-after-lost-receipt.json` | retry repeats an already-successful remote effect |
| `retained-command-without-recovery.json` | durable custody blocks retry but offers no recovery |
| `local-completion-before-receipt.json` | local terminal state outruns durable remote evidence |
| `invalidation-orphans-remote-payload.json` | local invalidation abandons an uncompensated remote effect |

Those traces are normalized from Quint's ITF output into small, stable JSON
fixtures. The tests use only Node's standard library; an application can copy a
fixture into Vitest/Jest without adopting Quint.

## Run it

Requirements: Node.js 18+ and Java 17+.

```bash
npm install
npm run check
```

`npm run check` performs five independent checks:

1. typechecks the safe and unsafe Quint modules;
2. samples 10,000 executions of the safe model;
3. exhaustively verifies the seven safe invariants to a 12-step bound;
4. regenerates all four counterexample fixtures from failed verification runs;
5. tests the semantic final state of each fixture.

Quint 0.32 cannot directly spawn its downloaded `.bat` verifier on Windows.
The small runner in `scripts/quint-runner.mjs` starts the same local Apalache
server through `cmd.exe`, waits for its port, and tears down that process tree.
It also reuses Quint's pinned downloader on the first run, so the documented
command works from a clean checkout. Other platforms use Quint's normal server
behavior.

`npm audit` currently reports the pinned Quint CLI through its transitive
`adm-zip` dependency (GHSA-xcpc-8h2w-3j85). Quint 0.32 is the latest release and
has no patched dependency line yet. This lab only lets Quint unpack its own
downloaded Apalache distribution; it must not be repurposed to process
untrusted ZIP input while that advisory remains open.

## Read the model

- `model/CamFlow.qnt` is the safe protocol and its invariants.
- `model/KnownFailures.qnt` imports that protocol and adds one unsafe transition
  per failure.
- `scripts/export-fixtures.mjs` turns raw ITF counterexamples into the stable
  files under `fixtures/generated/`.
- `test/fixtures.test.mjs` is the example of consuming them from an ordinary
  runtime test suite.

The important distinction is `PayloadSaved` versus `receiptRecorded`. A remote
effect and durable local knowledge of that effect are not the same state. Retry
must reconcile those facts before repeating the effect. Likewise, retention of
a durable completion command is safe only if it remains resumable or explicitly
releasable.

## Integration seam

The fixture schema is intentionally boring:

```json
{
  "scenario": "duplicate-save-after-lost-receipt",
  "source": { "violatedInvariant": "uploadIsDeduplicated" },
  "trace": [
    { "step": 0, "action": "init", "state": {} }
  ]
}
```

A production adapter should map each fixture action to the application's pure
transition reducer or mocked API boundary, then assert the application refuses
the final unsafe action. That conformance adapter is intentionally outside this
bakeoff: until it exists, the checker proves the model, not production code.

## Promote or kill

Promote the approach if the fixture adapter catches a real transition regression
or makes a workflow review materially clearer. Kill it if maintaining the model
and adapter costs more than direct property/state-machine tests or if the two
drift without detection.
