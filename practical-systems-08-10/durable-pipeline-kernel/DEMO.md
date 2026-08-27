# Demo

From this directory:

```powershell
.\check.ps1
```

Watch the final two JSON records:

1. `PipelineRunViewV1` ends on `rev-b`, attempt `2`, and retains both the
   refuted `rev-a` E2E receipt and the passing `rev-b` receipts.
2. `EvidenceReproductionV1` is emitted by a separate process after rereading and
   hashing the committed E2E input fixture. It reports `reproduced: true`.

Then change any input check to `false` without changing the manifest digest and
rerun `gleam run -m replay_evidence`. Reproduction refuses to inherit the old
passing verdict.
