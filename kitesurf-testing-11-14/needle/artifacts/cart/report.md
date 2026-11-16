# Cart quote belongs to the previous quantity

Reduced 11 events to 4.

Expected `8`; observed `2`.

Engine: Kitesurf (CDP product Chrome/145.0.0.0; marker @kitesurf), source revision 4a2903813a22b08d4f3cc64a4b35339d0293a5d1.

## Replay

`node src/cli.mjs replay artifacts/cart/minimal.json`

## Remaining events

- older-request: fill
- current-request: fill
- current-response: respond
- stale-response: respond

## Deletion witnesses

- Remove older-request: pass
- Remove current-request: pass
- Remove current-response: unresolved (1 fetch request(s) remain undelivered)
- Remove stale-response: unresolved (1 fetch request(s) remain undelivered)

Each single event deletion, including dependent closure, was tested twice and did not consistently reproduce the exact mismatch in both runs. Unresolved witnesses do not establish semantic necessity; unstable witnesses are inconclusive.
