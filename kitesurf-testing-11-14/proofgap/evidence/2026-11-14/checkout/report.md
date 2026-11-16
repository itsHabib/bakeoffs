# Proofgap · Checkout: the wrong total still passes

Engine: Kitesurf/0.0.1. Upstream: 4a2903813a22b08d4f3cc64a4b35339d0293a5d1.

smoke: 2 of 2 observed faults missed · behavior: 0 of 2 observed faults missed

## tenfold-total

api/quote /totalCents → 49000

Witness `#total`: "USD 49.00" → "USD 490.00"

- smoke: **survived**
- behavior: **detected**


## wrong-currency

api/quote /currency → "EUR"

Witness `#total`: "USD 49.00" → "EUR 49.00"

- smoke: **survived**
- behavior: **detected**


## unused-metadata

api/quote /trace → "mutated"

Witness `#total`: "USD 49.00" → "USD 49.00"

- smoke: **no-observed-change**
- behavior: **no-observed-change**


## unvisited-discount

api/discount /percent → 99

Witness `#total`: "USD 49.00" → "USD 49.00"

- smoke: **not-exercised**
- behavior: **not-exercised**
