# 60-second demo

1. Run `./demo.sh` with local Kitesurf running.
2. Open `artifacts/report.html`. “Quote ready” and “total exists” both sound reasonable. Both pass the original checkout.
3. Point to **tenfold-total**: the server actually returned 49000 cents. The page actually displayed USD 490.00. The smoke assertions stayed green. The exact total assertion failed.
4. Point to **wrong-currency**: same pattern, a different visible mistake. This is the assertion the reviewer should add, not another screenshot to inspect by hand.
5. Point to **unused-metadata** and **unvisited-discount**. They earn no credit. One changed no observed output; one was never fetched. A runner error similarly cannot count as detection.
6. Open `artifacts/shipping/report.html`: change the weight to three kilograms in the journey. A zero weight rate and a thirty-day delivery promise both evade the smoke test and fail the behavior checks.

The reusable output is the evidence JSON plus the human-readable report. To use it on another small static tool, copy a workload, supply its local app, JSON route and meaningful wrong values, and specify the DOM witness and checks. The product tests a declared assertion contract; it does not import arbitrary Playwright tests or claim whole-app coverage.
