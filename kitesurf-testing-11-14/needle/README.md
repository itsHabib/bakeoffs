# Needle

**Shrink a stale-response bug into a replay a teammate can understand.**

For a frontend engineer debugging “search results went backwards” or “the cart total belongs to the previous quantity,” Needle removes unrelated clicks, inputs and fetch deliveries, then reruns the remaining sequence in **real local Kitesurf**. It exports the reduced JSON, the exact DOM mismatch, every attempted reduction and a deletion witness for each retained event.

No dependencies beyond Node 22+. No cloud account or paid call. The browser runtime must already be running; Needle never substitutes Chromium for Kitesurf.

```sh
./demo.sh
```

This runs policy tests, reduces two different pages, replays both reduced failures, and checks the corresponding fixed pages. Read `artifacts/search/report.md` and `artifacts/cart/report.md`. Demo server port: 8892. Engine endpoint defaults to `ws://127.0.0.1:18892/devtools/browser`; override with `KITESURF_ENDPOINT`. `KITESURF_REVISION` records the caller-supplied source revision, separately from the engine's live CDP identity.

## Bring a local page

Start your application yourself. Copy `examples/search.json`, change the URL, actions and oracle, then run:

```sh
node src/cli.mjs reduce my-scenario.json artifacts/my-bug
node src/cli.mjs replay artifacts/my-bug/minimal.json
```

Only loopback app URLs and engine endpoints are accepted. Input JSON is trusted local input: your application still runs its own scripts and can make network requests outside the stub's scope. This is a reproduction tool, not a network sandbox.

The narrow input interface:

- `fill`: CSS `selector`, string `value`; sets `.value` and dispatches a bubbling `input` event.
- `click`: CSS `selector`; invokes the DOM element's `.click()`.
- `respond`: exact `url` as passed to `fetch`, JSON `body`; resolves the oldest pending request to that URL.
- `requires`: optional IDs of earlier prerequisite events. Deleting a prerequisite deletes its dependent events transitively. Declare a response's triggering action here.
- `oracle.actual` and `oracle.expected`: a CSS selector plus either `attribute`, `property`, or neither (text content). They are read from the rendered page DOM and compared as strings.
- `readySelector`: optional page readiness marker; defaults to the expected-value selector. Use a marker that appears after application handlers are installed.

The fetch stub is installed after navigation, before actions. It controls **subsequent calls to `window.fetch`**. It implements only successful `json()` and `text()` responses. All issued requests must be delivered before the oracle can judge a run. Initial-load fetch, cached fetch references, XHR, service workers, WebSockets, streaming bodies, HTTP failure semantics, AbortSignal behavior and native keyboard/pointer fidelity are outside this version. No sleep durations are used to simulate the response race: delivery order is explicit.

## What the evidence establishes

Needle requires the full input to produce the same mismatch three times. Every candidate is run twice in fresh page targets. Only the **same actual and expected strings** count as the original failure; a different failure cannot silently replace it. Missing targets, absent requests and undelivered responses are `unresolved`; inconsistent repeats are `unstable`. CDP, navigation and page execution errors stop the run.

Reduction removes chunks, increases granularity, and finishes by testing each remaining event deletion and its dependency closure again. The artifact reports the result of every single deletion **within this event/dependency model**; it does not claim a globally shortest reproduction. An unresolved deletion witness establishes a validity constraint, not semantic necessity; an unstable witness is inconclusive. Only pass or different-failure witnesses establish that a valid removal did not reproduce the original mismatch in those runs. Two or three repeats detect some instability; they cannot prove a race is deterministic under all conditions. The failure is a declared DOM invariant, not an inferred product requirement.

## Why this instead of a trace or HAR?

[Playwright can mock individual routes and replay HAR files](https://playwright.dev/docs/mock). A plain Playwright script is the practical baseline for writing and replaying the four-event race directly. Needle adds automated reduction, strict failure preservation, repeated probes, prerequisite handling and inspectable deletion evidence. Its minimization comes from established [delta debugging](https://www.cs.columbia.edu/~junfeng/08fa-e6998/sched/readings/delta-debug-2.pdf); it does not claim a new algorithm.

[Kitesurf](https://github.com/cloudflare/kitesurf) executes the actual page JavaScript, DOM events and promise continuations for every meaningful candidate. Its role here is a local non-Chromium execution target. Needle does not establish that Kitesurf is faster, smaller or more faithful than Chromium, and the reduction algorithm does not inherently require it. The dependency earns its place when this is the engine you need to reproduce against; Chromium parity needs separate verification.

## Development

```sh
npm test
npm run check
node src/server.mjs             # inspect http://127.0.0.1:8892/search or /cart
EXPECT_STATUS=pass node src/cli.mjs replay examples/search-fixed.json
```

The Node-only tests cover multi-event causes, dependency closure, malformed inputs, exact failure signatures, unresolved probes, unstable runs, infrastructure failures and nonmonotonic predicates. The demo is the real-engine integration check. See `RESULTS.md` for captured outcomes and `DEMO.md` for the 60-second walkthrough.

Runtime build and startup instructions: [RUNTIME.md](../RUNTIME.md).
