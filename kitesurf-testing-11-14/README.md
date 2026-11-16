# Kitesurf testing tools

Three small developer-tool prototypes using a real local Kitesurf browser through CDP. All bundled applications and fixture data are synthetic.

| Tool | Job | Recorded result |
|---|---|---|
| [Fixture Pruner](fixture-pruner/) | Remove unnecessary JSON mock fields while preserving explicit UI assertions. | Invoice 735 → 150 bytes; warehouse 561 → 95 bytes. |
| [Needle](needle/) | Reduce a stale-response reproduction while preserving the exact DOM mismatch. | Two 11-event schedules reduced to four events each. |
| [Proofgap](proofgap/) | Inject bad API values and identify assertions that miss visible mistakes. | Weak checks miss four faults; stronger checks detect all four. |

Start with [Fixture Pruner's invoice report](fixture-pruner/evidence/invoices/index.html), [Needle's reduction](needle/artifacts/search/report.md), or [Proofgap's audit](proofgap/evidence/2026-11-14/checkout/report.md). GitHub displays HTML source; download or serve the directory locally to view HTML reports.

## Run

Requires Node 22+, npm and a built local Kitesurf checkout. Follow [RUNTIME.md](RUNTIME.md) first, then run from this directory:

```sh
KITESURF_SOURCE=/absolute/path/to/kitesurf ./demo.sh
```

This starts missing local engines on ports 18891–18893, runs all three hands-free demos sequentially, and stops only engines it started. Fixture servers use ports 8891–8893. Fixture Pruner installs its locked Node dependency if missing; the other entries need no npm dependencies. Individual entry READMEs document their input interfaces and endpoint overrides.

Each demo runs deterministic tests, syntax checks and two real browser workloads. Fresh reports go under each entry's `artifacts/`; Needle rewrites its saved sample artifacts. Engines are never silently replaced by Chromium.

## Scope

These are archived prototypes, not maintained packages or arbitrary Playwright-suite adapters. Fixture Pruner supports one load-time JSON API and explicit DOM checks. Needle supports declared actions and a restricted controlled fetch stub. Proofgap supports local static apps with relative JSON APIs and explicit fault/witness definitions. Results do not establish production readiness, cross-browser parity, or a Kitesurf speed advantage. The reduction/auditing policies could use another browser adapter.

See [RESULTS.md](RESULTS.md) for verification and [RULES.md](RULES.md) for the build constraints. [The scorecard](scorecard.md) is intentionally unscored. Code is covered by the archive's [MIT license](../LICENSE); Kitesurf is a separately installed upstream dependency.
