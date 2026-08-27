# Flow State Lab — 60-second demo

From this directory:

```bash
npm run check
```

Narrate the output:

1. Both Quint files typecheck.
2. The safe protocol survives 10,000 sampled executions.
3. Apalache finds no violation within the 12-step exhaustive bound.
4. Each of the four unsafe modules is expected to fail verification; its minimal ITF trace is
   normalized into a JSON fixture.
5. Node tests confirm the last state is the intended failure, not merely any
   model-checker error.

Then open `fixtures/generated/duplicate-save-after-lost-receipt.json`. Its last
transition shows `PayloadSaved`, `receiptRecorded: false`, and
`flowSaveCount: 2`: the exact retry ambiguity the model is meant to expose.
