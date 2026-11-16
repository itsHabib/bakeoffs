# 60-second story

“Our invoice screen only needs a handful of values, but its fixture contains customer metadata, billing details, audit history and links. Manually trimming it risks deleting the flag that enables approval.”

1. Run `./demo.sh` with the local Kitesurf endpoint running. Point out the explicit contract in `examples/invoices/scenario.json`: the customer, formatted amount, overdue state, approval availability and reference.
2. Watch `PASS /debug` or `PASS /invoice/audit`: the consumer still satisfies the contract. A deletion such as `/invoice/canApprove` fails because approval becomes disabled.
3. Open `artifacts/invoices/index.html`. The report shows the response shrinking, the final fixture, and one failing control for each retained field. Open the “Failing mock” next to `canApprove` to see exactly what was removed.
4. Replay the passing fixture with `node src/verify.mjs examples/invoices/scenario.json artifacts/invoices/fixture.min.json`. Replay any exported control and see exit 2 with the failing assertion.
5. Open `artifacts/warehouse/index.html`: same tool, different API shape and business rule. This confirms it is not an invoice-specific answer.

The live run may take longer than the narrated 60 seconds depending on engine startup and candidate count. The reports are the prepared demonstration artifacts. Scope the promise: preserved explicit DOM outcomes in these scenarios, not all API consumers and not cross-browser parity.
