# Proofgap

**Find browser checks that stay green when the displayed answer is wrong.**

For the frontend engineer reviewing a checkout or calculator test, Proofgap replaces manually corrupting API fixtures, rerunning a journey, and checking whether the test noticed. It produces a response-to-DOM-to-assertion evidence report. The canned checkout has a plausible green smoke test that misses a tenfold price error and a wrong currency; a behavior check catches both.

Proofgap runs the meaningful page JavaScript and DOM observations in **real Kitesurf**. It does not substitute Chromium or infer success from CDP connectivity. The report's policy is deterministic JavaScript; no models or API keys participate.

## Run

Node 22+ with built-in WebSocket. No npm dependencies or install step.

Start a local [Kitesurf](https://github.com/cloudflare/kitesurf) runtime separately (its first build needs Bazel, pinned pnpm and Rust/Wasm dependencies). The default endpoint uses port 18893; see [runtime setup](../RUNTIME.md).

```sh
./demo.sh
```

This runs tests, syntax checks, and two real browser campaigns with known controls, plus live strong-only, broken-baseline and readiness-error checks. Exit zero requires every declared demo outcome to match the observed outcome. It writes `artifacts/report.html`, `.json`, `.md`, and the corresponding shipping reports under `artifacts/shipping/`. Open the HTML file in your browser.

```sh
KITESURF_ENDPOINT=ws://127.0.0.1:18893/devtools/browser \
KITESURF_SOURCE=/path/to/kitesurf ./demo.sh
```

`KITESURF_SOURCE` records the local checkout revision. It is provenance, not cryptographic proof of the running build. Alternatively supply `KITESURF_REVISION`. The runtime identity always comes from its live `/json/version`; a non-Kitesurf identity is rejected. `PROOFGAP_PORT` defaults to 8893. The demo uses the port sequentially and closes only its own servers and browser contexts.

## Your own workload

Copy `workloads/shipping/workload.json` and its `app/`, then adapt them. This deliberately narrow v0 supports a static local application with relative same-origin JSON API calls, like `fetch('api/rates')`. It serves each run at a unique URL prefix, so use relative script, stylesheet and API URLs. Bundled local assets and classic scripts are the tested format. An existing remote app, SPA absolute asset paths, framework dev server, arbitrary Playwright test, login or real backend is outside this interface.

```sh
node src/cli.mjs /absolute/path/to/workload.json
PROOFGAP_OUT=artifacts/my-audit node src/cli.mjs /path/to/workload.json --enforce
```

The configuration fields are:

- `name`, `app` (directory relative to config), `ready` (selector created after work completes).
- `steps`: `click`, `fill` with `value`, or `wait`, each with a CSS `selector`. The implementation uses DOM `.click()` and input/change events, **not physical pointer or keyboard simulation**.
- `routes`: relative `api/...` path and JSON `body`. Successful JSON responses only; no proxy or live upstream calls.
- `faults`: unique `id`, `route`, JSON `pointer`, replacement `value`, and a `witness` CSS selector expected to change. `why` is optional explanatory text. Faults are explicit developer-chosen counterexamples, not automatically assumed defects.
- `suites`: named arrays of `{selector, op, value}` checks. `op` is `equals`, `includes`, or `exists` (no value). Text is `textContent.trim()`. Suites share the same observation, making weak and strong assertion comparisons directly comparable.
- Optional `timeoutMs` (default 12000) for each selector wait. `expected` is only for `--demo-check` canned controls and never determines an audit verdict.

The runner starts a fresh browser context for baseline and every mutation. API hit records contain the response SHA-256 and whether the mutation was actually served. Reports contain baseline and mutated DOM text, all assertion results, timing and execution errors.

## Verdicts and exit codes

| Verdict | Meaning |
|---|---|
| `survived` | Baseline checks pass, changed response was requested, witness changed, assertions still pass. Review the suggested assertion gap. |
| `detected` | Same prerequisites, but at least one assertion fails. |
| `not-exercised` | No mutated response was requested; no detection credit. |
| `no-observed-change` | Response was served, but declared DOM witness did not change; no detection credit. This does **not** prove semantic equivalence. |
| `baseline-red` | The unmodified page already fails this suite; audit is inconclusive. |
| `baseline-error` / `execution-error` | Navigation, readiness, evaluation or protocol failed; not a killed mutation. |

Normal report mode returns 0 for a completed campaign even when it finds gaps. `--enforce` returns 1 if any declared fault is not detected, including unexercised/unchanged controls. Use a workload containing only relevant faults for this mode. Invalid configuration or execution failures return 2. `--demo-check` returns 0 only when the known controls match exactly. An unavailable engine writes `failure.json` with `verdict: incomplete`; it cannot produce a passing report. Old report files in that output directory are removed before each run to avoid showing stale success.

## Practical alternatives and the extra job

[Playwright API mocking](https://playwright.dev/docs/mock) already provides request routing and response replacement. A short Playwright script is a strong alternative. Proofgap adds the *assertion audit*: baseline gating, controlled fault campaigns, consumed-response evidence, a declared observable witness, suite comparison and precise inconclusive states. This can also be implemented with Playwright; the project claims no unique invention.

[Stryker](https://stryker-mutator.io/docs/) mutates source code and reports surviving mutants. Proofgap mutates JSON at a browser-visible boundary without rebuilding the application. It is narrower and does not replace source mutation testing.

[Kitesurf](https://github.com/cloudflare/kitesurf) executes the app's fetch, JavaScript event handlers and DOM updates. Its dependency earns a concrete role as a small independently implemented browser runtime for local boundary tests. It is **not technically necessary** for the audit policy; Chromium could run the same campaign through another adapter. No speed, fidelity or resource advantage is claimed or measured here.

## Limits

The developer declares both the mutation and its expected witness. A weak/wrong witness can hide a real effect; an irrelevant mutation can produce an irrelevant gap. A changed text witness establishes the declared DOM difference, not universal business correctness. Results apply only to the specified journey and faults, not overall test quality. Readiness must mean the app has finished the relevant work; this tool does not discover asynchronous quiescence. DOM actions and text checks do not establish physical interaction, CSS visibility, accessibility or production browser parity. External resources in a user-supplied application are not sandboxed by Proofgap: use self-contained local assets for this local workflow.

`npm test` tests policy, JSON pointer mutation, real local HTTP server isolation and HTML escaping. `npm run check` checks syntax. `./demo.sh` additionally requires real Kitesurf. See [DEMO.md](DEMO.md) for the 60-second story and [RESULTS.md](RESULTS.md) for this build's evidence.

Runtime build and startup instructions: [RUNTIME.md](../RUNTIME.md).
