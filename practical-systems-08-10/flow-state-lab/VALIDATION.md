# Validation

Validated on 2026-08-10 with Node.js 22.22.2, Java 17.0.16, Quint 0.32.0,
and Apalache 0.56.1 on Windows.

```text
npm run typecheck
  PASS — CamFlow.qnt and KnownFailures.qnt

npm run simulate
  PASS — 10,000 sampled traces, no invariant violation

npm run verify
  PASS — no violation found within 12 steps

npm run fixtures
  PASS — 4 expected counterexamples exported

npm test
  PASS — 4 tests
```

`npm audit` reports two high-severity entries for one underlying, currently
unpatched transitive advisory: Quint 0.32 depends on `adm-zip < 0.6.0`
(GHSA-xcpc-8h2w-3j85). It is confined to this development-only model-checking
toolchain and is not accepted as a production dependency.

The exhaustive claim is bounded to 12 transitions and applies to the Quint
model only. The generated JSON is the explicit handoff point for conformance
tests against running application code.
