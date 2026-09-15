# Exactly 60 seconds

Preparation, outside the presentation clock: have a terminal open in this repository. No inputs are required once `./demo.sh` starts. The build and complete canned run typically take a few seconds on the recorded machine; if a build is cold, prepare with `go build -o .bin/resume .` first. This is a 60-second speaking and inspection script, not a claim that compilation has a fixed duration.

| Clock | Action and words |
| --- | --- |
| 00–08 (8s) | Run `./demo.sh`. “This worker builds and checks an artifact, including a planted defect. The controller kills real processes at acknowledged boundaries.” |
| 08–20 (12s) | Point to `lost-ack`: effects 1, completions 1. “The sink committed, withheld its receipt, and the worker died. Its replacement queried the same operation key and recovered the receipt. There is one effect.” |
| 20–32 (12s) | Point to `opaque-lost-ack`: effects 1, completions 0. “This sink cannot answer that query. The replacement records an unresolved operation. Even another restart does not repeat the effect or invent completion.” |
| 32–42 (10s) | Point to `late-terminal` and `source-change`. “An old worker's terminal cannot advance the new incarnation. S1 remains history; S2 earns its own evidence and delivery.” |
| 42–51 (9s) | Point to `candidate-change` and the three mutant rows. “Changed bytes cannot borrow old checks. Deliberately broken retries, incarnation checks, and evidence reuse are each caught from artifacts and effect records.” |
| 51–60 (9s) | Point to the printed `ARTIFACTS` path. “The journal, candidate bytes, receipts, failures, and replay commands remain here. This is a one-host worker-crash experiment, with explicit uncertainty where the sink cannot resolve it.” Stop at 60. |

Durations total 60 seconds. The command itself is hands-free and completes without waiting for this narration. For inspection afterward, open the printed run directory's `results.json`, then `lost-ack/process-01.log`, `lost-ack/journal.jsonl`, `opaque-lost-ack/observation.json`, and each mutant's `observation.json`.
