# Fixture Pruner

**Turn an oversized JSON mock into a small, behavior-checked fixture.**

For a frontend engineer maintaining invoice, inventory or admin screens: point the tool at a local static app, one JSON response and a few explicit DOM assertions. Fixture Pruner deletes candidate fields, loads the real consumer in Kitesurf, and keeps a deletion only when every assertion still passes. It exports the reduced response, a readable report, and failing fixtures that explain why each retained field matters.

The recurring job is removing irrelevant response data from committed mocks without silently changing what the screen does. No model judges the result.

Recorded real-engine runs: invoice fixture **735 → 150 bytes**, warehouse fixture **561 → 95 bytes**, with **14 replayable failing controls** between them. Browse the saved [invoice report](evidence/invoices/index.html) and [warehouse report](evidence/warehouse/index.html), or rerun the demo below.

## Run

Requires Node 22+, npm, and an already running **local Kitesurf** engine. No cloud account, paid service or browser download is used. This repository does not build or start the engine.

```sh
# In this directory, with the Kitesurf engine running:
export KITESURF_ENDPOINT=ws://127.0.0.1:18891/devtools/browser
export KITESURF_REVISION=4a2903813a22b08d4f3cc64a4b35339d0293a5d1
./demo.sh
```

One command installs the locked Node dependencies if missing, checks syntax, runs tests, and runs both workloads hands-free. It prints each accepted/rejected deletion. Read `artifacts/invoices/index.html` and `artifacts/warehouse/index.html` afterward. The engine must be running before the command; discovery failure exits 3, without fabricated results or a Chromium fallback.

```sh
npm ci
node src/cli.mjs examples/invoices/scenario.json artifacts/my-run
node src/verify.mjs examples/invoices/scenario.json artifacts/my-run/fixture.min.json
# Expected failure (exit 2); deletes a field whose absence broke the consumer:
node src/verify.mjs examples/invoices/scenario.json artifacts/my-run/controls/01.json
```

The HTTP fixture host binds `127.0.0.1:8891` by default; set `PORT` if occupied. Serve reports with `python3 -m http.server 8892 --bind 127.0.0.1 --directory artifacts`, then open `http://127.0.0.1:8892/invoices/`. Static HTML reports contain no scripts or remote resources.

## Bring a consumer

The narrow supported interface is a **locally built/static web app that fetches one same-origin JSON API path on page load**. The tool serves the app directory and substitutes that path. Build a framework app to static files first. It does not proxy production, intercept third-party traffic, emulate an authenticated backend, or orchestrate clicks. Use synthetic or already-approved fixture data.

Place a scenario JSON beside the fixture and app directory:

```json
{
  "name": "Customer summary",
  "appDir": "./dist",
  "fixture": "./customer.json",
  "apiPath": "/api/customer",
  "ready": "#loaded",
  "maxRuns": 100,
  "checks": [
    { "selector": "#name", "text": "Ada" },
    { "selector": ".account-row", "count": 2 },
    { "selector": "#save", "disabled": false }
  ]
}
```

`appDir` and `fixture` resolve relative to the scenario. Optional `entry` defaults to `/`; optional `timeoutMs` defaults to 5000. `ready` must appear after the app has finished rendering, including its error handling. Assertions support exact trimmed `text`, substring `includes`, exact `count` (defaults to 1), and HTML `disabled` presence. Multiple matches join trimmed text with a newline. These are DOM checks, not visibility or accessibility checks. Include assertions for business outcomes and action availability; an existence-only check may preserve a broken screen.

Run `node src/cli.mjs path/to/scenario.json artifacts/name`. Run a different scenario in a separate command; runs own one local fixture server and execute sequentially. Exit 0 means a verified fixture was written; inspect `complete` in the report for budget exhaustion. Exit 2 is a contract/configuration problem; exit 3 is an engine/wiring failure.

## What you receive

- `fixture.min.json`: reduced response that passed a final fresh-page check.
- `report.json`: original response SHA-256, engine discovery, caller-reported source revision, contract, baseline, every candidate observation, final result and controls.
- `controls/*.json`: one deletion from the final fixture that failed the contract. The report names the field and observed difference. Emitted only after a complete reduction.
- `index.html`: a self-contained report with size reduction, failure explanations and artifact links.

The algorithm tries whole object-key deletions before descendants, restarting after each accepted deletion. A complete final pass means no **single remaining object-key deletion** passes. This is greedy, order-dependent 1-minimality, not a globally smallest fixture. Arrays retain length and order; objects inside arrays can shrink. Values are never synthesized or changed.

Each evaluation creates and closes its own browser context and page. Baseline runs twice and must yield identical recorded checks. Accepted reductions get a final fresh-context verification. Transport/protocol errors abort; they never count as proof that a field is needed. A missing ready marker is a consumer failure after navigation succeeded. Known unsupported browser behavior can still look like a consumer failure; baseline controls and the engine identity help bound that evidence.

## Why not just use Playwright or MSW?

[Playwright supports response routing and HAR replay/update](https://playwright.dev/docs/mock). [MSW supports HTTP response mocking](https://mswjs.io/docs/http/mocking-responses/). Both are useful ways to supply a mock. A plain script can delete fields too. Fixture Pruner adds the repeated **consumer-preserving reduction**, a replayable failure map, and an explicit evidence boundary. This is a focused workflow, not a claim to a new mocking primitive or a replacement for those tools.

[Kitesurf](https://github.com/cloudflare/kitesurf) provides the real page/JavaScript engine for every reduction. It earns a role as an independently implemented, local browser evaluator using familiar CDP. This product does not require a Kitesurf-only API and could support another engine later. No fidelity, speed or cost advantage over Chromium has been measured. CDP compatibility does not establish rendering parity.

## Boundaries

Only the provided scenario and assertions are preserved. A field may matter to another route, role, interaction, hidden state, schema validator or browser. Do not replace a shared canonical API response with one scenario's reduced fixture. This is **not redaction**: retained values, console/page errors and reports can contain sensitive data. Empty object keys, arrays, optional branches and fallback behavior are intentionally allowed when the contract permits them. No production traffic was used in the bundled examples.

## Development

```sh
npm test
npm run check
# Optional report presentation check using installed normal Google Chrome:
node scripts/inspect-reports.mjs
```

Tests cover reduction, fallback revisiting, array preservation, budget limits, broken DOM controls, config validation, infrastructure aborts, and static-root containment. There is no transpilation/build step. See `DEMO.md` for the walkthrough and `RESULTS.md` for recorded verification.

Runtime build and startup instructions: [RUNTIME.md](../RUNTIME.md).
