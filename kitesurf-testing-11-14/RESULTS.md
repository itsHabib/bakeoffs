# Verification

The publication copy was exported from Git's staged tree into a fresh directory and its combined `./demo.sh` was run against a separately built, pinned local Kitesurf checkout. This tests the portable launcher and archived code rather than depending on the original development repositories.

| Entry | Deterministic tests | Real-browser workloads |
|---|---:|---|
| Fixture Pruner | 8 | Invoice 735 → 150 bytes; warehouse 561 → 95 bytes. Final fixtures pass; retained-key deletions produce 8 and 6 failing controls. |
| Needle | 12 | Search and cart each reduce 11 → 4 events. Reduced replay reproduces the exact mismatch; fixed implementations pass. |
| Proofgap | 13 | Weak checks miss four visible faults across checkout and shipping; stronger checks detect all four. Unused/unchanged mutations receive no credit; negative baseline/readiness controls behave as expected. |

All syntax checks and the combined demo passed. See `validation.json` for the engine revision, verification outcome and hash of the executable demo inputs. The final executable inputs were compared with the tested export; subsequent edits were documentation and verification records only.

Historical synthetic reports are included for inspection. Absolute development paths were normalized; raw machine logs, session instructions, private repository history and screenshots are omitted. The public Git history and new files were scanned with Gitleaks with no findings. This is a scan result, not a guarantee that arbitrary future fixtures are safe to publish.

The fixture data and applications were authored for these demos. No real organization or customer dataset is included. These runs establish behavior on the bundled scenarios, not production readiness, browser parity, global minimality or a Kitesurf performance advantage. Each tool's README explains its narrower contract.

To reproduce, follow [RUNTIME.md](RUNTIME.md) and run the combined demo. Outputs are inspectable HTML, Markdown and JSON; the original reduction policies do not use models to judge correctness. The scorecard remains blank.
